package meta

import "testing"

func TestPagingAfterCursor_requiresNext(t *testing.T) {
	page := map[string]any{
		"paging": map[string]any{
			"cursors": map[string]any{"after": "abc"},
		},
	}
	if after, ok := pagingAfterCursor(page); ok || after != "" {
		t.Fatalf("expected no pagination without next, got %q %v", after, ok)
	}
	page["paging"].(map[string]any)["next"] = "https://graph.facebook.com/next"
	after, ok := pagingAfterCursor(page)
	if !ok || after != "abc" {
		t.Fatalf("got %q %v", after, ok)
	}
}
