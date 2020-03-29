package sub

import (
	"bytes"
	"errors"
	"net/http"
	"runtime/debug"

	log "github.com/sirupsen/logrus"
)

// displayError is a function that uses usual respond for errors that happen on server side.
func (e *env) displayError(w http.ResponseWriter, err error, status int) {
	w.WriteHeader(status)

	log.Error(err)

	e.render(w, "error.page.tmpl", http.StatusText(http.StatusInternalServerError))
}

// catchPanic is an adapter for catching panic.
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

// render renders a template from a cache.
func (e *env) render(w http.ResponseWriter, name string, td interface{}) {
	// Retrieve the appropriate template set from the cache based on the page n
	// (like 'home.page.tmpl'). If no entry exists in the cache with the
	// provided name, call the serverError helper method that we made earlier.
	ts, ok := e.templateCache[name]
	if !ok {
		w.Write([]byte("Internal Server Error"))
		log.Printf("The template %s does not exist", name)
		return
	}

	// Initialize a new buffer.
	buf := new(bytes.Buffer)

	// Execute the template set, passing in any dynamic data.
	err := ts.Execute(buf, td)
	if err != nil {
		w.Write([]byte("Internal Server Error"))
		log.Error(err)
		return
	}

	// Write the contents of the buffer to the http.ResponseWriter. Again, this
	// is another time where we pass our http.ResponseWriter to a function that
	// takes an io.Writer.
	buf.WriteTo(w)
}
