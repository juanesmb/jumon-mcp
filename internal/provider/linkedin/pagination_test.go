package linkedin

import (
	"context"
	"testing"
)

type stubLinkedInUpstream struct {
	pages []any
	calls int
}

func (s *stubLinkedInUpstream) requestJSON(
	_ context.Context,
	_, _, _, _ string,
	_ map[string]string,
	_ any,
	_ map[string]string,
) (any, error) {
	idx := s.calls
	s.calls++
	if idx >= len(s.pages) {
		return s.pages[len(s.pages)-1], nil
	}
	return s.pages[idx], nil
}

func TestFetchSearchPages_autoPaginateMergesElements(t *testing.T) {
	t.Parallel()

	stub := &stubLinkedInUpstream{
		pages: []any{
			map[string]any{
				"elements": []any{map[string]any{"id": "1"}},
				"metadata": map[string]any{"nextPageToken": "page2"},
			},
			map[string]any{
				"elements": []any{map[string]any{"id": "2"}},
				"metadata": map[string]any{},
			},
		},
	}

	result, err := fetchSearchPages(context.Background(), stub, "user", "tool", "path", map[string]string{"q": "search"}, true)
	if err != nil {
		t.Fatalf("fetchSearchPages() error = %v", err)
	}

	page, ok := result.(map[string]any)
	if !ok {
		t.Fatalf("result type = %T", result)
	}
	elements, ok := page["elements"].([]any)
	if !ok || len(elements) != 2 {
		t.Fatalf("elements = %#v, want 2 items", page["elements"])
	}
	if stub.calls != 2 {
		t.Fatalf("request calls = %d, want 2", stub.calls)
	}
	meta, ok := page["metadata"].(map[string]any)
	if !ok {
		t.Fatal("expected metadata")
	}
	if _, ok := meta["nextPageToken"]; ok {
		t.Fatal("nextPageToken should be stripped from merged metadata")
	}
}

func TestFetchSearchPages_singlePageWhenAutoPaginateDisabled(t *testing.T) {
	t.Parallel()

	stub := &stubLinkedInUpstream{
		pages: []any{
			map[string]any{
				"elements": []any{map[string]any{"id": "1"}},
				"metadata": map[string]any{"nextPageToken": "page2"},
			},
		},
	}

	result, err := fetchSearchPages(context.Background(), stub, "user", "tool", "path", map[string]string{"q": "search"}, false)
	if err != nil {
		t.Fatalf("fetchSearchPages() error = %v", err)
	}

	page, ok := result.(map[string]any)
	if !ok {
		t.Fatalf("result type = %T", result)
	}
	elements, ok := page["elements"].([]any)
	if !ok || len(elements) != 1 {
		t.Fatalf("elements = %#v, want 1 item", page["elements"])
	}
	if stub.calls != 1 {
		t.Fatalf("request calls = %d, want 1", stub.calls)
	}
}
