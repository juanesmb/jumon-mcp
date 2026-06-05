package meta

import (
	"encoding/json"
	"fmt"
	"strings"
)

func formatMetaAPIError(statusCode int, body string) error {
	msg := parseGraphErrorMessage(body)
	if msg == "" {
		msg = body
	}
	if len(msg) > 2000 {
		msg = msg[:2000] + "..."
	}
	switch statusCode {
	case 429:
		if msg != "" {
			return fmt.Errorf("meta api returned status 429 (rate limit): %s — retry after a short wait", msg)
		}
		return fmt.Errorf("meta api returned status 429 (rate limit) — retry after a short wait")
	default:
		if msg != "" {
			return fmt.Errorf("meta api returned status %d: %s", statusCode, msg)
		}
		return fmt.Errorf("meta api returned status %d", statusCode)
	}
}

func parseGraphErrorMessage(body string) string {
	body = strings.TrimSpace(body)
	if body == "" || body[0] != '{' {
		return body
	}
	var root map[string]any
	if err := json.Unmarshal([]byte(body), &root); err != nil {
		return body
	}
	errObj, ok := root["error"].(map[string]any)
	if !ok {
		return body
	}
	if userMsg, ok := errObj["error_user_msg"].(string); ok && strings.TrimSpace(userMsg) != "" {
		return strings.TrimSpace(userMsg)
	}
	if message, ok := errObj["message"].(string); ok && strings.TrimSpace(message) != "" {
		return strings.TrimSpace(message)
	}
	return body
}
