// Package configuration is in charge of the validation and extraction of all
// the configuration details from a configuration file or environment variables.
package configuration

import (
	"crypto/rsa"
	"os"
	"strings"

	errs "github.com/pkg/errors"
	"github.com/spf13/viper"
)

const (
	// EnvPrefix will be used for environment variable name prefixing.
	EnvPrefix = "RHDEV"

	// Constants for viper variable names. Will be used to set
	// default values as well as to get each value
	varHTTPAddress = "http.address"
	// DefaultHTTPAddress is the address and port string that your service will
	// be exported to by default.
	DefaultHTTPAddress = "0.0.0.0:8001"

	varLogLevel = "log.level"
	// DefaultLogLevel is the default log level used in your service.
	DefaultLogLevel = "info"

	varLogJSON = "log.json"
	// DefaultLogJSON is a switch to toggle on and off JSON log output.
	DefaultLogJSON = false
)

// Registry encapsulates the Viper configuration registry which stores the
// configuration data in-memory.
type Registry struct {
	v               *viper.Viper
	tokenPublicKey  *rsa.PublicKey
	tokenPrivateKey *rsa.PrivateKey
}

// New creates a configuration reader object using a configurable configuration
// file path.
func New(configFilePath string) (*Registry, error) {
	c := Registry{
		v: viper.New(),
	}
	c.v.SetEnvPrefix(EnvPrefix)
	c.v.AutomaticEnv()
	c.v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	c.v.SetTypeByDefaultValue(true)
	c.setConfigDefaults()

	if configFilePath != "" {
		c.v.SetConfigType("yaml")
		c.v.SetConfigFile(configFilePath)
		err := c.v.ReadInConfig() // Find and read the config file
		if err != nil {           // Handle errors reading the config file
			return nil, errs.Wrap(err, "failed to read config file")
		}
	}
	return &c, nil
}

func getConfigFilePath() string {
	// This was either passed as a env var Or, set inside main.go from --config
	envConfigPath, ok := os.LookupEnv(EnvPrefix + "_CONFIG_FILE_PATH")
	if !ok {
		return ""
	}
	return envConfigPath
}

// Get is a wrapper over New() which reads configuration file path from the
// environment variable.
func Get() (*Registry, error) {
	cd, err := New(getConfigFilePath())
	return cd, err
}

func (c *Registry) setConfigDefaults() {
	c.v.SetTypeByDefaultValue(true)

	c.v.SetDefault(varHTTPAddress, DefaultHTTPAddress)
	c.v.SetDefault(varLogLevel, DefaultLogLevel)
	c.v.SetDefault(varLogJSON, DefaultLogJSON)
}

// GetHTTPAddress returns the HTTP address (as set via default, config file, or
// environment variable) that the wit server binds to (e.g. "0.0.0.0:8080")
func (c *Registry) GetHTTPAddress() string {
	return c.v.GetString(varHTTPAddress)
}

// GetLogLevel returns the loggging level (as set via config file or environment
// variable)
func (c *Registry) GetLogLevel() string {
	return c.v.GetString(varLogLevel)
}

// IsLogJSON returns if we should log json format (as set via config file or
// environment variable)
func (c *Registry) IsLogJSON() bool {
	if c.v.IsSet(varLogJSON) {
		return c.v.GetBool(varLogJSON)
	}
	return DefaultLogJSON
}
