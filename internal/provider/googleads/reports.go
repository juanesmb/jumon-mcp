package googleads

import "context"

func (s *service) searchKeywords(ctx context.Context, userID, mcpTool string, in reportFilters) (any, error) {
	return s.runWrappedReportSearch(ctx, userID, mcpTool, in, "keyword_view", emptyResultKeywords, buildKeywordsQuery(in))
}

func (s *service) searchSearchTerms(ctx context.Context, userID, mcpTool string, in reportFilters) (any, error) {
	return s.runWrappedReportSearch(ctx, userID, mcpTool, in, "search_term_view", emptyResultSearchTerms, buildSearchTermsQuery(in))
}

func (s *service) searchPmaxSearchTerms(ctx context.Context, userID, mcpTool string, in reportFilters) (any, error) {
	return s.runWrappedReportSearch(ctx, userID, mcpTool, in, "campaign_search_term_view", emptyResultPmaxSearchTerms, buildPmaxSearchTermsQuery(in))
}

func (s *service) listConversionActions(ctx context.Context, userID, mcpTool string, in reportFilters) (any, error) {
	return s.runPaginatedReportSearch(ctx, userID, mcpTool, in, "conversion_action", buildConversionActionsQuery(in))
}

func (s *service) searchConversionPerformance(ctx context.Context, userID, mcpTool string, in reportFilters) (any, error) {
	return s.runPaginatedReportSearch(ctx, userID, mcpTool, in, "campaign", buildConversionPerformanceQuery(in))
}

func (s *service) listOfflineConversionUploadSummaries(ctx context.Context, userID, mcpTool string, in reportFilters) (any, error) {
	return s.runPaginatedReportSearch(ctx, userID, mcpTool, in, "offline_conversion_upload_conversion_action_summary", buildOfflineConversionUploadSummariesQuery(in))
}
