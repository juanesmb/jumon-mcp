package googleads

import (
	_ "embed"
	"fmt"
	"strings"
)

//go:embed gaql_resources.txt
var gaqlResourcesFile string

const maxChangeEventLimit = 10000

var gaqlResourceAllowlist = loadGAQLResourceAllowlist()

func loadGAQLResourceAllowlist() map[string]struct{} {
	out := make(map[string]struct{})
	for _, line := range strings.Split(gaqlResourcesFile, "\n") {
		name := strings.TrimSpace(line)
		if name == "" || strings.HasPrefix(name, "#") {
			continue
		}
		out[name] = struct{}{}
	}
	return out
}

func isAllowedGAQLResource(resource string) bool {
	_, ok := gaqlResourceAllowlist[resource]
	return ok
}

func commonGAQLResources() []string {
	return []string{
		"campaign",
		"ad_group",
		"ad_group_ad",
		"keyword_view",
		"search_term_view",
		"campaign_search_term_view",
		"conversion_action",
		"customer",
		"customer_client",
		"change_event",
		"shopping_performance_view",
		"asset_group",
		"offline_conversion_upload_conversion_action_summary",
	}
}

func normalizeGAQLResourceName(raw string) (string, error) {
	name := strings.TrimSpace(raw)
	if name == "" {
		return "", fmt.Errorf("resource_name is required")
	}
	if strings.ContainsAny(name, " \t\n\r'\"%;") {
		return "", fmt.Errorf("resource_name contains invalid characters")
	}
	if !isAllowedGAQLResource(name) {
		return "", fmt.Errorf("resource %q is not in the GAQL resource allowlist; see docs/google-ads-tools.md", name)
	}
	return name, nil
}

type gaqlSearchInput struct {
	customerContext
	resource     string
	fields       []string
	conditions   []string
	orderings    []string
	limit        int
	autoPaginate bool
}

func validateGAQLSearchInput(in gaqlSearchInput) (string, error) {
	resource, err := normalizeGAQLResourceName(in.resource)
	if err != nil {
		return "", err
	}
	if len(in.fields) == 0 {
		return "", fmt.Errorf("fields is required and must be non-empty")
	}
	for _, field := range in.fields {
		if err := validateGAQLFieldName(field, resource); err != nil {
			return "", err
		}
	}
	for _, condition := range in.conditions {
		if err := validateGAQLFragment(condition, "condition"); err != nil {
			return "", err
		}
	}
	for _, ordering := range in.orderings {
		if err := validateGAQLFragment(ordering, "ordering"); err != nil {
			return "", err
		}
	}
	return resource, nil
}

func validateGAQLFieldName(field, resource string) error {
	trimmed := strings.TrimSpace(field)
	if trimmed == "" {
		return fmt.Errorf("fields must not contain empty strings")
	}
	if strings.Contains(trimmed, "*") {
		return fmt.Errorf("field %q must not contain wildcards", field)
	}
	if strings.Contains(trimmed, ";") {
		return fmt.Errorf("field %q must not contain semicolons", field)
	}
	if trimmed == "id" || !strings.Contains(trimmed, ".") {
		return fmt.Errorf("field %q must be fully qualified (e.g. %s.id, metrics.clicks)", field, resource)
	}
	prefix := strings.SplitN(trimmed, ".", 2)[0]
	if prefix == resource || prefix == "metrics" || prefix == "segments" {
		return nil
	}
	if isAllowedGAQLAttributedPrefix(prefix) {
		return nil
	}
	return fmt.Errorf("field %q is not allowed by local guard; call google_get_resource_metadata and use only listed selectable fields", field)
}

var gaqlAttributedPrefixes = map[string]struct{}{
	"campaign":    {},
	"ad_group":    {},
	"customer":    {},
	"ad_group_ad": {},
	"asset_group": {},
}

func isAllowedGAQLAttributedPrefix(prefix string) bool {
	_, ok := gaqlAttributedPrefixes[prefix]
	return ok
}

func validateGAQLFragment(value, label string) error {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return fmt.Errorf("%s must not contain empty strings", label)
	}
	if strings.Contains(trimmed, ";") {
		return fmt.Errorf("%s must not contain semicolons", label)
	}
	return nil
}

func normalizeGenericSearchLimit(resource string, limit int) int {
	if limit <= 0 {
		return defaultReportLimit
	}
	if resource == "change_event" && limit > maxChangeEventLimit {
		return maxChangeEventLimit
	}
	if limit > maxReportLimit {
		return maxReportLimit
	}
	return limit
}

func buildGenericSearchQuery(in gaqlSearchInput, resource string) string {
	parts := []string{
		fmt.Sprintf("SELECT %s FROM %s", strings.Join(in.fields, ","), resource),
	}
	if len(in.conditions) > 0 {
		clean := make([]string, 0, len(in.conditions))
		for _, condition := range in.conditions {
			if trimmed := strings.TrimSpace(condition); trimmed != "" {
				clean = append(clean, trimmed)
			}
		}
		if len(clean) > 0 {
			parts = append(parts, " WHERE "+strings.Join(clean, " AND "))
		}
	}
	if len(in.orderings) > 0 {
		clean := make([]string, 0, len(in.orderings))
		for _, ordering := range in.orderings {
			if trimmed := strings.TrimSpace(ordering); trimmed != "" {
				clean = append(clean, trimmed)
			}
		}
		if len(clean) > 0 {
			parts = append(parts, " ORDER BY "+strings.Join(clean, ","))
		}
	}
	limit := normalizeGenericSearchLimit(resource, in.limit)
	parts = append(parts, googleLimitClause(limit))
	parts = append(parts, " PARAMETERS omit_unselected_resource_names=true")
	return strings.Join(parts, "")
}

func metricsDateHint(in gaqlSearchInput) string {
	hasMetricOrSegment := false
	for _, field := range in.fields {
		trimmed := strings.TrimSpace(field)
		if strings.HasPrefix(trimmed, "metrics.") || strings.HasPrefix(trimmed, "segments.") {
			hasMetricOrSegment = true
			break
		}
	}
	if !hasMetricOrSegment {
		return ""
	}
	for _, condition := range in.conditions {
		lower := strings.ToLower(condition)
		if strings.Contains(lower, "segments.date") {
			return ""
		}
	}
	return "Hint: metric/segment queries often require a segments.date filter (YYYY-MM-DD)."
}
