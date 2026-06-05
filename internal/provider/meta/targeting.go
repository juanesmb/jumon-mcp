package meta

import (
	"context"
	"fmt"
	"strings"
)

func (s *service) searchInterests(ctx context.Context, mcpTool, userID string, in searchInterestsInput) (any, error) {
	q := strings.TrimSpace(in.query)
	if q == "" {
		return nil, fmt.Errorf("meta: q is required")
	}
	query := map[string]string{
		"type": "adinterest",
		"q":    q,
	}
	if in.limit > 0 {
		query["limit"] = toString(in.limit)
	}
	raw, err := s.proxy.getWithRefresh(ctx, mcpTool, userID, "search", query)
	if err != nil {
		return nil, err
	}
	return unmarshalPayload(raw)
}

func (s *service) searchGeoLocations(ctx context.Context, mcpTool, userID string, in searchGeoInput) (any, error) {
	q := strings.TrimSpace(in.query)
	if q == "" {
		return nil, fmt.Errorf("meta: q is required")
	}
	query := map[string]string{
		"type": "adgeolocation",
		"q":    q,
	}
	if len(in.locationTypes) > 0 {
		query["location_types"] = jsonEncode(in.locationTypes)
	}
	if in.limit > 0 {
		query["limit"] = toString(in.limit)
	}
	raw, err := s.proxy.getWithRefresh(ctx, mcpTool, userID, "search", query)
	if err != nil {
		return nil, err
	}
	return unmarshalPayload(raw)
}

func (s *service) estimateAudienceSize(ctx context.Context, mcpTool, userID, actID string, in estimateAudienceInput) (any, error) {
	normalized, err := normalizeActID(actID)
	if err != nil {
		return nil, err
	}
	goal := strings.TrimSpace(in.optimizationGoal)
	if goal == "" {
		goal = "REACH"
	}
	query := map[string]string{
		"targeting_spec":    jsonEncode(in.targeting),
		"optimization_goal": goal,
	}
	raw, err := s.proxy.getWithRefresh(ctx, mcpTool, userID, normalized+"/delivery_estimate", query)
	if err != nil {
		return nil, err
	}
	return unmarshalPayload(raw)
}

type searchInterestsInput struct {
	query string
	limit int
}

type searchGeoInput struct {
	query         string
	locationTypes []string
	limit         int
}

type estimateAudienceInput struct {
	targeting         map[string]any
	optimizationGoal  string
}

func parseSearchInterestsInput(params map[string]any) searchInterestsInput {
	in := searchInterestsInput{query: strings.TrimSpace(toString(params["q"]))}
	if limit, ok := toInt(params["limit"]); ok && limit > 0 {
		in.limit = limit
	} else {
		in.limit = defaultListLimit
	}
	return in
}

func parseSearchGeoInput(params map[string]any) searchGeoInput {
	in := searchGeoInput{
		query:         strings.TrimSpace(toString(params["q"])),
		locationTypes: toStringSlice(params["location_types"]),
	}
	if limit, ok := toInt(params["limit"]); ok && limit > 0 {
		in.limit = limit
	} else {
		in.limit = defaultListLimit
	}
	return in
}

func parseEstimateAudienceInput(params map[string]any) (estimateAudienceInput, error) {
	targeting, err := parseTargetingObject(params["targeting"])
	if err != nil {
		return estimateAudienceInput{}, err
	}
	return estimateAudienceInput{
		targeting:        targeting,
		optimizationGoal: strings.TrimSpace(toString(params["optimization_goal"])),
	}, nil
}

func parseTargetingObject(value any) (map[string]any, error) {
	if value == nil {
		return nil, fmt.Errorf("meta: targeting is required")
	}
	m, ok := value.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("meta: targeting must be an object")
	}
	return m, nil
}
