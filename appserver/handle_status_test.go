package appserver_test

import (
	"net/http"
	"path/filepath"
	"testing"

	"github.com/redhat-developer/boilerplate-app/testutils"
)

func TestAppServer_HandleStatus(t *testing.T) {
	testDir := filepath.Join("golden-files", "status")

	tt := []testcase{
		{
			desc: "200 GET /status?format=json",
			req: request{
				path:   "/status?format=json",
				method: http.MethodGet,
			},
			res: expectedResponse{
				code: http.StatusOK,
				body: body{
					goldenFile: goldenFile{
						path: filepath.Join(testDir, "200_ok.res.body.json"),
						opts: testutils.CompareOptions{UUIDAgnostic: true, DateTimeAgnostic: true},
					},
				},
				headers: headers{
					kv: map[string]string{
						"Content-Type": "application/json",
					},
					goldenFile: goldenFile{
						path: filepath.Join(testDir, "200_ok.res.headers.json"),
					},
				},
			},
		},
		{
			desc: "200 GET /status?format=yaml",
			req: request{
				path:   "/status?format=yaml",
				method: http.MethodGet,
			},
			res: expectedResponse{
				code: http.StatusOK,
				body: body{
					goldenFile: goldenFile{
						path: filepath.Join(testDir, "200_ok.res.body.yaml"),
						opts: testutils.CompareOptions{UUIDAgnostic: true, DateTimeAgnostic: true},
					},
				},
				headers: headers{
					kv: map[string]string{
						"Content-Type": "application/yaml",
					},
					goldenFile: goldenFile{
						path: filepath.Join(testDir, "200_ok.res.headers.yaml"),
					},
				},
			},
		},
		{
			desc: "404 GET /status?format=foobar",
			req: request{
				path:   "/status?format=foobar",
				method: http.MethodGet,
			},
			res: expectedResponse{
				code: http.StatusNotFound,
			},
		},
		{
			desc: "405 POST /status?format=json",
			req: request{
				path:   "/status?format=json",
				method: http.MethodPost,
			},
			res: expectedResponse{
				code: http.StatusMethodNotAllowed,
			},
		},
		{
			desc: "405 POST /status?format=yaml",
			req: request{
				path:   "/status?format=yaml",
				method: http.MethodPost,
			},
			res: expectedResponse{
				code: http.StatusMethodNotAllowed,
			},
		},
		{
			desc: "404 GET /status/",
			req: request{
				path:   "/status/",
				method: http.MethodGet,
			},
			res: expectedResponse{
				code: http.StatusNotFound,
			},
		},
		{
			desc: "404 GET /status",
			req: request{
				path:   "/status",
				method: http.MethodGet,
			},
			res: expectedResponse{
				code: http.StatusNotFound,
			},
		},
	}

	runHandlerTests(t, tt)
}
