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
	listenAddr = flag.String("l", "", "the listen addr")
	pemPath    = flag.String("pem", "", "your pem path")
	serverAddr = flag.String("s", "", "the server addr")
)

// Configuration represents a configuration
type Configuration struct {
	ListenAddr string                `toml:"listen_addr"`
	Logger     astilog.Configuration `toml:"logger"`
	PEMPath    string                `toml:"pem_path"`
	ServerAddr string                `toml:"server_addr"`
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
		ServerAddr: ServerAddr,
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
		PEMPath:    *pemPath,
		ServerAddr: *serverAddr,
	}

	// Merge configs
	if err := mergo.Merge(&c, gc); err != nil {
		xlog.Fatalf("%v while merging configs", err)
	}

	// Return
	return c
}
