package crud

import (
	"encoding/json"
	"os"
	"regexp"

	"github.com/bejaneps/go-git-webapp/internal/util"
	"github.com/pkg/errors"
)

// RegexpConfig is a struct that holds a regexp info about each config file type.
type RegexpConfig struct {
	X map[string]string `json:"-"`
}

// ToJSON returns a json representation of a Git collection in a file.
func (c *GitCollection) ToJSON() (*os.File, error) {
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

// Filter applies regexp on content of each config file that is specified, and returns new collection with filtered result.
func (c *GitCollection) Filter(rc *RegexpConfig) (*GitCollection, error) {
	op := "crud.GitCollectionMatch"

	newColl := &GitCollection{
		BaseURL:  c.BaseURL,
		BaseHash: c.BaseHash,
		BaseDir:  c.BaseDir,
	}

	var err error
	reg := &regexp.Regexp{}
	for _, f := range c.Coll {
		for key, val := range rc.X {
			reg, err = regexp.Compile(val)
			if err != nil {
				return nil, errors.Wrapf(err, "(%s): invalid %s regexp", op, key)
			}

			if reg.MatchString(f.Name) {
				f.Type = key // make the type same as a key of regex
				newColl.Coll = append(newColl.Coll, f)
				break
			}
		}
	}

	return newColl, nil
}
