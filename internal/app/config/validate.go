package config

import (
	"fmt"

	"modelrail/internal/constant"
)

// Validate checks the merged configuration for values the runtime cannot work
// with. The first violation found is returned.
func (c Config) Validate() (err error) {
	if err = validateApp(c.App); err != nil {
		return
	}
	if err = validateLogging(c.Logging); err != nil {
		return
	}
	if err = validateServer(c.Server); err != nil {
		return
	}
	if err = validateMock(c.Providers.Mock); err != nil {
		return
	}
	if err = validateHTTPProvider(constant.ProviderDeepSeek, c.Providers.DeepSeek); err != nil {
		return
	}
	err = validateHTTPProvider(constant.ProviderGemini, c.Providers.Gemini)
	return
}

// validateApp requires a process name and a known environment.
func validateApp(app AppConfig) (err error) {
	if app.Name == constant.EmptyString {
		err = fmt.Errorf("app.name must not be empty")
		return
	}
	switch app.Environment {
	case constant.EnvDev, constant.EnvTest, constant.EnvProd:
	default:
		err = fmt.Errorf("app.environment must be dev, test or prod, got %q", app.Environment)
	}
	return
}

// validateLogging requires a known log level and encoding.
func validateLogging(logging LoggingConfig) (err error) {
	switch logging.Level {
	case constant.LogLevelDebug, constant.LogLevelInfo, constant.LogLevelWarn, constant.LogLevelError:
	default:
		err = fmt.Errorf("logging.level must be debug, info, warn or error, got %q", logging.Level)
		return
	}
	switch logging.Encoding {
	case constant.LogEncodingJSON, constant.LogEncodingConsole:
	default:
		err = fmt.Errorf("logging.encoding must be json or console, got %q", logging.Encoding)
	}
	return
}

// validateServer requires both listeners and non-negative timeouts.
func validateServer(server ServerConfig) (err error) {
	if server.PublicAddr == constant.EmptyString {
		err = fmt.Errorf("server.public_addr must not be empty")
		return
	}
	if server.AdminAddr == constant.EmptyString {
		err = fmt.Errorf("server.admin_addr must not be empty")
		return
	}
	if server.ReadHeaderTimeout.Std() <= 0 {
		err = fmt.Errorf("server.read_header_timeout must be positive")
		return
	}
	if server.ShutdownGrace.Std() < 0 {
		err = fmt.Errorf("server.shutdown_grace must not be negative")
	}
	return
}

// validateMock requires non-negative pacing values; a zero chunk size falls
// back to the mock's own default.
func validateMock(mock MockProviderConfig) (err error) {
	if mock.Name == constant.EmptyString {
		err = fmt.Errorf("providers.mock.name must not be empty")
		return
	}
	if mock.ChunkSize < 0 {
		err = fmt.Errorf("providers.mock.chunk_size must not be negative")
		return
	}
	if mock.Latency.Std() < 0 || mock.TTFT.Std() < 0 || mock.ChunkDelay.Std() < 0 {
		err = fmt.Errorf("providers.mock durations must not be negative")
	}
	return
}

// validateHTTPProvider enforces all-or-nothing provider entries: either the
// endpoint is fully described, or the entry is left empty and disabled.
func validateHTTPProvider(name string, provider HTTPProviderConfig) (err error) {
	empty := provider.BaseURL == constant.EmptyString &&
		provider.Model == constant.EmptyString &&
		provider.APIKeyEnv == constant.EmptyString
	if empty {
		return
	}
	if provider.BaseURL == constant.EmptyString {
		err = fmt.Errorf("providers.%s.base_url must not be empty when the provider is configured", name)
		return
	}
	if provider.Model == constant.EmptyString {
		err = fmt.Errorf("providers.%s.model must not be empty when the provider is configured", name)
		return
	}
	if provider.APIKeyEnv == constant.EmptyString {
		err = fmt.Errorf("providers.%s.api_key_env must not be empty when the provider is configured", name)
	}
	return
}
