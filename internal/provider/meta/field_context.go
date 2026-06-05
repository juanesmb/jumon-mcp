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

// embeddedFieldCatalog documents common Insights / entity fields for agent-safe filtering.
// Meta does not expose a stable public Graph field-metadata endpoint for ads_read tokens.
var embeddedFieldCatalog = []fieldContextEntry{
	{Name: "impressions", Levels: []string{"account", "campaign", "adset", "ad"}, Filterable: true, Sortable: true, Description: "Number of times ads were on screen."},
	{Name: "reach", Levels: []string{"account", "campaign", "adset", "ad"}, Filterable: true, Sortable: true, Description: "Number of people who saw ads at least once."},
	{Name: "clicks", Levels: []string{"account", "campaign", "adset", "ad"}, Filterable: true, Sortable: true, Description: "Total clicks on ads."},
	{Name: "spend", Levels: []string{"account", "campaign", "adset", "ad"}, Filterable: true, Sortable: true, Description: "Estimated total amount spent (account currency)."},
	{Name: "ctr", Levels: []string{"account", "campaign", "adset", "ad"}, Filterable: false, Sortable: true, Description: "Click-through rate."},
	{Name: "cpc", Levels: []string{"account", "campaign", "adset", "ad"}, Filterable: false, Sortable: true, Description: "Average cost per click."},
	{Name: "cpm", Levels: []string{"account", "campaign", "adset", "ad"}, Filterable: false, Sortable: true, Description: "Average cost per 1,000 impressions."},
	{Name: "frequency", Levels: []string{"account", "campaign", "adset", "ad"}, Filterable: false, Sortable: true, Description: "Average impressions per person."},
	{Name: "campaign_id", Levels: []string{"campaign", "adset", "ad"}, Filterable: true, Sortable: false, Description: "Campaign ID."},
	{Name: "campaign_name", Levels: []string{"campaign", "adset", "ad"}, Filterable: true, Sortable: true, Description: "Campaign name."},
	{Name: "adset_id", Levels: []string{"adset", "ad"}, Filterable: true, Sortable: false, Description: "Ad set ID."},
	{Name: "adset_name", Levels: []string{"adset", "ad"}, Filterable: true, Sortable: true, Description: "Ad set name."},
	{Name: "ad_id", Levels: []string{"ad"}, Filterable: true, Sortable: false, Description: "Ad ID."},
	{Name: "ad_name", Levels: []string{"ad"}, Filterable: true, Sortable: true, Description: "Ad name."},
	{Name: "objective", Levels: []string{"campaign"}, Filterable: true, Sortable: false, Description: "ODAX campaign objective (OUTCOME_*)."},
	{Name: "effective_status", Levels: []string{"campaign", "adset", "ad"}, Filterable: true, Sortable: false, Description: "Delivery status considering parent state."},
	{Name: "date_start", Levels: []string{"account", "campaign", "adset", "ad"}, Filterable: false, Sortable: false, Description: "Start date for the report row."},
	{Name: "date_stop", Levels: []string{"account", "campaign", "adset", "ad"}, Filterable: false, Sortable: false, Description: "End date for the report row."},
	{Name: "actions:omni_purchase", Levels: []string{"account", "campaign", "adset", "ad"}, Filterable: false, Sortable: false, Description: "Purchase actions across channels (not valid in default insights fields; request explicitly if needed)."},
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
