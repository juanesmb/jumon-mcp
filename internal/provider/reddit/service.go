package reddit

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

const (
	maxPageSizeBusinesses     = 700
	maxPageSizeAdAccounts     = 1000
	defaultPageSizeBusinesses = 700
	defaultPageSizeAdAccounts = 100
)

type service struct {
	proxy redditGETPort
}

func newService(proxy redditGETPort) *service {
	return &service{proxy: proxy}
}

type listBusinessesInput struct {
	adAccountID string
	role        string
	pageSize    int
	pageToken   string
}

type listAdAccountsInput struct {
	pageSize  int
	pageToken string
}

func (s *service) listMyBusinesses(ctx context.Context, userID string, in listBusinessesInput) (json.RawMessage, error) {
	query := buildBusinessesQuery(in)
	return s.redditGET(ctx, userID, pathMeBusinesses, query, in.pageToken)
}

func (s *service) listAdAccountsByBusiness(ctx context.Context, userID, businessID string, in listAdAccountsInput) (json.RawMessage, error) {
	id := strings.TrimSpace(businessID)
	if id == "" {
		return nil, fmt.Errorf("reddit: business_id is required; call reddit_list_businesses first to obtain a business id")
	}
	path := pathBusinessAdAccounts(id)
	query := buildAdAccountsQuery(in)
	return s.redditGET(ctx, userID, path, query, in.pageToken)
}

// redditGET uses a plain gateway GET when pageToken is set (pagination follow-up), otherwise refresh-on-unauthorized.
func (s *service) redditGET(ctx context.Context, userID, apiPath string, query map[string]string, pageToken string) (json.RawMessage, error) {
	if strings.TrimSpace(pageToken) != "" {
		return s.proxy.get(ctx, userID, apiPath, query)
	}
	return s.proxy.getWithRefresh(ctx, userID, apiPath, query)
}

func buildBusinessesQuery(in listBusinessesInput) map[string]string {
	query := map[string]string{}
	if v := strings.TrimSpace(in.adAccountID); v != "" {
		query[queryKeyAdAccountFilter] = v
	}
	if v := strings.TrimSpace(in.role); v != "" {
		query[queryKeyRole] = v
	}
	pageSize := in.pageSize
	if pageSize <= 0 {
		pageSize = defaultPageSizeBusinesses
	}
	if pageSize > maxPageSizeBusinesses {
		pageSize = maxPageSizeBusinesses
	}
	query[queryKeyPageSize] = strconv.Itoa(pageSize)
	if pt := strings.TrimSpace(in.pageToken); pt != "" {
		query[queryKeyPageToken] = pt
	}
	return query
}

func buildAdAccountsQuery(in listAdAccountsInput) map[string]string {
	query := map[string]string{}
	pageSize := in.pageSize
	if pageSize <= 0 {
		pageSize = defaultPageSizeAdAccounts
	}
	if pageSize > maxPageSizeAdAccounts {
		pageSize = maxPageSizeAdAccounts
	}
	query[queryKeyPageSize] = strconv.Itoa(pageSize)
	if pt := strings.TrimSpace(in.pageToken); pt != "" {
		query[queryKeyPageToken] = pt
	}
	return query
}
