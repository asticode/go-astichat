package builder

import "flag"

// Flags
var (
	PathWorkingDirectory = flag.String("working-directory", "", "the working directory path")
	ServerAddr           = flag.String("server-addr", "", "the server addr")
)

// Configuration represents a configuration
type Configuration struct {
	PathWorkingDirectory string `toml:"path_working_directory"`
	ServerAddr           string `toml:"server_addr"`
}

// FlagConfig returns a configuration based on flags
func FlagConfig() Configuration {
	return Configuration{
		PathWorkingDirectory: *PathWorkingDirectory,
		ServerAddr:           *ServerAddr,
	}
}
