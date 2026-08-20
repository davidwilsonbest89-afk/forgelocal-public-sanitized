package templates

import (
	"context"
	"errors"
	"strings"
	"testing"

	"forgelocal/internal/profile"
)

// T21-TEST-EVIDENCE-CORRECTION: this file proves validation already present in
// the TemplateRepository. It must not change Template, Profile, or SQLite behaviour.
func TestT21TemplateAcceptsAllT20NCFFieldTypes(t *testing.T) {
	store := openTestStore(t)
	_, err := store.Create(context.Background(), "all field types", Content{
		CustomFields: map[string]profile.CustomField{
			"text_field":    {Type: "text", Value: "plain-value"},
			"number_field":  {Type: "number", Value: float64(42)},
			"boolean_field": {Type: "boolean", Value: true},
			"select_field":  {Type: "select", Value: "gold", Options: []string{"silver", "gold"}},
		},
	}, "corr-t21-four-types")
	if err != nil {
		t.Fatalf("four T20-NCF field types must be accepted: %v", err)
	}
}

func TestT21TemplateRejectsInvalidMetadataFailClosed(t *testing.T) {
	longText := strings.Repeat("x", 2049)
	tooManyTags := make([]string, 21)
	for i := range tooManyTags {
		tooManyTags[i] = "tag" + string(rune('a'+i))
	}
	tooManyOptions := make([]string, 21)
	for i := range tooManyOptions {
		tooManyOptions[i] = "option" + string(rune('a'+i))
	}
	cases := []struct {
		name    string
		content Content
	}{
		{
			name:    "invalid custom field type",
			content: Content{CustomFields: map[string]profile.CustomField{"field": {Type: "unknown", Value: "value"}}},
		},
		{
			name:    "invalid custom field value type",
			content: Content{CustomFields: map[string]profile.CustomField{"field": {Type: "text", Value: float64(1)}}},
		},
		{
			name:    "forbidden custom field key",
			content: Content{CustomFields: map[string]profile.CustomField{"bad\nkey": {Type: "text", Value: "value"}}},
		},
		{
			name:    "overlong text value",
			content: Content{CustomFields: map[string]profile.CustomField{"field": {Type: "text", Value: longText}}},
		},
		{
			name:    "tag limit",
			content: Content{Tags: &tooManyTags},
		},
		{
			name:    "select option limit",
			content: Content{CustomFields: map[string]profile.CustomField{"field": {Type: "select", Value: "optiona", Options: tooManyOptions}}},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			store := openTestStore(t)
			_, err := store.Create(context.Background(), "invalid "+testCase.name, testCase.content, "corr-t21-invalid")
			if !errors.Is(err, ErrInvalidTemplate) {
				t.Fatalf("expected fail-closed invalid template error, got %v", err)
			}
		})
	}
}
