# CostAffective — Code Intelligence MCP Server

CostAffective is an MCP server that gives AI coding tools (Cursor, Claude Code, OpenCode, etc.) deep understanding of your codebase — without sending your code to any cloud.

Just run one command to install and connect it to all your AI coding tools.

---

## Quick Start (30 seconds)

```bash
# 1. Get the binary
git clone https://github.com/your-org/costaffective.git
cd costaffective
go build -o costaffective ./cmd/mycli/

# 2. Install and connect to all your AI coding tools
./costaffective install --all
```

That's it. The installer will:
- Copy the binary to `~/.local/bin/costaffective`
- Detect which AI coding tools you have installed
- Configure each one with the right settings

---

## Installation

### Option 1: Automatic Install (Recommended)

After building the binary, run:

```bash
./costaffective install
```

This interactive command will:
1. Build and install the binary into your PATH
2. Detect your AI coding tools (Cursor, Claude Code, OpenCode, etc.)
3. Ask which ones to connect
4. Configure everything automatically

For a non-interactive install (configure all detected tools):

```bash
./costaffective install --all
```

### Option 2: Manual Install

#### Linux — Ubuntu / Debian

```bash
# 1. Build the binary
cd costaffective
go build -o costaffective ./cmd/mycli/

# 2. Install to PATH
mkdir -p ~/.local/bin
cp costaffective ~/.local/bin/
chmod +x ~/.local/bin/costaffective

# 3. Make sure ~/.local/bin is in your PATH
#    Add this line to ~/.bashrc or ~/.zshrc:
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc

# 4. Verify it works
costaffective --version
```

#### Linux — Arch Linux

```bash
# 1. Build
cd costaffective
go build -o costaffective ./cmd/mycli/

# 2. Install to PATH
mkdir -p ~/.local/bin
cp costaffective ~/.local/bin/
chmod +x ~/.local/bin/costaffective

# 3. Add to PATH (for bash)
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc

#    Or for fish shell:
#    fish_add_path ~/.local/bin

# 4. Verify
costaffective --version
```

#### Linux — Fedora

```bash
# 1. Build
cd costaffective
go build -o costaffective ./cmd/mycli/

# 2. Install to PATH
mkdir -p ~/.local/bin
cp costaffective ~/.local/bin/
chmod +x ~/.local/bin/costaffective

# 3. Add to PATH
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc

# 4. Verify
costaffective --version
```

#### macOS

```bash
# 1. Build
cd costaffective
go build -o costaffective ./cmd/mycli/

# 2. Install to PATH
mkdir -p ~/.local/bin
cp costaffective ~/.local/bin/
chmod +x ~/.local/bin/costaffective

# 3. Add to PATH (adds to ~/.zshrc automatically on macOS)
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc

# 4. Verify
costaffective --version
```

#### Windows

```powershell
# 1. Build
cd costaffective
go build -o costaffective.exe ./cmd/mycli/

# 2. Create a directory for it
mkdir "$env:USERPROFILE\.local\bin" -Force

# 3. Copy the binary
copy costaffective.exe "$env:USERPROFILE\.local\bin\"

# 4. Add to PATH (System-wide, restart terminal after)
[Environment]::SetEnvironmentVariable(
    "Path",
    [Environment]::GetEnvironmentVariable("Path", "User") + ";$env:USERPROFILE\.local\bin",
    "User"
)

# 5. Verify
costaffective --version
```

---

## Connect to AI Coding Tools

### Option A: One-Command Setup (Recommended)

```bash
costaffective install
```

This detects your installed AI tools and configures them automatically.

### Option B: Configure a specific tool

```bash
costaffective install --target cursor     # Only configure Cursor
costaffective install --target claude     # Only configure Claude Code
costaffective install --target opencode   # Only configure OpenCode
costaffective install --target codex      # Only configure Codex CLI
costaffective install --target antigravity  # Only configure Antigravity/Gemini
```

### Option C: Fix issues if something broke

```bash
costaffective install --repair
```

This rebuilds the binary and fixes all MCP configurations.

---

## Check That Everything Works

```bash
costaffective doctor
```

This runs 12 checks and shows:

```
PASS  Binary Found
PASS  Binary Permissions
PASS  Binary Version
PASS  Binary in PATH
PASS  Cursor Config
PASS  Claude Code Config
PASS  OpenCode Config
PASS  MCP Startup
PASS  Repository
PASS  Index Directory

Results: 10 PASS, 0 WARN, 0 FAIL
Status: READY
```

See something marked `FAIL`? Run `costaffective install --repair` to fix it.

> **Per-client details**: See `docs/mcp/` for configuration guides for each AI coding tool.

---

## Remove CostAffective

```bash
costaffective uninstall
```

This removes the MCP configuration from all your AI tools. To also delete the binary:

```bash
rm ~/.local/bin/costaffective
```

---

## MCP Tools

Once connected, your AI coding tool gets these commands:

| Tool | What it does |
|------|-------------|
| `search_code` | Ask questions about your code in plain English |
| `find_symbol` | Find where a function, class, or variable is defined |
| `find_references` | Find everywhere a function or variable is used |
| `find_callers` | Find what calls a specific function |
| `grep_code` | Search for text patterns in code |
| `get_repository_summary` | Get an overview of your project structure |
| `index_repository` | (Re)build the search index |

---

## Requirements

- **Go 1.21+** (needed only to build; the binary has no dependencies)
- **git** (for repository analysis)
- **An MCP-compatible AI coding tool** (Cursor, Claude Code, OpenCode, Antigravity, or Codex CLI)

---

## Production Features

| Feature | Command | What it does |
|---------|---------|-------------|
| Auto-install | `costaffective install` | Builds + installs + configures |
| Doctor | `costaffective doctor` | 12 checks, PASS/WARN/FAIL |
| Repair | `costaffective install --repair` | Fixes broken installs |
| Version | `costaffective --version` | Shows version |

---

## Testing

```bash
go test ./...
```
