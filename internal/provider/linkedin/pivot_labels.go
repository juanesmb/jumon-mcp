package linkedin

import (
	_ "embed"
	"encoding/json"
	"strings"
)

//go:embed data/industries.json
var embeddedIndustriesJSON []byte

const (
	fieldPivotValues = "pivotValues"
	fieldPivotLabels = "pivotLabels"

	pivotMemberIndustry = "MEMBER_INDUSTRY"
	urnPrefixIndustry   = "urn:li:industry:"
)

var industryLabels = mustLoadIndustryLabels()

func mustLoadIndustryLabels() map[string]string {
	labels := map[string]string{}
	if err := json.Unmarshal(embeddedIndustriesJSON, &labels); err != nil {
		panic("linkedin: failed to load embedded industries.json: " + err.Error())
	}
	return labels
}

func industryLabelForURN(urn string) (string, bool) {
	id, ok := industryIDFromURN(urn)
	if !ok {
		return "", false
	}
	label, ok := industryLabels[id]
	return label, ok
}

func industryIDFromURN(urn string) (string, bool) {
	trimmed := strings.TrimSpace(urn)
	if !strings.HasPrefix(trimmed, urnPrefixIndustry) {
		return "", false
	}
	id := strings.TrimPrefix(trimmed, urnPrefixIndustry)
	if id == "" {
		return "", false
	}
	return id, true
}

func labelForPivotValue(pivot, value string) (string, bool) {
	switch strings.TrimSpace(pivot) {
	case pivotMemberIndustry:
		return industryLabelForURN(value)
	default:
		return "", false
	}
}

func pivotAt(pivots []string, index int) string {
	if index < len(pivots) {
		return pivots[index]
	}
	if len(pivots) == 1 {
		return pivots[0]
	}
	return ""
}

func enrichPivotLabels(row map[string]any, pivots []string) {
	pivots = trimPivots(pivots)
	if len(pivots) == 0 {
		return
	}

	rawValues, ok := row[fieldPivotValues]
	if !ok || rawValues == nil {
		return
	}

	valueStrings := stringSliceFromAny(rawValues)
	if len(valueStrings) == 0 {
		return
	}

	labels := make([]any, len(valueStrings))
	resolved := false
	for i, value := range valueStrings {
		if label, ok := labelForPivotValue(pivotAt(pivots, i), value); ok {
			labels[i] = label
			resolved = true
		}
	}
	if !resolved {
		return
	}
	row[fieldPivotLabels] = labels
}

func stringSliceFromAny(raw any) []string {
	switch values := raw.(type) {
	case []string:
		return values
	case []any:
		out := make([]string, 0, len(values))
		for _, item := range values {
			s, ok := item.(string)
			if !ok {
				continue
			}
			out = append(out, s)
		}
		return out
	default:
		return nil
	}
}
