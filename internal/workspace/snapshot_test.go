package workspace

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

type captureFunc func(context.Context, SnapshotRequest) (Capture, error)

func (f captureFunc) Capture(ctx context.Context, request SnapshotRequest) (Capture, error) {
	return f(ctx, request)
}

func TestSnapshotExtensionBuildsOwnedDeterministicRawSnapshot(t *testing.T) {
	request := validRequest(t, ScopeArtifactIdentities)
	request.Paths = []WorkspacePath{mustPath(t, ".sdd/work/export/work.yaml"), mustPath(t, ".sdd/workspace.yaml")}
	bytes := []byte("workspaceId: ws-1\n")
	allocated := uint64(3)
	current := uint64(8)
	artifactPath := mustPath(t, ".sdd/work/export/requirements.md")

	calls := 0
	reader := mustReader(t, captureFunc(func(_ context.Context, got SnapshotRequest) (Capture, error) {
		calls++
		if !reflect.DeepEqual(got.Paths, request.Paths) {
			t.Fatalf("request paths were not normalized deterministically: %v", got.Paths)
		}
		return Capture{
			SnapshotID:       "snap-1",
			Workspace:        WorkspaceObservation{WorkspaceID: "ws-1", CanonicalRoot: "/repo", FormatVersion: 1, Revision: 11},
			Work:             &WorkObservation{WorkID: "work-1", Slug: "export", Revision: 5},
			GenerationBefore: 9,
			GenerationAfter:  9,
			Paths: []PathObservation{
				{Path: request.Paths[1], Presence: Present, Kind: FileKindRegular, Bytes: bytes, Size: uint64(len(bytes)), Mode: 0o644},
				{Path: request.Paths[0], Presence: Absent, Kind: FileKindNone},
			},
			ArtifactIdentities: []ArtifactIdentityObservation{{
				WorkspaceID: "ws-1", WorkID: "work-1", Namespace: "requirement", ArtifactID: "REQ-001",
				State: ArtifactAllocated, AllocationRevision: &allocated,
				AllocationProvenance:    &ProvenanceRef{Kind: "transaction", ID: "tx-1", Digest: "sha256:allocation"},
				CurrentArtifactRevision: &current, CurrentPath: &artifactPath,
			}},
		}, nil
	}))

	snapshot, err := reader.Snapshot(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	if calls != 1 {
		t.Fatalf("snapshot bridge captured %d times; expected exactly once", calls)
	}
	if snapshot.ReadSetDigest == "" || snapshot.Generation != 9 {
		t.Fatalf("snapshot identity is incomplete: %#v", snapshot)
	}
	if snapshot.Paths[0].Path != request.Paths[0] || snapshot.Paths[1].Path != request.Paths[1] {
		t.Fatalf("paths are not in canonical order: %v", snapshot.Paths)
	}
	if snapshot.Paths[1].BlobDigest == "" {
		t.Fatal("present file did not receive a raw blob digest")
	}

	bytes[0] = 'X'
	if string(snapshot.Paths[1].Bytes) != "workspaceId: ws-1\n" {
		t.Fatal("snapshot aliases adapter-owned bytes")
	}

	second, err := acceptCapture(request, Capture{
		SnapshotID:         "snap-2",
		Workspace:          snapshot.Workspace,
		Work:               snapshot.Work,
		GenerationBefore:   9,
		GenerationAfter:    9,
		Paths:              []PathObservation{snapshot.Paths[1], snapshot.Paths[0]},
		ArtifactIdentities: snapshot.ArtifactIdentities,
	})
	if err != nil {
		t.Fatal(err)
	}
	if second.ReadSetDigest != snapshot.ReadSetDigest {
		t.Fatalf("enumeration order changed read-set identity: %s != %s", second.ReadSetDigest, snapshot.ReadSetDigest)
	}
}

func TestSnapshotExtensionRejectsMixedGeneration(t *testing.T) {
	request := validRequest(t, ScopeSelection)
	reader := mustReader(t, captureFunc(func(context.Context, SnapshotRequest) (Capture, error) {
		capture := validCapture(request)
		capture.GenerationAfter++
		return capture, nil
	}))

	_, err := reader.Snapshot(context.Background(), request)
	if !errors.Is(err, ErrSnapshotUnstable) {
		t.Fatalf("got %v; want ErrSnapshotUnstable", err)
	}
}

func TestSnapshotExtensionRejectsActiveTransaction(t *testing.T) {
	request := validRequest(t, ScopeSelection)
	reader := mustReader(t, captureFunc(func(context.Context, SnapshotRequest) (Capture, error) {
		capture := validCapture(request)
		capture.ActiveTransactionID = "tx-prepared"
		return capture, nil
	}))

	_, err := reader.Snapshot(context.Background(), request)
	if !errors.Is(err, ErrRecoveryRequired) {
		t.Fatalf("got %v; want ErrRecoveryRequired", err)
	}
}

func TestSnapshotExtensionHonorsCancellationBeforeCapture(t *testing.T) {
	request := validRequest(t, ScopeSelection)
	called := false
	reader := mustReader(t, captureFunc(func(context.Context, SnapshotRequest) (Capture, error) {
		called = true
		return Capture{}, nil
	}))
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := reader.Snapshot(ctx, request)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("got %v; want context cancellation", err)
	}
	if called {
		t.Fatal("capture adapter was called after cancellation")
	}
}

func TestSnapshotExtensionRejectsInventedArtifactAllocation(t *testing.T) {
	request := validRequest(t, ScopeArtifactIdentities)
	reader := mustReader(t, captureFunc(func(context.Context, SnapshotRequest) (Capture, error) {
		capture := validCapture(request)
		fakeRevision := uint64(0)
		capture.ArtifactIdentities = []ArtifactIdentityObservation{{
			WorkspaceID: "ws-1", WorkID: "work-1", Namespace: "requirement", ArtifactID: "REQ-404",
			State: ArtifactMissingRecord, AllocationRevision: &fakeRevision,
		}}
		return capture, nil
	}))

	_, err := reader.Snapshot(context.Background(), request)
	if !errors.Is(err, ErrSnapshotInvalid) {
		t.Fatalf("got %v; want missing-record integrity rejection", err)
	}
}

func TestSnapshotExtensionRejectsZeroPersistedArtifactRevision(t *testing.T) {
	request := validRequest(t, ScopeArtifactIdentities)
	reader := mustReader(t, captureFunc(func(context.Context, SnapshotRequest) (Capture, error) {
		capture := validCapture(request)
		zero := uint64(0)
		current := uint64(1)
		artifactPath := mustPath(t, ".sdd/work/export/requirements.md")
		capture.ArtifactIdentities = []ArtifactIdentityObservation{{
			WorkspaceID: "ws-1", WorkID: "work-1", Namespace: "requirement", ArtifactID: "REQ-001",
			State: ArtifactAllocated, AllocationRevision: &zero,
			AllocationProvenance:    &ProvenanceRef{Kind: "transaction", ID: "tx-1", Digest: "sha256:allocation"},
			CurrentArtifactRevision: &current, CurrentPath: &artifactPath,
		}}
		return capture, nil
	}))

	_, err := reader.Snapshot(context.Background(), request)
	if !errors.Is(err, ErrSnapshotInvalid) {
		t.Fatalf("got %v; want zero persisted revision rejection", err)
	}
}

func TestSnapshotExtensionEnforcesBoundedReadSet(t *testing.T) {
	request := validRequest(t, ScopeSelection)
	request.Limits.MaxBytes = 1
	reader := mustReader(t, captureFunc(func(context.Context, SnapshotRequest) (Capture, error) {
		return validCapture(request), nil
	}))

	_, err := reader.Snapshot(context.Background(), request)
	if !errors.Is(err, ErrLimitExceeded) {
		t.Fatalf("got %v; want ErrLimitExceeded", err)
	}
}

func TestReadSetDigestIgnoresNativeRootAndBindsSafetyFacts(t *testing.T) {
	request := validRequest(t, ScopeArtifactIdentities)
	base := validCapture(request)
	linkCount := uint64(1)
	base.Paths[0].LinkCount = &linkCount
	allocated := uint64(2)
	current := uint64(4)
	artifactPath := mustPath(t, ".sdd/work/export/requirements.md")
	base.ArtifactIdentities = []ArtifactIdentityObservation{{
		WorkspaceID: "ws-1", WorkID: "work-1", Namespace: "requirement", ArtifactID: "REQ-001",
		State: ArtifactAllocated, AllocationRevision: &allocated,
		AllocationProvenance:    &ProvenanceRef{Kind: "transaction", ID: "tx-1", Digest: "sha256:allocation"},
		CurrentArtifactRevision: &current, CurrentPath: &artifactPath,
	}}

	first, err := acceptCapture(request, base)
	if err != nil {
		t.Fatal(err)
	}

	relocated := base
	relocated.Workspace.CanonicalRoot = "/relocated/repo"
	second, err := acceptCapture(request, relocated)
	if err != nil {
		t.Fatal(err)
	}
	if first.ReadSetDigest != second.ReadSetDigest {
		t.Fatal("native root leaked into read-set identity")
	}

	linked := base
	linked.Paths = clonePaths(base.Paths)
	moreLinks := uint64(2)
	linked.Paths[0].LinkCount = &moreLinks
	third, err := acceptCapture(request, linked)
	if err != nil {
		t.Fatal(err)
	}
	if first.ReadSetDigest == third.ReadSetDigest {
		t.Fatal("hard-link safety fact did not affect read-set identity")
	}

	revised := base
	revised.ArtifactIdentities = cloneArtifactIdentities(base.ArtifactIdentities)
	newRevision := uint64(5)
	revised.ArtifactIdentities[0].CurrentArtifactRevision = &newRevision
	fourth, err := acceptCapture(request, revised)
	if err != nil {
		t.Fatal(err)
	}
	if first.ReadSetDigest == fourth.ReadSetDigest {
		t.Fatal("artifact revision did not affect read-set identity")
	}
}

func TestParseWorkspacePathRejectsHostDependentOrEscapingForms(t *testing.T) {
	t.Parallel()

	for _, value := range []string{"", "/absolute", "../escape", "dir/../escape", `dir\escape`, "C:/absolute", "dir//file", "dir/\x00file", "dir/\nfile"} {
		if _, err := ParseWorkspacePath(value); !errors.Is(err, ErrSnapshotInvalid) {
			t.Fatalf("ParseWorkspacePath(%q) error = %v", value, err)
		}
	}
	if got, err := ParseWorkspacePath(".sdd/work/item/work.yaml"); err != nil || got != ".sdd/work/item/work.yaml" {
		t.Fatalf("canonical path = %q, %v", got, err)
	}
}

func validRequest(t *testing.T, scope SnapshotScope) SnapshotRequest {
	t.Helper()
	return SnapshotRequest{
		Selection: Selection{WorkspaceID: "ws-1", WorkID: "work-1"},
		Scope:     scope,
		Paths:     []WorkspacePath{mustPath(t, ".sdd/workspace.yaml")},
		Limits:    Limits{MaxPaths: 16, MaxBytes: 1 << 20, StableReadRetries: 2},
	}
}

func validCapture(request SnapshotRequest) Capture {
	content := []byte("workspaceId: ws-1\n")
	return Capture{
		SnapshotID:       "snap-1",
		Workspace:        WorkspaceObservation{WorkspaceID: "ws-1", CanonicalRoot: "/repo", FormatVersion: 1, Revision: 1},
		Work:             &WorkObservation{WorkID: "work-1", Slug: "export", Revision: 1},
		GenerationBefore: 1,
		GenerationAfter:  1,
		Paths: []PathObservation{{
			Path: request.Paths[0], Presence: Present, Kind: FileKindRegular,
			Bytes: content, Size: uint64(len(content)), Mode: 0o644,
		}},
	}
}

func mustPath(t *testing.T, value string) WorkspacePath {
	t.Helper()
	result, err := ParseWorkspacePath(value)
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func mustReader(t *testing.T, capturer Capturer) *Reader {
	t.Helper()
	reader, err := NewReader(capturer)
	if err != nil {
		t.Fatal(err)
	}
	return reader
}
