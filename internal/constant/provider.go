package constant

import "time"

// Stable provider names reported by adapters and used in configuration.
const (
	ProviderMock     = "mock"
	ProviderDeepSeek = "deepseek"
	ProviderGemini   = "gemini"
)

// Provider diagnostic codes returned inside ProviderError.ProviderCode. They
// are not client-visible error codes.
const (
	ProviderCodeContextCanceled         = "context_canceled"
	ProviderCodeContextDeadlineExceeded = "context_deadline_exceeded"
)

// Default mock provider settings.
const (
	DefaultProviderName     = ProviderMock
	DefaultChunkSize        = 16
	DefaultContextWindow    = 8192
	DefaultMockLatency      = 200 * time.Millisecond
	DefaultMockTTFT         = 100 * time.Millisecond
	DefaultMockChunkDelay   = 20 * time.Millisecond
	DefaultMockResponseText = "Hello from the ModelRail mock provider."
)

// Default external provider settings. Credentials are never stored here, only
// the name of the environment variable that holds them.
const (
	DefaultDeepSeekBaseURL   = "https://api.deepseek.com"
	DefaultDeepSeekModel     = "deepseek-chat"
	DefaultDeepSeekAPIKeyEnv = "DEEPSEEK_API_KEY"

	DefaultGeminiBaseURL   = "https://generativelanguage.googleapis.com"
	DefaultGeminiModel     = "gemini-2.5-flash"
	DefaultGeminiAPIKeyEnv = "GEMINI_API_KEY"
)
