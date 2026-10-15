# Detailed Research Report: Modern Autonomous Agent Task Execution and Spec-Driven Architectures

This report details state-of-the-art patterns for orchestrating autonomous AI agents, standardizing codebase context (Google OKF), executing automated experiment loops (Karpathy's AutoResearch), and managing specification-driven development (OpenSpec, GitHub Spec Kit, Kiro), compiled in June 2026.

---

## Part 1: Modern Autonomous Agent Execution Patterns

### 1. Google OKF (Open Knowledge Format)
Google published **Open Knowledge Format (OKF) v0.1** (June 12, 2026, authored by Sam McVeety and Amir Hormati) to standardize how organizational and codebase knowledge is structured for both humans and AI agents. It addresses the "fragmented context problem" where internal documentation is scattered across wikis, code comments, and shared files.

*   **Intentionally Minimalist**: An OKF bundle is simply a directory of standard Markdown files. It avoids database registries and proprietary SDKs.
*   **YAML Frontmatter**: Every concept file must start with a YAML block. The only mandatory field is `type` (e.g. `type: BigQuery Table`, `type: API Endpoint`, `type: Metric`, `type: Playbook`). Recommended fields include `title`, `description`, `resource` (URI of the asset), `tags`, and `timestamp` (ISO 8601).
*   **Graph-based Backlinks**: Markdown files link to each other using relative links (e.g., `[Database Guidelines](database.md)`). Consumers treat these links as directed edges of an untyped relationship, generating a traversable knowledge graph.
*   **Reserved Filenames**:
    *   `index.md`: Enumerates directory contents for progressive disclosure.
    *   `log.md`: Stores update histories using date-grouped headers (`YYYY-MM-DD`).
*   **Permissive Consumption**: Consumers are instructed to tolerate broken links (representing written-but-not-yet-realized knowledge) and unknown metadata fields.

### 2. Karpathy's AutoResearch Pattern
Andrej Karpathy's **AutoResearch** loop establishes an autonomous machine learning experimentation cycle that prevents coding agents from drifting away from their goals.

*   **Three-File Architecture**:
    1.  **`prepare.py` (Locked)**: Installs dependencies, sets up the dataset, and calculates evaluation metrics (like validation bits-per-byte `val_bpb`). The agent cannot modify this, ensuring the ground truth for success remains deterministic.
    2.  **`train.py` (Writable)**: The script containing model parameters, optimizers, and architectural loops that the agent is instructed to modify.
    3.  **`program.md` (Agenda)**: Context, experimental constraints, and high-level goals.
*   **Propose-Train-Evaluate-Revert Loop**:
    *   **Propose**: The agent reads the results of previous iterations and modifies `train.py`.
    *   **Train/Evaluate**: The system runs a short training session (e.g., 5 minutes) and evaluates the metric.
    *   **Keep/Revert**: If the metric improves, changes are committed via `git`. If they fail or degrade performance, the code is rolled back using `git checkout -- train.py` and the agent proposes a different angle.

### 3. Loop Engineering
Loop engineering is the discipline of building reliable, long-running agentic loops (*context → decide → tool → result → context*) that survive failures and scale horizontally.

*   **Durable Execution (Checkpoint/Resume)**: Agents persist their intermediate memory, task indices, and execution logs to a database or local state file (e.g., `.kiro/state.json`). If the agent encounters rate limits (429 errors), network timeouts, or runtime crashes, it loads the checkpoint on reboot and resumes from the last completed subtask.
*   **Decoupled Verifiers**: Hard-coded, deterministic check routines (unit test suites, static analysis, linter checks) run independently of LLM reasoning, validating the code before it is committed.
*   **Stop Rules & Watchdogs**: Loops must have strict boundary caps (e.g., time limits, failure counters, token budgets) to prevent infinite loops when the agent gets stuck. External runners monitor modification timestamps to restart hanging agents.

### 4. Dynamic Skill Selection
Hard-coded tool lists cause agent prompt bloat and decision confusion. Dynamic selection addresses this by introducing registries:

*   **Skill Registries**: Reusable "skill" definitions stored with YAML metadata defining prerequisites and domains of applicability (similar to Model Cards).
*   **Context-Aware Routing**: A meta-planner analyzes the active task phase (e.g., proposal, tasks, implementation) and folder layout to dynamically bind only the required tools to the agent's active system prompt.

---

## Part 2: Open-Source Spec-Driven Development Frameworks

### 1. OpenSpec (by Fission-AI)
OpenSpec targets **brownfield/quick changes (≤ 2 hours)** by storing specifications directly inside the repository using Git-versioned folders.

*   **Living Specs Layout**: Uses a central `openspec/specs/` directory representing the "source of truth".
*   **Changes Directory**: Active tasks are worked on under `openspec/changes/<change-name>/`, grouping all related assets:
    *   `proposal.md` (why, what, capabilities)
    *   `design.md` (technical design)
    *   `tasks.md` (linear checklist)
    *   `specs/` (delta spec files outlining updates to existing domains)
*   **Merger/Archive Pipeline**: When a change is completed, running `openspec archive` parses the delta spec blocks (`## ADDED|MODIFIED|REMOVED Requirements`) and merges them directly back into the main spec files in `openspec/specs/`, maintaining a clean Git history.

### 2. GitHub Spec Kit
Spec Kit focuses on product-scenario driven workflows, using CLI tools (`specify-cli`) to generate predictable code from specs.

*   **Development Phases (Slash Commands)**:
    *   `/speckit.constitution`: Establishes governing principles (rules, UX, performance).
    *   `/speckit.specify`: Captures requirements and user stories (what/why).
    *   `/speckit.plan`: Sets technical stacks and files structure (how).
    *   `/speckit.tasks`: Compiles plans into ordered checklists.
    *   `/speckit.implement`: Instructs the agent to implement code for each task.
    *   `/speckit.converge`: Compares implementation with specifications to track remaining work.
*   **Extensions & Presets**: Features templates that can be customized locally (`.specify/templates/overrides/`) or overridden via preset bundles.

### 3. Kiro Specification Format
Kiro is optimized for **deep greenfield architectures (> 2 hours)**. It has a strict namespace structure:

*   **Directory Names**: `.kiro/specs/<feature-slug>/`
*   **Abstractions**:
    *   `spec.meta.json`: Estimated duration, complexity, loop kind.
    *   `requirements.md`: Formulated using **EARS** syntax (triggers, constraints, actors).
    *   `design.md`: Mermaid sequence/flow charts, verification plans, file structure plans.
    *   `tasks.md`: Checklist of implementation tasks mapping back to requirements using traceability references (e.g. `_Requirements: 1.1_`) and an optional topological dependency graph in JSON format to support parallel compilation.

### EARS Requirement Patterns
- **Ubiquitous**: `<Actor> SHALL ...` (e.g., "The system SHALL encrypt tokens")
- **Event-driven**: `WHEN <trigger> the system SHALL ...` or `WHEN <trigger> THEN <system> SHALL ...`
- **State-driven**: `WHILE <state> the system SHALL ...`
- **Conditional**: `IF <condition> the system SHALL ...`
- **Complex**: Combining conditions, triggers, and states (e.g., `WHEN <trigger> WHILE <state> IF <condition> the system SHALL ...`)

---

## Part 3: Comparative Analysis

| Feature | OpenSpec | GitHub Spec Kit | Kiro |
| :--- | :--- | :--- | :--- |
| **Best For** | Brownfield changes, quick deltas (≤ 2h) | Scenario-driven workflow validation | Greenfield, complex architectures (> 2h) |
| **Task Model** | Flat, linear checklists (`tasks.md`) | Interactive slash command sequences | Hierarchical DAG waves with parallel compilation |
| **Requirements Format** | Delta blocks (`ADDED`/`MODIFIED`) | Natural language specifications | EARS (Ubiquitous, Event, State, Conditional) |
| **Traceability** | Domain specs merged in `domains/` | Context folders | Explicit task-requirement links (`_Requirements: N.M_`) |
| **Integration** | CLI + Cursor custom commands | `specify-cli` + preset templates | DAG Wave compiler |

---

## Part 4: Recommendations for `sdd` Engine Integration

Our implementation of `sdd` bridges the best of these worlds:
1.  **Google OKF integration**: Scaffolds `.kiro/knowledge/` bundles automatically and runs link validations to verify documentation health.
2.  **Karpathy's AutoResearch Integration**: Enables agents to run benchmarks/tests with `sdd autoresearch run --revert-on-fail`, automatically rolling back code modifications if compilation or tests fail.
3.  **Spec-Kit and OpenSpec Alignment**: Provides `sdd specify` (matching Spec-Kit's `/specify`) and `sdd plan` (matching `/plan`), while retaining OpenSpec's delta spec and archive merging capabilities.
4.  **Loop State Persistence**: Utilizes `.kiro/state.json` to store loop history and guide agents dynamically through context-aware instructions.
