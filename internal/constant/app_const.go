// Package constant holds the small, dependency-free value types and constants
// shared by every layer of ModelRail.
package constant

const (
	// EmptyString is a named empty string used where a literal "" would be less
	// self-describing.
	EmptyString = ""
	// AppName is the process and service name used in logs and metadata.
	AppName = "modelrail"
)

// Command-line flags and their defaults.
const (
	// FlagConfig names the flag that selects the configuration file.
	FlagConfig = "config"
	// DefaultConfigPath is the configuration file loaded when none is supplied.
	DefaultConfigPath = "config.yml"
)
