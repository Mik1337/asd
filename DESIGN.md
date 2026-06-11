# asd — design

`asd` turns a natural-language request into a shell command, in the terminal,
without leaving your flow. You describe what you want; it prints a short
explanation and leaves the command ready to edit or run.

```
$ asd git show me origin
shows the remote named "origin" and the branches it tracks
git remote show origin
(copied to clipboard)
```

The goal is to stay *inline* — no alt-tabbing to a chat app — and to **teach**:
every answer explains what the command does so you learn it, not just run it.

---

## Decisions (and the tradeoffs we accepted)

### 1. Print-only binary with one contract
`asd <words>` prints **the bare command to stdout** and **the description to
stderr**. stdout is *always* command-only, so it (and the clipboard copy) is
clean and runnable. Nothing else is ever written to stdout. You trigger a
translation the same way you run any command: type `asd <request>` and press
Enter — no special keybinding to learn.

### 2. One mode: print + clipboard
`asd <words>` + Enter prints the explanation and the command, and copies the
command to your clipboard so it's one paste away.

We considered injecting the command *editable into the next prompt*, but that
requires a shell-side key-binding — a child binary cannot write into its parent
shell's edit buffer, and fish offers no `print -z` equivalent. We chose **not**
to use a hotkey, so there is no shim: you paste from the clipboard instead. This
keeps a single mode, no shell integration to install, and the same review-before-
run safety (you see the command, then choose to paste/run it).

*Accepted tradeoff:* you press Cmd+V to run, rather than the command auto-filling
your prompt.

### 3. Engine: one OpenAI-compatible wire format
The tool speaks `POST /v1/chat/completions`. That single path covers OpenAI,
Claude (via Anthropic's OpenAI-compatible endpoint), DeepSeek, and every local
server (Ollama, LM Studio, llama.cpp, vLLM) plus aggregators (OpenRouter). The
job — short request in, one command out — needs no provider-specific features,
so the lowest common denominator *is* the whole requirement.

*Accepted tradeoff:* "Claude mode" rides Anthropic's compat endpoint rather than
the native API. The `provider` config field is the seam where a native adapter
could slot in later.

### 4. Config: a file, written by a wizard, hand-editable too
`~/.config/asd/config.toml` (respects `$XDG_CONFIG_HOME`) is the source of
truth. First run (or `asd config`) launches an interactive wizard that picks a
provider preset, prompts for a key, **makes a live test call**, and writes the
file. Power users skip the wizard and drop their own file.

Key resolution precedence: `ASD_API_KEY` env → `api_key_env` → inline `api_key`.
The wizard writes inline keys with `chmod 600`. Single provider in v1 (no
profiles yet, but the file can grow `[providers.x]` tables later).

### 5. Model behaviour
- **Context:** the default `OS` + `shell` are injected into the prompt — the
  single highest-leverage accuracy win, with no sensitive leakage. If the
  request names a *different* target ("in powershell", "on linux", "as a python
  one-liner"), the model targets that instead of the default.
- **Output:** labeled tags — `[description]:` / `[command]:` — parsed
  leniently (order-independent, fences stripped, multi-line description for rich
  mode). If the tags are missing entirely, the first non-empty line is taken as
  the command so you're never left with nothing.
- **One answer.** No candidate picker — the shim injects one line anyway, and
  the command lands *editable*, so "close enough" is fixed in place. Unhappy?
  Re-ask.

### 6. Explanation levels — always on by default, tunable
The explanation is a core feature (learning, not a black box), so it is **on by
default**. Levels: `off | brief | rich`.
- bare (default) → **brief**, one concise line.
- `-e` / `--explain` → **rich**, flag-by-flag teaching breakdown.
- `-q` / `--quiet` → **off**, command only.

`-e` and `-q` are opposites (you use one or the other); the grammar is standard
POSIX short flags, so independent flags stack (`asd -em gpt-5.1 "…"`).

### 7. Failure behaviour — the hard invariant
**On any failure, stdout stays empty** and nothing is copied to the clipboard,
so a broken call can never hand you garbage to paste. Beyond that: errors go to
stderr; a `401` points at `asd config`; the timeout is configurable (default
30s) for slow local models.

*Deliberately excluded:* there is **no `auto_run` setting**. The whole safety
story is "command lands editable, human presses Enter"; an auto-execute knob
would quietly delete that safety net.

---

## Configuration

```toml
# ~/.config/asd/config.toml

# Provider (required)
provider = "claude"          # openai | claude | deepseek | custom
model    = "claude-sonnet-4-6"
api_key  = "sk-ant-..."      # inline (chmod 600); or:
# api_key_env = "ANTHROPIC_API_KEY"   # read key from this env var
# base_url = "http://localhost:11434/v1"   # required only for provider = custom
#   (env var ASD_API_KEY overrides everything above)

# Behaviour
explain     = "brief"        # off | brief | rich
timeout     = 30             # seconds (raise for slow local models)

# Advanced
temperature = 0.2            # low = deterministic commands
max_tokens  = 300            # responses are short
# system_prompt = "..."      # override the built-in prompt entirely
```

`OS`/`shell` are auto-detected, not configured.

---

## Layout

```
asd/
├── main.go                       CLI: arg parsing, orchestration, clipboard
├── internal/
│   ├── config/   config.go       load/save config.toml (BurntSushi/toml)
│   ├── llm/      client.go        OpenAI-compatible chat client
│   │             prompt.go        prompt assembly + OS/shell detection
│   │             prompt.txt       the editable system prompt (tweak freely)
│   ├── parse/    parse.go         [command]/[description] tag extraction
│   └── wizard/   wizard.go        interactive setup + live test call
└── DESIGN.md
```

The system prompt lives in **`internal/llm/prompt.txt`** as plain text
(embedded at build time) precisely so anyone can edit asd's behaviour without
touching Go.

---

## Status / TODO

- [ ] Verify Anthropic's OpenAI-compatible base URL + a sensible default model.
- [ ] `goreleaser` config for the Homebrew tap + prebuilt binaries.
- [ ] Optional `-n` multi-candidate mode if single answers prove insufficient.
