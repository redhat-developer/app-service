package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/iancoleman/strcase"
	errs "github.com/pkg/errors"
)

type templateParams struct {
	HandlerName string
	Path        string

	// OptsOverride set to true to override already existing files when generating code
	OptsOverride         bool
	OptsCleanupOnFailure bool
}

func (p *templateParams) sanitize() error {
	// sanitize handler name
	p.HandlerName = strings.Replace(p.HandlerName, "/", "", -1)
	p.HandlerName = strings.Replace(p.HandlerName, "'", "", -1)
	p.HandlerName = strings.Replace(p.HandlerName, "\"", "", -1)
	p.HandlerName = strings.TrimSpace(p.HandlerName)
	p.HandlerName = strcase.ToCamel(p.HandlerName)
	if p.HandlerName == "" {
		return errs.New("handler name must not be empty")
	}

	// sanitize path
	if p.Path == "" {
		p.Path = strcase.ToLowerCamel(p.HandlerName)
	}
	p.Path = strings.TrimPrefix(p.Path, "/")
	p.Path = "/" + p.Path
	p.Path = strings.TrimSuffix(p.Path, "/")
	p.Path = strcase.ToSnake(p.Path)

	// validate if path can be used as in gorilla
	router := mux.NewRouter()
	return router.HandleFunc(p.Path, func(http.ResponseWriter, *http.Request) {}).GetError()
}

func (p templateParams) String() string {
	return fmt.Sprintf("Handler name: %q\n"+
		"Path: %q\n"+
		"OptsOverride: %s\n"+
		"OptsCleanupOnFailure: %s\n",
		p.HandlerName,
		p.Path,
		strconv.FormatBool(p.OptsOverride),
		strconv.FormatBool(p.OptsCleanupOnFailure))
}

func (p templateParams) handlerFileName() string {
	return "../../appserver/handle_" + strcase.ToSnake(p.HandlerName) + ".go"
}

func (p templateParams) handlerTestFileName() string {
	return "../../appserver/handle_" + strcase.ToSnake(p.HandlerName) + "_test.go"
}
