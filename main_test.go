package main

import (
	"net/http"
	"os"
	"syscall"
	"testing"

	"github.com/redhat-developer/app-service/testutils"
	"github.com/stretchr/testify/require"
)

func TestMainFunc(t *testing.T) {
	// TODO(kwk): This test should be improved but for now it gives a short
	// overview of how one can call main with arguments.

	os.Args = []string{
		"noop",
	}

	// If we need to set a config variable
	restoreFunc := testutils.UnsetEnvVarAndRestore("foo")
	defer restoreFunc()

	go main()

	// GET /status?format=yaml
	for i := 0; i < 100; i++ {
		resp, err := http.Get("http://0.0.0.0:8080/status?format=yaml")
		if err == nil && resp != nil && resp.Body != nil {
			defer resp.Body.Close()
			break
		}
	}

	// send SIGTERM to current process to trigger server shutdown
	pid := os.Getpid()
	err := syscall.Kill(pid, syscall.SIGTERM)
	require.NoError(t, err)
}
