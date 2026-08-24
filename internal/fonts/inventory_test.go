package fonts

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestT35DeclaredInventoryIsDeterministicAndNonRedistributable(t *testing.T) {
	inventory := DeclaredInventory()
	if inventory.Version != InventoryVersion || inventory.ObservationMode != ObservationMode {
		t.Fatalf("unexpected inventory header: %#v", inventory)
	}
	if len(inventory.Entries) != 1 {
		t.Fatalf("want one declared entry, got %d", len(inventory.Entries))
	}
	entry := inventory.Entries[0]
	if entry.Status != StatusNotIncluded || entry.Provenance != "NOT_SUPPLIED" || entry.LicenseStatus != "PENDING_REVIEW" || entry.Redistribution != "BLOCKED" {
		t.Fatalf("inventory must remain closed and non-redistributable: %#v", entry)
	}
	body, err := json.Marshal(inventory)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(body), "/usr/share") || strings.Contains(string(body), "fontfile") {
		t.Fatalf("inventory leaked host/font path: %s", body)
	}
}

func TestT35ValidateEntryFailsClosed(t *testing.T) {
	for _, entry := range []Entry{
		{ID: "", LicenseStatus: "APPROVED"},
		{ID: "../font", LicenseStatus: "APPROVED"},
		{ID: "/tmp/font", LicenseStatus: "APPROVED"},
		{ID: "font", LicenseStatus: "UNKNOWN"},
		{ID: "font", LicenseStatus: "", Redistribution: "GRANTED"},
		{ID: "font", LicenseStatus: "PENDING_REVIEW", Redistribution: "GRANTED"},
	} {
		if err := ValidateEntry(entry); err == nil {
			t.Fatalf("entry %#v unexpectedly accepted", entry)
		}
	}
}

func TestT35ValidateApprovedEntry(t *testing.T) {
	entry := Entry{ID: "font-bundle", Provenance: "DECLARED_BUNDLE", LicenseStatus: "APPROVED", Redistribution: "GRANTED", Status: StatusDeclared}
	if err := ValidateEntry(entry); err != nil {
		t.Fatal(err)
	}
}
