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
	addrHTTP      = flag.String("http-addr", "", "the HTTP listen addr")
	addrUDP       = flag.String("udp-addr", "", "the UDP listen addr")
	configPath    = flag.String("c", "", "the config path")
	pathStatic    = flag.String("static", "", "the static path")
	pathTemplates = flag.String("templates", "", "the templates path")
)

// Configuration represents a configuration
// TODO Find a way not to put the mongo configuration here so that people who want to use another storage can
type Configuration struct {
	Addr          ConfigurationAddr     `toml:"addr"`
	Builder       builder.Configuration `toml:"builder"`
	Logger        astilog.Configuration `toml:"logger"`
	Mongo         astimgo.Configuration `toml:"mongo"`
	PathStatic    string                `toml:"path_static"`
	PathTemplates string                `toml:"path_templates"`
}

// ConfigurationAddr represents an addr configuration
type ConfigurationAddr struct {
	HTTP string `toml:"http"`
	UDP  string `toml:"udp"`
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
		Addr: ConfigurationAddr{
			HTTP: *addrHTTP,
			UDP:  *addrUDP,
		},
		Builder:       builder.FlagConfig(),
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
