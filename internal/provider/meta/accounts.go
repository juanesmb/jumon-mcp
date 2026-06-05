package meta

import (
	"context"
	"fmt"
)

func (s *service) listAdAccounts(ctx context.Context, mcpTool, userID string) (any, error) {
	query := map[string]string{
		"fields": fmt.Sprintf("adaccounts{%s}", joinCSV(defaultAdAccountsListFields)),
	}
	raw, err := s.proxy.getWithRefresh(ctx, mcpTool, userID, "me", query)
	if err != nil {
		return nil, err
	}
	return unmarshalPayload(raw)
}

func (s *service) getAdAccount(ctx context.Context, mcpTool, userID string, in getAdAccountInput) (any, error) {
	actID, err := normalizeActID(in.actID)
	if err != nil {
		return nil, err
	}
	fields := in.fields
	if len(fields) == 0 {
		fields = append([]string(nil), defaultAdAccountFields...)
	}
	query := map[string]string{"fields": joinCSV(fields)}
	raw, err := s.proxy.getWithRefresh(ctx, mcpTool, userID, actID, query)
	if err != nil {
		return nil, err
	}
	return unmarshalPayload(raw)
}
