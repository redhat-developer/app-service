package appserver_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/ghodss/yaml"
	"github.com/redhat-developer/boilerplate-app/appserver"
	"github.com/redhat-developer/boilerplate-app/configuration"
	"github.com/redhat-developer/boilerplate-app/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAppServer_HandleStatus(t *testing.T) {
	// enable gzip compression
	key := configuration.EnvPrefix + "_" + "HTTP_COMPRESS"
	resetFunc := testutils.UnsetEnvVar(key)
	defer resetFunc()
	os.Setenv(key, strconv.FormatBool(true))

	// where we store golden files for this test
	testDir := filepath.Join("golden-files", "status")

	// Setup the server
	srv, err := appserver.New("")
	require.NoError(t, err)
	err = srv.SetupRoutes()
	require.NoError(t, err)

	routesToPrint, err := srv.GetRegisteredRoutes()
	require.NoError(t, err)
	fmt.Println(routesToPrint)

	// Prepare expected response in JSON and YAML format
	expectedResponse := struct {
		Commit    string `json:"commit"`
		BuildTime string `json:"build_time"`
		StartTime string `json:"start_time"`
	}{
		BuildTime: appserver.BuildTime,
		StartTime: appserver.StartTime,
		Commit:    appserver.Commit,
	}
	expectedJSONBytes, err := json.Marshal(&expectedResponse)
	require.NoError(t, err)
	expectedYAMLBytes, err := yaml.Marshal(&expectedResponse)
	require.NoError(t, err)

	tt := []struct {
		desc                 string
		path                 string
		method               string
		expectedResponseCode int
		expectedHeaders      map[string]string
		matchBody            bool
		expectedResponseBody string
		goldenFilePath       string
		goldenFileOptions    testutils.CompareOptions
	}{
		{
			desc:                 "200 GET /status/json",
			path:                 "/status/json",
			method:               http.MethodGet,
			expectedResponseCode: http.StatusOK,
			expectedHeaders:      map[string]string{"Content-Type": "application/json"},
			matchBody:            true,
			expectedResponseBody: string(expectedJSONBytes),
			goldenFilePath:       filepath.Join(testDir, "ok.json"),
			goldenFileOptions:    testutils.CompareOptions{UUIDAgnostic: true, DateTimeAgnostic: true},
		},
		{
			desc:                 "200 GET /status/yaml",
			path:                 "/status/yaml",
			method:               http.MethodGet,
			expectedResponseCode: http.StatusOK,
			expectedHeaders:      map[string]string{"Content-Type": "application/yaml"},
			matchBody:            true,
			expectedResponseBody: string(expectedYAMLBytes),
			goldenFilePath:       filepath.Join(testDir, "ok.yaml"),
			goldenFileOptions:    testutils.CompareOptions{UUIDAgnostic: true, DateTimeAgnostic: true},
		},
		{
			desc:                 "405 POST /status/json",
			path:                 "/status/json",
			method:               http.MethodPost,
			expectedResponseCode: http.StatusMethodNotAllowed,
		},
		{
			desc:                 "405 POST /status/yaml",
			path:                 "/status/json",
			method:               http.MethodPost,
			expectedResponseCode: http.StatusMethodNotAllowed,
		},
		{
			desc:                 "404 Get /status/foobar",
			path:                 "/status/foobar",
			method:               http.MethodGet,
			expectedResponseCode: http.StatusNotFound,
		},
		{
			desc:                 "404 Get /status/",
			path:                 "/status/",
			method:               http.MethodGet,
			expectedResponseCode: http.StatusNotFound,
		},
	}

	for _, tc := range tt {
		// capture range variable for parllel testing
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()
			// Create a request to pass to our handler. We don't have any query
			// parameters for now, so we'll pass 'nil' as the third parameter.
			req, err := http.NewRequest(tc.method, tc.path, nil)
			require.NoError(t, err)

			// We create a ResponseRecorder (which satisfies http.ResponseWriter) to
			// record the response.
			rr := httptest.NewRecorder()
			// handler := http.HandlerFunc(srv.HandleStatus())

			// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
			// directly and pass in our Request and ResponseRecorder.
			srv.Router().ServeHTTP(rr, req)

			// Check the status code is what we expect.
			assert.Equal(t, tc.expectedResponseCode, rr.Code, "unexpected status code: want %d (%s) but got %d (%s)",
				tc.expectedResponseCode, http.StatusText(tc.expectedResponseCode), rr.Code, http.StatusText(rr.Code))

			// Check the content-type code is what we expect.
			for k, v := range tc.expectedHeaders {
				assert.Equal(t, v, rr.Header().Get(http.CanonicalHeaderKey(k)))
			}

			// Check the response body is what we expect.
			if tc.matchBody {
				require.Equal(t, tc.expectedResponseBody, rr.Body.String(), "handler returned unexpected body: got %v want %v")
				if tc.goldenFilePath != "" {
					testutils.CompareWithGolden(t, tc.goldenFilePath, rr.Body.String(), testutils.CompareOptions{UUIDAgnostic: true, DateTimeAgnostic: true})
				}
			}
		})
	}
}
