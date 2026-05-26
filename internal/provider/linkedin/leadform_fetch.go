package linkedin

import (
	"context"
	"fmt"
	"strings"
)

type leadFormSummary struct {
	name string
}

// fetchLeadFormsByIDs batch-fetches lead form metadata using GET /rest/leadForms?ids=List(...).
// Returns a map of formID → leadFormSummary. Fetch failures are silently ignored so the
// caller never fails the parent creative search.
func fetchLeadFormsByIDs(
	ctx context.Context,
	proxy linkedinUpstreamPort,
	userID, mcpTool string,
	ids []string,
) map[string]leadFormSummary {
	if len(ids) == 0 {
		return nil
	}
	deduped := uniqueStrings(ids)
	query := map[string]string{
		"ids": fmt.Sprintf("List(%s)", strings.Join(deduped, ",")),
	}
	raw, err := proxy.requestJSON(ctx, userID, mcpTool, "GET", "leadForms", query, nil, nil)
	if err != nil {
		return nil
	}
	return parseLeadFormBatchResponse(raw)
}

func parseLeadFormBatchResponse(raw any) map[string]leadFormSummary {
	root, ok := raw.(map[string]any)
	if !ok {
		return nil
	}
	results := make(map[string]leadFormSummary)

	// LinkedIn batch-GET returns a "results" map keyed by ID string when using ids=List(...).
	if resultsMap, ok := root["results"].(map[string]any); ok {
		for id, item := range resultsMap {
			form, ok := item.(map[string]any)
			if !ok {
				continue
			}
			if name := extractLeadFormName(form); name != "" {
				results[id] = leadFormSummary{name: name}
			}
		}
		return results
	}

	// Fall back to "elements" list format used by some LinkedIn API versions.
	elements, ok := root["elements"].([]any)
	if !ok {
		return nil
	}
	for _, item := range elements {
		form, ok := item.(map[string]any)
		if !ok {
			continue
		}
		id := extractLeadFormResponseID(form)
		if id == "" {
			continue
		}
		if name := extractLeadFormName(form); name != "" {
			results[id] = leadFormSummary{name: name}
		}
	}
	return results
}

// extractLeadFormName reads the form name from either a plain string or a MultiLocaleString
// as returned by the LinkedIn Marketing API.
func extractLeadFormName(form map[string]any) string {
	switch v := form["name"].(type) {
	case string:
		return strings.TrimSpace(v)
	case map[string]any:
		// MultiLocaleString: {"localized": {"en_US": "Form Name"}, "preferredLocale": {...}}
		if localized, ok := v["localized"].(map[string]any); ok {
			for _, val := range localized {
				if s, ok := val.(string); ok {
					if trimmed := strings.TrimSpace(s); trimmed != "" {
						return trimmed
					}
				}
			}
		}
	}
	return ""
}

func extractLeadFormResponseID(form map[string]any) string {
	switch v := form["id"].(type) {
	case string:
		return strings.TrimSpace(v)
	case float64:
		return fmt.Sprintf("%d", int64(v))
	}
	return ""
}

func uniqueStrings(ss []string) []string {
	seen := make(map[string]struct{}, len(ss))
	out := make([]string, 0, len(ss))
	for _, s := range ss {
		if _, ok := seen[s]; !ok {
			seen[s] = struct{}{}
			out = append(out, s)
		}
	}
	return out
}
