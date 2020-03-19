package sub

import (
	"net/http"
)

func (e *env) routes() {
	// route for static files
	e.router.PathPrefix("/static/").Handler(http.StripPrefix("/static", http.FileServer(http.Dir("./ui/static/"))))

	e.router.HandleFunc("/", e.catchPanic(e.handleFiles))
	e.router.HandleFunc("/configs", e.catchPanic(e.handleConfigs))
	e.router.HandleFunc("/search", e.catchPanic(e.handleSearch))

	e.router.HandleFunc("/search/", e.catchPanic(e.handleSearchQuery)).Queries("url", "", "commit", "", "dir", "")
	e.router.HandleFunc("/regexp", e.catchPanic(e.handleRegexpGET)).Methods("GET")
	e.router.HandleFunc("/regexp", e.catchPanic(e.handleRegexpPOST)).Methods("POST")

	e.router.HandleFunc("/filter", e.catchPanic(e.handleFilter))

}
