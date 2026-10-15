import { mkdir, writeFile } from "node:fs/promises";
import { join } from "node:path";
import { configYamlTemplate } from "../core/config.js";
import { getOkfArchitectureTemplate, getOkfDevelopmentTemplate } from "../core/okf.js";

const PROPOSAL_TEMPLATE = `# Proposal: {{id}}

## Why

Describe the problem or opportunity (1-2 sentences).

## What Changes

- 

## Capabilities

### New Capabilities

- 

### Modified Capabilities

- 

## Impact

- 
`;

const DELTA_TEMPLATE = `# Delta Spec

## ADDED Requirements

### Requirement: Example capability

The system SHALL ...

#### Scenario: Happy path

- **WHEN** ...
- **THEN** ...

## MODIFIED Requirements

## REMOVED Requirements

## RENAMED Requirements
`;

const CHANGE_TASKS_TEMPLATE = `# Tasks

## 1. Implementation

- [ ] 1.1 First task
- [ ] 1.2 Second task
`;

const SPEC_REQUIREMENTS_TEMPLATE = `# Requirements Document

## Introduction

## Glossary

## Goal

Describe the user or business outcome.

## Requirements

### Requirement 1: Example

**User Story:** As a developer, I want ...

#### Acceptance Criteria

1. WHEN a condition occurs, THE SYSTEM SHALL respond
2. THE SYSTEM SHALL maintain consistency

## Edge Cases

- Describe invalid, risky, or boundary conditions.

## Non-Goals

- State explicitly out-of-scope behavior.

## Invariants

1. WHEN an adjacent unchanged condition occurs, THE SYSTEM SHALL CONTINUE TO preserve existing behavior.

## Verification

- List automated checks, targeted tests, and manual verification.
`;

const SPEC_DESIGN_TEMPLATE = `# Design Document

## Overview

## Architecture

\`\`\`mermaid
graph TD
    A[Client] --> B[Service]
\`\`\`

## File Structure

## Verification Strategy
`;

const SPEC_TASKS_TEMPLATE = `# Implementation Plan

## Tasks

- [ ] 1. Setup project structure
  - _Requirements: 1.1_

## Task Dependency Graph

\`\`\`json
{
  "tasks": [
    { "id": "1", "depends_on": [] }
  ]
}
\`\`\`
`;

const SPEC_BEHAVIOR_TEMPLATE = `Feature: Example capability

  Scenario: Happy path
    Given a configured system
    When a user performs the supported action
    Then the system SHALL complete the action
`;

export async function scaffoldInit(workspace: string, okf?: boolean): Promise<void> {
  const kiro = join(workspace, ".kiro");
  await mkdir(join(kiro, "domains"), { recursive: true });
  await mkdir(join(kiro, "changes", "archive"), { recursive: true });
  await mkdir(join(kiro, "specs"), { recursive: true });
  await mkdir(join(kiro, "steering"), { recursive: true });
  await writeFile(join(kiro, "config.yaml"), configYamlTemplate(), "utf-8");
  await writeFile(join(kiro, "steering", "product.md"), "# Product\n\nDescribe users, outcomes, and product constraints.\n", "utf-8");
  await writeFile(join(kiro, "steering", "tech.md"), "# Technical standards\n\nRecord languages, runtime, testing, linting, and deployment constraints.\n", "utf-8");
  await writeFile(join(kiro, "steering", "structure.md"), "# Repository structure\n\nRecord major modules, ownership, and dependency boundaries.\n", "utf-8");

  if (okf) {
    const knowledge = join(kiro, "knowledge");
    await mkdir(knowledge, { recursive: true });
    await writeFile(join(knowledge, "architecture.md"), getOkfArchitectureTemplate(), "utf-8");
    await writeFile(join(knowledge, "development.md"), getOkfDevelopmentTemplate(), "utf-8");
  }
}

export async function scaffoldChange(workspace: string, id: string): Promise<string> {
  const dir = join(workspace, ".kiro", "changes", id);
  await mkdir(join(dir, "specs", "example"), { recursive: true });
  await writeFile(
    join(dir, "change.meta.json"),
    JSON.stringify({ loop: "change", budgetHours: 2, status: "active" }, null, 2),
    "utf-8",
  );
  await writeFile(join(dir, "proposal.md"), PROPOSAL_TEMPLATE.replace(/\{\{id\}\}/g, id), "utf-8");
  await writeFile(join(dir, "specs", "example", "spec.md"), DELTA_TEMPLATE, "utf-8");
  await writeFile(join(dir, "tasks.md"), CHANGE_TASKS_TEMPLATE, "utf-8");
  return dir;
}

export async function scaffoldSpecify(workspace: string, slug: string): Promise<string> {
  const dir = join(workspace, ".kiro", "specs", slug);
  await mkdir(dir, { recursive: true });
  const metaPath = join(dir, "spec.meta.json");
  const { access } = await import("node:fs/promises");
  let exists = false;
  try {
    await access(metaPath);
    exists = true;
  } catch {}
  if (!exists) {
    await writeFile(
      metaPath,
      JSON.stringify({ loop: "spec", estimatedDays: 5, complexity: "high" }, null, 2),
      "utf-8",
    );
  }
  await writeFile(join(dir, "requirements.md"), SPEC_REQUIREMENTS_TEMPLATE, "utf-8");
  return dir;
}

export async function scaffoldPlan(workspace: string, slug: string): Promise<string> {
  const dir = join(workspace, ".kiro", "specs", slug);
  await mkdir(dir, { recursive: true });
  await writeFile(join(dir, "design.md"), SPEC_DESIGN_TEMPLATE, "utf-8");
  await writeFile(join(dir, "tasks.md"), SPEC_TASKS_TEMPLATE, "utf-8");
  await writeFile(join(dir, "behavior.feature"), SPEC_BEHAVIOR_TEMPLATE, "utf-8");
  return dir;
}

export async function scaffoldSpec(workspace: string, slug: string): Promise<string> {
  await scaffoldSpecify(workspace, slug);
  await scaffoldPlan(workspace, slug);
  return join(workspace, ".kiro", "specs", slug);
}

export async function scaffoldCursorSkills(workspace: string): Promise<void> {
  const changeSkill = `# SDD Change Loop

Apply change loop tasks from \`.kiro/changes/<id>/\`.

Research tools (both loops):
- **graphify query** — codebase architecture
- **ast-grep** — structural code search
- **firecrawl search** — web research
- **context7 MCP** — library documentation

Run \`sdd instructions <change-id> --json\` for next steps.
`;
  const specSkill = `# SDD Spec Loop

Execute spec loop with DAG waves via \`sdd compile-waves <slug>\`.

Research tools:
- **graphify query** — deep architecture mapping
- **ast-grep** — pattern inventory
- **firecrawl search** — external benchmarks and papers
- **context7 MCP** — framework docs

Run \`sdd instructions <slug> --json --type spec\` for next steps.
`;
  await mkdir(join(workspace, ".cursor", "skills", "sdd-change"), { recursive: true });
  await mkdir(join(workspace, ".cursor", "skills", "sdd-spec"), { recursive: true });
  await writeFile(join(workspace, ".cursor", "skills", "sdd-change", "SKILL.md"), changeSkill, "utf-8");
  await writeFile(join(workspace, ".cursor", "skills", "sdd-spec", "SKILL.md"), specSkill, "utf-8");
}
