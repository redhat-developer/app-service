// Package appserver provides a Server structure that let's you package things you
// need all around your service
package appserver

import (
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	errs "github.com/pkg/errors"
	"github.com/redhat-developer/boilerplate-app/configuration"
)

var (
	// Commit current build commit set by build script
	Commit = "0"
	// BuildTime set by build script in ISO 8601 (UTC) format: YYYY-MM-DDThh:mm:ssTZD (see https://www.w3.org/TR/NOTE-datetime for details)
	BuildTime = "0"
	// StartTime in ISO 8601 (UTC) format
	StartTime = time.Now().UTC().Format("2006-01-02T15:04:05Z")
)

// AppServer bundles configuration, logging, and HTTP server objects in a single
// location.
type AppServer struct {
	config     *configuration.Registry
	router     *mux.Router
	httpServer *http.Server

	logger      *log.Logger
	routesSetup sync.Once
}

// New creates a new AppServer object with reasonable defaults.
func New(configFilePath string) (*AppServer, error) {
	srv := &AppServer{
		router: mux.NewRouter(),
		logger: log.New(os.Stderr, "", 0),
	}
	config, err := configuration.New(configFilePath)
	if err != nil {
		return nil, errs.Wrapf(err, "failed to create a new configuration registry from file %q", configFilePath)
	}
	srv.config = config
	srv.httpServer = &http.Server{
		Addr: srv.config.GetHTTPAddress(),
		// Good practice to set timeouts to avoid Slowloris attacks.
		WriteTimeout: srv.config.GetHTTPWriteTimeout(),
		ReadTimeout:  srv.config.GetHTTPReadTimeout(),
		IdleTimeout:  srv.config.GetHTTPIdleTimeout(),
		Handler:      handlers.LoggingHandler(os.Stdout, srv.router),
	}
	return srv, nil
}

// Logger returns the app server's log object
func (srv *AppServer) Logger() *log.Logger {
	return srv.logger
}

// Config returns the app server's config object
func (srv *AppServer) Config() *configuration.Registry {
	return srv.config
}

// HTTPServer returns the app server's HTTP server
func (srv *AppServer) HTTPServer() *http.Server {
	return srv.httpServer
}

// Router returns the app server's HTTP router
func (srv *AppServer) Router() *mux.Router {
	return srv.router
}

// GetRegisteredRoutes returns all registered routes formatted with their
// methods, paths, queries and names.
func (srv *AppServer) GetRegisteredRoutes() (string, error) {
	var sb strings.Builder
	err := srv.router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		sb.WriteString("ROUTE: ")
		pathTemplate, err := route.GetPathTemplate()
		if err == nil {
			sb.WriteString("\tPath template: ")
			sb.WriteString(pathTemplate)
		}
		name := route.GetName()
		if name != "" {
			sb.WriteString("\n\tName: ")
			sb.WriteString(name)
		}
		pathRegexp, err := route.GetPathRegexp()
		if err == nil {
			sb.WriteString("\n\tPath regexp: ")
			sb.WriteString(pathRegexp)
		}
		queriesTemplates, err := route.GetQueriesTemplates()
		if err == nil {
			sb.WriteString("\n\tQueries templates: ")
			sb.WriteString(strings.Join(queriesTemplates, ","))
		}
		queriesRegexps, err := route.GetQueriesRegexp()
		if err == nil {
			sb.WriteString("\n\tQueries regexps: ")
			sb.WriteString(strings.Join(queriesRegexps, ","))
		}
		methods, err := route.GetMethods()
		if err == nil {
			sb.WriteString("\n\tMethods: ")
			sb.WriteString(strings.Join(methods, ","))
		}
		return nil
	})
	return sb.String(), err
}
