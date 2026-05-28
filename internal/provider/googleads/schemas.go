package googleads

func customerContextSchema() map[string]any {
	return map[string]any{
		"customer_id":       map[string]any{"type": "string", "description": "Google Ads customer id (digits only, no hyphens)."},
		"login_customer_id": map[string]any{"type": "string", "description": "Manager (MCC) customer id when querying a client account."},
	}
}

func reportFiltersSchema(extra map[string]any) map[string]any {
	props := map[string]any{
		"customer_id":       customerContextSchema()["customer_id"],
		"login_customer_id": customerContextSchema()["login_customer_id"],
		"campaign_ids":      map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
		"ad_group_ids":      map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
		"statuses":          map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
		"date_range_start":  map[string]any{"type": "string", "description": "YYYY-MM-DD"},
		"date_range_end":    map[string]any{"type": "string", "description": "YYYY-MM-DD"},
		"limit":          map[string]any{"type": "integer", "description": "Max rows (default 500, max 1000)."},
		"auto_paginate": map[string]any{
			"type":        "boolean",
			"description": "When true (default), follows nextPageToken to merge up to 10 pages of results.",
		},
	}
	for key, value := range extra {
		props[key] = value
	}
	return map[string]any{
		"type":       "object",
		"required":   []string{"customer_id"},
		"properties": props,
	}
}

func listClientAccountsSchema() map[string]any {
	return map[string]any{
		"type":     "object",
		"required": []string{"customer_id"},
		"properties": map[string]any{
			"customer_id":          customerContextSchema()["customer_id"],
			"login_customer_id":    customerContextSchema()["login_customer_id"],
			"client_name_contains": map[string]any{"type": "string", "description": "Optional filter on client descriptive_name."},
		},
	}
}

func resolveCustomerSchema() map[string]any {
	return map[string]any{
		"type":     "object",
		"required": []string{"account_name"},
		"properties": map[string]any{
			"account_name": map[string]any{"type": "string", "description": "Account name to match (descriptive_name)."},
			"match_mode": map[string]any{
				"type":        "string",
				"enum":        []string{"contains", "exact"},
				"description": "Default contains.",
			},
			"search_under_managers": map[string]any{
				"type":        "boolean",
				"description": "When true (default), also search client accounts under accessible MCCs (max 10).",
			},
		},
	}
}

func searchCampaignsSchema() map[string]any {
	return reportFiltersSchema(map[string]any{
		"campaign_name_contains": map[string]any{"type": "string"},
	})
}

func searchAdGroupsSchema() map[string]any {
	return reportFiltersSchema(nil)
}

func searchAdsSchema() map[string]any {
	return reportFiltersSchema(map[string]any{
		"ad_ids": map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
	})
}

func searchKeywordsSchema() map[string]any {
	return reportFiltersSchema(map[string]any{
		"keyword_contains": map[string]any{"type": "string", "description": "Filter keyword text (LIKE)."},
	})
}

func searchSearchTermsSchema() map[string]any {
	return reportFiltersSchema(map[string]any{
		"search_term_contains": map[string]any{"type": "string", "description": "Filter search term text (LIKE)."},
	})
}

func searchPmaxSearchTermsSchema() map[string]any {
	return reportFiltersSchema(map[string]any{
		"search_term_contains": map[string]any{"type": "string", "description": "Filter PMax search term text (LIKE)."},
	})
}

func listConversionActionsSchema() map[string]any {
	return reportFiltersSchema(map[string]any{
		"name_contains": map[string]any{"type": "string", "description": "Filter conversion action name (LIKE)."},
	})
}

func listOfflineConversionUploadSummariesSchema() map[string]any {
	return listConversionActionsSchema()
}

func searchConversionPerformanceSchema() map[string]any {
	return reportFiltersSchema(map[string]any{
		"conversion_action_ids": map[string]any{
			"type":        "array",
			"items":       map[string]any{"type": "string"},
			"description": "Numeric conversion action ids or full resource names.",
		},
	})
}

func getResourceMetadataSchema() map[string]any {
	return map[string]any{
		"type":     "object",
		"required": []string{"resource_name"},
		"properties": map[string]any{
			"resource_name": map[string]any{
				"type":        "string",
				"description": "Google Ads GAQL resource (e.g. campaign, keyword_view, search_term_view).",
			},
		},
	}
}

func searchGAQLSchema() map[string]any {
	return map[string]any{
		"type":     "object",
		"required": []string{"customer_id", "resource", "fields"},
		"properties": map[string]any{
			"customer_id":       customerContextSchema()["customer_id"],
			"login_customer_id": customerContextSchema()["login_customer_id"],
			"resource": map[string]any{
				"type":        "string",
				"description": "GAQL FROM resource; must be in the embedded allowlist.",
			},
			"fields": map[string]any{
				"type":        "array",
				"items":       map[string]any{"type": "string"},
				"description": "Fully qualified field names (resource.*, metrics.*, segments.*).",
			},
			"conditions": map[string]any{
				"type":        "array",
				"items":       map[string]any{"type": "string"},
				"description": "Optional WHERE fragments combined with AND.",
			},
			"orderings": map[string]any{
				"type":        "array",
				"items":       map[string]any{"type": "string"},
				"description": "Optional ORDER BY fields.",
			},
			"limit": map[string]any{
				"type":        "integer",
				"description": "Max rows (default 500, max 1000; change_event max 10000).",
			},
			"auto_paginate": map[string]any{
				"type":        "boolean",
				"description": "When true (default), follows nextPageToken to merge up to 10 pages of results.",
			},
		},
	}
}
