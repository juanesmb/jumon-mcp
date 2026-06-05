package meta

const (
	defaultListLimit       = 25
	maxListLimit           = 100
	maxInsightsLimit       = 1000
	defaultDatePreset      = "last_30d"
	maxAutoPaginatePages   = 10
	maxDeliveryEntityIDs   = 50
)

var defaultAdAccountFields = []string{
	"name",
	"business_name",
	"account_status",
	"balance",
	"amount_spent",
	"currency",
	"account_id",
	"id",
}

var defaultAdAccountsListFields = []string{
	"id",
	"name",
	"account_id",
	"account_status",
	"currency",
}

var defaultCampaignListFields = []string{
	"id",
	"name",
	"objective",
	"effective_status",
	"status",
	"daily_budget",
	"lifetime_budget",
	"created_time",
}

var defaultAdSetListFields = []string{
	"id",
	"name",
	"campaign_id",
	"effective_status",
	"status",
	"daily_budget",
	"lifetime_budget",
	"optimization_goal",
}

var defaultAdListFields = []string{
	"id",
	"name",
	"adset_id",
	"campaign_id",
	"effective_status",
	"status",
	"creative",
}

var defaultInsightsFields = []string{
	"impressions",
	"reach",
	"clicks",
	"spend",
	"ctr",
	"cpc",
	"cpm",
	"frequency",
	"date_start",
	"date_stop",
}

var defaultSearchEntitiesFields = []string{
	"campaign_id",
	"campaign_name",
	"impressions",
	"reach",
	"clicks",
	"spend",
	"ctr",
	"cpc",
	"cpm",
	"date_start",
	"date_stop",
}

var defaultAccountPageFields = []string{
	"id",
	"name",
	"username",
	"leadgen_tos_accepted",
}

var deliveryErrorBaseFields = []string{
	"name",
	"effective_status",
}

var deliveryErrorAdFields = []string{
	"name",
	"effective_status",
	"failed_delivery_checks",
}

var deliveryErrorStructureFields = []string{
	"name",
	"effective_status",
	"issues_info",
}
