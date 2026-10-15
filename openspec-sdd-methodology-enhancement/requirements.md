# Requirements Document

## Introduction

This feature enhances the OpenSpec Spec-Driven Development (SDD) methodology with a subagent-native architecture. The enhancement introduces specialized agents that coordinate via a shared message board, enabling autonomous requirements analysis, RFC-style design generation, hostile QA audits, and git integration while maintaining compatibility with existing OpenSpec CLI workflows.

Key architectural decisions:

- **Decentralized Orchestration**: No monolithic orchestrator; specialized subagents coordinate via message board
- **Per-Agent Configuration**: Each subagent has configurable system prompts, models, permissions, tools, MCP servers, and skills
- **Four Specialist Agents**: Searcher (read-only), Implementer (write-only), Verifier (tests only), Reviewer (semantic review)
- **Brevity Protocol**: Matt Pocock-style concise responses enforced in all system prompts
- **Worktree Isolation**: Each workflow runs in its own git worktree for parallel execution
- **ADR Workflow**: Architecture Decision Records generated automatically for all decisions
- **Incident-Why-Outcome Comments**: All rules annotated with incident context and evidence

## Glossary

- **Message Board**: Central coordination layer providing shared state, async messaging, agent registration, and status tracking.
- **Subagent**: Specialized autonomous agent with dedicated responsibilities (Searcher, Implementer, Verifier, Reviewer, Refinement Coordinator).
- **Searcher Agent**: Read-only discovery agent that searches codebase, fetches external docs, and produces structured JSON findings.
- **Implementer Agent**: Write-only implementation agent that consumes SearchResults and implements tasks. Worktree-isolated.
- **Verifier Agent**: Test/lint/type-check runner that produces pass/fail verdicts. Never edits files.
- **Reviewer Agent**: Semantic code reviewer that produces structured reviews linked to requirement IDs.
- **System Prompt**: Custom instructions injected into a subagent's context at invocation time.
- **Permission Model**: Allowlist of file paths, tools, and operations each subagent can access.
- **Worktree**: Isolated git working directory for parallel workflow execution.
- **ADR**: Architecture Decision Record documenting context, decision, consequences, and alternatives.
- **Brevity Protocol**: Response contract enforcing ≤3 sentence answers, no preamble/recap, structured outputs preferred.
- **Incident-Why-Outcome**: Comment framework documenting incident, mechanism, evidence, and commit reference.
- **EARS Methodology**: Requirements specification methodology with six patterns: Ubiquitous, Event-driven, State-driven, Unwanted behavior, Optional features, Complex.
- **RFC-Style Design**: Standardized design document format with 7 required sections.

## Requirements

### Section 1: Subagent Configuration

### Requirement 1: Per-Subagent System Prompt Injection

**User Story:** As a workflow designer, I want each subagent to have configurable system prompts, so that I can customize agent behavior for different methodologies.

#### Acceptance Criteria

1. THE System SHALL support per-subagent system prompt configuration
2. WHEN a subagent is invoked, THE System SHALL inject the configured system prompt into the agent context
3. THE System Prompt SHALL be loaded from the subagent configuration file
4. WHERE no custom system prompt is defined, THE System SHALL use the default prompt

#### Scenario: System Prompt Loading

WHEN the Implementer agent is invoked for feature user-auth
THEN the system SHALL load the system prompt from config.agents.implementer.systemPrompt

#### Correctness Property

For any subagent invocation, the configured system prompt SHALL be injected before any other context.

**Validates:** Requirement 1
### Requirement 2: Per-Subagent Model Selection

**User Story:** As a workflow operator, I want to configure which LLM model each subagent uses, so that I can optimize cost and capability.

#### Acceptance Criteria

1. THE System SHALL support per-subagent model selection
2. THE Model Configuration SHALL include: primary model, fallback model, temperature, max tokens
3. WHEN the primary model is unavailable, THE System SHALL use the fallback model
4. WHERE no model is configured, THE System SHALL use the default model

#### Scenario: Model Fallback

WHEN the configured primary model is unavailable
THEN the system SHALL use the fallback model

#### Correctness Property

For any subagent, if the primary model fails, the fallback model SHALL be used within 5 seconds.

**Validates:** Requirement 2

### Requirement 3: Per-Subagent Permission Model

**User Story:** As a security engineer, I want each subagent to have an allowlist of permitted operations, so that agents cannot perform unauthorized actions.

#### Acceptance Criteria

1. THE System SHALL enforce per-subagent permission allowlists
2. THE Permission Model SHALL include: allowed file paths, allowed tools, allowed bash commands
3. WHEN an agent attempts an operation, THE System SHALL check against the allowlist
4. IF the operation is not in the allowlist, THE System SHALL reject it with PermissionDeniedError

#### Scenario: Permission Denial

WHEN the Searcher agent attempts to write to a file
THEN the system SHALL reject the operation with PermissionDeniedError

#### Correctness Property

For any operation by any subagent, the operation SHALL be rejected if not in the allowlist.

**Validates:** Requirement 3

### Requirement 4: Per-Subagent Tool Bindings

**User Story:** As a workflow optimizer, I want to configure which tools each subagent can access, so that agents only load necessary tools.

#### Acceptance Criteria

1. THE System SHALL support per-subagent tool bindings
2. WHEN an agent is invoked, THE System SHALL load only the bound tools
3. WHERE no tools are configured, THE System SHALL use the default tool set for that agent type

#### Scenario: Tool Loading

WHEN the Verifier agent is invoked
THEN the system SHALL load only: ruff, ty, pytest, ast-grep

#### Correctness Property

For any subagent invocation, only the bound tools SHALL be available.

**Validates:** Requirement 4

### Requirement 5: Per-Subagent MCP Server Bindings

**User Story:** As a workflow designer, I want to bind MCP servers to specific subagents, so that agents have access to required external tools.

#### Acceptance Criteria

1. THE System SHALL support per-subagent MCP server bindings
2. WHEN an agent is invoked, THE System SHALL connect to bound MCP servers
3. WHERE an MCP server is unavailable, THE System SHALL log a warning and proceed with degraded functionality

#### Scenario: MCP Connection

WHEN the Searcher agent is invoked
THEN the system SHALL connect to: Context7, Firecrawl, codegraph

#### Correctness Property

For any subagent, bound MCP servers SHALL be connected before agent execution starts.

**Validates:** Requirement 5

### Requirement 6: Per-Subagent Skill Bindings

**User Story:** As a workflow designer, I want to bind named skills to subagents, so that agents use appropriate specialized capabilities.

#### Acceptance Criteria

1. THE System SHALL support per-subagent skill bindings
2. WHEN an agent is invoked, THE System SHALL load the bound skills
3. WHERE a skill is unavailable, THE System SHALL log a warning and proceed without that skill

#### Scenario: Skill Loading

WHEN the Searcher agent is invoked
THEN the system SHALL load: orient, graphify, firecrawl-search

#### Correctness Property

For any subagent, bound skills SHALL be loaded before agent execution starts.

**Validates:** Requirement 6
### Section 2: Core Specialist Agents

### Requirement 7: Searcher Agent - Read-Only Discovery

**User Story:** As a workflow orchestrator, I want a read-only discovery agent, so that the search phase never mutates repository state.

#### Acceptance Criteria

1. THE Searcher Agent SHALL operate in read-only mode
2. THE Searcher Agent SHALL NOT have access to file write operations
3. WHEN invoked, THE Searcher Agent SHALL use: rg, ast-grep, graphify query, codegraph_explore, Context7 MCP, Firecrawl
4. THE Searcher Agent Output SHALL be structured JSON with search_id, findings, and sources
5. WHERE no findings are produced, THE Searcher Agent SHALL return an empty findings array with a reason

#### Scenario: Search Execution

WHEN the Searcher agent is invoked with query authentication flow
THEN the agent SHALL search using bound tools and MCP servers

#### Correctness Property

For any Searcher agent invocation, no file write operations SHALL be performed.

**Validates:** Requirement 7

### Requirement 8: Implementer Agent - Worktree-Isolated Implementation

**User Story:** As a workflow orchestrator, I want a write-only implementation agent isolated to a worktree, so that implementation cannot corrupt other branches.

#### Acceptance Criteria

1. THE Implementer Agent SHALL operate only within its assigned worktree path
2. THE Implementer Agent SHALL NOT have access to operations outside the worktree
3. WHEN invoked, THE Implementer Agent SHALL consume SearchResult from the message board
4. THE Implementer Agent SHALL implement one task per invocation from tasks.md
5. WHEN a task is complete, THE Implementer Agent SHALL mark the checkbox and publish ImplementationResult JSON

#### Scenario: Task Implementation

WHEN the Implementer agent is invoked with task 1.2 Create message type enum
THEN the agent SHALL read the SearchResult and implement the enum

#### Correctness Property

For any Implementer agent invocation, all writes SHALL be within the assigned worktree path.

**Validates:** Requirement 8

### Requirement 9: Verifier Agent - Test Runners Only

**User Story:** As a workflow orchestrator, I want a verifier agent that only runs tests, so that verification cannot modify implementation.

#### Acceptance Criteria

1. THE Verifier Agent SHALL have access only to: ruff check, ty check, pytest, ast-grep scan
2. THE Verifier Agent SHALL NOT have access to file write operations
3. WHEN invoked, THE Verifier Agent SHALL run all configured checks
4. THE Verifier Agent Output SHALL be structured JSON with verification_id and checks array
5. WHERE a check fails, THE Verifier Agent SHALL include failure details in the output

#### Scenario: Verification Execution

WHEN the Verifier agent is invoked after implementation
THEN the agent SHALL run: ruff check, ty check, pytest, ast-grep scan

#### Correctness Property

For any Verifier agent invocation, no file write operations SHALL be performed.

**Validates:** Requirement 9

### Requirement 10: Reviewer Agent - Semantic Review with Requirement Links

**User Story:** As a workflow orchestrator, I want a reviewer agent that produces structured reviews linked to requirements, so that findings are actionable.

#### Acceptance Criteria

1. THE Reviewer Agent SHALL read: git diff, specs, ADRs, requirements
2. THE Reviewer Agent SHALL NOT have access to file write operations
3. WHEN invoked, THE Reviewer Agent SHALL produce findings with: severity, requirement_id, file, line, description
4. THE Reviewer Agent Output SHALL be structured JSON with review_id and findings array
5. WHERE a finding is BLOCKER, THE Reviewer Agent SHALL reference the specific requirement violated

#### Scenario: Review Execution

WHEN the Reviewer agent is invoked after verification passes
THEN the agent SHALL read the diff and compare against specs and ADRs

#### Correctness Property

For any Reviewer finding, the requirement_id field SHALL reference a valid requirement.

**Validates:** Requirement 10
### Section 3: Brevity Protocol

### Requirement 11: Matt Pocock Brevity Protocol

**User Story:** As a token optimizer, I want all subagent system prompts to enforce brevity, so that responses are concise.

#### Acceptance Criteria

1. ALL subagent system prompts SHALL include the Brevity Protocol
2. THE Brevity Protocol SHALL specify: ≤3 sentences for conversational responses, code blocks exempt, no preamble/recap, structured outputs preferred
3. WHEN an agent produces a response exceeding 3 sentences of prose, THE System SHALL log a warning
4. WHERE structured output is possible, THE Agent SHALL use structured output

#### Scenario: Brevity Enforcement

WHEN an agent produces a prose response
THEN the system SHALL count sentences

#### Correctness Property

For any prose response from any subagent, the sentence count SHALL NOT exceed 3.

**Validates:** Requirement 11

### Section 4: Incident-Why-Outcome Comments

### Requirement 12: Block-Level HTML Comments with Incident-Why-Outcome

**User Story:** As a future maintainer, I want all rules annotated with incident context, so that I understand why each rule exists.

#### Acceptance Criteria

1. ALL rules in ALL SKILL.md files SHALL have block-level HTML comments
2. THE Comment SHALL include all 4 fields: INCIDENT, WHY THIS RULE HELPS, OUTCOME, COMMIT-REF
3. THE INCIDENT field SHALL include a specific date and description
4. THE OUTCOME field SHALL include a date range and measurable evidence

#### Scenario: Comment Validation

WHEN a new rule is added to a SKILL.md
THEN the Validator agent SHALL check for the Incident-Why-Outcome comment

#### Correctness Property

For any rule in any SKILL.md, the Incident-Why-Outcome comment SHALL be present with all 4 fields.

**Validates:** Requirement 12

### Section 5: Git Worktree Workflow

### Requirement 13: Worktree Creation and Isolation

**User Story:** As a parallel workflow runner, I want each workflow in its own worktree, so that concurrent workflows do not conflict.

#### Acceptance Criteria

1. WHEN an SDD workflow starts, THE System SHALL create a worktree: git worktree add ../sdd-feature_name -b sdd/feature_name
2. THE Worktree SHALL be created in a sibling directory to the repository root
3. THE Branch SHALL be named sdd/feature_name with sdd prefix
4. WHERE a worktree already exists for that feature, THE System SHALL reuse it

#### Scenario: Worktree Creation

WHEN a workflow starts for user-auth
THEN the system SHALL create worktree ../sdd-user-auth on branch sdd/user-auth

#### Correctness Property

For any workflow, a unique worktree SHALL exist before any implementation starts.

**Validates:** Requirement 13

### Requirement 14: Agent-to-Worktree Binding

**User Story:** As a workflow orchestrator, I want agents bound to specific worktrees, so that writes are isolated.

#### Acceptance Criteria

1. WHEN an agent is invoked, THE System SHALL pass the actionRoot parameter with the worktree path
2. WHEN the agent performs file operations, THE Agent SHALL resolve all paths relative to actionRoot
3. IF the agent attempts to write outside actionRoot, THE System SHALL reject the operation

#### Scenario: Worktree Binding

WHEN the Implementer agent is invoked for user-auth
THEN the system SHALL pass actionRoot as the worktree path

#### Correctness Property

For any file operation by any agent, the path SHALL be within the assigned actionRoot.

**Validates:** Requirement 14

### Requirement 15: Worktree Teardown and PR Merge

**User Story:** As a disk space manager, I want worktrees torn down after workflow completion, so that orphans do not accumulate.

#### Acceptance Criteria

1. WHEN a workflow completes and the PR is merged, THE System SHALL remove the worktree
2. THE System SHALL delete the feature branch after merge
3. WHERE the PR is closed without merge, THE System SHALL preserve the worktree for 7 days then remove

#### Scenario: Worktree Teardown

WHEN the PR for user-auth is merged
THEN the system SHALL run git worktree remove ../sdd-user-auth

#### Correctness Property

For any merged PR, the associated worktree SHALL be removed within 5 minutes.

**Validates:** Requirement 15
### Section 6: ADR Workflow

### Requirement 16: ADR Generation from Architect Decisions

**User Story:** As a future maintainer, I want all architecture decisions documented as ADRs, so that rationale is preserved.

#### Acceptance Criteria

1. WHEN the Architect agent produces an ArchitectureDecision, THE System SHALL generate an ADR
2. THE ADR SHALL be stored at: openspec/specs/feature_name/adr/ADR-NNN-slug.md
3. THE ADR Number SHALL be sequential within the feature
4. THE ADR Template SHALL include: Status, Context, Decision, Consequences, Alternatives

#### Scenario: ADR Generation

WHEN the Architect decides to use InMemoryMessageBoard for local development
THEN an ADR SHALL be generated at openspec/specs/user-auth/adr/ADR-001-in-memory-message-board.md

#### Correctness Property

For any ArchitectureDecision from the Architect, an ADR SHALL be generated.

**Validates:** Requirement 16

### Requirement 17: ADR Status Lifecycle

**User Story:** As a system architect, I want ADR status to reflect current state, so that I know which decisions are active.

#### Acceptance Criteria

1. THE ADR Status SHALL be one of: Proposed, Accepted, Deprecated, Superseded
2. WHEN an ADR is created, THE Status SHALL be Proposed
3. WHEN implementation starts, THE Status SHALL change to Accepted
4. WHEN a decision is no longer recommended, THE Status SHALL change to Deprecated
5. WHEN a new ADR replaces an old one, THE Old ADR Status SHALL change to Superseded with a reference

#### Scenario: Status Transition

WHEN ADR-002 supersedes ADR-001
THEN ADR-001 status SHALL change to Superseded by ADR-002

#### Correctness Property

For any ADR, the status SHALL accurately reflect the decision current state.

**Validates:** Requirement 17

### Requirement 18: ADR Supersession Chain

**User Story:** As a decision historian, I want superseded ADRs to link to their replacements, so that I can navigate decision evolution.

#### Acceptance Criteria

1. WHEN an ADR is superseded, THE ADR SHALL reference the superseding ADR number
2. THE Reference SHALL be in the format: Superseded by ADR-NNN
3. THE Superseding ADR SHALL reference the deprecated ADR in its Context section
4. WHERE multiple ADRs are superseded, THE New ADR SHALL list all deprecated ADRs

#### Scenario: Supersession Reference

WHEN ADR-003 supersedes ADR-001 and ADR-002
THEN ADR-001 SHALL contain Superseded by ADR-003

#### Correctness Property

For any superseded ADR, a reference to the superseding ADR SHALL exist.

**Validates:** Requirement 18

### Section 7: Message Board

### Requirement 19: Message Board Coordination System

**User Story:** As a system architect, I want the SDD system to use a shared message board for agent coordination, so that agents can communicate asynchronously without direct coupling.

#### Acceptance Criteria

1. THE Message Board SHALL support publishing and subscribing to messages
2. WHEN a message is published, THE Message Board SHALL deliver it to all subscribers within 500ms
3. THE Message Board SHALL maintain a registry of active agents and their capabilities
4. THE Message Board SHALL track workflow progress through phases
5. THE Message Board SHALL store and retrieve artifacts by ID

#### Scenario: Message Delivery

WHEN a subagent publishes a workflow_started message
THEN the Refinement Coordinator SHALL receive the message within 500ms

#### Correctness Property

For any valid message published to the message board, the message SHALL be delivered to all subscribers within the configured timeout.

**Validates:** Requirement 19

### Requirement 20: Subagent Registration and Lifecycle

**User Story:** As a system operator, I want subagents to register and deregister dynamically, so that the system can track available capabilities and manage agent lifecycle.

#### Acceptance Criteria

1. WHEN an agent starts, THE Agent SHALL publish an agent_registered message to the message board
2. WHEN an agent completes its task, THE Agent SHALL publish an agent_deregistered message
3. THE Message Board SHALL track each agent registration status and capabilities
4. WHILE an agent is registered, THE Message Board SHALL accept messages addressed to that agent

#### Scenario: Agent Registration

WHEN the Refinement Coordinator starts a new workflow phase
THEN the appropriate specialized agent SHALL be registered

#### Correctness Property

For any agent ID, the message board SHALL allow only one active registration at a time.

**Validates:** Requirement 20

### Requirement 21: Message Type Support

**User Story:** As a developer, I want the message board to support 14 distinct message types, so that different coordination patterns can be implemented.

#### Acceptance Criteria

1. THE Message Board SHALL support the following message types:
   - workflow_started, workflow_completed
   - agent_task, task_complete
   - artifact_created, artifact_updated
   - validation_result, audit_finding
   - feedback_request, feedback_response
   - agent_registered, agent_deregistered
   - status_update, error
2. WHEN a message is published, THE Message Type SHALL be one of the 14 defined types
3. THE Message Type SHALL be included in every message payload

#### Scenario: Status Updates

WHEN a workflow phase exceeds 30 seconds
THEN the system SHALL publish a status_update message

#### Correctness Property

For any message type in use, exactly one of the 14 defined types SHALL be used.

**Validates:** Requirement 21
