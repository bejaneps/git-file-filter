package crud

import (
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"gopkg.in/src-d/go-git.v4/plumbing"

	"gopkg.in/src-d/go-git.v4/plumbing/object"

	git "gopkg.in/src-d/go-git.v4"
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
}

// GitCollection is a struct that holds a commit hash and filename in a git repository
type GitCollection struct {
	BaseURL  string `json:"-"`
	BaseHash string `json:"-"`
	BaseDir  string `json:"-"`

	Coll []file `json:"file"`
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

	// retrieve files from root
	return coll, nil
}

// retrieveFromDir returns a collection that has all files from a repo in a specific dir,
// it also marks if a file is a config type.
func retrieveFromDir(url, hash, dir string, tree *object.Tree) ([]file, error) {
	var coll []file
	var err error

	tree.Files().ForEach(func(f *object.File) error {
		if strings.Contains(f.Name, dir) {
			co := file{}

			co.Hash = f.Hash.String()

			co.Name = f.Name[strings.Index(f.Name, "/")+1:]
			co.URL = url + "/blob/" + hash + "/" + dir + "/" + co.Name

			co.Content, err = f.Contents()
			if err != nil {
				return err
			}

			coll = append(coll, co)
		}

		return nil
	})

	return coll, nil
}

// retrieveFromRoot returns a collection that has all files from a repo in a root dir,
// it also marks if a file is a config type.
func retrieveFromRoot(url, hash string, tree *object.Tree) ([]file, error) {
	var coll []file
	var err error

	tree.Files().ForEach(func(f *object.File) error {
		co := file{}

		co.Hash = f.Hash.String()

		co.Name = f.Name
		co.URL = url + "/blob/" + hash + "/" + f.Name

		co.Content, err = f.Contents()
		if err != nil {
			return err
		}

		coll = append(coll, co)

		return nil
	})

	return coll, nil
}
