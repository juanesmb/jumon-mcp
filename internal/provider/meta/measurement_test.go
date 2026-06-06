package meta

import "testing"

func TestBuildDatasetQualityQuery(t *testing.T) {
	q := buildDatasetQualityQuery(datasetQualityInput{datasetID: "123456"})
	if q["dataset_id"] != "123456" {
		t.Fatalf("dataset_id = %q", q["dataset_id"])
	}
	if q["fields"] != defaultDatasetQualityFields {
		t.Fatalf("default fields = %q", q["fields"])
	}

	q = buildDatasetQualityQuery(datasetQualityInput{datasetID: "99", fields: "web{event_name}"})
	if q["fields"] != "web{event_name}" {
		t.Fatalf("custom fields = %q", q["fields"])
	}
}

func TestBuildDatasetStatsQuery(t *testing.T) {
	q := buildDatasetStatsQuery(datasetStatsInput{
		startTime:   "1700000000",
		endTime:     "1700086400",
		eventName:   "Purchase",
		aggregation: "event_total_counts",
	})
	if q["event_name"] != "Purchase" {
		t.Fatalf("event_name = %q", q["event_name"])
	}
	if q["aggregation"] != "event_total_counts" {
		t.Fatalf("aggregation = %q", q["aggregation"])
	}
}

func TestParseDatasetStatsInputRequiresDatasetID(t *testing.T) {
	if _, err := parseDatasetStatsInput(map[string]any{}); err == nil {
		t.Fatal("expected error for missing dataset_id")
	}
	_, err := parseDatasetStatsInput(map[string]any{"dataset_id": "pixel_1"})
	if err != nil {
		t.Fatal(err)
	}
}

func TestParseListCustomConversionsInputDatasetFilter(t *testing.T) {
	in := parseListCustomConversionsInput(map[string]any{
		"act_id":      "123",
		"dataset_id":  "999",
		"auto_paginate": false,
	})
	if in.datasetID != "999" {
		t.Fatalf("dataset_id = %q", in.datasetID)
	}
	if in.autoPaginate {
		t.Fatal("expected auto_paginate false")
	}
}
