package googleads

import (
	"context"
	"testing"
)

type stubGoogleProxy struct {
	pages []any
	calls int
}

func (s *stubGoogleProxy) requestJSON(
	_ context.Context,
	_, _, _, _ string,
	_ any,
	_ map[string]string,
) (any, error) {
	idx := s.calls
	s.calls++
	if idx >= len(s.pages) {
		return map[string]any{"results": []any{}}, nil
	}
	return s.pages[idx], nil
}

func TestGoogleSearchPaginated_autoPaginateMergesResults(t *testing.T) {
	proxy := &stubGoogleProxy{
		pages: []any{
			map[string]any{
				"results":       []any{map[string]any{"id": "1"}},
				"nextPageToken": "page-2",
			},
			map[string]any{
				"results": []any{map[string]any{"id": "2"}},
			},
		},
	}
	svc := &service{proxy: proxy, apiVersion: "v24"}

	out, err := svc.googleSearchPaginated(
		context.Background(), "user", "google_search_keywords", "123", "", "SELECT 1", true,
	)
	if err != nil {
		t.Fatal(err)
	}
	root, ok := out.(map[string]any)
	if !ok {
		t.Fatalf("unexpected type %T", out)
	}
	results, ok := root["results"].([]any)
	if !ok || len(results) != 2 {
		t.Fatalf("results = %+v", root["results"])
	}
	meta, ok := root["metadata"].(map[string]any)
	if !ok || meta["pages_fetched"] != 2 {
		t.Fatalf("metadata = %+v", root["metadata"])
	}
}
