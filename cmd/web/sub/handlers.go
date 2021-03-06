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

// request is a struct that holds a json config from filter page in web app
type request struct {
	Config []crud.Config `json:"config"`
}

var (
	json = jsoniter.ConfigCompatibleWithStandardLibrary
)

// handleRegexpGET handles upcoming requests from webapp filter page,
// when posting a json form not file.
func (e *env) handleRegexpGET(w http.ResponseWriter, r *http.Request) {
	// check if user searched a repository or no
	if e.gitCollectionFiles == nil {
		http.Redirect(w, r, "/search", http.StatusTemporaryRedirect)
		return
	}

	pattern := r.FormValue("pattern") // get the json from request

	var err error

	// convert all backslashes to forward slashes in string
	pattern = strings.ReplaceAll(pattern, "\\", "\\\\")

	// decode query value to json
	conf := &request{}
	err = json.NewDecoder(strings.NewReader(pattern)).Decode(&conf)
	if err != nil {
		e.displayError(w, err, http.StatusInternalServerError)
		return
	}

	// filter files by regexp
	coll, err := e.gitCollectionFiles.Filter(conf.Config)
	if err != nil {
		e.displayError(w, err, http.StatusInternalServerError)
		return
	}

	e.gitCollectionConfigs = coll // save to cache

	// write in json file also the file count in repo and count of programming langs used
	coll.FileCount = e.gitCollectionFiles.FileCount
	coll.Language = e.gitCollectionFiles.Language

	// create a json file
	f, err := coll.ToJSONFile()
	if err != nil {
		e.displayError(w, err, http.StatusInternalServerError)
		return
	}

	// create a download link for json file
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", f.Name()))
	w.Header().Set("Content-Type", "application/json")
	http.ServeContent(w, r, f.Name(), time.Now(), f)
}

// handleRegexpPOST handles upcoming requests from webapp filter page,
// when posting a json file, not form.
func (e *env) handleRegexpPOST(w http.ResponseWriter, r *http.Request) {
	// check if user searched a repository or no
	if e.gitCollectionFiles == nil {
		http.Redirect(w, r, "/search", http.StatusTemporaryRedirect)
		return
	}

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
	conf := &request{}
	err = json.NewDecoder(strings.NewReader(pattern)).Decode(&conf)
	if err != nil {
		e.displayError(w, err, http.StatusInternalServerError)
		return
	}

	// filter files by regexp
	coll, err := e.gitCollectionFiles.Filter(conf.Config)
	if err != nil {
		e.displayError(w, err, http.StatusInternalServerError)
		return
	}

	e.gitCollectionConfigs = coll // save to cache

	// write in json file also the file count in repo and count of programming langs used
	coll.FileCount = e.gitCollectionFiles.FileCount
	coll.Language = e.gitCollectionFiles.Language

	// create a json file
	f, err := coll.ToJSONFile()
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
