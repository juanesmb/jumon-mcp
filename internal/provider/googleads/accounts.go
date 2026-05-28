package googleads

import (
	"context"
	"fmt"
	"strings"
)

const listAdAccountsMessage = "Use descriptive_name to pick customer_id. For client accounts under an MCC, pass login_customer_id from google_resolve_customer or google_list_client_accounts_under_manager."

func (s *service) listAdAccounts(ctx context.Context, userID, mcpTool string) (any, error) {
	accounts, truncated, skipped, err := s.fetchAccessibleAccountRecords(ctx, userID, mcpTool)
	if err != nil {
		return nil, err
	}
	message := listAdAccountsMessage
	if truncated {
		message = listAdAccountsMessage + " " + listAdAccountsTruncatedMessage
	}
	return map[string]any{
		"accounts":            accounts,
		"truncated":           truncated,
		"skipped_unavailable": skipped,
		"message":             message,
	}, nil
}

func (s *service) resolveCustomer(ctx context.Context, userID, mcpTool string, in resolveCustomerInput) (any, error) {
	accounts, _, _, err := s.fetchAccessibleAccountRecords(ctx, userID, mcpTool)
	if err != nil {
		return nil, err
	}

	matches := make([]customerMatch, 0)
	seen := make(map[string]struct{})

	for _, account := range accounts {
		if !matchAccountName(account.DescriptiveName, in.accountName, in.matchMode) {
			continue
		}
		match := customerMatch{
			CustomerID:      account.CustomerID,
			DescriptiveName: account.DescriptiveName,
			Manager:         account.Manager,
			MatchType:       "direct",
		}
		if account.LoginCustomerID != nil {
			match.LoginCustomerID = *account.LoginCustomerID
		}
		key := match.CustomerID + "|" + match.LoginCustomerID
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		matches = append(matches, match)
	}

	if in.searchUnderManagers && len(matches) == 0 {
		managerCount := 0
		for _, account := range accounts {
			if !account.Manager {
				continue
			}
			if managerCount >= s.maxManagerScan {
				break
			}
			managerCount++
			clientMatches, err := s.searchManagerClientsByName(ctx, userID, mcpTool, account.CustomerID, in)
			if err != nil {
				return nil, err
			}
			for _, match := range clientMatches {
				key := match.CustomerID + "|" + match.LoginCustomerID
				if _, ok := seen[key]; ok {
					continue
				}
				seen[key] = struct{}{}
				matches = append(matches, match)
			}
		}
	}

	return map[string]any{
		"matches": matches,
		"message": "Use customer_id and login_customer_id (when present) in subsequent Google tools.",
	}, nil
}

func (s *service) listClientAccounts(ctx context.Context, userID, mcpTool string, in listClientAccountsInput) (any, error) {
	query := buildClientAccountsQuery(in.clientNameContains)
	loginID := in.loginCustomerID
	if loginID == "" {
		loginID = in.customerID
	}
	return s.googleSearch(ctx, userID, mcpTool, in.customerID, loginID, query)
}

func (s *service) searchManagerClientsByName(
	ctx context.Context,
	userID, mcpTool, managerCustomerID string,
	in resolveCustomerInput,
) ([]customerMatch, error) {
	query := buildClientAccountsResolveQuery(in.accountName, in.matchMode)

	raw, err := s.googleSearch(ctx, userID, mcpTool, managerCustomerID, managerCustomerID, query)
	if err != nil {
		return nil, err
	}

	rows := extractSearchRows(raw)
	matches := make([]customerMatch, 0, len(rows))
	for _, row := range rows {
		client := nestedMap(row, "customerClient")
		id := stringFromMap(client, "id")
		name := stringFromMap(client, "descriptiveName")
		if id == "" {
			continue
		}
		matches = append(matches, customerMatch{
			CustomerID:      id,
			DescriptiveName: name,
			Manager:         false,
			LoginCustomerID: managerCustomerID,
			MatchType:       "manager_client",
		})
	}
	return matches, nil
}

func (s *service) fetchAccessibleAccountRecords(ctx context.Context, userID, mcpTool string) ([]accountRecord, bool, int, error) {
	raw, err := s.listAccessibleCustomers(ctx, userID, mcpTool)
	if err != nil {
		return nil, false, 0, err
	}

	ids := extractAccessibleCustomerIDs(raw)
	truncated := len(ids) > s.maxAccessibleAccounts
	if truncated {
		ids = ids[:s.maxAccessibleAccounts]
	}

	accounts := make([]accountRecord, 0, len(ids))
	skipped := 0
	for _, id := range ids {
		record, err := s.fetchCustomerRecord(ctx, userID, mcpTool, id)
		if err != nil {
			skipped++
			accounts = append(accounts, accountRecord{
				CustomerID:      id,
				DescriptiveName: "",
				Manager:         false,
				LoginCustomerID: nil,
			})
			continue
		}
		accounts = append(accounts, record)
	}
	return accounts, truncated, skipped, nil
}

func (s *service) fetchCustomerRecord(ctx context.Context, userID, mcpTool, customerID string) (accountRecord, error) {
	query := strings.Join([]string{
		"SELECT customer.id, customer.descriptive_name, customer.manager, customer.currency_code, customer.time_zone",
		"FROM customer",
	}, " ")
	raw, err := s.googleSearch(ctx, userID, mcpTool, customerID, "", query)
	if err != nil {
		return accountRecord{}, err
	}

	rows := extractSearchRows(raw)
	record := accountRecord{
		CustomerID:      customerID,
		DescriptiveName: "",
		Manager:         false,
		LoginCustomerID: nil,
	}
	if len(rows) == 0 {
		return record, nil
	}

	customer := nestedMap(rows[0], "customer")
	if id := stringFromMap(customer, "id"); id != "" {
		record.CustomerID = id
	}
	record.DescriptiveName = stringFromMap(customer, "descriptiveName")
	record.Manager = boolFromMap(customer, "manager")
	record.CurrencyCode = stringFromMap(customer, "currencyCode")
	record.TimeZone = stringFromMap(customer, "timeZone")
	return record, nil
}

func buildClientAccountsQuery(clientNameContains string) string {
	return buildCustomerClientQuery(googleLikeClause("customer_client.descriptive_name", clientNameContains))
}

func buildClientAccountsResolveQuery(accountName, matchMode string) string {
	nameFilter := googleLikeClause("customer_client.descriptive_name", accountName)
	if matchMode == "exact" {
		nameFilter = googleExactClause("customer_client.descriptive_name", accountName)
	}
	return buildCustomerClientQuery(nameFilter)
}

func buildCustomerClientQuery(nameFilter string) string {
	return strings.Join([]string{
		"SELECT customer_client.id, customer_client.descriptive_name, customer_client.currency_code, customer_client.time_zone, customer_client.manager",
		"FROM customer_client",
		googleBuildWhereClause([]string{
			"customer_client.manager = false",
			nameFilter,
		}),
		"ORDER BY customer_client.id",
	}, " ")
}

func extractAccessibleCustomerIDs(raw any) []string {
	root, ok := raw.(map[string]any)
	if !ok {
		return nil
	}
	names := googleToStringSlice(root["resourceNames"])
	ids := make([]string, 0, len(names))
	for _, name := range names {
		id := googleNormalizeCustomerID(name)
		if id != "" {
			ids = append(ids, id)
		}
	}
	return ids
}

func extractSearchRows(raw any) []map[string]any {
	root, ok := raw.(map[string]any)
	if !ok {
		return nil
	}
	items, ok := root["results"].([]any)
	if !ok {
		return nil
	}
	rows := make([]map[string]any, 0, len(items))
	for _, item := range items {
		if row, ok := item.(map[string]any); ok {
			rows = append(rows, row)
		}
	}
	return rows
}

func nestedMap(root map[string]any, key string) map[string]any {
	if root == nil {
		return map[string]any{}
	}
	if nested, ok := root[key].(map[string]any); ok {
		return nested
	}
	return map[string]any{}
}

func stringFromMap(root map[string]any, key string) string {
	if root == nil {
		return ""
	}
	switch v := root[key].(type) {
	case string:
		return v
	case float64:
		return fmt.Sprintf("%.0f", v)
	default:
		return ""
	}
}

func boolFromMap(root map[string]any, key string) bool {
	if root == nil {
		return false
	}
	if v, ok := root[key].(bool); ok {
		return v
	}
	return false
}
