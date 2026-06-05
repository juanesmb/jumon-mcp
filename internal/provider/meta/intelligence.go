package meta

import (
	"context"
)

func (s *service) getOpportunityScore(ctx context.Context, mcpTool, userID, actID string) (any, error) {
	normalized, err := normalizeActID(actID)
	if err != nil {
		return nil, err
	}
	raw, err := s.proxy.getWithRefresh(ctx, mcpTool, userID, normalized+"/recommendations", nil)
	if err != nil {
		return nil, err
	}
	return unmarshalPayload(raw)
}
