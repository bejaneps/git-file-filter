package sub

import (
	"errors"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/bejaneps/go-git-webapp/internal/crud"
)

var (
	errFormURLEmpty = errors.New("url form value is empty")
)

// TODO: display repo, commit hash and directory on top of table
func (e *env) handleData(w http.ResponseWriter, r *http.Request) {
	// parse values from url query
	url := r.URL.Query().Get("url")
	hash := r.URL.Query().Get("commit")
	dir := r.URL.Query().Get("dir")

	if url == "" {
		err := e.templateFiles.ExecuteTemplate(w, "error.html", errFormURLEmpty)
		if err != nil {
			log.Error(err)
		}
		return
	}

	// get hash and file collections
	coll, err := crud.GetGitCollection(url, hash, dir)
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
