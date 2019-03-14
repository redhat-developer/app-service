package main

import (
	"testing"
)

func TestMainFunc(t *testing.T) {
	// p := templateParams{
	// 	HandlerName:          "hello / world dfoo",
	// 	OptsCleanupOnFailure: true,
	// }
	// _, err := os.Stat(p.handlerFileName())
	// require.FileExists(t, os.IsExist(err), "test file already exists: %q", p.handlerFileName())

	// // delete generated test file on completion of test
	// defer func(p templateParams) {
	// 	fmt.Printf("removing %q if it exists\n", p.handlerFileName())
	// 	_, err := os.Stat(p.handlerFileName())
	// 	if os.IsExist(err) {
	// 		os.Remove(p.handlerFileName())
	// 	}
	// }(p)
	// // io.Stat(handlerFileName(""))
	// os.Args = []string{"noop", fmt.Sprintf("-name=%q", p.HandlerName)}
	// main()
}
