package filesystem

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"sdd.dev/sdd/internal/workspace"
)

func TestCapturerReadsOneStableBoundedGeneration(t *testing.T) {
	t.Parallel()

	root := makeWorkspace(t, workspace.NewGenerationRecord(7, ""))
	writeFixture(t, root, ".sdd/workspace.yaml", []byte("workspaceId: ws-1\n"))
	path := mustWorkspacePath(t, ".sdd/workspace.yaml")
	capturer := mustCapturer(t, CaptureConfig{
		Root: root,
		Workspace: workspace.WorkspaceObservation{
			WorkspaceID: "ws-1", CanonicalRoot: root, FormatVersion: 1, Revision: 3,
		},
		NewID: func() (string, error) { return "snap-fixed", nil },
	})
	reader, err := workspace.NewReader(capturer)
	if err != nil {
		t.Fatal(err)
	}

	snapshot, err := reader.Snapshot(context.Background(), workspace.SnapshotRequest{
		Selection: workspace.Selection{ExplicitRoot: root, WorkspaceID: "ws-1"},
		Scope:     workspace.ScopeSelection,
		Paths:     []workspace.WorkspacePath{path},
		Limits:    workspace.Limits{MaxPaths: 4, MaxBytes: 4096, StableReadRetries: 1},
	})
	if err != nil {
		t.Fatal(err)
	}
	if snapshot.SnapshotID != "snap-fixed" || snapshot.Generation != 7 || snapshot.ReadSetDigest == "" {
		t.Fatalf("snapshot identity = %#v", snapshot)
	}
	if got := string(snapshot.Paths[0].Bytes); got != "workspaceId: ws-1\n" {
		t.Fatalf("captured bytes = %q", got)
	}
	if snapshot.Paths[0].NativeID == (workspace.NativeFileIdentity{}) && runtime.GOOS == "linux" {
		t.Fatal("linux capture omitted native file identity")
	}
}

func TestCapturerReportsActiveTransactionWithoutReadingDomainFiles(t *testing.T) {
	t.Parallel()

	root := makeWorkspace(t, workspace.NewGenerationRecord(8, "tx-prepared"))
	capturer := mustCapturer(t, CaptureConfig{
		Root: root,
		Workspace: workspace.WorkspaceObservation{
			WorkspaceID: "ws-1", CanonicalRoot: root, FormatVersion: 1, Revision: 3,
		},
	})
	reader, err := workspace.NewReader(capturer)
	if err != nil {
		t.Fatal(err)
	}
	_, err = reader.Snapshot(context.Background(), workspace.SnapshotRequest{
		Selection: workspace.Selection{WorkspaceID: "ws-1"},
		Scope:     workspace.ScopeRecoveryStatus,
		Paths:     []workspace.WorkspacePath{mustWorkspacePath(t, "missing")},
		Limits:    workspace.Limits{MaxPaths: 4, MaxBytes: 4096, StableReadRetries: 0},
	})
	if !errors.Is(err, workspace.ErrRecoveryRequired) {
		t.Fatalf("Snapshot() error = %v", err)
	}
}

func TestCapturerRejectsSymlinkComponents(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink creation requires privileges on some Windows runners")
	}
	t.Parallel()

	root := makeWorkspace(t, workspace.NewGenerationRecord(1, ""))
	outside := t.TempDir()
	writeFixture(t, outside, "secret", []byte("outside"))
	if err := os.Symlink(outside, filepath.Join(root, "linked")); err != nil {
		t.Fatal(err)
	}
	capturer := mustCapturer(t, CaptureConfig{
		Root: root,
		Workspace: workspace.WorkspaceObservation{
			WorkspaceID: "ws-1", CanonicalRoot: root, FormatVersion: 1, Revision: 1,
		},
	})

	_, err := capturer.Capture(context.Background(), workspace.SnapshotRequest{
		Selection: workspace.Selection{WorkspaceID: "ws-1"},
		Scope:     workspace.ScopeSelection,
		Paths:     []workspace.WorkspacePath{mustWorkspacePath(t, "linked/secret")},
		Limits:    workspace.Limits{MaxPaths: 4, MaxBytes: 4096, StableReadRetries: 0},
	})
	if !errors.Is(err, ErrPathUnsafe) {
		t.Fatalf("Capture() error = %v", err)
	}
}

func TestCapturerDoesNotWriteDuringCapture(t *testing.T) {
	t.Parallel()

	root := makeWorkspace(t, workspace.NewGenerationRecord(2, ""))
	writeFixture(t, root, "input.txt", []byte("value"))
	before := treeState(t, root)
	capturer := mustCapturer(t, CaptureConfig{
		Root: root,
		Workspace: workspace.WorkspaceObservation{
			WorkspaceID: "ws-1", CanonicalRoot: root, FormatVersion: 1, Revision: 1,
		},
	})
	_, err := capturer.Capture(context.Background(), workspace.SnapshotRequest{
		Selection: workspace.Selection{WorkspaceID: "ws-1"},
		Scope:     workspace.ScopeSelection,
		Paths:     []workspace.WorkspacePath{mustWorkspacePath(t, "input.txt")},
		Limits:    workspace.Limits{MaxPaths: 4, MaxBytes: 4096, StableReadRetries: 0},
	})
	if err != nil {
		t.Fatal(err)
	}
	after := treeState(t, root)
	if string(before) != string(after) {
		t.Fatalf("read-only capture changed workspace\nbefore: %s\nafter: %s", before, after)
	}
}

func makeWorkspace(t *testing.T, generation workspace.GenerationRecord) string {
	t.Helper()
	root := t.TempDir()
	data, err := json.Marshal(generation)
	if err != nil {
		t.Fatal(err)
	}
	writeFixture(t, root, string(workspace.GenerationRecordPath), data)
	return root
}

func writeFixture(t *testing.T, root, name string, data []byte) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(name))
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
}

func mustCapturer(t *testing.T, config CaptureConfig) *Capturer {
	t.Helper()
	result, err := NewCapturer(config)
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func mustWorkspacePath(t *testing.T, value string) workspace.WorkspacePath {
	t.Helper()
	result, err := workspace.ParseWorkspacePath(value)
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func treeState(t *testing.T, root string) []byte {
	t.Helper()
	var result []byte
	err := filepath.WalkDir(root, func(name string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(root, name)
		if err != nil {
			return err
		}
		result = append(result, []byte(relative)...)
		result = append(result, 0)
		if !entry.IsDir() {
			data, err := os.ReadFile(name)
			if err != nil {
				return err
			}
			result = append(result, data...)
			result = append(result, 0)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return result
}
