package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_templateParams_String(t *testing.T) {
	require.Equal(t, "Handler name: \"\"\nPath: \"\"\nOptsOverride: false\nOptsCleanupOnFailure: false\n", templateParams{}.String())
	require.Equal(t, "Handler name: \"foo\"\nPath: \"/foo\"\nOptsOverride: true\nOptsCleanupOnFailure: true\n", templateParams{HandlerName: "foo", Path: "/foo", OptsOverride: true, OptsCleanupOnFailure: true}.String())
}

func Test_templateParams_handlerFileName(t *testing.T) {
	require.Equal(t, "../../appserver/handle_.go", templateParams{}.handlerFileName())
	require.Equal(t, "../../appserver/handle_foo_bar.go", templateParams{HandlerName: "foo bar", Path: "/foo/bar", OptsOverride: true, OptsCleanupOnFailure: true}.handlerFileName())
}

func Test_templateParams_handlerTestFileName(t *testing.T) {
	require.Equal(t, "../../appserver/handle__test.go", templateParams{}.handlerTestFileName())
	require.Equal(t, "../../appserver/handle_foo_bar_test.go", templateParams{HandlerName: "foo bar", Path: "/foo/bar", OptsOverride: true, OptsCleanupOnFailure: true}.handlerTestFileName())
}

func Test_templateParams_sanitize(t *testing.T) {
	type testcase struct {
		name    string
		before  templateParams
		after   templateParams
		wantErr bool
	}
	tt := []testcase{
		{
			name:    "empty handler name",
			before:  templateParams{},
			wantErr: true,
		},
		{
			name:   "default path",
			before: templateParams{HandlerName: "FooBar"},
			after:  templateParams{HandlerName: "FooBar", Path: "/foo_bar"},
		},
		{
			name:   "given path",
			before: templateParams{HandlerName: "foo bar", Path: "/foo bar /"},
			after:  templateParams{HandlerName: "FooBar", Path: "/foo_bar"},
		},
		{
			name:   "given path with param",
			before: templateParams{HandlerName: "foo bar", Path: "/foo/{bar}/"},
			after:  templateParams{HandlerName: "FooBar", Path: "/foo/{bar}"},
		},
		{
			name:   "given path",
			before: templateParams{HandlerName: "myFooBarTestHandler"},
			after:  templateParams{HandlerName: "MyFooBarTestHandler", Path: "/my_foo_bar_test_handler"},
		},
	}
	for _, tc := range tt {
		// capture range variable to allow for parallel testing
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := tc.before.sanitize()
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.after, tc.before)
			}
		})
	}
}
