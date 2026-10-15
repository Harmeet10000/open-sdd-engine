# Research Notes: Autonomous Agent Task Execution Patterns

This document covers modern patterns and architectures for orchestrating autonomous AI agent task execution, compiled in June 2026.

---

## 1. Google OKF (Open Knowledge Format)

Google's **Open Knowledge Format (OKF)** (v0.1, published June 2026) is an open, vendor-neutral specification designed to standardize how organizational knowledge and codebase context are structured for AI agents.

### Core Structure & Schema
- **File-based & Git-friendly**: Avoids heavy database systems or RAG middlewares. All knowledge is stored as standard Markdown (`.md`) files in a repository directory.
- **YAML Frontmatter**: The only strictly mandatory field is `type`. Recommended fields include `title`, `description`, `resource` (URIs pointing to consoles, APIs, docs), `tags`, and `timestamp`.
- **Graph-based Backlinks**: Concepts are interconnected using standard relative Markdown links (e.g. `[Development Guide](development.md)`). By following these links, agents can traverse a structured knowledge graph.
- **Reserved Filenames**:
  - `index.md`: Standard index file for entry points and progressive disclosure.
  - `log.md`: Chronological log for updates and audit trails.

### Application to SDD (Spec-Driven Development)
- Enables scaffolded codebase guides (`.kiro/knowledge/`) to tell the agent about design patterns, linter settings, and system rules without bloat.
- Eliminates context confusion: the agent uses the knowledge graph to navigate domains rather than parsing the entire codebase dynamically on every turn.

---

## 2. Karpathy's AutoResearch Loop

Andrej Karpathy's **AutoResearch** is an autonomous experiment execution loop designed to iteratively propose, evaluate, and keep or discard machine learning experiments.

### Loop Architecture & Files
1. **`prepare.py` (Locked)**: Handles the ground-truth setup, download of datasets, tokenization, and metric calculation (such as bits-per-byte `val_bpb`). The agent cannot modify this, ensuring evaluation remains strict and uncorrupted.
2. **`train.py` (Writable)**: The primary file containing layers, learning rate, training parameters, etc. This is the playground for the agent to optimize.
3. **`program.md` (Agenda)**: Context and explicit instructions specifying goals and constraints.

### The Execution Cycle
- **Propose**: The agent proposes and makes a code tweak to `train.py`.
- **Run/Evaluate**: The system runs the script for a short duration and captures performance metrics.
- **Commit/Revert**: If performance improves, the system runs `git commit` to save the checkpoint. If it fails or degrades, it automatically runs `git checkout -- train.py` (revert) and tries a different proposal.

---

## 3. Loop Engineering

**Loop Engineering** focuses on building reliable, long-running agentic loops (*context → decide → tool → result → context*) that survive failures and scale horizontally.

### Key Best Practices
- **Triggers**: Hooks that launch agent loops based on schedules, repository commits, or events.
- **Decoupled Verifiers**: Hard-coded, deterministic check layers (e.g. static analyzers, compiler checks, syntax linters) that evaluate agent outputs before they are written or executed.
- **Termination/Stop Rules**: Clear conditions (budget thresholds, run durations, max token consumption, iteration limits) to prevent runaway execution or loop traps.

---

## 4. Dynamic Skill Selection

Moving from hardcoded tools to dynamic capability matching.

### Design Patterns
- **Skill Registries**: Reusable "skill" definitions stored with YAML metadata describing prerequisites and domains of applicability (similar to Model Cards).
- **Meta-Planners**: LLM classifiers that review the active task, check file extensions and directories, and query the registry to load only relevant tools into the active context window. This minimizes token consumption and prevents tool-choice confusion.

---

## 5. Long-Running Task Patterns

Primitives that enable agents to execute multi-hour architectures without losing progress.

### Design Patterns
- **Checkpoint & Resume (Durable Execution)**:
  - Persisting intermediate states (e.g., active task index, file list, verification state) in database rows or state files (like `.kiro/state.json`).
  - Upon network failure, rate-limits (like 429 errors), or memory crashes, the agent re-initializes and loads the state file to resume execution instantly.
- **Watchdog/Heartbeat**: External runner monitors the agent's workspace. If no updates or file modifications are written within a window (e.g., 5-10 minutes), the runner restarts or halts the agent.

---

## Source Material & Links
- [Google OKF Specification (knowledge-catalog)](https://github.com/google-cloud-platform/knowledge-catalog)
- [Karpathy AutoResearch Repository](https://github.com/karpathy/autoresearch)
- [LangGraph / Durable Execution frameworks](https://github.com/langchain-ai/langgraph)
