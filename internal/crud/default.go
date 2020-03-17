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

// collection holds info individual files commit hash and names
type collection struct {
	Hash string
	File string
}

// GitCollection is a struct that holds a commit hash and filename in a git repository
type GitCollection struct {
	BaseURL  string
	BaseHash string
	BaseDir  string

	Coll []collection
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
		tree.Files().ForEach(func(f *object.File) error {
			if strings.Contains(f.Name, dir) {
				co := collection{}

				co.File = f.Name[strings.Index(f.Name, "/")+1:]
				co.Hash = f.Hash.String()

				coll.Coll = append(coll.Coll, co)
			}

			return nil
		})

		coll.BaseDir = dir
	} else { // retrieve files from root dir
		tree.Files().ForEach(func(f *object.File) error {
			co := collection{}

			co.File = f.Name
			co.Hash = f.Hash.String()

			coll.Coll = append(coll.Coll, co)

			return nil
		})

		coll.BaseDir = "/"
	}

	// retrieve files from root
	return coll, nil
}
