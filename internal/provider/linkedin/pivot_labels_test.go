package linkedin

import "testing"

func TestIndustryLabelForURN_knownIDs(t *testing.T) {
	t.Parallel()

	cases := map[string]string{
		"urn:li:industry:6":    "Technology, Information and Internet",
		"urn:li:industry:96":   "IT Services and IT Consulting",
		"urn:li:industry:129":  "Capital Markets",
		"urn:li:industry:11":   "Business Consulting and Services",
		"urn:li:industry:1810": "Professional Services",
		"urn:li:industry:43":   "Financial Services",
	}
	for urn, want := range cases {
		got, ok := industryLabelForURN(urn)
		if !ok {
			t.Fatalf("industryLabelForURN(%q) = !ok", urn)
		}
		if got != want {
			t.Fatalf("industryLabelForURN(%q) = %q, want %q", urn, got, want)
		}
	}
}

func TestIndustryLabelForURN_unknownID(t *testing.T) {
	t.Parallel()

	if _, ok := industryLabelForURN("urn:li:industry:999999999"); ok {
		t.Fatal("expected unknown industry ID to miss")
	}
	if _, ok := industryLabelForURN("urn:li:seniority:1"); ok {
		t.Fatal("expected non-industry URN to miss")
	}
}

func TestEnrichPivotLabels_memberIndustry(t *testing.T) {
	t.Parallel()

	row := map[string]any{
		"clicks":      int64(143),
		"impressions": int64(5814),
		"pivotValues": []any{"urn:li:industry:6", "urn:li:industry:96"},
	}
	enrichPivotLabels(row, []string{"MEMBER_INDUSTRY"})

	labels, ok := row["pivotLabels"].([]any)
	if !ok {
		t.Fatalf("pivotLabels = %#v, want []any", row["pivotLabels"])
	}
	if len(labels) != 2 {
		t.Fatalf("len(pivotLabels) = %d, want 2", len(labels))
	}
	if labels[0] != "Technology, Information and Internet" {
		t.Fatalf("pivotLabels[0] = %v", labels[0])
	}
	if labels[1] != "IT Services and IT Consulting" {
		t.Fatalf("pivotLabels[1] = %v", labels[1])
	}
}

func TestEnrichPivotLabels_skipsNonIndustryPivot(t *testing.T) {
	t.Parallel()

	row := map[string]any{
		"pivotValues": []any{"urn:li:sponsoredCampaign:123"},
	}
	enrichPivotLabels(row, []string{"CAMPAIGN"})

	if _, ok := row["pivotLabels"]; ok {
		t.Fatal("expected pivotLabels to be omitted for CAMPAIGN pivot")
	}
}
