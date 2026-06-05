package meta

import (
	"context"
	"strings"
)

func (s *service) listCampaigns(ctx context.Context, mcpTool, userID, actID string, in listCampaignsInput) (any, error) {
	normalized, err := normalizeActID(actID)
	if err != nil {
		return nil, err
	}
	path := normalized + "/campaigns"
	query := buildListQuery(in.listPaginationInput)
	return s.graphGETPaginated(ctx, mcpTool, userID, path, query, in.autoPaginate)
}

func (s *service) getCampaign(ctx context.Context, mcpTool, userID string, in getCampaignInput) (any, error) {
	campaignID, err := requireCampaignID(in.campaignID)
	if err != nil {
		return nil, err
	}
	return s.getNode(ctx, mcpTool, userID, campaignID, in.fields, defaultCampaignListFields)
}

func (s *service) getAdSet(ctx context.Context, mcpTool, userID string, in getAdSetInput) (any, error) {
	adSetID, err := requireAdSetID(in.adSetID)
	if err != nil {
		return nil, err
	}
	return s.getNode(ctx, mcpTool, userID, adSetID, in.fields, defaultAdSetListFields)
}

func (s *service) getAd(ctx context.Context, mcpTool, userID string, in getAdInput) (any, error) {
	adID, err := requireAdID(in.adID)
	if err != nil {
		return nil, err
	}
	return s.getNode(ctx, mcpTool, userID, adID, in.fields, defaultAdListFields)
}

func (s *service) getNode(ctx context.Context, mcpTool, userID, nodeID string, fields, defaultFields []string) (any, error) {
	if len(fields) == 0 {
		fields = append([]string(nil), defaultFields...)
	}
	query := map[string]string{"fields": joinCSV(fields)}
	raw, err := s.proxy.getWithRefresh(ctx, mcpTool, userID, nodeID, query)
	if err != nil {
		return nil, err
	}
	return unmarshalPayload(raw)
}

func (s *service) listAdSets(ctx context.Context, mcpTool, userID, actID string, in listAdSetsInput) (any, error) {
	normalized, err := normalizeActID(actID)
	if err != nil {
		return nil, err
	}
	path := normalized + "/adsets"
	query := buildListQuery(in.listPaginationInput)
	if campaignID := strings.TrimSpace(in.campaignID); campaignID != "" {
		query["filtering"] = jsonEncode([]map[string]any{
			{"field": "campaign.id", "operator": "EQUAL", "value": campaignID},
		})
	}
	return s.graphGETPaginated(ctx, mcpTool, userID, path, query, in.autoPaginate)
}

func (s *service) listAds(ctx context.Context, mcpTool, userID, actID string, in listAdsInput) (any, error) {
	normalized, err := normalizeActID(actID)
	if err != nil {
		return nil, err
	}
	path := normalized + "/ads"
	query := buildListQuery(in.listPaginationInput)
	var filters []map[string]any
	if campaignID := strings.TrimSpace(in.campaignID); campaignID != "" {
		filters = append(filters, map[string]any{"field": "campaign.id", "operator": "EQUAL", "value": campaignID})
	}
	if adSetID := strings.TrimSpace(in.adSetID); adSetID != "" {
		filters = append(filters, map[string]any{"field": "adset.id", "operator": "EQUAL", "value": adSetID})
	}
	if len(filters) > 0 {
		query["filtering"] = jsonEncode(filters)
	}
	return s.graphGETPaginated(ctx, mcpTool, userID, path, query, in.autoPaginate)
}
