package googleads

import (
	"strings"
	"testing"
)

func TestExtractAccessibleCustomerIDs(t *testing.T) {
	raw := map[string]any{
		"resourceNames": []any{"customers/111", "customers/222"},
	}
	ids := extractAccessibleCustomerIDs(raw)
	if len(ids) != 2 || ids[0] != "111" || ids[1] != "222" {
		t.Fatalf("unexpected ids: %#v", ids)
	}
}

func TestExtractSearchRows(t *testing.T) {
	raw := map[string]any{
		"results": []any{
			map[string]any{"customer": map[string]any{"id": "1", "descriptiveName": "Acme"}},
		},
	}
	rows := extractSearchRows(raw)
	if len(rows) != 1 {
		t.Fatalf("expected 1 row got %d", len(rows))
	}
	customer := nestedMap(rows[0], "customer")
	if stringFromMap(customer, "descriptiveName") != "Acme" {
		t.Fatal("unexpected customer name")
	}
}

func TestBuildClientAccountsQueryWithNameFilter(t *testing.T) {
	query := buildClientAccountsQuery("Retail")
	if query == "" {
		t.Fatal("expected query")
	}
	if !strings.Contains(query, "customer_client.descriptive_name LIKE '%Retail%'") {
		t.Fatalf("unexpected query: %s", query)
	}
	if !strings.Contains(query, "customer_client.manager = false") {
		t.Fatalf("unexpected query: %s", query)
	}
}

func TestBuildClientAccountsResolveQueryExact(t *testing.T) {
	query := buildClientAccountsResolveQuery("Acme", "exact")
	if !strings.Contains(query, "customer_client.descriptive_name = 'Acme'") {
		t.Fatalf("unexpected query: %s", query)
	}
}
