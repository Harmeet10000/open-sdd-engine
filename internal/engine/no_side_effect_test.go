package engine_test

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"sdd.dev/sdd/internal/engine"
	"sdd.dev/sdd/internal/testkit"
)

type treeFact struct {
	Mode    fs.FileMode
	Size    int64
	ModTime time.Time
	Bytes   []byte
}

func TestReadOnlyAndDryRunCommandsPreserveNativeTreeFacts(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	marker := filepath.Join(root, ".sdd")
	if err := os.MkdirAll(filepath.Join(marker, "work"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(marker, "workspace.yaml"), []byte(discoveryWorkspace), 0o640); err != nil {
		t.Fatal(err)
	}
	// Windows may publish directory modification times lazily after child
	// creation. Pin every timestamp after setup so the before-snapshot cannot
	// race delayed metadata propagation and falsely attribute it to a read.
	fixedTime := time.Unix(1_700_000_000, 0)
	for _, path := range []string{
		filepath.Join(marker, "workspace.yaml"),
		filepath.Join(marker, "work"),
		marker,
		root,
	} {
		if err := os.Chtimes(path, fixedTime, fixedTime); err != nil {
			t.Fatalf("stabilize %s timestamps: %v", path, err)
		}
	}
	policy, err := engine.NewPolicySnapshot(engine.PolicySnapshotInput{
		Source: "test-policy",
		Floor:  []byte("floor: strict\n"),
	})
	if err != nil {
		t.Fatal(err)
	}
	eng, err := engine.NewEngine(engine.Dependencies{
		Facts:   engine.OSFileFacts{},
		IDs:     testkit.NewSequenceIDSource("work"),
		Policy:  testkit.NewFixedPolicySnapshotter(policy),
		Workdir: root,
		Build:   engine.BuildMetadata{Version: "0.1.0-test"},
	})
	if err != nil {
		t.Fatal(err)
	}
	start, err := engine.NewStartDryRunCommand("add export", "quick-spec", "")
	if err != nil {
		t.Fatal(err)
	}

	commands := []engine.Command{
		engine.NewStatusCommand("", ""),
		engine.NewDoctorCommand(""),
		start,
	}
	for _, command := range commands {
		before := captureTreeFacts(t, root)
		outcome := eng.Execute(context.Background(), command)
		if !outcome.OK() {
			t.Fatalf("%s failed: %#v", command.Kind(), outcome.Diagnostics())
		}
		after := captureTreeFacts(t, root)
		if !reflect.DeepEqual(before, after) {
			t.Fatalf("%s changed native tree facts\nbefore: %#v\nafter:  %#v", command.Kind(), before, after)
		}
	}
}

func captureTreeFacts(t *testing.T, root string) map[string]treeFact {
	t.Helper()
	facts := map[string]treeFact{}
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		fact := treeFact{Mode: info.Mode(), Size: info.Size(), ModTime: info.ModTime()}
		if info.Mode().IsRegular() {
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			fact.Bytes = content
		}
		facts[filepath.ToSlash(relative)] = fact
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return facts
}
