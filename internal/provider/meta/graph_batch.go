package meta

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

type graphBatchRequest struct {
	Method      string `json:"method"`
	RelativeURL string `json:"relative_url"`
}

type graphBatchResponseItem struct {
	Code int             `json:"code"`
	Body json.RawMessage `json:"body"`
}

func buildDeliveryRelativeURL(entityID, fieldsCSV string) string {
	q := url.Values{}
	q.Set("fields", fieldsCSV)
	return entityID + "?" + q.Encode()
}

func (s *service) graphBatchGET(
	ctx context.Context,
	mcpTool, userID string,
	relativeURLs []string,
) ([]graphBatchResponseItem, error) {
	if len(relativeURLs) == 0 {
		return nil, nil
	}
	if len(relativeURLs) > maxDeliveryEntityIDs {
		return nil, fmt.Errorf("meta: batch size %d exceeds max %d", len(relativeURLs), maxDeliveryEntityIDs)
	}

	requests := make([]graphBatchRequest, 0, len(relativeURLs))
	for _, rel := range relativeURLs {
		requests = append(requests, graphBatchRequest{
			Method:      "GET",
			RelativeURL: rel,
		})
	}
	batchJSON, err := json.Marshal(requests)
	if err != nil {
		return nil, err
	}

	raw, err := s.proxy.postFormWithRefresh(ctx, mcpTool, userID, "", nil, map[string]any{
		"batch": string(batchJSON),
	})
	if err != nil {
		return nil, err
	}

	var items []graphBatchResponseItem
	if err := json.Unmarshal(raw, &items); err != nil {
		return nil, err
	}
	return items, nil
}

func parseBatchItemBody(item graphBatchResponseItem) (map[string]any, error) {
	if item.Code < 200 || item.Code >= 300 {
		return nil, fmt.Errorf("batch item status %d: %s", item.Code, strings.TrimSpace(string(item.Body)))
	}
	var row map[string]any
	if err := json.Unmarshal(item.Body, &row); err != nil {
		return nil, err
	}
	return row, nil
}
