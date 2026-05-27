package googleads

import (
	"context"
	"fmt"
	"sort"
	"strings"
)

const fieldServicePageSize = 1000

func (s *service) getResourceMetadata(ctx context.Context, userID, mcpTool, resourceName string) (any, error) {
	resource, err := normalizeGAQLResourceName(resourceName)
	if err != nil {
		return nil, err
	}

	selectable := make(map[string]struct{})
	filterable := make(map[string]struct{})
	sortable := make(map[string]struct{})

	attributesQuery := fmt.Sprintf(
		"SELECT name, selectable, filterable, sortable WHERE name LIKE '%s.%%' AND category = 'ATTRIBUTE'",
		resource,
	)
	if err := s.collectFieldServiceRows(ctx, userID, mcpTool, attributesQuery, resource, selectable, filterable, sortable); err != nil {
		fallbackQuery := fmt.Sprintf(
			"SELECT name, selectable, filterable, sortable WHERE name LIKE '%s.%%'",
			resource,
		)
		if fallbackErr := s.collectFieldServiceRows(ctx, userID, mcpTool, fallbackQuery, resource, selectable, filterable, sortable); fallbackErr != nil {
			return nil, fmt.Errorf("google ads field metadata query failed: %w", err)
		}
	}

	metricsQuery := fmt.Sprintf(
		"SELECT name, selectable, filterable, sortable WHERE selectable_with CONTAINS ANY('%s')",
		resource,
	)
	_ = s.collectFieldServiceRows(ctx, userID, mcpTool, metricsQuery, resource, selectable, filterable, sortable)

	if len(selectable) == 0 && len(filterable) == 0 && len(sortable) == 0 {
		return nil, fmt.Errorf("no metadata found for resource %q; use a resource from the GAQL allowlist", resource)
	}

	return map[string]any{
		"resource":   resource,
		"selectable": sortedKeys(selectable),
		"filterable": sortedKeys(filterable),
		"sortable":   sortedKeys(sortable),
	}, nil
}

func (s *service) collectFieldServiceRows(
	ctx context.Context,
	userID, mcpTool, query, resource string,
	selectable, filterable, sortable map[string]struct{},
) error {
	pageToken := ""
	for {
		payload, err := s.googleAdsFieldSearch(ctx, userID, mcpTool, query, pageToken)
		if err != nil {
			return err
		}
		rows, nextToken, err := parseFieldServiceResponse(payload)
		if err != nil {
			return err
		}
		for _, row := range rows {
			name := strings.TrimSpace(row.name)
			if name == "" {
				continue
			}
			if !strings.HasPrefix(name, resource+".") &&
				!strings.HasPrefix(name, "metrics.") &&
				!strings.HasPrefix(name, "segments.") {
				continue
			}
			if row.selectable {
				selectable[name] = struct{}{}
			}
			if row.filterable {
				filterable[name] = struct{}{}
			}
			if row.sortable {
				sortable[name] = struct{}{}
			}
		}
		if nextToken == "" {
			return nil
		}
		pageToken = nextToken
	}
}

type fieldServiceRow struct {
	name       string
	selectable bool
	filterable bool
	sortable   bool
}

func parseFieldServiceResponse(payload any) ([]fieldServiceRow, string, error) {
	root, ok := payload.(map[string]any)
	if !ok {
		return nil, "", fmt.Errorf("unexpected field service response type")
	}
	nextToken, _ := root["nextPageToken"].(string)

	rawResults, ok := root["results"].([]any)
	if !ok {
		return nil, nextToken, nil
	}

	rows := make([]fieldServiceRow, 0, len(rawResults))
	for _, item := range rawResults {
		rowMap, ok := item.(map[string]any)
		if !ok {
			continue
		}
		fieldObj, ok := rowMap["googleAdsField"].(map[string]any)
		if !ok {
			fieldObj = rowMap
		}
		name, _ := fieldObj["name"].(string)
		rows = append(rows, fieldServiceRow{
			name:       name,
			selectable: boolValue(fieldObj["selectable"]),
			filterable: boolValue(fieldObj["filterable"]),
			sortable:   boolValue(fieldObj["sortable"]),
		})
	}
	return rows, nextToken, nil
}

func boolValue(raw any) bool {
	v, ok := raw.(bool)
	return ok && v
}

func sortedKeys(set map[string]struct{}) []string {
	names := make([]string, 0, len(set))
	for name := range set {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
