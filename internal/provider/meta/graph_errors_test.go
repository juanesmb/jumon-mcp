package meta

import (
	"strings"
	"testing"
)

func TestParseGraphErrorMessage_userMsg(t *testing.T) {
	body := `{"error":{"message":"Invalid parameter","error_user_msg":"You can't run lead ads until your Page accepts Lead Generation Terms."}}`
	got := parseGraphErrorMessage(body)
	if !strings.Contains(got, "Lead Generation Terms") {
		t.Fatalf("got %q", got)
	}
}

func TestFormatMetaAPIError_rateLimit(t *testing.T) {
	err := formatMetaAPIError(429, `{"error":{"message":"(#80004) Too many calls"}}`)
	if err == nil || !strings.Contains(err.Error(), "rate limit") {
		t.Fatalf("got %v", err)
	}
}
