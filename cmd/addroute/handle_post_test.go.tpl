// NOTE: This file was generated and should be modified by you!

package appserver_test

import (
	"net/http"
	"path/filepath"
	"testing"

	"github.com/kwk/boilerplate-app/testutils"
)

func TestAppServer_{{.HandlerName}}(t *testing.T) {
	testDir := filepath.Join("golden-files", "{{toSnake .HandlerName}}")

	tt := []testcase{
		{
			desc: "200 POST /{{.Path}}/english",
			req: request{
				path:   "/{{.Path}}/english",
				method: http.MethodPost,
				body: body{
					text: `{"name": "John"}`,
					goldenFile: goldenFile{
						path: filepath.Join(testDir, "english", "200_ok.req.body.json"),
						opts: testutils.CompareOptions{UUIDAgnostic: true, DateTimeAgnostic: true},
					},
				},
			},
			res: expectedResponse{
				compareBody: true,
				code:        http.StatusOK,
				body: body{
					text: `{"greeting":"Hello John!"}`,
					goldenFile: goldenFile{
						path: filepath.Join(testDir, "english", "200_ok.res.body.json"),
					},
				},
				headers: headers{
					kv: map[string]string{
						"Content-Type": "application/json",
					},
					goldenFile: goldenFile{
						path: filepath.Join(testDir, "english", "200_ok.res.headers.json"),
						opts: testutils.CompareOptions{MarshalInputAsJSON: true},
					},
				},
			},
		},
		{
			desc: "200 POST /{{.Path}}/german",
			req: request{
				path:   "/{{.Path}}/german",
				method: http.MethodPost,
				body: body{
					text: `{"name": "Klaus"}`,
					goldenFile: goldenFile{
						path: filepath.Join(testDir, "german", "200_ok.req.body.json"),
						opts: testutils.CompareOptions{UUIDAgnostic: true, DateTimeAgnostic: true},
					},
				},
			},
			res: expectedResponse{
				compareBody: true,
				code:        http.StatusOK,
				body: body{
					text: `{"greeting":"Hallo Klaus!"}`,
					goldenFile: goldenFile{
						path: filepath.Join(testDir, "german", "200_ok.res.body.json"),
					},
				},
				headers: headers{
					kv: map[string]string{
						"Content-Type": "application/json",
					},
					goldenFile: goldenFile{
						path: filepath.Join(testDir, "german", "200_ok.res.headers.json"),
						opts: testutils.CompareOptions{MarshalInputAsJSON: true},
					},
				},
			},
		},
		{
			desc: "400 POST /{{.Path}}/english (empty name)",
			req: request{
				path:   "/{{.Path}}/english",
				method: http.MethodPost,
				body: body{
					text: `{"name": ""}`,
					goldenFile: goldenFile{
						path: filepath.Join(testDir, "english", "400_empty_name.req.body.json"),
						opts: testutils.CompareOptions{UUIDAgnostic: true, DateTimeAgnostic: true},
					},
				},
			},
			res: expectedResponse{
				code: http.StatusBadRequest,
				body: body{
					goldenFile: goldenFile{
						path: filepath.Join(testDir, "english", "400_empty_name.res.body.txt"),
					},
				},
				headers: headers{
					goldenFile: goldenFile{
						path: filepath.Join(testDir, "english", "400_empty_name.res.headers.json"),
						opts: testutils.CompareOptions{MarshalInputAsJSON: true},
					},
				},
			},
		},
		{
			desc: "400 POST /{{.Path}}/english (missing closing brace)",
			req: request{
				path:   "/{{.Path}}/english",
				method: http.MethodPost,
				body: body{
					text: `{"name": "missing closing brace"`,
					goldenFile: goldenFile{
						path: filepath.Join(testDir, "english", "400_missing_closing_brace.req.body.json"),
						opts: testutils.CompareOptions{UUIDAgnostic: true, DateTimeAgnostic: true},
					},
				},
			},
			res: expectedResponse{
				code: http.StatusBadRequest,
				body: body{
					goldenFile: goldenFile{
						path: filepath.Join(testDir, "english", "400_missing_closing_brace.res.body.txt"),
					},
				},
				headers: headers{
					goldenFile: goldenFile{
						path: filepath.Join(testDir, "english", "400_missing_closing_brace.res.headers.json"),
						opts: testutils.CompareOptions{MarshalInputAsJSON: true},
					},
				},
			},
		},
	}

	runHandlerTests(t, tt)
}
