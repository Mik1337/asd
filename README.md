# asd

Describe a command, get the command — inline in your terminal.

```
$ asd git show me origin
shows the remote named "origin" and the branches it tracks
git remote show origin
(copied to clipboard)
```

`asd` translates a natural-language request into a shell command, prints a short
explanation so you actually learn it, and copies the command to your clipboard
ready to paste. It works with OpenAI, Claude, DeepSeek, or any
local/OpenAI-compatible model — your key, your model.

## Install

```sh
# Homebrew (planned)
brew install asd

# or build from source
go build -o asd .
```

First run sets up a provider interactively:

```
$ asd
Let's connect a provider:
  1) OpenAI
  2) Claude (Anthropic)
  3) DeepSeek
  4) OpenAI-compatible / local (custom URL)
> _
```

## Usage

```
asd <natural language request>     translate a request
asd config                         change provider/model/key

  -e, --explain     expressive, flag-by-flag explanation
  -q, --quiet       command only, no explanation
  -m, --model NAME  use a different model for this call
```

The command is printed and copied to your clipboard — paste it to run or edit.

You can name a different target in the request itself:

```
$ asd how do I list args in powershell
```

## Config

Lives at `~/.config/asd/config.toml`. See [DESIGN.md](DESIGN.md) for the full
schema and the reasoning behind every decision. The system prompt is plain text
at `internal/llm/prompt.txt` — edit it to taste.
