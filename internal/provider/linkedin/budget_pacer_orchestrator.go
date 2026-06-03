package linkedin

import (
	"context"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"
)

func (s *service) getBudgetPacerReport(
	ctx context.Context,
	userID, mcpTool string,
	in BudgetPacerInput,
) (BudgetPacerReport, error) {
	periodStart, periodEnd, err := resolvePeriodDates(in.PeriodStart, in.PeriodEnd)
	if err != nil {
		return BudgetPacerReport{}, err
	}

	report := BudgetPacerReport{
		AccountID: in.AccountID,
		Period: DatePeriod{
			Start: in.PeriodStart,
			End:   formatDate(periodEnd),
		},
		Metadata:        PacerReportMetadata{Warnings: []string{}},
		Rows:            []PacerRow{},
		ByCampaignGroup: []CampaignGroupPacerRollup{},
	}

	snapshots, truncated, err := s.fetchCampaignSnapshotsForPacer(ctx, userID, mcpTool, in)
	if err != nil {
		return BudgetPacerReport{}, err
	}
	if truncated {
		report.Metadata.Truncated = true
		report.Metadata.Warnings = append(report.Metadata.Warnings,
			"campaign list may be incomplete (pagination cap reached); narrow with campaign_ids or campaign_group_ids",
		)
	}

	snapshots = filterSnapshotsByTestFlag(snapshots, in.IncludeTestCampaigns)
	snapshots = filterSnapshotsByCampaignIDs(snapshots, in.CampaignIDs)
	if len(in.CampaignGroupIDs) > 0 {
		snapshots = filterSnapshotsByGroup(snapshots, in.CampaignGroupIDs)
	}

	report.Metadata.CampaignsIncluded = len(snapshots)
	if len(snapshots) == 0 {
		return report, nil
	}

	spendCurrent, err := s.fetchCampaignSpendMap(ctx, userID, mcpTool, in.AccountID, snapshots, in.PeriodStart, formatDate(periodEnd))
	if err != nil {
		return BudgetPacerReport{}, err
	}

	var spendPrior map[string]float64
	if strings.TrimSpace(in.ComparePeriodStart) != "" {
		compareEnd := in.ComparePeriodEnd
		if compareEnd == "" {
			_, compareEndTime, err := resolvePeriodDates(in.ComparePeriodStart, "")
			if err != nil {
				return BudgetPacerReport{}, err
			}
			compareEnd = formatDate(compareEndTime)
		}
		report.ComparePeriod = &DatePeriod{Start: in.ComparePeriodStart, End: compareEnd}
		spendPrior, err = s.fetchCampaignSpendMap(ctx, userID, mcpTool, in.AccountID, snapshots, in.ComparePeriodStart, compareEnd)
		if err != nil {
			return BudgetPacerReport{}, err
		}
	}

	groupNames, groupErr := s.fetchCampaignGroupNames(ctx, userID, mcpTool, in.AccountID, snapshots)
	if groupErr != nil {
		report.Metadata.Warnings = append(report.Metadata.Warnings,
			"campaign group names unavailable: "+groupErr.Error(),
		)
	}

	report.Rows = buildPacerRows(snapshots, spendCurrent, spendPrior, periodStart, periodEnd, in.PacingThresholds)
	report.ByCampaignGroup = buildGroupRollups(report.Rows, groupNames, in.PacingThresholds)
	report.Summary = buildAccountSummary(report.Rows, in.PacingThresholds)

	return report, nil
}

func (s *service) fetchCampaignSnapshotsForPacer(
	ctx context.Context,
	userID, mcpTool string,
	in BudgetPacerInput,
) ([]CampaignSnapshot, bool, error) {
	apiPath := fmt.Sprintf("adAccounts/%s/adCampaigns", url.PathEscape(in.AccountID))
	query := buildLinkedInSearchQuery(linkedInSearchQuery{
		statusFilter: in.StatusFilter,
		pageSize:     defaultCampaignsPageSize,
	})
	autoPaginate := resolveAutoPaginate(in.AutoPaginate, "", len(in.CampaignGroupIDs) > 0)

	raw, truncated, err := fetchSearchPagesWithTruncation(ctx, s.proxy, userID, mcpTool, apiPath, query, autoPaginate, nil)
	if err != nil {
		return nil, false, err
	}
	snapshots, err := campaignSnapshotsFromResponse(raw)
	if err != nil {
		return nil, false, err
	}
	return snapshots, truncated, nil
}

func filterSnapshotsByGroup(snapshots []CampaignSnapshot, groupIDs []string) []CampaignSnapshot {
	allowed := make(map[string]struct{}, len(groupIDs))
	for _, urn := range toCampaignGroupURNs(groupIDs) {
		allowed[urn] = struct{}{}
	}
	filtered := make([]CampaignSnapshot, 0, len(snapshots))
	for _, snap := range snapshots {
		if _, ok := allowed[snap.CampaignGroupURN]; ok {
			filtered = append(filtered, snap)
		}
	}
	return filtered
}

func (s *service) fetchCampaignSpendMap(
	ctx context.Context,
	userID, mcpTool, accountID string,
	snapshots []CampaignSnapshot,
	startDate, endDate string,
) (map[string]float64, error) {
	campaignIDs := make([]string, 0, len(snapshots))
	for _, snap := range snapshots {
		campaignIDs = append(campaignIDs, snap.ID)
	}

	analyticsIn := getAnalyticsInput{
		accountID:       accountID,
		startDate:       startDate,
		endDate:         endDate,
		finderType:      finderTypeAnalytics,
		pivots:          []string{"CAMPAIGN"},
		timeGranularity: "ALL",
		fields:          []string{"costInLocalCurrency"},
		campaignIDs:     campaignIDs,
		autoPaginate:    true,
		pageSize:        defaultAnalyticsPageSize,
	}

	query, err := buildAdAnalyticsQuery(analyticsIn)
	if err != nil {
		return nil, err
	}

	raw, err := fetchAnalyticsPages(ctx, s.proxy, userID, mcpTool, query, true)
	if err != nil {
		return nil, err
	}
	return parseSpendByCampaignID(raw), nil
}

func parseSpendByCampaignID(raw any) map[string]float64 {
	out := make(map[string]float64)
	pageMap, ok := raw.(map[string]any)
	if !ok {
		return out
	}
	elements, ok := pageMap["elements"].([]any)
	if !ok {
		return out
	}
	for _, item := range elements {
		row, ok := item.(map[string]any)
		if !ok {
			continue
		}
		campaignID := campaignIDFromPivotValues(row["pivotValues"])
		if campaignID == "" {
			continue
		}
		spend, ok := numericField(row, "costInLocalCurrency")
		if !ok {
			continue
		}
		out[campaignID] += spend
	}
	return out
}

func buildPacerRows(
	snapshots []CampaignSnapshot,
	spendCurrent, spendPrior map[string]float64,
	periodStart, periodEnd time.Time,
	thresholds PacingThresholds,
) []PacerRow {
	rows := make([]PacerRow, 0, len(snapshots))
	for _, snap := range snapshots {
		spend := spendCurrent[snap.ID]
		pacing := computePacing(PacingInputs{
			BudgetType:       snap.BudgetType,
			BudgetAmount:     snap.BudgetAmount,
			RunScheduleStart: snap.RunScheduleStart,
			RunScheduleEnd:   snap.RunScheduleEnd,
			SpendToDate:      spend,
			PeriodStart:      periodStart,
			PeriodEnd:        periodEnd,
		}, thresholds)

		row := PacerRow{
			CampaignID:           snap.ID,
			CampaignURN:          snap.URN,
			Name:                 snap.Name,
			Status:               snap.Status,
			CampaignGroupURN:     snap.CampaignGroupURN,
			BudgetType:           snap.BudgetType,
			BudgetAmount:         snap.BudgetAmount,
			CurrencyCode:         snap.CurrencyCode,
			SpendToDate:          round2(spend),
			ExpectedSpendToDate:  pacing.ExpectedSpendToDate,
			PacingPercent:        pacing.PacingPercent,
			PacingStatus:         pacing.PacingStatus,
			ProjectedPeriodSpend: pacing.ProjectedPeriodSpend,
		}
		if spendPrior != nil {
			prior := spendPrior[snap.ID]
			p := round2(prior)
			row.SpendPrior = &p
			row.SpendChangePercent = spendChangePercent(spend, prior)
		}
		rows = append(rows, row)
	}
	sort.Slice(rows, func(i, j int) bool {
		return rows[i].CampaignID < rows[j].CampaignID
	})
	return rows
}

func buildGroupRollups(
	rows []PacerRow,
	groupNames map[string]string,
	thresholds PacingThresholds,
) []CampaignGroupPacerRollup {
	byGroup := make(map[string]*CampaignGroupPacerRollup)

	for _, row := range rows {
		groupURN := row.CampaignGroupURN
		if groupURN == "" {
			groupURN = "_ungrouped"
		}
		rollup, ok := byGroup[groupURN]
		if !ok {
			rollup = &CampaignGroupPacerRollup{
				CampaignGroupURN:  groupURN,
				CampaignGroupName: groupNames[groupURN],
				PacingStatus:      pacingStatusUnknown,
			}
			byGroup[groupURN] = rollup
		}
		rollup.SpendToDate += row.SpendToDate
		if row.ExpectedSpendToDate != nil {
			if rollup.ExpectedSpendToDate == nil {
				v := 0.0
				rollup.ExpectedSpendToDate = &v
			}
			*rollup.ExpectedSpendToDate += *row.ExpectedSpendToDate
		}
		if row.CurrencyCode != "" && rollup.CurrencyCode == "" {
			rollup.CurrencyCode = row.CurrencyCode
		} else if row.CurrencyCode != "" && rollup.CurrencyCode != row.CurrencyCode {
			rollup.CurrencyCode = ""
		}
		rollup.CampaignCount++
	}

	rollups := make([]CampaignGroupPacerRollup, 0, len(byGroup))
	for _, rollup := range byGroup {
		r := *rollup
		if r.ExpectedSpendToDate != nil && *r.ExpectedSpendToDate > 0 {
			pct := (r.SpendToDate / *r.ExpectedSpendToDate) * 100
			r.PacingPercent = floatPtr(round2(pct))
			r.PacingStatus = classifyPacing(pct/100, thresholds)
		}
		if r.CampaignGroupURN == "_ungrouped" {
			r.CampaignGroupURN = ""
		}
		rollups = append(rollups, r)
	}
	sort.Slice(rollups, func(i, j int) bool {
		return rollups[i].CampaignGroupURN < rollups[j].CampaignGroupURN
	})
	return rollups
}

func buildAccountSummary(rows []PacerRow, thresholds PacingThresholds) *AccountPacerSummary {
	if len(rows) == 0 {
		return nil
	}
	currency := rows[0].CurrencyCode
	var totalSpend, totalExpected float64
	hasExpected := false

	for _, row := range rows {
		if row.CurrencyCode != "" && row.CurrencyCode != currency {
			return nil
		}
		totalSpend += row.SpendToDate
		if row.ExpectedSpendToDate != nil {
			totalExpected += *row.ExpectedSpendToDate
			hasExpected = true
		}
	}

	summary := &AccountPacerSummary{
		SpendToDate:  round2(totalSpend),
		PacingStatus: pacingStatusUnknown,
		CurrencyCode: currency,
	}
	if hasExpected && totalExpected > 0 {
		summary.ExpectedSpendToDate = floatPtr(round2(totalExpected))
		pct := (totalSpend / totalExpected) * 100
		summary.PacingPercent = floatPtr(round2(pct))
		summary.PacingStatus = classifyPacing(pct/100, thresholds)
	}
	return summary
}

func (s *service) fetchCampaignGroupNames(
	ctx context.Context,
	userID, mcpTool, accountID string,
	snapshots []CampaignSnapshot,
) (map[string]string, error) {
	hasGroup := false
	for _, snap := range snapshots {
		if snap.CampaignGroupURN != "" {
			hasGroup = true
			break
		}
	}
	if !hasGroup {
		return map[string]string{}, nil
	}

	in := getCampaignGroupsInput{
		accountID:    accountID,
		autoPaginate: true,
		pageSize:     defaultCampaignsPageSize,
	}
	raw, err := s.getCampaignGroups(ctx, userID, mcpTool, in)
	if err != nil {
		return nil, err
	}
	pageMap, ok := raw.(map[string]any)
	if !ok {
		return map[string]string{}, nil
	}
	elements, ok := pageMap["elements"].([]any)
	if !ok {
		return map[string]string{}, nil
	}

	names := make(map[string]string)
	for _, item := range elements {
		row, ok := item.(map[string]any)
		if !ok {
			continue
		}
		urn := campaignGroupURNFromRow(row)
		name := strings.TrimSpace(toString(row["name"]))
		if urn != "" && name != "" {
			names[urn] = name
		}
	}
	return names, nil
}

func campaignGroupURNFromRow(row map[string]any) string {
	return normalizeCampaignGroupURN(campaignIDFromRow(row))
}

func formatDate(t time.Time) string {
	return t.Format("2006-01-02")
}
