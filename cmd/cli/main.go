package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/pkg/errors"
	"gopkg.in/src-d/go-git.v4/plumbing"

	"gopkg.in/src-d/go-git.v4/plumbing/object"

	git "gopkg.in/src-d/go-git.v4"
)

// example freelancer: micronaut-projects/micronaut-examples/9669e10633ec7bf81488952d015bf36e900f8bca/hello-world-java
// example arg: https://github.com/micronaut-projects/micronaut-examples 9669e10633ec7bf81488952d015bf36e900f8bca hello-world-java

func main() {
	// check if args are empty
	if len(os.Args) < 1 || len(os.Args) != 4 {
		log.Fatalf("usage: .%s url commit_hash directory", os.Args[0])
	}

	url := os.Args[1]
	hash := os.Args[2]
	dir := os.Args[3]

	// initialize vars, so we don't recreate them
	r := &git.Repository{}
	var err error

	// clone a repo, cleanups directory name filepath.Clean() implicitly
	r, err = git.PlainClone(url, false, &git.CloneOptions{
		URL: url,
	})
	if errors.Cause(err) == git.ErrRepositoryAlreadyExists { // check if repo exists, then just open it
		r, err = git.PlainOpen(url)
		if err != nil {
			log.Fatalf("[ERROR]: %v", err)
		}
	} else if err != nil {
		log.Fatalf("[ERROR]: %v", err)
	}

	// retrieve a commit
	commit, err := r.CommitObject(plumbing.NewHash(hash))
	if err != nil {
		log.Fatalf("[ERROR]: %v", err)
	}

	tree, err := commit.Tree()
	if err != nil {
		log.Fatalf("[ERROR]: %v", err)
	}

	// print just files in a given directory
	tree.Files().ForEach(func(f *object.File) error {
		if strings.Contains(f.Name, dir) {
			name := f.Name[strings.Index(f.Name, "/")+1:]
			fmt.Printf("Hash: %s\t File: %s\n", f.Hash, name)
		}
		return nil
	})
}
