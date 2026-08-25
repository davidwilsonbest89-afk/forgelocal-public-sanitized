package extensions

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func syntheticZIP(t *testing.T, files map[string]string) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for name, body := range files {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := w.Write([]byte(body)); err != nil {
			t.Fatal(err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func highRiskManifest() string {
	return `{"name":"Synthetic Extension","version":"1.0.0","manifest_version":3,"permissions":["cookies","storage"],"optional_permissions":["unknownPermission"],"host_permissions":["*://*/*"],"optional_host_permissions":["file:///*"],"content_scripts":[{"matches":["https://example.test/*"]}]}`
}

func openTestRepository(t *testing.T) *Repository {
	t.Helper()
	r, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = r.Close() })
	return r
}

func TestT28ImportPreservesPermissionsAndRequiresExactHighRiskAcknowledgement(t *testing.T) {
	r := openTestRepository(t)
	v, err := r.Import(context.Background(), bytes.NewReader(syntheticZIP(t, map[string]string{"manifest.json": highRiskManifest(), "payload.js": "not executed"})), "", "corr-import")
	if err != nil {
		t.Fatal(err)
	}
	if v.State != "imported" || v.RiskState != "HIGH_RISK" {
		t.Fatalf("unexpected import projection: %+v", v)
	}
	if len(v.Manifest.Permissions) != 2 || len(v.Manifest.OptionalPermissions) != 1 || len(v.Manifest.HostPermissions) != 1 || len(v.Manifest.OptionalHostPermissions) != 1 || len(v.Manifest.ContentScriptMatches) != 1 {
		t.Fatalf("manifest declarations were not preserved: %+v", v.Manifest)
	}
	if _, err := r.Approve(context.Background(), v.ID, nil, false, "corr-refused"); !errors.Is(err, ErrPermissionAck) {
		t.Fatalf("expected exact acknowledgement refusal, got %v", err)
	}
	ack := []string{"cookies", "storage", "unknownPermission", "*://*/*", "file:///*", "https://example.test/*"}
	if _, err := r.Approve(context.Background(), v.ID, ack, false, "corr-high-risk"); !errors.Is(err, ErrHighRiskAck) {
		t.Fatalf("expected high-risk acknowledgement refusal, got %v", err)
	}
	approved, err := r.Approve(context.Background(), v.ID, ack, true, "corr-approved")
	if err != nil {
		t.Fatal(err)
	}
	if approved.State != "approved" || approved.ApprovedAt == "" {
		t.Fatalf("approval not persisted: %+v", approved)
	}
}

func TestT28LifecycleAssignmentUpdateRollbackRevokeAndPersistence(t *testing.T) {
	r := openTestRepository(t)
	ctx := context.Background()
	v1, err := r.Import(ctx, bytes.NewReader(syntheticZIP(t, map[string]string{"manifest.json": `{"name":"One","version":"1"}`})), "", "corr-1")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.Approve(ctx, v1.ID, nil, false, "corr-1-approve"); err != nil {
		t.Fatal(err)
	}
	assigned, err := r.Assign(ctx, v1.ID, "profile-1", "corr-1-assign", func(context.Context, string) (bool, error) { return true, nil })
	if err != nil {
		t.Fatal(err)
	}
	if assigned.State != "ready" {
		t.Fatalf("unexpected assignment: %+v", assigned)
	}
	v2, err := r.Update(ctx, v1.SeriesID, bytes.NewReader(syntheticZIP(t, map[string]string{"manifest.json": `{"name":"One","version":"2"}`})), "corr-2")
	if err != nil {
		t.Fatal(err)
	}
	if v2.ID == v1.ID || v2.Number != 2 {
		t.Fatalf("version was not immutable/incremented: v1=%+v v2=%+v", v1, v2)
	}
	if _, err := r.Approve(ctx, v2.ID, nil, false, "corr-2-approve"); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Assign(ctx, v2.ID, "profile-1", "corr-2-assign", func(context.Context, string) (bool, error) { return true, nil }); err != nil {
		t.Fatal(err)
	}
	if err := r.Rollback(ctx, v1.SeriesID, v1.ID, "corr-rollback"); err != nil {
		t.Fatal(err)
	}
	series, err := r.GetSeries(ctx, v1.SeriesID)
	if err != nil {
		t.Fatal(err)
	}
	if series.ActiveVersionID != v1.ID {
		t.Fatalf("rollback did not repoint active version: %+v", series)
	}
	if err := r.Revoke(ctx, v1.ID, "corr-revoke"); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Assign(ctx, v1.ID, "profile-2", "corr-refused", func(context.Context, string) (bool, error) { return true, nil }); !errors.Is(err, ErrRevoked) {
		t.Fatalf("revoked version was assignable: %v", err)
	}
	before := len(series.Versions)
	if err := r.Close(); err != nil {
		t.Fatal(err)
	}
	r2, err := Open(filepath.Dir(r.baseDir))
	if err != nil {
		t.Fatal(err)
	}
	defer r2.Close()
	persisted, err := r2.GetSeries(ctx, v1.SeriesID)
	if err != nil {
		t.Fatal(err)
	}
	if len(persisted.Versions) != before {
		t.Fatalf("restart changed version history: before=%d after=%d", before, len(persisted.Versions))
	}
}

func TestT28RejectsUnsafeArchivesAndCompensatesDatabaseFailure(t *testing.T) {
	r := openTestRepository(t)
	ctx := context.Background()
	cases := []struct {
		name  string
		files map[string]string
		want  error
	}{
		{"missing manifest", map[string]string{"payload.js": "x"}, ErrManifestInvalid},
		{"zip slip", map[string]string{"manifest.json": `{}`, "../escape": "x"}, ErrInvalidArchive},
		{"invalid manifest", map[string]string{"manifest.json": `[]`}, ErrManifestInvalid},
		{"trailing json", map[string]string{"manifest.json": `{} {}`}, ErrManifestInvalid},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := r.Import(ctx, bytes.NewReader(syntheticZIP(t, tc.files)), "", "corr-refuse")
			if !errors.Is(err, tc.want) {
				t.Fatalf("expected %v, got %v", tc.want, err)
			}
		})
	}
	r.SetFailBeforeCommitForTest(func() error { return errors.New("injected commit failure") })
	_, err := r.Import(ctx, bytes.NewReader(syntheticZIP(t, map[string]string{"manifest.json": `{}`})), "", "corr-db-failure")
	if err == nil {
		t.Fatal("expected injected failure")
	}
	r.SetFailBeforeCommitForTest(nil)
	result, err := r.List(ctx, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if result.Total != 0 {
		t.Fatalf("failed transaction published a series: %+v", result)
	}
}

func TestT28ConcurrentAssignmentCannotDoubleCreateRelation(t *testing.T) {
	r := openTestRepository(t)
	ctx := context.Background()
	v, err := r.Import(ctx, bytes.NewReader(syntheticZIP(t, map[string]string{"manifest.json": `{}`})), "", "corr-concurrent")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.Approve(ctx, v.ID, nil, false, "corr-concurrent-approve"); err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	var mu sync.Mutex
	var successes, failures int
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, assignErr := r.Assign(ctx, v.ID, "profile-concurrent", "corr-concurrent-assign", func(context.Context, string) (bool, error) { return true, nil })
			mu.Lock()
			defer mu.Unlock()
			if assignErr == nil {
				successes++
			} else {
				failures++
			}
		}()
	}
	wg.Wait()
	if successes != 1 || failures != 1 {
		t.Fatalf("expected one assignment winner, successes=%d failures=%d", successes, failures)
	}
	series, err := r.GetSeries(ctx, v.SeriesID)
	if err != nil {
		t.Fatal(err)
	}
	if len(series.Assignments) != 1 {
		t.Fatalf("double assignment created: %+v", series.Assignments)
	}
}

func TestT28RejectsSymlinkArchiveEntry(t *testing.T) {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	manifest, err := zw.Create("manifest.json")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := manifest.Write([]byte(`{}`)); err != nil {
		t.Fatal(err)
	}
	h := &zip.FileHeader{Name: "link", Method: zip.Store}
	h.SetMode(os.ModeSymlink | 0777)
	link, err := zw.CreateHeader(h)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := link.Write([]byte("/tmp/not-followed")); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	r := openTestRepository(t)
	if _, err := r.Import(context.Background(), bytes.NewReader(buf.Bytes()), "", "corr-symlink"); !errors.Is(err, ErrInvalidArchive) {
		t.Fatalf("symlink entry was accepted: %v", err)
	}
}

func TestT28StoreUsesManagedObjectPath(t *testing.T) {
	r := openTestRepository(t)
	v, err := r.Import(context.Background(), bytes.NewReader(syntheticZIP(t, map[string]string{"manifest.json": `{}`})), "", "corr-path")
	if err != nil {
		t.Fatal(err)
	}
	var rel string
	if err := r.db.QueryRow(`SELECT blob_relpath FROM extension_versions WHERE id=?`, v.ID).Scan(&rel); err != nil {
		t.Fatal(err)
	}
	if filepath.IsAbs(rel) || filepath.Dir(rel) == "." {
		t.Fatalf("blob path is not managed/derived: %q", rel)
	}
	if _, err := os.Stat(filepath.Join(r.baseDir, rel)); err != nil {
		t.Fatalf("managed blob missing: %v", err)
	}
}

func TestT28RejectsCorruptAndOversizedArchives(t *testing.T) {
	r := openTestRepository(t)
	if _, err := r.Import(context.Background(), bytes.NewReader([]byte("not a zip")), "", "corr-corrupt"); !errors.Is(err, ErrInvalidArchive) {
		t.Fatalf("corrupt ZIP was not rejected: %v", err)
	}
	oversized := make([]byte, MaxZIPBytes+1)
	if _, err := r.Import(context.Background(), bytes.NewReader(oversized), "", "corr-oversized"); !errors.Is(err, ErrArchiveLimit) {
		t.Fatalf("oversized archive was not rejected: %v", err)
	}
}

func TestT28PreservesAllAuthorizedPermissionsAndIgnoresUpdateURL(t *testing.T) {
	r := openTestRepository(t)
	manifest := `{"name":"All permissions","version":"1","permissions":["cookies","webRequest","webRequestBlocking","debugger","nativeMessaging","management","proxy","downloads","clipboardRead"],"host_permissions":["<all_urls>","*://*/*","file:///*"],"update_url":"https://updates.example.invalid/manifest.xml"}`
	v, err := r.Import(context.Background(), bytes.NewReader(syntheticZIP(t, map[string]string{"manifest.json": manifest})), "", "corr-update-url")
	if err != nil {
		t.Fatal(err)
	}
	if v.RiskState != "HIGH_RISK" {
		t.Fatalf("authorized sensitive permissions were not marked high-risk: %+v", v)
	}
	if len(v.Manifest.Permissions) != 9 || len(v.Manifest.HostPermissions) != 3 {
		t.Fatalf("permissions were not preserved: %+v", v.Manifest)
	}
	for _, permission := range []string{"cookies", "webRequest", "webRequestBlocking", "debugger", "nativeMessaging", "management", "proxy", "downloads", "clipboardRead"} {
		if !contains(v.Manifest.Permissions, permission) {
			t.Fatalf("permission %q was lost", permission)
		}
	}
	for _, host := range []string{"<all_urls>", "*://*/*", "file:///*"} {
		if !contains(v.Manifest.HostPermissions, host) {
			t.Fatalf("host pattern %q was lost", host)
		}
	}
}

func TestT28PurgeRequiresSafeLifecycleState(t *testing.T) {
	r := openTestRepository(t)
	ctx := context.Background()
	v, err := r.Import(ctx, bytes.NewReader(syntheticZIP(t, map[string]string{"manifest.json": `{}`})), "", "corr-purge")
	if err != nil {
		t.Fatal(err)
	}
	if !errors.Is(r.Purge(ctx, v.ID, "corr-purge-imported"), ErrPurgeNotAllowed) {
		t.Fatal("imported version was purged")
	}
	if _, err := r.Approve(ctx, v.ID, nil, false, "corr-purge-approve"); err != nil {
		t.Fatal(err)
	}
	if !errors.Is(r.Purge(ctx, v.ID, "corr-purge-approved"), ErrPurgeNotAllowed) {
		t.Fatal("active approved version was purged")
	}
	if err := r.Revoke(ctx, v.ID, "corr-purge-revoke"); err != nil {
		t.Fatal(err)
	}
	if err := r.Purge(ctx, v.ID, "corr-purge-final"); err != nil {
		t.Fatalf("revoked unassigned version was not purgeable: %v", err)
	}
	var count int
	if err := r.DB().QueryRow(`SELECT COUNT(*) FROM extension_versions WHERE id=?`, v.ID).Scan(&count); err != nil || count != 0 {
		t.Fatalf("purge left version metadata: count=%d err=%v", count, err)
	}
}

func TestT28RejectsPackageModifiedAfterImport(t *testing.T) {
	r := openTestRepository(t)
	ctx := context.Background()
	v, err := r.Import(ctx, bytes.NewReader(syntheticZIP(t, map[string]string{"manifest.json": `{}`})), "", "corr-tamper")
	if err != nil {
		t.Fatal(err)
	}
	var rel string
	if err := r.DB().QueryRow(`SELECT blob_relpath FROM extension_versions WHERE id=?`, v.ID).Scan(&rel); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(r.baseDir, rel), []byte("tampered after import"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Approve(ctx, v.ID, nil, false, "corr-tamper-approve"); !errors.Is(err, ErrIntegrity) {
		t.Fatalf("tampered package was approved: %v", err)
	}
}

func contains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
