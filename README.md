# asd

<img width="3852" height="2150" alt="nano-banana-2-kn7dy8d2t1dapetace3w8x3b7n88f21b-upscaled" src="https://github.com/user-attachments/assets/acf46e21-7d81-4ed2-a082-3485bf43b8dd" />

For when `man` is a wall of text and your memory is worse.

`asd` helps when you roughly know the command, but not the exact syntax.
Instead of opening a browser or chat window, ask from your terminal and
get back the command you meant to type.

It prints the command to stdout, shows a short explanation on stderr, and
copies the command to your clipboard. Nothing is executed.

````console
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
````

Command goes to stdout (and the clipboard); the explanation goes to stderr.
Nothing is executed — you review and run it yourself.

## Install

```fish
go install github.com/Mik1337/asd@latest
fish_add_path ~/go/bin                  # once, if ~/go/bin isn't on PATH

# Homebrew
brew install Mik1337/tap/asd
```

Works with any OpenAI-compatible API — OpenAI, Claude, DeepSeek, or a local
model (Ollama, LM Studio, llama.cpp, etc.).

## Usage

```text
asd <request>      get the command for what you're trying to do
asd config         set up or change your provider

  -e, --explain    flag-by-flag explanation
  -q, --quiet      print command only
  -m, --model      override the configured model for this call
```

On first run, if no config exists, `asd` will prompt for a provider and write the config file for you.

## Configuration

`~/.config/asd/config.toml` (or `$XDG_CONFIG_HOME/asd/config.toml`):

```toml
# Provider: openai | claude | deepseek | custom
# Named providers fill in base_url automatically.
# Use "custom" if you want to set base_url yourself.
provider = "claude"
model    = "claude-sonnet-4-6"
api_key  = "sk-ant-..."

explain     = "brief"   # off | brief | rich
timeout     = 30        # seconds
temperature = 0.2
max_tokens  = 300
```

### Options

| Key             | Type   | Default  | Description                                               |
| --------------- | ------ | -------- | --------------------------------------------------------- |
| `provider`      | string | —        | `openai`, `claude`, `deepseek`, or `custom`               |
| `base_url`      | string | —        | OpenAI-compatible endpoint; required for `custom`         |
| `model`         | string | —        | Model name expected by the endpoint                       |
| `api_key`       | string | —        | Inline API key; omit for local models that don't need one |
| `api_key_env`   | string | —        | Env var name to read the key from instead of inlining it  |
| `explain`       | string | `brief`  | Explanation verbosity: `off`, `brief`, or `rich`          |
| `timeout`       | int    | `30`     | Request timeout in seconds                                |
| `temperature`   | float  | `0.2`    | Sampling temperature; lower is more deterministic         |
| `max_tokens`    | int    | `300`    | Response cap                                              |
| `system_prompt` | string | built-in | Override the default system prompt entirely               |

**API key resolution** (first match wins):

```text
ASD_API_KEY -> api_key_env -> api_key
```

**Default base URLs** (filled in from `provider`):

| `provider` | `base_url`                     |
| ---------- | ------------------------------ |
| `openai`   | `https://api.openai.com/v1`    |
| `claude`   | `https://api.anthropic.com/v1` |
| `deepseek` | `https://api.deepseek.com/v1`  |
| `custom`   | you set `base_url`             |

`asd` works with any OpenAI-compatible API, including local models via Ollama, LM Studio, llama.cpp, and similar tools.

The system prompt lives at [`internal/llm/prompt.txt`](internal/llm/prompt.txt).
`OS` and `shell` are detected at runtime; mention another target in your
request if you need one (`asd list env vars in powershell`).

## Update

```fish
go install github.com/Mik1337/asd@latest    # Go
brew upgrade asd                            # Homebrew
```

## License

[MIT](LICENSE)
