package crud

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	jsoniter "github.com/json-iterator/go"
	"github.com/open-policy-agent/opa/rego"

	"github.com/bejaneps/go-git-webapp/internal/util"
	"github.com/pkg/errors"
	"gopkg.in/src-d/go-git.v4/plumbing"

	"gopkg.in/src-d/go-git.v4/plumbing/object"

	git "gopkg.in/src-d/go-git.v4"

	enry "github.com/src-d/enry/v2"
)

var (
	json = jsoniter.ConfigCompatibleWithStandardLibrary
)

var (
	reposDir      = "repositories"
	defaultPolicy = filepath.Join("config", "default.rego")
)

// file holds info individual files commit hash and names
type file struct {
	Hash      string `json:"-"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	URL       string `json:"-"`
	Config    bool   `json:"-"`
	Extension string `json:"extension"`

	OutputPolicy  string `json:"output_policy"`  // output of the opa applied
	AppliedPolicy string `json:"applied_policy"` // name of the policy

	Content string `json:"content"`

	Reader io.ReadCloser `json:"-"`
}

// GitCollection is a struct that holds a commit hash and filename in a git repository
type GitCollection struct {
	BaseURL  string `json:"-"`
	BaseHash string `json:"-"`
	BaseDir  string `json:"-"`

	FileCount                   int      `json:"file_count"`
	ProgrammingLanguages        []string `json:"file_extensions"`
	UnknownProgrammingLanguages []string `json:"unknown_file_extensions"`

	Policy *file `json:"-"` // string representation of content of a .rego file

	Coll []file `json:"file"`
}

// Config holds an info about each config file filtering, name: "Docker", filter: "\bDockerfile\b", policy: "https://example.com/1"
type Config struct {
	Name      string `json:"name"`
	Filter    string `json:"filter"`
	PolicyURL string `json:"policy"`
}

// GetGitCollection returns a filled GitCollection struct
func GetGitCollection(url, hash, dir string) (*GitCollection, error) {
	var op = "crud.GetGitCollection"

	// initialize vars, so we don't recreate them
	r := &git.Repository{}
	coll := &GitCollection{}
	var err error

	// join and cleanup dir where all repos will be saved
	path := filepath.Join(reposDir, filepath.Clean(url))

	// clone a repo, cleanups directory name filepath.Clean() implicitly
	r, err = git.PlainClone(path, false, &git.CloneOptions{
		URL: url,
	})
	if errors.Cause(err) == git.ErrRepositoryAlreadyExists { // check if repo exists, then just open it
		r, err = git.PlainOpen(path)
		if err != nil {
			return nil, errors.Wrapf(err, "(%s): opening a git repo", op)
		}
	} else if err != nil {
		return nil, errors.Wrapf(err, "(%s): cloning a git repo", op)
	}
	coll.BaseURL = url // for template

	// convert user supplied hash to SHA1 git hash
	var gitHash plumbing.Hash
	if hash != "" {
		gitHash = plumbing.NewHash(hash)
	} else {
		ref, err := r.Head()
		if err != nil {
			return nil, errors.Wrapf(err, "(%s): retrieving head commit", op)
		}
		gitHash = ref.Hash()
	}
	coll.BaseHash = gitHash.String() // for template

	// retrieve a commit
	commit, err := r.CommitObject(gitHash)
	if err != nil {
		return nil, errors.Wrapf(err, "(%s): retrieving a commit object", op)
	}

	// retreive a file structure of specific commit
	tree, err := commit.Tree()
	if err != nil {
		return nil, errors.Wrapf(err, "(%s): retrieving a commit file structure", op)
	}

	var name, content string
	// retrieve files from specific dir
	if dir != "" {
		coll.Coll, coll.FileCount, coll.ProgrammingLanguages, coll.UnknownProgrammingLanguages, err = retrieveFromDir(url, coll.BaseHash, dir, tree)
		if err != nil {
			return nil, errors.Wrapf(err, "(%s): retrieving list of files", op)
		}

		// search for a policy file
		name, content, err = findPolicyFromDir(dir, tree)
		if err != nil {
			return nil, errors.Wrapf(err, "(%s): searching a .rego file", op)
		}

		coll.BaseDir = dir
	} else { // retrieve files from root dir
		coll.Coll, coll.FileCount, coll.ProgrammingLanguages, coll.UnknownProgrammingLanguages, err = retrieveFromRoot(url, coll.BaseHash, tree)
		if err != nil {
			return nil, errors.Wrapf(err, "(%s): retrieving list of files", op)
		}

		// search for a policy file
		name, content, err = findPolicyFromRoot(tree)
		if err != nil {
			return nil, errors.Wrapf(err, "(%s): searching a .rego file", op)
		}

		coll.BaseDir = "/"
	}
	if name != "" && content != "" { // for applying policy
		coll.Policy = &file{}
		coll.Policy.Content = content
		coll.Policy.Name = name
	}

	// retrieve files from root
	return coll, nil
}

// retrieveFromDir returns a collection that has all files from a repo in a specific dir.
// It returns the collection of files in a git specific dir, the count of files, the slice of programming languages, and slice of unknown pr langs.
func retrieveFromDir(url, hash, dir string, tree *object.Tree) ([]file, int, []string, []string, error) {
	var coll []file
	var err error
	var count int
	var langs []string
	var unknownLangs []string

	err = tree.Files().ForEach(func(f *object.File) error {
		if strings.Contains(f.Name, dir) {
			co := file{}

			co.Hash = f.Hash.String()

			co.Name = f.Name[strings.Index(f.Name, "/")+1:]
			co.URL = url + "/blob/" + hash + "/" + dir + "/" + co.Name

			co.Reader, err = f.Reader()
			if err != nil {
				return err
			}

			co.Content, err = f.Contents()
			if err != nil {
				return err
			}

			co.Extension, _ = enry.GetLanguageByExtension(f.Name)
			if co.Extension == "" { // if can't determine ext by name then lookup it's content
				co.Extension, _ = enry.GetLanguageByContent(f.Name, []byte(co.Content))
				if co.Extension == "" {
					unknownLangs = append(unknownLangs, f.Name)
					co.Extension = "Unknown"
				}
			}

			// check if language is already added to list, if no then add it
			if !util.In(langs, co.Extension) {
				langs = append(langs, co.Extension)
			}

			coll = append(coll, co)
			count++
		}

		return nil
	})
	if err != nil {
		return nil, 0, nil, nil, err
	}

	return coll, count, langs, unknownLangs, nil
}

// retrieveFromRoot returns a collection that has all files from a repo in a root dir.
// It returns the collection of files in a git root dir, the count of files, the slice of programming languages, and slice of unknown pr langs.
func retrieveFromRoot(url, hash string, tree *object.Tree) ([]file, int, []string, []string, error) {
	var coll []file
	var err error
	var count int
	var langs []string
	var unknownLangs []string

	err = tree.Files().ForEach(func(f *object.File) error {
		co := file{}

		co.Hash = f.Hash.String()

		co.Name = f.Name
		co.URL = url + "/blob/" + hash + "/" + f.Name

		co.Reader, err = f.Reader()
		if err != nil {
			return err
		}

		co.Content, err = f.Contents()
		if err != nil {
			return err
		}

		co.Extension, _ = enry.GetLanguageByExtension(f.Name)
		if co.Extension == "" { // if can't determine ext by name then lookup it's content
			co.Extension, _ = enry.GetLanguageByContent(f.Name, []byte(co.Content))
			if co.Extension == "" {
				unknownLangs = append(unknownLangs, f.Name)
				co.Extension = "Unknown"
			}
		}

		// check if language is already added to list, if no then add it
		if !util.In(langs, co.Extension) {
			langs = append(langs, co.Extension)
		}

		coll = append(coll, co)
		count++

		return nil
	})
	if err != nil {
		return nil, 0, nil, nil, err
	}

	return coll, count, langs, unknownLangs, nil
}

// findPolicy searches for a policy .rego file in a git repository specific dir,
// if it finds a file, it returns a file's content converted to string,
// else nil.
func findPolicyFromDir(dir string, tree *object.Tree) (name, content string, err error) {
	tree.Files().ForEach(func(f *object.File) error {
		if strings.Contains(f.Name, dir) {
			if strings.Contains(f.Name, ".rego") {
				content, err = f.Contents()

				name = f.Name

				return nil
			}
		}
		return nil
	})

	return
}

// findPolicy searches for a policy .rego file in a git repository,
// if it finds a file, it returns a file's content converted to string,
// else nil.
func findPolicyFromRoot(tree *object.Tree) (name, content string, err error) {
	tree.Files().ForEach(func(f *object.File) error {
		if strings.Contains(f.Name, ".rego") {
			content, err = f.Contents()

			name = f.Name

			return nil
		}
		return nil
	})

	return
}

// Filter applies regexp on content of each config file that is specified, and returns new collection with filtered result.
func (c *GitCollection) Filter(confs []Config) (*GitCollection, error) {
	op := "crud.GitCollectionFilter"

	newColl := &GitCollection{
		BaseURL:  c.BaseURL,
		BaseHash: c.BaseHash,
		BaseDir:  c.BaseDir,
	}

	var err error
	reg := &regexp.Regexp{}
	for _, coll := range c.Coll {
		for _, conf := range confs {
			// 1: Filter by regex
			reg, err = regexp.Compile(conf.Filter)
			if err != nil {
				return nil, errors.Wrapf(err, "(%s): invalid %s regexp", op, conf.Name)
			}

			if !reg.MatchString(coll.Name) {
				continue
			}
			coll.Type = conf.Name // make the type same as a name of regex

			// 2: Filter by policy
			var policy string
			var err error

			if conf.PolicyURL != "" { // get a policy from url
				policy, err = getPolicyFromURL(conf.PolicyURL)
				if err != nil {
					return nil, errors.Wrapf(err, "(%s): retrieving policy file from %s", op, conf.PolicyURL)
				}

				coll.AppliedPolicy = conf.PolicyURL
			} else if c.Policy != nil { // use a policy from git repo
				policy = c.Policy.Content

				coll.AppliedPolicy = c.Policy.Name
			} else { // use default policy
				temp, err := ioutil.ReadFile(defaultPolicy)
				if err != nil {
					return nil, errors.Wrapf(err, "(%s): reading default policy file", op)
				}

				policy = string(temp)

				coll.AppliedPolicy = "not found"
			}

			input, err := util.ToJSON(coll.Name, coll.Reader) // convert a config file to json, and then pass it to OPA.s
			if errors.Cause(err) == util.ErrUnsupportedFileType {
				continue
			} else if err != nil || len(input) == 0 || input == nil {
				return nil, errors.Wrapf(err, "(%s): converting %s file to json", op, coll.Name)
			}

			// create a new rego object
			r := rego.New(
				rego.Query("data"),
				rego.Module(conf.PolicyURL, policy),
				rego.Input(input),
			)

			ctx := context.Background()

			// evaluate a policy and query on config file
			rs, err := r.Eval(ctx)
			if err != nil {
				return nil, errors.Wrapf(err, "(%s): evaluating a query of a %s", op, coll.Name)
			}

			// display all variables from rego file if result set isn't 0
			if len(rs) != 0 {
				for _, busu := range rs {
					for _, miki := range busu.Expressions {
						mapi := miki.Value.(map[string]interface{})
						for _, pur := range mapi {
							mupi := pur.(map[string]interface{})
							for ind, shind := range mupi {
								coll.OutputPolicy = fmt.Sprintf("%s: %v", ind, shind)
							}
						}
					}
				}

				newColl.Coll = append(newColl.Coll, coll)
			}
		}
	}

	return newColl, nil
}

// getPolicyFromURL retrieves a policy from url
// and returns it's content in string format.
func getPolicyFromURL(url string) (content string, err error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	buf := &strings.Builder{}
	b, err := io.Copy(buf, resp.Body)
	if err != nil {
		return "", err
	} else if b == 0 {
		return "", errors.New("no bytes copied from response")
	}

	return buf.String(), nil
}

// ToJSONFile returns a json representation of a Git collection in a file.
func (c *GitCollection) ToJSONFile() (*os.File, error) {
	op := "crud.GitCollectionToJSON"

	// create a json file for serving
	f, err := os.Create(util.RandomString(10) + ".json")
	if err != nil {
		return nil, errors.Wrapf(err, "(%s): creating new json file", op)
	}

	// create a new json file with config files
	err = json.NewEncoder(f).Encode(c)
	if err != nil {
		return nil, errors.Wrapf(err, "(%s): encoding file to json", op)
	}

	return f, nil
}
