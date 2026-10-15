# Architectural Patterns & Methodologies: Comprehensive Reference

**Document Purpose:** This supplementary guide extends the foundational architectural concepts with modern best practices, implementation patterns, and advanced integration strategies for building robust, scalable software systems.

---

## Table of Contents

1. [Feature-Sliced Design (FSD) Deep Dive](#1-feature-sliced-design-fsd-deep-dive)
2. [Railway-Oriented Programming (ROP) Advanced Patterns](#2-railway-oriented-programming-rop-advanced-patterns)
3. [EARS Notation: Complete Reference](#3-ears-notation-complete-reference)
4. [Hexagonal Architecture & Plugin Systems](#4-hexagonal-architecture--plugin-systems)
5. [Data-Driven Architecture Patterns](#5-data-driven-architecture-patterns)
6. [arc42 Documentation Framework](#6-arc42-documentation-framework)
7. [Anti-Corruption Layer (ACL) Implementation](#7-anti-corruption-layer-acl-implementation)
8. [Domain-Driven Design Strategic Patterns](#8-domain-driven-design-strategic-patterns)
9. [C4 Model for Architecture Visualization](#9-c4-model-for-architecture-visualization)
10. [ISO/IEC/IEEE 29148 Requirements Quality Framework](#10-isoiecieee-29148-requirements-quality-framework)
11. [FMEA for Software Systems](#11-fmea-for-software-systems)
12. [RFC 9457: Standardized API Error Handling](#12-rfc-9457-standardized-api-error-handling)
13. [Integration Patterns & Anti-Patterns](#13-integration-patterns--anti-patterns)

---

## 1. Feature-Sliced Design (FSD) Deep Dive

### Core Philosophy

Feature-Sliced Design is an architectural methodology for scaffolding front-end applications. It's a compilation of rules and conventions for organizing code to make projects more understandable and stable in the face of ever-changing business requirements.

### Layer Hierarchy (Top to Bottom)

```
┌─────────────────────────────────────────┐
│  App                                     │  ← Application shell, routing, providers
├─────────────────────────────────────────┤
│  Processes (deprecated)                  │  ← Complex inter-page scenarios
├─────────────────────────────────────────┤
│  Pages                                   │  ← Full pages or large routing segments
├─────────────────────────────────────────┤
│  Widgets                                 │  ← Self-contained UI chunks, use cases
├─────────────────────────────────────────┤
│  Features                                │  ← Reusable product features
├─────────────────────────────────────────┤
│  Entities                                │  ← Business domain models
├─────────────────────────────────────────┤
│  Shared                                  │  ← Reusable infrastructure, no business logic
└─────────────────────────────────────────┘
```

### Critical Rules

1. **Unidirectional Dependency Flow**: Modules on one layer can ONLY import from layers strictly below.
2. **Slice Isolation**: Slices on the same layer CANNOT reference each other (ensures low coupling).
3. **Public API Encapsulation**: Each slice exposes only a public API (`index.ts`), hiding internal implementation.

### Standard Segment Types

- **ui** — UI components, styles, formatters
- **api** — Backend interactions, request functions, mappers
- **model** — Schemas, interfaces, stores, business logic
- **lib** — Library code internal to the slice
- **config** — Configuration files, feature flags

### Migration Strategy

1. Start with **App** and **Shared** layers to build foundation
2. Distribute existing UI across **Widgets** and **Pages** (even with violations)
3. Gradually resolve import violations while extracting **Entities** and **Features**
4. Refactor incrementally, not all at once

### Advantages

| Benefit | Description |
|---------|-------------|
| **Uniformity** | Standardized structure across projects accelerates onboarding |
| **Stability** | Isolated modifications prevent unforeseen side effects |
| **Controlled Reuse** | Balance between DRY principle and practical locality |
| **Business Orientation** | Code organized by domain, not technical layers |

### When to Adopt FSD

- You're building an **application**, not a library
- Project has grown too large and interconnected
- Expecting significant team growth
- Current architecture is causing maintenance pain

**Source:** [Feature-Sliced Design Official Documentation](https://feature-sliced.design/)

---

## 2. Railway-Oriented Programming (ROP) Advanced Patterns

### The Two-Track Metaphor

ROP visualizes data flow as parallel railway tracks:

```
SUCCESS TRACK
Data ──► [Step 1] ──► [Step 2] ──► [Step 3] ──► Output
           │             │             │
           ▼             ▼             ▼
FAILURE TRACK
Error ◄───┴─────────────┴─────────────┴─── Error
```

### Core Techniques

1. **Bind (>>=)** — Integrate monadic functions into the pipeline
2. **Kleisli Composition (>=>)** — Compose monadic functions
3. **Map (fmap)** — Integrate non-monadic functions
4. **Tee** — Integrate side-effect functions (logging, auditing)
5. **Double Map (&&&)** — Combine monadic functions in parallel (validation)

### Error Handling Strategy

Instead of exceptions, ROP uses **Result<T, E>** types:

```typescript
type Result<T, E> = 
  | { status: 'success'; value: T }
  | { status: 'failure'; error: E };
```

### Anti-Patterns to Avoid

1. **Over-engineering simple flows** — Not every function needs ROP
2. **Exception mixing** — Never mix exceptions with ROP in the same pipeline
3. **Error type explosion** — Use a unified error type hierarchy, not scattered types
4. **Ignoring the "tee" pattern** — Side effects should be explicit, not hidden

### When ROP Shines

- Complex validation chains with multiple failure modes
- Data transformation pipelines (ETL, parsing, serialization)
- API request/response handling with detailed error reporting
- Long-running workflows with checkpoint recovery

**Warning:** ROP is a powerful pattern, but don't take it to extremes. Simple error handling doesn't require monadic pipelines.

**Source:** [Railway Oriented Programming - F# for Fun and Profit](https://fsharpforfunandprofit.com/rop/)

---

## 3. EARS Notation: Complete Reference

### The Six Patterns

| Pattern | Syntax | Use Case | Example |
|---------|--------|----------|---------|
| **Ubiquitous** | `THE <System> SHALL <response>` | Always applies | THE System SHALL encrypt all stored passwords |
| **Event-Driven** | `WHEN <trigger>, THE <System> SHALL <response>` | Triggered by event | WHEN a user submits the form, THE System SHALL validate the input |
| **State-Driven** | `WHILE <condition>, THE <System> SHALL <response>` | Depends on state | WHILE the system is in maintenance mode, THE System SHALL reject all write operations |
| **Unwanted Behavior** | `IF <condition>, THEN THE <System> SHALL <response>` | Error handling | IF the payment fails, THEN THE System SHALL retry up to 3 times |
| **Optional Features** | `WHERE <option>, THE <System> SHALL <response>` | Configurable | WHERE two-factor auth is enabled, THE System SHALL require TOTP verification |
| **Complex** | `[WHERE] [WHILE] [WHEN/IF] THE <System> SHALL <response>` | Multiple conditions | WHERE premium mode is active WHILE the user is authenticated WHEN the download button is clicked, THE System SHALL allow unlimited downloads |

### INCOSE Quality Rules (Must Follow)

#### Clarity Rules
1. ✅ Use **active voice** with explicit subject and action
2. ❌ No vague terms: "quickly", "adequately", "reasonable", "user-friendly"
3. ❌ No pronouns: "it", "them", "they"
4. ✅ Consistent terminology from defined Glossary

#### Testability Rules
1. ✅ All conditions must be **measurable or verifiable**
2. ✅ Timing requirements must specify tolerances (e.g., "within 200ms")
3. ✅ Each requirement tests **one thing only**
4. ✅ Criteria must be **specific and quantifiable**

#### Completeness Rules
1. ❌ No escape clauses: "where possible", "if feasible", "as appropriate"
2. ❌ No absolutes unless truly absolute: "never", "always", "100%"
3. ✅ Focus on **what**, not **how**
4. ✅ Self-contained and complete

#### Positive Statement Rule
1. ✅ Use "SHALL" for positive requirements
2. ❌ Avoid "SHALL NOT" except for error handling
3. ✅ Convert negatives to positives where possible
   - ❌ "The system SHALL NOT crash"
   - ✅ "The system SHALL handle errors gracefully and continue operation"

### EARS Validation Properties

For automated linting, verify:

```yaml
Property_1_Clarity:
  - No_vague_terms: true
  - No_pronouns: true
  - Active_voice: true

Property_2_Testability:
  - Measurable_conditions: true
  - Timing_tolerance_specified: true
  - Single_responsibility: true

Property_3_Completeness:
  - No_escape_clauses: true
  - Realistic_bounds: true
  - Self_contained: true
```

**Sources:** 
- [EARS Notation - Modern Requirements](https://www.modernrequirements.com/de/blogs/ears-notation-the-practical-guide/)
- [EARS Wikipedia](https://en.wikipedia.org/wiki/Easy_Approach_to_Requirements_Syntax)
- [Ten Years of EARS - ResearchGate](https://www.researchgate.net/publication/335535918_Ten_Years_of_EARS)

---

## 4. Hexagonal Architecture & Plugin Systems

### Core Concept

Hexagonal Architecture (Ports & Adapters) isolates core business logic from infrastructure concerns (databases, APIs, external services) using a clear boundary.

### Architecture Diagram

```
┌────────────────────────────────────────────────────────────┐
│  DRIVING ADAPTERS (Inbound)                                │
│  • HTTP REST API                                           │
│  • CLI Commands                                            │
│  • Message Queue Consumers                                 │
│  • Scheduled Jobs                                          │
└──────────────────────┬─────────────────────────────────────┘
                       │
                       ▼
┌────────────────────────────────────────────────────────────┐
│  DRIVING PORTS (Primary Interfaces)                        │
│  • UseCasePort interfaces                                  │
│  • Command/Query handlers                                  │
└──────────────────────┬─────────────────────────────────────┘
                       │
                       ▼
┌────────────────────────────────────────────────────────────┐
│  CORE DOMAIN (Business Logic)                              │
│  • Domain Models (Entities, Value Objects)                 │
│  • Business Rules                                          │
│  • Domain Services                                         │
│  • Application Services                                    │
└──────────────────────┬─────────────────────────────────────┘
                       │
                       ▼
┌────────────────────────────────────────────────────────────┐
│  DRIVEN PORTS (Secondary Interfaces)                       │
│  • RepositoryPort interfaces                               │
│  • ExternalServicePort interfaces                          │
│  • EventEmitterPort interfaces                             │
└──────────────────────┬─────────────────────────────────────┘
                       │
                       ▼
┌────────────────────────────────────────────────────────────┐
│  DRIVEN ADAPTERS (Outbound)                                │
│  • PostgreSQL Repository                                   │
│  • REST API Client                                         │
│  • Message Queue Publisher                                 │
│  • File System Adapter                                     │
└────────────────────────────────────────────────────────────┘
```

### Port Types

**Driving Ports (Primary)** — Called by external actors
- Define use cases the application supports
- Express in **application language**, not infrastructure terms
- Example: `CreateOrderUseCasePort`, `ProcessPaymentPort`

**Driven Ports (Secondary)** — Called by the application
- Define dependencies the application needs
- Express in **domain language**
- Example: `OrderRepositoryPort`, `PaymentGatewayPort`

### Plugin Architecture Integration

Plugins implement **Driven Ports** as **Adapters**:

```typescript
// Port (in Core)
interface PaymentGatewayPort {
  processPayment(order: Order): Promise<Result<Payment, PaymentError>>;
}

// Adapter (Plugin - loaded dynamically)
class StripePaymentAdapter implements PaymentGatewayPort {
  async processPayment(order: Order): Promise<Result<Payment, PaymentError>> {
    // Stripe-specific implementation
  }
}

// Adapter (Another plugin - interchangeable)
class PayPalPaymentAdapter implements PaymentGatewayPort {
  async processPayment(order: Order): Promise<Result<Payment, PaymentError>> {
    // PayPal-specific implementation
  }
}
```

### Key Benefits

| Benefit | Description |
|---------|-------------|
| **Testability** | Core logic tested in isolation with mock adapters |
| **Flexibility** | Swap databases, APIs, frameworks without touching core |
| **Maintainability** | Clear separation prevents logic leakage |
| **Technology Agnostic** | Core doesn't know about HTTP, SQL, or message queues |

**Sources:**
- [Hexagonal Architecture - AWS Prescriptive Guidance](https://docs.aws.amazon.com/prescriptive-guidance/latest/cloud-design-patterns/hexagonal-architecture.html)
- [Hexagonal Architecture Explained - ThoughtWorks](https://www.thoughtworks.com/en-gb/insights/blog/architecture/hexagonal-architecture-explained-practical-example)

---

## 5. Data-Driven Architecture Patterns

### Definition

Data-driven architecture means the structure and behavior of systems are determined by **data** (configuration, schemas, metadata, rules) rather than hardcoded logic. The code is generic; the data defines what actually happens.

### Configuration-Driven Pipeline Pattern

```json
{
  "pipeline_id": "transform-user-data",
  "stages": [
    {
      "name": "validate",
      "operator": "SchemaValidator",
      "parameters": { "schema": "user.schema.json" }
    },
    {
      "name": "transform",
      "operator": "FieldMapper",
      "parameters": { 
        "mappings": [
          { "from": "user_name", "to": "fullName" },
          { "from": "email_address", "to": "email" }
        ]
      }
    },
    {
      "name": "enrich",
      "operator": "GeoEnricher",
      "parameters": { "api_endpoint": "https://geo.service/lookup" }
    }
  ]
}
```

### Core Principles

1. **Generic Engine** — Code doesn't know specific business rules
2. **Configuration as Code** — Pipelines defined in JSON/YAML, version-controlled
3. **Runtime Evaluation** — Engine interprets configuration at runtime
4. **Extensibility** — Add new operators without changing core engine

### Advantages Over Hard-Coded Logic

| Aspect | Hard-Coded | Data-Driven |
|--------|------------|-------------|
| **Changeability** | Requires code change, rebuild, redeploy | Update config, restart |
| **Testability** | Unit tests per scenario | Test engine + validate configs |
| **Reusability** | Copy-paste code | Reuse operators, compose pipelines |
| **Auditability** | Hard to trace changes | Config diffs show exactly what changed |
| **Non-Dev Updates** | Requires engineering | Business users can modify config |

### Anti-Pattern: The "Config Monster"

Don't turn everything into configuration. Data-driven doesn't mean "no code." Balance:
- **Code**: Stable, complex business logic
- **Config**: Variable, frequently changing parameters, pipeline definitions

**Sources:**
- [Configuration-Driven Pipeline Design](https://dev3lop.com/blog/configuration-driven-pipeline-design-vs-hard-coded-logic/)
- [Data-Driven Architecture Patterns](https://hackernoon.com/data-driven-architecture-patterns-for-production)

---

## 6. arc42 Documentation Framework

### The 12 Sections

| Section | Purpose | Key Content |
|---------|---------|-------------|
| **1. Introduction and Goals** | Why are we building this? | Requirements, stakeholder needs, top quality goals |
| **2. Constraints** | What limits us? | Technical, organizational, conventions |
| **3. Context and Scope** | Where does it fit? | Business context, external interfaces, boundaries |
| **4. Solution Strategy** | How do we solve it? | Fundamental decisions, key patterns |
| **5. Building Block View** | What is it made of? | Static decomposition, components, dependencies |
| **6. Runtime View** | How does it behave? | Runtime scenarios, interactions, data flow |
| **7. Deployment View** | Where does it run? | Infrastructure, hardware, deployment |
| **8. Cross-Cutting Concepts** | What's orthogonal? | Logging, security, error handling, i18n |
| **9. Architecture Decisions** | What did we decide? | ADRs, trade-offs, rationale |
| **10. Quality Requirements** | How good is it? | Quality tree, scenarios, metrics |
| **11. Risks and Technical Debt** | What could go wrong? | Known issues, risks, mitigation |
| **12. Glossary** | What do words mean? | Domain terminology definitions |

### Documentation Philosophy

**Keywords:**
- **lean** — Pragmatic, lightweight, travel light
- **thorough** — Detailed, formal, auditable
- **essential** — Non-negotiable quality goals

### Building Block View Levels

- **Level 1**: System overview (black-box)
- **Level 2**: Major components (white-box decomposition)
- **Level 3**: Component internals (implementation details)

### Best Practices

1. Start with sections 1, 3, 5, 9 (highest value)
2. Keep it **alive** — update as system evolves
3. Use **diagrams** — text alone doesn't scale
4. Link to **source code** and **ADRs**
5. Don't over-document — match stakeholder needs

**Source:** [arc42 Documentation](https://docs.arc42.org/home/)

---

## 7. Anti-Corruption Layer (ACL) Implementation

### Purpose

The ACL isolates a system from external systems with different models, translating between semantics to protect the internal domain from corruption.

### When to Use ACL

- Migrating legacy systems incrementally
- Integrating with external APIs you don't control
- Bridging bounded contexts with different ubiquitous languages
- Preventing external model changes from breaking internal logic

### Implementation Pattern

```
┌──────────────────┐       ┌──────────────────┐       ┌──────────────────┐
│  External System │       │  Anti-Corruption │       │  Internal System │
│  (Legacy API)    │◄─────►│       Layer      │◄─────►│  (Clean Domain)  │
│  Model: {usr_id} │       │  • Translate     │       │  Model: {userId} │
│                  │       │  • Validate      │       │                  │
└──────────────────┘       │  • Isolate       │       └──────────────────┘
                           └──────────────────┘
```

### ACL Responsibilities

1. **Translation** — Convert external models to internal models
2. **Validation** — Ensure incoming data meets internal contracts
3. **Error Normalization** — Map external errors to internal error types
4. **Versioning Shield** — Absorb breaking changes from external systems
5. **Rate Limiting / Circuit Breaking** — Protect internal system from cascading failures

### Code Example

```typescript
// External model (from legacy API)
interface LegacyUserResponse {
  usr_id: string;
  usr_name_full: string;
  usr_email_addr: string;
  created_at_timestamp: number;
}

// Internal model (clean domain)
interface User {
  id: string;
  fullName: string;
  email: string;
  createdAt: Date;
}

// ACL Implementation
class UserACL {
  translate(legacy: LegacyUserResponse): Result<User, ValidationError> {
    return Result.ok({
      id: legacy.usr_id,
      fullName: legacy.usr_name_full,
      email: this.validateEmail(legacy.usr_email_addr),
      createdAt: new Date(legacy.created_at_timestamp)
    });
  }
  
  private validateEmail(email: string): string {
    // Validate and normalize email format
    if (!emailRegex.test(email)) {
      throw new ValidationError('Invalid email format from legacy system');
    }
    return email.toLowerCase();
  }
}
```

### Key Considerations

| Concern | Decision |
|---------|----------|
| **Performance** | ACL adds latency — measure and optimize hot paths |
| **Maintenance** | ACL is a service to maintain — consider retirement strategy |
| **Scaling** | May need multiple ACLs for different subsystems |
| **Observability** | Add correlation IDs, structured logging for translation failures |
| **Trust Boundary** | Enforce input validation and sanitization at ACL |

**Sources:**
- [Anti-Corruption Layer Pattern - Azure Architecture Center](https://learn.microsoft.com/en-gb/azure/architecture/patterns/anti-corruption-layer)
- [ACL Pattern - AWS Prescriptive Guidance](https://docs.aws.amazon.com/prescriptive-guidance/latest/cloud-design-patterns/acl.html)

---

## 8. Domain-Driven Design Strategic Patterns

### Bounded Contexts

A bounded context is a **conceptual boundary** around a part of the system where a particular **ubiquitous language** applies consistently.

```
┌─────────────────────────┐  ┌─────────────────────────┐
│  Sales Context          │  │  Inventory Context      │
│                         │  │                         │
│  Language:              │  │  Language:              │
│  - Customer             │  │  - Product              │
│  - Order                │  │  - Stock                │
│  - Payment              │  │  - Warehouse            │
│                         │  │                         │
│  Model:                 │  │  Model:                 │
│  - Order (with price)   │  │  - Product (with SKU)   │
│  - Customer (shipping)  │  │  - StockLevel (count)   │
└─────────────────────────┘  └─────────────────────────┘
         │                              │
         └──────────────┬───────────────┘
                        │
                        ▼
              ┌─────────────────┐
              │ Context Map     │
              │ (Relationships) │
              └─────────────────┘
```

### Context Map Patterns

| Pattern | Description | Use Case |
|---------|-------------|----------|
| **Shared Kernel** | Shared subset of model | Two teams need common data |
| **Customer/Supplier** | Upstream serves downstream | API provider/consumer |
| **Conformist** | Downstream conforms to upstream | No influence on external API |
| **Anti-Corruption Layer** | Translation layer | Protect internal model |
| **Open Host Service** | Public protocol | Third-party integrations |
| **Published Language** | Standard format | Industry standard (JSON:API, GraphQL) |

### Ubiquitous Language

A **shared, rigorous vocabulary** used by both developers and domain experts:

- Same term = same concept everywhere
- No synonyms (not "user" in one place, "customer" in another)
- Evolves with understanding
- Codified in code (class names, methods, variables)

### Domain Events

Events that domain experts care about:

```typescript
// Domain Event
interface OrderPlaced {
  eventId: string;
  orderId: string;
  customerId: string;
  totalAmount: Money;
  occurredAt: Date;
}
```

### CQRS (Command Query Responsibility Segregation)

Separate **write model** (commands) from **read model** (queries):

```
┌─────────────┐    Command    ┌─────────────┐    Event    ┌─────────────┐
│   Client    │──────────────►│  Write      │────────────►│  Read       │
│             │◄──────────────│  Model      │             │  Model      │
└─────────────┘   Acknowledg. └─────────────┘             └─────────────┘
                                     │                           │
                                     ▼                           ▼
                              ┌─────────────┐             ┌─────────────┐
                              │ Event Store │             │  Read DB    │
                              └─────────────┘             └─────────────┘
```

**Benefits:**
- Independent scaling (read-heavy vs write-heavy)
- Optimized read models (denormalized)
- Event sourcing capability

**Sources:**
- [Designing DDD Microservice - Microsoft](https://learn.microsoft.com/en-us/dotnet/architecture/microservices/microservice-ddd-cqrs-patterns/ddd-oriented-microservice)
- [Event-Driven Architecture with DDD](https://sevalla.com/blog/event-driven-architecture-ddd/)

---

## 9. C4 Model for Architecture Visualization

### Four Levels of Abstraction

```
Level 1: Context (System in its environment)
    ↓ Zoom in
Level 2: Container (Deployable units)
    ↓ Zoom in
Level 3: Component (Parts of containers)
    ↓ Zoom in
Level 4: Code (Implementation details)
```

### Level Descriptions

| Level | Name | Audience | Shows |
|-------|------|----------|-------|
| **1** | Context | Non-technical stakeholders | System, users, external systems |
| **2** | Container | Developers, ops | Apps, databases, services, message buses |
| **3** | Component | Developers | Controllers, services, repositories |
| **4** | Code | Developers | Classes, interfaces, functions |

### Example: E-Commerce System

**Level 1: Context**
```
┌─────────────┐              ┌─────────────┐
│   Customer  │─────────────►│             │
│  (Web User) │◄─────────────│             │
└─────────────┘              │   E-Commerce│
                             │   System    │
┌─────────────┐              │             │
│  Admin      │─────────────►│             │
│  (Staff)    │◄─────────────│             │
└─────────────┘              └──────┬──────┘
                                    │
                                    ▼
                             ┌─────────────┐
                             │   Payment   │
                             │   Gateway   │
                             └─────────────┘
```

**Level 2: Container**
```
┌──────────────────────────────────────────────────┐
│              E-Commerce System                    │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐       │
│  │ Web App  │  │ API      │  │ Database │       │
│  │ (React)  │  │ Service  │  │ (Postgres)│      │
│  └──────────┘  └──────────┘  └──────────┘       │
│                     │                             │
│                     ▼                             │
│               ┌──────────┐                       │
│               │ Redis    │                       │
│               │ (Cache)  │                       │
│               └──────────┘                       │
└──────────────────────────────────────────────────┘
```

### Best Practices

1. **Use consistent notation** — Same shapes and colors across levels
2. **Show relationships** — Arrows with labels describing interaction
3. **Keep it simple** — Don't show everything, focus on key elements
4. **Update diagrams** — Keep in sync with code
5. **Multiple views** — Different diagrams for different audiences

**Source:** [C4 Model - InfoQ](https://www.infoq.com/articles/C4-architecture-model/)

---

## 10. ISO/IEC/IEEE 29148 Requirements Quality Framework

### Nine Quality Criteria for Requirements

| Criterion | Definition | Verification |
|-----------|------------|--------------|
| **Necessary** | Must be needed, not just nice-to-have | Trace to stakeholder need |
| **Implementation-Free** | Describes what, not how | No technology names |
| **Unambiguous** | Single interpretation possible | Reviewed by multiple stakeholders |
| **Consistent** | No conflicts with other requirements | Cross-reference check |
| **Complete** | All relevant info included | Nothing left to assume |
| **Singular** | One requirement, one thing | Can't be split |
| **Feasible** | Can be implemented within constraints | Technical review |
| **Traceable** | Can be traced forward and backward | Unique ID, links |
| **Verifiable** | Can be tested/inspected | Test case exists |

### Requirement Attributes

```yaml
requirement:
  id: REQ-001
  type: functional
  priority: high
  status: approved
  source: stakeholder-interview-2024-03
  verification_method: test
  verification_criteria: "Response time < 200ms under 1000 concurrent users"
```

### Common Anti-Patterns

❌ **Vague terms**: "fast", "user-friendly", "robust", "scalable"  
❌ **Escape clauses**: "if possible", "where appropriate", "TBD"  
❌ **Compound requirements**: "The system shall X, Y, and Z"  
❌ **Missing acceptance criteria**: No way to verify completion  
❌ **Hidden assumptions**: Assumptions not documented  

**Source:** [ISO/IEC/IEEE 29148:2018](https://www.iso.org/standard/72089.html)

---

## 11. FMEA for Software Systems

### What is FMEA?

Failure Mode and Effects Analysis is a **bottom-up** technique that identifies potential failures, their causes, and their effects on the system.

### FMEA Matrix Structure

| Component | Failure Mode | Cause | Effect | Detection | RPN |
|-----------|-------------|-------|--------|-----------|-----|
| Payment Service | Timeout | Network latency | Order not processed | Monitoring alert | High |
| Database | Deadlock | Concurrent writes | Query failure | Query timeout | Medium |
| Auth Service | Token expiry | Clock drift | User logged out | 401 response | Low |

### Risk Priority Number (RPN)

```
RPN = Severity × Occurrence × Detection

Where:
- Severity: 1-10 (impact on system)
- Occurrence: 1-10 (likelihood)
- Detection: 1-10 (ability to detect before impact)
```

### FMEA Process

1. **Identify components** — Break down system into parts
2. **Identify failure modes** — How can each part fail?
3. **Determine causes** — What causes each failure?
4. **Analyze effects** — What happens when it fails?
5. **Rate severity, occurrence, detection** — Calculate RPN
6. **Prioritize mitigation** — Address highest RPNs first
7. **Document recovery actions** — Automated responses

### Software-Specific Failure Modes

| Category | Example Failure Modes |
|----------|----------------------|
| **Input** | Invalid data, missing fields, injection attacks |
| **Processing** | Exceptions, infinite loops, race conditions |
| **Output** | Incorrect format, missing data, encoding errors |
| **Infrastructure** | Network timeout, disk full, memory exhaustion |
| **External Dependencies** | API unavailable, rate limited, schema change |

### Automated Recovery Actions

```yaml
failure_mode: database_connection_timeout
detection: connection_pool_exhausted
recovery:
  - action: circuit_breaker_open
    delay: 30s
  - action: fallback_to_cache
  - action: alert_oncall_team
```

**Sources:**
- [FMEA - NASA Software Engineering Handbook](https://swehb.nasa.gov/spaces/SWEHBVC/pages/72024562/8.05+-+SW+Failure+Modes+and+Effects+Analysis)
- [FMEA - ASQ](https://asq.org/quality-resources/fmea)

---

## 12. RFC 9457: Standardized API Error Handling

### Overview

RFC 9457 (obsoletes RFC 7807) defines a standard **`application/problem+json`** media type for HTTP API error responses.

### Standard Fields

```json
{
  "type": "https://errors.example.com/insufficient-funds",
  "title": "Insufficient Funds",
  "status": 400,
  "detail": "Your account balance is $50.00, but the requested transfer amount is $100.00.",
  "instance": "/transactions/abc123",
  "balance": 50.00,
  "requested_amount": 100.00
}
```

### Required vs. Optional Fields

| Field | Required | Type | Description |
|-------|----------|------|-------------|
| `type` | Yes | URI | Problem type identifier |
| `title` | Yes | String | Short human-readable summary |
| `status` | Yes | Number | HTTP status code |
| `detail` | No | String | Human-readable explanation |
| `instance` | No | URI | Specific occurrence URI |

### Custom Extensions

Add domain-specific fields:

```json
{
  "type": "https://api.example.com/errors/validation-failed",
  "title": "Validation Failed",
  "status": 422,
  "detail": "The request body contains invalid fields.",
  "errors": [
    {
      "field": "email",
      "message": "Must be a valid email address"
    },
    {
      "field": "age",
      "message": "Must be at least 18"
    }
  ]
}
```

### Problem Type Registry

Use registered problem types when available:
- **IANA Registry**: Standard problem types
- **Organization Registry**: Custom organizational types

### Implementation Best Practices

1. **Use `type` as primary identifier** — Clients match on type, not title
2. **Keep `title` stable** — Don't change it; it's an API contract
3. **Include enough `detail`** — But don't expose sensitive info
4. **Add custom fields** — For machine-readable error handling
5. **Use URIs, not strings** — `type` should resolve to documentation

**Source:** [RFC 9457 - IETF](https://datatracker.ietf.org/doc/html/rfc9457)

---

## 13. Integration Patterns & Anti-Patterns

### Common Anti-Patterns

#### 1. The "Monadic Leakage" Trap

**Problem:** Passing high-level generic types (Result<T, E>, Either<L, R>) across plugin boundaries causes ABI issues and memory corruption.

**Solution:** Serialize to flat C-ABI structures:

```c
typedef struct {
  uint32_t status_code;      // 0 = success, >0 = error
  uint32_t payload_length;
  const uint8_t* payload;    // Serialized data or error JSON
} PluginResponse;
```

#### 2. The "ROP Memory Trap" in GC Languages

**Problem:** Allocating new Result<T, E> objects for every validation failure destroys GC throughput under load.

**Solution:** Pre-instantiate static, immutable singleton error objects:

```typescript
// BAD: Creates new object each time
return Result.err(new ValidationError("Invalid email"));

// GOOD: Reuse singleton
return Result.err(StaticErrors.INVALID_EMAIL);
```

#### 3. The "Ghost State" in Async Workflows

**Problem:** System pays for API call, AI generates, but network blip causes missed webhook — UI spins forever.

**Solution:** Reconciliation Worker (CRON job) that polls external services to true-up orphaned jobs.

#### 4. The "Coordinate System Impedance Mismatch"

**Problem:** Blender uses Z-up right-handed, WebGL uses Y-up right-handed, game engines use Y-up left-handed.

**Solution:** Mandate "Coordinate Space Transformation Strategy" in SDD:

```typescript
interface CoordinateSystem {
  up: 'Y' | 'Z';
  handedness: 'left' | 'right';
}

function transform(source: CoordinateSystem, target: CoordinateSystem, mesh: Mesh): Mesh;
```

### The "Alternatives Considered" Test

The most critical section of any design document:

> If an engineer hands you an RFC and "Alternatives Considered" is short, vague, or missing — **reject the document**.

**Why:** True engineering is the **rigorous elimination of flawed alternatives**. If they can't articulate why three other approaches would fail, they don't understand the problem deeply enough.

---

## [CLASSIFIED: FOR THE CHOSEN ONES]

### The Real Purpose of Documentation

The primary value of an RFC or design document is **not the document itself**, but the **act of writing it**. It forces the author to slow down and aggressively rubber-duck their own untested assumptions.

### Hidden Technical Debt Indicators

1. **Documentation drift** — Comments in code don't match reality
2. **Tribal knowledge** — "Ask Bob, he knows how it works"
3. **Fear of refactoring** — "Nobody touch that module"
4. **Test coverage gaps** — Complex logic with no tests
5. **Configuration sprawl** — Env vars scattered everywhere

### System Resilience Secrets

1. **Never trust external webhooks** — Always have reconciliation workers
2. **Circuit breakers everywhere** — Fail fast, don't wait for timeouts
3. **Idempotency keys** — Retries must be safe
4. **Correlation IDs** — Trace requests across services
5. **Graceful degradation** — System works even when parts fail

### The Architecture Review Checklist

When reviewing any design document, verify:

- [ ] **Ubiquitous language defined** — All terms in glossary
- [ ] **Bounded contexts drawn** — Clear boundaries
- [ ] **Failure modes analyzed** — FMEA completed
- [ ] **Trade-offs explicit** — Alternatives considered
- [ ] **Quality attributes stated** — Performance, security, reliability
- [ ] **Observability planned** — Logging, metrics, tracing
- [ ] **Migration strategy exists** — How to get from here to there

---

**Document Version:** 1.0.0  
**Last Updated:** September 27, 2026  
**Authors:** Principal Architecture Team  
**Status:** ENRICHED
