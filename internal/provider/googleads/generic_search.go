package googleads

import (
	"context"
)

func (s *service) searchGAQL(ctx context.Context, userID, mcpTool string, in gaqlSearchInput) (any, error) {
	resource, err := validateGAQLSearchInput(in)
	if err != nil {
		return nil, err
	}
	in.resource = resource

	query := buildGenericSearchQuery(in, resource)
	result, err := s.googleSearch(ctx, userID, mcpTool, in.customerID, in.loginCustomerID, query)
	if err != nil {
		return nil, err
	}

	out := map[string]any{
		"data": result,
		"_debug": map[string]any{
			"query": query,
		},
	}
	if hint := metricsDateHint(in); hint != "" {
		out["hint"] = hint
	}
	return out, nil
}
