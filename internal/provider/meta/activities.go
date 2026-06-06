package meta

import (
	"context"
)

func (s *service) getAccountActivities(ctx context.Context, mcpTool, userID, actID string, in activitiesInput) (any, error) {
	normalized, err := normalizeActID(actID)
	if err != nil {
		return nil, err
	}
	query := buildActivitiesQuery(in)
	raw, err := s.proxy.getWithRefresh(ctx, mcpTool, userID, normalized+"/activities", query)
	if err != nil {
		return nil, err
	}
	return unmarshalPayload(raw)
}

func (s *service) getAdSetActivities(ctx context.Context, mcpTool, userID, adSetID string, in activitiesInput) (any, error) {
	normalized, err := requireAdSetID(adSetID)
	if err != nil {
		return nil, err
	}
	query := buildActivitiesQuery(in)
	raw, err := s.proxy.getWithRefresh(ctx, mcpTool, userID, normalized+"/activities", query)
	if err != nil {
		return nil, err
	}
	return unmarshalPayload(raw)
}
