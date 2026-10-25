package constant

// Log levels accepted by the logger and by configuration validation.
const (
	LogLevelDebug = "debug"
	LogLevelInfo  = "info"
	LogLevelWarn  = "warn"
	LogLevelError = "error"
)

// Log encodings accepted by the logger and by configuration validation.
const (
	LogEncodingJSON    = "json"
	LogEncodingConsole = "console"
)

// Log output streams used by the logger.
const (
	LogOutputStdout = "stdout"
	LogOutputStderr = "stderr"
)
