WHEN a requirement uses "SHALL NOT"

THEN the Validator SHALL flag it unless it is an error handling requirement

WHEN a non-error requirement uses negative language

THEN the Validator SHALL suggest a positive rephrasing

**Correctness Properties:**

- Property 1: For any non-error requirement, the requirement SHALL use "SHALL" not "SHALL NOT"

- Property 2: For any error handling requirement, negative statements SHALL be used only for describing error conditions

**User Story:** As a system designer, I want design documents to follow RFC-style structure, so that technical specifications are standardized and comprehensive.

1. ALL design documents SHALL include the following sections:

   - Overview & Context

   - Requirements & Constraints

   - Proposed Architecture (with diagrams)

   - Detailed Technical Design

   - Trade-offs & Alternatives Considered

   - Observability, Rollout & Migration

   - Open Questions

2. WHEN a design document is generated, THE System SHALL include all required sections

3. THE Architecture section SHALL include at least one diagram

4. WHERE a design decision is documented, THE System SHALL include rationale

5. WHILE reviewing design documents, THE Validator SHALL check for RFC structure compliance

WHEN a design document is missing a required section

THEN the Validator SHALL flag it as a structural violation

WHEN the Architecture section has no diagram

THEN the Validator SHALL flag it as a structural violation

**Correctness Properties:**

- Property 1: For any design document, all 7 required sections SHALL be present

- Property 2: For any design document, the Architecture section SHALL contain at least one diagram

**User Story:** As a system architect, I want agents to coordinate via message-passing, so that the system is loosely coupled and extensible.

1. ALL agent communication SHALL occur via the message board

2. NO agents SHALL communicate via direct function calls

3. WHEN an agent needs to request work, THE Agent SHALL publish an agent_task message

4. WHEN an agent completes work, THE Agent SHALL publish a task_complete message

5. WHILE coordination occurs, THE System SHALL maintain message delivery guarantees

WHEN the Refinement Coordinator needs requirements analysis

THEN it SHALL publish an agent_task message to the message board

WHEN the Requirements Analyst completes analysis

THEN it SHALL publish a task_complete message

**Correctness Properties:**

- Property 1: For any agent interaction, communication SHALL occur only via message board messages

- Property 2: For any agent task request, the corresponding task_complete SHALL be published

**User Story:** As a system component, I want artifacts to be stored and retrieved from the message board, so that agents can share design outputs.

1. WHEN an artifact is generated, THE System SHALL store it in the message board by ID

2. WHEN an artifact is requested, THE System SHALL retrieve it by ID

3. THE Stored Artifact SHALL be accessible to all agents

4. WHERE an artifact is updated, THE System SHALL store the new version

5. WHILE retrieving artifacts, THE System SHALL return the latest version

WHEN the Architect generates a design document

THEN it SHALL be stored with ID "<feature_name>/design.md"

WHEN another agent requests the design document

THEN it SHALL receive the stored content

**Correctness Properties:**

- Property 1: For any artifact storage operation, the content SHALL be stored by ID

- Property 2: For any artifact retrieval operation, the content SHALL match the stored content


**User Story:** As a system orchestrator, I want workflows to progress through defined phases, so that the SDD process is systematic and auditable

1. THE System SHALL enforce the following phase order:

   - Phase 1: Requirements

   - Phase 2: Design

   - Phase 3: Tasks

   - Phase 4: Git Integration

2. NO phase SHALL be skipped or reordered

3. WHEN a phase completes, THE System SHALL publish workflow progress

4. IF a phase fails validation, THE System SHALL return to previous phase

5. WHILE a phase is in progress, THE System SHALL track completion status

WHEN Phase 1 (Requirements) completes

THEN Phase 2 (Design) SHALL start

WHEN Phase 2 completes

THEN Phase 3 (Tasks) SHALL start

**Correctness Properties:**

- Property 1: For any workflow, phases SHALL complete in the defined order

- Property 2: For any workflow, Phase N+1 SHALL NOT start until Phase N completes

**User Story:** As a QA engineer, I want automated adversarial review generation, so that design flaws are discovered systematically.

3. EACH finding SHALL include category (ambiguity, missing edge cases, blast radius, testability

THEN it SHALL generate findings with severity levels

THEN the workflow SHALL be blocked until the finding is addressed

**Correctness Properties:**

- Property 1: For any design document, the QA Auditor SHALL generate at least one finding per category checked

- Property 2: For any CRITICAL finding, the workflow SHALL be blocked

1. ALL audit findings SHALL have one of three severity levels:

   - CRITICAL: Blocker that must be fixed before implementation

   - WARNING: Important issue that should be addressed

   - SUGGESTION: Improvement suggestion for consideration

2. WHEN a finding is classified as CRITICAL, THE System SHALL block progression to the next phase

3. WHEN a finding is classified as WARNING, THE System SHALL warn the user but allow progression

4. WHERE no CRITICAL or WARNING findings exist, THE System SHALL allow workflow progression

**Correctness Properties:**

- Property 1: For any CRITICAL finding, the workflow SHALL be blocked

- Property 2: For any WARNING finding, the workflow SHALL allow progression with notification

**User Story:** As a system user, I want an interactive loop of Make → Validate → Review → Refine, so that artifacts improve through iteration.

THEN the artifact SHALL be revalidated



**Correctness Properties:**

- Property 1: For any artifact, each refinement iteration SHALL be tracked

- Property 2: For any artifact, the final version SHALL be marked after completion



### Requirement 21: Git Branch Management



**User Story:** As a developer, I want git branch-per-feature management, so that SDD artifacts are isolated and traceable.



#### Acceptance Criteria

1. WHEN an SDD workflow starts, THE Git Integrator SHALL create a branch named "sdd-{feature_name}"

2. NO commits SHALL be made to the main branch during SDD workflows

3. WHEN artifacts are committed, THE System SHALL use commit-per-artifact pattern

4. WHERE a workflow completes, THE System SHALL leave the branch for review

5. WHILE a workflow is in progress, THE System SHALL track branch status

WHEN a new SDD workflow starts for "user-authentication"

THEN the Git Integrator SHALL create branch "sdd-user-authentication"

WHEN artifacts are committed

THEN they SHALL be committed to the feature branch

**Correctness Properties:**
- Property 1: For any SDD workflow, a branch SHALL be created with the naming pattern "sdd-{feature_name}"
- Property 2: For any SDD workflow, all commits SHALL be on the feature branch

**Correctness Properties:**

- Property 1: For any artifact, each version SHALL have exactly one commit

- Property 2: For any artifact, the commit message SHALL include artifact ID and action

**Correctness Properties:**

- Property 1: For any completed workflow, a PR SHALL be generated

- Property 2: For any PR, the description SHALL include all artifacts summary

**Correctness Properties:**

- Property 1: For any workflow, configuration SHALL be loaded before execution

- Property 2: For any configured strictness level, all rules SHALL be enforced accordingly


**Correctness Properties:**

- Property 1: For any artifact, the template SHALL be applied if defined

- Property 2: For any template, all variables SHALL be substituted before output



### Requirement 26: Language Detection



**User Story:** As a system user, I want automatic language detection from user requests, so that code examples use the correct language.



#### Acceptance Criteria



1. WHEN a user request is received, THE System SHALL detect programming language or pseudocode style

2. THE Detected Language SHALL be one of: Python, TypeScript, Java, Go, Rust, Structured Pseudocode, Mathematical Pseudocode, or Lean

3. WHERE no language is detected, THE System SHALL use Structured Pseudocode as default

4. THE Detection Confidence SHALL be reported

5. WHILE generating artifacts, THE System SHALL use detected language



#### Scenario: Language Detection



WHEN a user request contains "Python code example"

THEN the system SHALL detect Python as the language

WHEN no language indicators exist

THEN the system SHALL default to Structured Pseudocode



**Correctness Properties:**

- Property 1: For any user request, a language SHALL be detected or defaulted

- Property 2: For any detected language, the confidence SHALL be reported



### Requirement 27: Agent Capability Matching



**User Story:** As a system orchestrator, I want agent capability matching, so that tasks are assigned to appropriate agents.



#### Acceptance Criteria



1. WHEN a task needs to be assigned, THE System SHALL match agent capabilities to task requirements

2. NO task SHALL be assigned to an agent without required capabilities

3. WHERE multiple agents match, THE System SHALL select the best match

4. THE Capability Matching SHALL be documented in workflow logs

5. WHILE matching, THE System SHALL consider agent availability



#### Scenario: Capability Matching



WHEN a requirements analysis task is created

THEN the Refinement Coordinator SHALL assign it to an agent with "ears_analysis" capability

WHEN no agent has the capability

THEN the system SHALL fail with a capability error



**Correctness Properties:**

- Property 1: For any task, the assigned agent SHALL have all required capabilities

- Property 2: For any task assignment, the selection SHALL be logged



### Requirement 28: Feedback Integration



**User Story:** As a system user, I want feedback integration into artifacts, so that improvements are tracked and applied.



#### Acceptance Criteria



1. WHEN feedback is received, THE System SHALL integrate it into the artifact

2. EACH feedback item SHALL be tracked with status (open, addressed, rejected)

3. WHERE feedback is addressed, THE System SHALL create a new artifact version

4. THE Feedback History SHALL be maintained for audit purposes

5. WHILE integrating feedback, THE System SHALL preserve existing content



#### Scenario: Feedback Integration

#### Correctness Property



WHEN feedback is received on a design document

THEN the system SHALL create a feedback entry with status "open"

WHEN the feedback is addressed

THEN a new version SHALL be created with status "addressed"

For any artifact, the template SHALL be applied if defined.



**Correctness Properties:**

- Property 1: For any feedback item, the status SHALL be tracked

- Property 2: For any addressed feedback, a new version SHALL be created
1. WHEN a workflow starts, THE System SHALL publish workflow_started message
2. WHILE a workflow is in progress, THE System SHALL publish status_update messages
3. WHEN a workflow completes, THE System SHALL publish workflow_completed message
4. WHERE a workflow fails, THE System SHALL publish error message
5. THE Status Update SHALL include phase name and completion percentage

#### Scenario: Progress Monitoring

WHEN a workflow enters the design phase
THEN the system SHALL publish status_update with phase "design"
WHEN 50% of the phase is complete
THEN the status_update SHALL include 50% completion

**Correctness Properties:**
- Property 1: For any workflow, status updates SHALL be published at regular intervals
- Property 2: For any workflow, completion percentage SHALL be accurate
 



**User Story:** As a system component, I want proper error handling, so that failures are reported and handled gracefully.

1. WHEN an error occurs, THE System SHALL publish an error message

2. EACH error message SHALL include error type and description

3. WHERE an agent fails, THE System SHALL deregister the agent

4. THE Error Message SHALL include suggested recovery action

5. WHILE handling errors, THE System SHALL maintain data integrity

WHEN a git operation fails

THEN the system SHALL publish an error message with type "GitException"

WHEN an agent fails during task execution

THEN the agent SHALL be deregistered

**Correctness Properties:**

- Property 1: For any error, an error message SHALL be published

- Property 2: For any agent failure, the agent SHALL be deregistered

1. THE Message Board SHALL support publishing and subscribing to messages

2. WHEN a message is published, THE Message Board SHALL deliver it to all subscribers within 500ms

3. THE Message Board SHALL maintain a registry of active agents and their capabilities

4. THE Message Board SHALL track workflow progress through phases

5. THE Message Board SHALL store and retrieve artifacts by ID

WHEN a subagent publishes a workflow_started message

THEN the Refinement Coordinator SHALL receive the message within 500ms

WHEN the message is delivered

THEN all registered subscribers SHALL receive the same message content



**Correctness Properties:**

- Property 1: For any valid message published to the message board, the message SHALL be delivered to all subscribers within the configured timeout

- Property 2: For any agent ID, the message board SHALL allow only one active registration at a time

- Property 3: For any artifact stored in the message board, all subscribers SHALL see the same content when reading the artifact after storage

1. WHEN an agent starts, THE Agent SHALL publish an agent_registered message to the message board

2. WHEN an agent completes its task, THE Agent SHALL publish an agent_deregistered message

3. THE Message Board SHALL track each agent's registration status and capabilities

4. WHILE an agent is registered, THE Message Board SHALL accept messages addressed to that agent

5. IF a registration conflict occurs, THE Message Board SHALL return an error message

WHEN the Refinement Coordinator starts a new workflow phase

THEN it SHALL register the appropriate specialized agent

WHEN the agent completes its task

THEN it SHALL deregister and the coordinator SHALL receive agent_deregistered confirmation

WHEN the Refinement Coordinator starts a new workflow phase

THEN it SHALL register the appropriate specialized agent

WHEN the agent completes its task

THEN it SHALL deregister and the coordinator SHALL receive agent_deregistered confirmation

**User Story:** As a developer, I want the message board to support 14 distinct message types, so that different coordination patterns can be implemented.

THE Message Board SHALL support the following message types:

   - workflow_started, workflow_completed

   - agent_task, task_complete

   - artifact_created, artifact_updated

   - validation_result, audit_finding

   - feedback_request, feedback_response

   - agent_registered, agent_deregistered

   - status_update, error

2. WHEN a message is published, THE Message Type SHALL be one of the 14 defined types

3. THE Message Type SHALL be included in every message payload

4. WHERE an error occurs, THE System SHALL publish an error message with details

5. WHILE a workflow is in progress, THE System SHALL publish status_update messages every 30 seconds

WHEN a workflow phase exceeds 30 seconds

THEN the system SHALL publish a status_update message

WHEN the status_update is published

THEN it SHALL include the current phase name and completion percentage

**Correctness Properties:**

- Property 1: For any message type in use, exactly one of the 14 defined types SHALL be used

- Property 2: For any status_update message, the timestamp difference from the previous update SHALL be between 25s and 35s



**User Story:** As a requirements analyst, I want to express core functionality using ubiquitous requirements, so that essential behaviors are clearly defined and always apply.

1. FOR ALL requirements that always apply regardless of context, THE System SHALL use the ubiquitous pattern: "THE <System> SHALL <response>"

2. WHEN a ubiquitous requirement is written, THE Description SHALL not contain condition clauses

3. THE Description SHALL use active voice with explicit subject and action

4. WHERE a requirement describes a core capability, THE System SHALL use ubiquitous format

5. WHILE reviewing requirements, THE Validator SHALL check for ubiquitous pattern compliance



WHEN a requirement describes core system functionality

THEN it SHALL use "THE <System> SHALL <response>" format

WHEN the format is applied

THEN the requirement SHALL be classified as EARS_UBIQUITOUS

**Correctness Properties:**

- Property 1: For any ubiquitous requirement, the format SHALL be "THE <System> SHALL <action>" with no condition clauses

- Property 2: For any ubiquitous requirement, the description SHALL use active voice with no pronouns

**User Story:** As a requirements analyst, I want to express event-triggered behaviors using event-driven requirements, so that reactive behaviors are clearly defined.



1. FOR ALL requirements triggered by specific events, THE System SHALL use the event-driven pattern: "WHEN <trigger>, THE <System> SHALL <response>"

2. WHEN an event-driven requirement is written, THE Trigger SHALL describe a measurable event

3. THE Trigger SHALL not contain vague terms like "quickly" or "adequately"

4. WHERE a requirement responds to an external event, THE System SHALL use event-driven format

5. WHILE reviewing requirements, THE Validator SHALL check for event-driven pattern compliance



WHEN a requirement describes behavior triggered by an event

THEN it SHALL use "WHEN <trigger>, THE <System> SHALL <response>" format

WHEN the format is applied

THEN the requirement SHALL be classified as EARS_EVENT_DRIVEN

**Correctness Properties:**

- Property 1: For any event-driven requirement, the trigger SHALL be measurable or verifiable

- Property 2: For any event-driven requirement, the response SHALL be deterministic given the trigger



**User Story:** As a requirements analyst, I want to express state-dependent behaviors using state-driven requirements, so that behaviors that depend on system state are clearly defined.



1. FOR ALL requirements that apply during specific states, THE System SHALL use the state-driven pattern: "WHILE <condition>, THE <System> SHALL <response>"

2. WHEN a state-driven requirement is written, THE Condition SHALL describe a measurable state

3. THE Condition SHALL not contain vague terms like "user-friendly" or "reasonable"

4. WHERE a requirement applies only in certain system states, THE System SHALL use state-driven format

5. WHILE reviewing requirements, THE Validator SHALL check for state-driven pattern compliance



WHEN a requirement describes behavior that depends on system state

THEN it SHALL use "WHILE <condition>, THE <System> SHALL <response>" format

WHEN the format is applied

THEN the requirement SHALL be classified as EARS_STATE_DRIVEN



- Property 1: For any state-driven requirement, the condition SHALL be measurable or verifiable at runtime

- Property 2: For any state-driven requirement, the response SHALL be consistent while the condition holds



**User Story:** As a requirements analyst, I want to express error handling and edge cases using unwanted behavior requirements, so that failure modes are clearly defined.

1. FOR ALL requirements describing unwanted situations, THE System SHALL use the unwanted behavior pattern: "IF <condition>, THEN THE <System> SHALL <response>"

2. WHEN an unwanted behavior requirement is written, THE Condition SHALL describe an error or edge case

3. THE Response SHALL describe how the system handles the unwanted situation

4. WHERE a requirement describes error handling, THE System SHALL use unwanted behavior format

5. WHILE reviewing requirements, THE Validator SHALL check for unwanted behavior pattern compliance

W

HEN a requirement describes error handling or edge case behavior

THEN it SHALL use "IF <condition>, THEN THE <System> SHALL <response>" format

WHEN the format is applied

THEN the requirement SHALL be classified as EARS_UNWANTED_BEHAVIOR

**Correctness Properties:**

- Property 1: For any unwanted behavior requirement, the condition SHALL be a specific, identifiable error or edge case

- Property 2: For any unwanted behavior requirement, the response SHALL allow the system to continue operation (unless it is a fatal error)



**User Story:** As a requirements analyst, I want to express optional or configurable behaviors using optional features requirements, so that optional functionality is clearly defined.

1. FOR ALL requirements describing optional behaviors, THE System SHALL use the optional features pattern: "WHERE <option>, THE <System> SHALL <response>"

2. WHEN an optional features requirement is written, THE Option SHALL describe a configurable feature

3. THE Option SHALL not contain absolute terms like "always" or "never"

4. WHERE a requirement describes optional functionality, THE System SHALL use optional features format

5. WHILE reviewing requirements, THE Validator SHALL check for optional features pattern compliance



WHEN a requirement describes optional or configurable behavior

THEN it SHALL use "WHERE <option>, THE <System> SHALL <response>" format

WHEN the format is applied

THEN the requirement SHALL be classified as EARS_OPTIONAL_FEATURES

**Correctness Properties:**

- Property 1: For any optional features requirement, the option SHALL be configurable or conditional

- Property 2: For any optional features requirement, the response SHALL be valid when the option is enabled

*User Story:** As a requirements analyst, I want to express complex requirements with multiple conditions using the complex pattern, so that multi-condition behaviors are clearly defined.

1. FOR ALL requirements with multiple conditions, THE System SHALL use the complex pattern with clause order: [WHERE] [WHILE] [WHEN/IF] THE <System> SHALL <response>

2. WHEN a complex requirement is written, THE Clause Order SHALL be: WHERE → WHILE → WHEN/IF → THE → SHALL

3. WHERE multiple conditions apply to a requirement, THE System SHALL use complex format

4. WHILE reviewing requirements, THE Validator SHALL check for complex pattern clause order compliance

5. THE System SHALL NOT use complex format for requirements with only one condition

WHEN a requirement has multiple conditions (e.g., state + event)

THEN it SHALL use "WHERE <optional> WHILE <state> WHEN <event> THE <System> SHALL <response>" format

WHEN the format is applied

THEN the requirement SHALL be classified as EARS_COMPLEX

**Correctness Properties:**

- Property 1: For any complex requirement, the clause order SHALL be WHERE → WHILE → WHEN/IF → THE → SHALL

- Property 2: For any complex requirement, the number of condition clauses SHALL be at least 2


**User Story:** As a requirements analyst, I want all requirements to follow INCOSE quality rules for clarity and precision, so that requirements are unambiguous and testable.

. ALL requirements SHALL use active voice with explicit subject and action

2. NO requirements SHALL contain vague terms like "quickly", "adequately", "reasonable", or "user-friendly"

3. NO requirements SHALL use pronouns like "it", "them", or "they"

4. ALL requirements SHALL use consistent terminology from the Glossary

5. WHILE reviewing requirements, THE Validator SHALL flag requirements violating clarity rules

WHEN a requirement contains vague terms

THEN the Validator SHALL flag it as a clarity violation

WHEN a requirement uses pronouns

THEN the Validator SHALL flag it as a clarity violation

**Correctness Properties:**

- Property 1: For any valid requirement, the description SHALL not contain vague terms

- Property 2: For any valid requirement, the description SHALL not use pronouns

**User Story:** As a requirements analyst, I want all requirements to follow INCOSE quality rules for testability, so that requirements can be verified and validated.

1. ALL conditions in requirements SHALL be measurable or verifiable

2. ALL criteria in requirements SHALL be specific and quantifiable where applicable

3. ALL timing requirements SHALL specify realistic tolerances (e.g., "within 200ms" not "quickly")

4. EACH requirement SHALL test one thing only

5. WHILE reviewing requirements, THE Validator SHALL check for testability violations

WHEN a requirement has unmeasurable conditions

THEN the Validator SHALL flag it as a testability violation

WHEN a requirement combines multiple checks

THEN the Validator SHALL flag it as a testability violation

**Correctness Properties:**

- Property 1: For any valid requirement, all conditions SHALL be measurable or verifiable

- Property 2: For any timing requirement, the tolerance SHALL be specified with a numeric value and unit

**User Story:** As a requirements analyst, I want all requirements to follow INCOSE quality rules for completeness, so that requirements are fully specified without escape clauses.

**Validates:** Requirement 11

1. NO requirements SHALL contain escape clauses like "where possible", "if feasible", or "as appropriate"

2. NO requirements SHALL use absolutes like "never", "always", or "100%" unless truly absolute

3. ALL requirements SHALL focus on what the system should do, not how it should do it

4. EACH requirement SHALL be complete and self-contained

5. WHILE reviewing requirements, THE Validator SHALL check for completeness violations

WHEN a requirement contains an escape clause

THEN the Validator SHALL flag it as a completeness violation

WHEN a requirement uses absolute terms in non-absolute contexts

THEN the Validator SHALL flag it as a completeness violation

**Correctness Properties:**

- Property 1: For any valid requirement, there SHALL be no escape clauses

- Property 2: For any timing requirement, the bound SHALL be realistic and achievable



**User Story:** As a requirements analyst, I want all requirements to follow INCOSE quality rules with positive statements, so that requirements define desired behavior.

1. ALL requirements SHALL use positive statements with "SHALL" rather than "SHALL NOT"

2. NO requirements SHALL state what the system should not do (except for error handling)

3. ERROR handling requirements MAY use negative statements when necessary

4. WHILE reviewing requirements, THE Validator SHALL check for positive statement compliance

5. THE System SHALL convert negative statements to positive where possible

- **Message Board**: Central coordination layer providing shared state, async messaging, agent registration, and status tracking. Acts as the single source of truth for artifacts and workflow progress.

- **Subagent**: Specialized autonomous agent with dedicated responsibilities (Requirements Analyst, Architect, QA Auditor, Git Integrator, Customizer, Validator, Task Generator, Refinement Coordinator).

- **Refinement Coordinator**: Main orchestrator that parses user requests, delegates to specialized agents, manages feedback loops, and coordinates with OpenSpec CLI.

- **EARS Methodology**: Requirements specification methodology with six patterns: Ubiquitous, Event-driven, State-driven, Unwanted behavior, Optional features, and Complex.

- **RFC-Style Design**: Standardized design document format with Overview, Requirements, Architecture, Trade-offs, and Open Questions sections.

- **Hostile QA Audit**: Adversarial review process with automated checks, edge case analysis, and adversarial attack patterns to discover design flaws.

- **Artifact**: Design artifact stored in the message board (requirements.md, design.md, tasks.md, review.md, etc.).

- **Message Type**: Categorized message format for inter-agent communication (workflow_started, agent_task, task_complete, etc.).

Property is missing **Validates: Requirements X.Y** reference
 
need to make a garbage document so that i can have all the errors and warning raised by the linter
 