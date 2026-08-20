package cookies

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func fixtureInput(value string) Input {
	return Input{Name: "theme", Value: value, Domain: "app.test", Path: "/", SameSite: "lax"}
}

func TestReplacePersistsOnlySyntheticDigest(t *testing.T) {
	dir := t.TempDir()
	store, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	fixtures, err := store.Replace("profile-a", []Input{fixtureInput("fixture:blue")})
	if err != nil {
		t.Fatal(err)
	}
	if len(fixtures) != 1 || fixtures[0].ValueDigest == "" {
		t.Fatalf("fixtures=%#v", fixtures)
	}
	data, err := os.ReadFile(filepath.Join(dir, "cookie-fixtures.json"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "fixture:blue") {
		t.Fatalf("fixture marker persisted: %s", data)
	}
	if strings.Contains(string(data), "\"value\"") {
		t.Fatalf("raw value field persisted: %s", data)
	}
	reopened, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	if got := reopened.Export("profile-a"); len(got) != 1 || got[0].ValueDigest != fixtures[0].ValueDigest {
		t.Fatalf("reopened=%#v", got)
	}
}

func TestReplaceRejectsRealOrInvalidFixturesAtomically(t *testing.T) {
	store, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	previous, err := store.Replace("profile-a", []Input{fixtureInput("fixture:initial")})
	if err != nil {
		t.Fatal(err)
	}
	invalid := []Input{fixtureInput("fixture:next"), {Name: "sid", Value: "real-cookie-value", Domain: "app.test", Path: "/"}}
	if _, err := store.Replace("profile-a", invalid); err == nil {
		t.Fatal("real value accepted")
	}
	got := store.Export("profile-a")
	if len(got) != 1 || got[0].ValueDigest != previous[0].ValueDigest {
		t.Fatalf("atomicity lost: %#v", got)
	}
	for _, input := range []Input{{Name: "bad", Value: "fixture:x", Domain: "example.com", Path: "/"}, {Name: "bad", Value: "fixture:x", Domain: "app.test", Path: "relative"}} {
		if _, err := store.Replace("profile-b", []Input{input}); err == nil {
			t.Fatalf("invalid fixture accepted: %#v", input)
		}
	}
}
