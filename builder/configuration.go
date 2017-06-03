package builder

import "flag"

// Flags
var (
	ServerHTTPAddr       = flag.String("server-http-addr", "", "the HTTP server addr")
	ServerUDPAddr        = flag.String("server-ud-addr", "", "the UDP server addr")
	WorkingDirectoryPath = flag.String("working-directory", "", "the working directory path")
)

// Configuration represents a configuration
type Configuration struct {
	ServerHTTPAddr       string `toml:"server_http_addr"`
	ServerUDPAddr        string `toml:"server_udp_addr"`
	WorkingDirectoryPath string `toml:"working_directory_path"`
}

// FlagConfig returns a configuration based on flags
func FlagConfig() Configuration {
	return Configuration{
		ServerHTTPAddr:       *ServerHTTPAddr,
		ServerUDPAddr:        *ServerUDPAddr,
		WorkingDirectoryPath: *WorkingDirectoryPath,
	}
}
