package meta

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

func (s *service) getDeliveryErrors(ctx context.Context, mcpTool, userID string, in deliveryErrorsInput) (any, error) {
	ids := normalizeEntityIDs(in.entityIDs)
	if len(ids) == 0 {
		return map[string]any{"results": []any{}}, nil
	}
	if len(ids) == 1 {
		row, err := s.fetchDeliveryErrorRow(ctx, mcpTool, userID, ids[0])
		if err != nil {
			return nil, err
		}
		if row == nil {
			return map[string]any{"results": []map[string]any{{"entity_id": ids[0]}}}, nil
		}
		return map[string]any{"results": []map[string]any{buildDeliveryErrorEntry(ids[0], row)}}, nil
	}

	results, err := s.fetchDeliveryErrorsBatch(ctx, mcpTool, userID, ids)
	if err != nil {
		results, err = s.fetchDeliveryErrorsSequential(ctx, mcpTool, userID, ids)
		if err != nil {
			return nil, err
		}
	}
	return map[string]any{"results": results}, nil
}

func normalizeEntityIDs(ids []string) []string {
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		if trimmed := strings.TrimSpace(id); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

func (s *service) fetchDeliveryErrorsBatch(ctx context.Context, mcpTool, userID string, ids []string) ([]map[string]any, error) {
	results := make([]map[string]any, 0, len(ids))
	for start := 0; start < len(ids); start += maxDeliveryEntityIDs {
		end := start + maxDeliveryEntityIDs
		if end > len(ids) {
			end = len(ids)
		}
		chunk := ids[start:end]
		urls := make([]string, len(chunk))
		for i, id := range chunk {
			urls[i] = buildDeliveryRelativeURL(id, joinCSV(deliveryErrorBatchFields))
		}
		items, err := s.graphBatchGET(ctx, mcpTool, userID, urls)
		if err != nil {
			return nil, err
		}
		if len(items) != len(chunk) {
			return nil, fmt.Errorf("meta: batch response length mismatch")
		}
		for i, item := range items {
			requestedID := chunk[i]
			row, err := parseBatchItemBody(item)
			if err != nil {
				return nil, err
			}
			if row == nil {
				results = append(results, map[string]any{"entity_id": requestedID})
				continue
			}
			results = append(results, buildDeliveryErrorEntry(requestedID, row))
		}
	}
	return results, nil
}

func (s *service) fetchDeliveryErrorsSequential(ctx context.Context, mcpTool, userID string, ids []string) ([]map[string]any, error) {
	results := make([]map[string]any, 0, len(ids))
	for _, id := range ids {
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
	return results, nil
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
	for _, key := range deliveryErrorBatchFields {
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
