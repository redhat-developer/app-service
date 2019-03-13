package appserver_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"

	"github.com/redhat-developer/boilerplate-app/appserver"
	"github.com/redhat-developer/boilerplate-app/configuration"
	"github.com/redhat-developer/boilerplate-app/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testcase struct {
	desc string
	req  request
	res  expectedResponse
}
type request struct {
	path    string
	method  string
	body    body
	headers headers
}
type expectedResponse struct {
	code        int
	compareBody bool
	body        body
	headers     headers
}
type body struct {
	text       string
	goldenFile goldenFile
}
type goldenFile struct {
	path string
	opts testutils.CompareOptions
}
type headers struct {
	kv         map[string]string
	goldenFile goldenFile
}

func runHandlerTests(t *testing.T, tt []testcase) {
	// enable gzip compression to also go along this path
	// TODO(kwk): This was only added to get more coverage.
	key := configuration.EnvPrefix + "_" + "HTTP_COMPRESS"
	restoreFunc := testutils.UnsetEnvVarAndRestore(key)
	defer restoreFunc()
	os.Setenv(key, strconv.FormatBool(true))

	// Setup the server
	srv, err := appserver.New("")
	require.NoError(t, err)
	err = srv.SetupRoutes()
	require.NoError(t, err)

	for _, tc := range tt {
		// capture range variable for parallel testing
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			// t.Parallel()

			// Check if we need to store a golden file for the request body and header
			if tc.req.body.goldenFile.path != "" {
				testutils.CompareWithGolden(t, tc.req.body.goldenFile.path, tc.req.body.text, tc.req.body.goldenFile.opts)
			}
			if tc.req.headers.goldenFile.path != "" {
				testutils.CompareWithGolden(t, tc.req.headers.goldenFile.path, tc.req.headers.kv, tc.req.headers.goldenFile.opts)
			}

			// Create a request to pass to our handler
			req, err := http.NewRequest(tc.req.method, tc.req.path, bytes.NewBufferString(tc.req.body.text))
			require.NoError(t, err)

			// Use request headers as defined in the test case
			if len(tc.req.headers.kv) > 0 {
				req.Header = http.Header{}
				for k, v := range tc.req.headers.kv {
					req.Header.Set(http.CanonicalHeaderKey(k), v)
				}
			}

			// We create a ResponseRecorder (which satisfies http.ResponseWriter) to
			// record the response.
			rr := httptest.NewRecorder()

			// Our handlers satisfy http.Handler, so we can call their ServeHTTP
			// method directly and pass in our Request and ResponseRecorder.
			srv.Router().ServeHTTP(rr, req)

			// Check the status code is what we expect.
			assert.Equal(t, tc.res.code, rr.Code, "unexpected status code: want %d (%s) but got %d (%s)",
				tc.res.code, http.StatusText(tc.res.code), rr.Code, http.StatusText(rr.Code))

			// Check if the expected headers have been set.
			for k, v := range tc.res.headers.kv {
				assert.Equal(t, v, rr.Header().Get(http.CanonicalHeaderKey(k)))
			}

			// Check the response body is what we expect.
			if tc.res.compareBody {
				require.Equal(t, tc.res.body.text, rr.Body.String())
			}
			if tc.res.body.goldenFile.path != "" {
				testutils.CompareWithGolden(t, tc.res.body.goldenFile.path, rr.Body.String(), tc.res.body.goldenFile.opts)
			}
			if tc.res.headers.goldenFile.path != "" {
				testutils.CompareWithGolden(t, tc.res.headers.goldenFile.path, rr.Header(), tc.res.headers.goldenFile.opts)
			}
		})
	}
}
