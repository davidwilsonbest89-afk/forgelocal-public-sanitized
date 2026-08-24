// Package geolocation contains synthetic-only QA helpers.
//
// It deliberately has no network, browser, profile, persistence, or runtime
// dependencies. It validates fixture coordinates and emits only redacted
// statuses; it never returns coordinates in a result.
package geolocation

import (
	"errors"
	"math"
	"sort"
)

const ObservationModeSyntheticOnly = "SYNTHETIC_FIXTURES_ONLY"

type Point struct {
	Latitude  float64
	Longitude float64
}

type Sample struct {
	Name  string
	Point Point
}

type Status string

const (
	StatusPass Status = "PASS"
	StatusFail Status = "FAIL"
)

type Result struct {
	Name            string `json:"name"`
	Status          Status `json:"status"`
	ObservationMode string `json:"observation_mode"`
	Reason          string `json:"reason,omitempty"`
}

var (
	ErrInvalidSampleName = errors.New("geolocation: invalid synthetic sample name")
	ErrInvalidCoordinate = errors.New("geolocation: invalid synthetic coordinate")
)

// ValidateSyntheticPoint validates a fixture point without classifying or
// exposing any real location. NaN and Inf are refused fail-closed.
func ValidateSyntheticPoint(p Point) error {
	if math.IsNaN(p.Latitude) || math.IsInf(p.Latitude, 0) || math.IsNaN(p.Longitude) || math.IsInf(p.Longitude, 0) {
		return ErrInvalidCoordinate
	}
	if p.Latitude < -90 || p.Latitude > 90 || p.Longitude < -180 || p.Longitude > 180 {
		return ErrInvalidCoordinate
	}
	return nil
}

// Evaluate checks named fixtures in deterministic name order. Coordinates are
// used only for validation and are never copied into Result.
func Evaluate(samples []Sample) []Result {
	ordered := append([]Sample(nil), samples...)
	sort.SliceStable(ordered, func(i, j int) bool { return ordered[i].Name < ordered[j].Name })
	results := make([]Result, 0, len(ordered))
	for _, sample := range ordered {
		result := Result{Name: sample.Name, ObservationMode: ObservationModeSyntheticOnly, Status: StatusPass}
		if sample.Name == "" || len(sample.Name) > 64 {
			result.Status = StatusFail
			result.Reason = "invalid synthetic fixture name"
		} else if err := ValidateSyntheticPoint(sample.Point); err != nil {
			result.Status = StatusFail
			result.Reason = "synthetic coordinate rejected"
		}
		results = append(results, result)
	}
	return results
}
