package main

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMainFunc(t *testing.T) {
	p := templateParams{
		HandlerName:          "my hello world",
		Path:                 "/my/hello/world",
		OptsCleanupOnFailure: true,
	}
	// _, err := os.Stat(p.handlerFileName())
	// require.FileExists(t, os.IsExist(err), "test file already exists: %q", p.handlerFileName())

	// delete generated test file on completion of test
	// defer func(p templateParams) {
	// 	fmt.Printf("removing %q if it exists\n", p.handlerFileName())
	// 	_, err := os.Stat(p.handlerFileName())
	// 	if os.IsExist(err) {
	// 		os.Remove(p.handlerFileName())
	// 	}
	// }(p)
	// io.Stat(handlerFileName(""))

	os.Args = []string{
		"noop",
		"-name=" + p.HandlerName,
		"-path=" + p.Path,
		"-override",
	}

	main()

	if fileExists("../../appserver/handle_my_hello_world.go") {
		err := os.Remove("../../appserver/handle_my_hello_world.go")
		require.NoError(t, err)
	}
	if fileExists("../../appserver/handle_my_hello_world_test.go") {
		err := os.Remove("../../appserver/handle_my_hello_world_test.go")
		require.NoError(t, err)
	}
}

func Test_generate(t *testing.T) {
	// create temp files to write out generated output to
	handlerFile, err := ioutil.TempFile(".", "my_foo_barhandler.go")
	require.NoError(t, err)
	defer os.Remove(handlerFile.Name())

	handlerTestFile, err := ioutil.TempFile(".", "my_foo_barhandler_test.go")
	require.NoError(t, err)
	defer os.Remove(handlerTestFile.Name())

	p := templateParams{HandlerName: "myFooBarTest\nHandler", Path: "/my/foo/bar/test/handler"}
	err = p.sanitize()
	require.NoError(t, err)

	t.Run("template:"+p.handlerTemplateFileName(), func(t *testing.T) {
		err := generate(handlerFile, p, p.handlerTemplateFileName())
		require.NoError(t, err)
	})
	t.Run("template:"+p.handlerTestTemplateFileName(), func(t *testing.T) {
		err := generate(handlerTestFile, p, p.handlerTestTemplateFileName())
		require.NoError(t, err)
	})
}
