// Package config loads and saves asd's configuration from
// ~/.config/asd/config.toml (or $XDG_CONFIG_HOME/asd/config.toml).
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
)

// ErrNotConfigured means no usable provider config was found.
var ErrNotConfigured = errors.New("no provider configured — run `asd config`")

// Config is the resolved runtime configuration.
type Config struct {
	Provider     string  `toml:"provider"`      // openai | claude | deepseek | custom
	BaseURL      string  `toml:"base_url"`      // OpenAI-compatible endpoint
	Model        string  `toml:"model"`         // model name the endpoint expects
	APIKey       string  `toml:"api_key"`       // inline key (may be empty for local)
	APIKeyEnv    string  `toml:"api_key_env"`   // name of an env var holding the key
	Explain      string  `toml:"explain"`       // off | brief | rich
	Timeout      int     `toml:"timeout"`       // seconds before giving up
	Temperature  float64 `toml:"temperature"`   // sampling temperature
	MaxTokens    int     `toml:"max_tokens"`    // response cap
	SystemPrompt string  `toml:"system_prompt"` // optional prompt override
}

// Default returns a Config populated with the behavioural defaults.
// The provider trio (BaseURL/Model/APIKey) is left empty.
func Default() *Config {
	return &Config{
		Explain:     "brief",
		Timeout:     30,
		Temperature: 0.2,
		MaxTokens:   300,
	}
}

// presets maps a provider shorthand to its OpenAI-compatible base URL.
var presets = map[string]string{
	"openai":   "https://api.openai.com/v1",
	"claude":   "https://api.anthropic.com/v1", // Anthropic's OpenAI-compatible endpoint
	"deepseek": "https://api.deepseek.com/v1",
}

// PresetURL returns the base URL for a known provider shorthand.
func PresetURL(provider string) (string, bool) {
	u, ok := presets[provider]
	return u, ok
}

// Dir is the directory holding config.toml.
func Dir() string {
	if x := os.Getenv("XDG_CONFIG_HOME"); x != "" {
		return filepath.Join(x, "asd")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "asd")
}

// Path is the absolute path to config.toml.
func Path() string { return filepath.Join(Dir(), "config.toml") }

// Load reads and resolves the configuration. It returns ErrNotConfigured
// when the file is missing or yields no usable endpoint. Defaults fill any
// keys the file omits.
func Load() (*Config, error) {
	c := Default()
	if _, err := toml.DecodeFile(Path(), c); err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNotConfigured
		}
		return nil, err
	}

	// Fill the base URL from the provider preset when not given explicitly.
	if c.BaseURL == "" && c.Provider != "" {
		if u, ok := presets[c.Provider]; ok {
			c.BaseURL = u
		}
	}
	if c.BaseURL == "" {
		return nil, ErrNotConfigured
	}
	return c, nil
}

// ResolveKey returns the API key using the precedence:
// ASD_API_KEY env > api_key_env > inline api_key.
func (c *Config) ResolveKey() string {
	if v := os.Getenv("ASD_API_KEY"); v != "" {
		return v
	}
	if c.APIKeyEnv != "" {
		if v := os.Getenv(c.APIKeyEnv); v != "" {
			return v
		}
	}
	return c.APIKey
}

// Save writes the config to disk with 0600 permissions (it may hold a key).
// Hand-formatted so the file stays readable and commented.
func (c *Config) Save() error {
	if err := os.MkdirAll(Dir(), 0o755); err != nil {
		return err
	}

	var b strings.Builder
	b.WriteString("# asd configuration\n\n")
	if c.Provider != "" {
		fmt.Fprintf(&b, "provider = %q\n", c.Provider)
	}
	// Only persist base_url for custom providers; presets derive it.
	if _, isPreset := presets[c.Provider]; !isPreset {
		fmt.Fprintf(&b, "base_url = %q\n", c.BaseURL)
	}
	fmt.Fprintf(&b, "model    = %q\n", c.Model)
	if c.APIKey != "" {
		fmt.Fprintf(&b, "api_key  = %q\n", c.APIKey)
	}
	if c.APIKeyEnv != "" {
		fmt.Fprintf(&b, "api_key_env = %q\n", c.APIKeyEnv)
	}
	b.WriteString("\n")
	fmt.Fprintf(&b, "explain     = %q\n", c.Explain)
	fmt.Fprintf(&b, "timeout     = %d\n", c.Timeout)
	fmt.Fprintf(&b, "temperature = %s\n", strconv.FormatFloat(c.Temperature, 'g', -1, 64))
	fmt.Fprintf(&b, "max_tokens  = %d\n", c.MaxTokens)
	if c.SystemPrompt != "" {
		fmt.Fprintf(&b, "system_prompt = %q\n", c.SystemPrompt)
	}

	return os.WriteFile(Path(), []byte(b.String()), 0o600)
}
