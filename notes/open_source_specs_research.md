# Research Notes: Open Source Specifications & Kiro Spec Format

This document summarizes findings on the architectures of Fission-AI/OpenSpec, GitHub Spec Kit, and Kiro specs, compiled in June 2026.

---

## 1. OpenSpec (by Fission-AI)

OpenSpec is a framework that standardizes Spec-Driven Development (SDD) for AI agents by storing living specification files directly inside repositories.

### Architecture Overview
- **`openspec/specs/`**: A dictionary containing living specification Markdown files organized by capability or domain. Acts as the permanent source of truth.
- **`openspec/changes/`**: A working area where delta changes are drafted.
- **`openspec/config.yaml`**: Standard configuration file specifying active schemas, project-wide rules (such as tech stack constraints), and validations.

### Key Parser & Validation Patterns
- **Delta Spec Parsing**: Standardizes change logs with blocks under headers `## ADDED Requirements`, `## MODIFIED Requirements`, `## REMOVED Requirements`, `## RENAMED Requirements`.
- **Requirement Blocks**: Consist of `### Requirement: [name]` followed by scenario lists `#### Scenario: [scenario name]` containing Given/When/Then.
- **Archive Pipeline**: Automates merging delta specs back into the master specs under `openspec/specs/` upon completion, keeping history in Git version control.

---

## 2. Spec Kit (by GitHub)

GitHub Spec Kit focuses on repo-native workflows for AI coding agents, treating specifications as executable steps rather than static documents.

### Architecture Overview
- Organizes workflows using slash commands that match developer intent:
  - `/constitution`: Outlines project-wide non-negotiable standards.
  - `/specify`: Defines feature requirements (the "what" and "why").
  - `/plan`: Outlines architectural design and implementation strategies (the "how").
  - `/tasks`: Breaks down implementation into ordered task lists.
  - `/implement`: Triggers agents to write code fulfilling the tasks.

### Integration Approach
- Integrates with standard agent workflows (Cursor, Claude Code, GitHub Copilot).
- Ensures that code generation is driven strictly by specifications rather than interactive agent "vibe coding".

---

## 3. Kiro Spec Format

The Kiro specification standard is designed for complex, long-running greenfield designs.

### Core Structure
Within `.kiro/specs/<feature-slug>/`:
1. **`spec.meta.json`**: Metainfo like estimated complexity, target duration, and loop kind.
2. **`requirements.md`**: Captures specifications using **EARS** (Easy Approach to Requirements Syntax).
3. **`design.md`**: Technical design describing high-level architecture with Mermaid charts, file structure plans, and verification plans.
4. **`tasks.md`**: Checklist of implementation tasks mapping back to requirements using traceability references (e.g. `_Requirements: 1.1_`) and an optional topological dependency graph in JSON.

### EARS Requirement Patterns
- **Ubiquitous**: `<Actor> SHALL ...` (e.g., "The system SHALL encrypt tokens")
- **Event-driven**: `WHEN <trigger> the system SHALL ...` or `WHEN <trigger> THEN <system> SHALL ...`
- **State-driven**: `WHILE <state> the system SHALL ...`
- **Conditional**: `IF <condition> the system SHALL ...`
- **Complex**: Combining conditions, triggers, and states (e.g., `WHEN <trigger> WHILE <state> IF <condition> the system SHALL ...`)
