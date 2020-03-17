package sub

func (e *env) routes() {
	e.router.HandleFunc("/", e.catchPanic(e.handleDashboard))
	e.router.HandleFunc("/data", e.catchPanic(e.handleData))
	e.router.HandleFunc("/download", e.catchPanic(e.handleDownload))
	e.router.HandleFunc("/regexp", e.catchPanic(e.handleRegexp))
}
