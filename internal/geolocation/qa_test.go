package geolocation

import (
	"encoding/json"
	"math"
	"strings"
	"testing"
)

func TestT33EvaluateSortsAndRedactsSyntheticFixtures(t *testing.T) {
	results := Evaluate([]Sample{
		{Name: "synthetic-b", Point: Point{Latitude: 48.8566, Longitude: 2.3522}},
		{Name: "synthetic-a", Point: Point{Latitude: 0, Longitude: 0}},
	})
	if len(results) != 2 || results[0].Name != "synthetic-a" || results[1].Name != "synthetic-b" {
		t.Fatalf("unexpected deterministic order: %#v", results)
	}
	for _, result := range results {
		if result.Status != StatusPass || result.ObservationMode != ObservationModeSyntheticOnly {
			t.Fatalf("unexpected result: %#v", result)
		}
	}
	body, err := json.Marshal(results)
	if err != nil {
		t.Fatal(err)
	}
	encoded := string(body)
	for _, forbidden := range []string{"48.8566", "2.3522", "latitude", "longitude", "coordinates"} {
		if strings.Contains(strings.ToLower(encoded), strings.ToLower(forbidden)) {
			t.Fatalf("result leaked coordinate data %q: %s", forbidden, encoded)
		}
	}
}

func TestT33ValidateSyntheticPointFailsClosed(t *testing.T) {
	for _, point := range []Point{
		{Latitude: -91, Longitude: 0},
		{Latitude: 91, Longitude: 0},
		{Latitude: 0, Longitude: -181},
		{Latitude: 0, Longitude: 181},
		{Latitude: math.NaN(), Longitude: 0},
		{Latitude: 0, Longitude: math.Inf(1)},
	} {
		if err := ValidateSyntheticPoint(point); err != ErrInvalidCoordinate {
			t.Fatalf("point %#v: want ErrInvalidCoordinate, got %v", point, err)
		}
	}
}

func TestT33EvaluateMarksInvalidFixtureWithoutExposingCause(t *testing.T) {
	results := Evaluate([]Sample{{Name: "synthetic-invalid", Point: Point{Latitude: 100, Longitude: 0}}})
	if len(results) != 1 || results[0].Status != StatusFail {
		t.Fatalf("unexpected invalid result: %#v", results)
	}
	if results[0].Reason != "synthetic coordinate rejected" {
		t.Fatalf("unexpected reason: %#v", results[0])
	}
}
