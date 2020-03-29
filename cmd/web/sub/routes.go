package sub

import (
	"net/http"
)

// routes injects dependency into http route handlers and also assigns handlers to each route.
func (e *env) routes() {
	// route for static files
	e.router.PathPrefix("/static/").Handler(http.StripPrefix("/static", http.FileServer(http.Dir("./ui/static/"))))

	// route for index page
	e.router.HandleFunc("/", e.catchPanic(e.handleFiles))

	// route for list of config files page
	e.router.HandleFunc("/configs", e.catchPanic(e.handleConfigs))

	// route for search page
	e.router.HandleFunc("/search", e.catchPanic(e.handleSearch))

	// route for search query page
	e.router.HandleFunc("/search/", e.catchPanic(e.handleSearchQuery)).Queries("url", "", "commit", "", "dir", "")

	// routes for filter query page
	e.router.HandleFunc("/regexp", e.catchPanic(e.handleRegexpGET)).Methods("GET")
	e.router.HandleFunc("/regexp", e.catchPanic(e.handleRegexpPOST)).Methods("POST")

	// route for filter page
	e.router.HandleFunc("/filter", e.catchPanic(e.handleFilter))
}
