package main

import (
	"flag"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/asticode/go-astichat/builder"
	"github.com/asticode/go-astilog"
	"github.com/asticode/go-astimgo"
	"github.com/imdario/mergo"
	"github.com/rs/xlog"
)

// Flags
var (
	configPath    = flag.String("c", "", "the config path")
	listenAddr    = flag.String("l", "", "the listen addr")
	pathStatic    = flag.String("s", "", "the static path")
	pathTemplates = flag.String("t", "", "the templates path")
)

// Configuration represents a configuration
type Configuration struct {
	Builder       builder.Configuration `toml:"builder"`
	ListenAddr    string                `toml:"listen_addr"`
	Logger        astilog.Configuration `toml:"logger"`
	Mongo         astimgo.Configuration `toml:"mongo"`
	PathStatic    string                `toml:"path_static"`
	PathTemplates string                `toml:"path_templates"`
}

// TOMLDecodeFile allows testing functions using it
var TOMLDecodeFile = func(fpath string, v interface{}) (toml.MetaData, error) {
	return toml.DecodeFile(fpath, v)
}

// NewConfiguration creates a new configuration object
func NewConfiguration() Configuration {
	// Global config
	gc := Configuration{
		Builder: builder.Configuration{
			KeyBits: 4096,
		},
		Logger: astilog.Configuration{
			AppName: "go-astichat-server",
		},
		Mongo: astimgo.Configuration{
			Timeout: 10 * time.Second,
		},
		PathStatic:    "static",
		PathTemplates: "templates",
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
		Builder:       builder.FlagConfig(),
		ListenAddr:    *listenAddr,
		Logger:        astilog.FlagConfig(),
		Mongo:         astimgo.FlagConfig(),
		PathStatic:    *pathStatic,
		PathTemplates: *pathTemplates,
	}

	// Merge configs
	if err := mergo.Merge(&c, gc); err != nil {
		xlog.Fatalf("%v while merging configs", err)
	}

	// Return
	return c
}
