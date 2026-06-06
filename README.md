# CostAffective

Code intelligence for AI coding tools — fully local, no cloud, one command.

CostAffective is an MCP server that gives Cursor, Claude Code, OpenCode, Codex CLI, and Antigravity deep codebase understanding without sending your code anywhere.

```bash
go install github.com/okyashgajjar/costaffective-mcp@latest
costaffective install --all    # connect to every AI tool you have
```

## Quick Start

```bash
# 1. Install the binary
go install github.com/okyashgajjar/costaffective-mcp@latest

# 2. Connect to your AI coding tools
costaffective install --all
```

The installer builds the binary, detects your installed AI tools, and writes the correct MCP configuration for each one.

**No Go installed?** Clone the repo and run `go build -o costaffective ./cmd/mycli/`, then copy the binary to your PATH.

## Usage

| Command | What it does |
|---------|-------------|
| `costaffective install` | Interactive install — build, detect, prompt, configure |
| `costaffective install --all` | Non-interactive: configure all detected tools |
| `costaffective install --target <name>` | Configure a specific tool (claude, cursor, opencode, codex, antigravity) |
| `costaffective install --repair` | Rebuild binary and fix all MCP configs |
| `costaffective doctor` | Run diagnostics — checks binary, PATH, MCP configs, startup |
| `costaffective uninstall` | Remove MCP configs from all configured tools |
| `costaffective serve` | Start the MCP stdio server (used internally by AI tools) |

## Doctor

```bash
costaffective doctor
```

Checks binary existence, permissions, PATH, MCP configs for each client, server startup, and repository state. All checks report PASS/WARN/FAIL with a summary status.

## MCP Tools

Once connected, your AI coding assistant can use these tools:

| Tool | Description |
|------|-------------|
| `search_code` | Ask questions about your code in plain English |
| `find_symbol` | Find where a function, class, or variable is defined |
| `find_references` | Find everywhere a function or variable is used |
| `find_callers` | Find what calls a specific function |
| `grep_code` | Search for text patterns in code |
| `get_repository_summary` | Get an overview of your project structure |
| `index_repository` | (Re)build the search index |

## Requirements

- **Go 1.21+** (to build; the binary has no runtime deps)
- **git** (for repository analysis)
- An MCP-compatible AI coding tool

## Uninstall

```bash
costaffective uninstall
rm ~/.local/bin/costaffective
```

## Development

```bash
git clone https://github.com/okyashgajjar/costaffective-mcp.git
cd costaffective-mcp
go build ./...
go test ./...
```
