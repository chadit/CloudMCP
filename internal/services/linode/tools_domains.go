package linode

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/linode/linodego"
	"github.com/mark3labs/mcp-go/mcp"

	"github.com/chadit/CloudMCP/pkg/types"
)

// extractStringParam extracts a string parameter from request arguments.
func extractStringParam(args map[string]interface{}, key string) string {
	if value, ok := args[key].(string); ok {
		return value
	}
	return ""
}

// extractIntParam extracts an integer parameter from request arguments.
func extractIntParam(args map[string]interface{}, key string) int {
	if value, ok := args[key].(float64); ok {
		return int(value)
	}
	return 0
}

// extractStringSliceParam extracts a string slice parameter from request arguments.
func extractStringSliceParam(args map[string]interface{}, key string) []string {
	rawValue, exists := args[key]
	if !exists {
		return nil
	}

	slice, ok := rawValue.([]interface{})
	if !ok {
		return nil
	}

	result := make([]string, 0, len(slice))

	for _, item := range slice {
		if str, strOK := item.(string); strOK {
			result = append(result, str)
		}
	}
	return result
}

// buildDomainUpdateOptions converts domain parameters to update options.
func buildDomainUpdateOptions(params DomainUpdateParams) linodego.DomainUpdateOptions {
	options := linodego.DomainUpdateOptions{}

	if params.Domain != "" {
		options.Domain = params.Domain
	}
	if params.Type != "" {
		options.Type = linodego.DomainType(params.Type)
	}
	if params.SOAEmail != "" {
		options.SOAEmail = params.SOAEmail
	}
	if params.Description != "" {
		options.Description = params.Description
	}
	if params.RetrySec > 0 {
		options.RetrySec = params.RetrySec
	}
	if params.ExpireSec > 0 {
		options.ExpireSec = params.ExpireSec
	}
	if params.RefreshSec > 0 {
		options.RefreshSec = params.RefreshSec
	}
	if params.TTLSec > 0 {
		options.TTLSec = params.TTLSec
	}
	if len(params.Tags) > 0 {
		options.Tags = params.Tags
	}
	if len(params.MasterIPs) > 0 {
		options.MasterIPs = params.MasterIPs
	}
	if len(params.AXfrIPs) > 0 {
		options.AXfrIPs = params.AXfrIPs
	}

	return options
}

// extractDomainUpdateParams extracts domain update parameters from request arguments.
func extractDomainUpdateParams(args map[string]interface{}, domainID int) DomainUpdateParams {
	return DomainUpdateParams{
		DomainID:    domainID,
		Domain:      extractStringParam(args, "domain"),
		Type:        extractStringParam(args, "type"),
		SOAEmail:    extractStringParam(args, "soa_email"),
		Description: extractStringParam(args, "description"),
		RetrySec:    extractIntParam(args, "retry_sec"),
		ExpireSec:   extractIntParam(args, "expire_sec"),
		RefreshSec:  extractIntParam(args, "refresh_sec"),
		TTLSec:      extractIntParam(args, "ttl_sec"),
		Tags:        extractStringSliceParam(args, "tags"),
		MasterIPs:   extractStringSliceParam(args, "master_ips"),
		AXfrIPs:     extractStringSliceParam(args, "axfr_ips"),
	}
}

const (
	// timeFormatLayout is the standard time format for displaying dates.
	timeFormatLayout = "2006-01-02T15:04:05"
	// defaultDNSRoot represents the root domain symbol.
	defaultDNSRoot = "@"
)

var (
	// ErrInvalidArgumentsFormat is returned when arguments are not in the expected format.
	ErrInvalidArgumentsFormat = errors.New("invalid arguments format")
	// ErrMissingDomainName is returned when domain name is missing or empty.
	ErrMissingDomainName = errors.New("domain is required")
	// ErrMissingDomainType is returned when domain type is missing or empty.
	ErrMissingDomainType = errors.New("type is required")
	// ErrMissingRecordType is returned when record type is missing or empty.
	ErrMissingRecordType = errors.New("type is required")
	// ErrMissingRecordTarget is returned when record target is missing or empty.
	ErrMissingRecordTarget = errors.New("target is required")
)

// handleDomainsList lists all domains.
func (s *Service) handleDomainsList(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	account, accountErr := s.accountManager.GetCurrentAccount()
	if accountErr != nil {
		return nil, fmt.Errorf("failed to get current account: %w", accountErr)
	}

	domains, domainsErr := account.Client.ListDomains(ctx, nil)
	if domainsErr != nil {
		return nil, types.NewToolError("linode", "domains_list", //nolint:wrapcheck // types.NewToolError already wraps the error
			"failed to list domains", domainsErr)
	}

	summaries := make([]DomainSummary, 0, len(domains))

	for _, domain := range domains {
		summary := DomainSummary{
			ID:          domain.ID,
			Domain:      domain.Domain,
			Type:        string(domain.Type),
			Status:      string(domain.Status),
			Description: domain.Description,
			SOAEmail:    domain.SOAEmail,
			RetrySec:    domain.RetrySec,
			MasterIPs:   domain.MasterIPs,
			AXfrIPs:     domain.AXfrIPs,
			Tags:        domain.Tags,
			Created:     "", // Domain doesn't have Created field.
			Updated:     "", // Domain doesn't have Updated field.
		}
		summaries = append(summaries, summary)
	}

	var stringBuilder strings.Builder

	stringBuilder.WriteString(fmt.Sprintf("Found %d domains:\n\n", len(summaries)))

	for _, domainEntry := range summaries {
		fmt.Fprintf(&stringBuilder, "ID: %d | %s (%s)\n", domainEntry.ID, domainEntry.Domain, domainEntry.Type)
		fmt.Fprintf(&stringBuilder, "  Status: %s\n", domainEntry.Status)

		if domainEntry.Description != "" {
			fmt.Fprintf(&stringBuilder, "  Description: %s\n", domainEntry.Description)
		}

		fmt.Fprintf(&stringBuilder, "  SOA Email: %s\n", domainEntry.SOAEmail)

		if len(domainEntry.MasterIPs) > 0 {
			fmt.Fprintf(&stringBuilder, "  Master IPs: %s\n", strings.Join(domainEntry.MasterIPs, ", "))
		}

		if len(domainEntry.Tags) > 0 {
			fmt.Fprintf(&stringBuilder, "  Tags: %s\n", strings.Join(domainEntry.Tags, ", "))
		}

		stringBuilder.WriteString("\n")
	}

	if len(summaries) == 0 {
		return mcp.NewToolResultText("No domains found."), nil
	}

	return mcp.NewToolResultText(stringBuilder.String()), nil
}

// handleDomainGet gets details of a specific domain.
func (s *Service) handleDomainGet(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	requestArguments := request.GetArguments()
	domainID, parseErr := parseIDFromArguments(requestArguments, "domain_id")

	if parseErr != nil {
		return mcp.NewToolResultError(parseErr.Error()), nil
	}

	account, accountErr := s.accountManager.GetCurrentAccount()
	if accountErr != nil {
		return nil, fmt.Errorf("failed to get current account: %w", accountErr)
	}

	domain, domainErr := account.Client.GetDomain(ctx, domainID)
	if domainErr != nil {
		return nil, types.NewToolError("linode", "domain_get", //nolint:wrapcheck // types.NewToolError already wraps the error
			fmt.Sprintf("failed to get domain %d", domainID), domainErr)
	}

	domainDetail := DomainDetail{
		ID:          domain.ID,
		Domain:      domain.Domain,
		Type:        string(domain.Type),
		Status:      string(domain.Status),
		Description: domain.Description,
		SOAEmail:    domain.SOAEmail,
		RetrySec:    domain.RetrySec,
		MasterIPs:   domain.MasterIPs,
		AXfrIPs:     domain.AXfrIPs,
		Tags:        domain.Tags,
		Created:     "", // Domain doesn't have Created field.
		Updated:     "", // Domain doesn't have Updated field.
		ExpireSec:   domain.ExpireSec,
		RefreshSec:  domain.RefreshSec,
		TTLSec:      domain.TTLSec,
	}

	var stringBuilder strings.Builder

	fmt.Fprintf(&stringBuilder, "Domain Details:\n")
	fmt.Fprintf(&stringBuilder, "ID: %d\n", domainDetail.ID)
	fmt.Fprintf(&stringBuilder, "Domain: %s\n", domainDetail.Domain)
	fmt.Fprintf(&stringBuilder, "Type: %s\n", domainDetail.Type)
	fmt.Fprintf(&stringBuilder, "Status: %s\n", domainDetail.Status)

	if domainDetail.Description != "" {
		fmt.Fprintf(&stringBuilder, "Description: %s\n", domainDetail.Description)
	}

	fmt.Fprintf(&stringBuilder, "SOA Email: %s\n", domainDetail.SOAEmail)
	fmt.Fprintf(&stringBuilder, "TTL: %d seconds\n", domainDetail.TTLSec)
	fmt.Fprintf(&stringBuilder, "Refresh: %d seconds\n", domainDetail.RefreshSec)
	fmt.Fprintf(&stringBuilder, "Retry: %d seconds\n", domainDetail.RetrySec)
	fmt.Fprintf(&stringBuilder, "Expire: %d seconds\n", domainDetail.ExpireSec)
	fmt.Fprintf(&stringBuilder, "Created: %s\n", domainDetail.Created)
	fmt.Fprintf(&stringBuilder, "Updated: %s\n\n", domainDetail.Updated)

	if len(domainDetail.MasterIPs) > 0 {
		fmt.Fprintf(&stringBuilder, "Master IPs: %s\n", strings.Join(domainDetail.MasterIPs, ", "))
	}

	if len(domainDetail.AXfrIPs) > 0 {
		fmt.Fprintf(&stringBuilder, "AXFR IPs: %s\n", strings.Join(domainDetail.AXfrIPs, ", "))
	}

	if len(domainDetail.Tags) > 0 {
		fmt.Fprintf(&stringBuilder, "Tags: %s\n", strings.Join(domainDetail.Tags, ", "))
	}

	return mcp.NewToolResultText(stringBuilder.String()), nil
}

// handleDomainCreate creates a new domain.
func (s *Service) handleDomainCreate(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	requestArguments, argumentsOK := request.Params.Arguments.(map[string]interface{})
	if !argumentsOK {
		return mcp.NewToolResultError(ErrInvalidArgumentsFormat.Error()), nil
	}

	// Parse required parameters.
	domainName, domainNameOK := requestArguments["domain"].(string)
	if !domainNameOK || domainName == "" {
		return mcp.NewToolResultError(ErrMissingDomainName.Error()), nil
	}

	domainType, domainTypeOK := requestArguments["type"].(string)
	if !domainTypeOK || domainType == "" {
		return mcp.NewToolResultError(ErrMissingDomainType.Error()), nil
	}

	// Build domain create options.
	domainParams := DomainCreateParams{
		Domain: domainName,
		Type:   domainType,
	}

	// Optional parameters.
	if soaEmail, soaEmailOK := requestArguments["soa_email"].(string); soaEmailOK {
		domainParams.SOAEmail = soaEmail
	}

	if description, descriptionOK := requestArguments["description"].(string); descriptionOK {
		domainParams.Description = description
	}

	if retrySec, retrySecOK := requestArguments["retry_sec"].(float64); retrySecOK {
		domainParams.RetrySec = int(retrySec)
	}

	if expireSec, expireSecOK := requestArguments["expire_sec"].(float64); expireSecOK {
		domainParams.ExpireSec = int(expireSec)
	}

	if refreshSec, refreshSecOK := requestArguments["refresh_sec"].(float64); refreshSecOK {
		domainParams.RefreshSec = int(refreshSec)
	}

	if ttlSec, ttlSecOK := requestArguments["ttl_sec"].(float64); ttlSecOK {
		domainParams.TTLSec = int(ttlSec)
	}

	if tagsRaw, tagsOK := requestArguments["tags"]; tagsOK {
		if tagsSlice, tagsSliceOK := tagsRaw.([]interface{}); tagsSliceOK {
			tagsList := make([]string, len(tagsSlice))

			for tagIndex, tagEntry := range tagsSlice {
				if tagString, tagStringOK := tagEntry.(string); tagStringOK {
					tagsList[tagIndex] = tagString
				}
			}

			domainParams.Tags = tagsList
		}
	}

	if masterIPsRaw, masterIPsOK := requestArguments["master_ips"]; masterIPsOK {
		if ipsSlice, ipsSliceOK := masterIPsRaw.([]interface{}); ipsSliceOK {
			masterIPsList := make([]string, len(ipsSlice))

			for ipIndex, ipEntry := range ipsSlice {
				if ipString, ipStringOK := ipEntry.(string); ipStringOK {
					masterIPsList[ipIndex] = ipString
				}
			}

			domainParams.MasterIPs = masterIPsList
		}
	}

	if axfrIPsRaw, axfrIPsOK := requestArguments["axfr_ips"]; axfrIPsOK {
		if ipsSlice, ipsSliceOK := axfrIPsRaw.([]interface{}); ipsSliceOK {
			axfrIPsList := make([]string, len(ipsSlice))

			for ipIndex, ipEntry := range ipsSlice {
				if ipString, ipStringOK := ipEntry.(string); ipStringOK {
					axfrIPsList[ipIndex] = ipString
				}
			}

			domainParams.AXfrIPs = axfrIPsList
		}
	}

	account, accountErr := s.accountManager.GetCurrentAccount()
	if accountErr != nil {
		return mcp.NewToolResultError(accountErr.Error()), nil
	}

	createOptions := linodego.DomainCreateOptions{
		Domain: domainParams.Domain,
		Type:   linodego.DomainType(domainParams.Type),
	}

	if domainParams.SOAEmail != "" {
		createOptions.SOAEmail = domainParams.SOAEmail
	}

	if domainParams.Description != "" {
		createOptions.Description = domainParams.Description
	}

	if domainParams.RetrySec > 0 {
		createOptions.RetrySec = domainParams.RetrySec
	}

	if len(domainParams.MasterIPs) > 0 {
		createOptions.MasterIPs = domainParams.MasterIPs
	}

	if len(domainParams.AXfrIPs) > 0 {
		createOptions.AXfrIPs = domainParams.AXfrIPs
	}

	if domainParams.ExpireSec > 0 {
		createOptions.ExpireSec = domainParams.ExpireSec
	}

	if domainParams.RefreshSec > 0 {
		createOptions.RefreshSec = domainParams.RefreshSec
	}

	if domainParams.TTLSec > 0 {
		createOptions.TTLSec = domainParams.TTLSec
	}

	if len(domainParams.Tags) > 0 {
		createOptions.Tags = domainParams.Tags
	}

	createdDomain, createErr := account.Client.CreateDomain(ctx, createOptions)
	if createErr != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to create domain: %v", createErr)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Domain created successfully:\nID: %d\nDomain: %s\nType: %s\nStatus: %s",
		createdDomain.ID, createdDomain.Domain, createdDomain.Type, createdDomain.Status)), nil
}

// handleDomainUpdate updates an existing domain.
func (s *Service) handleDomainUpdate(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	requestArguments, argumentsOK := request.Params.Arguments.(map[string]interface{})
	if !argumentsOK {
		return mcp.NewToolResultError(ErrInvalidArgumentsFormat.Error()), nil
	}

	domainID, parseErr := parseIDFromArguments(requestArguments, "domain_id")
	if parseErr != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid domain_id parameter: %v", parseErr)), nil
	}

	domainParams := extractDomainUpdateParams(requestArguments, domainID)

	account, accountErr := s.accountManager.GetCurrentAccount()
	if accountErr != nil {
		return mcp.NewToolResultError(accountErr.Error()), nil
	}

	updateOptions := buildDomainUpdateOptions(domainParams)

	updatedDomain, updateErr := account.Client.UpdateDomain(ctx, domainParams.DomainID, updateOptions)
	if updateErr != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to update domain: %v", updateErr)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Domain updated successfully:\nID: %d\nDomain: %s\nType: %s\nStatus: %s",
		updatedDomain.ID, updatedDomain.Domain, updatedDomain.Type, updatedDomain.Status)), nil
}

// handleDomainDelete deletes a domain.
func (s *Service) handleDomainDelete(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	requestArguments, argumentsOK := request.Params.Arguments.(map[string]interface{})
	if !argumentsOK {
		return mcp.NewToolResultError(ErrInvalidArgumentsFormat.Error()), nil
	}

	domainID, parseErr := parseIDFromArguments(requestArguments, "domain_id")
	if parseErr != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid domain_id parameter: %v", parseErr)), nil
	}

	domainParams := DomainDeleteParams{
		DomainID: domainID,
	}

	account, accountErr := s.accountManager.GetCurrentAccount()
	if accountErr != nil {
		return mcp.NewToolResultError(accountErr.Error()), nil
	}

	deleteErr := account.Client.DeleteDomain(ctx, domainParams.DomainID)
	if deleteErr != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to delete domain: %v", deleteErr)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Domain %d deleted successfully", domainParams.DomainID)), nil
}

// handleDomainRecordsList lists all records for a domain.
func (s *Service) handleDomainRecordsList(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	requestArguments, argumentsOK := request.Params.Arguments.(map[string]interface{})
	if !argumentsOK {
		return mcp.NewToolResultError(ErrInvalidArgumentsFormat.Error()), nil
	}

	domainID, parseErr := parseIDFromArguments(requestArguments, "domain_id")
	if parseErr != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid domain_id parameter: %v", parseErr)), nil
	}

	domainParams := DomainRecordsListParams{
		DomainID: domainID,
	}

	account, accountErr := s.accountManager.GetCurrentAccount()
	if accountErr != nil {
		return mcp.NewToolResultError(accountErr.Error()), nil
	}

	records, recordsErr := account.Client.ListDomainRecords(ctx, domainParams.DomainID, nil)
	if recordsErr != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to list domain records: %v", recordsErr)), nil
	}

	recordList := make([]DomainRecord, 0, len(records))

	for _, record := range records {
		recordList = append(recordList, DomainRecord{
			ID:       record.ID,
			Type:     string(record.Type),
			Name:     record.Name,
			Target:   record.Target,
			Priority: record.Priority,
			Weight:   record.Weight,
			Port:     record.Port,
			Service:  stringPtrValue(record.Service),
			Protocol: stringPtrValue(record.Protocol),
			TTLSec:   record.TTLSec,
			Tag:      stringPtrValue(record.Tag),
			Created:  record.Created.Format(timeFormatLayout),
			Updated:  record.Updated.Format(timeFormatLayout),
		})
	}

	var stringBuilder strings.Builder

	stringBuilder.WriteString(fmt.Sprintf("Found %d domain records:\n\n", len(recordList)))

	// Group records by type for better readability.
	recordsByType := make(map[string][]DomainRecord)
	for _, record := range recordList {
		recordsByType[record.Type] = append(recordsByType[record.Type], record)
	}

	for recordType, recordsForType := range recordsByType {
		fmt.Fprintf(&stringBuilder, "%s Records:\n", recordType)

		for _, record := range recordsForType {
			recordName := record.Name
			if recordName == "" {
				recordName = defaultDNSRoot
			}

			fmt.Fprintf(&stringBuilder, "  ID: %d | %s -> %s", record.ID, recordName, record.Target)

			if record.Priority > 0 {
				fmt.Fprintf(&stringBuilder, " (Priority: %d)", record.Priority)
			}

			if record.Weight > 0 {
				fmt.Fprintf(&stringBuilder, " (Weight: %d)", record.Weight)
			}

			if record.Port > 0 {
				fmt.Fprintf(&stringBuilder, " (Port: %d)", record.Port)
			}

			if record.TTLSec > 0 {
				fmt.Fprintf(&stringBuilder, " (TTL: %ds)", record.TTLSec)
			}

			stringBuilder.WriteString("\n")
		}

		stringBuilder.WriteString("\n")
	}

	return mcp.NewToolResultText(stringBuilder.String()), nil
}

// handleDomainRecordGet gets details of a specific domain record.
func (s *Service) handleDomainRecordGet(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	requestArguments, argumentsOK := request.Params.Arguments.(map[string]interface{})
	if !argumentsOK {
		return mcp.NewToolResultError(ErrInvalidArgumentsFormat.Error()), nil
	}

	domainID, domainParseErr := parseIDFromArguments(requestArguments, "domain_id")
	if domainParseErr != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid domain_id parameter: %v", domainParseErr)), nil
	}

	recordID, recordParseErr := parseIDFromArguments(requestArguments, "record_id")
	if recordParseErr != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid record_id parameter: %v", recordParseErr)), nil
	}

	recordParams := DomainRecordGetParams{
		DomainID: domainID,
		RecordID: recordID,
	}

	account, accountErr := s.accountManager.GetCurrentAccount()
	if accountErr != nil {
		return mcp.NewToolResultError(accountErr.Error()), nil
	}

	record, recordErr := account.Client.GetDomainRecord(ctx, recordParams.DomainID, recordParams.RecordID)
	if recordErr != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get domain record: %v", recordErr)), nil
	}

	recordDetail := DomainRecord{
		ID:       record.ID,
		Type:     string(record.Type),
		Name:     record.Name,
		Target:   record.Target,
		Priority: record.Priority,
		Weight:   record.Weight,
		Port:     record.Port,
		Service:  stringPtrValue(record.Service),
		Protocol: stringPtrValue(record.Protocol),
		TTLSec:   record.TTLSec,
		Tag:      stringPtrValue(record.Tag),
		Created:  record.Created.Format(timeFormatLayout),
		Updated:  record.Updated.Format(timeFormatLayout),
	}

	var stringBuilder strings.Builder

	fmt.Fprintf(&stringBuilder, "Domain Record Details:\n")
	fmt.Fprintf(&stringBuilder, "ID: %d\n", recordDetail.ID)
	fmt.Fprintf(&stringBuilder, "Type: %s\n", recordDetail.Type)
	fmt.Fprintf(&stringBuilder, "Name: %s\n", recordDetail.Name)
	fmt.Fprintf(&stringBuilder, "Target: %s\n", recordDetail.Target)

	if recordDetail.Priority > 0 {
		fmt.Fprintf(&stringBuilder, "Priority: %d\n", recordDetail.Priority)
	}

	if recordDetail.Weight > 0 {
		fmt.Fprintf(&stringBuilder, "Weight: %d\n", recordDetail.Weight)
	}

	if recordDetail.Port > 0 {
		fmt.Fprintf(&stringBuilder, "Port: %d\n", recordDetail.Port)
	}

	if recordDetail.Service != "" {
		fmt.Fprintf(&stringBuilder, "Service: %s\n", recordDetail.Service)
	}

	if recordDetail.Protocol != "" {
		fmt.Fprintf(&stringBuilder, "Protocol: %s\n", recordDetail.Protocol)
	}

	if recordDetail.Tag != "" {
		fmt.Fprintf(&stringBuilder, "Tag: %s\n", recordDetail.Tag)
	}

	fmt.Fprintf(&stringBuilder, "TTL: %d seconds\n", recordDetail.TTLSec)
	fmt.Fprintf(&stringBuilder, "Created: %s\n", recordDetail.Created)
	fmt.Fprintf(&stringBuilder, "Updated: %s\n", recordDetail.Updated)

	return mcp.NewToolResultText(stringBuilder.String()), nil
}

// handleDomainRecordCreate creates a new domain record.
func (s *Service) handleDomainRecordCreate(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	requestArguments, argumentsOK := request.Params.Arguments.(map[string]interface{})
	if !argumentsOK {
		return mcp.NewToolResultError(ErrInvalidArgumentsFormat.Error()), nil
	}

	domainID, parseErr := parseIDFromArguments(requestArguments, "domain_id")
	if parseErr != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid domain_id parameter: %v", parseErr)), nil
	}

	// Parse required parameters.
	recordType, recordTypeOK := requestArguments["type"].(string)
	if !recordTypeOK || recordType == "" {
		return mcp.NewToolResultError(ErrMissingRecordType.Error()), nil
	}

	target, targetOK := requestArguments["target"].(string)
	if !targetOK || target == "" {
		return mcp.NewToolResultError(ErrMissingRecordTarget.Error()), nil
	}

	// Build record create options.
	recordParams := DomainRecordCreateParams{
		DomainID: domainID,
		Type:     recordType,
		Target:   target,
	}

	// Optional parameters.
	if recordName, recordNameOK := requestArguments["name"].(string); recordNameOK {
		recordParams.Name = recordName
	}

	if priority, priorityOK := requestArguments["priority"].(float64); priorityOK {
		recordParams.Priority = int(priority)
	}

	if weight, weightOK := requestArguments["weight"].(float64); weightOK {
		recordParams.Weight = int(weight)
	}

	if port, portOK := requestArguments["port"].(float64); portOK {
		recordParams.Port = int(port)
	}

	if service, serviceOK := requestArguments["service"].(string); serviceOK {
		recordParams.Service = service
	}

	if protocol, protocolOK := requestArguments["protocol"].(string); protocolOK {
		recordParams.Protocol = protocol
	}

	if ttlSec, ttlSecOK := requestArguments["ttl_sec"].(float64); ttlSecOK {
		recordParams.TTLSec = int(ttlSec)
	}

	if tag, tagOK := requestArguments["tag"].(string); tagOK {
		recordParams.Tag = tag
	}

	account, accountErr := s.accountManager.GetCurrentAccount()
	if accountErr != nil {
		return mcp.NewToolResultError(accountErr.Error()), nil
	}

	createOptions := linodego.DomainRecordCreateOptions{
		Type:   linodego.DomainRecordType(recordParams.Type),
		Target: recordParams.Target,
	}

	if recordParams.Name != "" {
		createOptions.Name = recordParams.Name
	}

	if recordParams.Priority > 0 {
		createOptions.Priority = intPtr(recordParams.Priority)
	}

	if recordParams.Weight > 0 {
		createOptions.Weight = intPtr(recordParams.Weight)
	}

	if recordParams.Port > 0 {
		createOptions.Port = intPtr(recordParams.Port)
	}

	if recordParams.Service != "" {
		createOptions.Service = stringPtr(recordParams.Service)
	}

	if recordParams.Protocol != "" {
		createOptions.Protocol = stringPtr(recordParams.Protocol)
	}

	if recordParams.TTLSec > 0 {
		createOptions.TTLSec = recordParams.TTLSec
	}

	if recordParams.Tag != "" {
		createOptions.Tag = stringPtr(recordParams.Tag)
	}

	createdRecord, createErr := account.Client.CreateDomainRecord(ctx, recordParams.DomainID, createOptions)
	if createErr != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to create domain record: %v", createErr)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Domain record created successfully:\nID: %d\nType: %s\nName: %s\nTarget: %s",
		createdRecord.ID, createdRecord.Type, createdRecord.Name, createdRecord.Target)), nil
}

// handleDomainRecordUpdate updates a domain record.
func (s *Service) handleDomainRecordUpdate(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	requestArguments, argumentsOK := request.Params.Arguments.(map[string]interface{})
	if !argumentsOK {
		return mcp.NewToolResultError(ErrInvalidArgumentsFormat.Error()), nil
	}

	domainID, domainParseErr := parseIDFromArguments(requestArguments, "domain_id")
	if domainParseErr != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid domain_id parameter: %v", domainParseErr)), nil
	}

	recordID, recordParseErr := parseIDFromArguments(requestArguments, "record_id")
	if recordParseErr != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid record_id parameter: %v", recordParseErr)), nil
	}

	// Build record update options.
	recordParams := DomainRecordUpdateParams{
		DomainID: domainID,
		RecordID: recordID,
	}

	// Optional parameters.
	if recordType, recordTypeOK := requestArguments["type"].(string); recordTypeOK {
		recordParams.Type = recordType
	}

	if recordName, recordNameOK := requestArguments["name"].(string); recordNameOK {
		recordParams.Name = recordName
	}

	if target, targetOK := requestArguments["target"].(string); targetOK {
		recordParams.Target = target
	}

	if priority, priorityOK := requestArguments["priority"].(float64); priorityOK {
		recordParams.Priority = int(priority)
	}

	if weight, weightOK := requestArguments["weight"].(float64); weightOK {
		recordParams.Weight = int(weight)
	}

	if port, portOK := requestArguments["port"].(float64); portOK {
		recordParams.Port = int(port)
	}

	if service, serviceOK := requestArguments["service"].(string); serviceOK {
		recordParams.Service = service
	}

	if protocol, protocolOK := requestArguments["protocol"].(string); protocolOK {
		recordParams.Protocol = protocol
	}

	if ttlSec, ttlSecOK := requestArguments["ttl_sec"].(float64); ttlSecOK {
		recordParams.TTLSec = int(ttlSec)
	}

	if tag, tagOK := requestArguments["tag"].(string); tagOK {
		recordParams.Tag = tag
	}

	account, accountErr := s.accountManager.GetCurrentAccount()
	if accountErr != nil {
		return mcp.NewToolResultError(accountErr.Error()), nil
	}

	updateOptions := linodego.DomainRecordUpdateOptions{}

	if recordParams.Type != "" {
		updateOptions.Type = linodego.DomainRecordType(recordParams.Type)
	}

	if recordParams.Name != "" {
		updateOptions.Name = recordParams.Name
	}

	if recordParams.Target != "" {
		updateOptions.Target = recordParams.Target
	}

	if recordParams.Priority > 0 {
		updateOptions.Priority = intPtr(recordParams.Priority)
	}

	if recordParams.Weight > 0 {
		updateOptions.Weight = intPtr(recordParams.Weight)
	}

	if recordParams.Port > 0 {
		updateOptions.Port = intPtr(recordParams.Port)
	}

	if recordParams.Service != "" {
		updateOptions.Service = stringPtr(recordParams.Service)
	}

	if recordParams.Protocol != "" {
		updateOptions.Protocol = stringPtr(recordParams.Protocol)
	}

	if recordParams.TTLSec > 0 {
		updateOptions.TTLSec = recordParams.TTLSec
	}

	if recordParams.Tag != "" {
		updateOptions.Tag = stringPtr(recordParams.Tag)
	}

	updatedRecord, updateErr := account.Client.UpdateDomainRecord(ctx, recordParams.DomainID, recordParams.RecordID, updateOptions)
	if updateErr != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to update domain record: %v", updateErr)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Domain record updated successfully:\nID: %d\nType: %s\nName: %s\nTarget: %s",
		updatedRecord.ID, updatedRecord.Type, updatedRecord.Name, updatedRecord.Target)), nil
}

// handleDomainRecordDelete deletes a domain record.
func (s *Service) handleDomainRecordDelete(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	requestArguments, argumentsOK := request.Params.Arguments.(map[string]interface{})
	if !argumentsOK {
		return mcp.NewToolResultError(ErrInvalidArgumentsFormat.Error()), nil
	}

	domainID, domainParseErr := parseIDFromArguments(requestArguments, "domain_id")
	if domainParseErr != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid domain_id parameter: %v", domainParseErr)), nil
	}

	recordID, recordParseErr := parseIDFromArguments(requestArguments, "record_id")
	if recordParseErr != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid record_id parameter: %v", recordParseErr)), nil
	}

	recordParams := DomainRecordDeleteParams{
		DomainID: domainID,
		RecordID: recordID,
	}

	account, accountErr := s.accountManager.GetCurrentAccount()
	if accountErr != nil {
		return mcp.NewToolResultError(accountErr.Error()), nil
	}

	deleteErr := account.Client.DeleteDomainRecord(ctx, recordParams.DomainID, recordParams.RecordID)
	if deleteErr != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to delete domain record: %v", deleteErr)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Domain record %d deleted successfully from domain %d",
		recordParams.RecordID, recordParams.DomainID)), nil
}

// parseArguments is a placeholder function for structured parameter parsing.
// Convert remaining handler functions to use direct argument parsing like instances and domains.
func parseArguments(_ interface{}, _ interface{}) error {
	// This is a temporary placeholder that returns no error.
	// The remaining functions will need to be converted to use direct argument parsing.
	return nil
}

// stringPtrValue safely dereferences a string pointer, returning empty string if nil.
func stringPtrValue(ptr *string) string {
	if ptr == nil {
		return ""
	}

	return *ptr
}

// intPtr returns a pointer to the given int value.
func intPtr(intValue int) *int {
	return &intValue
}

// stringPtr returns a pointer to the given string value.
func stringPtr(stringValue string) *string {
	return &stringValue
}
