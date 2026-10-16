package engine_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"sdd.dev/sdd/internal/engine"
	"sdd.dev/sdd/internal/testkit"
	"sdd.dev/sdd/internal/workspace"
)

func TestClockPortsAreControllable(t *testing.T) {
	t.Parallel()

	fixed := testkit.NewFixedClock(time.Date(2026, 10, 16, 12, 0, 0, 0, time.FixedZone("IST", 5*3600+1800)))
	if got := fixed.Now(); got.Location() != time.UTC {
		t.Fatalf("fixed clock did not normalize to UTC: %v", got)
	}
	before := fixed.Now()
	fixed.Advance(90 * time.Minute)
	if got, want := fixed.Now(), before.Add(90*time.Minute); !got.Equal(want) {
		t.Fatalf("fixed clock advance: got %v want %v", got, want)
	}

	var asPort engine.Clock = fixed
	if !asPort.Now().Equal(fixed.Now()) {
		t.Fatal("fixed clock does not satisfy the engine Clock port")
	}

	system := engine.SystemClock{}
	var _ engine.Clock = system
	if got := system.Now(); got.Location() != time.UTC || time.Since(got) > time.Minute {
		t.Fatalf("system clock is not a current UTC instant: %v", got)
	}
}

func TestIDSourcesMintStableNonIntentIdentities(t *testing.T) {
	t.Parallel()

	sequenced := testkit.NewSequenceIDSource("work")
	first, err := sequenced.NewID()
	if err != nil {
		t.Fatal(err)
	}
	second, err := sequenced.NewID()
	if err != nil {
		t.Fatal(err)
	}
	if first != "work-0" || second != "work-1" {
		t.Fatalf("sequence identities are not deterministic: %q %q", first, second)
	}

	injected := errors.New("namespace exhausted")
	sequenced.InjectFault(injected)
	if _, err := sequenced.NewID(); !errors.Is(err, injected) {
		t.Fatalf("injected identity fault did not propagate: %v", err)
	}
	sequenced.ResetFault()
	if got, err := sequenced.NewID(); err != nil || got != "work-2" {
		t.Fatalf("fault consumed a sequence value: got %q err %v", got, err)
	}

	random := engine.RandomIDSource{}
	seen := map[string]struct{}{}
	for range 64 {
		identity, err := random.NewID()
		if err != nil {
			t.Fatal(err)
		}
		if identity == "" {
			t.Fatal("random identity source minted an empty ID")
		}
		if _, duplicate := seen[identity]; duplicate {
			t.Fatalf("random identity source repeated %q", identity)
		}
		seen[identity] = struct{}{}
	}
}

func TestSnapshotterPropagatesCallerContext(t *testing.T) {
	t.Parallel()

	var _ engine.Snapshotter = (*workspace.Reader)(nil)
	fake := testkit.NewFixedSnapshotter(workspace.Snapshot{SnapshotID: "snap-1"})

	snapshot, err := fake.Snapshot(context.Background(), workspace.SnapshotRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if snapshot.SnapshotID != "snap-1" {
		t.Fatalf("snapshotter did not return the canned snapshot: %#v", snapshot)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := fake.Snapshot(ctx, workspace.SnapshotRequest{}); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled context did not propagate through the port: %v", err)
	}

	injected := errors.New("generation moved")
	fake.InjectFault(injected)
	if _, err := fake.Snapshot(context.Background(), workspace.SnapshotRequest{}); !errors.Is(err, injected) {
		t.Fatalf("injected snapshot fault did not propagate: %v", err)
	}
}

func TestPolicySnapshotIsImmutableAndFaultable(t *testing.T) {
	t.Parallel()

	floor := []byte("floor: strict\n")
	snapshot, err := engine.NewPolicySnapshot(engine.PolicySnapshotInput{
		Source: "ci-policy/v3",
		Floor:  floor,
	})
	if err != nil {
		t.Fatal(err)
	}
	floor[0] = 'X'
	if got := string(snapshot.Floor()); got != "floor: strict\n" {
		t.Fatalf("policy snapshot aliases caller bytes: %q", got)
	}
	got := snapshot.Floor()
	got[0] = 'X'
	if string(snapshot.Floor())[0] == 'X' {
		t.Fatal("policy floor accessor aliases stored bytes")
	}
	if snapshot.Source() != "ci-policy/v3" {
		t.Fatal("policy snapshot accessors lost input")
	}
	if snapshot.Digest() != engine.BlobDigest("sdd.policy-floor/1", []byte("floor: strict\n")) {
		t.Fatalf("policy digest does not bind the observed bytes: %q", snapshot.Digest())
	}

	if _, emptyErr := engine.NewPolicySnapshot(engine.PolicySnapshotInput{}); !errors.Is(emptyErr, engine.ErrInvalidProtocolValue) {
		t.Fatalf("empty policy snapshot error = %v", emptyErr)
	}

	fake := testkit.NewFixedPolicySnapshotter(snapshot)
	observed, err := fake.PolicySnapshot(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if observed.Digest() != snapshot.Digest() {
		t.Fatalf("policy port did not return the canned snapshot: %#v", observed)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := fake.PolicySnapshot(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled context did not propagate through the policy port: %v", err)
	}
	injected := errors.New("policy unverifiable")
	fake.InjectFault(injected)
	if _, err := fake.PolicySnapshot(context.Background()); !errors.Is(err, injected) {
		t.Fatalf("injected policy fault did not propagate: %v", err)
	}
}

func TestMemFileFactsModelBehaviorAndFaults(t *testing.T) {
	t.Parallel()

	content := []byte("workId: work-1\n")
	fake := testkit.NewMemFileFacts(map[string]testkit.MemFile{
		"/repo/.sdd/work.yaml": {Bytes: content, Mode: 0o644, ModTime: time.Date(2026, 10, 16, 0, 0, 0, 0, time.UTC)},
		"/repo/.sdd":           {Kind: engine.FileKindDirectory, Mode: 0o755},
	})
	content[0] = 'X'

	fact, err := fake.Stat(context.Background(), "/repo/.sdd/work.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if fact.Kind() != engine.FileKindRegular || fact.Size() != 15 || fact.Mode() != 0o644 {
		t.Fatalf("stat fact is wrong: %#v", fact)
	}
	if fact.ModTime().Location() != time.UTC {
		t.Fatalf("stat timestamp is not UTC: %v", fact.ModTime())
	}
	directory, err := fake.Stat(context.Background(), "/repo/.sdd")
	if err != nil {
		t.Fatal(err)
	}
	if directory.Kind() != engine.FileKindDirectory {
		t.Fatalf("directory kind is wrong: %#v", directory)
	}

	read, err := fake.ReadFile(context.Background(), "/repo/.sdd/work.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if string(read) != "workId: work-1\n" {
		t.Fatalf("read bytes alias constructor input: %q", read)
	}
	read[0] = 'X'
	again, err := fake.ReadFile(context.Background(), "/repo/.sdd/work.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if string(again) != "workId: work-1\n" {
		t.Fatal("read bytes alias stored content")
	}

	if _, err := fake.Stat(context.Background(), "/repo/.sdd/missing.yaml"); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("missing stat error = %v", err)
	}
	if _, err := fake.ReadFile(context.Background(), "/repo/.sdd/missing.yaml"); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("missing read error = %v", err)
	}
	if _, err := fake.Stat(context.Background(), ""); !errors.Is(err, engine.ErrInvalidProtocolValue) {
		t.Fatalf("empty path stat error = %v", err)
	}
	if _, err := fake.ReadFile(context.Background(), ""); !errors.Is(err, engine.ErrInvalidProtocolValue) {
		t.Fatalf("empty path read error = %v", err)
	}

	injected := errors.New("disk fault")
	fake.InjectFault("/repo/.sdd/work.yaml", injected)
	if _, err := fake.Stat(context.Background(), "/repo/.sdd/work.yaml"); !errors.Is(err, injected) {
		t.Fatalf("injected stat fault did not propagate: %v", err)
	}
	if _, err := fake.ReadFile(context.Background(), "/repo/.sdd/work.yaml"); !errors.Is(err, injected) {
		t.Fatalf("injected read fault did not propagate: %v", err)
	}
}

func TestBlockedReadHonorsCancellation(t *testing.T) {
	t.Parallel()

	fake := testkit.NewMemFileFacts(map[string]testkit.MemFile{
		"/repo/slow.yaml": {Bytes: []byte("slow\n")},
	})
	fake.BlockRead("/repo/slow.yaml")

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		_, err := fake.ReadFile(ctx, "/repo/slow.yaml")
		done <- err
	}()
	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("blocked read did not honor cancellation: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("blocked read ignored caller cancellation")
	}

	type readResult struct {
		content []byte
		err     error
	}
	released := make(chan readResult, 1)
	go func() {
		content, err := fake.ReadFile(context.Background(), "/repo/slow.yaml")
		released <- readResult{content: content, err: err}
	}()
	fake.ReleaseRead("/repo/slow.yaml")
	select {
	case result := <-released:
		if result.err != nil {
			t.Fatal(result.err)
		}
		if string(result.content) != "slow\n" {
			t.Fatalf("released read returned %q", result.content)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("released read did not return after ReleaseRead")
	}
}

func TestOSFileFactsReadNativeDirectory(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "work.yaml"), []byte("workId: work-1\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(root, "work"), 0o755); err != nil {
		t.Fatal(err)
	}

	native := engine.OSFileFacts{}
	var _ engine.FileFacts = native
	ctx := context.Background()

	fact, err := native.Stat(ctx, filepath.Join(root, "work.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if fact.Kind() != engine.FileKindRegular || fact.Size() != 15 {
		t.Fatalf("native stat fact is wrong: %#v", fact)
	}
	content, err := native.ReadFile(ctx, filepath.Join(root, "work.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "workId: work-1\n" {
		t.Fatalf("native read returned %q", content)
	}
	directory, err := native.Stat(ctx, filepath.Join(root, "work"))
	if err != nil {
		t.Fatal(err)
	}
	if directory.Kind() != engine.FileKindDirectory {
		t.Fatalf("native directory kind is wrong: %#v", directory)
	}

	if _, err := native.Stat(ctx, filepath.Join(root, "missing.yaml")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("missing native stat error = %v", err)
	}
	if _, err := native.ReadFile(ctx, filepath.Join(root, "missing.yaml")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("missing native read error = %v", err)
	}
	if _, err := native.Stat(ctx, ""); !errors.Is(err, engine.ErrInvalidProtocolValue) {
		t.Fatalf("empty native path error = %v", err)
	}

	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	if _, err := native.Stat(cancelled, filepath.Join(root, "work.yaml")); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled stat did not short-circuit: %v", err)
	}
	if _, err := native.ReadFile(cancelled, filepath.Join(root, "work.yaml")); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled read did not short-circuit: %v", err)
	}
}

func TestFixedClockZeroValueIsEpoch(t *testing.T) {
	t.Parallel()

	var clock testkit.FixedClock
	if got := clock.Now(); !got.Equal(time.Unix(0, 0).UTC()) {
		t.Fatalf("zero clock = %v, documented Unix epoch", got)
	}
}

func TestMemFileFactsRejectDirectoryReads(t *testing.T) {
	t.Parallel()

	fake := testkit.NewMemFileFacts(map[string]testkit.MemFile{
		"/repo/.sdd": {Kind: engine.FileKindDirectory},
	})
	if _, err := fake.ReadFile(context.Background(), "/repo/.sdd"); err == nil {
		t.Fatal("directory read succeeded, native adapter fails it")
	}
}

func TestMemFileFactsListsRootAndTrailingSlashes(t *testing.T) {
	t.Parallel()

	fake := testkit.NewMemFileFacts(map[string]testkit.MemFile{
		"/repo/.sdd/workspace.yaml": {Bytes: []byte("w\n")},
	})
	for _, path := range []string{"/", "/repo/", "/repo/.sdd/"} {
		entries, err := fake.ListDir(context.Background(), path)
		if err != nil {
			t.Fatalf("ListDir(%q) error = %v", path, err)
		}
		if len(entries) == 0 {
			t.Fatalf("ListDir(%q) missed recorded children", path)
		}
	}
	root, err := fake.ListDir(context.Background(), "/")
	if err != nil {
		t.Fatal(err)
	}
	if len(root) != 1 || root[0].Name() != "repo" || root[0].Kind() != engine.FileKindDirectory {
		t.Fatalf("root listing = %v", root)
	}
}

func TestFileFactRejectsUnknownKind(t *testing.T) {
	t.Parallel()

	if _, err := engine.NewFileFact(engine.FileFactInput{Kind: "symlink"}); !errors.Is(err, engine.ErrInvalidProtocolValue) {
		t.Fatalf("unknown file kind error = %v", err)
	}
}

func TestDirEntryRejectsInvalidNames(t *testing.T) {
	t.Parallel()

	for _, name := range []string{"", ".", "..", "a/b", `a\b`, "work.yaml/"} {
		if _, err := engine.NewDirEntry(engine.DirEntryInput{Name: name, Kind: engine.FileKindRegular}); !errors.Is(err, engine.ErrInvalidProtocolValue) {
			t.Fatalf("invalid entry name %q error = %v", name, err)
		}
	}
	entry, err := engine.NewDirEntry(engine.DirEntryInput{Name: "work.yaml", Kind: engine.FileKindRegular})
	if err != nil {
		t.Fatal(err)
	}
	if entry.Name() != "work.yaml" || entry.Kind() != engine.FileKindRegular {
		t.Fatalf("entry accessors lost input: %#v", entry)
	}
}

func TestListDirIsSortedAndFaultable(t *testing.T) {
	t.Parallel()

	fake := testkit.NewMemFileFacts(map[string]testkit.MemFile{
		"/repo/.sdd/work/zebra/work.yaml": {Bytes: []byte("z\n")},
		"/repo/.sdd/work/apple/work.yaml": {Bytes: []byte("a\n")},
		"/repo/.sdd/workspace.yaml":       {Bytes: []byte("w\n")},
		"/repo/.sdd/work":                 {Kind: engine.FileKindDirectory},
	})

	entries, err := fake.ListDir(context.Background(), "/repo/.sdd/work")
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 || entries[0].Name() != "apple" || entries[1].Name() != "zebra" {
		t.Fatalf("children are not sorted: %v", entries)
	}
	for _, entry := range entries {
		if entry.Kind() != engine.FileKindDirectory {
			t.Fatalf("nested work dir reported as %q", entry.Kind())
		}
	}

	top, err := fake.ListDir(context.Background(), "/repo/.sdd")
	if err != nil {
		t.Fatal(err)
	}
	if len(top) != 2 || top[0].Name() != "work" || top[1].Name() != "workspace.yaml" {
		t.Fatalf("top-level children are wrong: %v", top)
	}

	if _, err := fake.ListDir(context.Background(), "/repo/.sdd/workspace.yaml"); !errors.Is(err, engine.ErrDependencyFault) {
		t.Fatalf("file list error = %v", err)
	}
	if _, err := fake.ListDir(context.Background(), "/repo/absent"); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("missing list error = %v", err)
	}
	if _, err := fake.ListDir(context.Background(), ""); !errors.Is(err, engine.ErrInvalidProtocolValue) {
		t.Fatalf("empty list path error = %v", err)
	}
	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := fake.ListDir(cancelled, "/repo/.sdd"); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled list did not short-circuit: %v", err)
	}
	injected := errors.New("readdir fault")
	fake.InjectFault("/repo/.sdd", injected)
	if _, err := fake.ListDir(context.Background(), "/repo/.sdd"); !errors.Is(err, injected) {
		t.Fatalf("injected list fault did not propagate: %v", err)
	}
}

func TestOSListDirReadsNativeDirectory(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "b-dir"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "a-file"), []byte("a"), 0o644); err != nil {
		t.Fatal(err)
	}

	native := engine.OSFileFacts{}
	entries, err := native.ListDir(context.Background(), root)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 || entries[0].Name() != "a-file" || entries[0].Kind() != engine.FileKindRegular ||
		entries[1].Name() != "b-dir" || entries[1].Kind() != engine.FileKindDirectory {
		t.Fatalf("native children are wrong: %v", entries)
	}
	if _, err := native.ListDir(context.Background(), filepath.Join(root, "absent")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("missing native list error = %v", err)
	}
	if _, err := native.ListDir(context.Background(), filepath.Join(root, "a-file")); !errors.Is(err, engine.ErrDependencyFault) {
		t.Fatalf("file native list error = %v", err)
	}
}
