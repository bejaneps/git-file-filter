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

// GitCollection is a struct that holds a commit hash and filename in a git repository
type GitCollection struct {
	Hash string
	File string
}

// GetGitCollections returns a slice of git collection
func GetGitCollections(url, hash, dir string) ([]GitCollection, error) {
	var op = "crud.GetGitCollections"

	// initialize vars, so we don't recreate them
	r := &git.Repository{}
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
		return retrieveFromDir(tree, dir), nil
	}

	// retrieve files from root
	return retrieveFromRoot(tree), nil
}

func retrieveFromDir(tree *object.Tree, dir string) []GitCollection {
	var coll []GitCollection

	tree.Files().ForEach(func(f *object.File) error {
		if strings.Contains(f.Name, dir) {
			co := GitCollection{}

			co.File = f.Name[strings.Index(f.Name, "/")+1:]
			co.Hash = f.Hash.String()

			coll = append(coll, co)
		}
		return nil
	})

	return coll
}

func retrieveFromRoot(tree *object.Tree) []GitCollection {
	var coll []GitCollection

	tree.Files().ForEach(func(f *object.File) error {
		co := GitCollection{}

		co.File = f.Name[strings.Index(f.Name, "/")+1:]
		co.Hash = f.Hash.String()

		coll = append(coll, co)

		return nil
	})

	return coll
}
