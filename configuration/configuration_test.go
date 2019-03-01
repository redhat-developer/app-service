package configuration_test

import (
	"io/ioutil"
	"os"
	"strconv"
	"testing"

	"github.com/redhat-developer/boilerplate-app/configuration"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// unsetEnvVar unsets the given environment variable with the key (if present).
// It returns a function to be called whenever you want to restore the original
// environment.
func unsetEnvVar(key string) func() {
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

// getDefaultConfiguration returns a configuration registry without anything but
// defaults set. Remember that environment variables can overwrite defaults, so
// please ensure to properly unset envionment variables using unsetEnvVar().
func getDefaultConfiguration(t *testing.T) *configuration.Registry {
	config, err := configuration.New("")
	require.NoError(t, err)
	return config
}

// getFileConfiguration returns a configuration based on defaults, the given
// file content and overwrites by environment variables. As with
// getDefaultConfiguration() remember that environment variables can overwrite
// defaults, so please ensure to properly unset envionment variables using
// unsetEnvVar().
func getFileConfiguration(t *testing.T, content string) *configuration.Registry {
	tmpFile, err := ioutil.TempFile(os.TempDir(), "configFile-")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	_, err = tmpFile.Write([]byte(content))
	require.NoError(t, err)
	require.NoError(t, tmpFile.Close())
	config, err := configuration.New(tmpFile.Name())
	require.NoError(t, err)
	return config
}

func TestGetHTTPAddress(t *testing.T) {
	key := configuration.EnvPrefix + "_" + "HTTP_ADDRESS"

	t.Run("default", func(t *testing.T) {
		resetFunc := unsetEnvVar(key)
		defer resetFunc()
		config := getDefaultConfiguration(t)
		assert.Equal(t, configuration.DefaultHTTPAddress, config.GetHTTPAddress())
	})

	t.Run("file", func(t *testing.T) {
		resetFunc := unsetEnvVar(key)
		defer resetFunc()
		newVal := uuid.NewV4().String()
		config := getFileConfiguration(t, `http.address: "`+newVal+`"`)
		assert.Equal(t, newVal, config.GetHTTPAddress())
	})

	t.Run("env overwrite", func(t *testing.T) {
		newVal := uuid.NewV4().String()
		os.Setenv(key, newVal)
		config := getDefaultConfiguration(t)
		assert.Equal(t, newVal, config.GetHTTPAddress())
	})
}

func TestGetLogLevel(t *testing.T) {
	key := configuration.EnvPrefix + "_" + "LOG_LEVEL"
	resetFunc := unsetEnvVar(key)
	defer resetFunc()

	t.Run("default", func(t *testing.T) {
		resetFunc := unsetEnvVar(key)
		defer resetFunc()
		config := getDefaultConfiguration(t)
		assert.Equal(t, configuration.DefaultLogLevel, config.GetLogLevel())
	})

	t.Run("file", func(t *testing.T) {
		resetFunc := unsetEnvVar(key)
		defer resetFunc()
		newVal := uuid.NewV4().String()
		config := getFileConfiguration(t, `log.level: "`+newVal+`"`)
		assert.Equal(t, newVal, config.GetLogLevel())
	})

	t.Run("env overwrite", func(t *testing.T) {
		newVal := uuid.NewV4().String()
		os.Setenv(key, newVal)
		config := getDefaultConfiguration(t)
		assert.Equal(t, newVal, config.GetLogLevel())
	})
}

func TestIsLogJSON(t *testing.T) {
	key := configuration.EnvPrefix + "_" + "LOG_JSON"
	resetFunc := unsetEnvVar(key)
	defer resetFunc()

	t.Run("default", func(t *testing.T) {
		resetFunc := unsetEnvVar(key)
		defer resetFunc()
		config := getDefaultConfiguration(t)
		assert.Equal(t, configuration.DefaultLogJSON, config.IsLogJSON())
	})

	t.Run("file", func(t *testing.T) {
		resetFunc := unsetEnvVar(key)
		defer resetFunc()
		newVal := !configuration.DefaultLogJSON
		config := getFileConfiguration(t, `log.json: "`+strconv.FormatBool(newVal)+`"`)
		assert.Equal(t, newVal, config.IsLogJSON())
	})

	t.Run("env overwrite", func(t *testing.T) {
		newVal := !configuration.DefaultLogJSON
		os.Setenv(key, strconv.FormatBool(newVal))
		config := getDefaultConfiguration(t)
		assert.Equal(t, newVal, config.IsLogJSON())
	})
}
