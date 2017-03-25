package main

import (
	"flag"

	"github.com/BurntSushi/toml"
	"github.com/asticode/go-astilog"
	"github.com/imdario/mergo"
	"github.com/rs/xlog"
)

// Flags
var (
	configPath = flag.String("c", "", "the config path")
	passphrase = flag.String("p", "", "the passphrase")
	pemPath    = flag.String("pem", "", "your pem path")
)

// Configuration represents a configuration
type Configuration struct {
	Logger     astilog.Configuration `toml:"logger"`
	Passphrase string                `toml:"passphrase"`
	PEMPath    string                `toml:"pem_path"`
}

// TOMLDecodeFile allows testing functions using it
var TOMLDecodeFile = func(fpath string, v interface{}) (toml.MetaData, error) {
	return toml.DecodeFile(fpath, v)
}

// NewConfiguration creates a new configuration object
func NewConfiguration() Configuration {
	// Global config
	gc := Configuration{
		Logger: astilog.Configuration{
			AppName: "go-astichat-client",
		},
	}

	// Local config
	if *configPath != "" {
		// Decode local config
		if _, err := TOMLDecodeFile(*configPath, &gc); err != nil {
			xlog.Fatalf("%v while decoding the config path %s", err, *configPath)
		}
	}

	// Flag config
	c := Configuration{
		Logger:     astilog.FlagConfig(),
		Passphrase: *passphrase,
		PEMPath:    *pemPath,
	}

	// Merge configs
	if err := mergo.Merge(&c, gc); err != nil {
		xlog.Fatalf("%v while merging configs", err)
	}

	// Return
	return c
}
