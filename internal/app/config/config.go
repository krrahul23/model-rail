// Package config loads and validates the ModelRail runtime configuration from
// a YAML file. Secrets are never stored in YAML: provider entries reference the
// name of the environment variable that holds the credential.
package config

import (
	"fmt"
	"os"

	"modelrail/internal/constant"

	"gopkg.in/yaml.v3"
)

// Config is the root runtime configuration.
type Config struct {
	App       AppConfig       `yaml:"app"`
	Logging   LoggingConfig   `yaml:"logging"`
	Server    ServerConfig    `yaml:"server"`
	Providers ProvidersConfig `yaml:"providers"`
}

// AppConfig holds process-level identity.
type AppConfig struct {
	Name        string `yaml:"name"`
	Environment string `yaml:"environment"` // dev | test | prod
}

// LoggingConfig holds structured-logging settings.
type LoggingConfig struct {
	Level    string `yaml:"level"`    // debug | info | warn | error
	Encoding string `yaml:"encoding"` // json | console
}

// ServerConfig holds the HTTP listeners and their shutdown behaviour.
type ServerConfig struct {
	PublicAddr        string   `yaml:"public_addr"`
	AdminAddr         string   `yaml:"admin_addr"`
	ReadHeaderTimeout Duration `yaml:"read_header_timeout"`
	ShutdownGrace     Duration `yaml:"shutdown_grace"`
}

// ProvidersConfig groups one entry per provider adapter.
type ProvidersConfig struct {
	Mock     MockProviderConfig `yaml:"mock"`
	DeepSeek HTTPProviderConfig `yaml:"deepseek"`
	Gemini   HTTPProviderConfig `yaml:"gemini"`
}

// MockProviderConfig drives the deterministic mock adapter.
type MockProviderConfig struct {
	Name         string   `yaml:"name"`
	Latency      Duration `yaml:"latency"`
	TTFT         Duration `yaml:"ttft"`
	ChunkSize    int      `yaml:"chunk_size"`
	ChunkDelay   Duration `yaml:"chunk_delay"`
	ResponseText string   `yaml:"response_text"`
}

// HTTPProviderConfig describes an external provider. The credential itself is
// read from the environment variable named by APIKeyEnv at startup, never from
// this file.
type HTTPProviderConfig struct {
	BaseURL   string `yaml:"base_url"`
	Model     string `yaml:"model"`
	APIKeyEnv string `yaml:"api_key_env"`
}

// Default returns a fully populated configuration. Omitted YAML fields decode
// over these values, so a partial file stays valid.
func Default() (cfg Config) {
	cfg = Config{
		App: AppConfig{
			Name:        constant.AppName,
			Environment: constant.EnvDev,
		},
		Logging: LoggingConfig{
			Level:    constant.LogLevelInfo,
			Encoding: constant.LogEncodingJSON,
		},
		Server: ServerConfig{
			PublicAddr:        constant.DefaultPublicAddr,
			AdminAddr:         constant.DefaultAdminAddr,
			ReadHeaderTimeout: Duration(constant.DefaultReadHeaderTimeout),
			ShutdownGrace:     Duration(constant.DefaultShutdownGrace),
		},
		Providers: ProvidersConfig{
			Mock: MockProviderConfig{
				Name:         constant.DefaultProviderName,
				Latency:      Duration(constant.DefaultMockLatency),
				TTFT:         Duration(constant.DefaultMockTTFT),
				ChunkSize:    constant.DefaultChunkSize,
				ChunkDelay:   Duration(constant.DefaultMockChunkDelay),
				ResponseText: constant.DefaultMockResponseText,
			},
			DeepSeek: HTTPProviderConfig{
				BaseURL:   constant.DefaultDeepSeekBaseURL,
				Model:     constant.DefaultDeepSeekModel,
				APIKeyEnv: constant.DefaultDeepSeekAPIKeyEnv,
			},
			Gemini: HTTPProviderConfig{
				BaseURL:   constant.DefaultGeminiBaseURL,
				Model:     constant.DefaultGeminiModel,
				APIKeyEnv: constant.DefaultGeminiAPIKeyEnv,
			},
		},
	}
	return
}

// Load reads the YAML file at path, decodes it over the defaults and validates
// the merged result.
func Load(path string) (cfg Config, err error) {
	var data []byte
	if data, err = os.ReadFile(path); err != nil {
		return
	}

	cfg = Default()
	if err = yaml.Unmarshal(data, &cfg); err != nil {
		err = fmt.Errorf("decode config %s: %w", path, err)
		return
	}
	if err = cfg.Validate(); err != nil {
		err = fmt.Errorf("validate config %s: %w", path, err)
		return
	}
	return
}
