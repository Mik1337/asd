// Package llm talks to an OpenAI-compatible chat-completions endpoint and
// builds the system prompt. One wire format covers OpenAI, Claude (via
// Anthropic's compatible endpoint), DeepSeek, and local servers (Ollama,
// LM Studio, llama.cpp, ...).
package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client is a minimal OpenAI-compatible chat client.
type Client struct {
	BaseURL     string
	APIKey      string
	Model       string
	Temperature float64
	MaxTokens   int
	Timeout     time.Duration
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model       string    `json:"model"`
	Messages    []message `json:"messages"`
	Temperature float64   `json:"temperature"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
}

type chatResponse struct {
	Choices []struct {
		Message message `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

// Complete sends one system+user exchange and returns the raw assistant text.
func (c *Client) Complete(ctx context.Context, system, user string) (string, error) {
	body, err := json.Marshal(chatRequest{
		Model:       c.Model,
		Temperature: c.Temperature,
		MaxTokens:   c.MaxTokens,
		Messages: []message{
			{Role: "system", Content: system},
			{Role: "user", Content: user},
		},
	})
	if err != nil {
		return "", err
	}

	timeout := c.Timeout
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	url := strings.TrimRight(c.BaseURL, "/") + "/chat/completions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.APIKey)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("timed out after %s (raise `timeout` in config for slow local models)", timeout)
		}
		return "", fmt.Errorf("connection failed: %w", err)
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)

	var cr chatResponse
	_ = json.Unmarshal(raw, &cr)

	switch {
	case resp.StatusCode == http.StatusUnauthorized:
		return "", fmt.Errorf("auth failed (401) — check your key, run `asd config`")
	case resp.StatusCode >= 400:
		if cr.Error != nil && cr.Error.Message != "" {
			return "", fmt.Errorf("provider error (%d): %s", resp.StatusCode, cr.Error.Message)
		}
		return "", fmt.Errorf("provider error (%d)", resp.StatusCode)
	}
	if len(cr.Choices) == 0 {
		return "", fmt.Errorf("empty response from provider")
	}
	return strings.TrimSpace(cr.Choices[0].Message.Content), nil
}
