package sub

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/bejaneps/go-git-webapp/internal/crud"

	jsoniter "github.com/json-iterator/go"
)

var (
	json = jsoniter.ConfigCompatibleWithStandardLibrary

	c = &cache{}
)

// temp cache to save data between requests
type cache struct {
	*crud.GitCollection
}

func (e *env) handleRegexpGET(w http.ResponseWriter, r *http.Request) {
	pattern := r.FormValue("pattern")

	var err error

	// convert all backslashes to forward slashes in string
	pattern = strings.ReplaceAll(pattern, "\\", "/")

	// decode query value to json
	rt := &crud.RegexpConfig{}
	err = json.NewDecoder(strings.NewReader(pattern)).Decode(rt)
	if err != nil {
		e.displayError(w, err, http.StatusInternalServerError)
		return
	}

	// filter files by regexp
	coll, err := c.Filter(rt)
	if err != nil {
		e.displayError(w, err, http.StatusInternalServerError)
		return
	}

	// create a json file
	f, err := coll.ToJSON()
	if err != nil {
		e.displayError(w, err, http.StatusInternalServerError)
		return
	}

	// create a download link for json file
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", f.Name()))
	w.Header().Set("Content-Type", "application/json")
	http.ServeContent(w, r, f.Name(), time.Now(), f)
}

func (e *env) handleRegexpPOST(w http.ResponseWriter, r *http.Request) {
	// get json file from request
	file, _, err := r.FormFile("regFile")
	if err != nil {
		e.displayError(w, err, http.StatusInternalServerError)
		return
	}
	defer file.Close()

	// decode file into json struct
	rt := &crud.RegexpConfig{}
	err = json.NewDecoder(file).Decode(rt)
	if err != nil {
		e.displayError(w, err, http.StatusInternalServerError)
		return
	}

	// filter files by regexp
	coll, err := c.Filter(rt)
	if err != nil {
		e.displayError(w, err, http.StatusInternalServerError)
		return
	}

	// create a json file
	f, err := coll.ToJSON()
	if err != nil {
		e.displayError(w, err, http.StatusInternalServerError)
		return
	}

	// create a download link for json file
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", f.Name()))
	w.Header().Set("Content-Type", "application/json")
	http.ServeContent(w, r, f.Name(), time.Now(), f)
}

// TODO: create links for each file and commit hash
func (e *env) handleData(w http.ResponseWriter, r *http.Request) {
	// parse values from url query
	url := r.URL.Query().Get("url")
	hash := r.URL.Query().Get("commit")
	dir := r.URL.Query().Get("dir")

	// get hash and file collections
	coll, err := crud.GetGitCollection(url, hash, dir)
	if err != nil {
		e.displayError(w, err, http.StatusInternalServerError)
		return
	}

	// save collection to cache
	c.GitCollection = coll

	// execute collection on template
	e.templateFiles.ExecuteTemplate(w, "dashboard.html", coll)
}

func (e *env) handleDashboard(w http.ResponseWriter, r *http.Request) {
	e.templateFiles.ExecuteTemplate(w, "dashboard.html", nil)
}
