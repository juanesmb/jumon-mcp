package linkedin

import (
	"strings"
	"testing"
)

func TestBuildLinkedInSearchQuery_statusFilter(t *testing.T) {
	t.Parallel()

	query := buildLinkedInSearchQuery(linkedInSearchQuery{
		statusFilter: []string{"ACTIVE"},
		pageSize:     100,
	})
	if !strings.Contains(query["search"], "status:(values:List(ACTIVE))") {
		t.Fatalf("search = %q", query["search"])
	}
	if query["pageSize"] != "100" {
		t.Fatalf("pageSize = %q", query["pageSize"])
	}
}

func TestBuildLinkedInSearchQuery_pageToken(t *testing.T) {
	t.Parallel()

	query := buildLinkedInSearchQuery(linkedInSearchQuery{
		pageSize:  50,
		pageToken: "abc123",
	})
	if query["pageToken"] != "abc123" {
		t.Fatalf("pageToken = %q", query["pageToken"])
	}
}

func TestBuildLinkedInSearchQuery_testFilterUsesSearchDotTest(t *testing.T) {
	t.Parallel()

	testFilter := false
	query := buildLinkedInSearchQuery(linkedInSearchQuery{
		testFilter: &testFilter,
		pageSize:   100,
	})

	if query["search.test"] != "false" {
		t.Fatalf("search.test = %q, want false", query["search.test"])
	}
	if strings.Contains(query["search"], "test:") {
		t.Fatalf("test filter should not be embedded in search tuple: %q", query["search"])
	}
}

func TestResolveAutoPaginate(t *testing.T) {
	t.Parallel()

	if !resolveAutoPaginate(false, "", true) {
		t.Fatal("forceAllPages should enable auto pagination")
	}
	if resolveAutoPaginate(true, "token", false) {
		t.Fatal("page token should disable auto pagination")
	}
	if !resolveAutoPaginate(true, "", false) {
		t.Fatal("requested auto pagination should be preserved")
	}
}

func TestParseGetCampaignGroupsInput_requiresAccountID(t *testing.T) {
	t.Parallel()

	_, err := parseGetCampaignGroupsInput(map[string]any{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestParseGetCampaignGroupsInput_autoPaginateDefaultsTrue(t *testing.T) {
	t.Parallel()

	in, err := parseGetCampaignGroupsInput(map[string]any{"account_id": "999"})
	if err != nil {
		t.Fatalf("parseGetCampaignGroupsInput() error = %v", err)
	}
	if !in.autoPaginate {
		t.Fatal("autoPaginate should default to true")
	}
	if in.pageSize != defaultCampaignsPageSize {
		t.Fatalf("pageSize = %d, want %d", in.pageSize, defaultCampaignsPageSize)
	}
}

func TestFilterCampaignsByGroup_matchesCampaignGroupURN(t *testing.T) {
	t.Parallel()

	raw := map[string]any{
		"elements": []any{
			map[string]any{
				"id":            "1",
				"campaignGroup": "urn:li:sponsoredCampaignGroup:665256903",
			},
			map[string]any{
				"id":            "2",
				"campaignGroup": "urn:li:sponsoredCampaignGroup:999999999",
			},
		},
	}

	filtered := filterCampaignsByGroup(raw, []string{"665256903"}).(map[string]any)
	elements := filtered["elements"].([]any)
	if len(elements) != 1 {
		t.Fatalf("len(elements) = %d, want 1", len(elements))
	}
	row := elements[0].(map[string]any)
	if row["id"] != "1" {
		t.Fatalf("unexpected campaign id %#v", row["id"])
	}
}

func TestFilterCampaignsByGroup_noFilterReturnsAll(t *testing.T) {
	t.Parallel()

	raw := map[string]any{
		"elements": []any{
			map[string]any{"id": "1"},
			map[string]any{"id": "2"},
		},
	}

	filtered := filterCampaignsByGroup(raw, nil).(map[string]any)
	if len(filtered["elements"].([]any)) != 2 {
		t.Fatal("expected all elements when no group filter")
	}
}
