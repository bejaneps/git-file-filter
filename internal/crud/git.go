package crud

import (
	"context"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/ghodss/yaml"

	jsoniter "github.com/json-iterator/go"
	"github.com/open-policy-agent/opa/rego"

	"github.com/bejaneps/go-git-webapp/internal/util"
	"github.com/pkg/errors"
	"gopkg.in/src-d/go-git.v4/plumbing"

	"gopkg.in/src-d/go-git.v4/plumbing/object"

	git "gopkg.in/src-d/go-git.v4"
)

var (
	json = jsoniter.ConfigCompatibleWithStandardLibrary
)

var reposDir = "repositories"

// file holds info individual files commit hash and names
type file struct {
	Hash    string `json:"-"`
	Name    string `json:"name"`
	Type    string `json:"type"`
	URL     string `json:"-"`
	Content string `json:"content"`
	Config  bool   `json:"-"`

	Reader io.ReadCloser `json:"-"`
}

// GitCollection is a struct that holds a commit hash and filename in a git repository
type GitCollection struct {
	BaseURL  string `json:"-"`
	BaseHash string `json:"-"`
	BaseDir  string `json:"-"`

	PolicyFile string `json:"-"` // string representation of content of a .rego file

	Coll []file `json:"file"`
}

// Config holds an info about each config file filtering, name: "Docker", filter: "\bDockerfile\b", policy: "https://example.com/1"
type Config struct {
	Name   string `json:"name"`
	Filter string `json:"filter"`
	Policy string `json:"policy"`
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

	// retrieve files from specific dir
	if dir != "" {
		coll.Coll, err = retrieveFromDir(url, coll.BaseHash, dir, tree)
		if err != nil {
			return nil, errors.Wrapf(err, "(%s): retrieving list of files", op)
		}

		coll.BaseDir = dir
	} else { // retrieve files from root dir
		coll.Coll, err = retrieveFromRoot(url, coll.BaseHash, tree)
		if err != nil {
			return nil, errors.Wrapf(err, "(%s): retrieving list of files", op)
		}

		coll.BaseDir = "/"
	}

	// search for a policy file
	content, err := findPolicy(tree)
	if err != nil {
		return nil, errors.Wrapf(err, "(%s): searching a .rego file", op)
	}
	coll.PolicyFile = content // for applying policy

	// retrieve files from root
	return coll, nil
}

// retrieveFromDir returns a collection that has all files from a repo in a specific dir,
// it also marks if a file is a config type.
func retrieveFromDir(url, hash, dir string, tree *object.Tree) ([]file, error) {
	var coll []file
	var err error

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

			coll = append(coll, co)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return coll, nil
}

// retrieveFromRoot returns a collection that has all files from a repo in a root dir,
// it also marks if a file is a config type.
func retrieveFromRoot(url, hash string, tree *object.Tree) ([]file, error) {
	var coll []file
	var err error

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

		coll = append(coll, co)

		return nil
	})
	if err != nil {
		return nil, err
	}

	return coll, nil
}

// findPolicy searches for a policy .rego file in a git repository,
// if it finds a file, it returns a file's content converted to string,
// else nil.
func findPolicy(tree *object.Tree) (content string, err error) {
	tree.Files().ForEach(func(f *object.File) error {
		if strings.Contains(f.Name, ".rego") {
			content, err = f.Contents()
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

			if conf.Policy == "" {
				policy = c.PolicyFile
			} else { // get a policy from a url
				policy, err = getPolicyFromURL(conf.Policy)
				if err != nil {
					return nil, errors.Wrapf(err, "(%s): retrieving policy file from %s", op, conf.Policy)
				}
			}

			input, err := toJSON(coll.Name, coll.Reader)
			if err != nil {
				return nil, errors.Wrapf(err, "(%s): converting %s file to json", op, coll.Name)
			} else if len(input) == 0 {
				continue
			}

			// create a new rego object
			r := rego.New(
				rego.Query("data"),
				rego.Module(conf.Policy, policy),
				rego.Input(input),
			)

			ctx := context.Background()

			rs, err := r.Eval(ctx)
			if err != nil {
				return nil, errors.Wrapf(err, "(%s): evaluating a query of a %s", op, coll.Name)
			}

			if len(rs) != 0 {
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

// toJSON tries to convert a file to compatible JSON format.
func toJSON(name string, rc io.ReadCloser) ([]byte, error) {
	// obtain bytes of an input file for unmarshaling
	bs, err := ioutil.ReadAll(rc)
	if err != nil {
		return nil, err
	} else if len(bs) == 0 {
		return nil, errors.Errorf("no bytes read from file %s", name)
	}

	var js []byte
	if util.In([]string{".yaml", ".yml"}, name) {
		js, err = yaml.YAMLToJSON(bs)
		if err != nil {
			return nil, errors.New("converting yaml to json: " + name)
		}
	}

	return js, nil
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
