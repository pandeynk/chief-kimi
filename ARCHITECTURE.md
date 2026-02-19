# Chief — Architecture Overview

This document describes the high-level architecture of **Chief**, an autonomous coding agent that drives Claude Code to implement user stories from a Product Requirements Document (PRD).

---

## Project Overview

Chief is a Go CLI tool that:

1. Accepts a project description as a set of user stories (a **PRD**).
2. Runs [Claude Code](https://docs.anthropic.com/en/docs/claude-code) in a loop, implementing one story per iteration.
3. Persists state between iterations so every new Claude session starts with full context.
4. Exposes a **Terminal UI (TUI)** for monitoring progress, managing multiple PRDs in parallel, and reviewing results.

The core pattern is the [Ralph Wiggum loop](https://ghuntley.com/ralph/): each Claude invocation has a fresh context window, but all progress is persisted to disk so nothing is lost between runs. In other words, rather than fighting context-window limits with one ever-growing conversation, Chief runs Claude in a tight loop — each iteration starts fresh, reads persisted state from the previous run, and picks up exactly where it left off.

---

## Repository Layout

```
chief/
├── cmd/chief/          # Entry point (main.go, CLI flag parsing, subcommand dispatch)
├── embed/              # Embedded prompt templates (compiled into the binary)
├── internal/
│   ├── cmd/            # Subcommand logic: new, edit, status, list
│   ├── config/         # Project-level configuration (.chief/config.yaml)
│   ├── git/            # Git helpers: branches, worktrees, push, diff, merge
│   ├── loop/           # Core agent loop: Loop, Manager, Parser, event types
│   ├── notify/         # Sound notifications on completion
│   ├── prd/            # PRD types, loader/saver, file watcher, md→json converter
│   └── tui/            # Bubble Tea TUI: App model, views, overlay components
├── docs/               # VitePress documentation site
├── .chief/             # Runtime state directory (not committed, per-project)
├── CHIEF_TUI_SPEC.md   # Detailed TUI design specification
└── WORKTREES_SPEC.md   # Worktree isolation design specification
```

---

## Component Breakdown

### `cmd/chief` — Entry Point

`main.go` is the single binary entry point. It:

- Dispatches CLI subcommands (`new`, `edit`, `status`, `list`, `help`, `--version`).
- Parses TUI-mode flags (`--max-iterations`, `--no-sound`, `--verbose`, `--merge`, `--force`, `--no-retry`).
- Auto-discovers a PRD if none is specified (looks for `.chief/prds/main/prd.json`, then any available PRD).
- Runs the **first-time setup TUI** when no PRD exists at all.
- Checks whether `prd.md` is newer than `prd.json` and triggers conversion if needed.
- Creates the `tui.App`, wires up sound notifications, and runs the Bubble Tea event loop.
- After TUI exit, handles post-exit actions (`PostExitInit`, `PostExitEdit`) by launching subcommands and restarting the TUI.

### `embed` — Prompt Templates

All Claude prompts are embedded in the binary at compile time using Go's `//go:embed` directive. There are five templates:

| Template | Purpose |
|---|---|
| `prompt.txt` | Main agent loop prompt sent to Claude each iteration |
| `init_prompt.txt` | PRD creation prompt used by `chief new` |
| `edit_prompt.txt` | PRD editing prompt used by `chief edit` |
| `convert_prompt.txt` | `prd.md` → `prd.json` conversion prompt |
| `detect_setup_prompt.txt` | Detects project setup commands (e.g., `npm install`) |

Placeholders like `{{PRD_PATH}}`, `{{PRD_DIR}}`, and `{{CONTEXT}}` are substituted at runtime by the `embed` package helpers.

### `internal/prd` — PRD Types and State

The PRD is the central data structure that drives the loop.

**Types** (`types.go`):
- `PRD` — project name, description, and a list of user stories.
- `UserStory` — ID, title, description, acceptance criteria, priority, `passes` (complete flag), and `inProgress` (in-flight flag).
- `AllComplete()` — returns `true` when every story has `passes: true`.
- `NextStory()` — returns the next story to work on: first any `inProgress: true` story (crash recovery), then the lowest-priority incomplete story.

**Loader** (`loader.go`): `LoadPRD` / `Save` read and write `prd.json`.

**Watcher** (`watcher.go`): Uses `fsnotify` to watch `prd.json` for file-system changes and notify the TUI to reload.

**Generator** (`generator.go`): Invokes Claude with the init or edit prompt to create/update `prd.json` interactively.

### `internal/loop` — Core Agent Loop

This is the heart of Chief.

**`Loop`** (`loop.go`): Manages a single PRD execution.
- Spawns `claude --dangerously-skip-permissions -p <prompt> --output-format stream-json --verbose` as a subprocess.
- Reads stdout line-by-line and forwards parsed `Event` objects to an unbuffered channel.
- Supports pause (waits for current iteration to finish), stop (kills the subprocess immediately), and automatic retry on crash (up to 3 attempts with configurable delays).
- After each iteration, re-reads `prd.json`; exits the loop when `AllComplete()` is true or max iterations is reached.

**`Manager`** (`manager.go`): Manages multiple `Loop` instances in parallel.
- Registers PRDs by name with optional worktree metadata (path + branch).
- Starts, pauses, and stops individual loops; forwards all events as `ManagerEvent{PRDName, Event}` on a single merged channel.
- Fires `onComplete` and `onPostComplete` callbacks (push, PR creation) when a PRD finishes.
- Exposes state snapshots (`GetAllInstances`, `GetRunningPRDs`) for the TUI.

**`Parser`** (`parser.go`): Converts raw Claude `stream-json` lines into typed `Event` structs.

| Event Type | Trigger |
|---|---|
| `EventIterationStart` | `{"type":"system","subtype":"init"}` |
| `EventAssistantText` | Text block in assistant message |
| `EventToolStart` | Tool use block in assistant message |
| `EventToolResult` | Tool result in user message |
| `EventStoryStarted` | `<ralph-status>` tag in assistant text |
| `EventComplete` | `<chief-complete/>` tag in assistant text |
| `EventMaxIterationsReached` | Loop iteration counter hits the limit |
| `EventError` | Subprocess error |
| `EventRetrying` | Crash recovery retry attempt |

### `internal/tui` — Terminal User Interface

Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) (Elm-architecture) and styled with [Lip Gloss](https://github.com/charmbracelet/lipgloss).

**`App`** (`app.go`): The root Bubble Tea model. Holds:
- The active PRD and its path.
- A `loop.Manager` for all registered PRDs.
- A `prd.Watcher` for live file reload.
- A `TabBar` for switching between multiple PRDs.
- Active view mode and all overlay/component state.

**Views** (one `ViewMode` constant per view):

| View | File | Purpose |
|---|---|---|
| `ViewDashboard` | `dashboard.go` | Main status screen: story list, iteration counter, activity log |
| `ViewLog` | `log.go` | Scrollable raw Claude output log |
| `ViewDiff` | `diff.go` | Git diff viewer (branch vs. default branch) |
| `ViewPicker` | `picker.go` | PRD picker: list, create, switch, merge, clean |
| `ViewHelp` | `help.go` | Keyboard shortcut reference overlay |
| `ViewBranchWarning` | `branch_warning.go` | Warning dialog when on a protected branch |
| `ViewWorktreeSpinner` | `worktree_spinner.go` | Step-by-step worktree setup progress |
| `ViewCompletion` | `completion.go` | Post-completion screen with push/PR status |
| `ViewSettings` | `settings.go` | Settings overlay (push, PR, setup command) |

**`TabBar`** (`tabbar.go`): Shows one tab per registered PRD; displays branch name when a worktree is active.

**Key bindings** (approximate):
- `s` — start / resume loop
- `p` — pause after current story
- `x` — stop immediately
- `l` — toggle log view
- `d` — toggle diff view
- `n` / `1–9` — switch PRD
- `m` — merge branch (in picker)
- `c` — clean worktree (in picker)
- `,` — open settings
- `?` — help overlay
- `q` / `Ctrl-C` — quit

### `internal/git` — Git Utilities

Thin wrappers around `git` CLI invocations:

| File | Responsibility |
|---|---|
| `git.go` | Branch detection, creation, existence check, diff generation |
| `worktree.go` | Create/remove/list/prune worktrees; detect orphaned worktrees |
| `push.go` | Push branch to origin; create PR via `gh` CLI |
| `gitignore.go` | Check whether `.chief/` is already in `.gitignore`; add it |

### `internal/config` — Project Configuration

Reads and writes `.chief/config.yaml`. Current settings:

```yaml
worktree:
  setup: "npm install"   # Command run after worktree creation
onComplete:
  push: true             # Auto-push branch when PRD completes
  createPR: true         # Auto-create GitHub PR on completion
```

### `internal/cmd` — Subcommand Implementations

| Subcommand | File | What it does |
|---|---|---|
| `chief new` | `new.go` | Creates `.chief/prds/<name>/` and runs the init prompt interactively via Claude |
| `chief edit` | `edit.go` | Opens an existing PRD for interactive editing via Claude |
| `chief status` | `status.go` | Prints story completion counts for a PRD |
| `chief list` | (in `status.go`) | Lists all PRDs with their completion progress |

### `internal/notify` — Sound Notifications

Plays a short completion sound when a PRD finishes, using the `hajimehoshi/oto` audio library. Gracefully degrades (logs a warning) if audio initialization fails.

---

## Data Flow

```
  User
   │  chief [options] [prd-name]
   ▼
cmd/chief/main.go
   │  discovers prd.json
   │  creates tui.App
   ▼
tui.App  ─── prd.Watcher ──► prd.json (disk)
   │                           ▲
   │  s (start)                │  writes passes:true
   ▼                           │
loop.Manager
   │  Start(prdName)
   ▼
loop.Loop
   │  exec claude --output-format stream-json
   ▼
Claude subprocess
   │  stdout (stream-json lines)
   ▼
loop.Parser.ParseLine()
   │  returns typed Event
   ▼
loop.Loop.events channel
   ▼
loop.Manager.events channel  (ManagerEvent)
   ▼
tui.App.Update()  ─── renders ──► terminal
```

---

## State Persistence

Chief persists all state to `.chief/` in the project root:

```
.chief/
├── config.yaml           # Project-level settings
└── prds/
    └── <prd-name>/
        ├── prd.json       # Story list with passes/inProgress flags
        ├── prd.md         # Human-editable source (optional; converts to prd.json)
        ├── progress.md    # Running log written by Claude after each iteration
        └── claude.log     # Raw Claude subprocess output
```

The loop is fully resumable: stopping Chief and restarting it picks up from `NextStory()`, which reads `prd.json` to find the next incomplete story.

---

## Parallel PRD Execution and Worktrees

Each PRD can run in an isolated [git worktree](https://git-scm.com/docs/git-worktree):

1. A branch is created from the default branch (e.g., `chief/auth-system`).
2. A worktree is checked out at `.chief/worktrees/<prd-name>/`.
3. Claude runs with its working directory set to the worktree, so all file operations and commits go to that branch.
4. On completion, the branch can be pushed and/or a PR opened automatically.

This prevents multiple concurrent Claude instances from interfering with each other.

---

## Build and Test

```bash
# Build
go build ./cmd/chief/

# Run all tests
go test ./...

# Run tests for a specific package
go test ./internal/loop/...
```

There are no external test dependencies; all tests use the standard `testing` package.
