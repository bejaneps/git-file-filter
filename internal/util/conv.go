package util

import (
	json "github.com/json-iterator/go"

	"io"
	"io/ioutil"

	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
)

// ErrUnsupportedFileType is used when a config file is different than the ones that are supported.
var ErrUnsupportedFileType = errors.New("unsupported file type")

// ToJSON tries to convert a file to compatible JSON format.
func ToJSON(name string, rc io.ReadCloser) ([]byte, error) {
	op := "util.ToJSON"

	// obtain bytes of an input file for unmarshaling
	bs, err := ioutil.ReadAll(rc)
	if err != nil {
		return nil, errors.Wrapf(err, "(%s): reading file %s", op, name)
	} else if len(bs) == 0 {
		return nil, errors.Errorf("no bytes read from file %s", name)
	}

	var js []byte
	if In([]string{".json"}, name) { // for json
		return bs, nil
	} else if In([]string{".yaml", ".yml"}, name) { // for yaml
		js, err = yaml.YAMLToJSON(bs)
		if err != nil {
			return nil, err
		}
	} else if In([]string{".tf"}, name) { // for terraform
		var content interface{}
		content, err = getHclJSON(bs, name)
		if err != nil {
			return nil, err
		}

		js, err = json.Marshal(content)
		if err != nil {
			return nil, err
		}
	} else { // for unsupported types
		return nil, ErrUnsupportedFileType
	}

	return js, nil
}
