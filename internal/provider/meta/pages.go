package meta

import "context"

func (s *service) listAccountPages(ctx context.Context, mcpTool, userID string, in listAccountPagesInput) (any, error) {
	query := buildListQuery(in.listPaginationInput)
	if in.actID != "" {
		normalized, err := normalizeActID(in.actID)
		if err != nil {
			return nil, err
		}
		return s.graphGETPaginated(ctx, mcpTool, userID, normalized+"/promote_pages", query, in.autoPaginate)
	}
	return s.graphGETPaginated(ctx, mcpTool, userID, "me/accounts", query, in.autoPaginate)
}
