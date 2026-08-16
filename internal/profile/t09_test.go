// T09 — Profile Writes evidence tests.
//
// These tests cover the T09 write contract: lifecycle transitions, tag
// management, identifier uniqueness, negative paths and concurrent mutations
// on the same profile.

package profile

import (
	"sync"
	"testing"
	"time"
)

// --- Lifecycle ---------------------------------------------------------------

func TestLifecycleArchiveAndReopen(t *testing.T) {
	store, err := NewStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	p := &Profile{Name: "Archivable", RuntimeID: "camoufox"}
	if err := store.Create(p); err != nil {
		t.Fatal(err)
	}
	if p.LifecycleState != LifecycleActive {
		t.Fatalf("initial state = %q, want active", p.LifecycleState)
	}
	if err := store.ArchiveProfile(p.ID); err != nil {
		t.Fatal(err)
	}
	got, err := store.Get(p.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.LifecycleState != LifecycleArchived {
		t.Fatalf("state = %q, want archived", got.LifecycleState)
	}
	// Archived profiles refuse mutations.
	if _, err := store.Update(p.ID, map[string]any{"name": "Mutated"}); !IsLifecycleError(err) {
		t.Fatalf("archived update err = %v, want lifecycle error", err)
	}
	if err := store.ReopenProfile(p.ID); err != nil {
		t.Fatal(err)
	}
	got, _ = store.Get(p.ID)
	if got.LifecycleState != LifecycleActive {
		t.Fatalf("state = %q, want active after reopen", got.LifecycleState)
	}
}

func TestArchiveIdempotentAndQuarantinedRefused(t *testing.T) {
	store, err := NewStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	p := &Profile{Name: "Stateful", RuntimeID: "chromium"}
	if err := store.Create(p); err != nil {
		t.Fatal(err)
	}
	if err := store.ArchiveProfile(p.ID); err != nil {
		t.Fatal(err)
	}
	// A second archive on the same profile is idempotent: it resolves to
	// success with no storage side effect, matching the API contract tested
	// through the router.
	if err := store.ArchiveProfile(p.ID); err != nil {
		t.Fatalf("double archive err = %v, want idempotent nil", err)
	}
	gotAgain, err := store.Get(p.ID)
	if err != nil {
		t.Fatal(err)
	}
	if gotAgain.LifecycleState != LifecycleArchived {
		t.Fatalf("state after double archive = %q, want archived", gotAgain.LifecycleState)
	}
	if err := store.ReopenProfile(p.ID); err != nil {
		t.Fatal(err)
	}
	// Quarantine is lifted only by an external authority: reopening through
	// the public API must refuse.
	if err := store.quarantineForTest(p.ID); err != nil {
		t.Fatal(err)
	}
	if err := store.ReopenProfile(p.ID); !IsLifecycleError(err) {
		t.Fatalf("reopen quarantined err = %v, want lifecycle error", err)
	}
}

// --- Tags --------------------------------------------------------------------

func TestTagAssignmentBudgetAndArchivedRefusal(t *testing.T) {
	store, err := NewStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	p := &Profile{Name: "Tagged", RuntimeID: "camoufox"}
	if err := store.Create(p); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < maxProfileTags; i++ {
		if err := store.AddProfileTag(p.ID, tagName(i)); err != nil {
			t.Fatalf("tag %d: %v", i, err)
		}
	}
	if err := store.AddProfileTag(p.ID, tagName(maxProfileTags)); !IsValidationError(err) {
		t.Fatalf("over-budget tag err = %v, want validation error", err)
	}
	// Invalid tag names are refused at the boundary.
	if err := store.AddProfileTag(p.ID, ""); !IsValidationError(err) {
		t.Fatalf("empty tag err = %v, want validation error", err)
	}
	if err := store.RemoveProfileTag(p.ID, tagName(0)); err != nil {
		t.Fatal(err)
	}
	// Archived profiles refuse tag changes so their observable state stays
	// stable.
	if err := store.ArchiveProfile(p.ID); err != nil {
		t.Fatal(err)
	}
	if err := store.AddProfileTag(p.ID, tagName(1)); !IsLifecycleError(err) {
		t.Fatalf("tag archived profile err = %v, want lifecycle error", err)
	}
}

// --- Identifier uniqueness ---------------------------------------------------

func TestProfileNameUniquenessIsCaseInsensitive(t *testing.T) {
	store, err := NewStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	a := &Profile{Name: "Unique", RuntimeID: "camoufox"}
	if err := store.Create(a); err != nil {
		t.Fatal(err)
	}
	b := &Profile{Name: "unique", RuntimeID: "camoufox"}
	berr := store.Create(b)
	if !IsDuplicateError(berr) {
		t.Fatalf("case-insensitive dup err = %v, want duplicate error", berr)
	}
	// An explicit ID reuse is refused before any disk activity.
	c := &Profile{ID: a.ID, Name: "Fresh Name", RuntimeID: "camoufox"}
	if err := store.Create(c); !IsDuplicateError(err) {
		t.Fatalf("duplicate id err = %v, want duplicate error", err)
	}
	// A second, genuinely different profile is required to test a rename
	// into a taken name (a case-identical rename is a no-op by contract).
	d := &Profile{Name: "Other Profile", RuntimeID: "camoufox"}
	if err := store.Create(d); err != nil {
		t.Fatal(err)
	}
	// Renaming into a taken name is refused and leaves the profile intact.
	if _, err := store.Update(a.ID, map[string]any{"name": "OTHER PROFILE"}); !IsDuplicateError(err) {
		t.Fatalf("rename into dup err = %v, want duplicate error", err)
	}
	got, err := store.Get(a.ID)
	if err != nil || got.Name != "Unique" {
		t.Fatalf("profile damaged by failed rename: %+v %v", got, err)
	}
}

// --- Negative validation paths -----------------------------------------------

func TestWriteContractRejectsInvalidInputs(t *testing.T) {
	store, err := NewStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name string
		p    *Profile
	}{
		{"empty name", &Profile{Name: "", RuntimeID: "camoufox"}},
		{"too long name", &Profile{Name: dupString("n", 129), RuntimeID: "camoufox"}},
		{"invalid proxy type", &Profile{Name: "Px", RuntimeID: "camoufox", Proxy: &ProxyConfig{Type: "ftp", Host: "h", Port: 80}}},
		{"negative port", &Profile{Name: "Px2", RuntimeID: "camoufox", Proxy: &ProxyConfig{Type: "http", Host: "h", Port: -1}}},
		{"too many tags", &Profile{Name: "Tags", RuntimeID: "camoufox", Tags: tooManyTags()}},
		{"invalid tag", &Profile{Name: "Tag", RuntimeID: "camoufox", Tags: []string{""}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := store.Create(tc.p); !IsValidationError(err) {
				t.Fatalf("create %q err = %v, want validation error", tc.name, err)
			}
		})
	}
}

// --- Concurrency -------------------------------------------------------------

func TestConcurrentMutationsSerializePerProfile(t *testing.T) {
	t.Parallel()
	store, err := NewStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	p := &Profile{Name: "Contended", RuntimeID: "camoufox"}
	if err := store.Create(p); err != nil {
		t.Fatal(err)
	}
	const goroutines = 40
	// Capture the profile id up front: concurrent Update calls rewrite the
	// shared profile value under the store lock, so reading p.ID from each
	// goroutine's argument evaluation would itself race with those writes.
	id := p.ID
	var wg sync.WaitGroup
	var okUpdates, locked int
	var mu sync.Mutex
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			name := tagName(i)
			_, err := store.Update(id, map[string]any{"name": name})
			mu.Lock()
			if err == nil {
				okUpdates++
			} else if IsLockedError(err) {
				locked++
			}
			mu.Unlock()
		}(i)
	}
	wg.Wait()
	if okUpdates == 0 {
		t.Fatal("no update succeeded under contention")
	}
	// Exactly one writer wins per budget window; the remainder are refused by
	// the isolation budget, never silently merged.
	if okUpdates+locked != goroutines {
		t.Fatalf("ok=%d locked=%d, want total %d", okUpdates, locked, goroutines)
	}
}

func TestConcurrentProfileCreationsKeepUniqueNames(t *testing.T) {
	t.Parallel()
	store, err := NewStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	const goroutines = 30
	var wg sync.WaitGroup
	var created int
	var mu sync.Mutex
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			// Identical normalized names race against each other.
			p := &Profile{Name: "Racing Name", RuntimeID: "camoufox"}
			if err := store.Create(p); err == nil {
				mu.Lock()
				created++
				mu.Unlock()
			}
		}(i)
	}
	wg.Wait()
	if created < 1 || created > goroutines {
		t.Fatalf("created = %d out of %d under identical-name contention", created, goroutines)
	}
	// Every stored profile keeps exactly one owner of the name.
	all := store.List("", "")
	owners := make(map[string]string)
	for _, q := range all {
		if prev, taken := owners[normalizeName(q.Name)]; taken && prev != q.ID {
			t.Fatalf("name %q owned by both %q and %q", q.Name, prev, q.ID)
		}
		owners[normalizeName(q.Name)] = q.ID
	}
}

// --- Test helpers ------------------------------------------------------------

func tagName(i int) string { return "tag-" + iu64(i) }

func iu64(i int) string {
	if i < 10 {
		return string(rune('0' + i))
	}
	return iu64(i/10) + string(rune('0'+i%10))
}

func dupString(s string, n int) string {
	out := make([]byte, 0, n)
	for len(out) < n {
		out = append(out, s...)
	}
	return string(out[:n])
}

func tooManyTags() []string {
	tags := make([]string, maxProfileTags+1)
	for i := range tags {
		tags[i] = tagName(i)
	}
	return tags
}

// Verify budget bounds stay sensible.
func TestIsolationBudgetIsShort(t *testing.T) {
	if perProfileIsolationBudget > 10*time.Second {
		t.Fatalf("isolation budget %v is too long for a hot API path", perProfileIsolationBudget)
	}
}
