package testutils

import "os"

// UnsetEnvVar unsets the given environment variable with the key (if present).
// It returns a function to be called whenever you want to restore the original
// environment.
//
// In a test you can use this to temporarily set an environment variable:
//
//    func TestFoo(t *testing.T) {
//        resetFunc := testutils.UnsetEnvVar("foo")
//        defer resetFunc()
//        os.Setenv(key, "bar")
//
//        // continue as if foo=bar
//    }
func UnsetEnvVar(key string) func() {
	realEnvValue, present := os.LookupEnv(key)
	os.Unsetenv(key)
	return func() {
		if present {
			os.Setenv(key, realEnvValue)
		} else {
			os.Unsetenv(key)
		}
	}
}
