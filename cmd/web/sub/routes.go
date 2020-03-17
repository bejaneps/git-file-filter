package sub

func (e *env) routes() {
	e.router.HandleFunc("/", e.handleDashboard)
	e.router.HandleFunc("/data", e.handleData)
}
