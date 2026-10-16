package testkit_test

import (
	"sync"
	"testing"
	"time"

	"sdd.dev/sdd/internal/testkit"
)

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
