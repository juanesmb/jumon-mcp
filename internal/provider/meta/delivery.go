package meta

import (
	"context"
	"encoding/json"
	"strings"
)

func (s *service) getDeliveryErrors(ctx context.Context, mcpTool, userID string, in deliveryErrorsInput) (any, error) {
	results := make([]map[string]any, 0, len(in.entityIDs))

	for _, entityID := range in.entityIDs {
		id := strings.TrimSpace(entityID)
		if id == "" {
			continue
		}
		row, err := s.fetchDeliveryErrorRow(ctx, mcpTool, userID, id)
		if err != nil {
			return nil, err
		}
		if row == nil {
			results = append(results, map[string]any{"entity_id": id})
			continue
		}
		results = append(results, buildDeliveryErrorEntry(id, row))
	}

	return map[string]any{"results": results}, nil
}

func (s *service) fetchDeliveryErrorRow(ctx context.Context, mcpTool, userID, entityID string) (map[string]any, error) {
	raw, err := s.proxy.getWithRefresh(ctx, mcpTool, userID, entityID, map[string]string{
		"fields": joinCSV(deliveryErrorAdFields),
	})
	if err == nil {
		return decodeDeliveryRow(raw)
	}
	if !isGraphNonexistingFieldError(err) {
		return nil, err
	}

	raw, err = s.proxy.getWithRefresh(ctx, mcpTool, userID, entityID, map[string]string{
		"fields": joinCSV(deliveryErrorStructureFields),
	})
	if err != nil {
		return nil, err
	}
	return decodeDeliveryRow(raw)
}

func decodeDeliveryRow(raw json.RawMessage) (map[string]any, error) {
	payload, err := unmarshalPayload(raw)
	if err != nil {
		return nil, err
	}
	row, ok := payload.(map[string]any)
	if !ok {
		return nil, nil
	}
	return row, nil
}

func buildDeliveryErrorEntry(requestedID string, row map[string]any) map[string]any {
	entry := map[string]any{"entity_id": requestedID}
	if v, ok := row["id"]; ok {
		entry["entity_id"] = v
	}
	entry["entity_type"] = inferEntityType(row)
	for _, key := range append(append([]string{}, deliveryErrorBaseFields...), "failed_delivery_checks", "issues_info") {
		if v, ok := row[key]; ok {
			entry[key] = v
		}
	}
	return entry
}

func inferEntityType(row map[string]any) string {
	if _, ok := row["failed_delivery_checks"]; ok {
		return "ad"
	}
	if _, ok := row["issues_info"]; ok {
		return "campaign_or_adset"
	}
	return "unknown"
}

func isGraphNonexistingFieldError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "nonexisting field") ||
		strings.Contains(msg, "is not valid for fields param")
}
