# CostAffective - Code Intelligence MCP Server

CostAffective is a highly optimized Model Context Protocol (MCP) server that provides repository intelligence directly to LLM clients (like Cursor, Claude Code, Antigravity, OpenCode).

By pushing complex retrieval pipelines (like Auto Retrieval, Grep, Tree-Sitter AST Search, Reference/Caller tracking) and context compression down into a lightweight local server, CostAffective minimizes the token input provided to LLMs, reducing costs and preventing hallucinations.

## Features

- **Standard MCP Interface**: Instantly connectable to any MCP-compliant editor or agent.
- **Auto-Reindexing Watchdog**: Monitors your repository changes via `fsnotify` and incrementally reindexes them automatically in the background.
- **Answer-Type Budgets**: Classifies intent and compresses responses dynamically based on tight token budgets.
- **Triple-Layer Memory**: Session, Repository, and Discovery layers caching intermediate query results.
- **9 Intelligent Retrievers**: Naive, Grep, FTS5, Tree-Sitter, Architecture, and more.
- **Zero LLM Overhead**: The server doesn't execute LLM calls directly—it serves intelligent context *to* your LLMs via the MCP protocol.

## Installation

```bash
go build -o costaffective ./cmd/mycli/
# Optionally, move it to a location in your PATH
sudo mv costaffective /usr/local/bin/
```

## Setup

Add the server to your MCP client configuration (e.g., in Cursor or Claude desktop).

**Example Configuration (`mcp.json`):**

```json
{
  "mcpServers": {
    "costaffective": {
      "command": "costaffective",
      "args": ["serve"],
      "env": {}
    }
  }
}
```

## MCP Tools Available

CostAffective exposes the following tools to the LLM:

1. `search_code`: Search the repository with an intelligent query pipeline. Best for natural language questions, architecture, or general code search.
2. `find_symbol`: Find the exact definition location of a symbol (class, function, variable).
3. `find_references`: Find all references and usages of a symbol across the repository.
4. `find_callers`: Find all functions that call a specific function.
5. `grep_code`: Exact text search across the repository using ripgrep-like functionality.
6. `get_repository_summary`: Get a high-level overview of the entire repository (modules, languages, file counts).
7. `index_repository`: Manually trigger a re-index of the repository. (Usually unnecessary due to the file watcher).

## Testing

```bash
go test ./...
```

The testing suite contains unit tests, integration tests, and full system tests that ensure the server initializes properly over stdio without panics.
