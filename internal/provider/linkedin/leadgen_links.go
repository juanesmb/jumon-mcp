package linkedin

import "strings"

const (
	fieldLeadFormUrn      = "leadFormUrn"
	fieldLeadFormCtaLabel = "leadFormCtaLabel"
	fieldLeadFormName     = "leadFormName"

	adFormURNPrefix = "urn:li:adForm:"
)

// leadGenCTAFromCreative extracts the lead gen CTA label and form destination URN from a
// creative row. Returns (label, formURN, true) when a leadgenCallToAction with a non-empty
// destination is present; label may be an empty string if omitted by LinkedIn.
func leadGenCTAFromCreative(row map[string]any) (ctaLabel, formURN string, ok bool) {
	cta, hasCTA := row["leadgenCallToAction"].(map[string]any)
	if !hasCTA {
		return "", "", false
	}
	label, _ := cta["label"].(string)
	destination, _ := cta["destination"].(string)
	destination = strings.TrimSpace(destination)
	if destination == "" {
		return "", "", false
	}
	return strings.TrimSpace(label), destination, true
}

// formIDFromAdFormURN extracts the numeric string ID from a plain adForm URN
// (e.g. "urn:li:adForm:12345" → "12345"). Returns false for versioned lead gen form URNs
// or any other URN type—those cannot be resolved via batch GET leadForms?ids=List(...).
func formIDFromAdFormURN(urn string) (string, bool) {
	trimmed := strings.TrimSpace(urn)
	if !strings.HasPrefix(trimmed, adFormURNPrefix) {
		return "", false
	}
	id := strings.TrimSpace(trimmed[len(adFormURNPrefix):])
	if id == "" {
		return "", false
	}
	return id, true
}
