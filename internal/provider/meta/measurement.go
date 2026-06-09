package meta

import (
	"context"
)

func (s *service) listCustomConversions(ctx context.Context, mcpTool, userID, actID string, in listCustomConversionsInput) (any, error) {
	normalized, err := normalizeActID(actID)
	if err != nil {
		return nil, err
	}
	query := buildListQuery(in.listPaginationInput)
	if in.datasetID != "" {
		query["filtering"] = jsonEncode([]map[string]any{
			{"field": "pixel", "operator": "EQUAL", "value": in.datasetID},
		})
	}
	return s.graphGETPaginated(ctx, mcpTool, userID, normalized+"/customconversions", query, in.autoPaginate)
}

func (s *service) listDatasets(ctx context.Context, mcpTool, userID, actID string, in listPaginationInput) (any, error) {
	normalized, err := normalizeActID(actID)
	if err != nil {
		return nil, err
	}
	query := buildListQuery(in)
	return s.graphGETPaginated(ctx, mcpTool, userID, normalized+"/adspixels", query, in.autoPaginate)
}

func (s *service) getDataset(ctx context.Context, mcpTool, userID string, in getDatasetInput) (any, error) {
	datasetID, err := requireDatasetID(in.datasetID)
	if err != nil {
		return nil, err
	}
	return s.getNode(ctx, mcpTool, userID, datasetID, in.fields, defaultDatasetFields)
}

func (s *service) getDatasetQuality(ctx context.Context, mcpTool, userID string, in datasetQualityInput) (any, error) {
	datasetID, err := requireDatasetID(in.datasetID)
	if err != nil {
		return nil, err
	}
	query := buildDatasetQualityQuery(datasetQualityInput{datasetID: datasetID, fields: in.fields})
	raw, err := s.proxy.getWithRefresh(ctx, mcpTool, userID, "dataset_quality", query)
	if err != nil {
		return nil, err
	}
	return unmarshalPayload(raw)
}
