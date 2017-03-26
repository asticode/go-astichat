package main

import (
	"flag"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/asticode/go-astilog"
	"github.com/asticode/go-astimgo"
	"github.com/imdario/mergo"
	"github.com/rs/xlog"
)

// Flags
var (
	configPath = flag.String("c", "", "the config path")
	listenAddr = flag.String("l", "", "the listen addr")
)

// Configuration represents a configuration
// TODO Find a way not to put the mongo configuration here so that people who want to use another storage can
type Configuration struct {
	ListenAddr string                `toml:"listen_addr"`
	Logger     astilog.Configuration `toml:"logger"`
	Mongo      astimgo.Configuration `toml:"mongo"`
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
			AppName: "go-astichat-server",
		},
		Mongo: astimgo.Configuration{
			Timeout: 10 * time.Second,
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
		ListenAddr: *listenAddr,
		Logger:     astilog.FlagConfig(),
		Mongo:      astimgo.FlagConfig(),
	}

	// Merge configs
	if err := mergo.Merge(&c, gc); err != nil {
		xlog.Fatalf("%v while merging configs", err)
	}

	// Return
	return c
}
