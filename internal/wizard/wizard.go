// Package wizard runs the interactive first-time setup that writes config.toml.
package wizard

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"asd/internal/config"
	"asd/internal/llm"
)

type preset struct {
	name, label, url, model string
}

var presets = []preset{
	{"openai", "OpenAI", "https://api.openai.com/v1", "gpt-5.1"},
	{"claude", "Claude (Anthropic)", "https://api.anthropic.com/v1", "claude-sonnet-4-6"},
	{"deepseek", "DeepSeek", "https://api.deepseek.com/v1", "deepseek-chat"},
	{"custom", "OpenAI-compatible / local (custom URL)", "", ""},
}

// Run walks the user through provider selection and saves the result.
func Run() (*config.Config, error) {
	in := bufio.NewReader(os.Stdin)

	fmt.Fprintln(os.Stderr, "Let's connect a provider:")
	for i, p := range presets {
		fmt.Fprintf(os.Stderr, "  %d) %s\n", i+1, p.label)
	}
	fmt.Fprint(os.Stderr, "> ")
	idx := parseChoice(readLine(in), len(presets))
	if idx < 0 {
		return nil, fmt.Errorf("invalid choice")
	}
	p := presets[idx]

	cfg := config.Default()
	cfg.Provider = p.name
	cfg.BaseURL = p.url
	cfg.Model = p.model

	if p.name == "custom" {
		fmt.Fprint(os.Stderr, "Base URL (e.g. http://localhost:11434/v1): ")
		cfg.BaseURL = readLine(in)
		fmt.Fprint(os.Stderr, "Model: ")
		cfg.Model = readLine(in)
	} else {
		fmt.Fprintf(os.Stderr, "Model [%s]: ", cfg.Model)
		if m := readLine(in); m != "" {
			cfg.Model = m
		}
	}

	cfg.APIKey = readSecret(in, "API key (leave blank for none): ")

	// Eagerly validate so a bad key surfaces now, not mid-flow later.
	fmt.Fprint(os.Stderr, "Testing connection… ")
	if err := testCall(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "failed: %v\n", err)
		fmt.Fprint(os.Stderr, "Save anyway? [y/N]: ")
		if !yes(readLine(in)) {
			return nil, fmt.Errorf("aborted")
		}
	} else {
		fmt.Fprintln(os.Stderr, "✓ works")
	}

	if err := cfg.Save(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func testCall(cfg *config.Config) error {
	client := &llm.Client{
		BaseURL:     cfg.BaseURL,
		APIKey:      cfg.ResolveKey(),
		Model:       cfg.Model,
		Temperature: cfg.Temperature,
		MaxTokens:   16,
		Timeout:     15 * time.Second,
	}
	_, err := client.Complete(context.Background(), "Reply with the single word: ok", "ping")
	return err
}

func readLine(in *bufio.Reader) string {
	line, _ := in.ReadString('\n')
	return strings.TrimSpace(line)
}

// readSecret reads a line with terminal echo disabled when possible.
func readSecret(in *bufio.Reader, prompt string) string {
	fmt.Fprint(os.Stderr, prompt)
	off := setEcho(false)
	line := readLine(in)
	if off {
		setEcho(true)
		fmt.Fprintln(os.Stderr)
	}
	return line
}

// setEcho toggles terminal echo via stty; returns true if it took effect.
func setEcho(on bool) bool {
	arg := "-echo"
	if on {
		arg = "echo"
	}
	c := exec.Command("stty", arg)
	c.Stdin = os.Stdin
	return c.Run() == nil
}

func parseChoice(s string, n int) int {
	if s == "" {
		return -1
	}
	v := 0
	for _, r := range s {
		if r < '0' || r > '9' {
			return -1
		}
		v = v*10 + int(r-'0')
	}
	if v < 1 || v > n {
		return -1
	}
	return v - 1
}

func yes(s string) bool {
	return strings.HasPrefix(strings.ToLower(strings.TrimSpace(s)), "y")
}
