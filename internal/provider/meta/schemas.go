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
		"limit": map[string]any{
			"type":        "integer",
			"description": "Results per page (default 25, max 1000 for insights).",
		},
		"after":       map[string]any{"type": "string"},
		"auto_paginate": map[string]any{"type": "boolean", "description": docAutoPaginate},
	}
	if includeLevel {
		props["level"] = map[string]any{
			"type":        "string",
			"enum":        []string{"account", "campaign", "adset", "ad"},
			"description": docInsightsLevels + " Default campaign for meta_search_ad_entities.",
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

func getAdSetSchema() map[string]any {
	return map[string]any{
		"type":     "object",
		"required": []string{"adset_id"},
		"properties": map[string]any{
			"adset_id": map[string]any{"type": "string", "description": "Meta ad set id."},
			"fields":   map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
		},
	}
}

func getAdSchema() map[string]any {
	return map[string]any{
		"type":     "object",
		"required": []string{"ad_id"},
		"properties": map[string]any{
			"ad_id":  map[string]any{"type": "string", "description": "Meta ad id."},
			"fields": map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
		},
	}
}

func deliveryErrorsSchema() map[string]any {
	return map[string]any{
		"type":     "object",
		"required": []string{"entity_ids"},
		"properties": map[string]any{
			"entity_ids": map[string]any{
				"type":        "array",
				"items":       map[string]any{"type": "string"},
				"description": "Campaign, ad set, or ad ids to fetch delivery errors for (max 50).",
			},
		},
	}
}

func listAccountPagesSchema() map[string]any {
	props := listPaginationProperties()
	delete(props, "effective_status")
	props["act_id"] = map[string]any{
		"type":        "string",
		"description": "Optional. When set, lists Pages promotable from this ad account via GET /{act_id}/promote_pages. When omitted, uses GET /me/accounts.",
	}
	return map[string]any{
		"type":       "object",
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

func listCreativesSchema() map[string]any {
	props := listPaginationProperties()
	props["filtering"] = map[string]any{
		"type":        "array",
		"description": "Additional filter objects with field, operator, value.",
		"items":       map[string]any{"type": "object"},
	}
	return map[string]any{
		"type":       "object",
		"required":   []string{"act_id"},
		"properties": props,
	}
}

func getCreativeSchema() map[string]any {
	return map[string]any{
		"type":     "object",
		"required": []string{"creative_id"},
		"properties": map[string]any{
			"creative_id": map[string]any{"type": "string", "description": "Meta ad creative id."},
			"fields":      map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
			"thumbnail_width":  map[string]any{"type": "integer", "description": "Thumbnail width in pixels (default 64)."},
			"thumbnail_height": map[string]any{"type": "integer", "description": "Thumbnail height in pixels (default 64)."},
		},
	}
}

func adImagesSchema() map[string]any {
	props := listPaginationProperties()
	props["hashes"] = map[string]any{
		"type":        "array",
		"items":       map[string]any{"type": "string"},
		"description": "Filter by image hashes.",
	}
	props["name"] = map[string]any{"type": "string", "description": "Filter by image name (partial match)."}
	props["minwidth"] = map[string]any{"type": "integer", "description": "Minimum image width in pixels."}
	props["minheight"] = map[string]any{"type": "integer", "description": "Minimum image height in pixels."}
	return map[string]any{
		"type":       "object",
		"required":   []string{"act_id"},
		"properties": props,
	}
}

func adVideosSchema() map[string]any {
	props := listPaginationProperties()
	props["video_ids"] = map[string]any{
		"type":        "array",
		"items":       map[string]any{"type": "string"},
		"description": "Optional filter to specific video ids.",
	}
	return map[string]any{
		"type":       "object",
		"required":   []string{"act_id"},
		"properties": props,
	}
}

func adPreviewSchema() map[string]any {
	return map[string]any{
		"type":     "object",
		"required": []string{"ad_id"},
		"properties": map[string]any{
			"ad_id": map[string]any{"type": "string", "description": "Meta ad id to preview."},
			"ad_format": map[string]any{
				"type":        "string",
				"description": "Placement format e.g. DESKTOP_FEED_STANDARD, INSTAGRAM_STANDARD, INSTAGRAM_STORY, FACEBOOK_REELS_MOBILE.",
			},
			"locale":     map[string]any{"type": "string", "description": "Preview locale e.g. en_US."},
			"start_date": map[string]any{"type": "string", "description": "Preview start date (UNIX timestamp) for scheduled ads."},
			"end_date":   map[string]any{"type": "string", "description": "Preview end date (UNIX timestamp) for scheduled ads."},
		},
	}
}

func searchInterestsSchema() map[string]any {
	return map[string]any{
		"type":     "object",
		"required": []string{"q"},
		"properties": map[string]any{
			"q":     map[string]any{"type": "string", "description": "Interest search keyword."},
			"limit": map[string]any{"type": "integer", "description": "Max results (default 25)."},
		},
	}
}

func searchGeoLocationsSchema() map[string]any {
	return map[string]any{
		"type":     "object",
		"required": []string{"q"},
		"properties": map[string]any{
			"q": map[string]any{"type": "string", "description": "Geo search query e.g. New York, Japan."},
			"location_types": map[string]any{
				"type":        "array",
				"items":       map[string]any{"type": "string"},
				"description": "Optional types: country, region, city, zip, geo_market, electoral_district.",
			},
			"limit": map[string]any{"type": "integer", "description": "Max results (default 25)."},
		},
	}
}

func estimateAudienceSizeSchema() map[string]any {
	return map[string]any{
		"type":     "object",
		"required": []string{"act_id", "targeting"},
		"properties": map[string]any{
			"act_id": actIDProperty(),
			"targeting": map[string]any{
				"type":        "object",
				"description": "Full targeting spec (geo_locations, age_min, age_max, interests, flexible_spec, etc.).",
			},
			"optimization_goal": map[string]any{
				"type":        "string",
				"description": "Default REACH. Other values: LINK_CLICKS, LANDING_PAGE_VIEWS, OFFSITE_CONVERSIONS, IMPRESSIONS.",
			},
		},
	}
}

func listCustomAudiencesSchema() map[string]any {
	props := listPaginationProperties()
	props["subtype_filter"] = map[string]any{
		"type":        "string",
		"description": "Optional subtype filter: CUSTOM, WEBSITE, LOOKALIKE, APP, ENGAGEMENT, OFFLINE_CONVERSION.",
	}
	return map[string]any{
		"type":       "object",
		"required":   []string{"act_id"},
		"properties": props,
	}
}

func getCustomAudienceSchema() map[string]any {
	return map[string]any{
		"type":     "object",
		"required": []string{"custom_audience_id"},
		"properties": map[string]any{
			"custom_audience_id": map[string]any{"type": "string", "description": "Custom audience id."},
			"fields":             map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
		},
	}
}

func listCustomAudienceAdSetsSchema() map[string]any {
	props := listPaginationProperties()
	delete(props, "act_id")
	delete(props, "effective_status")
	props["custom_audience_id"] = map[string]any{"type": "string", "description": "Custom audience id."}
	return map[string]any{
		"type":       "object",
		"required":   []string{"custom_audience_id"},
		"properties": props,
	}
}

func opportunityScoreSchema() map[string]any {
	return map[string]any{
		"type":     "object",
		"required": []string{"act_id"},
		"properties": map[string]any{
			"act_id": actIDProperty(),
		},
	}
}

func listCustomConversionsSchema() map[string]any {
	props := listPaginationProperties()
	props["dataset_id"] = map[string]any{
		"type":        "string",
		"description": docDatasetID + " Optional filter to conversions tied to one pixel.",
	}
	return map[string]any{
		"type":       "object",
		"required":   []string{"act_id"},
		"properties": props,
	}
}

func listDatasetsSchema() map[string]any {
	props := listPaginationProperties()
	return map[string]any{
		"type":       "object",
		"required":   []string{"act_id"},
		"properties": props,
	}
}

func getDatasetSchema() map[string]any {
	return map[string]any{
		"type":     "object",
		"required": []string{"dataset_id"},
		"properties": map[string]any{
			"dataset_id": map[string]any{"type": "string", "description": docDatasetID},
			"fields":     map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
		},
	}
}

func listCreativeAdsSchema() map[string]any {
	props := listPaginationProperties()
	delete(props, "effective_status")
	props["creative_id"] = map[string]any{"type": "string", "description": "Meta ad creative id."}
	return map[string]any{
		"type":       "object",
		"required":   []string{"act_id", "creative_id"},
		"properties": props,
	}
}

func activitiesProperties() map[string]any {
	return map[string]any{
		"fields": map[string]any{
			"type":        "array",
			"items":       map[string]any{"type": "string"},
			"description": "Activity log fields (default: actor_name, object_type, translated_event_type, event_time, changed_data).",
		},
		"time_range": map[string]any{
			"type":        "object",
			"description": docActivitiesTime,
			"properties": map[string]any{
				"since": map[string]any{"type": "string"},
				"until": map[string]any{"type": "string"},
			},
		},
		"since": map[string]any{"type": "string", "description": "Start date YYYY-MM-DD (ignored when time_range is set)."},
		"until": map[string]any{"type": "string", "description": "End date YYYY-MM-DD (ignored when time_range is set)."},
		"limit": map[string]any{"type": "integer", "description": "Results per page (default 25, max 100)."},
		"after":  map[string]any{"type": "string"},
		"before": map[string]any{"type": "string"},
	}
}

func accountActivitiesSchema() map[string]any {
	props := activitiesProperties()
	props["act_id"] = actIDProperty()
	return map[string]any{
		"type":       "object",
		"required":   []string{"act_id"},
		"properties": props,
	}
}

func adSetActivitiesSchema() map[string]any {
	props := activitiesProperties()
	props["adset_id"] = map[string]any{"type": "string", "description": "Meta ad set id."}
	return map[string]any{
		"type":       "object",
		"required":   []string{"adset_id"},
		"properties": props,
	}
}

func interestSuggestionsSchema() map[string]any {
	return map[string]any{
		"type":     "object",
		"required": []string{"interest_list"},
		"properties": map[string]any{
			"interest_list": map[string]any{
				"type":        "array",
				"items":       map[string]any{"type": "string"},
				"description": docInterestSuggestions,
			},
			"limit": map[string]any{"type": "integer", "description": "Max results (default 25)."},
		},
	}
}

func datasetQualitySchema() map[string]any {
	return map[string]any{
		"type":     "object",
		"required": []string{"dataset_id"},
		"properties": map[string]any{
			"dataset_id": map[string]any{"type": "string", "description": docDatasetID},
			"fields": map[string]any{
				"type":        "string",
				"description": docDatasetQualityNote + " Default: web{event_match_quality,event_name}.",
			},
		},
	}
}
