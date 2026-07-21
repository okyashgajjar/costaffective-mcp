# CostWise MCP Cross-Platform E2E Test Suite

This directory contains the cross-platform end-to-end (E2E) test suite for the **CostWise MCP Server**.

It tests the full MCP protocol stack (JSON-RPC stdio initialization, handshake, capabilities, tool list discovery) and exercises all 12 MCP tools against a sample indexed code repository.

---

## Supported Operating Systems
- **Linux** (Ubuntu, Debian, Fedora, Arch, Alpine, etc.)
- **macOS** (Intel & Apple Silicon)
- **Windows** (Native PowerShell, CMD, or WSL)

---

## Prerequisites
- **Python 3.7+** (built-in standard libraries only; no external `pip` packages required).
- The compiled `costwise` (or `costwise.exe`) binary.

---

## Running the Tests

### Quick Run (Auto-detects binary in `~/.local/bin` or PATH)

#### Linux / macOS:
```bash
python3 tests/e2e/run_e2e_tests.py
```

#### Windows (PowerShell / CMD):
```powershell
python tests/e2e/run_e2e_tests.py
```

---

### Specifying a Custom Binary Path

If `costwise` is built in a custom path or build directory:

```bash
python3 tests/e2e/run_e2e_tests.py --binary=/path/to/costwise
```

On Windows:
```powershell
python tests/e2e/run_e2e_tests.py --binary=C:\Users\ysmak\.local\bin\costwise.exe
```

---

## Test Coverage
1. **JSON-RPC Handshake & Protocol Capabilities**: Verifies `initialize` and `notifications/initialized`.
2. **MCP Tool Discovery**: Asserts all 12 tools are registered (`search_code`, `find_symbol`, `read_symbol`, `find_references`, `find_callers`, `get_repository_summary`, `index_repository`, `remember`, `stash_context`, `recall`, `allow_dir`, `session_brief`).
3. **Repository Indexing**: Verifies auto-indexing and symbol extraction.
4. **Symbol Search & Reading**: Exercises `find_symbol`, `read_symbol`, and `search_code`.
5. **Session Memory & Stash**: Tests `remember`, `recall`, `stash_context`, and `session_brief`.
