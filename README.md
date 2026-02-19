# Chief

Build big projects with Claude or Kimi. Chief breaks your work into tasks and runs your AI agent CLI in a loop until they're done.

**[Documentation](https://minicodemonkey.github.io/chief/)** · **[Quick Start](https://minicodemonkey.github.io/chief/guide/quick-start)**

![Chief TUI](https://minicodemonkey.github.io/chief/images/tui-screenshot.png)

## Install

```bash
brew install minicodemonkey/chief/chief
```

Or via install script:

```bash
curl -fsSL https://raw.githubusercontent.com/MiniCodeMonkey/chief/refs/heads/main/install.sh | sh
```

## Usage

```bash
# Create a new project
chief new

# Launch the TUI and press 's' to start
chief
```

Chief runs your AI agent in a [Ralph Wiggum loop](https://ghuntley.com/ralph/): each iteration starts with a fresh context window, but progress is persisted between runs. This lets the agent work through large projects without hitting context limits.

## How It Works

1. **Describe your project** as a series of tasks
2. **Chief runs your AI agent** in a loop, one task at a time
3. **One commit per task** — clean git history, easy to review

See the [documentation](https://minicodemonkey.github.io/chief/concepts/how-it-works) for details.

## Agent Configuration

Chief defaults to using the [Claude Code CLI](https://docs.anthropic.com/en/docs/claude-code). To use [Kimi CLI](https://github.com/MoonshotAI/moonshot-kimi) instead, add an `agent` section to `.chief/config.yaml`:

```yaml
agent:
  provider: kimi
```

You can also change the provider interactively via the **Settings** overlay in the TUI (press `c` to open settings).

Supported providers:
- `claude` (default) — Claude Code CLI by Anthropic
- `kimi` — Kimi CLI by Moonshot AI

## Requirements

- An AI agent CLI installed and authenticated:
  - [Claude Code CLI](https://docs.anthropic.com/en/docs/claude-code), **or**
  - [Kimi CLI](https://github.com/MoonshotAI/moonshot-kimi)

## License

MIT

## Acknowledgments

- [snarktank/ralph](https://github.com/snarktank/ralph) — The original Ralph implementation that inspired this project
- [Geoffrey Huntley](https://ghuntley.com/ralph/) — For coining the "Ralph Wiggum loop" pattern
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) — TUI framework
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) — Terminal styling
