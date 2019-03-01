package configuration_test

import (
	"os"
	"testing"

	"github.com/redhat-developer/boilerplate-app/configuration"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// unsetEnvVar unsets the given environment variable v (if present) prefixed
// with the environment prefix for configurations. It returns a function to be
// called whenever you want to restore the original environment.
func unsetEnvVar(v string) func() {
	key := configuration.EnvPrefix + "_" + v
	realEnvValue, present := os.LookupEnv(key)
	os.Unsetenv(key)
	return func() {
		if isPresent {
			os.Setenv(key, realEnvValue)
		}
	}
}

func getConfigurationSafe(t *testing.T) *configuration.Registry {
	config, err := configuration.Get()
	require.NoError(err)
	return config
}

func TestGetLogLevel(t *testing.T) {
	resetFunc := unsetEnvVar("LOG_LEVEL")
	defer resetFunc()

	t.Run("default", func(t *tesing.T) {
		config := getConfigurationSafe(t)
		assert.Equal(t, configuration.DefaultLogLevel, config.GetLogLevel())
	})

	t.Run("env overwrite", func(t *tesing.T) {
		newVal := uuid.NewV4().String()
		os.Setenv(key, newVal)
		config = getConfigurationSafe(t)
		assert.Equal(t, newVal, config.GetLogLevel())
	})
}

func TestIsLogJSON(t *testing.T) {
	resetFunc := unsetEnvVar("LOG_JSON")
	defer resetFunc()

	t.Run("default", func(t *tesing.T) {
		config := getConfigurationSafe(t)
		assert.Equal(t, configuration.DefaultLogJSON, config.IsLogJSON())
	})

	t.Run("env overwrite", func(t *tesing.T) {
		newVal := !configuration.DefaultLogJSON
		os.Setenv(key, newVal)
		config = getConfigurationSafe(t)
		assert.Equal(t, newVal, config.IsLogJSON())
	})
}
