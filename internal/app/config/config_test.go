package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// repoConfigPath points at the config.yml shipped in the repository root, so
// the checked-in file stays valid as the loader evolves.
const repoConfigPath = "../../../config.yml"

// writeConfig writes content to a temporary YAML file and returns its path.
func writeConfig(t *testing.T, content string) (path string) {
	t.Helper()
	path = filepath.Join(t.TempDir(), "config.yml")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	return
}

func TestDefaultIsValid(t *testing.T) {
	t.Parallel()

	if err := Default().Validate(); err != nil {
		t.Fatalf("Default().Validate() = %v, want nil", err)
	}
}

func TestLoadRepositoryConfig(t *testing.T) {
	t.Parallel()

	cfg, err := Load(repoConfigPath)
	if err != nil {
		t.Fatalf("Load(%q) = %v", repoConfigPath, err)
	}

	if cfg.App.Name != "modelrail" {
		t.Fatalf("App.Name = %q, want %q", cfg.App.Name, "modelrail")
	}
	if got := cfg.Providers.Mock.Latency.Std(); got != 200*time.Millisecond {
		t.Fatalf("Mock.Latency = %s, want 200ms", got)
	}
	if cfg.Providers.Mock.ChunkSize != 16 {
		t.Fatalf("Mock.ChunkSize = %d, want 16", cfg.Providers.Mock.ChunkSize)
	}
	if cfg.Providers.DeepSeek.APIKeyEnv != "DEEPSEEK_API_KEY" {
		t.Fatalf("DeepSeek.APIKeyEnv = %q, want %q", cfg.Providers.DeepSeek.APIKeyEnv, "DEEPSEEK_API_KEY")
	}
}

func TestLoadAppliesDefaultsForOmittedFields(t *testing.T) {
	t.Parallel()

	path := writeConfig(t, "app:\n  name: custom\n")
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() = %v, want nil", err)
	}

	if cfg.App.Name != "custom" {
		t.Fatalf("App.Name = %q, want %q", cfg.App.Name, "custom")
	}
	// Omitted fields must keep the defaults rather than decode to zero.
	if cfg.Server.PublicAddr != ":8080" {
		t.Fatalf("Server.PublicAddr = %q, want %q", cfg.Server.PublicAddr, ":8080")
	}
	if cfg.Logging.Level != "info" {
		t.Fatalf("Logging.Level = %q, want %q", cfg.Logging.Level, "info")
	}
	if cfg.Logging.Encoding != "json" {
		t.Fatalf("Logging.Encoding = %q, want %q", cfg.Logging.Encoding, "json")
	}
}

func TestLoadRejectsUnknownLogLevel(t *testing.T) {
	t.Parallel()

	path := writeConfig(t, "logging:\n  level: verbose\n")
	_, err := Load(path)
	if err == nil {
		t.Fatal("Load() = nil, want validation error")
	}
	if !strings.Contains(err.Error(), "logging.level") {
		t.Fatalf("error %q does not mention logging.level", err)
	}
}

func TestLoadRejectsNegativeDuration(t *testing.T) {
	t.Parallel()

	path := writeConfig(t, "server:\n  shutdown_grace: -1s\n")
	_, err := Load(path)
	if err == nil {
		t.Fatal("Load() = nil, want validation error")
	}
	if !strings.Contains(err.Error(), "shutdown_grace") {
		t.Fatalf("error %q does not mention shutdown_grace", err)
	}
}

func TestLoadRejectsMalformedDuration(t *testing.T) {
	t.Parallel()

	path := writeConfig(t, "providers:\n  mock:\n    latency: fast\n")
	_, err := Load(path)
	if err == nil {
		t.Fatal("Load() = nil, want duration parse error")
	}
	if !strings.Contains(err.Error(), "invalid duration") {
		t.Fatalf("error %q does not mention an invalid duration", err)
	}
}

func TestLoadRejectsPartiallyConfiguredProvider(t *testing.T) {
	t.Parallel()

	path := writeConfig(t, "providers:\n  deepseek:\n    base_url: https://example.invalid\n    model: \"\"\n    api_key_env: \"\"\n")
	_, err := Load(path)
	if err == nil {
		t.Fatal("Load() = nil, want validation error")
	}
	if !strings.Contains(err.Error(), "providers.deepseek") {
		t.Fatalf("error %q does not mention providers.deepseek", err)
	}
}
