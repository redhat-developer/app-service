package main

import (
	"flag"
	"html/template"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/iancoleman/strcase"
	errs "github.com/pkg/errors"
)

// First we create a FuncMap with which to register the function.
var funcMap = template.FuncMap{
	// The name "title" is what the function will be called in the template text.
	"title":        strings.Title,
	"toCamel":      strcase.ToCamel,
	"toLowerCamel": strcase.ToLowerCamel,
	"toSnake":      strcase.ToSnake,
}

func main() {
	// parse flags
	p := &templateParams{}
	flag.StringVar(&p.HandlerName, "name", "", "name of the handler")
	flag.StringVar(&p.Path, "basePath", "", "optional: /path/to/the/endpoint (defaults to: /handler_name)")
	flag.BoolVar(&p.OptsOverride, "override", false, "optional: forces override of generated files (defaults to: false)")
	flag.BoolVar(&p.OptsCleanupOnFailure, "cleanupOnFailure", false, "optional: deletes generated files if an error happened (defaults to: false)")
	flag.Parse()

	if err := p.sanitize(); err != nil {
		log.Printf("failed to sanitize template parameters: %+v\n", err)
		os.Exit(1)
	}

	log.Println(p.String())

	if err := writeHandlerCode(p); err != nil {
		log.Println("failed to write handler code: ", err)
		os.Exit(1)
	}
}

func writeHandlerCode(p *templateParams) error {
	fileName := p.handlerFileName()
	// prevent overwriting
	if !p.OptsOverride {
		if _, err := os.Stat(fileName); err != nil {
			if os.IsExist(err) {
				return errs.Wrapf(err, "file already exists: %s", fileName)
			}
		}
	}
	outFile, err := os.Create(fileName)
	if err != nil {
		return errs.Wrapf(err, "failed to create file %q", fileName)
	}
	defer outFile.Close()
	err = generate(outFile, p)
	if err != nil {
		if p.OptsCleanupOnFailure {
			os.Remove(fileName)
		}
		return errs.Wrapf(err, "failed to generate %q", fileName)
	}
	return nil
}

func generate(outFile *os.File, p *templateParams) error {
	tmplName := filepath.Base(outFile.Name())
	tmpl := template.Must(template.New(tmplName).Funcs(funcMap).ParseFiles(outFile.Name()))
	err := tmpl.Execute(outFile, p)
	return errs.Wrap(err, "failed to execute template")
}
