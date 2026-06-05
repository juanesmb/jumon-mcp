package meta

import (
	"context"
	"fmt"
	"strings"
)

func (s *service) getAdImages(ctx context.Context, mcpTool, userID, actID string, in adImagesInput) (any, error) {
	normalized, err := normalizeActID(actID)
	if err != nil {
		return nil, err
	}
	query := buildListQuery(in.listPaginationInput)
	if len(in.hashes) > 0 {
		query["hashes"] = jsonEncode(in.hashes)
	}
	if in.name != "" {
		query["name"] = in.name
	}
	if in.minWidth > 0 {
		query["minwidth"] = toString(in.minWidth)
	}
	if in.minHeight > 0 {
		query["minheight"] = toString(in.minHeight)
	}
	return s.graphGETPaginated(ctx, mcpTool, userID, normalized+"/adimages", query, in.autoPaginate)
}

func (s *service) getAdVideos(ctx context.Context, mcpTool, userID, actID string, in adVideosInput) (any, error) {
	normalized, err := normalizeActID(actID)
	if err != nil {
		return nil, err
	}
	query := buildListQuery(in.listPaginationInput)
	if len(in.videoIDs) > 0 {
		query["filtering"] = jsonEncode([]map[string]any{
			{"field": "id", "operator": "IN", "value": in.videoIDs},
		})
	}
	return s.graphGETPaginated(ctx, mcpTool, userID, normalized+"/advideos", query, in.autoPaginate)
}

func (s *service) getAdPreview(ctx context.Context, mcpTool, userID string, in adPreviewInput) (any, error) {
	adID := strings.TrimSpace(in.adID)
	if adID == "" {
		return nil, fmt.Errorf("meta: ad_id is required")
	}
	query := map[string]string{}
	if in.adFormat != "" {
		query["ad_format"] = in.adFormat
	}
	if in.locale != "" {
		query["locale"] = in.locale
	}
	if in.startDate != "" {
		query["start_date"] = in.startDate
	}
	if in.endDate != "" {
		query["end_date"] = in.endDate
	}
	raw, err := s.proxy.getWithRefresh(ctx, mcpTool, userID, adID+"/previews", query)
	if err != nil {
		return nil, err
	}
	return unmarshalPayload(raw)
}

type adImagesInput struct {
	listPaginationInput
	hashes     []string
	name       string
	minWidth   int
	minHeight  int
}

type adVideosInput struct {
	listPaginationInput
	videoIDs []string
}

type adPreviewInput struct {
	adID      string
	adFormat  string
	locale    string
	startDate string
	endDate   string
}

func parseAdImagesInput(params map[string]any) adImagesInput {
	in := adImagesInput{
		listPaginationInput: parseListPagination(params, defaultAdImageFields),
		hashes:              toStringSlice(params["hashes"]),
		name:                strings.TrimSpace(toString(params["name"])),
	}
	if w, ok := toInt(params["minwidth"]); ok {
		in.minWidth = w
	}
	if h, ok := toInt(params["minheight"]); ok {
		in.minHeight = h
	}
	return in
}

func parseAdVideosInput(params map[string]any) adVideosInput {
	return adVideosInput{
		listPaginationInput: parseListPagination(params, defaultAdVideoFields),
		videoIDs:            toStringSlice(params["video_ids"]),
	}
}

func parseAdPreviewInput(params map[string]any) adPreviewInput {
	return adPreviewInput{
		adID:      strings.TrimSpace(toString(params["ad_id"])),
		adFormat:  strings.TrimSpace(toString(params["ad_format"])),
		locale:    strings.TrimSpace(toString(params["locale"])),
		startDate: strings.TrimSpace(toString(params["start_date"])),
		endDate:   strings.TrimSpace(toString(params["end_date"])),
	}
}
