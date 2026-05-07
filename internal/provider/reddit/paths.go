package reddit

import (
	"fmt"
	"net/url"
)

const (
	pathMeBusinesses        = "me/businesses"
	queryKeyPageSize        = "page.size"
	queryKeyPageToken       = "page.token"
	queryKeyAdAccountFilter = "ad_account_id"
	queryKeyRole            = "role"
)

func pathBusinessAdAccounts(businessID string) string {
	return fmt.Sprintf("businesses/%s/ad_accounts", url.PathEscape(businessID))
}
