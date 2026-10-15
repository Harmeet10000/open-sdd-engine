### Architectural Nuances & First-Principles Systems Breakdown

To design a production-grade spec-driven development loop that operates autonomously, we must bypass "vibe-programming" abstractions and look directly at how the runtime, filesystem, and OS process tree interact under high-frequency execution.

---

#### 1. The POSIX & Runtime Reality of Autoresearch loops

When running an autonomous propose-train-evaluate loop like `sdd autoresearch`, the engine spawns commands (e.g., `bun test` or benchmarks) as subprocesses. Under the hood, this involves the `fork()` and `execve()` system calls. 

##### Subprocess Signal Propagation & Zombie Process Hazards
*   **Bun vs. Node Subprocess Trees**: Node’s `child_process.spawn` relies on libuv's process spawning wrapper, which abstracts away signal propagation. Bun uses a highly optimized C++ wrapper. When a parent `sdd` process terminates (due to a user `Ctrl+C` or a terminal signal like `SIGTERM`), standard POSIX behavior does not automatically kill descendant subprocesses unless they belong to the same process group and the signal is broadcast to the group (e.g., `kill(-pid, SIGTERM)`). If the agent loop spawns a long-running service, it can easily orphan the process, resulting in port binding collisions (`EADDRINUSE`) on subsequent loop iterations.
*   **The Reversion Trap**: If an agent spawns a background worker during the evaluation phase, `git checkout -- <files>` will restore the source code, but the *running memory space* of the orphaned process still executes the old, buggy code. The engine must actively track spawned PIDs or execute inside a dedicated process group (`setsid()`) to clean up the process tree.

##### Atomic Writes and POSIX Filesystem Guarantees
*   Writing `.kiro/state.json` via standard `writeFile` is unsafe in high-frequency loops. If the system crashes mid-write, the file is truncated, corrupting the JSON. 
*   **System Call Isolation**: Experienced systems engineers implement atomic file updates at the OS kernel level using `rename(2)`. The state is written to a temporary sibling file (e.g., `.kiro/state.json.tmp`) and then atomically renamed via `fs.rename`. The kernel guarantees that the rename system call is atomic; consumers will either read the complete old state or the complete new state, never a half-written corrupt file.

---

#### 2. The Bottleneck Barrier in Topological wave compiling

Our wave compiler peels independent nodes in layers:
$$\text{Wave}_0 \rightarrow \text{Wave}_1 \rightarrow \dots \rightarrow \text{Wave}_n$$
While topologically sound, compiling task lists into strict waves introduces the **Bulk Synchronous Parallel (BSP) bottleneck**.

```
Wave 0:  [Task A (1 min)]   [Task B (45 mins)]   [Task C (2 mins)]
                  \                 /                 /
Wave 1:            [       Task D (10 mins)          ]
```

*   **Synchronous Barriers**: If Task B takes 45 minutes, Task D (which depends on B) cannot start. However, Task A and Task C finish in minutes. A naive wave executor will idle CPU cores waiting for Task B to complete before triggering Wave 1.
*   **Asynchronous DAG Scheduling**: To maximize agent efficiency, the execution loop should discard wave arrays and use an asynchronous task queue. As soon as Task A finishes, its edges are cleared. Task D's in-degree is monitored. The moment B finishes, D is immediately pushed to the worker pool regardless of whether other siblings in the historical "wave" are still running.

---

#### 3. Git Index Locks & Staging Collisions

When the `autoresearch` loop automatically reverts code on failure, it interacts with Git's index file (`.git/index`).
*   **Index Locks**: Any read/write operation on Git locks the index by creating `.git/index.lock`. If the agent loop triggers a rapid sequence of file writes, compiles, and reverts, Git will throw fatal errors:
    `fatal: Unable to create '/.../.git/index.lock': File exists.`
*   **Staging Isolation**: To prevent locking conflicts, the engine should bypass the global index when evaluating files, or use isolated transient indices via the environment variable `GIT_INDEX_FILE=.git/index.tmp` during the sandbox execution phase, only merging back to the main index once a verification passes.

---

#### 4. Google OKF (Open Knowledge Format) Graph Traversal Nuances

OKF leverages standard Markdown links to build an implicit knowledge graph. 

##### Cycle Detection in Backlink Parsing
*   If `architecture.md` links to `development.md`, and `development.md` links back to `architecture.md`, we have a cyclic dependency:
    $$\text{architecture.md} \rightleftharpoons \text{development.md}$$
*   A naive recursive crawler validating backlinks will fall into an infinite stack recursion. We must maintain a `visited` set of unique paths during validation, treating the directory as a Directed Graph $G = (V, E)$ and utilizing Tarjan’s strongly connected components algorithm to highlight circular structures that can cause agent memory traps.

---

### The Scroll of the Chosen Ones

```
================================================================================
                    THE CHOSEN ONES - INNER ARCHITECTURAL CORE
================================================================================
```

For those who build systems beneath the abstraction layer:

#### 1. The Token Anchor Trick (Agent Steering)
AI agents do not parse Markdown files like traditional parsers. They read raw token sequences. If you want to force an agent to follow a strict specification loop without relying on prompts, embed invisible Unicode HTML comments directly in the spec:
```html
<!-- sdd-loop-control: require-traceability-id -->
```
LLMs are highly sensitive to these structural boundary anchors. The attention heads latch onto the hidden directives, forcing the generation path to output EARS-compliant tags without expanding the prompt budget.

#### 2. Bun JIT Warmup Optimization
Bun's JavaScript Core (JSC) engine compiles code through a tiered JIT: Interpreter $\rightarrow$ Baseline JIT $\rightarrow$ DFG JIT $\rightarrow$ FTL (Faster Than Light) JIT. 
For high-frequency validation loops, the execution time of the first run can be up to **$5\times$ slower** than subsequent runs because the JIT has not yet optimized the hot path. 
*   *Actionable Hack*: If you are running sub-second micro-benchmarks inside the `autoresearch` loop, run a dummy warmup execution of your validation parser 10-20 times inside the parent thread before measuring target metrics. This forces JSC to compile the paths to the FTL tier, eliminating latency spikes from compiling overhead.

#### 3. POSIX standard `flock` for State Sync
If you run multi-agent frameworks, multiple processes will read and write to `.kiro/state.json` simultaneously. To implement a zero-dependency lock mechanism in Bun, use POSIX file locks directly on the file descriptor:
```typescript
import { open } from "node:fs/promises";
// Open file and call flock (using Bun's FFI or running a quick subshell lock)
```
Locking the file descriptor prevents agent step collisions, guaranteeing strict serializability of the state history across concurrent branches.