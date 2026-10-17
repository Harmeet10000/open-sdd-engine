package protocol

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"sdd.dev/sdd/internal/engine"
	"sdd.dev/sdd/internal/presentation/render"
	"sdd.dev/sdd/internal/testkit"
)

var errConformanceFault = errors.New("conformance fault injection")

func conformanceEngine(t *testing.T, files map[string]testkit.MemFile) *engine.Engine {
	t.Helper()
	policy, err := engine.NewPolicySnapshot(engine.PolicySnapshotInput{Source: "test-policy", Floor: []byte("floor: strict\n")})
	if err != nil {
		t.Fatal(err)
	}
	eng, err := engine.NewEngine(engine.Dependencies{
		Facts:   testkit.NewMemFileFacts(files),
		IDs:     testkit.NewSequenceIDSource("work"),
		Policy:  testkit.NewFixedPolicySnapshotter(policy),
		Workdir: testkit.NativePath("/repo"),
		Build:   engine.BuildMetadata{Version: "0.1.0-dev", Commit: "abc", BuildTime: "unknown", GoVersion: "go1.24.0"},
	})
	if err != nil {
		t.Fatal(err)
	}
	return eng
}

func conformanceLayoutFiles() map[string]testkit.MemFile {
	return map[string]testkit.MemFile{
		"/repo/.sdd/workspace.yaml":       {Bytes: []byte("formatVersion: 1\nworkspaceId: ws-1\nrevision: 11\n")},
		"/repo/.sdd/work/apple/work.yaml": {Bytes: []byte("schemaVersion: 1\nworkId: work-1\nslug: apple\nrevision: 5\nstatus: active\ncomposition: quick-spec\n")},
	}
}

func assertEngineOutcomeValid(t *testing.T, eng *engine.Engine, outcome engine.Outcome) {
	t.Helper()
	schema := compileFoundationSchema(t, "")
	var output bytes.Buffer
	input := render.NewInput(outcome, render.NewRedactor())
	if err := render.WriteJSON(&output, input); err != nil {
		t.Fatal(err)
	}
	if err := schema.Validate(decodeJSONBytes(t, output.Bytes())); err != nil {
		t.Fatalf("engine outcome rejected by published schema: %v\n%s", err, output.Bytes())
	}
	_ = eng
}

func TestEngineOutcomesConformToPublishedSchema(t *testing.T) {
	t.Parallel()

	eng := conformanceEngine(t, conformanceLayoutFiles())
	start, err := engine.NewStartDryRunCommand("add export", "quick-spec", "")
	if err != nil {
		t.Fatal(err)
	}
	outcomes := map[string]engine.Outcome{
		"version":        eng.Execute(context.Background(), engine.NewVersionCommand()),
		"status":         eng.Execute(context.Background(), engine.NewStatusCommand("", "")),
		"doctor":         eng.Execute(context.Background(), engine.NewDoctorCommand("")),
		"doctor-failure": conformanceEngine(t, nil).Execute(context.Background(), engine.NewDoctorCommand("")),
		"start-dry-run":  eng.Execute(context.Background(), start),
		"unknown":        eng.UnknownCommand("bogus"),
	}
	for name, outcome := range outcomes {
		outcome := outcome
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			assertEngineOutcomeValid(t, eng, outcome)
		})
	}
}

func TestEngineVersionIgnoresBrokenFilesystem(t *testing.T) {
	t.Parallel()

	facts := testkit.NewMemFileFacts(nil)
	facts.InjectFault("/nonexistent", errConformanceFault)
	policy, err := engine.NewPolicySnapshot(engine.PolicySnapshotInput{Source: "test-policy"})
	if err != nil {
		t.Fatal(err)
	}
	snapshotter := testkit.NewFixedPolicySnapshotter(policy)
	snapshotter.InjectFault(errConformanceFault)
	eng, err := engine.NewEngine(engine.Dependencies{
		Facts:   facts,
		IDs:     testkit.NewSequenceIDSource("work"),
		Policy:  snapshotter,
		Workdir: testkit.NativePath("/nonexistent"),
		Build:   engine.BuildMetadata{Version: "0.1.0-dev"},
	})
	if err != nil {
		t.Fatal(err)
	}
	outcome := eng.Execute(context.Background(), engine.NewVersionCommand())
	if !outcome.OK() {
		t.Fatalf("version depends on workspace or policy: %#v", outcome)
	}
	assertEngineOutcomeValid(t, eng, outcome)
}

func TestEngineDryRunIsDeterministicUnderFixedIDs(t *testing.T) {
	t.Parallel()

	renderOnce := func() []byte {
		eng := conformanceEngine(t, conformanceLayoutFiles())
		start, err := engine.NewStartDryRunCommand("add export", "quick-spec", "")
		if err != nil {
			t.Fatal(err)
		}
		var output bytes.Buffer
		input := render.NewInput(eng.Execute(context.Background(), start), render.NewRedactor())
		if err := render.WriteJSON(&output, input); err != nil {
			t.Fatal(err)
		}
		return output.Bytes()
	}
	first, second := renderOnce(), renderOnce()
	if !bytes.Equal(first, second) {
		t.Fatalf("fixed IDs produced different dry-run envelopes:\n%s\n%s", first, second)
	}
}
