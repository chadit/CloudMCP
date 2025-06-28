package linode

import (
	"context"
	"fmt"
	"strings"

	"github.com/linode/linodego"
	"github.com/mark3labs/mcp-go/mcp"
)


// formatMySQLDatabases formats MySQL database list for display
func formatMySQLDatabases(databases []linodego.MySQLDatabase) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d MySQL databases:\n\n", len(databases)))

	for _, db := range databases {
		fmt.Fprintf(&sb, "ID: %d | %s (MySQL %s)\n", db.ID, db.Label, db.Version)
		fmt.Fprintf(&sb, "  Region: %s | Type: %s | Status: %s\n", db.Region, db.Type, string(db.Status))
		fmt.Fprintf(&sb, "  Primary Host: %s | Port: %d\n", db.Hosts.Primary, db.Port)
		if db.Hosts.Secondary != "" {
			fmt.Fprintf(&sb, "  Secondary Host: %s\n", db.Hosts.Secondary)
		}
		fmt.Fprintf(&sb, "  Created: %s | Updated: %s\n", 
			db.Created.Format("2006-01-02T15:04:05"), 
			db.Updated.Format("2006-01-02T15:04:05"))
		sb.WriteString("\n")
	}

	return sb.String()
}

// formatPostgresDatabases formats PostgreSQL database list for display
func formatPostgresDatabases(databases []linodego.PostgresDatabase) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d PostgreSQL databases:\n\n", len(databases)))

	for _, db := range databases {
		fmt.Fprintf(&sb, "ID: %d | %s (PostgreSQL %s)\n", db.ID, db.Label, db.Version)
		fmt.Fprintf(&sb, "  Region: %s | Type: %s | Status: %s\n", db.Region, db.Type, string(db.Status))
		fmt.Fprintf(&sb, "  Primary Host: %s | Port: %d\n", db.Hosts.Primary, db.Port)
		if db.Hosts.Secondary != "" {
			fmt.Fprintf(&sb, "  Secondary Host: %s\n", db.Hosts.Secondary)
		}
		fmt.Fprintf(&sb, "  Created: %s | Updated: %s\n", 
			db.Created.Format("2006-01-02T15:04:05"), 
			db.Updated.Format("2006-01-02T15:04:05"))
		sb.WriteString("\n")
	}

	return sb.String()
}

// formatMySQLDatabaseDetail formats MySQL database detail for display
func formatMySQLDatabaseDetail(db *linodego.MySQLDatabase) string {
	return formatDatabaseDetail("MySQL", MySQLDatabaseDetail{
		ID:          db.ID,
		Label:       db.Label,
		Engine:      db.Engine,
		Version:     db.Version,
		Region:      db.Region,
		Type:        db.Type,
		Status:      string(db.Status),
		ClusterSize: db.ClusterSize,
		Hosts: DatabaseHosts{
			Primary:   db.Hosts.Primary,
			Secondary: db.Hosts.Secondary,
		},
		Port:      db.Port,
		AllowList: db.AllowList,
		Created:   db.Created.Format("2006-01-02T15:04:05"),
		Updated:   db.Updated.Format("2006-01-02T15:04:05"),
	})
}

// formatPostgresDatabaseDetail formats PostgreSQL database detail for display
func formatPostgresDatabaseDetail(db *linodego.PostgresDatabase) string {
	return formatDatabaseDetail("PostgreSQL", PostgresDatabaseDetail{
		ID:          db.ID,
		Label:       db.Label,
		Engine:      db.Engine,
		Version:     db.Version,
		Region:      db.Region,
		Type:        db.Type,
		Status:      string(db.Status),
		ClusterSize: db.ClusterSize,
		Hosts: DatabaseHosts{
			Primary:   db.Hosts.Primary,
			Secondary: db.Hosts.Secondary,
		},
		Port:      db.Port,
		AllowList: db.AllowList,
		Created:   db.Created.Format("2006-01-02T15:04:05"),
		Updated:   db.Updated.Format("2006-01-02T15:04:05"),
	})
}

// handleDatabasesList lists all databases (both MySQL and PostgreSQL).
func (s *Service) handleDatabasesList(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	account, err := s.accountManager.GetCurrentAccount()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	databases, err := account.Client.ListDatabases(ctx, nil)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to list databases: %v", err)), nil
	}

	var summaries []DatabaseSummary
	for _, db := range databases {
		summary := DatabaseSummary{
			ID:          db.ID,
			Label:       db.Label,
			Engine:      db.Engine,
			Version:     db.Version,
			Region:      db.Region,
			Type:        db.Type,
			Status:      string(db.Status),
			ClusterSize: db.ClusterSize,
			Created:     db.Created.Format("2006-01-02T15:04:05"),
			Updated:     db.Updated.Format("2006-01-02T15:04:05"),
		}
		summaries = append(summaries, summary)
	}

	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Found %d databases:\n\n", len(summaries)))

	for _, db := range summaries {
		fmt.Fprintf(&sb, "ID: %d | %s (%s %s)\n", db.ID, db.Label, db.Engine, db.Version)
		fmt.Fprintf(&sb, "  Region: %s | Type: %s | Status: %s\n", db.Region, db.Type, db.Status)
		fmt.Fprintf(&sb, "  Cluster Size: %d nodes\n", db.ClusterSize)
		fmt.Fprintf(&sb, "  Updated: %s\n", db.Updated)
		sb.WriteString("\n")
	}

	return mcp.NewToolResultText(sb.String()), nil
}

// handleMySQLDatabasesList lists all MySQL databases.
func (s *Service) handleMySQLDatabasesList(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	account, err := s.accountManager.GetCurrentAccount()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	databases, err := account.Client.ListMySQLDatabases(ctx, nil)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to list MySQL databases: %v", err)), nil
	}

	var summaries []MySQLDatabaseSummary
	for _, db := range databases {
		summary := MySQLDatabaseSummary{
			ID:          db.ID,
			Label:       db.Label,
			Engine:      db.Engine,
			Version:     db.Version,
			Region:      db.Region,
			Type:        db.Type,
			Status:      string(db.Status),
			ClusterSize: db.ClusterSize,
			Hosts: DatabaseHosts{
				Primary:   db.Hosts.Primary,
				Secondary: db.Hosts.Secondary,
			},
			Port:    db.Port,
			Created: db.Created.Format("2006-01-02T15:04:05"),
			Updated: db.Updated.Format("2006-01-02T15:04:05"),
		}
		summaries = append(summaries, summary)
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d MySQL databases:\n\n", len(summaries)))

	for _, db := range summaries {
		fmt.Fprintf(&sb, "ID: %d | %s (MySQL %s)\n", db.ID, db.Label, db.Version)
		fmt.Fprintf(&sb, "  Region: %s | Type: %s | Status: %s\n", db.Region, db.Type, db.Status)
		fmt.Fprintf(&sb, "  Primary Host: %s | Port: %d\n", db.Hosts.Primary, db.Port)
		if db.Hosts.Secondary != "" {
			fmt.Fprintf(&sb, "  Secondary Host: %s\n", db.Hosts.Secondary)
		}
		fmt.Fprintf(&sb, "  Cluster Size: %d nodes | Updated: %s\n", db.ClusterSize, db.Updated)
		sb.WriteString("\n")
	}

	return mcp.NewToolResultText(sb.String()), nil
}

// handlePostgresDatabasesList lists all PostgreSQL databases.
func (s *Service) handlePostgresDatabasesList(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	account, err := s.accountManager.GetCurrentAccount()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	databases, err := account.Client.ListPostgresDatabases(ctx, nil)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to list PostgreSQL databases: %v", err)), nil
	}

	var summaries []PostgresDatabaseSummary
	for _, db := range databases {
		summary := PostgresDatabaseSummary{
			ID:          db.ID,
			Label:       db.Label,
			Engine:      db.Engine,
			Version:     db.Version,
			Region:      db.Region,
			Type:        db.Type,
			Status:      string(db.Status),
			ClusterSize: db.ClusterSize,
			Hosts: DatabaseHosts{
				Primary:   db.Hosts.Primary,
				Secondary: db.Hosts.Secondary,
			},
			Port:    db.Port,
			Created: db.Created.Format("2006-01-02T15:04:05"),
			Updated: db.Updated.Format("2006-01-02T15:04:05"),
		}
		summaries = append(summaries, summary)
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d PostgreSQL databases:\n\n", len(summaries)))

	for _, db := range summaries {
		fmt.Fprintf(&sb, "ID: %d | %s (PostgreSQL %s)\n", db.ID, db.Label, db.Version)
		fmt.Fprintf(&sb, "  Region: %s | Type: %s | Status: %s\n", db.Region, db.Type, db.Status)
		fmt.Fprintf(&sb, "  Primary Host: %s | Port: %d\n", db.Hosts.Primary, db.Port)
		if db.Hosts.Secondary != "" {
			fmt.Fprintf(&sb, "  Secondary Host: %s\n", db.Hosts.Secondary)
		}
		fmt.Fprintf(&sb, "  Cluster Size: %d nodes | Updated: %s\n", db.ClusterSize, db.Updated)
		sb.WriteString("\n")
	}

	return mcp.NewToolResultText(sb.String()), nil
}

// formatDatabaseDetail formats database details in a standardized way.
func formatDatabaseDetail(engineType string, detail interface {
	GetID() int
	GetLabel() string
	GetEngine() string
	GetVersion() string
	GetRegion() string
	GetType() string
	GetStatus() string
	GetClusterSize() int
	GetHosts() DatabaseHosts
	GetPort() int
	GetAllowList() []string
	GetCreated() string
	GetUpdated() string
},
) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "%s Database Details:\n", engineType)
	fmt.Fprintf(&sb, "ID: %d\n", detail.GetID())
	fmt.Fprintf(&sb, "Label: %s\n", detail.GetLabel())
	fmt.Fprintf(&sb, "Engine: %s %s\n", detail.GetEngine(), detail.GetVersion())
	fmt.Fprintf(&sb, "Region: %s\n", detail.GetRegion())
	fmt.Fprintf(&sb, "Type: %s\n", detail.GetType())
	fmt.Fprintf(&sb, "Status: %s\n", detail.GetStatus())
	fmt.Fprintf(&sb, "Cluster Size: %d nodes\n", detail.GetClusterSize())
	fmt.Fprintf(&sb, "Primary Host: %s\n", detail.GetHosts().Primary)
	if detail.GetHosts().Secondary != "" {
		fmt.Fprintf(&sb, "Secondary Host: %s\n", detail.GetHosts().Secondary)
	}
	fmt.Fprintf(&sb, "Port: %d\n", detail.GetPort())
	fmt.Fprintf(&sb, "Created: %s\n", detail.GetCreated())
	fmt.Fprintf(&sb, "Updated: %s\n\n", detail.GetUpdated())

	if len(detail.GetAllowList()) > 0 {
		fmt.Fprintf(&sb, "Allow List (IP ranges with access):\n")

		for _, ip := range detail.GetAllowList() {
			fmt.Fprintf(&sb, "  - %s\n", ip)
		}
	} else {
		fmt.Fprintf(&sb, "Allow List: No IP restrictions (open access)\n")
	}

	return sb.String()
}

// GetID returns the database ID.
func (d MySQLDatabaseDetail) GetID() int { return d.ID }

// GetLabel returns the database label.
func (d MySQLDatabaseDetail) GetLabel() string { return d.Label }

// GetEngine returns the database engine.
func (d MySQLDatabaseDetail) GetEngine() string { return d.Engine }

// GetVersion returns the database version.
func (d MySQLDatabaseDetail) GetVersion() string { return d.Version }

// GetRegion returns the database region.
func (d MySQLDatabaseDetail) GetRegion() string { return d.Region }

// GetType returns the database type.
func (d MySQLDatabaseDetail) GetType() string { return d.Type }

// GetStatus returns the database status.
func (d MySQLDatabaseDetail) GetStatus() string { return d.Status }

// GetClusterSize returns the cluster size.
func (d MySQLDatabaseDetail) GetClusterSize() int { return d.ClusterSize }

// GetHosts returns the database hosts.
func (d MySQLDatabaseDetail) GetHosts() DatabaseHosts { return d.Hosts }

// GetPort returns the database port.
func (d MySQLDatabaseDetail) GetPort() int { return d.Port }

// GetAllowList returns the allow list.
func (d MySQLDatabaseDetail) GetAllowList() []string { return d.AllowList }

// GetCreated returns the creation time.
func (d MySQLDatabaseDetail) GetCreated() string { return d.Created }

// GetUpdated returns the update time.
func (d MySQLDatabaseDetail) GetUpdated() string { return d.Updated }

// GetID returns the database ID.
func (d PostgresDatabaseDetail) GetID() int { return d.ID }

// GetLabel returns the database label.
func (d PostgresDatabaseDetail) GetLabel() string { return d.Label }

// GetEngine returns the database engine.
func (d PostgresDatabaseDetail) GetEngine() string { return d.Engine }

// GetVersion returns the database version.
func (d PostgresDatabaseDetail) GetVersion() string { return d.Version }

// GetRegion returns the database region.
func (d PostgresDatabaseDetail) GetRegion() string { return d.Region }

// GetType returns the database type.
func (d PostgresDatabaseDetail) GetType() string { return d.Type }

// GetStatus returns the database status.
func (d PostgresDatabaseDetail) GetStatus() string { return d.Status }

// GetClusterSize returns the cluster size.
func (d PostgresDatabaseDetail) GetClusterSize() int { return d.ClusterSize }

// GetHosts returns the database hosts.
func (d PostgresDatabaseDetail) GetHosts() DatabaseHosts { return d.Hosts }

// GetPort returns the database port.
func (d PostgresDatabaseDetail) GetPort() int { return d.Port }

// GetAllowList returns the allow list.
func (d PostgresDatabaseDetail) GetAllowList() []string { return d.AllowList }

// GetCreated returns the creation time.
func (d PostgresDatabaseDetail) GetCreated() string { return d.Created }

// GetUpdated returns the update time.
func (d PostgresDatabaseDetail) GetUpdated() string { return d.Updated }

// handleMySQLDatabaseGet gets details of a specific MySQL database.
func (s *Service) handleMySQLDatabaseGet(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var params MySQLDatabaseGetParams
	if err := parseArguments(request.Params.Arguments, &params); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid parameters: %v", err)), nil
	}

	account, err := s.accountManager.GetCurrentAccount()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	db, err := account.Client.GetMySQLDatabase(ctx, params.DatabaseID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get MySQL database: %v", err)), nil
	}

	detail := MySQLDatabaseDetail{
		ID:          db.ID,
		Label:       db.Label,
		Engine:      db.Engine,
		Version:     db.Version,
		Region:      db.Region,
		Type:        db.Type,
		Status:      string(db.Status),
		ClusterSize: db.ClusterSize,
		Hosts: DatabaseHosts{
			Primary:   db.Hosts.Primary,
			Secondary: db.Hosts.Secondary,
		},
		Port:      db.Port,
		AllowList: db.AllowList,
		Created:   db.Created.Format("2006-01-02T15:04:05"),
		Updated:   db.Updated.Format("2006-01-02T15:04:05"),
	}

	return mcp.NewToolResultText(formatDatabaseDetail("MySQL", detail)), nil
}

// handlePostgresDatabaseGet gets details of a specific PostgreSQL database.
func (s *Service) handlePostgresDatabaseGet(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var params PostgresDatabaseGetParams
	if err := parseArguments(request.Params.Arguments, &params); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid parameters: %v", err)), nil
	}

	account, err := s.accountManager.GetCurrentAccount()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	db, err := account.Client.GetPostgresDatabase(ctx, params.DatabaseID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get PostgreSQL database: %v", err)), nil
	}

	detail := PostgresDatabaseDetail{
		ID:          db.ID,
		Label:       db.Label,
		Engine:      db.Engine,
		Version:     db.Version,
		Region:      db.Region,
		Type:        db.Type,
		Status:      string(db.Status),
		ClusterSize: db.ClusterSize,
		Hosts: DatabaseHosts{
			Primary:   db.Hosts.Primary,
			Secondary: db.Hosts.Secondary,
		},
		Port:      db.Port,
		AllowList: db.AllowList,
		Created:   db.Created.Format("2006-01-02T15:04:05"),
		Updated:   db.Updated.Format("2006-01-02T15:04:05"),
	}

	return mcp.NewToolResultText(formatDatabaseDetail("PostgreSQL", detail)), nil
}

// handleMySQLDatabaseCreate creates a new MySQL database.
func (s *Service) handleMySQLDatabaseCreate(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var params MySQLDatabaseCreateParams
	if err := parseArguments(request.Params.Arguments, &params); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid parameters: %v", err)), nil
	}

	account, err := s.accountManager.GetCurrentAccount()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	createOpts := linodego.MySQLCreateOptions{
		Label:       params.Label,
		Region:      params.Region,
		Type:        params.Type,
		Engine:      params.Engine,
		ClusterSize: params.ClusterSize,
		AllowList:   params.AllowList,
	}

	db, err := account.Client.CreateMySQLDatabase(ctx, createOpts)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to create MySQL database: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("MySQL database created successfully:\nID: %d\nLabel: %s\nEngine: %s %s\nRegion: %s\nType: %s\nStatus: %s\nPrimary Host: %s\nPort: %d",
		db.ID, db.Label, db.Engine, db.Version, db.Region, db.Type, db.Status, db.Hosts.Primary, db.Port)), nil
}

// handlePostgresDatabaseCreate creates a new PostgreSQL database.
func (s *Service) handlePostgresDatabaseCreate(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var params PostgresDatabaseCreateParams
	if err := parseArguments(request.Params.Arguments, &params); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid parameters: %v", err)), nil
	}

	account, err := s.accountManager.GetCurrentAccount()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	createOpts := linodego.PostgresCreateOptions{
		Label:       params.Label,
		Region:      params.Region,
		Type:        params.Type,
		Engine:      params.Engine,
		ClusterSize: params.ClusterSize,
		AllowList:   params.AllowList,
	}

	db, err := account.Client.CreatePostgresDatabase(ctx, createOpts)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to create PostgreSQL database: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("PostgreSQL database created successfully:\nID: %d\nLabel: %s\nEngine: %s %s\nRegion: %s\nType: %s\nStatus: %s\nPrimary Host: %s\nPort: %d",
		db.ID, db.Label, db.Engine, db.Version, db.Region, db.Type, db.Status, db.Hosts.Primary, db.Port)), nil
}

// formatDatabaseUpdateResult formats database update result in a standardized way.
func formatDatabaseUpdateResult(engineType string, db interface {
	GetID() int
	GetLabel() string
	GetStatus() string
	GetHosts() DatabaseHosts
},
) string {
	return fmt.Sprintf("%s database updated successfully:\nID: %d\nLabel: %s\nStatus: %s\nPrimary Host: %s",
		engineType, db.GetID(), db.GetLabel(), db.GetStatus(), db.GetHosts().Primary)
}

// DatabaseResult represents a database operation result.
type DatabaseResult struct {
	ID     int
	Label  string
	Status string
	Hosts  DatabaseHosts
}

// GetID returns the database ID.
func (d DatabaseResult) GetID() int { return d.ID }

// GetLabel returns the database label.
func (d DatabaseResult) GetLabel() string { return d.Label }

// GetStatus returns the database status.
func (d DatabaseResult) GetStatus() string { return d.Status }

// GetHosts returns the database hosts.
func (d DatabaseResult) GetHosts() DatabaseHosts { return d.Hosts }

// handleMySQLDatabaseUpdate updates an existing MySQL database.
func (s *Service) handleMySQLDatabaseUpdate(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var params MySQLDatabaseUpdateParams
	if err := parseArguments(request.Params.Arguments, &params); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid parameters: %v", err)), nil
	}

	account, err := s.accountManager.GetCurrentAccount()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	updateOpts := linodego.MySQLUpdateOptions{}
	if params.Label != "" {
		updateOpts.Label = params.Label
	}
	if params.AllowList != nil {
		updateOpts.AllowList = &params.AllowList
	}

	db, err := account.Client.UpdateMySQLDatabase(ctx, params.DatabaseID, updateOpts)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to update MySQL database: %v", err)), nil
	}

	result := DatabaseResult{
		ID:     db.ID,
		Label:  db.Label,
		Status: string(db.Status),
		Hosts: DatabaseHosts{
			Primary:   db.Hosts.Primary,
			Secondary: db.Hosts.Secondary,
		},
	}

	return mcp.NewToolResultText(formatDatabaseUpdateResult("MySQL", result)), nil
}

// handlePostgresDatabaseUpdate updates an existing PostgreSQL database.
func (s *Service) handlePostgresDatabaseUpdate(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var params PostgresDatabaseUpdateParams
	if err := parseArguments(request.Params.Arguments, &params); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid parameters: %v", err)), nil
	}

	account, err := s.accountManager.GetCurrentAccount()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	updateOpts := linodego.PostgresUpdateOptions{}
	if params.Label != "" {
		updateOpts.Label = params.Label
	}
	if params.AllowList != nil {
		updateOpts.AllowList = &params.AllowList
	}

	db, err := account.Client.UpdatePostgresDatabase(ctx, params.DatabaseID, updateOpts)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to update PostgreSQL database: %v", err)), nil
	}

	result := DatabaseResult{
		ID:     db.ID,
		Label:  db.Label,
		Status: string(db.Status),
		Hosts: DatabaseHosts{
			Primary:   db.Hosts.Primary,
			Secondary: db.Hosts.Secondary,
		},
	}

	return mcp.NewToolResultText(formatDatabaseUpdateResult("PostgreSQL", result)), nil
}

// handleMySQLDatabaseDelete deletes a MySQL database.
func (s *Service) handleMySQLDatabaseDelete(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var params MySQLDatabaseDeleteParams
	if err := parseArguments(request.Params.Arguments, &params); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid parameters: %v", err)), nil
	}

	account, err := s.accountManager.GetCurrentAccount()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	err = account.Client.DeleteMySQLDatabase(ctx, params.DatabaseID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to delete MySQL database: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("MySQL database %d deleted successfully", params.DatabaseID)), nil
}

// handlePostgresDatabaseDelete deletes a PostgreSQL database.
func (s *Service) handlePostgresDatabaseDelete(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var params PostgresDatabaseDeleteParams
	if err := parseArguments(request.Params.Arguments, &params); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid parameters: %v", err)), nil
	}

	account, err := s.accountManager.GetCurrentAccount()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	err = account.Client.DeletePostgresDatabase(ctx, params.DatabaseID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to delete PostgreSQL database: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("PostgreSQL database %d deleted successfully", params.DatabaseID)), nil
}

// handleMySQLDatabaseCredentials gets root credentials for a MySQL database.
func (s *Service) handleMySQLDatabaseCredentials(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var params MySQLDatabaseCredentialsParams
	if err := parseArguments(request.Params.Arguments, &params); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid parameters: %v", err)), nil
	}

	account, err := s.accountManager.GetCurrentAccount()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	creds, err := account.Client.GetMySQLDatabaseCredentials(ctx, params.DatabaseID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get MySQL database credentials: %v", err)), nil
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "MySQL Database Credentials:\n")
	fmt.Fprintf(&sb, "Username: %s\n", creds.Username)
	fmt.Fprintf(&sb, "Password: %s\n", creds.Password)
	fmt.Fprintf(&sb, "\nConnection Details:\n")
	fmt.Fprintf(&sb, "These are the root credentials for database %d.\n", params.DatabaseID)
	fmt.Fprintf(&sb, "Use these credentials to connect to your MySQL database.\n")
	fmt.Fprintf(&sb, "\nSecurity Note: Store these credentials securely and limit access.")

	return mcp.NewToolResultText(sb.String()), nil
}

// handlePostgresDatabaseCredentials gets root credentials for a PostgreSQL database.
func (s *Service) handlePostgresDatabaseCredentials(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var params PostgresDatabaseCredentialsParams
	if err := parseArguments(request.Params.Arguments, &params); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid parameters: %v", err)), nil
	}

	account, err := s.accountManager.GetCurrentAccount()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	creds, err := account.Client.GetPostgresDatabaseCredentials(ctx, params.DatabaseID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get PostgreSQL database credentials: %v", err)), nil
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "PostgreSQL Database Credentials:\n")
	fmt.Fprintf(&sb, "Username: %s\n", creds.Username)
	fmt.Fprintf(&sb, "Password: %s\n", creds.Password)
	fmt.Fprintf(&sb, "\nConnection Details:\n")
	fmt.Fprintf(&sb, "These are the root credentials for database %d.\n", params.DatabaseID)
	fmt.Fprintf(&sb, "Use these credentials to connect to your PostgreSQL database.\n")
	fmt.Fprintf(&sb, "\nSecurity Note: Store these credentials securely and limit access.")

	return mcp.NewToolResultText(sb.String()), nil
}

// handleMySQLDatabaseCredentialsReset resets root password for a MySQL database.
func (s *Service) handleMySQLDatabaseCredentialsReset(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var params MySQLDatabaseCredentialsResetParams
	if err := parseArguments(request.Params.Arguments, &params); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid parameters: %v", err)), nil
	}

	account, err := s.accountManager.GetCurrentAccount()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	err = account.Client.ResetMySQLDatabaseCredentials(ctx, params.DatabaseID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to reset MySQL database credentials: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("MySQL database %d root password reset successfully.\nRetrieve new credentials using the credentials command.", params.DatabaseID)), nil
}

// handlePostgresDatabaseCredentialsReset resets root password for a PostgreSQL database.
func (s *Service) handlePostgresDatabaseCredentialsReset(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var params PostgresDatabaseCredentialsResetParams
	if err := parseArguments(request.Params.Arguments, &params); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid parameters: %v", err)), nil
	}

	account, err := s.accountManager.GetCurrentAccount()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	err = account.Client.ResetPostgresDatabaseCredentials(ctx, params.DatabaseID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to reset PostgreSQL database credentials: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("PostgreSQL database %d root password reset successfully.\nRetrieve new credentials using the credentials command.", params.DatabaseID)), nil
}

// handleDatabaseEnginesList lists all available database engines.
func (s *Service) handleDatabaseEnginesList(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	account, err := s.accountManager.GetCurrentAccount()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	engines, err := account.Client.ListDatabaseEngines(ctx, nil)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to list database engines: %v", err)), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d database engines:\n\n", len(engines)))

	for _, engine := range engines {
		fmt.Fprintf(&sb, "Engine: %s\n", engine.Engine)
		fmt.Fprintf(&sb, "  ID: %s\n", engine.ID)
		fmt.Fprintf(&sb, "  Version: %s\n", engine.Version)
		// Note: Deprecated field may not be available in all linodego versions
		fmt.Fprintf(&sb, "  Status: Active\n")
		sb.WriteString("\n")
	}

	return mcp.NewToolResultText(sb.String()), nil
}

// handleDatabaseTypesList lists all available database types.
func (s *Service) handleDatabaseTypesList(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	account, err := s.accountManager.GetCurrentAccount()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	types, err := account.Client.ListDatabaseTypes(ctx, nil)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to list database types: %v", err)), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d database types:\n\n", len(types)))

	for _, dbType := range types {
		fmt.Fprintf(&sb, "Type: %s\n", dbType.ID)
		fmt.Fprintf(&sb, "  Label: %s\n", dbType.Label)
		fmt.Fprintf(&sb, "  Class: %s\n", dbType.Class)
		fmt.Fprintf(&sb, "  Disk: %d GB\n", dbType.Disk)
		fmt.Fprintf(&sb, "  Memory: %d MB\n", dbType.Memory)
		// Note: VCPUs and Price fields may not be available in all linodego versions
		// Note: Engines field structure may vary in linodego versions
		sb.WriteString("\n")
	}

	return mcp.NewToolResultText(sb.String()), nil
}
