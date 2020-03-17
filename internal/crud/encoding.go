package crud

import (
	"encoding/json"
	"os"

	"github.com/bejaneps/go-git-webapp/internal/util"
	"github.com/pkg/errors"
)

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
