# Chief Kimi

Build big projects with AI CLIs like Claude or Kimi. Chief breaks your work into tasks and runs the AI in a loop until they're done.

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

## Using with Kimi

Chief uses the `CHIEF_CLI` environment variable to select which AI CLI to invoke (default: `claude`). To use Kimi instead of Claude, set `CHIEF_CLI=kimi` before running any Chief command.

### Prerequisites

- **Kimi CLI** installed and authenticated. Install it with:

  ```bash
  npm install -g @moonshot-ai/kimi-coder
  kimi auth login
  ```

  Verify the installation:

  ```bash
  kimi --version
  ```

### Running Chief with Kimi

```bash
# Set the CLI for the current shell session
export CHIEF_CLI=kimi

# Create a new project (uses Kimi interactively)
chief new

# Launch the TUI and press 's' to start
chief
```

Or set it inline for a single command:

```bash
CHIEF_CLI=kimi chief new
CHIEF_CLI=kimi chief
```

To make the setting permanent, add the export to your shell profile (`~/.bashrc`, `~/.zshrc`, etc.):

```bash
echo 'export CHIEF_CLI=kimi' >> ~/.zshrc
```

## Usage

```bash
# Create a new project
chief new

# Launch the TUI and press 's' to start
chief
```

Chief runs the AI in a [Ralph Wiggum loop](https://ghuntley.com/ralph/): each iteration starts with a fresh context window, but progress is persisted between runs. This lets the AI work through large projects without hitting context limits.

## How It Works

1. **Describe your project** as a series of tasks
2. **Chief runs the AI** in a loop, one task at a time
3. **One commit per task** — clean git history, easy to review

See the [documentation](https://minicodemonkey.github.io/chief/concepts/how-it-works) for details.

## Requirements

- **[Claude Code CLI](https://docs.anthropic.com/en/docs/claude-code)** installed and authenticated (default), **or**
- **[Kimi CLI](https://www.moonshot.cn/)** installed and authenticated (set `CHIEF_CLI=kimi`)

## License

MIT

## Acknowledgments

- [snarktank/ralph](https://github.com/snarktank/ralph) — The original Ralph implementation that inspired this project
- [Geoffrey Huntley](https://ghuntley.com/ralph/) — For coining the "Ralph Wiggum loop" pattern
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) — TUI framework
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) — Terminal styling
