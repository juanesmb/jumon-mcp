package meta

import (
	"context"
	"fmt"
	"strings"
)

func (s *service) listCustomAudiences(ctx context.Context, mcpTool, userID, actID string, in listCustomAudiencesInput) (any, error) {
	normalized, err := normalizeActID(actID)
	if err != nil {
		return nil, err
	}
	query := buildListQuery(in.listPaginationInput)
	if in.subtypeFilter != "" {
		query["filtering"] = jsonEncode([]map[string]any{
			{"field": "subtype", "operator": "EQUAL", "value": in.subtypeFilter},
		})
	}
	return s.graphGETPaginated(ctx, mcpTool, userID, normalized+"/customaudiences", query, in.autoPaginate)
}

func (s *service) getCustomAudience(ctx context.Context, mcpTool, userID string, in getCustomAudienceInput) (any, error) {
	audienceID := strings.TrimSpace(in.audienceID)
	if audienceID == "" {
		return nil, fmt.Errorf("meta: custom_audience_id is required")
	}
	return s.getNode(ctx, mcpTool, userID, audienceID, in.fields, defaultCustomAudienceFields)
}

func (s *service) listCustomAudienceAdSets(ctx context.Context, mcpTool, userID string, in listCustomAudienceAdSetsInput) (any, error) {
	audienceID := strings.TrimSpace(in.audienceID)
	if audienceID == "" {
		return nil, fmt.Errorf("meta: custom_audience_id is required")
	}
	query := buildListQuery(in.listPaginationInput)
	return s.graphGETPaginated(ctx, mcpTool, userID, audienceID+"/adsets", query, in.autoPaginate)
}

type listCustomAudiencesInput struct {
	listPaginationInput
	subtypeFilter string
}

type getCustomAudienceInput struct {
	audienceID string
	fields     []string
}

type listCustomAudienceAdSetsInput struct {
	listPaginationInput
	audienceID string
}

func parseListCustomAudiencesInput(params map[string]any) listCustomAudiencesInput {
	return listCustomAudiencesInput{
		listPaginationInput: parseListPagination(params, defaultCustomAudienceListFields),
		subtypeFilter:       strings.TrimSpace(toString(params["subtype_filter"])),
	}
}

func parseGetCustomAudienceInput(params map[string]any) getCustomAudienceInput {
	return getCustomAudienceInput{
		audienceID: strings.TrimSpace(toString(params["custom_audience_id"])),
		fields:     toStringSlice(params["fields"]),
	}
}

func parseListCustomAudienceAdSetsInput(params map[string]any) listCustomAudienceAdSetsInput {
	return listCustomAudienceAdSetsInput{
		listPaginationInput: parseListPagination(params, defaultAdSetListFields),
		audienceID:          strings.TrimSpace(toString(params["custom_audience_id"])),
	}
}
