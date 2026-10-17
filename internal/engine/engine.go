package engine

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"path/filepath"
	"runtime"
	"strings"

	"gopkg.in/yaml.v3"
)

// errPolicySource marks a failure to observe required policy. It maps to a
// policy/permission denial, never to a workspace or usage error.
var errPolicySource = errors.New("policy source unavailable")

// errAssemble marks a failure to build a valid protocol value from
// engine-held inputs (plan operations, work references, check objects).
// Constructor validation still fires inside those builders, so the assembly
// marker is matched before ErrInvalidProtocolValue: the fault is the
// engine's, not the caller's input.
var errAssemble = errors.New("engine assembly fault")

// errIdentitySource marks failure of the entropy/identity capability. It is
// operationally distinct from workspace observation: repairing repository
// paths cannot make a failed identity provider safe to retry.
var errIdentitySource = errors.New("identity source unavailable")

// These embedded files are the exact, reviewable definition bytes used by
// the foundation preview. Their digests are decision inputs; changing either
// file necessarily changes every newly computed plan digest.
//
//go:embed definitions/foundation-preview-workflow.yaml
var foundationPreviewWorkflowDefinition []byte

//go:embed definitions/foundation-preview-validators.yaml
var foundationPreviewValidatorDefinition []byte

// BuildMetadata identifies one reproducible build. Empty Commit and
// BuildTime render as "unknown"; Version is required by NewEngine because a
// binary that cannot name itself is a broken build, not a runtime condition.
type BuildMetadata struct {
	Version     string
	Commit      string
	BuildTime   string
	GoVersion   string
	BuildInputs map[string]string
}

// Dependencies wires every port the engine consumes. Clock is deliberately
// absent: no foundation command stamps time, and time-stamping arrives with
// completion receipts and replay (DEC-03). The Clock port stays available in
// testkit for the slices that need it.
type Dependencies struct {
	Facts   FileFacts
	IDs     IDSource
	Policy  PolicySnapshotter
	Workdir string
	Build   BuildMetadata
}

// Engine is the single supported application seam. Construct it once at the
// composition root; callers cross Execute (and UnknownCommand for
// unparseable input) and nothing else.
type Engine struct {
	facts   FileFacts
	ids     IDSource
	policy  PolicySnapshotter
	workdir string
	build   BuildMetadata
}

// NewEngine validates mandatory dependencies. It never reads ambient state:
// Workdir is an explicit composition input, not a getwd call.
func NewEngine(deps Dependencies) (*Engine, error) {
	if deps.Facts == nil || deps.IDs == nil || deps.Policy == nil {
		return nil, fmt.Errorf("%w: file facts, identity source, and policy snapshotter are required", ErrInvalidProtocolValue)
	}
	if strings.TrimSpace(deps.Workdir) == "" {
		return nil, fmt.Errorf("%w: explicit workdir is required", ErrInvalidProtocolValue)
	}
	if strings.TrimSpace(deps.Build.Version) == "" {
		return nil, fmt.Errorf("%w: build version is required", ErrInvalidProtocolValue)
	}
	buildInputs := make(map[string]string, len(deps.Build.BuildInputs))
	for key, value := range deps.Build.BuildInputs {
		buildInputs[key] = value
	}
	build := deps.Build
	build.BuildInputs = buildInputs
	return &Engine{
		facts:   deps.Facts,
		ids:     deps.IDs,
		policy:  deps.Policy,
		workdir: deps.Workdir,
		build:   build,
	}, nil
}

// UsageError reports a CLI-layer input failure (bad flags, missing
// arguments, conflicting options) as a stable outcome. A zero command
// normalizes to an unknown-command envelope; an identified command keeps
// its attribution with a null result.
func (e *Engine) UsageError(cmd Command, message string) Outcome {
	return e.fail(cmd, fmt.Errorf("%w: %s", ErrInvalidProtocolValue, message))
}

// DevelopmentBuild is the fixed dev-build identity used when no linker
// metadata was injected. Values are reproducible literals, never wall-clock
// time, so local builds stay byte-comparable.
func DevelopmentBuild() BuildMetadata {
	return BuildMetadata{Version: "0.1.0-dev", Commit: "unknown", BuildTime: "unknown", GoVersion: runtime.Version()}
}

// UnknownCommand reports unparseable CLI input without trusting it as a
// command. The input schema's command discriminator is closed, so unknown
// input can never arrive through Execute; it arrives here with the raw
// requested text preserved for the outcome envelope.
func (e *Engine) UnknownCommand(requested string) Outcome {
	if strings.TrimSpace(requested) == "" {
		requested = "(unspecified)"
	}
	mapped := mapFailure(&unknownRequest{requested: requested})
	outcome, err := NewOutcome(OutcomeInput{
		Command:          CommandUnknown,
		RequestedCommand: requested,
		Diagnostics:      []Diagnostic{mapped.value},
		NextActions:      mapped.next,
		ExitCode:         mapped.exit,
	})
	if err != nil {
		panic(fmt.Sprintf("engine: valid unknown-command outcome rejected: %v", err))
	}
	return outcome
}

type unknownRequest struct{ requested string }

func (u *unknownRequest) Error() string { return "unknown command " + u.requested }

// Execute runs one foundation command through decode → snapshot → decide →
// plan → verify → outcome. Read-only commands stop before any effect;
// dry-run produces an inspectable plan and never applies it. Cancellation is
// honored before expensive work. Renderer write errors are the caller's to
// propagate; Execute only builds the outcome.
func (e *Engine) Execute(ctx context.Context, cmd Command) Outcome {
	if err := ctx.Err(); err != nil && validOutcomeCommand(cmd.Kind()) {
		return e.fail(cmd, err)
	}
	switch cmd.Kind() {
	case CommandVersion:
		return e.version(ctx, cmd)
	case CommandDoctor:
		return e.doctor(ctx, cmd)
	case CommandStatus:
		return e.status(ctx, cmd)
	case CommandStart:
		return e.startDryRun(ctx, cmd)
	default:
		return e.fail(cmd, fmt.Errorf("%w: unsupported command kind %q", ErrInvalidProtocolValue, cmd.Kind()))
	}
}

func (e *Engine) version(ctx context.Context, cmd Command) Outcome {
	if ctxErr := ctx.Err(); ctxErr != nil {
		return e.fail(cmd, ctxErr)
	}
	buildInputs := make(map[string]string, len(e.build.BuildInputs))
	for key, value := range e.build.BuildInputs {
		buildInputs[key] = value
	}
	commit := e.build.Commit
	if commit == "" {
		commit = "unknown"
	}
	buildTime := e.build.BuildTime
	if buildTime == "" {
		buildTime = "unknown"
	}
	result := NewVersionResult(VersionResultInput{
		Version:     e.build.Version,
		Commit:      commit,
		BuildTime:   buildTime,
		GoVersion:   e.build.GoVersion,
		BuildInputs: buildInputs,
	})
	outcome, err := NewOutcome(OutcomeInput{Command: cmd.Kind(), Result: &result, ExitCode: ExitSuccess})
	if err != nil {
		panic(fmt.Sprintf("engine: valid version outcome rejected: %v", err))
	}
	return outcome
}

func (e *Engine) doctor(ctx context.Context, cmd Command) Outcome {
	root := e.workdir
	if cmd.Root() != "" {
		root = cmd.Root()
	}
	checks := []map[string]any{}
	record := func(name string, ok bool, detail string) {
		checks = append(checks, map[string]any{"name": name, "ok": ok, "detail": detail})
	}

	record("executable", true, e.build.Version)
	if _, err := e.facts.Stat(ctx, e.workdir); err != nil {
		if ctx.Err() != nil {
			return e.fail(cmd, ctx.Err())
		}
		record("workdir", false, "current directory is not accessible")
		return e.doctorFailure(cmd, checks, fmt.Errorf("%w: workdir is not accessible: %w", ErrWorkspaceNotFound, err))
	}
	record("workdir", true, e.workdir)

	layout, err := DiscoverLayout(ctx, e.facts, root, cmd.Root())
	if err != nil {
		if ctx.Err() != nil {
			return e.fail(cmd, ctx.Err())
		}
		record("workspace", false, "workspace discovery failed")
		return e.doctorFailure(cmd, checks, err)
	}
	if ctxErr := ctx.Err(); ctxErr != nil {
		return e.fail(cmd, ctxErr)
	}
	record("workspace", true, layout.WorkspaceID())
	record("schema", true, "1.0")

	policy, err := e.policy.PolicySnapshot(ctx)
	if err != nil {
		if ctx.Err() != nil {
			return e.fail(cmd, ctx.Err())
		}
		record("policy", false, "required policy source is unavailable")
		return e.doctorFailure(cmd, checks, fmt.Errorf("%w: %w", errPolicySource, err))
	}
	if ctxErr := ctx.Err(); ctxErr != nil {
		return e.fail(cmd, ctxErr)
	}
	record("policy", true, policy.Source())

	result, resultErr := NewDoctorResult(checks)
	if resultErr != nil {
		return e.fail(cmd, resultErr)
	}
	workspace := workspaceRefOf(layout)
	outcome, err := NewOutcome(OutcomeInput{
		Command:   cmd.Kind(),
		Workspace: &workspace,
		Result:    &result,
		ExitCode:  ExitSuccess,
	})
	if err != nil {
		panic(fmt.Sprintf("engine: valid doctor outcome rejected: %v", err))
	}
	return outcome
}

func (e *Engine) doctorFailure(cmd Command, checks []map[string]any, cause error) Outcome {
	result, resultErr := NewDoctorResult(checks)
	if resultErr != nil {
		return e.fail(cmd, fmt.Errorf("%w: doctor checks: %w", errAssemble, resultErr))
	}
	mapped := mapFailure(cause)
	outcome, err := NewOutcome(OutcomeInput{
		Command:     cmd.Kind(),
		Result:      &result,
		Diagnostics: []Diagnostic{mapped.value},
		NextActions: mapped.next,
		ExitCode:    mapped.exit,
	})
	if err != nil {
		panic(fmt.Sprintf("engine: valid doctor failure outcome rejected: %v", err))
	}
	return outcome
}

func (e *Engine) status(ctx context.Context, cmd Command) Outcome {
	layout, err := e.discoverFor(ctx, cmd)
	if err != nil {
		return e.fail(cmd, err)
	}
	record, err := SelectWork(layout, cmd.WorkID())
	if err != nil {
		return e.fail(cmd, err)
	}
	items := []WorkRef{}
	if record.ID() != "" {
		work, workErr := NewWorkRef(WorkRefInput{ID: record.ID(), Slug: record.Slug(), Revision: record.Revision()})
		if workErr != nil {
			return e.fail(cmd, fmt.Errorf("%w: selected work: %w", errAssemble, workErr))
		}
		items = append(items, work)
	}
	result, resultErr := NewStatusResult(items)
	if resultErr != nil {
		return e.fail(cmd, fmt.Errorf("%w: status result: %w", errAssemble, resultErr))
	}
	workspace := workspaceRefOf(layout)
	outcome, err := NewOutcome(OutcomeInput{
		Command:   cmd.Kind(),
		Workspace: &workspace,
		Result:    &result,
		ExitCode:  ExitSuccess,
	})
	if err != nil {
		panic(fmt.Sprintf("engine: valid status outcome rejected: %v", err))
	}
	return outcome
}

func (e *Engine) startDryRun(ctx context.Context, cmd Command) Outcome {
	layout, err := e.discoverFor(ctx, cmd)
	if err != nil {
		return e.fail(cmd, err)
	}
	policy, err := e.policy.PolicySnapshot(ctx)
	if err != nil {
		if ctx.Err() != nil {
			return e.fail(cmd, ctx.Err())
		}
		return e.fail(cmd, fmt.Errorf("%w: %w", errPolicySource, err))
	}
	if ctxErr := ctx.Err(); ctxErr != nil {
		return e.fail(cmd, ctxErr)
	}
	planID, err := e.ids.NewID()
	if err != nil {
		return e.fail(cmd, errors.Join(errIdentitySource, fmt.Errorf("mint plan identity: %w", err)))
	}
	if ctxErr := ctx.Err(); ctxErr != nil {
		return e.fail(cmd, ctxErr)
	}
	workID, err := e.ids.NewID()
	if err != nil {
		return e.fail(cmd, errors.Join(errIdentitySource, fmt.Errorf("mint work identity: %w", err)))
	}
	if ctxErr := ctx.Err(); ctxErr != nil {
		return e.fail(cmd, ctxErr)
	}
	if planID == "" || workID == "" {
		return e.fail(cmd, fmt.Errorf("%w: identity source minted an empty ID", errAssemble))
	}
	// The slug is display data; safe-slug derivation arrives with journey
	// profiles. Until then the immutable work ID doubles as a valid slug so
	// the preview record decodes under the engine's own strict rules. IDs
	// containing path separators are rejected instead of sanitized: silent
	// rewriting could collide two identities into one directory.
	slug := workID
	if strings.ContainsAny(slug, `/\`) {
		return e.fail(cmd, fmt.Errorf("%w: derived slug is not a safe path component", errAssemble))
	}
	path := ".sdd/work/" + slug + "/work.yaml"
	composition := cmd.Workflow()
	if composition == "" {
		composition = "unspecified"
	}
	// Proposed bytes are marshaled, never concatenated: workflow text may
	// contain YAML-significant characters (quotes, colons, comments) that
	// string building would corrupt into an undecodable record.
	zero := uint64(0)
	contentBytes, err := yaml.Marshal(workManifest{
		SchemaVersion: 1,
		WorkID:        workID,
		Slug:          slug,
		Revision:      &zero,
		Status:        "active",
		Composition:   composition,
	})
	if err != nil {
		return e.fail(cmd, fmt.Errorf("%w: dry-run content: %w", errAssemble, err))
	}
	content := string(contentBytes)
	contentDigest := BlobDigest("sdd.raw-blob/1", contentBytes)
	operation, err := NewPlanOperation(PlanOperationInput{
		Kind:             OperationCreateFile,
		Path:             path,
		Content:          &content,
		ContentDigest:    contentDigest,
		ExpectedRevision: layout.Revision(),
	})
	if err != nil {
		return e.fail(cmd, fmt.Errorf("%w: dry-run operation: %w", errAssemble, err))
	}
	precondition, err := NewPlanPrecondition(PlanPreconditionInput{
		Kind:             PreconditionAbsent,
		Path:             path,
		ExpectedRevision: layout.Revision(),
	})
	if err != nil {
		return e.fail(cmd, fmt.Errorf("%w: dry-run precondition: %w", errAssemble, err))
	}
	workflowDefinitionDigest := BlobDigest("sdd.workflow-definition/1", foundationPreviewWorkflowDefinition)
	workflowRequestDigest := BlobDigest("sdd.workflow-request/1", []byte(cmd.Workflow()))
	workflowDigest, err := DerivationDigest(
		"sdd.effective-workflow-definition/1",
		workflowDefinitionDigest,
		workflowRequestDigest,
	)
	if err != nil {
		return e.fail(cmd, fmt.Errorf("%w: effective workflow definition: %w", errAssemble, err))
	}
	validatorDigest := BlobDigest("sdd.validator-definition/1", foundationPreviewValidatorDefinition)
	planDigest, err := ComputePlanDigest(PlanDigestInput{
		PlanID: planID, WorkspaceID: layout.WorkspaceID(), WorkID: workID,
		RevisionScope: RevisionWorkspace, ExpectedRevision: layout.Revision(),
		InputDigest:               BlobDigest("sdd.start-intent/1", []byte(cmd.Intent())),
		PolicyDigest:              policy.Digest(),
		WorkflowDefinitionDigest:  workflowDigest,
		ValidatorDefinitionDigest: validatorDigest,
		Operations:                []PlanOperation{operation},
		Preconditions:             []PlanPrecondition{precondition},
	})
	if err != nil {
		return e.fail(cmd, fmt.Errorf("%w: dry-run plan digest: %w", errAssemble, err))
	}
	plan, err := NewMutationPlan(MutationPlanInput{
		PlanID: planID, WorkspaceID: layout.WorkspaceID(), WorkID: workID,
		RevisionScope: RevisionWorkspace, ExpectedRevision: layout.Revision(),
		InputDigest:               BlobDigest("sdd.start-intent/1", []byte(cmd.Intent())),
		PolicyDigest:              policy.Digest(),
		WorkflowDefinitionDigest:  workflowDigest,
		ValidatorDefinitionDigest: validatorDigest,
		Operations:                []PlanOperation{operation},
		Preconditions:             []PlanPrecondition{precondition},
		Digest:                    planDigest,
	})
	if err != nil {
		return e.fail(cmd, fmt.Errorf("%w: dry-run plan: %w", errAssemble, err))
	}
	if verifyErr := VerifyPlanDigest(plan); verifyErr != nil {
		return e.fail(cmd, fmt.Errorf("%w: dry-run plan verification: %w", errAssemble, verifyErr))
	}
	result := NewStartDryRunResult(plan)
	source, _ := NewDiagnosticSource(DiagnosticSourceInput{Kind: "plan"})
	preview, _ := NewDiagnostic(DiagnosticInput{
		Code:     "DIAG_PLAN_NOT_AUTHORIZABLE",
		Severity: SeverityWarning,
		Message:  "Preview only; durable application is not implemented in foundation",
		Source:   source,
	})
	// The previewed work item does not exist, so suggesting `status --work`
	// would be a guaranteed failure. Doctor is a valid, read-only next command
	// that re-observes both workspace and policy readiness. A future applier
	// will replace this with revision-bound application once that capability
	// exists.
	revision := layout.Revision()
	verify, verifyErr := NewNextAction(NextActionInput{
		Command:          "doctor",
		Flags:            map[string]FlagValue{"root": StringFlag(layout.Root())},
		Reason:           "Re-check workspace and policy readiness after reviewing this non-authorizable preview",
		ExpectedRevision: &revision,
	})
	if verifyErr != nil {
		return e.fail(cmd, fmt.Errorf("%w: dry-run next action: %w", errAssemble, verifyErr))
	}
	workspace := workspaceRefOf(layout)
	outcome, err := NewOutcome(OutcomeInput{
		Command:     cmd.Kind(),
		Workspace:   &workspace,
		Result:      &result,
		Diagnostics: []Diagnostic{preview},
		NextActions: []NextAction{verify},
		ExitCode:    ExitSuccess,
	})
	if err != nil {
		panic(fmt.Sprintf("engine: valid dry-run outcome rejected: %v", err))
	}
	return outcome
}

func (e *Engine) discoverFor(ctx context.Context, cmd Command) (Layout, error) {
	if err := ctx.Err(); err != nil {
		return Layout{}, err
	}
	root := e.workdir
	explicit := ""
	if cmd.Root() != "" {
		root = cmd.Root()
		if !filepath.IsAbs(root) {
			root = filepath.Join(e.workdir, root)
		}
		explicit = root
	}
	return DiscoverLayout(ctx, e.facts, root, explicit)
}

func workspaceRefOf(layout Layout) WorkspaceRef {
	workspace, err := NewWorkspaceRef(layout.Root(), layout.FormatVersion())
	if err != nil {
		panic(fmt.Sprintf("engine: discovered layout produced an invalid workspace reference: %v", err))
	}
	return workspace
}

type mappedDiagnostic struct {
	value Diagnostic
	exit  ExitCode
	next  []NextAction
}

// fail translates any Execute failure into a stable outcome. Precedence is
// explicit: cancellation, assembly, usage, selection, workspace/schema,
// policy, observation faults, then residual operational faults. An invalid
// command kind is attributed to an unknown-command outcome so malformed
// input never panics.
func (e *Engine) fail(cmd Command, err error) Outcome {
	kind := cmd.Kind()
	input := OutcomeInput{
		Command:  kind,
		ExitCode: ExitUsage,
	}
	if !validOutcomeCommand(kind) {
		kind = CommandUnknown
		input.Command = kind
		input.RequestedCommand = "(unspecified)"
	}
	mapped := mapFailure(err)
	input.Diagnostics = []Diagnostic{mapped.value}
	input.NextActions = mapped.next
	input.ExitCode = mapped.exit
	outcome, outcomeErr := NewOutcome(input)
	if outcomeErr != nil {
		panic(fmt.Sprintf("engine: valid failure outcome rejected: %v", outcomeErr))
	}
	return outcome
}

type failureMapping struct {
	code        string
	severity    Severity
	remediation string
	sourceKind  string
	retryable   bool
	exit        ExitCode
}

// mapFailure renders one error into its stable diagnostic shape. The mapping
// is total: observation-path faults land in the workspace family and
// residual assembly faults land in the schema family, both exit 3 with a
// truthful message. A future internal-error extension may narrow the
// residual default; the messages and exit family are already stable.
func mapFailure(err error) mappedDiagnostic {
	mapping, message, metadata, next := classifyFailure(err)
	source, sourceErr := NewDiagnosticSource(DiagnosticSourceInput{Kind: mapping.sourceKind})
	if sourceErr != nil {
		panic(fmt.Sprintf("engine: diagnostic source rejected: %v", sourceErr))
	}
	diagnostic, diagnosticErr := NewDiagnostic(DiagnosticInput{
		Code:        mapping.code,
		Severity:    mapping.severity,
		Message:     message,
		Source:      source,
		Remediation: mapping.remediation,
		Metadata:    metadata,
		Retryable:   mapping.retryable,
	})
	if diagnosticErr != nil {
		panic(fmt.Sprintf("engine: failure diagnostic rejected: %v", diagnosticErr))
	}
	return mappedDiagnostic{value: diagnostic, exit: mapping.exit, next: next}
}

func classifyFailure(err error) (failureMapping, string, map[string]any, []NextAction) {
	switch {
	case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
		return failureMapping{
			code: "DIAG_INTERRUPTED", severity: SeverityError,
			remediation: "Retry the command; no effects were applied",
			sourceKind:  "command", retryable: true, exit: ExitInterrupted,
		}, "Command interrupted before completion", nil, nil
	case isUnknownRequest(err):
		return failureMapping{
			code: "DIAG_COMMAND_UNKNOWN", severity: SeverityError,
			remediation: "Run sdd --help for the supported commands",
			sourceKind:  "command", retryable: false, exit: ExitUsage,
		}, err.Error(), nil, nil
	case errors.Is(err, errAssemble):
		return failureMapping{
			code: "DIAG_WORKSPACE_VERSION_INCOMPATIBLE", severity: SeverityError,
			remediation: "Retry the command; a persistent failure means the installation is defective",
			sourceKind:  "workspace", retryable: false, exit: ExitWorkspace,
		}, err.Error(), nil, nil
	case errors.Is(err, ErrInvalidProtocolValue):
		return failureMapping{
			code: "DIAG_USAGE_ERROR", severity: SeverityError,
			remediation: "Correct the command input and retry",
			sourceKind:  "command", retryable: false, exit: ExitUsage,
		}, err.Error(), nil, nil
	case errors.Is(err, ErrSelectionAmbiguous):
		candidates := AmbiguousCandidates(err)
		items := make([]any, 0, len(candidates))
		for _, candidate := range candidates {
			items = append(items, candidate)
		}
		var next []NextAction
		if len(candidates) > 0 {
			action, actionErr := NewNextAction(NextActionInput{
				Command: "status",
				Flags:   map[string]FlagValue{"work": StringFlag(candidates[0])},
				Reason:  "Select one work item explicitly",
			})
			if actionErr == nil {
				next = []NextAction{action}
			}
		}
		return failureMapping{
			code: "DIAG_WORKSPACE_SELECTION_AMBIGUOUS", severity: SeverityError,
			remediation: "Repeat the command with --work <immutable-id>",
			sourceKind:  "workspace", retryable: false, exit: ExitWorkspace,
		}, err.Error(), map[string]any{"candidates": items}, next
	case errors.Is(err, ErrWorkNotFound):
		return failureMapping{
			code: "DIAG_WORK_ITEM_NOT_FOUND", severity: SeverityError,
			remediation: "List work items with sdd status from the workspace root",
			sourceKind:  "workspace", retryable: false, exit: ExitWorkspace,
		}, err.Error(), nil, nil
	case errors.Is(err, ErrWorkspaceIncompatible):
		return failureMapping{
			code: "DIAG_WORKSPACE_VERSION_INCOMPATIBLE", severity: SeverityError,
			remediation: "Repair the workspace manifest and work records to the documented native format",
			sourceKind:  "workspace", retryable: false, exit: ExitWorkspace,
		}, err.Error(), nil, nil
	case errors.Is(err, ErrWorkspaceNotFound), errors.Is(err, ErrInvalidRoot):
		return failureMapping{
			code: "DIAG_WORKSPACE_NOT_FOUND", severity: SeverityError,
			remediation: "Run from inside a workspace or pass a valid --root",
			sourceKind:  "workspace", retryable: false, exit: ExitWorkspace,
		}, err.Error(), nil, nil
	case errors.Is(err, errPolicySource):
		return failureMapping{
			code: "WORKFLOW_POLICY_UNAVAILABLE", severity: SeverityCritical,
			remediation: "Restore the trusted policy source and retry",
			sourceKind:  "policy", retryable: false, exit: ExitPolicy,
		}, err.Error(), nil, nil
	case errors.Is(err, errIdentitySource):
		return failureMapping{
			code: "DIAG_IDENTITY_UNAVAILABLE", severity: SeverityError,
			remediation: "Restore the system identity source and retry; no effects were applied",
			sourceKind:  "engine", retryable: true, exit: ExitExternalTool,
		}, err.Error(), nil, nil
	case errors.Is(err, ErrDependencyFault):
		return failureMapping{
			code: "DIAG_WORKSPACE_NOT_FOUND", severity: SeverityError,
			remediation: "Restore access to the workspace files and retry",
			sourceKind:  "workspace", retryable: false, exit: ExitWorkspace,
		}, err.Error(), nil, nil
	default:
		return failureMapping{
			code: "DIAG_WORKSPACE_VERSION_INCOMPATIBLE", severity: SeverityError,
			remediation: "Retry the command; a persistent failure means the installation is defective",
			sourceKind:  "workspace", retryable: false, exit: ExitWorkspace,
		}, err.Error(), nil, nil
	}
}

func isUnknownRequest(err error) bool {
	var unknown *unknownRequest
	return errors.As(err, &unknown)
}
