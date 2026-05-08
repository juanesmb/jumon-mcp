package reddit

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

const (
	maxPageSizeBusinesses      = 700
	defaultPageSizeBusinesses  = 700
	maxRedditPagedListSize     = 1000
	defaultRedditPagedListSize = 100
)

var hourlyUTCReportTimestamp = regexp.MustCompile(`^[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:00:00Z$`)

type service struct {
	proxy redditUpstreamPort
}

func newService(proxy redditUpstreamPort) *service {
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

type listCampaignsInput struct {
	pageSize  int
	pageToken string
}

type createReportInput struct {
	pageSize        int
	pageToken       string
	fields          []string
	breakdowns      []string
	customColumnIDs []string
	filter          string
	startsAt        string
	endsAt          string
	timeZoneID      string
}

type redditReportBody struct {
	Data redditReportData `json:"data"`
}

type redditReportData struct {
	Breakdowns      []string `json:"breakdowns,omitempty"`
	CustomColumnIDs []string `json:"custom_column_ids,omitempty"`
	Fields          []string `json:"fields"`
	Filter          string   `json:"filter,omitempty"`
	StartsAt        string   `json:"starts_at"`
	EndsAt          string   `json:"ends_at"`
	TimeZoneID      string   `json:"time_zone_id,omitempty"`
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

func (s *service) listCampaignsForAdAccount(ctx context.Context, userID, adAccountID string, in listCampaignsInput) (json.RawMessage, error) {
	id := strings.TrimSpace(adAccountID)
	if id == "" {
		return nil, fmt.Errorf("reddit: ad_account_id is required; use reddit_list_ad_accounts after reddit_list_businesses")
	}
	path := pathAdAccountCampaigns(id)
	query := buildCampaignsQuery(in)
	return s.redditGET(ctx, userID, path, query, in.pageToken)
}

func (s *service) createReportForAdAccount(ctx context.Context, userID, adAccountID string, in createReportInput) (json.RawMessage, error) {
	id := strings.TrimSpace(adAccountID)
	if id == "" {
		return nil, fmt.Errorf("reddit: ad_account_id is required; use reddit_list_ad_accounts after reddit_list_businesses")
	}
	fields := nonemptyStrings(in.fields)
	if len(fields) == 0 {
		return nil, fmt.Errorf("reddit: fields must include at least one metric field name")
	}
	if err := validateRedditHourlyUTC("starts_at", in.startsAt); err != nil {
		return nil, err
	}
	if err := validateRedditHourlyUTC("ends_at", in.endsAt); err != nil {
		return nil, err
	}

	body := redditReportBody{
		Data: redditReportData{
			Breakdowns:      nonemptyStrings(in.breakdowns),
			CustomColumnIDs: nonemptyStrings(in.customColumnIDs),
			Fields:          fields,
			Filter:          strings.TrimSpace(in.filter),
			StartsAt:        strings.TrimSpace(in.startsAt),
			EndsAt:          strings.TrimSpace(in.endsAt),
			TimeZoneID:      strings.TrimSpace(in.timeZoneID),
		},
	}
	query := buildReportQuery(in)
	path := pathAdAccountReports(id)
	return s.redditPOST(ctx, userID, path, query, body, in.pageToken)
}

func validateRedditHourlyUTC(field, value string) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return fmt.Errorf("reddit: %s is required", field)
	}
	if !hourlyUTCReportTimestamp.MatchString(value) {
		return fmt.Errorf("reddit: %s must be hourly UTC RFC3339, e.g. 2023-11-20T03:00:00Z", field)
	}
	return nil
}

func nonemptyStrings(slice []string) []string {
	if len(slice) == 0 {
		return nil
	}
	out := make([]string, 0, len(slice))
	for _, s := range slice {
		if t := strings.TrimSpace(s); t != "" {
			out = append(out, t)
		}
	}
	return out
}

// redditGET uses a plain gateway GET when pageToken is set (pagination follow-up), otherwise refresh-on-unauthorized.
func (s *service) redditGET(ctx context.Context, userID, apiPath string, query map[string]string, pageToken string) (json.RawMessage, error) {
	if strings.TrimSpace(pageToken) != "" {
		return s.proxy.get(ctx, userID, apiPath, query)
	}
	return s.proxy.getWithRefresh(ctx, userID, apiPath, query)
}

// redditPOST uses a plain gateway POST when pageToken is set (pagination follow-up), otherwise refresh-on-unauthorized.
func (s *service) redditPOST(ctx context.Context, userID, apiPath string, query map[string]string, body any, pageToken string) (json.RawMessage, error) {
	if strings.TrimSpace(pageToken) != "" {
		return s.proxy.postJSON(ctx, userID, apiPath, query, body)
	}
	return s.proxy.postJSONWithRefresh(ctx, userID, apiPath, query, body)
}

func buildBusinessesQuery(in listBusinessesInput) map[string]string {
	query := map[string]string{}
	if v := strings.TrimSpace(in.adAccountID); v != "" {
		query[queryKeyAdAccountFilter] = v
	}
	if v := strings.TrimSpace(in.role); v != "" {
		query[queryKeyRole] = v
	}
	appendRedditPage(query, in.pageSize, in.pageToken, defaultPageSizeBusinesses, maxPageSizeBusinesses)
	return query
}

func buildAdAccountsQuery(in listAdAccountsInput) map[string]string {
	query := map[string]string{}
	appendRedditPage(query, in.pageSize, in.pageToken, defaultRedditPagedListSize, maxRedditPagedListSize)
	return query
}

func buildCampaignsQuery(in listCampaignsInput) map[string]string {
	query := map[string]string{}
	appendRedditPage(query, in.pageSize, in.pageToken, defaultRedditPagedListSize, maxRedditPagedListSize)
	return query
}

func buildReportQuery(in createReportInput) map[string]string {
	query := map[string]string{}
	appendRedditPage(query, in.pageSize, in.pageToken, defaultRedditPagedListSize, maxRedditPagedListSize)
	return query
}

func appendRedditPage(query map[string]string, pageSize int, pageToken string, defaultSize, maxSize int) {
	if pageSize <= 0 {
		pageSize = defaultSize
	}
	if pageSize > maxSize {
		pageSize = maxSize
	}
	query[queryKeyPageSize] = strconv.Itoa(pageSize)
	if pt := strings.TrimSpace(pageToken); pt != "" {
		query[queryKeyPageToken] = pt
	}
}
