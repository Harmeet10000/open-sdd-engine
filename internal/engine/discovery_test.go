package engine_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"sdd.dev/sdd/internal/engine"
	"sdd.dev/sdd/internal/testkit"
)

const (
	discoveryWorkspace = `formatVersion: 1
workspaceId: ws-1
revision: 11
`
	discoveryWorkApple = `schemaVersion: 1
workId: work-b
slug: apple
revision: 5
status: active
composition: quick-spec
`
	discoveryWorkZebra = `schemaVersion: 1
workId: work-a
slug: zebra
revision: 7
status: active
composition: bug-fix
`
)

func discoveryFiles(extra map[string]testkit.MemFile) map[string]testkit.MemFile {
	files := map[string]testkit.MemFile{
		"/repo/.sdd/workspace.yaml":             {Bytes: []byte(discoveryWorkspace)},
		"/repo/.sdd/work/apple/work.yaml":       {Bytes: []byte(discoveryWorkApple)},
		"/repo/.sdd/work/zebra/work.yaml":       {Bytes: []byte(discoveryWorkZebra)},
		"/repo/.sdd/work/apple/requirements.md": {Bytes: []byte("extra file is not a work dir\n")},
	}
	for name, file := range extra {
		files[name] = file
	}
	return files
}

func TestDiscoverLayoutFindsNearestMarker(t *testing.T) {
	t.Parallel()

	facts := testkit.NewMemFileFacts(discoveryFiles(nil))
	layout, err := engine.DiscoverLayout(context.Background(), facts, "/repo/.sdd/work/apple", "")
	if err != nil {
		t.Fatal(err)
	}
	if layout.Root() != "/repo" || layout.WorkspaceID() != "ws-1" || layout.FormatVersion() != 1 || layout.Revision() != 11 {
		t.Fatalf("layout identity is wrong: %#v", layout)
	}
	works := layout.Works()
	if len(works) != 2 || works[0].ID() != "work-a" || works[1].ID() != "work-b" {
		t.Fatalf("works are not ordered by ID: %v", works)
	}
	if works[1].Slug() != "apple" || works[1].Status() != "active" || works[1].Composition() != "quick-spec" || works[1].Revision() != 5 {
		t.Fatalf("work record lost fields: %#v", works[1])
	}
	works[0] = engine.WorkRecord{}
	if layout.Works()[0].ID() != "work-a" {
		t.Fatal("layout works accessor aliases stored records")
	}
}

func TestDiscoverLayoutFindsZeroWorkItems(t *testing.T) {
	t.Parallel()

	for name, files := range map[string]map[string]testkit.MemFile{
		"missing work dir": {
			"/repo/.sdd/workspace.yaml": {Bytes: []byte(discoveryWorkspace)},
		},
		"empty work dir": {
			"/repo/.sdd/workspace.yaml": {Bytes: []byte(discoveryWorkspace)},
			"/repo/.sdd/work":           {Kind: engine.FileKindDirectory},
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			facts := testkit.NewMemFileFacts(files)
			layout, err := engine.DiscoverLayout(context.Background(), facts, "/repo", "")
			if err != nil {
				t.Fatal(err)
			}
			if len(layout.Works()) != 0 {
				t.Fatalf("expected zero works: %v", layout.Works())
			}
			if _, err := engine.SelectWork(layout, ""); err != nil {
				t.Fatalf("empty selection error = %v", err)
			}
		})
	}
}

func TestDiscoverLayoutRejectsInvalidLayouts(t *testing.T) {
	t.Parallel()

	workspaceCases := map[string]string{
		"malformed yaml":       "formatVersion: [1\n",
		"empty manifest":       "",
		"wrong version":        "formatVersion: 2\nworkspaceId: ws-1\nrevision: 1\n",
		"zero version":         "formatVersion: 0\nworkspaceId: ws-1\nrevision: 1\n",
		"missing workspace ID": "formatVersion: 1\nrevision: 1\n",
		"unknown field":        "formatVersion: 1\nworkspaceId: ws-1\nrevision: 1\nowner: alice\n",
		"duplicate key":        "formatVersion: 1\nformatVersion: 1\nworkspaceId: ws-1\nrevision: 1\n",
		"negative revision":    "formatVersion: 1\nworkspaceId: ws-1\nrevision: -1\n",
		"trailing document":    "formatVersion: 1\nworkspaceId: ws-1\nrevision: 1\n---\nformatVersion: 1\n",
	}
	for name, manifest := range workspaceCases {
		t.Run("workspace/"+name, func(t *testing.T) {
			t.Parallel()
			files := discoveryFiles(map[string]testkit.MemFile{
				"/repo/.sdd/workspace.yaml": {Bytes: []byte(manifest)},
			})
			assertIncompatible(t, files)
		})
	}

	workCases := map[string]string{
		"duplicate work IDs":    "schemaVersion: 1\nworkId: work-b\nslug: zebra\nrevision: 1\nstatus: active\ncomposition: bug-fix\n",
		"wrong work version":    "schemaVersion: 0\nworkId: work-9\nslug: zebra\nrevision: 1\nstatus: active\ncomposition: bug-fix\n",
		"malformed work record": "schemaVersion: [1\n",
		"missing work status":   "schemaVersion: 1\nworkId: work-9\nslug: zebra\nrevision: 1\ncomposition: bug-fix\n",
		"slug mismatch":         "schemaVersion: 1\nworkId: work-9\nslug: other\nrevision: 1\nstatus: active\ncomposition: bug-fix\n",
		"unknown work field":    "schemaVersion: 1\nworkId: work-9\nslug: zebra\nrevision: 1\nstatus: active\ncomposition: bug-fix\nowner: alice\n",
	}
	for name, record := range workCases {
		t.Run("work/"+name, func(t *testing.T) {
			t.Parallel()
			files := discoveryFiles(map[string]testkit.MemFile{
				"/repo/.sdd/work/zebra/work.yaml": {Bytes: []byte(record)},
			})
			assertIncompatible(t, files)
		})
	}
}

func assertIncompatible(t *testing.T, files map[string]testkit.MemFile) {
	t.Helper()
	facts := testkit.NewMemFileFacts(files)
	_, err := engine.DiscoverLayout(context.Background(), facts, "/repo", "")
	if !errors.Is(err, engine.ErrWorkspaceIncompatible) {
		t.Fatalf("layout error = %v", err)
	}
	if got := engine.LayoutDiagnosticCode(err); got != "DIAG_WORKSPACE_VERSION_INCOMPATIBLE" {
		t.Fatalf("diagnostic code = %q", got)
	}
}

func TestDiscoverLayoutReportsMissingWorkspace(t *testing.T) {
	t.Parallel()

	facts := testkit.NewMemFileFacts(map[string]testkit.MemFile{
		"/elsewhere/file.txt": {Bytes: []byte("unrelated\n")},
	})
	_, err := engine.DiscoverLayout(context.Background(), facts, "/elsewhere", "")
	if !errors.Is(err, engine.ErrWorkspaceNotFound) {
		t.Fatalf("missing workspace error = %v", err)
	}
	if got := engine.LayoutDiagnosticCode(err); got != "DIAG_WORKSPACE_NOT_FOUND" {
		t.Fatalf("diagnostic code = %q", got)
	}
}

func TestDiscoverLayoutNeverSkipsAnInvalidMarker(t *testing.T) {
	t.Parallel()

	facts := testkit.NewMemFileFacts(map[string]testkit.MemFile{
		"/repo/.sdd/workspace.yaml":        {Bytes: []byte(discoveryWorkspace)},
		"/repo/deep/.sdd/workspace.yaml":   {Bytes: []byte("formatVersion: 99\nworkspaceId: ws-x\nrevision: 1\n")},
		"/repo/deep/.sdd/work/a/work.yaml": {Bytes: []byte(discoveryWorkApple)},
	})
	_, err := engine.DiscoverLayout(context.Background(), facts, "/repo/deep", "")
	if !errors.Is(err, engine.ErrWorkspaceIncompatible) {
		t.Fatalf("invalid nearest marker was skipped or misreported: %v", err)
	}
}

func TestDiscoverLayoutHonorsExplicitRootWithoutFallback(t *testing.T) {
	t.Parallel()

	facts := testkit.NewMemFileFacts(discoveryFiles(map[string]testkit.MemFile{
		"/repo/child/note.txt": {Bytes: []byte("start here\n")},
	}))

	layout, err := engine.DiscoverLayout(context.Background(), facts, "/repo/child", "/repo")
	if err != nil {
		t.Fatal(err)
	}
	if layout.Root() != "/repo" || len(layout.Works()) != 2 {
		t.Fatalf("explicit root layout is wrong: %#v", layout)
	}

	_, err = engine.DiscoverLayout(context.Background(), facts, "/repo/child", "/repo/child")
	if !errors.Is(err, engine.ErrInvalidRoot) {
		t.Fatalf("invalid explicit root error = %v", err)
	}
	if got := engine.LayoutDiagnosticCode(err); got != "DIAG_WORKSPACE_NOT_FOUND" {
		t.Fatalf("diagnostic code = %q", got)
	}
}

func TestSelectWorkFollowsExplicitRootRules(t *testing.T) {
	t.Parallel()

	facts := testkit.NewMemFileFacts(discoveryFiles(nil))
	layout, err := engine.DiscoverLayout(context.Background(), facts, "/repo", "")
	if err != nil {
		t.Fatal(err)
	}

	explicit, err := engine.SelectWork(layout, "work-b")
	if err != nil {
		t.Fatal(err)
	}
	if explicit.Slug() != "apple" {
		t.Fatalf("explicit selection returned %q", explicit.Slug())
	}

	if _, slugErr := engine.SelectWork(layout, "apple"); !errors.Is(slugErr, engine.ErrWorkNotFound) {
		t.Fatalf("slug selection error = %v", slugErr)
	} else if got := engine.LayoutDiagnosticCode(slugErr); got != "DIAG_WORK_ITEM_NOT_FOUND" {
		t.Fatalf("diagnostic code = %q", got)
	}

	if _, missingErr := engine.SelectWork(layout, "work-missing"); !errors.Is(missingErr, engine.ErrWorkNotFound) {
		t.Fatalf("missing selection error = %v", missingErr)
	}

	_, err = engine.SelectWork(layout, "")
	if !errors.Is(err, engine.ErrSelectionAmbiguous) {
		t.Fatalf("ambiguous selection error = %v", err)
	}
	if got := engine.LayoutDiagnosticCode(err); got != "DIAG_WORKSPACE_SELECTION_AMBIGUOUS" {
		t.Fatalf("diagnostic code = %q", got)
	}
	candidates := engine.AmbiguousCandidates(err)
	if len(candidates) != 2 || candidates[0] != "work-a" || candidates[1] != "work-b" {
		t.Fatalf("ambiguous candidates = %v", candidates)
	}
	if got := engine.AmbiguousCandidates(errors.New("other")); got != nil {
		t.Fatalf("non-ambiguous error produced candidates: %v", got)
	}

	single, err := engine.DiscoverLayout(context.Background(), testkit.NewMemFileFacts(map[string]testkit.MemFile{
		"/repo/.sdd/workspace.yaml":       {Bytes: []byte(discoveryWorkspace)},
		"/repo/.sdd/work/apple/work.yaml": {Bytes: []byte(discoveryWorkApple)},
	}), "/repo", "")
	if err != nil {
		t.Fatal(err)
	}
	implied, err := engine.SelectWork(single, "")
	if err != nil {
		t.Fatal(err)
	}
	if implied.ID() != "work-b" {
		t.Fatalf("single implicit selection returned %q", implied.ID())
	}
}

func TestDiscoverLayoutPropagatesCancellationAndRejectsBadInput(t *testing.T) {
	t.Parallel()

	facts := testkit.NewMemFileFacts(discoveryFiles(nil))
	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := engine.DiscoverLayout(cancelled, facts, "/repo", ""); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled discovery error = %v", err)
	}

	if _, err := engine.DiscoverLayout(context.Background(), nil, "/repo", ""); !errors.Is(err, engine.ErrInvalidProtocolValue) {
		t.Fatalf("nil facts error = %v", err)
	}
	if _, err := engine.DiscoverLayout(context.Background(), facts, "", ""); !errors.Is(err, engine.ErrInvalidProtocolValue) {
		t.Fatalf("empty start dir error = %v", err)
	}
}

func TestDiscoverLayoutLeavesNoMutation(t *testing.T) {
	t.Parallel()

	files := discoveryFiles(nil)
	facts := testkit.NewMemFileFacts(files)
	if _, err := engine.DiscoverLayout(context.Background(), facts, "/repo/.sdd/work/apple", ""); err != nil {
		t.Fatal(err)
	}
	for name, file := range files {
		content, err := facts.ReadFile(context.Background(), name)
		if err != nil {
			t.Fatal(err)
		}
		if string(content) != string(file.Bytes) {
			t.Fatalf("discovery mutated %s", name)
		}
	}
}

func TestDiscoverLayoutReadsNativeDirectory(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	write := func(name, content string) {
		t.Helper()
		path := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write(".sdd/workspace.yaml", discoveryWorkspace)
	write(".sdd/work/apple/work.yaml", discoveryWorkApple)

	layout, err := engine.DiscoverLayout(context.Background(), engine.OSFileFacts{}, filepath.Join(root, ".sdd", "work"), "")
	if err != nil {
		t.Fatal(err)
	}
	if layout.Root() != root || len(layout.Works()) != 1 || layout.Works()[0].ID() != "work-b" {
		t.Fatalf("native discovery is wrong: %#v", layout)
	}
}

func TestLayoutDiagnosticCodeDefaultsToUsage(t *testing.T) {
	t.Parallel()

	if got := engine.LayoutDiagnosticCode(errors.New("boom")); got != "DIAG_USAGE_ERROR" {
		t.Fatalf("default code = %q", got)
	}
}
