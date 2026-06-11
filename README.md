# asd

Translate a natural-language description into a shell command. Prints the
command and a one-line explanation, and copies the command to the clipboard.

```console
$ asd undo my last commit but keep the changes
undoes the last commit, leaving its changes staged in your working tree
git reset --soft HEAD~1

$ asd kill whatever is hogging port 3000
finds the process bound to TCP port 3000 and force-kills it
lsof -ti tcp:3000 | xargs kill -9

$ asd squash the last 3 commits into one
interactively rebase the last three commits so you can squash them
git rebase -i HEAD~3

$ asd trim the first 10 seconds off video.mp4 without re-encoding
copies video.mp4 from the 10s mark with no re-encoding
ffmpeg -ss 10 -i video.mp4 -c copy trimmed.mp4
```

Command goes to stdout (and the clipboard); the explanation goes to stderr.
Nothing is executed — you review and run it yourself.

## Install

```fish
go install github.com/Mik1337/asd@latest
fish_add_path ~/go/bin                  # once, if ~/go/bin isn't on PATH

# Homebrew (after the first tagged release)
brew install Mik1337/tap/asd
```

Provider-agnostic: any OpenAI-compatible endpoint (OpenAI, Claude, DeepSeek, or
a local model via Ollama / LM Studio / llama.cpp).

## Usage

```text
asd <request>      translate a request into a command
asd config         (re)connect a provider

  -e, --explain    flag-by-flag explanation
  -q, --quiet      command only, no explanation
  -m, --model      override the configured model for this call
```

On first run with no config, `asd` prompts for a provider and writes the file.

## Configuration

`~/.config/asd/config.toml` (or `$XDG_CONFIG_HOME/asd/config.toml`):

```toml
# Provider — one of: openai | claude | deepseek | custom.
# A named provider fills in base_url for you; "custom" requires it explicitly.
provider = "claude"
model    = "claude-sonnet-4-6"
api_key  = "sk-ant-..."

# Behaviour
explain     = "brief"   # off | brief | rich
timeout     = 30        # seconds
temperature = 0.2
max_tokens  = 300
```

### Options

| Key             | Type   | Default  | Description                                                        |
| --------------- | ------ | -------- | ------------------------------------------------------------------ |
| `provider`      | string | —        | `openai`, `claude`, `deepseek`, or `custom`. Sets `base_url`.      |
| `base_url`      | string | —        | OpenAI-compatible endpoint. Required only when `provider = custom`. |
| `model`         | string | —        | Model name the endpoint expects.                                   |
| `api_key`       | string | —        | Inline API key. Omit for local models that need none.              |
| `api_key_env`   | string | —        | Name of an env var to read the key from instead of inlining it.    |
| `explain`       | string | `brief`  | Explanation verbosity: `off`, `brief`, or `rich`.                  |
| `timeout`       | int    | `30`     | Request timeout in seconds. Raise for slow local models.           |
| `temperature`   | float  | `0.2`    | Sampling temperature. Lower is more deterministic.                 |
| `max_tokens`    | int    | `300`    | Response cap.                                                      |
| `system_prompt` | string | built-in | Override the system prompt entirely.                              |

**API key resolution** (first match wins): `ASD_API_KEY` env var →
`api_key_env` → inline `api_key`.

**Provider base URLs** (set automatically by `provider`):

| `provider` | `base_url`                     |
| ---------- | ------------------------------ |
| `openai`   | `https://api.openai.com/v1`    |
| `claude`   | `https://api.anthropic.com/v1` |
| `deepseek` | `https://api.deepseek.com/v1`  |
| `custom`   | set `base_url` yourself        |

The system prompt is plain text at
[`internal/llm/prompt.txt`](internal/llm/prompt.txt). `OS` and `shell` are
detected at runtime; name another target in the request to override
(`asd list env vars in powershell`).

## Update

```fish
go install github.com/Mik1337/asd@latest    # Go
brew upgrade asd                            # Homebrew
```

## License

[MIT](LICENSE)
