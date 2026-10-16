package testkit_test

import (
	"sync"
	"testing"
	"time"

	"sdd.dev/sdd/internal/testkit"
	"sdd.dev/sdd/internal/workspace"
)

func TestFixedSnapshotterIsolatesSnapshots(t *testing.T) {
	t.Parallel()

	bytes := []byte("workspaceId: ws-1\n")
	fake := testkit.NewFixedSnapshotter(workspace.Snapshot{
		SnapshotID: "snap-1",
		Paths: []workspace.PathObservation{
			{Path: ".sdd/workspace.yaml", Presence: workspace.Present, Kind: workspace.FileKindRegular, Bytes: bytes, Size: uint64(len(bytes))},
		},
	})
	bytes[0] = 'X'

	first, err := fake.Snapshot(t.Context(), workspace.SnapshotRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if string(first.Paths[0].Bytes) != "workspaceId: ws-1\n" {
		t.Fatal("snapshotter aliases constructor input")
	}
	first.Paths[0].Bytes[0] = 'X'
	second, err := fake.Snapshot(t.Context(), workspace.SnapshotRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if string(second.Paths[0].Bytes) != "workspaceId: ws-1\n" {
		t.Fatal("snapshotter aliases returned snapshots")
	}
}

func TestFakesAreSafeForConcurrentUse(t *testing.T) {
	t.Parallel()

	clock := testkit.NewFixedClock(time.Unix(0, 0))
	identities := testkit.NewSequenceIDSource("id")
	files := testkit.NewMemFileFacts(map[string]testkit.MemFile{
		"/repo/a.yaml": {Bytes: []byte("a\n")},
	})

	const workers = 16
	var group sync.WaitGroup
	for range workers {
		group.Add(1)
		go func() {
			defer group.Done()
			clock.Advance(time.Second)
			if _, err := identities.NewID(); err != nil {
				t.Error(err)
				return
			}
			if _, err := files.ReadFile(t.Context(), "/repo/a.yaml"); err != nil {
				t.Error(err)
			}
		}()
	}
	group.Wait()

	if got := clock.Now(); !got.Equal(time.Unix(workers, 0)) {
		t.Fatalf("concurrent clock advances lost updates: %v", got)
	}
	seen := map[string]struct{}{}
	for range workers {
		identity, err := identities.NewID()
		if err != nil {
			t.Fatal(err)
		}
		seen[identity] = struct{}{}
	}
	if len(seen) != workers {
		t.Fatalf("concurrent identity minting produced duplicates: %v", seen)
	}
}
