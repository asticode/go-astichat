package builder

import "flag"

// Flags
var (
	PathWorkingDirectory = flag.String("working-directory", "", "the working directory path")
	ServerHTTPAddr       = flag.String("server-http-addr", "", "the ĤTTP server addr")
	ServerUDPAddr        = flag.String("server-ud-addr", "", "the UDP server addr")
)

// Configuration represents a configuration
type Configuration struct {
	PathWorkingDirectory string `toml:"path_working_directory"`
	ServerHTTPAddr       string `toml:"server_http_addr"`
	ServerUDPAddr        string `toml:"server_udp_addr"`
}

// FlagConfig returns a configuration based on flags
func FlagConfig() Configuration {
	return Configuration{
		PathWorkingDirectory: *PathWorkingDirectory,
		ServerHTTPAddr:       *ServerHTTPAddr,
		ServerUDPAddr:        *ServerUDPAddr,
	}
}
