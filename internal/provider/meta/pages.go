package meta

import (
	"context"
)

func (s *service) listAccountPages(ctx context.Context, mcpTool, userID string, in listPaginationInput) (any, error) {
	query := buildListQuery(in)
	return s.graphGETPaginated(ctx, mcpTool, userID, "me/accounts", query, in.autoPaginate)
}
