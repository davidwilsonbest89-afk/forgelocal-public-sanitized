// Package fonts contains a documentary, redacted font-bundle contract.
// It does not inspect host fonts or include any font binary.
package fonts

import (
	"errors"
	"path/filepath"
	"strings"
)

const (
	InventoryVersion = "font-bundle-projection-v1"
	ObservationMode  = "DECLARED_METADATA_ONLY"
)

type Status string

const (
	StatusNotIncluded Status = "NOT_INCLUDED"
	StatusRejected    Status = "REJECTED"
	StatusDeclared    Status = "DECLARED"
)

type Entry struct {
	ID             string `json:"id"`
	Provenance     string `json:"provenance"`
	LicenseStatus  string `json:"license_status"`
	Redistribution string `json:"redistribution"`
	Status         Status `json:"status"`
}

type Inventory struct {
	Version         string  `json:"version"`
	ObservationMode string  `json:"observation_mode"`
	Entries         []Entry `json:"entries"`
}

var (
	ErrInvalidEntryID = errors.New("fonts: invalid entry id")
	ErrMissingLicense = errors.New("fonts: license review required")
)

// DeclaredInventory is intentionally empty of font files and host observations.
func DeclaredInventory() Inventory {
	return Inventory{
		Version:         InventoryVersion,
		ObservationMode: ObservationMode,
		Entries: []Entry{{
			ID:             "font-bundle",
			Provenance:     "NOT_SUPPLIED",
			LicenseStatus:  "PENDING_REVIEW",
			Redistribution: "BLOCKED",
			Status:         StatusNotIncluded,
		}},
	}
}

// ValidateEntry applies path and license guards to supplied documentary
// metadata. It never opens the referenced path.
func ValidateEntry(entry Entry) error {
	if entry.ID == "" || filepath.IsAbs(entry.ID) || strings.Contains(entry.ID, "..") || strings.ContainsAny(entry.ID, "/\\\x00") {
		return ErrInvalidEntryID
	}
	if entry.LicenseStatus == "" || entry.LicenseStatus == "UNKNOWN" {
		return ErrMissingLicense
	}
	if entry.Redistribution == "GRANTED" && entry.LicenseStatus != "APPROVED" {
		return ErrMissingLicense
	}
	return nil
}
