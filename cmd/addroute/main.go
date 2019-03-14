package main

import (
	"flag"
	"html/template"
	"io/ioutil"
	"log"
	"os"
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
	p := templateParams{}
	flag.StringVar(&p.HandlerName, "name", "", "name of the handler")
	flag.StringVar(&p.Path, "path", "", "optional: /path/to/the/endpoint (defaults to: /handler_name)")
	flag.BoolVar(&p.OptsOverride, "override", false, "optional: forces override of generated files (defaults to: false)")
	flag.BoolVar(&p.OptsCleanupOnFailure, "cleanupOnFailure", false, "optional: deletes generated files if an error happened (defaults to: false)")
	flag.Parse()

	if err := p.sanitize(); err != nil {
		log.Printf("failed to sanitize template parameters: %+v\n", err)
		os.Exit(1)
	}

	log.Println(p.String())

	// create output files for handler and it's test

	handlerFile, err := ioutil.TempFile(".", "")
	if err != nil {
		log.Printf("failed to create temporary handler file: %+v\n", err)
		os.Exit(1)
	}
	log.Printf("created temporary handler file: %q", handlerFile.Name())

	handlerTestFile, err := ioutil.TempFile(".", "")
	if err != nil {
		log.Printf("failed to create temporary handler test file: %+v\n", err)
		os.Exit(1)
	}
	log.Printf("created temporary handler test file: %q\n", handlerTestFile.Name())

	// generate handler and it's test code

	if err = generate(handlerFile, p, p.handlerTemplateFileName()); err != nil {
		log.Printf("failed to create handler code: %+v\n", err)
		os.Exit(1)
	}

	if err = generate(handlerTestFile, p, p.handlerTestTemplateFileName()); err != nil {
		log.Printf("failed to create handler test code: %+v\n", err)
		os.Exit(1)
	}

	// Move temporary files to final destination

	if !p.OptsOverride {
		if fileExists(p.handlerFileName()) {
			log.Printf("cannot move %q to %q because the target file already exists and you explicitly requested not to override\n", handlerFile.Name(), p.handlerFileName())
			os.Exit(1)
		}
	}
	err = os.Rename(handlerFile.Name(), p.handlerFileName())
	if err != nil {
		log.Printf("failed to move %q to %q: %+v\n", handlerFile.Name(), p.handlerFileName(), err)
	}

	if !p.OptsOverride {
		if fileExists(p.handlerTestFileName()) {
			log.Printf("cannot move %q to %q because the target file already exists and you explicitly requested not to override\n", handlerTestFile.Name(), p.handlerTestFileName())
			os.Exit(1)
		}
	}
	os.Rename(handlerTestFile.Name(), p.handlerTestFileName())
	if err != nil {
		log.Printf("failed to move %q to %q: %+v\n", handlerTestFile.Name(), p.handlerTestFileName(), err)
	}
}

// fileExists checks if a file exists and is not a directory before we
// try using it to prevent further errors.
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func generate(outFile *os.File, p templateParams, templateFile string) error {
	tmpl, err := template.New(templateFile).Funcs(funcMap).ParseFiles(templateFile)
	if err != nil {
		return errs.Wrapf(err, "failed to parse template file %q", templateFile)
	}
	err = tmpl.Execute(outFile, p)
	return errs.Wrap(err, "failed to execute template")
}
