package linkedin

import (
	"strconv"
	"strings"
)

type linkedInSearchQuery struct {
	statusFilter []string
	typeFilter   []string
	nameFilter   []string
	testFilter   *bool
	sortOrder    string
	pageSize     int
	pageToken    string
}

type searchPagination struct {
	autoPaginate bool
	pageSize     int
	pageToken    string
}

func buildLinkedInSearchQuery(in linkedInSearchQuery) map[string]string {
	query := map[string]string{"q": "search"}
	appendListFinder(query, "status", in.statusFilter)
	appendListFinder(query, "type", in.typeFilter)
	appendListFinder(query, "name", in.nameFilter)
	if in.testFilter != nil {
		query["search.test"] = strconv.FormatBool(*in.testFilter)
	}
	if sortOrder := strings.TrimSpace(in.sortOrder); sortOrder != "" {
		query["sortOrder"] = sortOrder
	}
	query["pageSize"] = strconv.Itoa(in.pageSize)
	if token := strings.TrimSpace(in.pageToken); token != "" {
		query["pageToken"] = token
	}
	return query
}

func resolveAutoPaginate(requested bool, pageToken string, forceAllPages bool) bool {
	if strings.TrimSpace(pageToken) != "" {
		return false
	}
	if forceAllPages {
		return true
	}
	return requested
}

func parseSearchPagination(params map[string]any) searchPagination {
	return parsePagination(params, defaultCampaignsPageSize)
}

func parsePagination(params map[string]any, defaultPageSize int) searchPagination {
	autoPaginate := true
	if v, ok := params["auto_paginate"].(bool); ok {
		autoPaginate = v
	}
	pageSize := defaultPageSize
	if ps, ok := toInt(params["page_size"]); ok && ps > 0 {
		pageSize = ps
	}
	return searchPagination{
		autoPaginate: autoPaginate,
		pageSize:     pageSize,
		pageToken:    strings.TrimSpace(toString(params["page_token"])),
	}
}

func parseOptionalTestFilter(params map[string]any) *bool {
	if testValue, ok := params["test_filter"].(bool); ok {
		return &testValue
	}
	return nil
}

func parseOptionalBoolParam(value any, defaultValue bool) bool {
	if v, ok := value.(bool); ok {
		return v
	}
	return defaultValue
}

func filterCampaignsByGroup(raw any, groupIDs []string) any {
	if len(groupIDs) == 0 {
		return raw
	}

	allowed := make(map[string]struct{}, len(groupIDs))
	for _, urn := range toCampaignGroupURNs(groupIDs) {
		allowed[urn] = struct{}{}
	}

	pageMap, ok := raw.(map[string]any)
	if !ok {
		return raw
	}

	elements, ok := pageMap["elements"].([]any)
	if !ok {
		return raw
	}

	filtered := make([]any, 0, len(elements))
	for _, element := range elements {
		row, ok := element.(map[string]any)
		if !ok {
			continue
		}
		campaignGroup, _ := row["campaignGroup"].(string)
		if _, matches := allowed[campaignGroup]; matches {
			filtered = append(filtered, element)
		}
	}

	pageMap["elements"] = filtered
	return pageMap
}
