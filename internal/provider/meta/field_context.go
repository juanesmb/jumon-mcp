package meta

import (
	"strings"
)

type fieldContextEntry struct {
	Name        string   `json:"name"`
	Levels      []string `json:"levels"`
	Filterable  bool     `json:"filterable"`
	Sortable    bool     `json:"sortable"`
	Description string   `json:"description,omitempty"`
}

func buildFieldContextResponse(in fieldContextInput) any {
	names := in.fieldNames
	level := strings.TrimSpace(in.level)
	out := make([]fieldContextEntry, 0)
	for _, entry := range embeddedFieldCatalog {
		if len(names) > 0 && !containsString(names, entry.Name) {
			continue
		}
		if level != "" && !levelMatches(entry.Levels, level) {
			continue
		}
		out = append(out, entry)
	}
	return map[string]any{
		"fields": out,
		"source": "embedded_catalog",
		"hint":   "Call meta_get_field_context before filtering or sorting in meta_search_ad_entities. Do not request standalone actions or action_values fields.",
	}
}

func containsString(values []string, target string) bool {
	for _, v := range values {
		if strings.EqualFold(strings.TrimSpace(v), target) {
			return true
		}
	}
	return false
}

func levelMatches(levels []string, level string) bool {
	for _, l := range levels {
		if l == level {
			return true
		}
	}
	return false
}
