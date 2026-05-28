package googleads

import "context"

func (s *service) runPaginatedReportSearch(
	ctx context.Context,
	userID, mcpTool string,
	in reportFilters,
	gaqlResource, query string,
) (any, error) {
	annotateGoogleSpan(ctx, in.customerID, gaqlResource, mcpTool)
	raw, err := s.googleSearchPaginated(ctx, userID, mcpTool, in.customerID, in.loginCustomerID, query, in.autoPaginate)
	if err != nil {
		logGoogleSearchFailure(ctx, mcpTool, in.customerID, gaqlResource, query, err)
		return nil, err
	}
	return raw, nil
}

func (s *service) runWrappedReportSearch(
	ctx context.Context,
	userID, mcpTool string,
	in reportFilters,
	gaqlResource string,
	kind emptyResultKind,
	query string,
) (any, error) {
	annotateGoogleSpan(ctx, in.customerID, gaqlResource, mcpTool)
	raw, err := s.googleSearchPaginated(ctx, userID, mcpTool, in.customerID, in.loginCustomerID, query, in.autoPaginate)
	if err != nil {
		logGoogleSearchFailure(ctx, mcpTool, in.customerID, gaqlResource, query, err)
		return nil, err
	}
	return s.wrapReportSearchResponse(ctx, userID, mcpTool, kind, in.customerContext, raw)
}
