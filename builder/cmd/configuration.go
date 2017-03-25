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
	configPath      = flag.String("c", "", "the config path")
	keyBits         = flag.Int("kb", 0, "the private key bits")
	rootProjectPath = flag.String("r", "", "the root project path")
)

// Configuration represents a configuration
type Configuration struct {
	KeyBits         int                   `toml:"key_bits"`
	Logger          astilog.Configuration `toml:"logger"`
	RootProjectPath string                `toml:"root_project_path"`
}

// TOMLDecodeFile allows testing functions using it
var TOMLDecodeFile = func(fpath string, v interface{}) (toml.MetaData, error) {
	return toml.DecodeFile(fpath, v)
}

// NewConfiguration creates a new configuration object
func NewConfiguration() Configuration {
	// Global config
	gc := Configuration{
		KeyBits: 4096,
		Logger: astilog.Configuration{
			AppName: "go-astichat-builder",
		},
		RootProjectPath: ".",
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
		KeyBits:         *keyBits,
		Logger:          astilog.FlagConfig(),
		RootProjectPath: *rootProjectPath,
	}

	// Merge configs
	if err := mergo.Merge(&c, gc); err != nil {
		xlog.Fatalf("%v while merging configs", err)
	}

	// Return
	return c
}
