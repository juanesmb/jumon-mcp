package googleads

import "context"

func (s *service) searchKeywords(ctx context.Context, userID, mcpTool string, in reportFilters) (any, error) {
	return s.googleSearch(ctx, userID, mcpTool, in.customerID, in.loginCustomerID, buildKeywordsQuery(in))
}

func (s *service) searchSearchTerms(ctx context.Context, userID, mcpTool string, in reportFilters) (any, error) {
	return s.googleSearch(ctx, userID, mcpTool, in.customerID, in.loginCustomerID, buildSearchTermsQuery(in))
}

func (s *service) listConversionActions(ctx context.Context, userID, mcpTool string, in reportFilters) (any, error) {
	return s.googleSearch(ctx, userID, mcpTool, in.customerID, in.loginCustomerID, buildConversionActionsQuery(in))
}

func (s *service) searchConversionPerformance(ctx context.Context, userID, mcpTool string, in reportFilters) (any, error) {
	return s.googleSearch(ctx, userID, mcpTool, in.customerID, in.loginCustomerID, buildConversionPerformanceQuery(in))
}
