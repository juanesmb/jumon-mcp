package linkedin

import (
	"slices"
	"testing"
)

func TestDeriveAverageFrequency(t *testing.T) {
	t.Parallel()

	got := deriveAverageFrequency(8512, 6454)
	if got != 1.32 {
		t.Fatalf("deriveAverageFrequency() = %v, want 1.32", got)
	}
}

func TestDeriveAverageFrequency_zeroReach(t *testing.T) {
	t.Parallel()

	if got := deriveAverageFrequency(100, 0); got != 0 {
		t.Fatalf("deriveAverageFrequency() = %v, want 0", got)
	}
}

func TestEnrichAnalyticsElement_addsAverageFrequency(t *testing.T) {
	t.Parallel()

	row := map[string]any{
		"impressions":            int64(8512),
		"approximateMemberReach": int64(6454),
	}
	enrichAnalyticsElement(row)

	freq, ok := row[fieldAverageFrequency].(float64)
	if !ok {
		t.Fatalf("averageFrequency missing or wrong type: %#v", row[fieldAverageFrequency])
	}
	if freq != 1.32 {
		t.Fatalf("averageFrequency = %v, want 1.32", freq)
	}
}

func TestEnrichAnalyticsElement_skipsWithoutReach(t *testing.T) {
	t.Parallel()

	row := map[string]any{"impressions": int64(100)}
	enrichAnalyticsElement(row)

	if _, ok := row[fieldAverageFrequency]; ok {
		t.Fatal("expected averageFrequency to be omitted when reach is missing")
	}
}

func TestEnrichAnalyticsElement_usesLegacyReachField(t *testing.T) {
	t.Parallel()

	row := map[string]any{
		"impressions":                  int64(1000),
		"approximateUniqueImpressions": int64(500),
	}
	enrichAnalyticsElement(row)

	if row[fieldAverageFrequency] != 2.0 {
		t.Fatalf("averageFrequency = %v, want 2.0", row[fieldAverageFrequency])
	}
}

func TestEnrichAnalyticsResponse_enrichesElements(t *testing.T) {
	t.Parallel()

	payload := map[string]any{
		"elements": []any{
			map[string]any{
				"impressions":            int64(8512),
				"approximateMemberReach": int64(6454),
			},
		},
	}
	enrichAnalyticsResponse(payload, nil)

	row := payload["elements"].([]any)[0].(map[string]any)
	if row[fieldAverageFrequency] != 1.32 {
		t.Fatalf("averageFrequency = %v, want 1.32", row[fieldAverageFrequency])
	}
}

func TestEnrichAnalyticsResponse_preservesNonMapPayload(t *testing.T) {
	t.Parallel()

	payload := "unexpected"
	if got := enrichAnalyticsResponse(payload, nil); got != payload {
		t.Fatalf("enrichAnalyticsResponse() = %#v, want unchanged payload", got)
	}
}

func TestEnrichAnalyticsResponse_addsIndustryPivotLabels(t *testing.T) {
	t.Parallel()

	payload := map[string]any{
		"elements": []any{
			map[string]any{
				"pivotValues": []any{"urn:li:industry:6"},
			},
		},
	}
	enrichAnalyticsResponse(payload, []string{"MEMBER_INDUSTRY"})

	row := payload["elements"].([]any)[0].(map[string]any)
	labels := row["pivotLabels"].([]any)
	if labels[0] != "Technology, Information and Internet" {
		t.Fatalf("pivotLabels[0] = %v", labels[0])
	}
}

func TestEnsureDeliveryMetricFields(t *testing.T) {
	t.Parallel()

	fields := ensureDeliveryMetricFields([]string{"clicks", "approximateMemberReach"})
	for _, name := range []string{"approximateMemberReach", "impressions", "clicks"} {
		if !slices.Contains(fields, name) {
			t.Fatalf("ensureDeliveryMetricFields() = %v, missing %q", fields, name)
		}
	}
}
