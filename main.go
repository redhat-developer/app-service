package main

import (
	"flag"
	"log"
	"os"

	"github.com/redhat-developer/boilerplate-app/configuration"
)

func main() {
	// Parse flags
	var configFilePath string
	flag.StringVar(&configFilePath, "config", "", "path to the config file to read (if none is given, defaults will be used)")
	flag.Parse()

	// Override default -config switch with environment variable only if -config
	// switch was not explicitly given via the command line.
	configSwitchIsSet := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == "config" {
			configSwitchIsSet = true
		}
	})
	if !configSwitchIsSet {
		if envConfigPath, ok := os.LookupEnv(configuration.EnvPrefix + "_CONFIG_FILE_PATH"); ok {
			configFilePath = envConfigPath
		}
	}

	config, err := configuration.New(configFilePath)
	if err != nil {
		log.Panic(nil, map[string]interface{}{
			"config_file_path": configFilePath,
			"err":              err,
		}, "failed to setup the configuration")
	}
	_ = config
}
