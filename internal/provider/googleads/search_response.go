package googleads

import (
	"context"
	"fmt"
	"strings"
)

const channelSniffQuery = "SELECT campaign.advertising_channel_type, campaign.name FROM campaign WHERE campaign.status != 'REMOVED' LIMIT 20"

type emptyResultKind int

const (
	emptyResultKeywords emptyResultKind = iota
	emptyResultSearchTerms
	emptyResultPmaxSearchTerms
)

func (s *service) wrapReportSearchResponse(
	ctx context.Context,
	userID, mcpTool string,
	kind emptyResultKind,
	in customerContext,
	raw any,
) (any, error) {
	if !isEmptySearchResults(raw) {
		return raw, nil
	}

	out := cloneSearchRoot(raw)
	hint := emptyResultHint(kind)
	out["hint"] = hint

	summary, err := s.summarizeCampaignChannels(ctx, userID, mcpTool, in.customerID, in.loginCustomerID)
	if err == nil && len(summary) > 0 {
		out["channel_summary"] = summary
		out["hint"] = hint + " Active campaign channels: " + formatChannelSummary(summary) + "."
	}
	return out, nil
}

func isEmptySearchResults(raw any) bool {
	return len(extractSearchRows(raw)) == 0
}

func cloneSearchRoot(raw any) map[string]any {
	root, ok := raw.(map[string]any)
	if !ok {
		return map[string]any{"results": []any{}}
	}
	out := make(map[string]any, len(root)+2)
	for k, v := range root {
		out[k] = v
	}
	if _, ok := out["results"]; !ok {
		out["results"] = []any{}
	}
	return out
}

func emptyResultHint(kind emptyResultKind) string {
	switch kind {
	case emptyResultKeywords:
		return "No keyword_view rows returned. keyword_view only includes Search campaigns with active keywords. If this account runs Demand Gen, Video, Display, or Performance Max only, try google_resolve_customer for a Search client account or use google_search_pmax_search_terms for PMax search terms."
	case emptyResultSearchTerms:
		return "No search_term_view rows returned. search_term_view only includes Search campaigns. If this account has no Search inventory, try another client under the MCC or use google_search_pmax_search_terms for Performance Max search terms (campaign_search_term_view)."
	case emptyResultPmaxSearchTerms:
		return "No campaign_search_term_view rows returned. This view covers Performance Max search terms only. If the account has no PMax campaigns or PMax search term reporting is unavailable, confirm campaign type with google_search_campaigns or broaden the date range."
	default:
		return "No rows returned for this report."
	}
}

func (s *service) summarizeCampaignChannels(
	ctx context.Context,
	userID, mcpTool, customerID, loginCustomerID string,
) (map[string]int, error) {
	raw, err := s.googleSearch(ctx, userID, mcpTool, customerID, loginCustomerID, channelSniffQuery)
	if err != nil {
		return nil, err
	}
	counts := make(map[string]int)
	for _, row := range extractSearchRows(raw) {
		campaign := nestedMap(row, "campaign")
		channelType := stringFromMap(campaign, "advertisingChannelType")
		if channelType == "" {
			channelType = "UNKNOWN"
		}
		counts[channelType]++
	}
	return counts, nil
}

func formatChannelSummary(counts map[string]int) string {
	if len(counts) == 0 {
		return "none detected (no active campaigns or account may be empty)"
	}
	parts := make([]string, 0, len(counts))
	for channel, count := range counts {
		label := strings.ReplaceAll(channel, "_", " ")
		parts = append(parts, fmt.Sprintf("%s (%d)", label, count))
	}
	return strings.Join(parts, ", ")
}
