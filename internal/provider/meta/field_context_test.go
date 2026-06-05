package meta

import "testing"

func TestBuildFieldContextFilterByName(t *testing.T) {
	root := buildFieldContextResponse(fieldContextInput{fieldNames: []string{"spend"}}).(map[string]any)
	fields := root["fields"].([]fieldContextEntry)
	if len(fields) != 1 || fields[0].Name != "spend" {
		t.Fatalf("fields %#v", root["fields"])
	}
}

func TestLevelMatchesFieldContext(t *testing.T) {
	root := buildFieldContextResponse(fieldContextInput{level: "campaign"}).(map[string]any)
	fields := root["fields"].([]fieldContextEntry)
	if len(fields) == 0 {
		t.Fatal("expected campaign-level fields")
	}
	for _, f := range fields {
		if !levelMatches(f.Levels, "campaign") {
			t.Fatalf("field %s not valid at campaign level", f.Name)
		}
	}
}

func TestEmbeddedFieldCatalogSize(t *testing.T) {
	if len(embeddedFieldCatalog) < 30 {
		t.Fatalf("expected expanded catalog, got %d entries", len(embeddedFieldCatalog))
	}
}
