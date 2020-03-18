package sub

func (e *env) routes() {
	e.router.HandleFunc("/", e.catchPanic(e.handleDashboard))
	e.router.HandleFunc("/data", e.catchPanic(e.handleData))
	e.router.HandleFunc("/regexp", e.catchPanic(e.handleRegexpGET)).Methods("GET")
	e.router.HandleFunc("/regexp", e.catchPanic(e.handleRegexpPOST)).Methods("POST")
}
