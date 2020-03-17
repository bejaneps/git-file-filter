package sub

import (
	"errors"
	"net/http"
	"runtime/debug"

	log "github.com/sirupsen/logrus"
)

func (e *env) displayError(w http.ResponseWriter, err error, status int) {
	w.WriteHeader(status)

	log.Error(err)

	err2 := e.templateFiles.ExecuteTemplate(w, "error.html", http.StatusText(http.StatusInternalServerError))
	if err2 != nil {
		log.Error(err2)
	}
}

func (e *env) catchPanic(f http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				err := errors.New(rec.(error).Error() + "\n\n\n" + string(debug.Stack()))

				e.displayError(w, err, http.StatusInternalServerError)
			}
		}()

		f.ServeHTTP(w, r)
	})
}
