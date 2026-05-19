package linkedin

import (
	"strings"
	"testing"
)

func TestBuildAnalyticsFieldsDescription_includesCatalogAndConstraints(t *testing.T) {
	t.Parallel()

	desc := buildAnalyticsFieldsDescription()
	for _, needle := range []string{
		"approximateMemberReach",
		"oneClickLeads",
		"videoViews",
		"max 20 per request",
		"averageFrequency",
		"92 days",
	} {
		if !strings.Contains(desc, needle) {
			t.Fatalf("description missing %q:\n%s", needle, desc)
		}
	}
}

func TestAnalyticsFieldsSchemaProperty_hasExamples(t *testing.T) {
	t.Parallel()

	prop := analyticsFieldsSchemaProperty()
	examples, ok := prop["examples"].([][]string)
	if !ok || len(examples) < 3 {
		t.Fatalf("expected examples [][]string, got %#v", prop["examples"])
	}
}
