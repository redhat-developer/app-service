package appserver

// SetupRoutes registers handlers for various URL paths. You can call this
// function more than once but only the first call will have an effect.
func (srv *AppServer) SetupRoutes() error {
	var err error
	srv.routesSetup.Do(func() {

		// /status is something you should always have in any of your services,
		// please leave it as is.
		srv.router.HandleFunc("/status", srv.HandleStatus()).Queries("format", "{format:(?:json|yaml)}").
			Name("status").
			Methods("GET")

		// ADD YOUR OWN ROUTES HERE
		srv.router.HandleFunc("/topology", srv.HandleTopology()).
			Name("topology").
			Methods("GET")
	})
	return err
}
