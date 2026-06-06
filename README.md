CostAffective

Supercharge Claude Code, Cursor, OpenCode, Codex CLI, and Antigravity with Token-Efficient Repository Intelligence

⚡ Up to 45.9% fewer tokens
⚡ Up to 54.3% fewer exploration loops
⚡ Up to 42.1% fewer tool interactions
🔒 100% local

«Spend tokens on reasoning, not retrieval.»

CostAffective is a local-first MCP server that helps AI coding assistants understand large codebases with less context, fewer retrieval loops, and lower token usage.

Works with:

- Claude Code
- Cursor
- OpenCode
- Codex CLI
- Antigravity
- MCP-compatible clients

No cloud services.

No code uploads.

No external infrastructure.

Everything runs locally on your machine.

---

Benchmark Highlights

Continue OSS Repository

- 3,203 files
- 1,985 source files

CostAffective

- 4.7M tokens
- 89 API calls
- 43 exploration calls

Alternative Semantic Code Intelligence Benchmark

- 8.7M tokens
- 134 API calls
- 94 exploration calls

Observed Results

✅ 45.9% fewer tokens

✅ 54.3% fewer exploration loops

✅ 42.1% fewer tool interactions

✅ Comparable repository-analysis deliverables

Generated outputs included:

- Unit Catalog
- Integration Map
- Architecture Overview
- Benchmark Harness

Additional repository benchmarks are currently being expanded across multiple open-source projects.

---

Why CostAffective?

Modern coding agents spend a large portion of their context budget:

- repeatedly reading files
- rediscovering architecture
- searching the same symbols
- exploring identical code paths
- re-running retrieval loops

CostAffective builds repository intelligence locally and provides compressed, repository-aware context through MCP.

This allows coding assistants to spend more context on solving problems and less context on searching repositories.

---

What CostAffective Optimizes

Reduce Token Waste

Retrieve relevant context instead of repeatedly reading files.

Reduce Exploration Loops

Help coding assistants find the correct code paths faster.

Compress Context

Deliver smaller and more useful context windows.

Improve Repository Understanding

Provide architecture-aware repository intelligence.

Stay Local

No code leaves your machine.

---

Features

Repository Intelligence

- Repository indexing
- Symbol search
- Reference tracking
- Caller discovery
- Repository summaries
- Architecture overviews
- Cross-file navigation

Retrieval Optimization

- Context compression
- Budget-aware retrieval
- Result ranking
- Retrieval filtering
- Context reduction

MCP Integration

- Claude Code
- Cursor
- OpenCode
- Codex CLI
- Antigravity
- MCP-compatible clients

Installation Experience

- One-command setup
- Automatic client detection
- Automatic MCP configuration
- Self-diagnostics
- Repair mode

---

Installation

Install

go install github.com/okyashgajjar/costaffective-mcp@latest

Connect to All Supported Clients

costaffective install --all

The installer automatically:

- detects installed AI coding tools
- installs MCP configurations
- validates startup
- verifies installation

---

Quick Start

# Install
go install github.com/okyashgajjar/costaffective-mcp@latest

# Configure all supported clients
costaffective install --all

# Verify installation
costaffective doctor

---

Supported Clients

Client| Status
Claude Code| ✅
Cursor| ✅
OpenCode| ✅
Codex CLI| ✅
Antigravity| ✅
MCP-Compatible Clients| ✅

---

MCP Tools

search_code

Repository-aware code search.

Examples:

How does authentication work?

Explain the retrieval pipeline.

Where is repository indexing implemented?

find_symbol

Locate definitions.

Examples:

SearchCode

CompressForAnswerType

FilterResults

find_references

Find symbol usages across a repository.

find_callers

Find functions that invoke a specific function.

grep_code

Perform exact text searches.

get_repository_summary

Generate a repository overview including:

- major modules
- architecture structure
- repository organization

index_repository

Build or refresh repository indexes.

---

Commands

Command| Description
"costaffective install"| Interactive installation
"costaffective install --all"| Configure all detected clients
"costaffective install --target <client>"| Configure a specific client
"costaffective install --repair"| Repair binary and MCP configuration
"costaffective doctor"| Validate installation and MCP setup
"costaffective uninstall"| Remove MCP configurations
"costaffective serve"| Start MCP server

---

Doctor

Validate installation health.

costaffective doctor

Checks:

- binary installation
- permissions
- MCP configuration
- startup validation
- repository access

Example:

PASS Binary
PASS Permissions
PASS Claude Code
PASS Cursor
PASS MCP Startup

Status: READY

---

Local-First Architecture

Repository
    ↓
Indexing
    ↓
Repository Intelligence
    ↓
Context Compression
    ↓
MCP Server
    ↓
AI Coding Assistant

CostAffective focuses on delivering the smallest useful context while preserving repository understanding.

---

Philosophy

Most AI coding tools optimize for finding more code.

CostAffective optimizes for sending less code.

The goal is not to retrieve more files.

The goal is to retrieve the right context with the smallest useful token budget.

«Spend tokens on reasoning, not retrieval.»

---

Development

git clone https://github.com/okyashgajjar/costaffective-mcp.git

cd costaffective-mcp

go build ./...

go test ./...

---

License

MIT