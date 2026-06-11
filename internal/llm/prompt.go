package llm

import (
	_ "embed"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// promptTemplate is the editable system prompt. Tweak prompt.txt to change
// asd's behaviour without touching code; {OS}, {SHELL} and {DETAIL} are filled
// in at runtime.
//
//go:embed prompt.txt
var promptTemplate string

// detail is the per-level guidance appended via the {DETAIL} placeholder.
var detail = map[string]string{
	"off":   "Keep the description to a terse half-sentence.",
	"brief": "Keep the description to one concise line.",
	"rich":  "Make the description a teaching breakdown: explain the command and what each flag and argument does (it may span several lines before the [command]: tag).",
}

// DetectOS returns a human label for the current OS.
func DetectOS() string {
	switch runtime.GOOS {
	case "darwin":
		return "macOS"
	case "linux":
		return "Linux"
	case "windows":
		return "Windows"
	default:
		return runtime.GOOS
	}
}

// DetectShell returns the basename of $SHELL, defaulting to "sh".
func DetectShell() string {
	if s := os.Getenv("SHELL"); s != "" {
		return filepath.Base(s)
	}
	return "sh"
}

// SystemPrompt fills the template for the given explanation level
// (off | brief | rich).
func SystemPrompt(level string) string {
	d, ok := detail[level]
	if !ok {
		d = detail["brief"]
	}
	out := promptTemplate
	out = strings.ReplaceAll(out, "{OS}", DetectOS())
	out = strings.ReplaceAll(out, "{SHELL}", DetectShell())
	out = strings.ReplaceAll(out, "{DETAIL}", d)
	return strings.TrimSpace(out)
}
