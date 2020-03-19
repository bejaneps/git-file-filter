package sub

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/bejaneps/go-git-webapp/internal/crud"

	jsoniter "github.com/json-iterator/go"
)

var (
	json = jsoniter.ConfigCompatibleWithStandardLibrary
)

func (e *env) handleRegexpGET(w http.ResponseWriter, r *http.Request) {
	pattern := r.FormValue("pattern")

	var err error

	// convert all backslashes to forward slashes in string
	pattern = strings.ReplaceAll(pattern, "\\", "\\\\")

	// decode query value to json
	rt := &crud.RegexpConfig{}
	err = json.NewDecoder(strings.NewReader(pattern)).Decode(&rt.X)
	if err != nil {
		e.displayError(w, err, http.StatusInternalServerError)
		return
	}

	// filter files by regexp
	coll, err := e.gitCollectionFiles.Filter(rt)
	if err != nil {
		e.displayError(w, err, http.StatusInternalServerError)
		return
	}

	e.gitCollectionConfigs = coll // save to cache

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
	file, _, err := r.FormFile("pattern")
	if err != nil {
		e.displayError(w, err, http.StatusInternalServerError)
		return
	}
	defer file.Close()

	// read a file into a buffer, and then replace all \ to /
	buf := &strings.Builder{}
	b, err := io.Copy(buf, file)
	if err != nil {
		e.displayError(w, err, http.StatusInternalServerError)
		return
	} else if b == 0 {
		e.displayError(w, errors.New("no bytes copied from file to buffer"), http.StatusInternalServerError)
		return
	}
	pattern := strings.ReplaceAll(buf.String(), "\\", "\\\\")

	// decode file into json struct
	rt := &crud.RegexpConfig{}
	err = json.NewDecoder(strings.NewReader(pattern)).Decode(&rt.X)
	if err != nil {
		e.displayError(w, err, http.StatusInternalServerError)
		return
	}

	// filter files by regexp
	coll, err := e.gitCollectionFiles.Filter(rt)
	if err != nil {
		e.displayError(w, err, http.StatusInternalServerError)
		return
	}

	e.gitCollectionConfigs = coll // save to cache

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

func (e *env) handleSearch(w http.ResponseWriter, r *http.Request) {
	e.render(w, "create.page.tmpl", nil)
}

func (e *env) handleSearchQuery(w http.ResponseWriter, r *http.Request) {
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
	e.gitCollectionFiles = coll

	http.Redirect(w, r, "/", http.StatusFound)
}

func (e *env) handleFiles(w http.ResponseWriter, r *http.Request) {
	if e.gitCollectionFiles == nil {
		e.render(w, "home.page.tmpl", nil)
	} else {
		e.render(w, "home.page.tmpl", e.gitCollectionFiles)
	}
}

func (e *env) handleFilter(w http.ResponseWriter, r *http.Request) {
	e.render(w, "filter.page.tmpl", nil)
}

func (e *env) handleConfigs(w http.ResponseWriter, r *http.Request) {
	if e.gitCollectionConfigs == nil {
		e.render(w, "configs.page.tmpl", nil)
	} else {
		e.render(w, "configs.page.tmpl", e.gitCollectionConfigs)
	}
}
