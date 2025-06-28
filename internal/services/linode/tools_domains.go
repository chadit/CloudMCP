package linode

import (
	"context"
	"fmt"
	"strings"

	"github.com/linode/linodego"
	"github.com/mark3labs/mcp-go/mcp"

	"github.com/chadit/CloudMCP/pkg/types"
)

// handleDomainsList lists all domains.
func (s *Service) handleDomainsList(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	account, err := s.accountManager.GetCurrentAccount()
	if err != nil {
		return nil, err
	}

	domains, err := account.Client.ListDomains(ctx, nil)
	if err != nil {
		return nil, types.NewToolError("linode", "domains_list", //nolint:wrapcheck // types.NewToolError already wraps the error
			"failed to list domains", err)
	}

	var summaries []DomainSummary
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
			Created:     "", // Domain doesn't have Created field
			Updated:     "", // Domain doesn't have Updated field
		}
		summaries = append(summaries, summary)
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d domains:\n\n", len(summaries)))

	for _, domain := range summaries {
		fmt.Fprintf(&sb, "ID: %d | %s (%s)\n", domain.ID, domain.Domain, domain.Type)
		fmt.Fprintf(&sb, "  Status: %s\n", domain.Status)
		if domain.Description != "" {
			fmt.Fprintf(&sb, "  Description: %s\n", domain.Description)
		}
		fmt.Fprintf(&sb, "  SOA Email: %s\n", domain.SOAEmail)
		if len(domain.MasterIPs) > 0 {
			fmt.Fprintf(&sb, "  Master IPs: %s\n", strings.Join(domain.MasterIPs, ", "))
		}
		if len(domain.Tags) > 0 {
			fmt.Fprintf(&sb, "  Tags: %s\n", strings.Join(domain.Tags, ", "))
		}
		sb.WriteString("\n")
	}

	if len(summaries) == 0 {
		return mcp.NewToolResultText("No domains found."), nil
	}

	return mcp.NewToolResultText(sb.String()), nil
}

// handleDomainGet gets details of a specific domain.
func (s *Service) handleDomainGet(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	arguments := request.GetArguments()
	domainID, err := parseIDFromArguments(arguments, "domain_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	account, err := s.accountManager.GetCurrentAccount()
	if err != nil {
		return nil, err
	}

	domain, err := account.Client.GetDomain(ctx, domainID)
	if err != nil {
		return nil, types.NewToolError("linode", "domain_get", //nolint:wrapcheck // types.NewToolError already wraps the error
			fmt.Sprintf("failed to get domain %d", domainID), err)
	}

	detail := DomainDetail{
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
		Created:     "", // Domain doesn't have Created field
		Updated:     "", // Domain doesn't have Updated field
		ExpireSec:   domain.ExpireSec,
		RefreshSec:  domain.RefreshSec,
		TTLSec:      domain.TTLSec,
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "Domain Details:\n")
	fmt.Fprintf(&sb, "ID: %d\n", detail.ID)
	fmt.Fprintf(&sb, "Domain: %s\n", detail.Domain)
	fmt.Fprintf(&sb, "Type: %s\n", detail.Type)
	fmt.Fprintf(&sb, "Status: %s\n", detail.Status)
	if detail.Description != "" {
		fmt.Fprintf(&sb, "Description: %s\n", detail.Description)
	}
	fmt.Fprintf(&sb, "SOA Email: %s\n", detail.SOAEmail)
	fmt.Fprintf(&sb, "TTL: %d seconds\n", detail.TTLSec)
	fmt.Fprintf(&sb, "Refresh: %d seconds\n", detail.RefreshSec)
	fmt.Fprintf(&sb, "Retry: %d seconds\n", detail.RetrySec)
	fmt.Fprintf(&sb, "Expire: %d seconds\n", detail.ExpireSec)
	fmt.Fprintf(&sb, "Created: %s\n", detail.Created)
	fmt.Fprintf(&sb, "Updated: %s\n\n", detail.Updated)

	if len(detail.MasterIPs) > 0 {
		fmt.Fprintf(&sb, "Master IPs: %s\n", strings.Join(detail.MasterIPs, ", "))
	}
	if len(detail.AXfrIPs) > 0 {
		fmt.Fprintf(&sb, "AXFR IPs: %s\n", strings.Join(detail.AXfrIPs, ", "))
	}
	if len(detail.Tags) > 0 {
		fmt.Fprintf(&sb, "Tags: %s\n", strings.Join(detail.Tags, ", "))
	}

	return mcp.NewToolResultText(sb.String()), nil
}

// handleDomainCreate creates a new domain.
func (s *Service) handleDomainCreate(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	arguments, ok := request.Params.Arguments.(map[string]interface{})
	if !ok {
		return mcp.NewToolResultError("Invalid arguments format"), nil
	}

	// Parse required parameters
	domainName, ok := arguments["domain"].(string)
	if !ok || domainName == "" {
		return mcp.NewToolResultError("domain is required"), nil
	}

	domainType, ok := arguments["type"].(string)
	if !ok || domainType == "" {
		return mcp.NewToolResultError("type is required"), nil
	}

	// Build domain create options
	params := DomainCreateParams{
		Domain: domainName,
		Type:   domainType,
	}

	// Optional parameters
	if soaEmail, ok := arguments["soa_email"].(string); ok {
		params.SOAEmail = soaEmail
	}
	if description, ok := arguments["description"].(string); ok {
		params.Description = description
	}
	if retrySec, ok := arguments["retry_sec"].(float64); ok {
		params.RetrySec = int(retrySec)
	}
	if expireSec, ok := arguments["expire_sec"].(float64); ok {
		params.ExpireSec = int(expireSec)
	}
	if refreshSec, ok := arguments["refresh_sec"].(float64); ok {
		params.RefreshSec = int(refreshSec)
	}
	if ttlSec, ok := arguments["ttl_sec"].(float64); ok {
		params.TTLSec = int(ttlSec)
	}
	if tagsRaw, ok := arguments["tags"]; ok {
		if tagsSlice, ok := tagsRaw.([]interface{}); ok {
			tags := make([]string, len(tagsSlice))
			for i, tag := range tagsSlice {
				if tagStr, ok := tag.(string); ok {
					tags[i] = tagStr
				}
			}
			params.Tags = tags
		}
	}
	if masterIPsRaw, ok := arguments["master_ips"]; ok {
		if ipsSlice, ok := masterIPsRaw.([]interface{}); ok {
			ips := make([]string, len(ipsSlice))
			for i, ip := range ipsSlice {
				if ipStr, ok := ip.(string); ok {
					ips[i] = ipStr
				}
			}
			params.MasterIPs = ips
		}
	}
	if axfrIPsRaw, ok := arguments["axfr_ips"]; ok {
		if ipsSlice, ok := axfrIPsRaw.([]interface{}); ok {
			ips := make([]string, len(ipsSlice))
			for i, ip := range ipsSlice {
				if ipStr, ok := ip.(string); ok {
					ips[i] = ipStr
				}
			}
			params.AXfrIPs = ips
		}
	}

	account, err := s.accountManager.GetCurrentAccount()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	createOpts := linodego.DomainCreateOptions{
		Domain: params.Domain,
		Type:   linodego.DomainType(params.Type),
	}

	if params.SOAEmail != "" {
		createOpts.SOAEmail = params.SOAEmail
	}
	if params.Description != "" {
		createOpts.Description = params.Description
	}
	if params.RetrySec > 0 {
		createOpts.RetrySec = params.RetrySec
	}
	if len(params.MasterIPs) > 0 {
		createOpts.MasterIPs = params.MasterIPs
	}
	if len(params.AXfrIPs) > 0 {
		createOpts.AXfrIPs = params.AXfrIPs
	}
	if params.ExpireSec > 0 {
		createOpts.ExpireSec = params.ExpireSec
	}
	if params.RefreshSec > 0 {
		createOpts.RefreshSec = params.RefreshSec
	}
	if params.TTLSec > 0 {
		createOpts.TTLSec = params.TTLSec
	}
	if len(params.Tags) > 0 {
		createOpts.Tags = params.Tags
	}

	createdDomain, err := account.Client.CreateDomain(ctx, createOpts)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to create domain: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Domain created successfully:\nID: %d\nDomain: %s\nType: %s\nStatus: %s",
		createdDomain.ID, createdDomain.Domain, createdDomain.Type, createdDomain.Status)), nil
}

// handleDomainUpdate updates an existing domain.
func (s *Service) handleDomainUpdate(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	arguments, ok := request.Params.Arguments.(map[string]interface{})
	if !ok {
		return mcp.NewToolResultError("Invalid arguments format"), nil
	}

	domainID, err := parseIDFromArguments(arguments, "domain_id")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid domain_id parameter: %v", err)), nil
	}

	// Build domain update options
	params := DomainUpdateParams{
		DomainID: domainID,
	}

	// Optional parameters
	if domain, ok := arguments["domain"].(string); ok {
		params.Domain = domain
	}
	if domainType, ok := arguments["type"].(string); ok {
		params.Type = domainType
	}
	if soaEmail, ok := arguments["soa_email"].(string); ok {
		params.SOAEmail = soaEmail
	}
	if description, ok := arguments["description"].(string); ok {
		params.Description = description
	}
	if retrySec, ok := arguments["retry_sec"].(float64); ok {
		params.RetrySec = int(retrySec)
	}
	if expireSec, ok := arguments["expire_sec"].(float64); ok {
		params.ExpireSec = int(expireSec)
	}
	if refreshSec, ok := arguments["refresh_sec"].(float64); ok {
		params.RefreshSec = int(refreshSec)
	}
	if ttlSec, ok := arguments["ttl_sec"].(float64); ok {
		params.TTLSec = int(ttlSec)
	}
	if tagsRaw, ok := arguments["tags"]; ok {
		if tagsSlice, ok := tagsRaw.([]interface{}); ok {
			tags := make([]string, len(tagsSlice))
			for i, tag := range tagsSlice {
				if tagStr, ok := tag.(string); ok {
					tags[i] = tagStr
				}
			}
			params.Tags = tags
		}
	}
	if masterIPsRaw, ok := arguments["master_ips"]; ok {
		if ipsSlice, ok := masterIPsRaw.([]interface{}); ok {
			ips := make([]string, len(ipsSlice))
			for i, ip := range ipsSlice {
				if ipStr, ok := ip.(string); ok {
					ips[i] = ipStr
				}
			}
			params.MasterIPs = ips
		}
	}
	if axfrIPsRaw, ok := arguments["axfr_ips"]; ok {
		if ipsSlice, ok := axfrIPsRaw.([]interface{}); ok {
			ips := make([]string, len(ipsSlice))
			for i, ip := range ipsSlice {
				if ipStr, ok := ip.(string); ok {
					ips[i] = ipStr
				}
			}
			params.AXfrIPs = ips
		}
	}

	account, err := s.accountManager.GetCurrentAccount()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	updateOpts := linodego.DomainUpdateOptions{}

	if params.Domain != "" {
		updateOpts.Domain = params.Domain
	}
	if params.Type != "" {
		updateOpts.Type = linodego.DomainType(params.Type)
	}
	if params.SOAEmail != "" {
		updateOpts.SOAEmail = params.SOAEmail
	}
	if params.Description != "" {
		updateOpts.Description = params.Description
	}
	if params.RetrySec > 0 {
		updateOpts.RetrySec = params.RetrySec
	}
	if len(params.MasterIPs) > 0 {
		updateOpts.MasterIPs = params.MasterIPs
	}
	if len(params.AXfrIPs) > 0 {
		updateOpts.AXfrIPs = params.AXfrIPs
	}
	if params.ExpireSec > 0 {
		updateOpts.ExpireSec = params.ExpireSec
	}
	if params.RefreshSec > 0 {
		updateOpts.RefreshSec = params.RefreshSec
	}
	if params.TTLSec > 0 {
		updateOpts.TTLSec = params.TTLSec
	}
	if len(params.Tags) > 0 {
		updateOpts.Tags = params.Tags
	}

	domain, err := account.Client.UpdateDomain(ctx, params.DomainID, updateOpts)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to update domain: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Domain updated successfully:\nID: %d\nDomain: %s\nType: %s\nStatus: %s",
		domain.ID, domain.Domain, domain.Type, domain.Status)), nil
}

// handleDomainDelete deletes a domain.
func (s *Service) handleDomainDelete(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	arguments, ok := request.Params.Arguments.(map[string]interface{})
	if !ok {
		return mcp.NewToolResultError("Invalid arguments format"), nil
	}

	domainID, err := parseIDFromArguments(arguments, "domain_id")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid domain_id parameter: %v", err)), nil
	}

	params := DomainDeleteParams{
		DomainID: domainID,
	}

	account, err := s.accountManager.GetCurrentAccount()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	err = account.Client.DeleteDomain(ctx, params.DomainID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to delete domain: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Domain %d deleted successfully", params.DomainID)), nil
}

// handleDomainRecordsList lists all records for a domain.
func (s *Service) handleDomainRecordsList(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	arguments, ok := request.Params.Arguments.(map[string]interface{})
	if !ok {
		return mcp.NewToolResultError("Invalid arguments format"), nil
	}

	domainID, err := parseIDFromArguments(arguments, "domain_id")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid domain_id parameter: %v", err)), nil
	}

	params := DomainRecordsListParams{
		DomainID: domainID,
	}

	account, err := s.accountManager.GetCurrentAccount()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	records, err := account.Client.ListDomainRecords(ctx, params.DomainID, nil)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to list domain records: %v", err)), nil
	}

	var recordList []DomainRecord
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
			Created:  record.Created.Format("2006-01-02T15:04:05"),
			Updated:  record.Updated.Format("2006-01-02T15:04:05"),
		})
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d domain records:\n\n", len(recordList)))

	// Group records by type for better readability
	recordsByType := make(map[string][]DomainRecord)
	for _, record := range recordList {
		recordsByType[record.Type] = append(recordsByType[record.Type], record)
	}

	for recordType, records := range recordsByType {
		fmt.Fprintf(&sb, "%s Records:\n", recordType)
		for _, record := range records {
			name := record.Name
			if name == "" {
				name = "@"
			}
			fmt.Fprintf(&sb, "  ID: %d | %s -> %s", record.ID, name, record.Target)
			if record.Priority > 0 {
				fmt.Fprintf(&sb, " (Priority: %d)", record.Priority)
			}
			if record.Weight > 0 {
				fmt.Fprintf(&sb, " (Weight: %d)", record.Weight)
			}
			if record.Port > 0 {
				fmt.Fprintf(&sb, " (Port: %d)", record.Port)
			}
			if record.TTLSec > 0 {
				fmt.Fprintf(&sb, " (TTL: %ds)", record.TTLSec)
			}
			sb.WriteString("\n")
		}
		sb.WriteString("\n")
	}

	return mcp.NewToolResultText(sb.String()), nil
}

// handleDomainRecordGet gets details of a specific domain record.
func (s *Service) handleDomainRecordGet(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	arguments, ok := request.Params.Arguments.(map[string]interface{})
	if !ok {
		return mcp.NewToolResultError("Invalid arguments format"), nil
	}

	domainID, err := parseIDFromArguments(arguments, "domain_id")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid domain_id parameter: %v", err)), nil
	}

	recordID, err := parseIDFromArguments(arguments, "record_id")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid record_id parameter: %v", err)), nil
	}

	params := DomainRecordGetParams{
		DomainID: domainID,
		RecordID: recordID,
	}

	account, err := s.accountManager.GetCurrentAccount()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	record, err := account.Client.GetDomainRecord(ctx, params.DomainID, params.RecordID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get domain record: %v", err)), nil
	}

	detail := DomainRecord{
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
		Created:  record.Created.Format("2006-01-02T15:04:05"),
		Updated:  record.Updated.Format("2006-01-02T15:04:05"),
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "Domain Record Details:\n")
	fmt.Fprintf(&sb, "ID: %d\n", detail.ID)
	fmt.Fprintf(&sb, "Type: %s\n", detail.Type)
	fmt.Fprintf(&sb, "Name: %s\n", detail.Name)
	fmt.Fprintf(&sb, "Target: %s\n", detail.Target)
	if detail.Priority > 0 {
		fmt.Fprintf(&sb, "Priority: %d\n", detail.Priority)
	}
	if detail.Weight > 0 {
		fmt.Fprintf(&sb, "Weight: %d\n", detail.Weight)
	}
	if detail.Port > 0 {
		fmt.Fprintf(&sb, "Port: %d\n", detail.Port)
	}
	if detail.Service != "" {
		fmt.Fprintf(&sb, "Service: %s\n", detail.Service)
	}
	if detail.Protocol != "" {
		fmt.Fprintf(&sb, "Protocol: %s\n", detail.Protocol)
	}
	if detail.Tag != "" {
		fmt.Fprintf(&sb, "Tag: %s\n", detail.Tag)
	}
	fmt.Fprintf(&sb, "TTL: %d seconds\n", detail.TTLSec)
	fmt.Fprintf(&sb, "Created: %s\n", detail.Created)
	fmt.Fprintf(&sb, "Updated: %s\n", detail.Updated)

	return mcp.NewToolResultText(sb.String()), nil
}

// handleDomainRecordCreate creates a new domain record.
func (s *Service) handleDomainRecordCreate(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	arguments, ok := request.Params.Arguments.(map[string]interface{})
	if !ok {
		return mcp.NewToolResultError("Invalid arguments format"), nil
	}

	domainID, err := parseIDFromArguments(arguments, "domain_id")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid domain_id parameter: %v", err)), nil
	}

	// Parse required parameters
	recordType, ok := arguments["type"].(string)
	if !ok || recordType == "" {
		return mcp.NewToolResultError("type is required"), nil
	}

	target, ok := arguments["target"].(string)
	if !ok || target == "" {
		return mcp.NewToolResultError("target is required"), nil
	}

	// Build record create options
	params := DomainRecordCreateParams{
		DomainID: domainID,
		Type:     recordType,
		Target:   target,
	}

	// Optional parameters
	if name, ok := arguments["name"].(string); ok {
		params.Name = name
	}
	if priority, ok := arguments["priority"].(float64); ok {
		params.Priority = int(priority)
	}
	if weight, ok := arguments["weight"].(float64); ok {
		params.Weight = int(weight)
	}
	if port, ok := arguments["port"].(float64); ok {
		params.Port = int(port)
	}
	if service, ok := arguments["service"].(string); ok {
		params.Service = service
	}
	if protocol, ok := arguments["protocol"].(string); ok {
		params.Protocol = protocol
	}
	if ttlSec, ok := arguments["ttl_sec"].(float64); ok {
		params.TTLSec = int(ttlSec)
	}
	if tag, ok := arguments["tag"].(string); ok {
		params.Tag = tag
	}

	account, err := s.accountManager.GetCurrentAccount()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	createOpts := linodego.DomainRecordCreateOptions{
		Type:   linodego.DomainRecordType(params.Type),
		Target: params.Target,
	}

	if params.Name != "" {
		createOpts.Name = params.Name
	}
	if params.Priority > 0 {
		createOpts.Priority = intPtr(params.Priority)
	}
	if params.Weight > 0 {
		createOpts.Weight = intPtr(params.Weight)
	}
	if params.Port > 0 {
		createOpts.Port = intPtr(params.Port)
	}
	if params.Service != "" {
		createOpts.Service = stringPtr(params.Service)
	}
	if params.Protocol != "" {
		createOpts.Protocol = stringPtr(params.Protocol)
	}
	if params.TTLSec > 0 {
		createOpts.TTLSec = params.TTLSec
	}
	if params.Tag != "" {
		createOpts.Tag = stringPtr(params.Tag)
	}

	record, err := account.Client.CreateDomainRecord(ctx, params.DomainID, createOpts)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to create domain record: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Domain record created successfully:\nID: %d\nType: %s\nName: %s\nTarget: %s",
		record.ID, record.Type, record.Name, record.Target)), nil
}

// handleDomainRecordUpdate updates a domain record.
func (s *Service) handleDomainRecordUpdate(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	arguments, ok := request.Params.Arguments.(map[string]interface{})
	if !ok {
		return mcp.NewToolResultError("Invalid arguments format"), nil
	}

	domainID, err := parseIDFromArguments(arguments, "domain_id")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid domain_id parameter: %v", err)), nil
	}

	recordID, err := parseIDFromArguments(arguments, "record_id")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid record_id parameter: %v", err)), nil
	}

	// Build record update options
	params := DomainRecordUpdateParams{
		DomainID: domainID,
		RecordID: recordID,
	}

	// Optional parameters
	if recordType, ok := arguments["type"].(string); ok {
		params.Type = recordType
	}
	if name, ok := arguments["name"].(string); ok {
		params.Name = name
	}
	if target, ok := arguments["target"].(string); ok {
		params.Target = target
	}
	if priority, ok := arguments["priority"].(float64); ok {
		params.Priority = int(priority)
	}
	if weight, ok := arguments["weight"].(float64); ok {
		params.Weight = int(weight)
	}
	if port, ok := arguments["port"].(float64); ok {
		params.Port = int(port)
	}
	if service, ok := arguments["service"].(string); ok {
		params.Service = service
	}
	if protocol, ok := arguments["protocol"].(string); ok {
		params.Protocol = protocol
	}
	if ttlSec, ok := arguments["ttl_sec"].(float64); ok {
		params.TTLSec = int(ttlSec)
	}
	if tag, ok := arguments["tag"].(string); ok {
		params.Tag = tag
	}

	account, err := s.accountManager.GetCurrentAccount()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	updateOpts := linodego.DomainRecordUpdateOptions{}

	if params.Type != "" {
		updateOpts.Type = linodego.DomainRecordType(params.Type)
	}
	if params.Name != "" {
		updateOpts.Name = params.Name
	}
	if params.Target != "" {
		updateOpts.Target = params.Target
	}
	if params.Priority > 0 {
		updateOpts.Priority = intPtr(params.Priority)
	}
	if params.Weight > 0 {
		updateOpts.Weight = intPtr(params.Weight)
	}
	if params.Port > 0 {
		updateOpts.Port = intPtr(params.Port)
	}
	if params.Service != "" {
		updateOpts.Service = stringPtr(params.Service)
	}
	if params.Protocol != "" {
		updateOpts.Protocol = stringPtr(params.Protocol)
	}
	if params.TTLSec > 0 {
		updateOpts.TTLSec = params.TTLSec
	}
	if params.Tag != "" {
		updateOpts.Tag = stringPtr(params.Tag)
	}

	record, err := account.Client.UpdateDomainRecord(ctx, params.DomainID, params.RecordID, updateOpts)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to update domain record: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Domain record updated successfully:\nID: %d\nType: %s\nName: %s\nTarget: %s",
		record.ID, record.Type, record.Name, record.Target)), nil
}

// handleDomainRecordDelete deletes a domain record.
func (s *Service) handleDomainRecordDelete(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	arguments, ok := request.Params.Arguments.(map[string]interface{})
	if !ok {
		return mcp.NewToolResultError("Invalid arguments format"), nil
	}

	domainID, err := parseIDFromArguments(arguments, "domain_id")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid domain_id parameter: %v", err)), nil
	}

	recordID, err := parseIDFromArguments(arguments, "record_id")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid record_id parameter: %v", err)), nil
	}

	params := DomainRecordDeleteParams{
		DomainID: domainID,
		RecordID: recordID,
	}

	account, err := s.accountManager.GetCurrentAccount()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	err = account.Client.DeleteDomainRecord(ctx, params.DomainID, params.RecordID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to delete domain record: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Domain record %d deleted successfully from domain %d",
		params.RecordID, params.DomainID)), nil
}

// parseArguments is a placeholder function for structured parameter parsing.
// TODO: Convert remaining handler functions to use direct argument parsing like instances and domains.
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
func intPtr(i int) *int {
	return &i
}

// stringPtr returns a pointer to the given string value.
func stringPtr(s string) *string {
	return &s
}
