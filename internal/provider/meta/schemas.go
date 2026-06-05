package meta

func listAdAccountsSchema() map[string]any {
	return map[string]any{
		"type":       "object",
		"properties": map[string]any{},
	}
}

func actIDProperty() map[string]any {
	return map[string]any{
		"type":        "string",
		"description": docActID,
	}
}

func listPaginationProperties() map[string]any {
	return map[string]any{
		"act_id": actIDProperty(),
		"fields": map[string]any{
			"type":        "array",
			"items":       map[string]any{"type": "string"},
			"description": "Graph fields to return. Defaults are curated per tool.",
		},
		"effective_status": map[string]any{
			"type":        "array",
			"items":       map[string]any{"type": "string"},
			"description": "Filter by effective_status (ACTIVE, PAUSED, etc.).",
		},
		"limit": map[string]any{
			"type":        "integer",
			"description": "Results per page (default 25, max 100).",
		},
		"after": map[string]any{
			"type":        "string",
			"description": "Pagination cursor from paging.cursors.after.",
		},
		"before": map[string]any{
			"type":        "string",
			"description": "Pagination cursor from paging.cursors.before.",
		},
		"auto_paginate": map[string]any{
			"type":        "boolean",
			"description": docAutoPaginate,
		},
	}
}

func getAdAccountSchema() map[string]any {
	return map[string]any{
		"type":     "object",
		"required": []string{"act_id"},
		"properties": map[string]any{
			"act_id": actIDProperty(),
			"fields": map[string]any{
				"type":  "array",
				"items": map[string]any{"type": "string"},
			},
		},
	}
}

func listCampaignsSchema() map[string]any {
	return map[string]any{
		"type":       "object",
		"required":   []string{"act_id"},
		"properties": listPaginationProperties(),
	}
}

func getCampaignSchema() map[string]any {
	return map[string]any{
		"type":     "object",
		"required": []string{"campaign_id"},
		"properties": map[string]any{
			"campaign_id": map[string]any{"type": "string", "description": "Meta campaign id."},
			"fields":      map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
		},
	}
}

func listAdSetsSchema() map[string]any {
	props := listPaginationProperties()
	props["campaign_id"] = map[string]any{
		"type":        "string",
		"description": "Optional. Filter ad sets to one campaign.",
	}
	return map[string]any{
		"type":       "object",
		"required":   []string{"act_id"},
		"properties": props,
	}
}

func listAdsSchema() map[string]any {
	props := listPaginationProperties()
	props["campaign_id"] = map[string]any{"type": "string", "description": "Optional campaign filter."}
	props["adset_id"] = map[string]any{"type": "string", "description": "Optional ad set filter."}
	return map[string]any{
		"type":       "object",
		"required":   []string{"act_id"},
		"properties": props,
	}
}

func insightsProperties(includeLevel bool) map[string]any {
	props := map[string]any{
		"fields": map[string]any{
			"type":        "array",
			"items":       map[string]any{"type": "string"},
			"description": "Insights metrics and dimensions.",
		},
		"date_preset": map[string]any{
			"type":        "string",
			"description": "Relative window (default last_30d). Overridden by explicit time_range/time_ranges/since/until.",
		},
		"time_range": map[string]any{
			"type":        "object",
			"description": "Custom range {since: YYYY-MM-DD, until: YYYY-MM-DD}.",
			"properties": map[string]any{
				"since": map[string]any{"type": "string"},
				"until": map[string]any{"type": "string"},
			},
		},
		"time_ranges": map[string]any{
			"type":        "array",
			"description": "Array of time_range objects for period comparison.",
			"items": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"since": map[string]any{"type": "string"},
					"until": map[string]any{"type": "string"},
				},
			},
		},
		"time_increment": map[string]any{"type": "string", "description": "1-90 days, monthly, or all_days."},
		"breakdowns": map[string]any{
			"type":        "array",
			"items":       map[string]any{"type": "string"},
			"description": docPublisherPlatform,
		},
		"filtering": map[string]any{
			"type":        "array",
			"description": "Filter objects {field, operator, value}.",
			"items":       map[string]any{"type": "object"},
		},
		"sort":        map[string]any{"type": "string", "description": "Sort key e.g. impressions_descending."},
		"since":       map[string]any{"type": "string"},
		"until":       map[string]any{"type": "string"},
		"limit":       map[string]any{"type": "integer"},
		"after":       map[string]any{"type": "string"},
		"auto_paginate": map[string]any{"type": "boolean", "description": docAutoPaginate},
	}
	if includeLevel {
		props["level"] = map[string]any{
			"type":        "string",
			"enum":        []string{"account", "campaign", "adset", "ad"},
			"description": "Aggregation level (default campaign for meta_search_ad_entities).",
		}
	}
	return props
}

func adAccountInsightsSchema() map[string]any {
	props := insightsProperties(false)
	props["act_id"] = actIDProperty()
	return map[string]any{
		"type":       "object",
		"required":   []string{"act_id"},
		"properties": props,
	}
}

func campaignInsightsSchema() map[string]any {
	props := insightsProperties(false)
	props["campaign_id"] = map[string]any{"type": "string"}
	return map[string]any{
		"type":       "object",
		"required":   []string{"campaign_id"},
		"properties": props,
	}
}

func searchAdEntitiesSchema() map[string]any {
	props := insightsProperties(true)
	props["act_id"] = actIDProperty()
	return map[string]any{
		"type":       "object",
		"required":   []string{"act_id"},
		"properties": props,
	}
}

func fieldContextSchema() map[string]any {
	return map[string]any{
		"type":     "object",
		"properties": map[string]any{
			"field_names": map[string]any{
				"type":        "array",
				"items":       map[string]any{"type": "string"},
				"description": "Optional subset of field names to return.",
			},
			"level": map[string]any{
				"type":        "string",
				"enum":        []string{"account", "campaign", "adset", "ad"},
				"description": "Optional filter to fields valid at this level.",
			},
		},
	}
}
