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
	Docker    string `json:"Docker"`
	Terraform string `json:"Terraform"`
	Manifest  string `json:"Manifest"`
	Gradle    string `json:"Gradle"`
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
	op := "crud.ConfigFileMatch"

	newColl := &GitCollection{
		BaseURL:  c.BaseURL,
		BaseHash: c.BaseHash,
		BaseDir:  c.BaseDir,
	}

	var err error
	reg := &regexp.Regexp{}
	for _, v := range c.Coll {
		if rc.Docker != "" {
			reg, err = regexp.Compile(rc.Docker)
			if err != nil {
				return nil, errors.Wrapf(err, "(%s): invalid docker regexp", op)
			}

			if reg.MatchString(v.Name) {
				newColl.Coll = append(newColl.Coll, v)
				continue
			}
		}
		if rc.Gradle != "" {
			reg, err = regexp.Compile(rc.Gradle)
			if err != nil {
				return nil, errors.Wrapf(err, "(%s): invalid gradle regexp", op)
			}

			if reg.MatchString(v.Name) {
				newColl.Coll = append(newColl.Coll, v)
				continue
			}
		}
		if rc.Terraform != "" {
			reg, err = regexp.Compile(rc.Terraform)
			if err != nil {
				return nil, errors.Wrapf(err, "(%s): invalid terraform regexp", op)
			}
			if reg.MatchString(v.Name) {
				newColl.Coll = append(newColl.Coll, v)
				continue
			}
		}
		if rc.Manifest != "" {
			reg, err = regexp.Compile(rc.Manifest)
			if err != nil {
				return nil, errors.Wrapf(err, "(%s): invalid manifest regexp", op)
			}
			if reg.MatchString(v.Name) {
				newColl.Coll = append(newColl.Coll, v)
			}
		}
	}

	return newColl, nil
}
