package meta

import (
	"encoding/json"
	"testing"
)

func TestBuildDeliveryRelativeURL(t *testing.T) {
	got := buildDeliveryRelativeURL("123", "name,effective_status")
	if got != "123?fields=name%2Ceffective_status" {
		t.Fatalf("got %q", got)
	}
}

func TestParseBatchItemBody_success(t *testing.T) {
	row, err := parseBatchItemBody(graphBatchResponseItem{
		Code: 200,
		Body: json.RawMessage(`{"id":"1","name":"Ad"}`),
	})
	if err != nil {
		t.Fatal(err)
	}
	if row["id"] != "1" {
		t.Fatalf("row = %v", row)
	}
}

func TestParseBatchItemBody_error(t *testing.T) {
	_, err := parseBatchItemBody(graphBatchResponseItem{Code: 400, Body: json.RawMessage(`{"error":"bad"}`)})
	if err == nil {
		t.Fatal("expected error")
	}
}
