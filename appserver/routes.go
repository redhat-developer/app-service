package appserver

import "github.com/gorilla/handlers"

// SetupRoutes registers handlers for various URL paths. You can call this
// function more than once but only the first call will have an effect.
func (srv *AppServer) SetupRoutes() error {
	var err error
	srv.routesSetup.Do(func() {

		// /status is something you should always have in any of your services,
		// please leave it as is.
		srv.router.HandleFunc("/status/{format:(?:json|yaml)}", srv.HandleStatus()).Name("status").Methods("GET")

		// TODO: BEGIN TO PUT YOUR HANDLERS BELOW !!!

		// TODO: PUT YOUR HANDLERS ABOVE !!!

		if srv.config.GetHTTPCompressResponses() {
			srv.router.Use(handlers.CompressHandler)
		}

		// Check
		// err := srv.router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		// // TODO(kwk): one service with different handlers for different methods
	})
	return err
}
