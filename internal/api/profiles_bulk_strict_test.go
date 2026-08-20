// T24 — Strict request-shape and target-budget tests.

package api

import (
	"net/http"
	"strings"
	"testing"
)

func TestT24BulkRejectsStrictJSONAndTargetBudgetBeforeMutation(t *testing.T) {
	r, _, store, _, _ := testBulkRouter(t)
	id := createBulkProfile(t, store, "T24 Strict")
	for _, body := range []string{
		`{"operation":"archive","profile_ids":["` + id + `"],"unexpected":true}`,
		`{"operation":"archive","profile_ids":["` + id + `"]}{}`,
	} {
		_, rec := callBulk(t, r, body)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("strict request accepted status=%d body=%s", rec.Code, rec.Body.String())
		}
		assertErrorCode(t, rec, "INVALID_BULK_REQUEST")
	}
	tooMany := make([]string, maxBulkProfileTargets+1)
	for i := range tooMany {
		tooMany[i] = `"id-` + strings.Repeat("x", i+1) + `"`
	}
	_, rec := callBulk(t, r, `{"operation":"archive","profile_ids":[`+strings.Join(tooMany, ",")+`]}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("over-limit bulk accepted status=%d body=%s", rec.Code, rec.Body.String())
	}
	assertErrorCode(t, rec, "INVALID_BULK_REQUEST")
	p, err := store.Get(id)
	if err != nil || p.LifecycleState != "active" {
		t.Fatalf("rejected request mutated profile=%+v err=%v", p, err)
	}
}
