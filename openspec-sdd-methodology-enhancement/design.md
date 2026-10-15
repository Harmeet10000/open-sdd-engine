# Design Document: Subagent-Native Spec-Driven Development Enhancement

## Overview

This document describes the technical design for enhancing the OpenSpec Spec-Driven Development (SDD) methodology with a subagent-native architecture. The enhancement introduces specialized agents that coordinate via a shared message board, enabling autonomous requirements analysis, RFC-style design generation, hostile QA audits, and git integration while maintaining compatibility with existing OpenSpec CLI workflows.

The key architectural decisions are:

- **Decentralized Orchestration**: No monolithic orchestrator; specialized subagents coordinate via message board
- **Message-Driven Architecture**: Async communication using structured message formats
- **EARS Methodology Integration**: Ubiquitous, Event-driven, State-driven, Unwanted behavior, Optional features
- **RFC-Style Design Documents**: Standardized format with requirements, architecture, trade-offs
- **Hostile QA Review**: Automatic adversarial review with severity-based findings
- **Git Integration**: Branch management, commits per artifact, PR generation

## Architecture

### Architecture Overview

The enhanced SDD system follows a message-driven architecture with specialized subagents:

```
┌─────────────────────────────────────────────────────────────┐
│                  Message Board (Shared State)               │
│  • Artifact Store: proposal.md, spec.md, design.md, etc.    │
│  • Agent Registry: active agents, capabilities              │
│  • Status Tracker: workflow phase, completion state         │
│  • Message Queue: async coordination messages               │
└─────────────────────────────────────────────────────────────┘
                            ▲
                            │
        ┌───────────────────┴───────────────────┐
        │                                       │
┌───────▼────────┐              ┌──────────────▼──────────────┐
│ Refinement     │              │  Specialized Subagents      │
│ Coordinator    │              │                             │
│ • Workflow     │              │ Requirements Analyst        │
│   orchestration│              │   • EARS template analysis  │
│ • Agent        │              │   • Requirement extraction  │
│   selection    │              │   • Scenario generation     │
│ • Feedback     │              │                             │
│   integration  │              │ Architect                   │
│                │              │   • RFC-style design        │
│                │              │   • Architecture decisions  │
│                │              │   • Trade-off analysis      │
│                │              │                             │
│                │              │ QA Auditor                  │
│                │              │   • Adversarial review      │
│                │              │   • Edge case discovery     │
│                │              │   • Severity classification │
│                │              │                             │
│                │              │ Git Integrator              │
│                │              │   • Branch management       │
│                │              │   • Commit history          │
│                │              │   • PR generation           │
│                │              │                             │
│                │              │ Customizer                  │
│                │              │   • Config parsing          │
│                │              │   • Template application    │
│                │              │   • Methodology customization│
│                │              │                             │
│                │              │ Validator                   │
│                │              │   • Spec format validation  │
│                │              │   • Rule compliance check   │
│                │              │   • Consistency verification│
│                │              │                             │
│                │              │ Refinement Coordinator      │
│                │              │   • Loop control            │
│                │              │   • Feedback integration    │
│                │              │   • Refinement orchestration│
└─────────────────────────────────────────────────────────────┘
                            │
        ┌───────────────────┴───────────────────┐
        │                                       │
┌───────▼────────┐              ┌──────────────▼──────────────┐
│    Git         │              │    OpenSpec CLI             │
│  Repository    │              │  (External Interface)       │
│  • Branches    │              │  • validate command         │
│  • Commits     │              │  • apply command            │
│  • PRs         │              │  • view command             │
└────────────────┘              └─────────────────────────────┘
                            │
        ┌───────────────────┴───────────────────┐
        │                                       │
┌───────▼────────┐              ┌──────────────▼──────────────┐
│    Config      │              │    OpenSpec CLI Skills      │
│  (config.yaml) │              │  (Entrypoint Skills)        │
│  • Schema      │              │  • openspec-new-change      │
│  • Templates   │              │  • openspec-propose         │
│  • Rules       │              │  • openspec-apply-change    │
│  • Methods     │              │  • openspec-archive-change  │
└────────────────┘              └─────────────────────────────┘
```

### Component Responsibilities

**Message Board**: Central coordination layer providing shared state, async messaging, agent registration, and status tracking. Acts as the single source of truth for artifacts and workflow progress.

**Refinement Coordinator**: Main entry point for SDD workflow orchestration. Parses user requests, selects appropriate specialized agents, manages feedback loops, and coordinates with OpenSpec CLI for validation/apply phases.

**Specialized Subagents**: Each agent has a dedicated responsibility with clear interfaces and message contracts. Agents communicate via the message board rather than direct calls.

**OpenSpec CLI Skills**: Existing entrypoint skills (openspec-new-change, openspec-propose, openspec-apply-change, openspec-archive-change) provide the external interface and command-line functionality.

### Data Flow

#### Request Processing Flow

```
1. User submits feature request via OpenSpec skill
2. Refinement Coordinator receives request
3. Parse user request + detect language/style requirements
4. Publish workflow_started message to message board
5. Delegate to specialized agents in phase order:
   - Requirements Analyst → Validator
   - Architect → QA Auditor → Validator
   - Task Generation → Validator
   - Git Integrator
6. Each agent publishes progress messages
7. Feedback loops handled via message board
8. All artifacts committed to git branch
9. PR generated for review
```

#### Agent Coordination Flow

```
Agent A → Message Board: publish("agent_task", {task, payload})
Message Board → Agent B: subscribe("agent_task")
Agent B → Message Board: publish("task_complete", {taskId, result})
Message Board → Agent A: notify("task_complete", {taskId, result})
```

## Components and Interfaces

### Component 1: Message Board

#### Purpose
Central coordination layer providing shared state, async messaging, agent registration, and status tracking.

#### Interface

```python
from __future__ import annotations

import asyncio
from dataclasses import dataclass
from datetime import datetime
from enum import Enum
from typing import Any, Protocol


class MessageType(str, Enum):
    WORKFLOW_STARTED = "workflow_started"
    WORKFLOW_COMPLETED = "workflow_completed"
    AGENT_TASK = "agent_task"
    TASK_COMPLETE = "task_complete"
    ARTIFACT_CREATED = "artifact_created"
    ARTIFACT_UPDATED = "artifact_updated"
    VALIDATION_RESULT = "validation_result"
    AUDIT_FINDING = "audit_finding"
    FEEDBACK_REQUEST = "feedback_request"
    FEEDBACK_RESPONSE = "feedback_response"
    AGENT_REGISTERED = "agent_registered"
    AGENT_DEREGISTERED = "agent_deregistered"
    STATUS_UPDATE = "status_update"
    ERROR = "error"


@dataclass(frozen=True)
class Message:
    """Message format for inter-agent communication."""
    
    id: str
    type: MessageType
    timestamp: datetime
    sender_id: str
    payload: dict[str, Any]
    correlation_id: str | None = None  # For request/response chains
    priority: int = 0  # Higher priority processed first


class MessageBoardProtocol(Protocol):
    """Protocol for message board implementations."""
    
    async def publish(self, message: Message) -> None:
        """Publish a message to all subscribers."""
        ...
    
    async def subscribe(self, message_type: MessageType) -> asyncio.Queue[Message]:
        """Subscribe to a message type, returns a queue for consumption."""
        ...
    
    async def register_agent(self, agent_id: str, capabilities: list[str]) -> None:
        """Register an agent with its capabilities."""
        ...
    
    async def deregister_agent(self, agent_id: str) -> None:
        """Deregister an agent."""
        ...
    
    async def get_artifact(self, artifact_id: str) -> dict[str, Any] | None:
        """Get artifact by ID."""
        ...
    
    async def store_artifact(self, artifact_id: str, content: str) -> None:
        """Store artifact content."""
        ...
    
    async def get_status(self) -> dict[str, Any]:
        """Get current workflow status."""
        ...
    
    async def update_status(self, phase: str, completed: bool) -> None:
        """Update workflow phase status."""
        ...


class InMemoryMessageBoard:
    """In-memory message board for local development/testing."""
    
    def __init__(self) -> None:
        self._queues: dict[MessageType, asyncio.Queue[Message]] = {}
        self._agents: dict[str, dict[str, Any]] = {}
        self._artifacts: dict[str, str] = {}
        self._status: dict[str, Any] = {
            "phase": "initializing",
            "completed_phases": [],
            "current_agent": None,
        }
        self._lock = asyncio.Lock()
    
    async def publish(self, message: Message) -> None:
        """Publish a message to all subscribers."""
        async with self._lock:
            if message.type not in self._queues:
                self._queues[message.type] = asyncio.Queue()
            
            await self._queues[message.type].put(message)
    
    async def subscribe(self, message_type: MessageType) -> asyncio.Queue[Message]:
        """Subscribe to a message type, returns a queue for consumption."""
        async with self._lock:
            if message_type not in self._queues:
                self._queues[message_type] = asyncio.Queue()
            return self._queues[message_type]
    
    async def register_agent(self, agent_id: str, capabilities: list[str]) -> None:
        """Register an agent with its capabilities."""
        async with self._lock:
            self._agents[agent_id] = {
                "capabilities": capabilities,
                "registered_at": datetime.now(),
                "status": "active",
            }
    
    async def deregister_agent(self, agent_id: str) -> None:
        """Deregister an agent."""
        async with self._lock:
            if agent_id in self._agents:
                self._agents[agent_id]["status"] = "deregistered"
    
    async def get_artifact(self, artifact_id: str) -> dict[str, Any] | None:
        """Get artifact by ID."""
        async with self._lock:
            if artifact_id in self._artifacts:
                return {
                    "id": artifact_id,
                    "content": self._artifacts[artifact_id],
                    "updated_at": datetime.now(),
                }
            return None
    
    async def store_artifact(self, artifact_id: str, content: str) -> None:
        """Store artifact content."""
        async with self._lock:
            self._artifacts[artifact_id] = content
    
    async def get_status(self) -> dict[str, Any]:
        """Get current workflow status."""
        async with self._lock:
            return {
                **self._status,
                "agents": self._agents,
                "artifacts": list(self._artifacts.keys()),
            }
    
    async def update_status(self, phase: str, completed: bool) -> None:
        """Update workflow phase status."""
        async with self._lock:
            self._status["phase"] = phase
            if completed:
                if phase not in self._status["completed_phases"]:
                    self._status["completed_phases"].append(phase)
            self._status["updated_at"] = datetime.now()
```

#### Key Design Decisions

1. **In-Memory First**: Starts with in-memory implementation for simplicity, extensible to Redis/PostgreSQL for distributed setups
2. **Type-Safe Messages**: Enum for message types prevents invalid message creation
3. **Correlation IDs**: Enable request/response chains and traceability
4. **Priority Queues**: Higher priority messages processed first (e.g., errors, validation results)

### Component 2: Refinement Coordinator

#### Purpose
Main orchestrator that parses user requests, delegates to specialized agents, manages feedback loops, and coordinates with OpenSpec CLI.

#### Interface

```python
from __future__ import annotations

from dataclasses import dataclass
from datetime import datetime
from enum import Enum
from typing import TYPE_CHECKING

if TYPE_CHECKING:
    from app.config.settings import Settings
    from message_board import MessageBoardProtocol


class LanguageStyle(str, Enum):
    PYTHON = "python"
    TYPESCRIPT = "typescript"
    JAVA = "java"
    GO = "go"
    RUST = "rust"
    STRUCTURED_PSEUDO_CODE = "structured_pseudocode"
    MATHEMATICAL_PSEUDO_CODE = "mathematical_pseudocode"
    LEAN = "lean"


@dataclass(frozen=True)
class LanguageDetectionResult:
    detected: bool
    language: LanguageStyle | None = None
    reason: str = ""
    confidence: float = 0.0


class RefinementCoordinator:
    """Orchestrates SDD workflow with specialized agents."""
    
    def __init__(
        self,
        message_board: MessageBoardProtocol,
        settings: Settings,
    ) -> None:
        self._mb = message_board
        self._settings = settings
        self._task_id_counter = 0
        self._tasks: dict[str, dict[str, Any]] = {}
    
    async def detect_language(self, user_request: str) -> LanguageDetectionResult:
        """Detect programming language or pseudocode style from user request."""
        # Implementation details in Low-Level Design section
        ...
    
    async def orchestrate_workflow(
        self,
        feature_name: str,
        user_request: str,
    ) -> None:
        """Main orchestration entry point."""
        # Step 1: Parse request and detect language
        language_result = await self.detect_language(user_request)
        
        # Step 2: Publish workflow started
        await self._mb.publish(Message(
            id=self._generate_task_id(),
            type=MessageType.WORKFLOW_STARTED,
            timestamp=datetime.now(),
            sender_id="coordinator",
            payload={
                "feature_name": feature_name,
                "user_request": user_request,
                "detected_language": language_result.language.value if language_result.language else None,
            },
        ))
        
        # Step 3: Delegate to specialized agents
        # Phase 1: Requirements
        requirements_result = await self._orchestrate_requirements_phase(
            feature_name,
            language_result,
        )
        
        # Phase 2: Design
        design_result = await self._orchestrate_design_phase(
            feature_name,
            requirements_result,
        )
        
        # Phase 3: Tasks
        tasks_result = await self._orchestrate_tasks_phase(
            feature_name,
            design_result,
        )
        
        # Phase 4: Git Integration
        git_result = await self._orchestrate_git_phase(
            feature_name,
            tasks_result,
        )
        
        # Step 4: Publish workflow completed
        await self._mb.publish(Message(
            id=self._generate_task_id(),
            type=MessageType.WORKFLOW_COMPLETED,
            timestamp=datetime.now(),
            sender_id="coordinator",
            payload={
                "feature_name": feature_name,
                "phases_completed": ["requirements", "design", "tasks", "git"],
                "git_branch": git_result.get("branch_name"),
            },
        ))
    
    async def _orchestrate_requirements_phase(
        self,
        feature_name: str,
        language_result: LanguageDetectionResult,
    ) -> dict[str, Any]:
        """Orchestrate requirements analysis phase."""
        await self._mb.update_status("requirements", False)
        
        # Register Requirements Analyst agent
        await self._mb.register_agent(
            "requirements_analyst",
            ["ears_analysis", "requirement_extraction", "scenario_generation"],
        )
        
        # Publish task to message board
        await self._mb.publish(Message(
            id=self._generate_task_id(),
            type=MessageType.AGENT_TASK,
            timestamp=datetime.now(),
            sender_id="coordinator",
            payload={
                "agent_id": "requirements_analyst",
                "task_type": "analyze_requirements",
                "feature_name": feature_name,
                "detected_language": language_result.language.value if language_result.language else None,
            },
        ))
        
        # Wait for completion
        requirements_queue = await self._mb.subscribe(MessageType.TASK_COMPLETE)
        result_message = await requirements_queue.get()
        
        await self._mb.deregister_agent("requirements_analyst")
        await self._mb.update_status("requirements", True)
        
        return result_message.payload
    
    async def _orchestrate_design_phase(
        self,
        feature_name: str,
        requirements_result: dict[str, Any],
    ) -> dict[str, Any]:
        """Orchestrate design phase."""
        await self._mb.update_status("design", False)
        
        # Register Architect and QA Auditor agents
        await self._mb.register_agent(
            "architect",
            ["rfc_design", "architecture_decisions", "trade_off_analysis"],
        )
        await self._mb.register_agent(
            "qa_auditor",
            ["hostile_audit", "edge_case_discovery", "severity_classification"],
        )
        
        # Design creation
        await self._mb.publish(Message(
            id=self._generate_task_id(),
            type=MessageType.AGENT_TASK,
            timestamp=datetime.now(),
            sender_id="coordinator",
            payload={
                "agent_id": "architect",
                "task_type": "create_design",
                "feature_name": feature_name,
                "requirements": requirements_result,
            },
        ))
        
        # QA Audit
        await self._mb.publish(Message(
            id=self._generate_task_id(),
            type=MessageType.AGENT_TASK,
            timestamp=datetime.now(),
            sender_id="coordinator",
            payload={
                "agent_id": "qa_auditor",
                "task_type": "audit_design",
                "feature_name": feature_name,
                "design": "...",
            },
        ))
        
        # Wait for audit completion
        audit_queue = await self._mb.subscribe(MessageType.AUDIT_FINDING)
        audit_result = await audit_queue.get()
        
        await self._mb.deregister_agent("architect")
        await self._mb.deregister_agent("qa_auditor")
        await self._mb.update_status("design", True)
        
        return audit_result.payload
    
    async def _orchestrate_tasks_phase(
        self,
        feature_name: str,
        design_result: dict[str, Any],
    ) -> dict[str, Any]:
        """Orchestrate tasks generation phase."""
        await self._mb.update_status("tasks", False)
        
        # Register agent
        await self._mb.register_agent(
            "task_generator",
            ["task_breakdown", "dependency_ordering", "verification_planning"],
        )
        
        await self._mb.publish(Message(
            id=self._generate_task_id(),
            type=MessageType.AGENT_TASK,
            timestamp=datetime.now(),
            sender_id="coordinator",
            payload={
                "agent_id": "task_generator",
                "task_type": "generate_tasks",
                "feature_name": feature_name,
                "design": design_result,
            },
        ))
        
        tasks_queue = await self._mb.subscribe(MessageType.TASK_COMPLETE)
        result_message = await tasks_queue.get()
        
        await self._mb.deregister_agent("task_generator")
        await self._mb.update_status("tasks", True)
        
        return result_message.payload
    
    async def _orchestrate_git_phase(
        self,
        feature_name: str,
        tasks_result: dict[str, Any],
    ) -> dict[str, Any]:
        """Orchestrate git integration phase."""
        await self._mb.update_status("git", False)
        
        # Register agent
        await self._mb.register_agent(
            "git_integrator",
            ["branch_management", "Incident_history", "pr_generation"],
        )
        
        await self._mb.publish(Message(
            id=self._generate_task_id(),
            type=MessageType.AGENT_TASK,
            timestamp=datetime.now(),
            sender_id="coordinator",
            payload={
                "agent_id": "git_integrator",
                "task_type": "create_git_artifacts",
                "feature_name": feature_name,
                "tasks": tasks_result,
            },
        ))
        
        git_queue = await self._mb.subscribe(MessageType.TASK_COMPLETE)
        result_message = await git_queue.get()
        
        await self._mb.deregister_agent("git_integrator")
        await self._mb.update_status("git", True)
        
        return result_message.payload
    
    def _generate_task_id(self) -> str:
        """Generate unique task ID."""
        self._task_id_counter += 1
        return f"task-{self._task_id_counter:04d}"
```

#### Key Design Decisions

1. **Phase-Based Orchestration**: Clear separation of concerns with dedicated phases
2. **Agent Registration/Deregistration**: Ensures clean lifecycle management
3. **Message-Driven Coordination**: Agents communicate via message board, not direct calls
4. **Extensible**: Easy to add new agents or phases without changing core logic

### Component 3: Requirements Analyst

#### Purpose
Analyzes user requirements using EARS methodology and generates structured requirements documents.

#### Interface

```python
from __future__ import annotations

from dataclasses import dataclass
from enum import Enum
from typing import TYPE_CHECKING

if TYPE_CHECKING:
    from message_board import MessageBoardProtocol


class EarsCategory(str, Enum):
    UBIQUITOUS = "ubiquitous"
    EVENT_DRIVEN = "event_driven"
    STATE_DRIVEN = "state_driven"
    UNWANTED_BEHAVIOR = "unwanted_behavior"
    OPTIONAL_FEATURES = "optional_features"


@dataclass(frozen=True)
class Requirement:
    id: str
    category: EarsCategory
    description: str
    acceptance_criteria: list[str]
    priority: int
    dependencies: list[str]


class RequirementsAnalyst:
    """Analyzes user requirements using EARS methodology."""
    
    def __init__(self, message_board: MessageBoardProtocol) -> None:
        self._mb = message_board
    
    async def analyze_requirements(
        self,
        feature_name: str,
        user_request: str,
        detected_language: str | None,
    ) -> dict[str, Any]:
        """Main entry point for requirements analysis."""
        # Step 1: Parse user request
        parsed_request = await self._parse_user_request(user_request)
        
        # Step 2: Apply EARS methodology
        ears_analysis = await self._apply_ears_methodology(
            parsed_request,
            detected_language,
        )
        
        # Step 3: Generate requirements document
        requirements_doc = self._generate_requirements_doc(
            feature_name,
            ears_analysis,
            detected_language,
        )
        
        # Step 4: Store in message board
        await self._mb.store_artifact(f"{feature_name}/requirements.md", requirements_doc)
        
        return {
            "feature_name": feature_name,
            "requirements": requirements_doc,
            "ears_analysis": ears_analysis.model_dump(),
            "detected_language": detected_language,
        }
    
    async def _parse_user_request(self, user_request: str) -> dict[str, Any]:
        """Parse user request into structured format."""
        return {
            "raw_text": user_request,
            "features": self._extract_features(user_request),
            "constraints": self._extract_constraints(user_request),
            "language_indicators": self._extract_language_indicators(user_request),
        }
    
    async def _apply_ears_methodology(
        self,
        parsed_request: dict[str, Any],
        detected_language: str | None,
    ) -> dict[str, list[Requirement]]:
        """Apply EARS methodology to parsed request."""
        return {
            EarsCategory.UBIQUITOUS: await self._extract_ubiquitous_requirements(parsed_request, detected_language),
            EarsCategory.EVENT_DRIVEN: await self._extract_event_driven_requirements(parsed_request),
            EarsCategory.STATE_DRIVEN: await self._extract_state_driven_requirements(parsed_request),
            EarsCategory.UNWANTED_BEHAVIOR: await self._extract_unwanted_behavior_requirements(parsed_request),
            EarsCategory.OPTIONAL_FEATURES: await self._extract_optional_features_requirements(parsed_request),
        }
    
    def _generate_requirements_doc(
        self,
        feature_name: str,
        ears_analysis: dict[str, list[Requirement]],
        detected_language: str | None,
    ) -> str:
        """Generate requirements document in EARS format."""
        # Implementation details in Low-Level Design section
        ...
    
    def _extract_features(self, user_request: str) -> list[str]:
        """Extract feature descriptions from user request."""
        # Implementation details in Low-Level Design section
        ...
    
    def _extract_constraints(self, user_request: str) -> list[str]:
        """Extract constraints from user request."""
        # Implementation details in Low-Level Design section
        ...
    
    def _extract_language_indicators(self, user_request: str) -> str | None:
        """Extract programming language or pseudocode style indicators."""
        # Implementation details in Low-Level Design section
        ...
    
    async def _extract_ubiquitous_requirements(
        self,
        parsed_request: dict[str, Any],
        detected_language: str | None,
    ) -> list[Requirement]:
        """Extract ubiquitous requirements (core functionality)."""
        # Implementation details in Low-Level Design section
        ...
    
    async def _extract_event_driven_requirements(
        self,
        parsed_request: dict[str, Any],
    ) -> list[Requirement]:
        """Extract event-driven requirements (external events, triggers)."""
        # Implementation details in Low-Level Design section
        ...
    
    async def _extract_state_driven_requirements(
        self,
        parsed_request: dict[str, Any],
    ) -> list[Requirement]:
        """Extract state-driven requirements (state transitions, validations)."""
        # Implementation details in Low-Level Design section
        ...
    
    async def _extract_unwanted_behavior_requirements(
        self,
        parsed_request: dict[str, Any],
    ) -> list[Requirement]:
        """Extract unwanted behavior requirements (error handling, edge cases)."""
        # Implementation details in Low-Level Design section
        ...
    
    async def _extract_optional_features_requirements(
        self,
        parsed_request: dict[str, Any],
    ) -> list[Requirement]:
        """Extract optional features requirements (nice-to-haves)."""
        # Implementation details in Low-Level Design section
        ...
```

#### Key Design Decisions

1. **EARS Methodology**: Structured approach to requirements analysis
2. **Language Detection**: Integrates with coordinator for consistent language usage
3. **Extensible Categories**: Easy to add new EARS categories or sub-categories
4. **Requirements as Data**: Structured requirements for downstream processing

### Component 4: Architect

#### Purpose
Creates RFC-style design documents with architecture decisions, trade-offs, and implementation details.

#### Interface

```python
from __future__ import annotations

from dataclasses import dataclass
from typing import TYPE_CHECKING

if TYPE_CHECKING:
    from message_board import MessageBoardProtocol


@dataclass(frozen=True)
class ArchitectureDecision:
    decision_id: str
    context: str
    decision: str
    rationale: str
    alternatives_considered: list[str]
    consequences: dict[str, str]


class Architect:
    """Creates RFC-style design documents."""
    
    def __init__(self, message_board: MessageBoardProtocol) -> None:
        self._mb = message_board
    
    async def create_design(
        self,
        feature_name: str,
        requirements: dict[str, Any],
    ) -> dict[str, Any]:
        """Main entry point for design creation."""
        # Step 1: Parse requirements
        parsed_requirements = await self._parse_requirements(requirements)
        
        # Step 2: Create architecture diagrams
        diagrams = await self._create_diagrams(feature_name)
        
        # Step 3: Create components and interfaces
        components = await self._create_components(feature_name)
        
        # Step 4: Document data models
        data_models = await self._create_data_models(feature_name)
        
        # Step 5: Document architecture decisions
        decisions = await self._create_architecture_decisions(feature_name)
        
        # Step 6: Document trade-offs
        trade_offs = await self._create_trade_offs(feature_name)
        
        # Step 7: Generate RFC-style design document
        design_doc = self._generate_design_doc(
            feature_name,
            parsed_requirements,
            diagrams,
            components,
            data_models,
            decisions,
            trade_offs,
        )
        
        # Step 8: Store in message board
        await self._mb.store_artifact(f"{feature_name}/design.md", design_doc)
        
        return {
            "feature_name": feature_name,
            "design": design_doc,
            "diagrams": diagrams,
            "components": components,
            "data_models": data_models,
            "decisions": decisions,
            "trade_offs": trade_offs,
        }
    
    async def _parse_requirements(self, requirements: dict[str, Any]) -> dict[str, Any]:
        """Parse requirements for design creation."""
        # Implementation details in Low-Level Design section
        ...
    
    async def _create_diagrams(self, feature_name: str) -> dict[str, str]:
        """Create architecture and sequence diagrams."""
        return {
            "architecture": await self._create_architecture_diagram(feature_name),
            "data_flow": await self._create_data_flow_diagram(feature_name),
            "sequence": await self._create_sequence_diagram(feature_name),
        }
    
    async def _create_components(self, feature_name: str) -> dict[str, Any]:
        """Create component definitions and interfaces."""
        # Implementation details in Low-Level Design section
        ...
    
    async def _create_data_models(self, feature_name: str) -> dict[str, Any]:
        """Create data model definitions."""
        # Implementation details in Low-Level Design section
        ...
    
    async def _create_architecture_decisions(
        self,
        feature_name: str,
    ) -> list[ArchitectureDecision]:
        """Create architecture decision records."""
        # Implementation details in Low-Level Design section
        ...
    
    async def _create_trade_offs(self, feature_name: str) -> list[dict[str, str]]:
        """Document trade-offs and risks."""
        # Implementation details in Low-Level Design section
        ...
    
    def _generate_design_doc(
        self,
        feature_name: str,
        parsed_requirements: dict[str, Any],
        diagrams: dict[str, str],
        components: dict[str, Any],
        data_models: dict[str, Any],
        decisions: list[ArchitectureDecision],
        trade_offs: list[dict[str, str]],
    ) -> str:
        """Generate RFC-style design document."""
        # Implementation details in Low-Level Design section
        ...
```

#### Key Design Decisions

1. **RFC-Style Format**: Standardized format with clear sections
2. **Diagram Generation**: Automatic diagram creation for common patterns
3. **Architecture Decisions**: Explicit tracking of key decisions with rationale
4. **Trade-offs Documentation**: Explicit trade-offs and risks documented

### Component 5: QA Auditor

#### Purpose
Performs hostile audit / adversarial review with severity-based findings.

#### Interface

```python
from __future__ import annotations

from dataclasses import dataclass
from enum import Enum
from typing import TYPE_CHECKING

if TYPE_CHECKING:
    from message_board import MessageBoardProtocol


class Severity(str, Enum):
    CRITICAL = "CRITICAL"
    WARNING = "WARNING"
    SUGGESTION = "SUGGESTION"


@dataclass(frozen=True)
class AuditFinding:
    finding_id: str
    severity: Severity
    category: str
    description: str
    location: str | None
    recommendation: str


class QAAuditor:
    """Performs hostile audit / adversarial review."""
    
    def __init__(self, message_board: MessageBoardProtocol) -> None:
        self._mb = message_board
    
    async def audit_design(
        self,
        feature_name: str,
        design: dict[str, Any],
    ) -> dict[str, Any]:
        """Main entry point for QA audit."""
        # Step 1: Load design and requirements
        requirements = await self._mb.get_artifact(f"{feature_name}/requirements.md")
        design_content = design.get("design", "")
        
        # Step 2: Run automated checks
        automated_checks = await self._run_automated_checks(design_content, requirements)
        
        # Step 3: Run edge case analysis
        edge_cases = await self._analyze_edge_cases(design_content, requirements)
        
        # Step 4: Run adversarial review
        adversarial = await self._run_adversarial_review(design_content, requirements)
        
        # Step 5: Combine findings
        all_findings = self._combine_findings(automated_checks, edge_cases, adversarial)
        
        # Step 6: Generate audit report
        audit_report = self._generate_audit_report(feature_name, all_findings)
        
        # Step 7: Store in message board
        await self._mb.store_artifact(f"{feature_name}/review.md", audit_report)
        
        return {
            "feature_name": feature_name,
            "audit_report": audit_report,
            "findings": [f.model_dump() for f in all_findings],
        }
    
    async def _run_automated_checks(
        self,
        design_content: str,
        requirements: dict[str, Any] | None,
    ) -> list[AuditFinding]:
        """Run automated checks on design."""
        return await self._check_spec_format(design_content) + \
               await self._check_consistency(design_content, requirements) + \
               await self._check_completeness(design_content, requirements)
    
    async def _check_spec_format(self, design_content: str) -> list[AuditFinding]:
        """Check spec format compliance."""
        # Implementation details in Low-Level Design section
        ...
    
    async def _check_consistency(
        self,
        design_content: str,
        requirements: dict[str, Any] | None,
    ) -> list[AuditFinding]:
        """Check consistency between design and requirements."""
        # Implementation details in Low-Level Design section
        ...
    
    async def _check_completeness(
        self,
        design_content: str,
        requirements: dict[str, Any] | None,
    ) -> list[AuditFinding]:
        """Check completeness of design coverage."""
        # Implementation details in Low-Level Design section
        ...
    
    async def _analyze_edge_cases(
        self,
        design_content: str,
        requirements: dict[str, Any] | None,
    ) -> list[AuditFinding]:
        """Analyze for edge cases and error scenarios."""
        # Implementation details in Low-Level Design section
        ...
    
    async def _run_adversarial_review(
        self,
        design_content: str,
        requirements: dict[str, Any] | None,
    ) -> list[AuditFinding]:
        """Run adversarial review - try to break the design."""
        # Implementation details in Low-Level Design section
        ...
    
    def _combine_findings(
        self,
        automated_checks: list[AuditFinding],
        edge_cases: list[AuditFinding],
        adversarial: list[AuditFinding],
    ) -> list[AuditFinding]:
        """Combine findings and deduplicate."""
        all_findings = automated_checks + edge_cases + adversarial
        return list({f.finding_id: f for f in all_findings}.values())
    
    def _generate_audit_report(
        self,
        feature_name: str,
        findings: list[AuditFinding],
    ) -> str:
        """Generate audit report."""
        # Implementation details in Low-Level Design section
        ...
```

#### Key Design Decisions

1. **Severity-Based Classification**: CRITICAL/WARNING/SUGGESTION for triaging
2. **Multiple Review Approaches**: Automated checks, edge cases, adversarial
3. **Deduplication**: Combines findings from multiple sources
4. **Actionable Recommendations**: Each finding includes recommendation

### Component 6: Validator

#### Purpose
Validates spec format and rule compliance.

#### Interface

```python
from __future__ import annotations

from dataclasses import dataclass
from typing import TYPE_CHECKING

if TYPE_CHECKING:
    from message_board import MessageBoardProtocol


@dataclass(frozen=True)
class ValidationResult:
    is_valid: bool
    errors: list[str]
    warnings: list[str]
    suggestions: list[str]


class Validator:
    """Validates spec format and rule compliance."""
    
    def __init__(self, message_board: MessageBoardProtocol) -> None:
        self._mb = message_board
    
    async def validate_requirements(
        self,
        feature_name: str,
        requirements: str,
    ) -> ValidationResult:
        """Validate requirements document."""
        # Implementation details in Low-Level Design section
        ...
    
    async def validate_design(
        self,
        feature_name: str,
        design: str,
        requirements: str,
    ) -> ValidationResult:
        """Validate design document against requirements."""
        # Implementation details in Low-Level Design section
        ...
    
    async def validate_tasks(
        self,
        feature_name: str,
        tasks: str,
        design: str,
    ) -> ValidationResult:
        """Validate tasks document against design."""
        # Implementation details in Low-Level Design section
        ...
```

#### Key Design Decisions

1. **Format Validation**: Ensures artifacts match expected formats
2. **Consistency Checking**: Validates artifacts against each other
3. **Rule Compliance**: Checks against project rules (RESULT-PATTERN, EXCEPTION-RULES, etc.)

### Component 7: Git Integrator

#### Purpose
Manages git operations: branch creation, commits, PR generation.

#### Interface

```python
from __future__ import annotations

from dataclasses import dataclass
from typing import TYPE_CHECKING

if TYPE_CHECKING:
    from message_board import MessageBoardProtocol


@dataclass(frozen=True)
class GitOperationResult:
    success: bool
    branch_name: str | None
    commit_hash: str | None
    pr_url: str | None
    error: str | None = None


class GitIntegrator:
    """Manages git operations for SDD artifacts."""
    
    def __init__(self, message_board: MessageBoardProtocol) -> None:
        self._mb = message_board
    
    async def create_git_artifacts(
        self,
        feature_name: str,
        tasks: dict[str, Any],
    ) -> GitOperationResult:
        """Main entry point for git integration."""
        # Step 1: Create git branch
        branch_name = await self._create_branch(feature_name)
        
        if not branch_name:
            return GitOperationResult(
                success=False,
                branch_name=None,
                commit_hash=None,
                pr_url=None,
                error="Failed to create git branch",
            )
        
        # Step 2: Add and commit artifacts
        commit_hash = await self._commit_artifacts(feature_name, branch_name)
        
        if not commit_hash:
            return GitOperationResult(
                success=False,
                branch_name=branch_name,
                commit_hash=None,
                pr_url=None,
                error="Failed to commit artifacts",
            )
        
        # Step 3: Generate PR
        pr_url = await self._generate_pr(feature_name, branch_name)
        
        return GitOperationResult(
            success=True,
            branch_name=branch_name,
            commit_hash=commit_hash,
            pr_url=pr_url,
        )
    
    async def _create_branch(self, feature_name: str) -> str | None:
        """Create git branch for feature."""
        # Implementation details in Low-Level Design section
        ...
    
    async def _commit_artifacts(
        self,
        feature_name: str,
        branch_name: str,
    ) -> str | None:
        """Commit SDD artifacts to git."""
        # Implementation details in Low-Level Design section
        ...
    
    async def _generate_pr(
        self,
        feature_name: str,
        branch_name: str,
    ) -> str | None:
        """Generate pull request."""
        # Implementation details in Low-Level Design section
        ...
```

#### Key Design Decisions

1. **Branch-per-Feature**: Isolated branch for each feature
2. **Commit-per-Artifact**: Individual commits for traceability
3. **PR Generation**: Automatic PR creation for review

### Component 8: Customizer

#### Purpose
Parses configuration, applies templates, and enables methodology customization.

#### Interface

```python
from __future__ import annotations

from dataclasses import dataclass
from typing import TYPE_CHECKING

if TYPE_CHECKING:
    from message_board import MessageBoardProtocol


@dataclass(frozen=True)
class CustomizerConfig:
    schema: str
    rules: dict[str, Any]
    templates: dict[str, str]
    context: str


class Customizer:
    """Parses config and applies templates for methodology customization."""
    
    def __init__(self, message_board: MessageBoardProtocol) -> None:
        self._mb = message_board
    
    async def load_config(self) -> CustomizerConfig:
        """Load configuration from openspec/config.yaml."""
        # Implementation details in Low-Level Design section
        ...
    
    async def apply_template(
        self,
        template_name: str,
        context: dict[str, Any],
    ) -> str:
        """Apply template with context."""
        # Implementation details in Low-Level Design section
        ...
    
    async def customize_methodology(
        self,
        base_methodology: str,
        customizations: dict[str, Any],
    ) -> str:
        """Apply customizations to methodology."""
        # Implementation details in Low-Level Design section
        ...
```

#### Key Design Decisions

1. **Config as Source of Truth**: All methodology rules from config
2. **Template System**: Flexible template application
3. **Methodology Customization**: Extensible methodology system

## Data Models

### Message Format

```python
@dataclass(frozen=True)
class Message:
    """Message format for inter-agent communication."""
    
    id: str
    type: MessageType
    timestamp: datetime
    sender_id: str
    payload: dict[str, Any]
    correlation_id: str | None = None  # For request/response chains
    priority: int = 0  # Higher priority processed first
```

### Audit Finding

```python
@dataclass(frozen=True)
class AuditFinding:
    finding_id: str
    severity: Severity
    category: str
    description: str
    location: str | None
    recommendation: str
```

### Architecture Decision

```python
@dataclass(frozen=True)
class ArchitectureDecision:
    decision_id: str
    context: str
    decision: str
    rationale: str
    alternatives_considered: list[str]
    consequences: dict[str, str]
```

### Requirement

```python
@dataclass(frozen=True)
class Requirement:
    id: str
    category: EarsCategory
    description: str
    acceptance_criteria: list[str]
    priority: int
    dependencies: list[str]
```

## Correctness Properties

### Property 1: Message Delivery Reliability

*For any* valid message published to the message board, the message SHALL be delivered to all subscribers within the configured timeout.

**Validates: Requirement 1**

### Property 2: Agent Registration Uniqueness

*For any* agent ID, the message board SHALL allow only one active registration at a time.

**Validates: Requirement 2**

### Property 3: Artifact Consistency

*For any* artifact stored in the message board, all subscribers SHALL see the same content when reading the artifact after storage.

**Validates: Requirement 16**

### Property 4: Agent Lifecycle Management

*For any* agent registration, the message board SHALL ensure exactly one active registration exists at a time.

**Validates: Requirement 2**

### Property 5: EARS Format Validation

*For any* requirements document, the validator SHALL check at least the following:
- EARS format compliance for all 5 categories (ubiquitous, event-driven, state-driven, unwanted behavior, optional features)
- Scenario format (#### WHEN/THEN)
- Requirement numbering continuity
- No duplicate requirement IDs

**Validates: Requirement 10**

### Property 6: RFC Structure Validation

*For any* design document, all 7 required sections SHALL be present:
- Overview & Context
- Requirements & Constraints
- Proposed Architecture
- Detailed Technical Design
- Trade-offs & Alternatives Considered
- Observability, Rollout & Migration
- Open Questions

**Validates: Requirement 14**

### Property 7: Agent Coordination

*For any* agent interaction, communication SHALL occur only via message board messages.

**Validates: Requirement 15**

### Property 8: Artifact Storage

*For any* artifact storage operation, the content SHALL be stored by ID and retrievable by all agents.

**Validates: Requirement 16**

### Property 9: Phase Progression

*For any* SDD workflow, phases SHALL complete in the defined order:
1. Requirements → 2. Design → 3. Tasks → 4. Git Integration

**Validates: Requirement 17**

### Property 10: Audit Finding Generation

*For any* design document, the QA Auditor SHALL generate audit findings with severity classification:
- CRITICAL: Blocker that must be fixed before implementation
- WARNING: Important issue that should be addressed
- SUGGESTION: Improvement suggestion for consideration

**Validates: Requirement 18**

## Error Handling

### Exception Hierarchy

```python
class SDDException(Exception):
    """Base exception for SDD system."""
    pass


class MessageBoardException(SDDException):
    """Raised for message board operations."""
    pass


class AgentException(SDDException):
    """Raised for agent operations."""
    pass


class ValidationException(SDDException):
    """Raised for validation failures."""
    pass


class AuditException(SDDException):
    """Raised for audit failures."""
    pass


class GitException(SDDException):
    """Raised for git operations."""
    pass
```

### Error Context

All exceptions include:
- Operation that failed
- Contextual information (feature_name, agent_id, etc.)
- Suggested recovery action

## Testing Strategy

### Unit Testing Approach

Test each component in isolation:
- Message board message delivery and persistence
- Agent task execution and message publication
- Requirements analysis with various input formats
- Design document generation with template substitution
- Audit finding classification and severity assignment
- Git operations with mocked repository

### Property-Based Testing Approach

**Property Test Library**: `hypothesis` (Python)

**Property Tests**:
- Property 1: Message delivery reliability
- Property 2: Agent registration uniqueness
- Property 3: Artifact consistency
- Property 4: Validation coverage
- Property 5: Audit severity classification
- Property 6: Git branch isolation
- Property 7: PR description completeness
- Property 8: Language style consistency
- Property 9: Feedback loop completion
- Property 10: Workflow progression

### Integration Testing Approach

End-to-end workflow tests:
- Complete SDD workflow from user request to PR
- Multi-phase feedback loops
- Error recovery and retry
- Concurrent agent operations

## Performance Considerations

### Current Constraints

- In-memory message board suitable for local development
- No concurrent agent execution (sequential orchestration)

### Scalability Considerations

- Redis-backed message board for distributed setups
- Agent pool for parallel execution
- Artifact caching for large documents
- Async I/O for git operations

## Security Considerations

### Threat Model

- **Message Injection**: Message board validates message format and sender
- **Artifact Tampering**: Git commits with cryptographic signing
- **Agent Impersonation**: Agent registration with unique IDs
- **Git Repository Access**: Standard git authentication

### Mitigation Strategies

- Message format validation
- Git commit signing for critical artifacts
- Agent ID verification
- Repository access control via git authentication

## Dependencies

### External Dependencies

- **OpenSpec CLI**: Existing entrypoint skills (openspec-new-change, openspec-propose, openspec-apply-change, openspec-archive-change)
- **Git**: Version control operations
- **Asyncio**: Python async runtime

### Internal Dependencies

- **Settings**: Configuration and feature flags
- **Logging**: Structured logging for debugging
- **Exceptions**: Custom exception hierarchy

## Open Questions

1. **Message Board Storage**: Should the message board support persistence beyond in-memory? For distributed teams, Redis or PostgreSQL could be used.

2. **Agent Pool Size**: Should multiple instances of each agent type be supported for parallel execution? Currently assumes one instance per agent type.

3. **Feedback Loop Threshold**: What constitutes enough feedback to proceed? Currently requires all CRITICAL/WARNING findings to be addressed; could be configurable.

4. **Language Detection Accuracy**: The language detection algorithm could be enhanced with ML-based classification; current implementation uses regex patterns.

5. **Audit Finding Severity Thresholds**: Should severity thresholds be configurable per project? Currently uses CRITICAL/WARNING/SUGGESTION categories.