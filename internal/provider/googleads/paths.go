package googleads

import "fmt"

func pathListAccessibleCustomers(version string) string {
	return fmt.Sprintf("%s/customers:listAccessibleCustomers", version)
}

func pathGoogleAdsSearch(version, customerID string) string {
	return fmt.Sprintf("%s/customers/%s/googleAds:search", version, customerID)
}

