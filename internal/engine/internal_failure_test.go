package engine

import (
	"errors"
	"testing"

	"sdd.dev/sdd/internal/workspace"
)

func TestResidualFailureIsRedactedAndNeverMasqueradesAsWorkspaceDamage(t *testing.T) {
	t.Parallel()

	mapped := mapFailure(errors.New("secret low-level invariant detail"))
	if mapped.value.Code() != "DIAG_INTERNAL_ERROR" || mapped.exit != ExitWorkspace {
		t.Fatalf("residual mapping = %s/%d", mapped.value.Code(), mapped.exit)
	}
	if mapped.value.Source().Kind() != "engine" {
		t.Fatalf("residual source = %q", mapped.value.Source().Kind())
	}
	if mapped.value.Message() == "secret low-level invariant detail" {
		t.Fatal("residual mapping exposed the low-level error")
	}

	outcome := internalFailureOutcome(CommandStatus, "")
	if outcome.OK() || outcome.ExitCode() != ExitWorkspace {
		t.Fatalf("internal outcome = %#v", outcome)
	}
	if len(outcome.Diagnostics()) != 1 || outcome.Diagnostics()[0].Code() != "DIAG_INTERNAL_ERROR" {
		t.Fatalf("internal diagnostics = %#v", outcome.Diagnostics())
	}
}

func TestWorkspaceSnapshotFailuresKeepConflictAndConfigurationCategories(t *testing.T) {
	t.Parallel()

	tests := []struct {
		err      error
		wantCode string
		wantExit ExitCode
	}{
		{workspace.ErrSnapshotUnstable, "DIAG_SNAPSHOT_UNSTABLE", ExitConflict},
		{workspace.ErrRecoveryRequired, "DIAG_RECOVERY_REQUIRED", ExitConflict},
		{workspace.ErrLimitExceeded, "DIAG_LIMIT_EXCEEDED", ExitWorkspace},
		{workspace.ErrSnapshotInvalid, "DIAG_WORKSPACE_VERSION_INCOMPATIBLE", ExitWorkspace},
	}
	for _, test := range tests {
		mapped := mapFailure(test.err)
		if mapped.value.Code() != test.wantCode || mapped.exit != test.wantExit {
			t.Fatalf("%v mapped to %s/%d, want %s/%d", test.err, mapped.value.Code(), mapped.exit, test.wantCode, test.wantExit)
		}
	}
}
