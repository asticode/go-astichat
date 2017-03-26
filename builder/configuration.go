package builder

import "flag"

// Flags
var (
	KeyBits              = flag.Int("kb", 0, "the private key bits")
	PathRootProject      = flag.String("r", "", "the root project path")
	PathWorkingDirectory = flag.String("w", "", "the working directory path")
)

// Configuration represents a configuration
type Configuration struct {
	KeyBits              int    `toml:"key_bits"`
	PathRootProject      string `toml:"path_root_project"`
	PathWorkingDirectory string `toml:"path_working_directory"`
}

// FlagConfig returns a configuration based on flags
func FlagConfig() Configuration {
	return Configuration{
		KeyBits:              *KeyBits,
		PathRootProject:      *PathRootProject,
		PathWorkingDirectory: *PathWorkingDirectory,
	}
}
