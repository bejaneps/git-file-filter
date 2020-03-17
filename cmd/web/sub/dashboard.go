package sub

import (
	"errors"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/bejaneps/go-git-webapp/internal/crud"
)

var (
	errFormValuesEmpty = errors.New("form value is empty")
)

func (e *env) handleData(w http.ResponseWriter, r *http.Request) {
	// parse values from url query
	url := r.URL.Query().Get("url")
	hash := r.URL.Query().Get("commit")
	dir := r.URL.Query().Get("dir")

	if url == "" || hash == "" {
		err := e.templateFiles.ExecuteTemplate(w, "error.html", errFormValuesEmpty)
		if err != nil {
			log.Error(err)
		}
		return
	}

	// get hash and file collections
	coll, err := crud.GetGitCollections(url, hash, dir)
	if err != nil {
		err := e.templateFiles.ExecuteTemplate(w, "error.html", err)
		if err != nil {
			log.Error(err)
		}
		return
	}

	// execute collections on template
	e.templateFiles.ExecuteTemplate(w, "dashboard.html", coll)
}

func (e *env) handleDashboard(w http.ResponseWriter, r *http.Request) {
	e.templateFiles.ExecuteTemplate(w, "dashboard.html", nil)
}
