package appserver

import (
	"path/filepath"
	"testing"

	"github.com/redhat-developer/app-service/testutils"

	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	t.Parallel()
	t.Run("default configuration", func(t *testing.T) {
		t.Parallel()
		srv, err := New("")
		require.NoError(t, err)
		require.NotNil(t, srv)
	})
	t.Run("non existing file path", func(t *testing.T) {
		t.Parallel()
		srv, err := New(uuid.NewV4().String())
		require.Error(t, err)
		require.Nil(t, srv)
	})
}

func TestAppServer_Logger(t *testing.T) {
	t.Parallel()
	srv, err := New("")
	require.NoError(t, err)
	require.Equal(t, srv.logger, srv.Logger())
}

func TestAppServer_Router(t *testing.T) {
	t.Parallel()
	srv, err := New("")
	require.NoError(t, err)
	require.Equal(t, srv.router, srv.Router())
}

func TestAppServer_Config(t *testing.T) {
	t.Parallel()
	srv, err := New("")
	require.NoError(t, err)
	require.Equal(t, srv.config, srv.Config())
}

func TestAppServer_HTTPServer(t *testing.T) {
	t.Parallel()
	srv, err := New("")
	require.NoError(t, err)
	require.Equal(t, srv.httpServer, srv.HTTPServer())
}

func TestAppServer_GetRegisteredRoutes(t *testing.T) {
	t.Parallel()
	t.Run("no routes registered", func(t *testing.T) {
		t.Parallel()
		srv, err := New("")
		require.NoError(t, err)
		routes, err := srv.GetRegisteredRoutes()
		require.NoError(t, err)
		require.Equal(t, "", routes)
	})
	t.Run("no routes registered", func(t *testing.T) {
		t.Parallel()
		srv, err := New("")
		require.NoError(t, err)
		err = srv.SetupRoutes()
		require.NoError(t, err)
		// where we store golden files for this test
		testFile := filepath.Join("golden-files", "appserver", "get_registered_routes.json")
		routes, err := srv.GetRegisteredRoutes()
		require.NoError(t, err)
		testutils.CompareWithGolden(t, testFile, routes, testutils.CompareOptions{})
	})
}
