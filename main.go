// Command asd turns a natural-language request into a shell command.
//
//	asd git show me origin      → git remote show origin   (+ a description)
//
// stdout is always the bare command (so it copies to the clipboard cleanly);
// the human-readable description goes to stderr.
package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/Mik1337/asd/internal/config"
	"github.com/Mik1337/asd/internal/llm"
	"github.com/Mik1337/asd/internal/parse"
	"github.com/Mik1337/asd/internal/wizard"
)

// version is overridden at build time via -ldflags "-X main.version=...".
var version = "dev"

const usage = `asd — describe a command, get the command.

Usage:
  asd <natural language request>     translate a request into a shell command
  asd config                         set up or change your provider

The command is printed and copied to your clipboard, ready to paste.

Flags:
  -e, --explain        expressive, flag-by-flag explanation
  -q, --quiet          command only, no explanation
  -m, --model NAME     override the configured model for this call
  -v, --version        print version
  -h, --help           show this help

Issues & source: https://github.com/Mik1337/asd`

type args struct {
	query []string
	model string
	level string // "" = use config; otherwise off | rich
}

func main() {
	a := os.Args[1:]

	if len(a) == 0 {
		// Bare `asd`: set up if unconfigured, else show help.
		if _, err := config.Load(); err != nil {
			runConfig()
			return
		}
		fmt.Fprintln(os.Stderr, usage)
		os.Exit(2)
	}

	switch a[0] {
	case "config", "login":
		runConfig()
		return
	case "version", "--version", "-v":
		fmt.Println("asd", version)
		return
	case "help", "--help", "-h":
		fmt.Println(usage)
		fmt.Println("\nConfig file:", config.Path())
		return
	}

	p, err := parseArgs(a)
	if err != nil {
		fmt.Fprintln(os.Stderr, "asd:", err)
		os.Exit(2)
	}
	if len(p.query) == 0 {
		fmt.Fprintln(os.Stderr, usage)
		os.Exit(2)
	}
	run(p)
}

func parseArgs(a []string) (args, error) {
	var out args
	for i := 0; i < len(a); i++ {
		t := a[i]
		switch {
		case t == "--":
			out.query = append(out.query, a[i+1:]...)
			return out, nil
		case t == "--quiet":
			out.level = "off"
		case t == "--explain":
			out.level = "rich"
		case t == "--model":
			if i+1 >= len(a) {
				return out, fmt.Errorf("--model needs a value")
			}
			i++
			out.model = a[i]
		case strings.HasPrefix(t, "--model="):
			out.model = strings.TrimPrefix(t, "--model=")
		case len(t) > 1 && t[0] == '-' && t[1] != '-':
			// short flag cluster, e.g. -e, -q, -m, -em <model>
			cluster := t[1:]
			for j := 0; j < len(cluster); j++ {
				switch cluster[j] {
				case 'q':
					out.level = "off"
				case 'e':
					out.level = "rich"
				case 'm':
					if rest := cluster[j+1:]; rest != "" {
						out.model = rest
					} else {
						if i+1 >= len(a) {
							return out, fmt.Errorf("-m needs a value")
						}
						i++
						out.model = a[i]
					}
					j = len(cluster) // -m consumes the rest
				default:
					return out, fmt.Errorf("unknown flag -%c", cluster[j])
				}
			}
		default:
			out.query = append(out.query, t)
		}
	}
	return out, nil
}

func run(p args) {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, "asd: no provider configured — let's set one up.")
		if cfg, err = wizard.Run(); err != nil {
			fmt.Fprintln(os.Stderr, "asd:", err)
			os.Exit(1)
		}
	}

	level := cfg.Explain
	if p.level != "" {
		level = p.level
	}
	model := cfg.Model
	if p.model != "" {
		model = p.model
	}

	client := &llm.Client{
		BaseURL:     cfg.BaseURL,
		APIKey:      cfg.ResolveKey(),
		Model:       model,
		Temperature: cfg.Temperature,
		MaxTokens:   cfg.MaxTokens,
		Timeout:     time.Duration(cfg.Timeout) * time.Second,
	}

	sys := cfg.SystemPrompt
	if sys == "" {
		sys = llm.SystemPrompt(level)
	}

	raw, err := client.Complete(context.Background(), sys, strings.Join(p.query, " "))
	if err != nil {
		// Failure invariant: stdout stays empty so nothing corrupts the prompt.
		fmt.Fprintln(os.Stderr, "asd:", err)
		os.Exit(1)
	}

	cmd, desc := parse.Tags(raw)
	if cmd == "" {
		fmt.Fprintln(os.Stderr, "asd: no command in response")
		os.Exit(1)
	}

	// Description → stderr (shown, never captured); command → stdout (clean).
	if level != "off" && desc != "" {
		fmt.Fprintln(os.Stderr, desc)
	}
	fmt.Fprintln(os.Stdout, cmd)

	// Copy the command so it's one paste away.
	if clipboard(cmd) == nil {
		fmt.Fprintln(os.Stderr, "(copied to clipboard)")
	}
}

func runConfig() {
	if _, err := wizard.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "asd:", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "✓ Saved to %s\n", config.Path())
}

// clipboard copies s to the system clipboard, best-effort.
func clipboard(s string) error {
	var c *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		c = exec.Command("pbcopy")
	case "linux":
		if _, err := exec.LookPath("wl-copy"); err == nil {
			c = exec.Command("wl-copy")
		} else if _, err := exec.LookPath("xclip"); err == nil {
			c = exec.Command("xclip", "-selection", "clipboard")
		} else {
			return fmt.Errorf("no clipboard tool found")
		}
	default:
		return fmt.Errorf("clipboard unsupported on %s", runtime.GOOS)
	}
	c.Stdin = strings.NewReader(s)
	return c.Run()
}
