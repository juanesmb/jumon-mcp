package meta

import (
	"context"
	"fmt"
	"strings"
)

func (s *service) listCreatives(ctx context.Context, mcpTool, userID, actID string, in listCreativesInput) (any, error) {
	normalized, err := normalizeActID(actID)
	if err != nil {
		return nil, err
	}
	query := buildListQuery(in.listPaginationInput)
	if len(in.filtering) > 0 {
		query["filtering"] = jsonEncode(in.filtering)
	}
	return s.graphGETPaginated(ctx, mcpTool, userID, normalized+"/adcreatives", query, in.autoPaginate)
}

func (s *service) getCreative(ctx context.Context, mcpTool, userID string, in getCreativeInput) (any, error) {
	creativeID := strings.TrimSpace(in.creativeID)
	if creativeID == "" {
		return nil, fmt.Errorf("meta: creative_id is required")
	}
	fields := in.fields
	if len(fields) == 0 {
		fields = append([]string(nil), defaultCreativeFields...)
	}
	query := map[string]string{"fields": joinCSV(fields)}
	if in.thumbnailWidth > 0 {
		query["thumbnail_width"] = toString(in.thumbnailWidth)
	}
	if in.thumbnailHeight > 0 {
		query["thumbnail_height"] = toString(in.thumbnailHeight)
	}
	raw, err := s.proxy.getWithRefresh(ctx, mcpTool, userID, creativeID, query)
	if err != nil {
		return nil, err
	}
	return unmarshalPayload(raw)
}

type listCreativesInput struct {
	listPaginationInput
	filtering []map[string]any
}

type getCreativeInput struct {
	creativeID      string
	fields          []string
	thumbnailWidth  int
	thumbnailHeight int
}

func parseListCreativesInput(params map[string]any) (listCreativesInput, error) {
	filtering, err := parseFiltering(params["filtering"])
	if err != nil {
		return listCreativesInput{}, err
	}
	return listCreativesInput{
		listPaginationInput: parseListPagination(params, defaultCreativeListFields),
		filtering:           filtering,
	}, nil
}

func parseGetCreativeInput(params map[string]any) getCreativeInput {
	in := getCreativeInput{
		creativeID: strings.TrimSpace(toString(params["creative_id"])),
		fields:     toStringSlice(params["fields"]),
	}
	if w, ok := toInt(params["thumbnail_width"]); ok {
		in.thumbnailWidth = w
	}
	if h, ok := toInt(params["thumbnail_height"]); ok {
		in.thumbnailHeight = h
	}
	return in
}
