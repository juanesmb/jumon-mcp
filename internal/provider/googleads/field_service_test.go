package googleads

import "testing"

func TestParseFieldServiceResponse(t *testing.T) {
	payload := map[string]any{
		"results": []any{
			map[string]any{
				"googleAdsField": map[string]any{
					"name":       "campaign.id",
					"selectable": true,
					"filterable": true,
					"sortable":   false,
				},
			},
			map[string]any{
				"googleAdsField": map[string]any{
					"name":       "metrics.clicks",
					"selectable": true,
					"filterable": false,
					"sortable":   true,
				},
			},
		},
		"nextPageToken": "token-2",
	}

	rows, next, err := parseFieldServiceResponse(payload)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("rows = %d", len(rows))
	}
	if next != "token-2" {
		t.Fatalf("next = %q", next)
	}
	if !rows[0].selectable || !rows[0].filterable || rows[0].sortable {
		t.Fatalf("unexpected flags on row 0: %+v", rows[0])
	}
}

func TestParseResourceMetadataInput(t *testing.T) {
	name, err := parseResourceMetadataInput(map[string]any{"resource_name": "keyword_view"})
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if name != "keyword_view" {
		t.Fatalf("resource = %q", name)
	}
}

func TestParseGAQLSearchInput(t *testing.T) {
	in, err := parseGAQLSearchInput(map[string]any{
		"customer_id": "1234567890",
		"resource":    "campaign",
		"fields":      []any{"campaign.id", "metrics.clicks"},
		"conditions":  []any{"campaign.status = ENABLED"},
		"limit":       float64(250),
	})
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if in.customerID != "1234567890" || in.resource != "campaign" {
		t.Fatalf("unexpected input: %+v", in)
	}
	if len(in.fields) != 2 || in.limit != 250 {
		t.Fatalf("fields/limit: %+v", in)
	}
}
