package engine

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

var digestPattern = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)

type WorkspaceRef struct {
	root          string
	formatVersion uint64
}

func NewWorkspaceRef(root string, formatVersion uint64) (WorkspaceRef, error) {
	if strings.TrimSpace(root) == "" || formatVersion == 0 {
		return WorkspaceRef{}, fmt.Errorf("%w: workspace root and positive format version are required", ErrInvalidProtocolValue)
	}
	return WorkspaceRef{root: root, formatVersion: formatVersion}, nil
}

func (r WorkspaceRef) Root() string          { return r.root }
func (r WorkspaceRef) FormatVersion() uint64 { return r.formatVersion }

func (r WorkspaceRef) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Root          string `json:"root"`
		FormatVersion uint64 `json:"formatVersion"`
	}{Root: r.root, FormatVersion: r.formatVersion})
}

type WorkRef struct {
	id       string
	slug     string
	workflow string
	revision uint64
}

type WorkRefInput struct {
	ID       string
	Slug     string
	Workflow string
	Revision uint64
}

func NewWorkRef(input WorkRefInput) (WorkRef, error) {
	if strings.TrimSpace(input.ID) == "" || strings.TrimSpace(input.Slug) == "" {
		return WorkRef{}, fmt.Errorf("%w: work ID and slug are required", ErrInvalidProtocolValue)
	}
	return WorkRef{
		id:       input.ID,
		slug:     input.Slug,
		workflow: input.Workflow,
		revision: input.Revision,
	}, nil
}

func (r WorkRef) ID() string       { return r.id }
func (r WorkRef) Slug() string     { return r.slug }
func (r WorkRef) Workflow() string { return r.workflow }
func (r WorkRef) Revision() uint64 { return r.revision }

func (r WorkRef) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		ID       string `json:"id"`
		Slug     string `json:"slug"`
		Workflow string `json:"workflow,omitempty"`
		Revision uint64 `json:"revision"`
	}{ID: r.id, Slug: r.slug, Workflow: r.workflow, Revision: r.revision})
}

type ResultKind string

const (
	ResultStatus      ResultKind = "status"
	ResultVersion     ResultKind = "version"
	ResultDoctor      ResultKind = "doctor"
	ResultStartDryRun ResultKind = "start-dry-run"
)

type StatusResult struct {
	workItems []WorkRef
}

func NewStatusResult(workItems []WorkRef) Result {
	value := StatusResult{workItems: append([]WorkRef{}, workItems...)}
	return Result{kind: ResultStatus, status: &value}
}

func (r StatusResult) WorkItems() []WorkRef {
	return append([]WorkRef{}, r.workItems...)
}

func (r StatusResult) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Kind      ResultKind `json:"kind"`
		WorkItems []WorkRef  `json:"workItems"`
	}{Kind: ResultStatus, WorkItems: nonNilSlice(r.workItems)})
}

type VersionResult struct {
	version               string
	commit                string
	buildTime             string
	goVersion             string
	buildInputs           map[string]string
	protocolSchemaVersion string
}

type VersionResultInput struct {
	Version     string
	Commit      string
	BuildTime   string
	GoVersion   string
	BuildInputs map[string]string
}

func NewVersionResult(input VersionResultInput) Result {
	value := VersionResult{
		version:               input.Version,
		commit:                input.Commit,
		buildTime:             input.BuildTime,
		goVersion:             input.GoVersion,
		buildInputs:           cloneMap(input.BuildInputs),
		protocolSchemaVersion: SchemaVersion,
	}
	return Result{kind: ResultVersion, version: &value}
}

func (r VersionResult) Version() string                { return r.version }
func (r VersionResult) Commit() string                 { return r.commit }
func (r VersionResult) BuildTime() string              { return r.buildTime }
func (r VersionResult) GoVersion() string              { return r.goVersion }
func (r VersionResult) BuildInputs() map[string]string { return cloneMap(r.buildInputs) }
func (r VersionResult) ProtocolSchemaVersion() string  { return r.protocolSchemaVersion }

func (r VersionResult) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Kind                  ResultKind        `json:"kind"`
		Version               string            `json:"version"`
		Commit                string            `json:"commit"`
		BuildTime             string            `json:"buildTime"`
		GoVersion             string            `json:"goVersion"`
		BuildInputs           map[string]string `json:"buildInputs"`
		ProtocolSchemaVersion string            `json:"protocolSchemaVersion"`
	}{
		Kind:                  ResultVersion,
		Version:               r.version,
		Commit:                r.commit,
		BuildTime:             r.buildTime,
		GoVersion:             r.goVersion,
		BuildInputs:           nonNilMap(r.buildInputs),
		ProtocolSchemaVersion: r.protocolSchemaVersion,
	})
}

type DoctorResult struct {
	checks []map[string]any
}

func NewDoctorResult(checks []map[string]any) (Result, error) {
	cloned := make([]map[string]any, 0, len(checks))
	for index, check := range checks {
		value, err := cloneJSONObject(check)
		if err != nil {
			return Result{}, fmt.Errorf("%w: doctor check %d: %w", ErrInvalidProtocolValue, index, err)
		}
		cloned = append(cloned, value)
	}
	value := DoctorResult{checks: cloned}
	return Result{kind: ResultDoctor, doctor: &value}, nil
}

func (r DoctorResult) Checks() []map[string]any {
	checks := make([]map[string]any, 0, len(r.checks))
	for _, check := range r.checks {
		checks = append(checks, cloneStoredJSONObject(check))
	}
	return checks
}

func (r DoctorResult) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Kind   ResultKind       `json:"kind"`
		Checks []map[string]any `json:"checks"`
	}{Kind: ResultDoctor, Checks: nonNilSlice(r.checks)})
}

type OperationKind string

const (
	OperationCreateFile OperationKind = "create-file"
	OperationUpdateFile OperationKind = "update-file"
	OperationDeleteFile OperationKind = "delete-file"
)

type PlanOperation struct {
	kind             OperationKind
	path             string
	content          *string
	contentDigest    string
	expectedDigest   string
	expectedRevision uint64
	authorizable     bool
}

type PlanOperationInput struct {
	Kind             OperationKind
	Path             string
	Content          *string
	ContentDigest    string
	ExpectedDigest   string
	ExpectedRevision uint64
	Authorizable     bool
}

func NewPlanOperation(input PlanOperationInput) (PlanOperation, error) {
	if strings.TrimSpace(input.Path) == "" {
		return PlanOperation{}, fmt.Errorf("%w: plan operation path is required", ErrInvalidProtocolValue)
	}
	switch input.Kind {
	case OperationCreateFile:
		if input.Content == nil || !validDigest(input.ContentDigest) || input.ExpectedDigest != "" {
			return PlanOperation{}, fmt.Errorf("%w: invalid create-file operation", ErrInvalidProtocolValue)
		}
	case OperationUpdateFile:
		if input.Content == nil || !validDigest(input.ContentDigest) || !validDigest(input.ExpectedDigest) {
			return PlanOperation{}, fmt.Errorf("%w: invalid update-file operation", ErrInvalidProtocolValue)
		}
	case OperationDeleteFile:
		if input.Content != nil || input.ContentDigest != "" || !validDigest(input.ExpectedDigest) {
			return PlanOperation{}, fmt.Errorf("%w: invalid delete-file operation", ErrInvalidProtocolValue)
		}
	default:
		return PlanOperation{}, fmt.Errorf("%w: unsupported operation kind %q", ErrInvalidProtocolValue, input.Kind)
	}
	return PlanOperation{
		kind:             input.Kind,
		path:             input.Path,
		content:          clonePointer(input.Content),
		contentDigest:    input.ContentDigest,
		expectedDigest:   input.ExpectedDigest,
		expectedRevision: input.ExpectedRevision,
		authorizable:     input.Authorizable,
	}, nil
}

func (o PlanOperation) Kind() OperationKind      { return o.kind }
func (o PlanOperation) Path() string             { return o.path }
func (o PlanOperation) Content() *string         { return clonePointer(o.content) }
func (o PlanOperation) ContentDigest() string    { return o.contentDigest }
func (o PlanOperation) ExpectedDigest() string   { return o.expectedDigest }
func (o PlanOperation) ExpectedRevision() uint64 { return o.expectedRevision }
func (o PlanOperation) Authorizable() bool       { return o.authorizable }

func (o PlanOperation) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Kind             OperationKind `json:"kind"`
		Path             string        `json:"path"`
		Content          *string       `json:"content,omitempty"`
		ContentDigest    string        `json:"contentDigest,omitempty"`
		ExpectedDigest   string        `json:"expectedDigest,omitempty"`
		ExpectedRevision uint64        `json:"expectedRevision"`
		Authorizable     bool          `json:"authorizable"`
	}{
		Kind:             o.kind,
		Path:             o.path,
		Content:          o.content,
		ContentDigest:    o.contentDigest,
		ExpectedDigest:   o.expectedDigest,
		ExpectedRevision: o.expectedRevision,
		Authorizable:     o.authorizable,
	})
}

type PreconditionKind string

const (
	PreconditionRevision PreconditionKind = "revision"
	PreconditionDigest   PreconditionKind = "digest"
	PreconditionAbsent   PreconditionKind = "absent"
)

type PlanPrecondition struct {
	kind             PreconditionKind
	path             string
	expectedRevision uint64
	expectedDigest   string
}

type PlanPreconditionInput struct {
	Kind             PreconditionKind
	Path             string
	ExpectedRevision uint64
	ExpectedDigest   string
}

func NewPlanPrecondition(input PlanPreconditionInput) (PlanPrecondition, error) {
	if strings.TrimSpace(input.Path) == "" {
		return PlanPrecondition{}, fmt.Errorf("%w: plan precondition path is required", ErrInvalidProtocolValue)
	}
	switch input.Kind {
	case PreconditionRevision, PreconditionAbsent:
		if input.ExpectedDigest != "" {
			return PlanPrecondition{}, fmt.Errorf("%w: %s precondition cannot contain a digest", ErrInvalidProtocolValue, input.Kind)
		}
	case PreconditionDigest:
		if !validDigest(input.ExpectedDigest) {
			return PlanPrecondition{}, fmt.Errorf("%w: digest precondition requires a SHA-256 digest", ErrInvalidProtocolValue)
		}
	default:
		return PlanPrecondition{}, fmt.Errorf("%w: unsupported precondition kind %q", ErrInvalidProtocolValue, input.Kind)
	}
	return PlanPrecondition{
		kind:             input.Kind,
		path:             input.Path,
		expectedRevision: input.ExpectedRevision,
		expectedDigest:   input.ExpectedDigest,
	}, nil
}

func (p PlanPrecondition) Kind() PreconditionKind   { return p.kind }
func (p PlanPrecondition) Path() string             { return p.path }
func (p PlanPrecondition) ExpectedRevision() uint64 { return p.expectedRevision }
func (p PlanPrecondition) ExpectedDigest() string   { return p.expectedDigest }

func (p PlanPrecondition) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Kind             PreconditionKind `json:"kind"`
		Path             string           `json:"path"`
		ExpectedRevision uint64           `json:"expectedRevision"`
		ExpectedDigest   string           `json:"expectedDigest,omitempty"`
	}{
		Kind:             p.kind,
		Path:             p.path,
		ExpectedRevision: p.expectedRevision,
		ExpectedDigest:   p.expectedDigest,
	})
}

type RevisionScope string

const (
	RevisionWorkspace RevisionScope = "workspace"
	RevisionWork      RevisionScope = "work"
)

type MutationPlan struct {
	planID                    string
	workspaceID               string
	workID                    string
	revisionScope             RevisionScope
	expectedRevision          uint64
	inputDigest               string
	policyDigest              string
	workflowDefinitionDigest  string
	validatorDefinitionDigest string
	operations                []PlanOperation
	preconditions             []PlanPrecondition
	digest                    string
	idempotencyKey            string
	authorizable              bool
}

type MutationPlanInput struct {
	PlanID                    string
	WorkspaceID               string
	WorkID                    string
	RevisionScope             RevisionScope
	ExpectedRevision          uint64
	InputDigest               string
	PolicyDigest              string
	WorkflowDefinitionDigest  string
	ValidatorDefinitionDigest string
	Operations                []PlanOperation
	Preconditions             []PlanPrecondition
	Digest                    string
	IdempotencyKey            string
	Authorizable              bool
}

func NewMutationPlan(input MutationPlanInput) (MutationPlan, error) {
	if input.PlanID == "" || input.WorkspaceID == "" || input.WorkID == "" {
		return MutationPlan{}, fmt.Errorf("%w: plan, workspace, and work IDs are required", ErrInvalidProtocolValue)
	}
	if input.RevisionScope != RevisionWorkspace && input.RevisionScope != RevisionWork {
		return MutationPlan{}, fmt.Errorf("%w: unsupported revision scope %q", ErrInvalidProtocolValue, input.RevisionScope)
	}
	for name, digest := range map[string]string{
		"input":                input.InputDigest,
		"policy":               input.PolicyDigest,
		"workflow definition":  input.WorkflowDefinitionDigest,
		"validator definition": input.ValidatorDefinitionDigest,
		"plan":                 input.Digest,
	} {
		if !validDigest(digest) {
			return MutationPlan{}, fmt.Errorf("%w: %s digest is malformed", ErrInvalidProtocolValue, name)
		}
	}
	if len(input.Operations) == 0 || len(input.Preconditions) == 0 {
		return MutationPlan{}, fmt.Errorf("%w: plan operations and preconditions cannot be empty", ErrInvalidProtocolValue)
	}
	return MutationPlan{
		planID:                    input.PlanID,
		workspaceID:               input.WorkspaceID,
		workID:                    input.WorkID,
		revisionScope:             input.RevisionScope,
		expectedRevision:          input.ExpectedRevision,
		inputDigest:               input.InputDigest,
		policyDigest:              input.PolicyDigest,
		workflowDefinitionDigest:  input.WorkflowDefinitionDigest,
		validatorDefinitionDigest: input.ValidatorDefinitionDigest,
		operations:                append([]PlanOperation{}, input.Operations...),
		preconditions:             append([]PlanPrecondition{}, input.Preconditions...),
		digest:                    input.Digest,
		idempotencyKey:            input.IdempotencyKey,
		authorizable:              input.Authorizable,
	}, nil
}

func (p MutationPlan) PlanID() string                    { return p.planID }
func (p MutationPlan) WorkspaceID() string               { return p.workspaceID }
func (p MutationPlan) WorkID() string                    { return p.workID }
func (p MutationPlan) RevisionScope() RevisionScope      { return p.revisionScope }
func (p MutationPlan) ExpectedRevision() uint64          { return p.expectedRevision }
func (p MutationPlan) InputDigest() string               { return p.inputDigest }
func (p MutationPlan) PolicyDigest() string              { return p.policyDigest }
func (p MutationPlan) WorkflowDefinitionDigest() string  { return p.workflowDefinitionDigest }
func (p MutationPlan) ValidatorDefinitionDigest() string { return p.validatorDefinitionDigest }
func (p MutationPlan) Operations() []PlanOperation       { return append([]PlanOperation{}, p.operations...) }
func (p MutationPlan) Preconditions() []PlanPrecondition {
	return append([]PlanPrecondition{}, p.preconditions...)
}
func (p MutationPlan) Digest() string         { return p.digest }
func (p MutationPlan) IdempotencyKey() string { return p.idempotencyKey }
func (p MutationPlan) Authorizable() bool     { return p.authorizable }

func (p MutationPlan) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		PlanID                    string             `json:"planId"`
		WorkspaceID               string             `json:"workspaceId"`
		WorkID                    string             `json:"workId"`
		RevisionScope             RevisionScope      `json:"revisionScope"`
		ExpectedRevision          uint64             `json:"expectedRevision"`
		InputDigest               string             `json:"inputDigest"`
		PolicyDigest              string             `json:"policyDigest"`
		WorkflowDefinitionDigest  string             `json:"workflowDefinitionDigest"`
		ValidatorDefinitionDigest string             `json:"validatorDefinitionDigest"`
		Operations                []PlanOperation    `json:"operations"`
		Preconditions             []PlanPrecondition `json:"preconditions"`
		Digest                    string             `json:"digest"`
		IdempotencyKey            string             `json:"idempotencyKey,omitempty"`
		Authorizable              bool               `json:"authorizable"`
	}{
		PlanID:                    p.planID,
		WorkspaceID:               p.workspaceID,
		WorkID:                    p.workID,
		RevisionScope:             p.revisionScope,
		ExpectedRevision:          p.expectedRevision,
		InputDigest:               p.inputDigest,
		PolicyDigest:              p.policyDigest,
		WorkflowDefinitionDigest:  p.workflowDefinitionDigest,
		ValidatorDefinitionDigest: p.validatorDefinitionDigest,
		Operations:                nonNilSlice(p.operations),
		Preconditions:             nonNilSlice(p.preconditions),
		Digest:                    p.digest,
		IdempotencyKey:            p.idempotencyKey,
		Authorizable:              p.authorizable,
	})
}

type StartDryRunResult struct {
	plan MutationPlan
}

func NewStartDryRunResult(plan MutationPlan) Result {
	value := StartDryRunResult{plan: plan}
	return Result{kind: ResultStartDryRun, startDryRun: &value}
}

func (r StartDryRunResult) Plan() MutationPlan { return r.plan }

func (r StartDryRunResult) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Kind ResultKind   `json:"kind"`
		Plan MutationPlan `json:"plan"`
	}{Kind: ResultStartDryRun, Plan: r.plan})
}

// Result is a closed tagged value rather than an interface pointer. Exactly
// one payload is populated by its constructor.
type Result struct {
	kind        ResultKind
	status      *StatusResult
	version     *VersionResult
	doctor      *DoctorResult
	startDryRun *StartDryRunResult
}

func (r Result) Kind() ResultKind { return r.kind }

func (r Result) Status() (StatusResult, bool) {
	if r.status == nil {
		return StatusResult{}, false
	}
	return StatusResult{workItems: append([]WorkRef{}, r.status.workItems...)}, true
}

func (r Result) Version() (VersionResult, bool) {
	if r.version == nil {
		return VersionResult{}, false
	}
	value := *r.version
	value.buildInputs = cloneMap(value.buildInputs)
	return value, true
}

func (r Result) Doctor() (DoctorResult, bool) {
	if r.doctor == nil {
		return DoctorResult{}, false
	}
	return DoctorResult{checks: r.doctor.Checks()}, true
}

func (r Result) StartDryRun() (StartDryRunResult, bool) {
	if r.startDryRun == nil {
		return StartDryRunResult{}, false
	}
	return *r.startDryRun, true
}

func (r Result) MarshalJSON() ([]byte, error) {
	switch r.kind {
	case ResultStatus:
		return json.Marshal(r.status)
	case ResultVersion:
		return json.Marshal(r.version)
	case ResultDoctor:
		return json.Marshal(r.doctor)
	case ResultStartDryRun:
		return json.Marshal(r.startDryRun)
	default:
		return nil, fmt.Errorf("%w: unsupported result kind %q", ErrInvalidProtocolValue, r.kind)
	}
}

func validDigest(value string) bool {
	return digestPattern.MatchString(value)
}
