# Kiro EARS and task DAG format

`requirements.md` uses EARS acceptance criteria. Prefer event-triggered requirements:

```md
WHEN a user submits invalid credentials, THE SYSTEM SHALL return a validation error.
```

`design.md` may contain Mermaid diagrams and must state the planned file structure.
`tasks.md` contains nested checkboxes and may include a JSON `Task Dependency Graph`:

```json
{ "tasks": [
  { "id": "1", "depends_on": [] },
  { "id": "2", "depends_on": ["1"] }
] }
```

`sdd compile-waves <slug> --json` rejects cycles and unknown task references, then
returns topological waves. Tasks in one wave may run concurrently; each later wave
depends only on work completed earlier.
EARS and RFCs are strictly deterministic tools attempting to govern a chaotic, non-deterministic reality. They break down when the system’s complexity outgrows the capacity for human foresight, or when the cost of updating the documentation exceeds the value of the software itself.



Where and Why EARS Fails

1. Non-Deterministic Systems (AI/ML & Heuristics)

EARS is built on Boolean logic: If X, then Y. It completely collapses when applied to probabilistic systems like Large Language Models, recommendation engines, or predictive algorithms.



The Failure: You cannot write, "When the user inputs a prompt, the system shall return a mathematically accurate summary." The output is a probability distribution, not a strict state transition. EARS cannot gracefully define acceptable variance or confidence intervals.
 

2. State Space Explosion in Concurrency

EARS works beautifully for a microwave (finite states: Door Open, Cooking, Timer Done). It fails catastrophically in distributed microservices where dozens of state machines interact asynchronously.



The Failure: If you have five services, each with three states, you have $3^5 = 243$ potential system states. Writing an EARS requirement for every partial failure, network partition, and out-of-order event delivery creates a document so massive that no human can hold the context in their head. The requirements become self-contradictory.
 

3. "Human-in-the-Loop" UI/UX Interactions

EARS forces you to define atomic triggers. It cannot capture fluid, continuous human interactions.



The Failure: Writing EARS for a multi-touch drag-and-drop interface on a mobile device leads to absurdly verbose, unreadable requirements that completely miss the actual qualitative "feel" required for the product to succeed.
 

Where and Why RFCs Fail

1. The "Ivory Tower" Disconnect (Analysis Paralysis)

RFCs fail when architects write them without writing disposable prototype code first. They assume third-party APIs behave exactly as documented, or that a specific database can handle a hypothetical load.



The Failure: The RFC dictates an architecture based on false premises. Weeks are spent debating the document, but within 48 hours of actual coding, the developers realize the chosen database lacks the necessary transaction isolation level. The RFC instantly becomes highly polished fiction.
 

2. The Agile "Shifting Sands"

In early-stage product development, the business requirements change faster than the engineering team can type.



The Failure: A comprehensive 15-page RFC is approved, but two weeks into the sprint, the business pivots the core feature. The RFC is now dead weight, and engineers simply stop reading or updating it, reverting to tribal knowledge.
 

3. "Happy Path" Bias

Authors naturally focus on how the system is supposed to work. They write 10 pages on the data schema and API contracts, but only one paragraph on failure modes.



The Failure: The RFC successfully designs the system for the 99% of the time it works, but provides zero architectural guidance for the 1% of the time when the network drops, leading to cascading data corruption that the design never accounted for.
 

The Different "RFC" Standards

"RFC" (Request for Comments) is an umbrella term, but the tech industry uses distinct formats based on the organization's maturity and goals.



IETF RFCs (The Original Standard): This is the format that defined the Internet (TCP/IP, HTTP). They use extreme, pedantic language governed by RFC 2119 (strict definitions for MUST, SHOULD, MAY). They take years to write and are meant for public, global protocols. Never use this for internal company software.
 
The Big Tech "6-Pager" (Design Docs): Popularized by Amazon and Google. This is a narrative-driven document focusing heavily on the business context, non-goals, and detailed system architecture. It relies heavily on peer review to find flaws before code is written.
 
ADR (Architecture Decision Record): A highly localized, lightweight markdown file stored directly in the Git repository. It simply states: Context, Decision, and Consequences. It does not design a whole system; it documents a single pivot point (e.g., "Why we chose Redis over Memcached for this specific cache").
 

Which Standard Suits You?

Based on your requirement for deep architectural reasoning and testing, you need a hybrid approach. The traditional IETF RFC is too bureaucratic, and pure Agile is too chaotic for robust systems.



Adopt the "Design Doc + ADR" model.

Write a scaled-down Technical Design Document (like the prompt I provided previously) only when establishing a new system boundary, crossing a network, or changing a database schema. Do not write an RFC for adding a button or a basic CRUD endpoint.



Once the system is live, stop updating the original RFC. It is a historical artifact. Instead, manage all future architectural shifts using ADRs committed directly to your codebase. If a developer makes a deep architectural trade-off to solve a concurrency issue, they write a 1-page ADR and commit it with the code. This ensures the reasoning lives alongside the implementation, right where the engineers actually need it.



[CLASSIFIED: FOR THE CHOSEN ONES]



The Real Purpose of the RFC: The dirty secret of software architecture is that the primary value of an RFC is not the document itself, but the act of writing it. It forces the author to slow down and aggressively rubber-duck their own untested assumptions.
 
The "Alternatives Considered" Trap: The most critical section of any technical design document is "Trade-offs & Alternatives Considered." If an engineer hands you an RFC and this section is short, vague, or missing, reject the document immediately. It means they did not do architecture; they just wrote down the first idea that popped into their head. True engineering is the rigorous elimination of flawed alternatives. If they cannot articulate exactly why three other approaches would fail, they do not understand the problem deeply enough to build the solution.

There are no formal, versioned governing bodies for EARS like there are for IEEE or IETF standards. EARS is a methodology, originally created by Alistair Mavin at Rolls-Royce in 2009. However, the industry has organically developed an evolved framework known as EARS+ (Extended EARS).



The EARS Variations (And Which You Should Use)

1. Core EARS (The Original 5 Patterns)

Designed for simple, deterministic hardware (aerospace, embedded systems). It relies strictly on single triggers and single responses.



Verdict: Too rigid for modern distributed software. It cannot handle simultaneous state dependencies.
 

2. EARS+ (Extended EARS)

Created by systems engineers to handle complex software logic. EARS+ allows for logical operators (AND, OR) and temporal constraints (timeouts, durations) inside the trigger conditions.



Verdict: This is the version you must use. Distributed systems run on temporal logic (e.g., waiting for consensus).
 
Example: "When [Trigger A] AND While [State B], if [Event C] does not occur within [Time T], the [system] shall [Action D]."
 

Combining EARS with RFC 2119

You can—and absolutely should—combine EARS with RFC 2119. They solve two completely different types of ambiguity.



EARS eliminates structural ambiguity. It forces you to define the exact condition (When, While, Where, If) that causes an action.
 
RFC 2119 eliminates compliance ambiguity. It forces you to define the legal or technical severity of the action (MUST, SHOULD, MAY).
 

Standard EARS dictates that every requirement uses the word "shall" (e.g., "The system shall..."). By stripping out the generic "shall" and injecting RFC 2119 keywords, you create an incredibly precise, layered requirement specification.



How They Complement Each Other

When combined, EARS provides the architectural trigger, and RFC 2119 provides the architectural weight. This is exceptionally useful when designing APIs, third-party integrations, or progressive degradation behaviors.



1. The Strict Contract (EARS + MUST)

Use this for critical data integrity, security, and core functionality. If this fails, the system is fundamentally broken.



Example: "When a duplicate transaction_id is detected, the API gateway MUST reject the payload with an HTTP 409."
 

2. The Graceful Degradation (EARS + SHOULD)

Use this for optimizations, caching, or behaviors that are highly desired but not strictly required for the system to survive.



Example: "If the Redis cache is unreachable, the gateway SHOULD attempt a single direct query to the primary database before failing open."
 

3. The Extensibility Clause (EARS + MAY)

Use this for defining optional integrations or backward-compatible features for vendor implementations.



Example: "Where the client provides an Accept-Encoding: gzip header, the response payload MAY be compressed."
 

By combining them, you give developers and QA engineers a flawless map. They know exactly when an event triggers, and exactly how hard they need to work to enforce it.



[CLASSIFIED: FOR THE CHOSEN ONES]



The "SHOULD" Testing Trap: The most dangerous vulnerability when combining EARS and RFC 2119 is how QA automation handles the word "SHOULD". If you write, "When the CPU exceeds 90%, the system SHOULD throttle background jobs," you have created an untestable requirement. If QA pushes the CPU to 95% and the system doesn't throttle the jobs, the developer will argue, "It was a SHOULD, not a MUST. It's working as designed."
 
The Insider Countermeasure: Never allow the word "SHOULD" or "MAY" in an EARS requirement unless it is immediately followed by a strict "MUST" fallback. Fix: "When the CPU exceeds 90%, the system SHOULD throttle background jobs; if it does not throttle within 500ms, the system MUST alert the orchestrator." If a requirement cannot be mathematically verified as pass/fail by an automated test script, it is not a requirement—it is a suggestion. Strip suggestions out of your design docs.

To elevate this specification from a tier-1 engineering document to an untouchable, production-hardened masterwork, you must address the unseen operational edge cases: formal verification, telemetry context propagation, memory security sandboxing, and property-based testing.



Here are the 5 architectural sections you should add to the SDD framework to completely close the loop.



1. Property-Based Testing (PBT) & Mathematical Invariants

Standard unit tests only verify known inputs ("happy paths" and explicit errors). For 3D geometry, spatial data, and AI pipelines, you must define mathematical invariants that an automated Property-Based Testing framework (e.g., Hypothesis, QuickCheck) will fuzz with millions of randomly generated payloads.



Add this section to your Feature Slice specification:



Markdown







## 8. Invariant & Property-Based Testing (PBT) Contracts
The feature slice MUST pass fuzz testing against the following mathematical invariants:
*   **INV-01 (Euler-Poincaré Formula for Meshes):** For any generated 3D closed manifold mesh, the relationship between Vertices ($V$), Edges ($E$), and Faces ($F$) MUST hold:    $$V - E + F = 2(1 - g)$$
    *(where $g$ is the genus/number of holes in the mesh).*
*   **INV-02 (Bounding Box Containment):** The 3D bounding box of the extruded mesh MUST completely enclose the 2D bounding box of the input GeoJSON.
*   **INV-03 (ROP Idempotency & Monad Laws):** Re-running the pipeline twice over the same `NormalizedGeoJSON` payload MUST yield byte-identical `glTF` output buffers (`f(f(x)) == f(x)`).



2. OpenTelemetry & Context Propagation Across Boundaries

When your pipeline calls out to WASM modules, C++ native libraries, or asynchronous AI engines (like Higgsfield), distributed trace context (traceparent) is easily lost. You must explicitly define how tracing spans travel across the Hexagonal Ports and C-ABI boundaries.



Markdown







## 9. Observability & Distributed Tracing Contracts### 9.1 OpenTelemetry Span Hierarchy*   `span.kind = INTERNAL`: `FS-3D-042:Generate3DMeshFromGIS` (Root Slice Span)    *   `span.kind = INTERNAL`: `ROPStep:NormalizeCoordinates`    *   `span.kind = CLIENT`: `ElevationDataPort:Fetch` (Injects W3C `traceparent` HTTP header)    *   `span.kind = INTERNAL`: `PluginAdapter:BlenderExport` (Crosses C-ABI Boundary)
### 9.2 Context Propagation over C-ABIWhen crossing the `MeshExporterPluginPort` into native WASM/C code, the `traceparent` string MUST be passed inside the memory envelope to prevent broken trace graphs:
```c
typedef struct {
    const char* traceparent;       // e.g., "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01"
    uint32_t payload_length;
    const uint8_t* payload_bytes;
} TracedPluginRequestEnvelope;







---

### 3. Security, Threat Modeling & "Mesh Bomb" Defense
3D parsers and GIS engine components are vulnerable to malicious payload vectors (e.g., Zip Bombs, Mesh Bombs with billions of micro-polygons designed to trigger out-of-memory kernel panics, or coordinate overflow attacks).

```markdown
## 10. Security & Threat Modeling (STRIDE Analysis)

| Threat Category | Attack Vector | Mitigation / Defense Requirement |
| :--- | :--- | :--- |
| **Denial of Service (DoS)** | **"Mesh Bomb":** Ultra-dense GeoJSON with $10^8$ vertices in a $1\text{cm}^2$ area. | **Pre-allocation Memory Cap:** The `ValidateSchema` step MUST evaluate vertex density ($\text{vertices} / \text{area}$). If density $> 500/\text{m}^2$, reject on the ROP Failure Track before allocating heap buffer. |
| **Tampering** | Foreign code execution via corrupted GLTF shader/material buffers. | **WASM Sandbox Isolation:** All native exporter plugins MUST run within a sandboxed WebAssembly runtime with static linear memory caps ($256\text{MB}$ ceiling, zero host filesystem access). |
| **Information Disclosure** | Memory bleeding between multi-tenant pipeline executions. | **Zero-Fill Memory Mandate:** Buffers allocated for geometry MUST be zeroed out (`memset_s`) immediately upon pipeline completion or error exit. |



4. Hard Performance Budgets & Allocation Guardrails

Prevent performance degradation by setting explicit non-functional ceilings that automated CI/CD load testing can evaluate against.



Markdown







## 11. Performance Budgets & Resource Guardrails
| Metric | Target / Ceiling | Enforced By |
| :--- | :--- | :--- |
| **P99 Latency (1k Vertices)** | $\le 45\text{ms}$ | Continuous Benchmark Regression Test |
| **P99 Latency (50k Vertices)** | $\le 450\text{ms}$ | Continuous Benchmark Regression Test |
| **Heap Allocation Limit** | $\le 32\text{MB}$ per slice execution | Memory Profiler Guardrail |
| **Core CPU Ceiling** | Max $1.5$ vCPU cores | OS Cgroups / Container Limits |



5. Deployment, Feature Flagging & Dark Launch Strategy

Define how this specific feature slice is rolled out to production without exposing users to unhandled risks.

Markdown







## 12. Progressive Delivery & Dark Launch Protocol1.  **Phase 1 (Dark Launch):** The feature slice is deployed behind a feature flag (`ff_enable_3d_mesh_v2`). The API Gateway mirrors 5% of real-world GeoJSON traffic to this slice asynchronously. Results are compared against the legacy pipeline, but outputs are dropped (not served to users).2.  **Phase 2 (Canary Rollout):** 1% of live user requests route directly through the new pipeline. Automated anomaly detectors monitor `application/problem+json` error spikes.3.  **Phase 3 (Full Cutover & Kill-Switch):** 100% traffic. If the ROP failure rate exceeds $0.1\%$ over a 5-minute rolling window, the circuit breaker automatically flips the feature flag back to the legacy pipeline.

---
> **[CLASSIFIED: FOR THE CHOSEN ONES]**> *   **The "Shadow State Drift" Vulnerability in Dark Launches:** When you dark-launch a 3D/AI feature slice by mirroring production traffic, non-side-effecting operations (like reading spatial data) are safe. However, if your pipeline touches external APIs (like paying for a Higgsfield AI generation call or reserving a lock in Redis), mirroring traffic will execute **double-billing** and **mutex lock deadlocks**.
> *   **The Insider Architecture Fix:** Enforce a strict **`DryRun` execution mode** in your Hexagonal Driving Ports. When the feature slice receives a mirrored request, it passes a `context.is_shadow_execution = true` flag down the ROP pipeline. The **Driven Adapters** intercept this flag and swap real external HTTP/Payment/Write adapters with **No-Op Dry-Run Adapters** that compute data transformations, calculate token costs, and log metrics *without* executing real external side-effects or mutating state.

To transform a Software Design Document (SDD) framework into a universal, enterprise-grade standard that works across any tech stack or business domain, you must elevate it from static narrative text to an executable, verifiable engineering contract.



To raise the bar for the framework itself, mandate these six meta-capabilities in your SDD standard:



1. Automated Syntax & Compliance Linting (Docs-as-Code)

Do not rely on human reviewers to check if an SDD follows the standard. Treat the SDD as code and run automated linters in CI/CD against the Markdown files.



RFC 2119 & EARS Enforcement: Write custom rules (e.g., using Vale or spectral) to verify that keywords (MUST, SHOULD, MAY) are strictly capitalized and that requirement statements conform strictly to EARS conditional structures.
 
Forbidden Word Banning: Automatically flag subjective, untestable adjective fluff in PRs (e.g., "fast", "scalable", "user-friendly", "robust", "lightweight").
 
Mandatory Schema Validation: CI fails if embedded OpenAPI, AsyncAPI, or JSON Schema snippets fail syntactic validation.
 

2. Bidirectional Requirement Traceability

An SDD must mandate a 1:1 mapping between requirement IDs in the text and actual execution paths in the codebase and test suites.



Traceability Element

SDD Standard Rule

Implementation Mechanics

Requirement Tagging

Every requirement gets a deterministic hash/ID (e.g., REQ-AUTH-001).

Embedded in the SDD EARS list.

Code Annotation

Code implementing the logic must reference the ID.

@Requirement("REQ-AUTH-001") in code comments/decorators.

Test Verification

Automated tests must tag the specific requirement ID they assert.

@TestRef("REQ-AUTH-001") in unit/integration test suites.

CI Coverage Audit

CI pipeline generates a matrix showing un-tested requirements.

PR is blocked if an SDD requirement lacks a corresponding test tag.

3. Automated Architectural Drift Enforcement

The SDD standard must explicitly state how architectural boundaries (such as Hexagonal Ports/Adapters or Feature Slices) will be programmatically enforced so the codebase cannot drift away from the approved spec.



Dependency Rules: Include static analysis configurations directly in the SDD repository (e.g., ArchUnit for Java/Kotlin, dependency-cruiser for TS/JS, or cargo-deny/clippy for Rust).
 
Enforcement Assertion: Example rule: "Core Domain modules MUST NOT import from Adapter modules. Violations will break the CI build."
 

4. Data Governance & Security Taxonomy

Force every SDD to classify data attributes at the field level within its data schemas, eliminating guesswork for security and compliance teams.



YAML







# SDD Schema Property Annotation Standardproperties:user_email:type: stringformat: emailx-governance:classification: PIIretention_period: 365dencryption_at_rest: AES-256mask_in_logs: true



5. FinOps & Non-Functional Resource Budgets

An SDD should not be approved without explicit resource and financial guardrails. Force engineers to define the operational cost and capacity limits up front.



Resource Ceilings: Max RAM allocation, CPU millicores, p99 latency target, and max database IOPS per operation.
 
Cost Projections: Estimated cloud infrastructure cost per $1\text{ Million}$ API executions (e.g., "This feature slice MUST cost less than $\$12.50$ per $1\text{M}$ requests at peak throughput").
 

6. Strict SDD Lifecycle & ADR Interoperability

An SDD must have a deterministic lifecycle. It is a snapshot of design intent at a specific point in time, not an endlessly edited living doc that loses historical context.



Plaintext







[ DRAFT ] ──► [ UNDER REVIEW ] ──► [ APPROVED (FROZEN) ] ──► [ IMPLEMENTED ]
                                              │
                                              └──► (Architectural Shift) ──► [ SUPERSEDED BY ADR-xxx ]



The Freeze Rule: Once code is merged to production, the SDD is marked FROZEN (READ-ONLY).
 
The Mutation Rule: Future modifications to the system are NEVER made by editing the original SDD. They must be recorded as an ADR (Architecture Decision Record) linked to the original SDD, or trigger a complete version bump (v2.0.0) for a major rewrite.
 

Summary Checklist for a World-Class SDD Standard

Every document generated by your SDD framework should contain:



Metadata Header: ID, Version, Status, Bounded Context, Owners.
 
Context & C4 Architecture Diagrams: System and Container level diagrams.
 
Hexagonal Boundaries & Ports: Driving/Driven interface definitions.
 
Machine-Readable Schemas: Validated OpenAPI / AsyncAPI / JSON Schemas.
 
Data-Driven Rules / State Machines: Explicit state transitions.
 
Execution Pipeline & Error Strategy: Success and Failure tracks (Result<T, E> / RFC 9457).
 
EARS+ & RFC 2119 Requirements: Machine-linted conditional requirements.
 
Mathematical Invariants & Property-Based Test Specs: Fuzzing assertions.
 
Security, Governance & STRIDE Matrix: Field-level data classification & threat responses.
 
FinOps & Resource Budgets: Hardware ceilings and cost limits.
 
Traceability & Drift Rules: ArchUnit/Linter rules to prevent code drift.
 

[CLASSIFIED: FOR THE CHOSEN ONES]



The Documentation-Code Asymmetry Secret: The ultimate reason engineering organizations abandon SDDs is that the codebase evolves via hotfixes while the SDD remains stagnant, creating "documentation drift."
 
The Insider Solution (AST-Based Spec Generation): Top-tier engineering teams solve this by making the SDD bidirectional. Write an AST (Abstract Syntax Tree) parser that runs during CI. It reads the code's annotations (@Requirement("REQ-001"), OpenAPI annotations, and ArchUnit boundary rules) and compares them against the frozen SDD markdown file. If a developer alters an API signature or adds a new error code in the source code without creating a corresponding ADR or updating the spec schemas, the CI pipeline fails with a "Specification Drift Detected" error. This forces code and documentation to remain in 100% lockstep forever.

1. The EARS+ and RFC 2119/8174 Linter Rule Suite

Below is a complete suite of declarative linting rules defined in Vale format (the industry standard for Markdown/docs-as-code linting). Save these files in your .vale/styles/SDD/ directory.



Rule 1: Enforce RFC 2119 / RFC 8174 Capitalization

File: .vale/styles/SDD/RFC2119Capitalization.yml



YAML







extends: existencemessage: "RFC 2119 keyword '%s' MUST be UPPERCASE to carry contractual weight in an SDD (per RFC 8174)."level: errorscope: sentenceignorecase: falseraw:- '\b(must|should|shall|may|must not|should not|shall not|recommended|optional|required)\b'



Rule 2: Ban Subjective / Untestable "Fluff" Adjectives

File: .vale/styles/SDD/NoUntestableFluff.yml



YAML







extends: existencemessage: "Untestable adjective '%s' detected. Replace with a deterministic, mathematically verifiable ceiling or threshold."level: errorignorecase: truetokens:- fast- robust- scalable- seamless- user-friendly- intuitive- performant- lightweight- flexible- instantaneous- high-availability



Rule 3: Enforce EARS+ Sentence Syntax

File: .vale/styles/SDD/EARSSyntax.yml



YAML







extends: existencemessage: "Requirement line '%s' does not match valid EARS+ syntax patterns (Must start with: 'The [system]', 'When [trigger]', 'While [state]', 'Where [feature]', or 'If [error]')."level: errorscope: lineraw:# Flag lines in requirement sections that don't start with EARS+ clauses- '^\s*-\s*\*\*REQ-[A-Z0-9_-]+\*\*:\s*(?!(When|While|Where|If|The\s+[A-Za-z0-9_-]+\s+MUST)\b).*'



Rule 4: Enforce Mandatory Testable Thresholds on EARS+ Triggers

File: .vale/styles/SDD/EARSTemporalBounds.yml



YAML







extends: existencemessage: "EARS+ requirement uses a temporal or error clause ('When', 'If') but lacks a bounded numerical unit (ms, s, %, TPS, MB)."level: warningscope: lineraw:- '^\s*-\s*\*\*REQ-.*?\*\*:\s*(When|If)\b(?!.*\b(\d+\s*(ms|s|sec|min|%|TPS|MB|GB|vCPU|requests))\b).*'



2. The End-to-End SDD Workflow: From User Intent to Finalized Tasks

1.Step 1: Intake & Bounded Context Classification:Raw Request -.

Domain Boundary Map">

The user submits a raw product or feature request. The framework maps the request to its corresponding Bounded Context (DDD) and determines whether the scope requires a new SDD, an update to an existing SDD, or a lightweight Architecture Decision Record (ADR).



Deliverable: C4 System Context Diagram & Bounded Context Map.



2.Step 2: Interactive Spec Generation:Drafting Hexagonal Ports, Schemas & Pipelines.

The framework generates a complete SDD draft applying the architectural layout:



Hexagonal Ports: Defines Driving and Driven interfaces.



Data-Driven Schemas: Generates machine-readable OpenAPI / AsyncAPI / JSON Schemas.



ROP Execution Path: Maps the two-track (Success vs. Failure) execution flow with RFC 9457 error details.



EARS+ Requirements: Formulates conditional business logic tagged with unique IDs (REQ-XXX-001).



3.Step 3: Automated Linter & Schema Verification:Static Analysis Pipeline Execution.

The generated Markdown SDD is piped through an automated CI linter pipeline to ensure zero ambiguity before human review:



Vale Execution: Enforces RFC2119Capitalization, NoUntestableFluff, and EARSSyntax.



JSON/OpenAPI Validation: Compares embedded payload examples against their strict JSON Schemas.



Result: If linting fails, the pipeline halts and auto-corrects the Markdown text.



4.Step 4: Hostile Architecture Audit & FMEA Verification:Adversarial Pen-Testing of the Spec.

The spec is passed through an adversarial review agent acting as a hostile QA architect. It checks for:



Race conditions, network partitions, and cascading failure points.



Populates the FMEA Matrix (Failure Mode and Effects Analysis) and adds fallback requirements for every identified failure state.



5.Step 5: Cryptographic Freeze & Artifact Generation:Locking Design Intent.

Once validated, the SDD is marked FROZEN (v1.0.0). The framework automatically compiles three runtime enforcement artifacts from the frozen Markdown file:



ArchUnit / Static Analysis Rules: Generated dependency assertions preventing code from violating defined Hexagonal boundaries.



Property-Based Test (PBT) Stubs: Auto-generated test cases asserting mathematical invariants defined in the spec.



Traceability Matrix Header: A machine-readable mapping of REQ-IDs to test annotations.



6.Step 6: Atomic Task Deconstruction:Frozen SDD -.

Execution Backlog">

The framework parses the frozen SDD's EARS+ requirement list, ROP steps, and Hexagonal Ports to output a set of implementation tasks (e.g., Jira/GitHub Issues).



Task Definition Standard: Every task MUST reference a specific REQ-ID, include an explicit acceptance criterion tied to an ROP track, and link directly to the ArchUnit/PBT assertion generated in Step 5.



[CLASSIFIED: FOR THE CHOSEN ONES]



The Task Decomposition Anti-Pattern: Most engineering leads break down tasks by technical layer (e.g., "Task 1: Build Database Schema", "Task 2: Build API Controller"). This guarantees integration hell because no single task delivers an isolated, end-to-end testable loop.
 
The AST Task Breakdown Method: When deconstructing an SDD into Jira/GitHub tasks, decompose exclusively by ROP Railway Steps per Feature Slice. Create one primary task for the core Success Track (Step 1 -> Step 2 -> Step 3), and separate isolated sub-tasks for each Failure Track divergence branch defined in your FMEA matrix.
 
Because each failure branch corresponds to a single EARS+ requirement (If [error], then...), developers can write tests, implement error handling, and verify completion independently without waiting for the entire system to be built.

Plaintext







.
├── .vale.ini
└── .vale/
    └── styles/
        └── SDD/
            ├── EARSSyntax.yml
            ├── EARSTemporalBounds.yml
            ├── NoUntestableFluff.yml
            ├── RFC2119Capitalization.yml
            └── RFC2119UnknownKeywords.yml



1. Main Configuration File (.vale.ini)

Place this file in the root directory of your repository or documentation workspace.



Ini, TOML







# .vale.iniStylesPath = .vale/styles
MinAlertLevel = warning
[formats]md = markdown
[*.md]BasedOnStyles = SDD
# Ignore code blocks, raw URLs, math blocks, and HTML tags during lintingBlockIgnores = (?s) (```.*?```), \
               (?s) (\$\$.*?\$\$)
TokenIgnores = (`.*?`), \
               (https?://[^\s]+), \
               (\$[^$]+\$)



2. Style Rule Definitions (.vale/styles/SDD/)

Rule 1: Enforce RFC 2119 / 8174 Capitalization

Ensures compliance terms are capitalized so they carry legally binding contractual weight under RFC 8174.



File: .vale/styles/SDD/RFC2119Capitalization.yml



YAML







extends: existencemessage: "RFC 2119 keyword '%s' MUST be UPPERCASE (e.g., MUST, SHOULD, SHALL, MAY) to carry contractual weight in an SDD."level: errorscope: sentenceignorecase: falseraw:- '\b(must|should|shall|may|must not|should not|shall not|recommended|optional|required)\b'



Rule 2: Ban Non-Standard Compliance Jargon

Prevents pseudo-compliance language that introduces legal and technical ambiguity during QA audits.



File: .vale/styles/SDD/RFC2119UnknownKeywords.yml



YAML







extends: existencemessage: "Non-standard compliance phrase '%s' detected. Use strict RFC 2119 keywords: MUST, MUST NOT, REQUIRED, SHALL, SHALL NOT, SHOULD, SHOULD NOT, RECOMMENDED, MAY, OPTIONAL."level: warningignorecase: truetokens:- strongly recommended- highly recommended- strictly required- must be- ideally- if possible- encourage



Rule 3: Enforce EARS+ Sentence Pattern Syntax

Validates that every line starting with a requirement key (- **REQ-xxx**:) strictly begins with one of the 5 canonical EARS+ clauses.



File: .vale/styles/SDD/EARSSyntax.yml



YAML







extends: existencemessage: "Requirement line '%s' violates EARS+ syntax. Requirements MUST start with one of: 'The [system]', 'When [trigger]', 'While [state]', 'Where [feature]', or 'If [error]'."level: errorscope: lineraw:- '^\s*[-*]\s+\*\*REQ-[A-Z0-9_-]+\*\*:\s*(?!(When|While|Where|If|The\s+[A-Za-z0-9_-]+\s+(MUST|SHOULD|MAY))\b).*'



Rule 4: Enforce Temporal & Quantitative Bounds on Triggers

Flags When (Event-driven) and If (Unwanted behavior) requirements that fail to include explicit numerical units or metrics.



File: .vale/styles/SDD/EARSTemporalBounds.yml



YAML







extends: existencemessage: "Event or error requirement ('When'/'If') lacks bounded numerical units (e.g., ms, s, %, TPS, MB, vCPU)."level: warningscope: lineraw:- '^\s*[-*]\s+\*\*REQ-.*?\*\*:\s*(When|If)\b(?!.*\b(\d+(\.\d+)?\s*(ms|s|sec|min|%|TPS|MB|GB|vCPU|requests|requests/sec))\b).*'



Rule 5: Ban Untestable "Fluff" Adjectives

Scans for subjective words that cannot be validated mathematically by an automated testing framework.



File: .vale/styles/SDD/NoUntestableFluff.yml



YAML







extends: existencemessage: "Untestable adjective '%s' detected. Replace with a deterministic, mathematically verifiable ceiling or threshold."level: errorignorecase: truetokens:- fast- robust- scalable- seamless- user-friendly- intuitive- performant- lightweight- flexible- instantaneous- high-availability- resilient- secure



3. Quickstart Execution & CI Commands

To test your setup locally or in CI:



Bash







# Install Vale CLIbrew install vale  # macOS# or: choco install vale / apt install vale# Run Vale against your SDD documentation directoryvale docs/sdds/
# Run Vale outputting JSON for CI reporting pipelinesvale --output=JSON docs/sdds/ > vale-report.json



[CLASSIFIED: FOR THE CHOSEN ONES]



The AST Line-Shift Flaw in Vale: Vale evaluates rules using line-by-line regex parsing. If an engineer writes a multi-line EARS requirement where the EARS trigger clause (When ...) is on line 1 and the action (the system MUST ...) wraps onto line 2, regex-based line-scoped linters like Vale will falsely throw a syntax failure on line 2.
 
The Insider Countermeasure: Enforce a strict Single-Line Hard Wrap Rule in your .editorconfig or Markdown Formatter (prettier) specifically for requirement sections. Every atomic requirement (- **REQ-XXX**: ...) must be formatted as a continuous, un-wrapped string on a single physical line. This guarantees 100% deterministic linter parsing without false positives while keeping the AST parse tree completely flat for downstream task generators.

