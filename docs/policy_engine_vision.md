# CostWise Policy Engine: Vision & Implementation Plan

## 1. Vision and Motivation
Currently, **CostWise-MCP** is an exceptionally powerful observation tool. It uses a Tree-Sitter AST engine to natively understand the codebase, extracting exact symbols, imports, call graphs, and references. However, it is passive; it tells you *what* the architecture is, but doesn't judge if it is *correct*.

Conversely, **RepoGraph** (as used in the CCAI project) is an active enforcer. It validates strict layer boundaries, calculates a health score (100/100), and fails CI builds if rules are broken. But it has a fatal flaw: it relies on developers writing manual `@repograph` YAML comments on every file. If the comments don't match the code, RepoGraph validates a lie.

**The Vision:** We will port the strict enforcement features of RepoGraph directly into CostWise-MCP. By doing this, we eliminate the need for manual annotations. CostWise will enforce architectural boundaries by comparing a single, centralized rule file against the *actual, compiled AST imports*. 

This makes CostWise the ultimate architectural gatekeeper: strict, automatic, and impossible to lie to.

---

## 2. Core Concepts

### A. Advanced Centralized Configuration (`costwise-architecture.yaml`)
RepoGraph was limited to simple layer-to-layer `depends_on`. CostWise-MCP, powered by its deep AST understanding, can support a much richer and more granular dependency mapping. The central configuration will support directional dependencies, third-party package restrictions, and module-level isolation.

```yaml
# costwise-architecture.yaml
version: 1

# 1. External Package Restrictions
banned_imports:
  Domain: 
    - "package:flutter/**"  # Domain must be pure Dart, no UI frameworks
    - "dart:html"

# 2. Granular Layer Definitions
layers:
  Presentation:
    path: lib/features/*/presentation/**
    can_depend_on: [Application, Domain]
    allow_exports: false # Cannot be imported by any other layer
  
  Application:
    path: lib/features/*/application/**
    can_depend_on: [Domain]
    
  Domain:
    path: lib/features/*/domain/**
    can_depend_on: [] # Absolute core, depends on nothing
    
  Datasource:
    path: lib/features/*/data/**
    can_depend_on: [Domain]

# 3. Cross-Feature Isolation (Modular Monolith Support)
features:
  path_pattern: lib/features/*
  isolated: true
  allow_cross_feature_imports_via: "*/api/**" # Features can only communicate via their `api` folders.
```

### B. AST-Backed Validation
CostWise already has a `SymbolDB` tracking all files and their imports. The Policy Engine will:
1. Map every file to a defined layer using the `path` globs.
2. Extract all imports and outgoing calls for that file.
3. Check if any import points to a layer not explicitly allowed in `can_depend_on`.

### C. Health Scoring & CI/CD Enforcement
CostWise will introduce a new CLI command (`costwise validate`). This command will calculate a health score (deducting points for each layer violation) and exit with code `1` if the score is below 100, acting as a strict CI/CD gate.

### D. MCP Tooling (`validate_architecture`)
AI Agents will have access to a new MCP tool. Before an agent completes a task, it can run `validate_architecture` to ensure its proposed changes don't violate the project's structural rules, allowing it to self-correct.

---

## 3. Detailed Implementation Plan (CostWise-MCP Repository)

### Phase 1: Policy Parser & Core Validator
**Target Package:** `internal/policy/`

*   **`config.go`**: Create a YAML parser for `costwise-architecture.yaml`.
*   **`validator.go`**: Create the core engine.
    *   **Input**: The parsed `Config` and the active `RepoSession` (giving access to `SymbolDB`).
    *   **Execution**: Iterate over all indexed files. Use the existing Tree-Sitter reference index to resolve external imports for each file.
    *   **Evaluation**: If `File A` (Layer: Presentation) imports `File B` (Layer: Datasource), flag a `LayerViolation`.
    *   **Output**: A detailed `ValidationReport` struct containing the Health Score and a list of violations with file/line numbers.

### Phase 2: CLI Integration
**Target Package:** `cmd/` & `internal/doctor/`

*   **`cmd/validate.go`**: Register a new Cobra command `costwise validate`.
    *   Accepts a `--policy` flag (defaulting to `costwise-architecture.yaml`).
    *   Calls the `validator.go` engine.
    *   Prints a beautiful terminal report (using colors) showing the violations and final Health Score.
    *   `os.Exit(1)` if there are violations.

### Phase 3: MCP Tool Integration
**Target Package:** `internal/mcpserver/`

*   **`tools.go`**: Register the `validate_architecture` tool.
    *   This tool will execute the validator and return the `ValidationReport` as a formatted string to the LLM.
*   **`internal/skill/policy.md`**: Update the `costwise-session` skill guidance. Add an instruction: *"Before concluding a feature implementation, run `validate_architecture` to ensure your new files and imports do not violate the repository's strict layer boundaries."*

### Phase 4: CCAI Migration (The Consumer)
Once the engine is shipped in CostWise-MCP v2.x, the consumer repository (CCAI) will migrate:
1.  **Delete RepoGraph:** Remove the `tool/repograph` directory.
2.  **Strip Annotations:** Delete all `/// @repograph` comments from the 37 Dart files.
3.  **Create Policy:** Write `costwise-architecture.yaml` at the root of CCAI.
4.  **Update CI:** Replace `dart run tool/repograph/bin/repograph.dart validate` with `costwise validate` in the GitHub Actions pipeline.

---

## 4. Verification Plan
*   **Unit Tests (`internal/policy/validator_test.go`)**: Create a mock `SymbolDB` with known illegal imports and assert that the validator correctly identifies them and calculates the proper health score reduction.
*   **End-to-End Test**: Run `costwise validate` against a test fixture repository to verify CLI output and exit codes.
*   **Agent Test**: Connect an agent to CostWise, introduce a bad import, and ask the agent to check the architecture. Verify the agent receives the error and fixes the import.
