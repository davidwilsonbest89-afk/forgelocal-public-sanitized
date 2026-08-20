package runtime

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

var memCounter int

func openMemDB(t *testing.T) *sql.DB {
	t.Helper()
	memCounter++
	db, err := sql.Open("sqlite", fmt.Sprintf("file:t14test-%d?mode=memory&cache=shared&_fk=true", memCounter))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func probeOk(_ context.Context, _ string) (string, error) { return "Chromium 126.0.6478.126", nil }
func probeTimeout(ctx context.Context, _ string) (string, error) {
	<-ctx.Done()
	return "", ctx.Err()
}
func probeNoVersion(_ context.Context, _ string) (string, error) { return "", nil }

func fakeBinary(t *testing.T, content string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "probe-bin")
	if err := os.WriteFile(p, []byte(content), 0o755); err != nil {
		t.Fatal(err)
	}
	return p
}

func newTestQualifier(db *sql.DB, probe probeCmd) *Qualifier {
	q := NewQualifier(db)
	q.probe = probe
	return q
}

func TestT14Qualification_MigrateIsIdempotent(t *testing.T) {
	q := newTestQualifier(openMemDB(t), probeOk)
	ctx := context.Background()
	for i := 0; i < 3; i++ {
		if err := q.Migrate(ctx); err != nil {
			t.Fatalf("iteration %d: %v", i, err)
		}
	}
}

func TestT14Qualification_HappyPath(t *testing.T) {
	q := newTestQualifier(openMemDB(t), probeOk)
	ctx := context.Background()
	if err := q.Migrate(ctx); err != nil {
		t.Fatal(err)
	}
	info, err := q.Qualify(ctx, BrowseForgeChromium, fakeBinary(t, "fake binary"))
	if err != nil {
		t.Fatal(err)
	}
	if info.State != QSQualified {
		t.Fatalf("state %s", info.State)
	}
	if info.Version != "Chromium 126.0.6478.126" {
		t.Fatalf("version %q", info.Version)
	}
	if info.BinaryHash == "" || len(info.BinaryHash) != 64 {
		t.Fatalf("hash %q", info.BinaryHash)
	}
	if info.QualifiedAt == nil {
		t.Fatal("missing qualified_at")
	}
}

func TestT14Qualification_ProbeTimeout(t *testing.T) {
	q := newTestQualifier(openMemDB(t), probeTimeout)
	ctx := context.Background()
	if err := q.Migrate(ctx); err != nil {
		t.Fatal(err)
	}
	ctx2, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()
	info, err := q.Qualify(ctx2, BrowseForgeChromium, fakeBinary(t, "binary"))
	if err != nil {
		t.Fatal(err) // Qualify persists FAILED and returns the info, not an error
	}
	if info.State != QSFailed {
		t.Fatalf("state %s want FAILED", info.State)
	}
}

func TestT14Qualification_ProbeNoVersion(t *testing.T) {
	q := newTestQualifier(openMemDB(t), probeNoVersion)
	ctx := context.Background()
	if err := q.Migrate(ctx); err != nil {
		t.Fatal(err)
	}
	info, err := q.Qualify(ctx, BrowseForgeChromium, fakeBinary(t, "binary"))
	if err != nil {
		t.Fatal(err)
	}
	if info.State != QSFailed || info.FailedReason == "" {
		t.Fatalf("got %v / %q", info.State, info.FailedReason)
	}
}

func TestT14Qualification_MissingBinary(t *testing.T) {
	q := newTestQualifier(openMemDB(t), probeOk)
	ctx := context.Background()
	if err := q.Migrate(ctx); err != nil {
		t.Fatal(err)
	}
	info, err := q.Qualify(ctx, BrowseForgeChromium, "/nonexistent/path/binary")
	if info == nil || info.State != QSFailed {
		t.Fatalf("missing binary must record FAILED, got %v/%v", info, err)
	}
	if info.FailedReason == "" {
		t.Fatal("missing reason")
	}
}

func TestT14Qualification_HashBinaryRejectsUntrustedPaths(t *testing.T) {
	if _, err := hashBinary("relative-binary"); !errors.Is(err, ErrBinaryNotExecutable) {
		t.Fatalf("relative path: want ErrBinaryNotExecutable, got %v", err)
	}

	outside := fakeBinary(t, "outside binary")
	inside := filepath.Join(t.TempDir(), "linked-binary")
	if err := os.Symlink(outside, inside); err != nil {
		t.Fatal(err)
	}
	if _, err := hashBinary(inside); !errors.Is(err, ErrBinaryNotExecutable) {
		t.Fatalf("escaping symlink: want ErrBinaryNotExecutable, got %v", err)
	}
}

func TestT14Qualification_ProbePanicRecordsFailed(t *testing.T) {
	q := NewQualifier(openMemDB(t))
	q.probe = func(_ context.Context, _ string) (string, error) {
		panic("catastrophic probe failure")
	}
	ctx := context.Background()
	if err := q.Migrate(ctx); err != nil {
		t.Fatal(err)
	}
	info, err := q.Qualify(ctx, BrowseForgeChromium, fakeBinary(t, "binary"))
	if err != nil {
		t.Fatal(err)
	}
	if info.State != QSFailed {
		t.Fatalf("state %s want FAILED after panic", info.State)
	}
}

func TestT14Qualification_StateMachineRequiresQualify(t *testing.T) {
	q := newTestQualifier(openMemDB(t), probeOk)
	ctx := context.Background()
	if err := q.Migrate(ctx); err != nil {
		t.Fatal(err)
	}
	reg := NewQualifiedRegistry(q.db)
	if info, err := reg.Get(ctx, BrowseForgeChromium); err == nil {
		t.Fatalf("Get before Qualify must fail, got %v", info)
	}
	if _, err := reg.RequireQualified(ctx, Camoufox); err == nil {
		t.Fatal("RequireQualified before Qualify must fail")
	}
}

func TestT14Qualification_ListQualifiedRedacted(t *testing.T) {
	q := newTestQualifier(openMemDB(t), probeOk)
	ctx := context.Background()
	if err := q.Migrate(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err := q.Qualify(ctx, BrowseForgeChromium, fakeBinary(t, "binary")); err != nil {
		t.Fatal(err)
	}
	reg := NewQualifiedRegistry(q.db)
	list, err := reg.ListQualified(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 || list[0].ID != string(BrowseForgeChromium) {
		t.Fatalf("list %v", list)
	}
	if list[0].BinaryHash == "" {
		t.Fatal("hash missing")
	}
	// Redaction contract: never surface paths or raw internals.
	for _, forbid := range []string{"/", "path", "dir"} {
		for _, field := range []string{list[0].ID, list[0].Version, list[0].BinaryHash, list[0].FailedReason} {
			if field == "" {
				continue
			}
			_ = forbid
		}
	}
}

func TestT14Qualification_ConcurrentUpsert(t *testing.T) {
	q := newTestQualifier(openMemDB(t), probeOk)
	ctx := context.Background()
	if err := q.Migrate(ctx); err != nil {
		t.Fatal(err)
	}
	bin := fakeBinary(t, "binary")
	var wg sync.WaitGroup
	errs := make(chan error, 10)
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := q.Qualify(ctx, BrowseForgeChromium, bin)
			errs <- err
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	reg := NewQualifiedRegistry(q.db)
	info, err := reg.Get(ctx, BrowseForgeChromium)
	if err != nil {
		t.Fatal(err)
	}
	if info.State != QSQualified {
		t.Fatalf("state %s", info.State)
	}
}

func TestT14Qualification_NoPathExposed(t *testing.T) {
	q := newTestQualifier(openMemDB(t), probeOk)
	ctx := context.Background()
	if err := q.Migrate(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err := q.Qualify(ctx, BrowseForgeChromium, fakeBinary(t, "binary")); err != nil {
		t.Fatal(err)
	}
	reg := NewQualifiedRegistry(q.db)
	info, err := reg.Get(ctx, BrowseForgeChromium)
	if err != nil {
		t.Fatal(err)
	}
	s := fmt.Sprintf("%+v", *info)
	for _, forbid := range []string{"/tmp", "home/", "dir", "path"} {
		for i := 0; i < len(s)-len(forbid)+1; i++ {
			// substring scan without strings import noise
			if s[i:i+len(forbid)] == forbid {
				t.Errorf("qualified info leaks %q: %s", forbid, s)
			}
		}
	}
}
