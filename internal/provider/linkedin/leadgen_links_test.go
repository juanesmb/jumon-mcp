package linkedin

import "testing"

func TestLeadGenCTAFromCreative_extractsLabelAndDestination(t *testing.T) {
	t.Parallel()

	row := map[string]any{
		"leadgenCallToAction": map[string]any{
			"label":       "Download",
			"destination": "urn:li:adForm:12345",
		},
	}
	label, formURN, ok := leadGenCTAFromCreative(row)
	if !ok {
		t.Fatal("expected ok")
	}
	if label != "Download" {
		t.Fatalf("label = %q, want %q", label, "Download")
	}
	if formURN != "urn:li:adForm:12345" {
		t.Fatalf("formURN = %q, want %q", formURN, "urn:li:adForm:12345")
	}
}

func TestLeadGenCTAFromCreative_labelOptional(t *testing.T) {
	t.Parallel()

	row := map[string]any{
		"leadgenCallToAction": map[string]any{
			"destination": "urn:li:adForm:99",
		},
	}
	label, formURN, ok := leadGenCTAFromCreative(row)
	if !ok {
		t.Fatal("expected ok even when label is absent")
	}
	if label != "" {
		t.Fatalf("label = %q, want empty", label)
	}
	if formURN != "urn:li:adForm:99" {
		t.Fatalf("formURN = %q", formURN)
	}
}

func TestLeadGenCTAFromCreative_missingField(t *testing.T) {
	t.Parallel()

	row := map[string]any{"id": "urn:li:sponsoredCreative:1"}
	_, _, ok := leadGenCTAFromCreative(row)
	if ok {
		t.Fatal("expected !ok for row without leadgenCallToAction")
	}
}

func TestLeadGenCTAFromCreative_emptyDestination(t *testing.T) {
	t.Parallel()

	row := map[string]any{
		"leadgenCallToAction": map[string]any{
			"label":       "Download",
			"destination": "   ",
		},
	}
	_, _, ok := leadGenCTAFromCreative(row)
	if ok {
		t.Fatal("expected !ok when destination is blank")
	}
}

func TestFormIDFromAdFormURN(t *testing.T) {
	t.Parallel()

	id, ok := formIDFromAdFormURN("urn:li:adForm:12345")
	if !ok || id != "12345" {
		t.Fatalf("formIDFromAdFormURN() = %q, %v, want \"12345\", true", id, ok)
	}
}

func TestFormIDFromAdFormURN_rejectsOtherURNs(t *testing.T) {
	t.Parallel()

	cases := []string{
		"urn:li:versionedLeadGenForm:(urn:li:leadGenForm:3162,1)",
		"urn:li:sponsoredCreative:123",
		"urn:li:adForm:",
		"",
	}
	for _, urn := range cases {
		if _, ok := formIDFromAdFormURN(urn); ok {
			t.Fatalf("expected formIDFromAdFormURN(%q) = !ok", urn)
		}
	}
}
