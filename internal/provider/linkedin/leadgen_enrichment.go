package linkedin

import "context"

// enrichCreativesWithLeadForms adds leadFormUrn, leadFormCtaLabel, and leadFormName fields
// to elements that contain a leadgenCallToAction. Form names are resolved in a single batch
// request when at least one plain adForm URN is present. All failures are silently ignored
// so the parent creative search is never blocked.
func enrichCreativesWithLeadForms(
	ctx context.Context,
	proxy linkedinUpstreamPort,
	userID, mcpTool string,
	elements []any,
) {
	// First pass: annotate CTA label + form URN; collect adForm IDs for batch fetch.
	adFormIDs := make([]string, 0)
	for _, item := range elements {
		row, ok := item.(map[string]any)
		if !ok {
			continue
		}
		label, formURN, ok := leadGenCTAFromCreative(row)
		if !ok {
			continue
		}
		row[fieldLeadFormUrn] = formURN
		if label != "" {
			row[fieldLeadFormCtaLabel] = label
		}
		if id, ok := formIDFromAdFormURN(formURN); ok {
			adFormIDs = append(adFormIDs, id)
		}
	}

	if len(adFormIDs) == 0 {
		return
	}

	formsByID := fetchLeadFormsByIDs(ctx, proxy, userID, mcpTool, adFormIDs)
	if len(formsByID) == 0 {
		return
	}

	// Second pass: apply resolved form names.
	for _, item := range elements {
		row, ok := item.(map[string]any)
		if !ok {
			continue
		}
		formURN, _ := row[fieldLeadFormUrn].(string)
		if formURN == "" {
			continue
		}
		id, ok := formIDFromAdFormURN(formURN)
		if !ok {
			continue
		}
		if summary, ok := formsByID[id]; ok && summary.name != "" {
			row[fieldLeadFormName] = summary.name
		}
	}
}
