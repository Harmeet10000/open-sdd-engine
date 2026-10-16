package engine

import (
	"encoding/json"
	"fmt"
)

type Outcome struct {
	command          CommandKind
	requestedCommand string
	workspace        *WorkspaceRef
	work             *WorkRef
	result           *Result
	diagnostics      []Diagnostic
	nextActions      []NextAction
	exitCode         ExitCode
}

type OutcomeInput struct {
	Command          CommandKind
	RequestedCommand string
	Workspace        *WorkspaceRef
	Work             *WorkRef
	Result           *Result
	Diagnostics      []Diagnostic
	NextActions      []NextAction
	ExitCode         ExitCode
}

func NewOutcome(input OutcomeInput) (Outcome, error) {
	if !input.ExitCode.valid() {
		return Outcome{}, fmt.Errorf("%w: unsupported exit code %d", ErrInvalidProtocolValue, input.ExitCode)
	}
	if !validOutcomeCommand(input.Command) {
		return Outcome{}, fmt.Errorf("%w: unsupported outcome command %q", ErrInvalidProtocolValue, input.Command)
	}
	if input.ExitCode == ExitSuccess && input.Result == nil {
		return Outcome{}, fmt.Errorf("%w: successful outcome requires a result", ErrInvalidProtocolValue)
	}
	if input.ExitCode != ExitSuccess && len(input.Diagnostics) == 0 {
		return Outcome{}, fmt.Errorf("%w: failed outcome requires a diagnostic", ErrInvalidProtocolValue)
	}
	if input.Result != nil && !resultMatchesCommand(input.Command, input.Result.Kind()) {
		return Outcome{}, fmt.Errorf(
			"%w: result kind %q does not match command %q",
			ErrInvalidProtocolValue,
			input.Result.Kind(),
			input.Command,
		)
	}
	if input.Command == CommandUnknown {
		if input.RequestedCommand == "" || input.ExitCode != ExitUsage || input.Result != nil {
			return Outcome{}, fmt.Errorf("%w: unknown command requires requested input, usage exit, and null result", ErrInvalidProtocolValue)
		}
	} else if input.RequestedCommand != "" {
		return Outcome{}, fmt.Errorf("%w: requestedCommand is reserved for unknown commands", ErrInvalidProtocolValue)
	}
	if input.ExitCode == ExitSuccess && requiresWorkspace(input.Command) && input.Workspace == nil {
		return Outcome{}, fmt.Errorf("%w: successful %s outcome requires workspace context", ErrInvalidProtocolValue, input.Command)
	}

	return Outcome{
		command:          input.Command,
		requestedCommand: input.RequestedCommand,
		workspace:        clonePointer(input.Workspace),
		work:             clonePointer(input.Work),
		result:           cloneResultPointer(input.Result),
		diagnostics:      cloneDiagnostics(input.Diagnostics),
		nextActions:      cloneActions(input.NextActions),
		exitCode:         input.ExitCode,
	}, nil
}

func (o Outcome) SchemaVersion() string     { return SchemaVersion }
func (o Outcome) Command() CommandKind      { return o.command }
func (o Outcome) RequestedCommand() string  { return o.requestedCommand }
func (o Outcome) OK() bool                  { return o.exitCode == ExitSuccess }
func (o Outcome) Workspace() *WorkspaceRef  { return clonePointer(o.workspace) }
func (o Outcome) Work() *WorkRef            { return clonePointer(o.work) }
func (o Outcome) Result() *Result           { return cloneResultPointer(o.result) }
func (o Outcome) Diagnostics() []Diagnostic { return cloneDiagnostics(o.diagnostics) }
func (o Outcome) NextActions() []NextAction { return cloneActions(o.nextActions) }
func (o Outcome) ExitCode() ExitCode        { return o.exitCode }

func (o Outcome) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		SchemaVersion    string        `json:"schemaVersion"`
		Command          CommandKind   `json:"command"`
		RequestedCommand string        `json:"requestedCommand,omitempty"`
		OK               bool          `json:"ok"`
		Workspace        *WorkspaceRef `json:"workspace"`
		Work             *WorkRef      `json:"work"`
		Result           *Result       `json:"result"`
		Diagnostics      []Diagnostic  `json:"diagnostics"`
		NextActions      []NextAction  `json:"nextActions"`
		ExitCode         ExitCode      `json:"exitCode"`
	}{
		SchemaVersion:    SchemaVersion,
		Command:          o.command,
		RequestedCommand: o.requestedCommand,
		OK:               o.OK(),
		Workspace:        o.workspace,
		Work:             o.work,
		Result:           o.result,
		Diagnostics:      nonNilSlice(o.diagnostics),
		NextActions:      nonNilSlice(o.nextActions),
		ExitCode:         o.exitCode,
	})
}

// Redacted creates a new immutable outcome for presentation. The source
// outcome remains exact and may continue to participate in decisions and
// digest verification; only the presentation copy is transformed.
func (o Outcome) Redacted(redact func(string) string) Outcome {
	if redact == nil {
		return o
	}
	value := o
	value.requestedCommand = redact(value.requestedCommand)
	value.workspace = redactWorkspace(value.workspace, redact)
	value.work = redactWork(value.work, redact)
	value.result = redactResult(value.result, redact)
	value.diagnostics = redactDiagnostics(value.diagnostics, redact)
	value.nextActions = redactActions(value.nextActions, redact)
	return value
}

func validOutcomeCommand(command CommandKind) bool {
	switch command {
	case CommandStatus, CommandVersion, CommandDoctor, CommandStart, CommandUnknown:
		return true
	default:
		return false
	}
}

func resultMatchesCommand(command CommandKind, result ResultKind) bool {
	switch command {
	case CommandStatus:
		return result == ResultStatus
	case CommandVersion:
		return result == ResultVersion
	case CommandDoctor:
		return result == ResultDoctor
	case CommandStart:
		return result == ResultStartDryRun
	case CommandUnknown:
		return false
	default:
		return false
	}
}

func requiresWorkspace(command CommandKind) bool {
	return command == CommandStatus || command == CommandDoctor || command == CommandStart
}

func cloneResultPointer(value *Result) *Result {
	if value == nil {
		return nil
	}
	copyValue := Result{kind: value.kind}
	if status, ok := value.Status(); ok {
		copyValue.status = &status
	}
	if version, ok := value.Version(); ok {
		copyValue.version = &version
	}
	if doctor, ok := value.Doctor(); ok {
		copyValue.doctor = &doctor
	}
	if start, ok := value.StartDryRun(); ok {
		copyValue.startDryRun = &start
	}
	return &copyValue
}

func cloneDiagnostics(values []Diagnostic) []Diagnostic {
	cloned := make([]Diagnostic, 0, len(values))
	for _, value := range values {
		value.metadata = cloneStoredJSONObject(value.metadata)
		cloned = append(cloned, value)
	}
	return cloned
}

func cloneActions(values []NextAction) []NextAction {
	cloned := make([]NextAction, 0, len(values))
	for _, value := range values {
		value.args = append([]string{}, value.args...)
		value.flags = cloneMap(value.flags)
		value.expectedRevision = clonePointer(value.expectedRevision)
		cloned = append(cloned, value)
	}
	return cloned
}

func redactWorkspace(value *WorkspaceRef, redact func(string) string) *WorkspaceRef {
	if value == nil {
		return nil
	}
	copyValue := *value
	copyValue.root = redact(copyValue.root)
	return &copyValue
}

func redactWork(value *WorkRef, redact func(string) string) *WorkRef {
	if value == nil {
		return nil
	}
	copyValue := *value
	copyValue.id = redact(copyValue.id)
	copyValue.slug = redact(copyValue.slug)
	copyValue.workflow = redact(copyValue.workflow)
	return &copyValue
}

func redactDiagnostics(values []Diagnostic, redact func(string) string) []Diagnostic {
	redacted := cloneDiagnostics(values)
	for index := range redacted {
		diagnostic := &redacted[index]
		diagnostic.message = redact(diagnostic.message)
		diagnostic.remediation = redact(diagnostic.remediation)
		diagnostic.source.path = redact(diagnostic.source.path)
		diagnostic.source.field = redact(diagnostic.source.field)
		diagnostic.source.artifactID = redact(diagnostic.source.artifactID)
		if diagnostic.source.rangeValue != nil {
			diagnostic.source.rangeValue.path = redact(diagnostic.source.rangeValue.path)
		}
		diagnostic.metadata = redactJSONObject(diagnostic.metadata, redact)
	}
	return redacted
}

func redactActions(values []NextAction, redact func(string) string) []NextAction {
	redacted := cloneActions(values)
	for index := range redacted {
		action := &redacted[index]
		action.command = redact(action.command)
		action.reason = redact(action.reason)
		for argIndex := range action.args {
			action.args[argIndex] = redact(action.args[argIndex])
		}
		for name, flag := range action.flags {
			if flag.kind == flagString {
				flag.stringValue = redact(flag.stringValue)
				action.flags[name] = flag
			}
		}
	}
	return redacted
}

func redactResult(value *Result, redact func(string) string) *Result {
	cloned := cloneResultPointer(value)
	if cloned == nil {
		return nil
	}
	if cloned.status != nil {
		for index := range cloned.status.workItems {
			item := &cloned.status.workItems[index]
			item.id = redact(item.id)
			item.slug = redact(item.slug)
			item.workflow = redact(item.workflow)
		}
	}
	if cloned.version != nil {
		cloned.version.version = redact(cloned.version.version)
		cloned.version.commit = redact(cloned.version.commit)
		cloned.version.buildTime = redact(cloned.version.buildTime)
		cloned.version.goVersion = redact(cloned.version.goVersion)
		for key, item := range cloned.version.buildInputs {
			cloned.version.buildInputs[key] = redact(item)
		}
	}
	if cloned.doctor != nil {
		for index, check := range cloned.doctor.checks {
			cloned.doctor.checks[index] = redactJSONObject(check, redact)
		}
	}
	if cloned.startDryRun != nil {
		plan := &cloned.startDryRun.plan
		plan.planID = redact(plan.planID)
		plan.workspaceID = redact(plan.workspaceID)
		plan.workID = redact(plan.workID)
		plan.idempotencyKey = redact(plan.idempotencyKey)
		for index := range plan.operations {
			operation := &plan.operations[index]
			operation.path = redact(operation.path)
			if operation.content != nil {
				redactedContent := redact(*operation.content)
				operation.content = &redactedContent
			}
		}
		for index := range plan.preconditions {
			plan.preconditions[index].path = redact(plan.preconditions[index].path)
		}
	}
	return cloned
}

func redactJSONObject(value map[string]any, redact func(string) string) map[string]any {
	redacted := make(map[string]any, len(value))
	for key, item := range value {
		redacted[key] = redactJSONValue(item, redact)
	}
	return redacted
}

func redactJSONValue(value any, redact func(string) string) any {
	switch typed := value.(type) {
	case string:
		return redact(typed)
	case []any:
		items := make([]any, len(typed))
		for index, item := range typed {
			items[index] = redactJSONValue(item, redact)
		}
		return items
	case map[string]any:
		return redactJSONObject(typed, redact)
	default:
		return typed
	}
}
