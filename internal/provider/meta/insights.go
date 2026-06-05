package meta

import (
	"context"
)

func (s *service) getAdAccountInsights(ctx context.Context, mcpTool, userID, actID string, in insightsInput) (any, error) {
	normalized, err := normalizeActID(actID)
	if err != nil {
		return nil, err
	}
	path := normalized + "/insights"
	query := buildInsightsQuery(in)
	if in.level == "" {
		query["level"] = "account"
	}
	return s.graphGETPaginated(ctx, mcpTool, userID, path, query, in.autoPaginate)
}

func (s *service) getCampaignInsights(ctx context.Context, mcpTool, userID, campaignID string, in insightsInput) (any, error) {
	id, err := requireCampaignID(campaignID)
	if err != nil {
		return nil, err
	}
	path := id + "/insights"
	query := buildInsightsQuery(in)
	if in.level == "" {
		query["level"] = "campaign"
	}
	return s.graphGETPaginated(ctx, mcpTool, userID, path, query, in.autoPaginate)
}

func (s *service) searchAdEntities(ctx context.Context, mcpTool, userID string, in searchAdEntitiesInput) (any, error) {
	normalized, err := normalizeActID(in.actID)
	if err != nil {
		return nil, err
	}
	path := normalized + "/insights"
	query := buildInsightsQuery(in.insightsInput)
	level := in.level
	if level == "" {
		level = "campaign"
	}
	query["level"] = level
	return s.graphGETPaginated(ctx, mcpTool, userID, path, query, in.autoPaginate)
}
