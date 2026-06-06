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

var deliveryErrorBatchFields = []string{
	"name",
	"effective_status",
	"failed_delivery_checks",
	"issues_info",
}

var defaultCreativeListFields = []string{
	"id",
	"name",
	"status",
	"object_type",
	"thumbnail_url",
	"title",
	"body",
	"call_to_action_type",
}

var defaultCreativeFields = []string{
	"id",
	"name",
	"account_id",
	"status",
	"object_type",
	"object_story_spec",
	"asset_feed_spec",
	"thumbnail_url",
	"title",
	"body",
	"link_url",
	"call_to_action_type",
	"image_hash",
	"video_id",
}

var defaultAdImageFields = []string{
	"id",
	"hash",
	"name",
	"url",
	"width",
	"height",
	"status",
	"created_time",
}

var defaultAdVideoFields = []string{
	"id",
	"title",
	"source",
	"picture",
	"length",
	"created_time",
}

var defaultCustomAudienceListFields = []string{
	"id",
	"name",
	"subtype",
	"approximate_count_lower_bound",
	"approximate_count_upper_bound",
	"delivery_status",
	"time_created",
}

var defaultCustomAudienceFields = []string{
	"id",
	"name",
	"description",
	"subtype",
	"approximate_count_lower_bound",
	"approximate_count_upper_bound",
	"delivery_status",
	"operation_status",
	"time_created",
	"time_updated",
	"rule",
	"customer_file_source",
}

var defaultCustomConversionListFields = []string{
	"id",
	"name",
	"custom_event_type",
	"pixel",
	"rule",
	"default_conversion_value",
	"is_archived",
}

var defaultDatasetListFields = []string{
	"id",
	"name",
	"creation_time",
	"last_fired_time",
	"is_unavailable",
}

var defaultDatasetFields = []string{
	"id",
	"name",
	"creation_time",
	"last_fired_time",
	"is_unavailable",
	"owner_business",
	"code",
}

var defaultActivityFields = []string{
	"actor_name",
	"object_type",
	"translated_event_type",
	"event_time",
	"changed_data",
	"object_id",
	"object_name",
}

var defaultCreativeAdListFields = []string{
	"id",
	"name",
	"status",
	"effective_status",
	"adset_id",
	"campaign_id",
}

var defaultDatasetQualityFields = "web{event_match_quality,event_name}"

var validDemographicClasses = []string{
	"demographics",
	"life_events",
	"industries",
	"income",
	"family_statuses",
	"user_device",
	"user_os",
}
