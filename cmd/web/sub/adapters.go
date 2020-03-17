package sub

import (
	"net/http"

	log "github.com/sirupsen/logrus"
)

func (e *env) displayError(w http.ResponseWriter, err error, status int) {
	w.WriteHeader(status)

	err2 := e.templateFiles.ExecuteTemplate(w, "error.html", err)
	if err2 != nil {
		log.Error(err2)
	}
}
