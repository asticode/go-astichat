package builder

import "flag"

// Flags
var (
	KeyBits              = flag.Int("key-bits", 0, "the private key bits")
	PathProjectRoot      = flag.String("project-root", "", "the root project path")
	PathWorkingDirectory = flag.String("working-directory", "", "the working directory path")
	ServerAddr           = flag.String("server-addr", "", "the server addr")
)

// Configuration represents a configuration
type Configuration struct {
	KeyBits              int    `toml:"key_bits"`
	PathProjectRoot      string `toml:"path_project_root"`
	PathWorkingDirectory string `toml:"path_working_directory"`
	ServerAddr           string `toml:"server_addr"`
}

// FlagConfig returns a configuration based on flags
func FlagConfig() Configuration {
	return Configuration{
		KeyBits:              *KeyBits,
		PathProjectRoot:      *PathProjectRoot,
		PathWorkingDirectory: *PathWorkingDirectory,
		ServerAddr:           *ServerAddr,
	}
}
