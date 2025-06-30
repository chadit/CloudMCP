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

var (
	// ErrFailedToListDatabases is returned when listing databases fails.
	ErrFailedToListDatabases = errors.New("failed to list databases")
	// ErrFailedToListMySQLDatabases is returned when listing MySQL databases fails.
	ErrFailedToListMySQLDatabases = errors.New("failed to list MySQL databases")
	// ErrFailedToListPostgresDatabases is returned when listing PostgreSQL databases fails.
	ErrFailedToListPostgresDatabases = errors.New("failed to list PostgreSQL databases")
	// ErrFailedToGetMySQLDatabase is returned when getting MySQL database fails.
	ErrFailedToGetMySQLDatabase = errors.New("failed to get MySQL database")
	// ErrFailedToGetPostgresDatabase is returned when getting PostgreSQL database fails.
	ErrFailedToGetPostgresDatabase = errors.New("failed to get PostgreSQL database")
	// ErrFailedToCreateMySQLDatabase is returned when creating MySQL database fails.
	ErrFailedToCreateMySQLDatabase = errors.New("failed to create MySQL database")
	// ErrFailedToCreatePostgresDatabase is returned when creating PostgreSQL database fails.
	ErrFailedToCreatePostgresDatabase = errors.New("failed to create PostgreSQL database")
	// ErrFailedToUpdateMySQLDatabase is returned when updating MySQL database fails.
	ErrFailedToUpdateMySQLDatabase = errors.New("failed to update MySQL database")
	// ErrFailedToUpdatePostgresDatabase is returned when updating PostgreSQL database fails.
	ErrFailedToUpdatePostgresDatabase = errors.New("failed to update PostgreSQL database")
	// ErrFailedToDeleteMySQLDatabase is returned when deleting MySQL database fails.
	ErrFailedToDeleteMySQLDatabase = errors.New("failed to delete MySQL database")
	// ErrFailedToDeletePostgresDatabase is returned when deleting PostgreSQL database fails.
	ErrFailedToDeletePostgresDatabase = errors.New("failed to delete PostgreSQL database")
	// ErrFailedToGetMySQLCredentials is returned when getting MySQL credentials fails.
	ErrFailedToGetMySQLCredentials = errors.New("failed to get MySQL database credentials")
	// ErrFailedToGetPostgresCredentials is returned when getting PostgreSQL credentials fails.
	ErrFailedToGetPostgresCredentials = errors.New("failed to get PostgreSQL database credentials")
	// ErrFailedToResetMySQLCredentials is returned when resetting MySQL credentials fails.
	ErrFailedToResetMySQLCredentials = errors.New("failed to reset MySQL database credentials")
	// ErrFailedToResetPostgresCredentials is returned when resetting PostgreSQL credentials fails.
	ErrFailedToResetPostgresCredentials = errors.New("failed to reset PostgreSQL database credentials")
	// ErrFailedToListDatabaseEngines is returned when listing database engines fails.
	ErrFailedToListDatabaseEngines = errors.New("failed to list database engines")
	// ErrFailedToListDatabaseTypes is returned when listing database types fails.
	ErrFailedToListDatabaseTypes = errors.New("failed to list database types")
)

// createDatabaseResult creates a standardized database result.
func createDatabaseResult(database interface{}) DatabaseResult {
	switch databaseObj := database.(type) {
	case *linodego.MySQLDatabase:
		return DatabaseResult{
			ID:     databaseObj.ID,
			Label:  databaseObj.Label,
			Status: string(databaseObj.Status),
			Hosts: DatabaseHosts{
				Primary:   databaseObj.Hosts.Primary,
				Secondary: databaseObj.Hosts.Secondary,
			},
		}
	case *linodego.PostgresDatabase:
		return DatabaseResult{
			ID:     databaseObj.ID,
			Label:  databaseObj.Label,
			Status: string(databaseObj.Status),
			Hosts: DatabaseHosts{
				Primary:   databaseObj.Hosts.Primary,
				Secondary: databaseObj.Hosts.Secondary,
			},
		}
	default:
		return DatabaseResult{}
	}
}

// formatMySQLDetail formats MySQL database details.
func formatMySQLDetail(database *linodego.MySQLDatabase) MySQLDatabaseDetail {
	return MySQLDatabaseDetail{
		ID:          database.ID,
		Label:       database.Label,
		Engine:      database.Engine,
		Version:     database.Version,
		Region:      database.Region,
		Type:        database.Type,
		Status:      string(database.Status),
		ClusterSize: database.ClusterSize,
		Hosts: DatabaseHosts{
			Primary:   database.Hosts.Primary,
			Secondary: database.Hosts.Secondary,
		},
		Port:      database.Port,
		AllowList: database.AllowList,
		Created:   database.Created.Format(timeFormatLayout),
		Updated:   database.Updated.Format(timeFormatLayout),
	}
}

// formatPostgresDetail formats PostgreSQL database details.
func formatPostgresDetail(database *linodego.PostgresDatabase) PostgresDatabaseDetail {
	return PostgresDatabaseDetail{
		ID:          database.ID,
		Label:       database.Label,
		Engine:      database.Engine,
		Version:     database.Version,
		Region:      database.Region,
		Type:        database.Type,
		Status:      string(database.Status),
		ClusterSize: database.ClusterSize,
		Hosts: DatabaseHosts{
			Primary:   database.Hosts.Primary,
			Secondary: database.Hosts.Secondary,
		},
		Port:      database.Port,
		AllowList: database.AllowList,
		Created:   database.Created.Format(timeFormatLayout),
		Updated:   database.Updated.Format(timeFormatLayout),
	}
}

// handleDatabasesList lists all databases (both MySQL and PostgreSQL).
func (s *Service) handleDatabasesList(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	account, accountErr := s.accountManager.GetCurrentAccount()
	if accountErr != nil {
		return nil, fmt.Errorf("failed to get current account: %w", accountErr)
	}

	databases, databasesErr := account.Client.ListDatabases(ctx, nil)
	if databasesErr != nil {
		return nil, types.NewToolError("linode", "databases_list", //nolint:wrapcheck // types.NewToolError already wraps the error
			"failed to list databases", databasesErr)
	}

	summaries := make([]DatabaseSummary, 0, len(databases))

	for _, database := range databases {
		summary := DatabaseSummary{
			ID:          database.ID,
			Label:       database.Label,
			Engine:      database.Engine,
			Version:     database.Version,
			Region:      database.Region,
			Type:        database.Type,
			Status:      string(database.Status),
			ClusterSize: database.ClusterSize,
			Created:     database.Created.Format(timeFormatLayout),
			Updated:     database.Updated.Format(timeFormatLayout),
		}
		summaries = append(summaries, summary)
	}

	var stringBuilder strings.Builder

	stringBuilder.WriteString(fmt.Sprintf("Found %d databases:\n\n", len(summaries)))

	for _, database := range summaries {
		fmt.Fprintf(&stringBuilder, "ID: %d | %s (%s %s)\n", database.ID, database.Label, database.Engine, database.Version)
		fmt.Fprintf(&stringBuilder, "  Region: %s | Type: %s | Status: %s\n", database.Region, database.Type, database.Status)
		fmt.Fprintf(&stringBuilder, "  Cluster Size: %d nodes\n", database.ClusterSize)
		fmt.Fprintf(&stringBuilder, "  Updated: %s\n", database.Updated)
		stringBuilder.WriteString("\n")
	}

	return mcp.NewToolResultText(stringBuilder.String()), nil
}

// handleDatabasesListGeneric is a generic helper for listing databases.
func handleDatabasesListGeneric[T any](
	ctx context.Context,
	s *Service,
	listFunc func(context.Context, *linodego.ListOptions) ([]T, error),
	convertFunc func(T) TypedDatabaseSummary,
	dbType, toolName, errorMsg string,
) (*mcp.CallToolResult, error) {
	databases, databasesErr := listFunc(ctx, nil)
	if databasesErr != nil {
		return nil, types.NewToolError("linode", toolName, //nolint:wrapcheck // types.NewToolError already wraps the error
			errorMsg, databasesErr)
	}

	summaries := make([]TypedDatabaseSummary, 0, len(databases))
	for _, database := range databases {
		summaries = append(summaries, convertFunc(database))
	}

	return mcp.NewToolResultText(s.formatTypedDatabaseList(dbType, summaries)), nil
}

// convertMySQLDatabaseToSummary converts a MySQL database to a summary.
func convertMySQLDatabaseToSummary(database linodego.MySQLDatabase) TypedDatabaseSummary {
	return MySQLDatabaseSummary{
		ID:          database.ID,
		Label:       database.Label,
		Engine:      database.Engine,
		Version:     database.Version,
		Region:      database.Region,
		Type:        database.Type,
		Status:      string(database.Status),
		ClusterSize: database.ClusterSize,
		Hosts: DatabaseHosts{
			Primary:   database.Hosts.Primary,
			Secondary: database.Hosts.Secondary,
		},
		Port:    database.Port,
		Created: database.Created.Format(timeFormatLayout),
		Updated: database.Updated.Format(timeFormatLayout),
	}
}

// convertPostgresDatabaseToSummary converts a PostgreSQL database to a summary.
func convertPostgresDatabaseToSummary(database linodego.PostgresDatabase) TypedDatabaseSummary {
	return PostgresDatabaseSummary{
		ID:          database.ID,
		Label:       database.Label,
		Engine:      database.Engine,
		Version:     database.Version,
		Region:      database.Region,
		Type:        database.Type,
		Status:      string(database.Status),
		ClusterSize: database.ClusterSize,
		Hosts: DatabaseHosts{
			Primary:   database.Hosts.Primary,
			Secondary: database.Hosts.Secondary,
		},
		Port:    database.Port,
		Created: database.Created.Format(timeFormatLayout),
		Updated: database.Updated.Format(timeFormatLayout),
	}
}

// handleMySQLDatabasesList lists all MySQL databases.
func (s *Service) handleMySQLDatabasesList(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	account, err := s.accountManager.GetCurrentAccount()
	if err != nil {
		return nil, fmt.Errorf("failed to get current account: %w", err)
	}

	return handleDatabasesListGeneric(
		ctx,
		s,
		account.Client.ListMySQLDatabases,
		convertMySQLDatabaseToSummary,
		"MySQL",
		"mysql_databases_list",
		"failed to list MySQL databases",
	)
}

// handlePostgresDatabasesList lists all PostgreSQL databases.
func (s *Service) handlePostgresDatabasesList(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	account, err := s.accountManager.GetCurrentAccount()
	if err != nil {
		return nil, fmt.Errorf("failed to get current account: %w", err)
	}

	return handleDatabasesListGeneric(
		ctx,
		s,
		account.Client.ListPostgresDatabases,
		convertPostgresDatabaseToSummary,
		"PostgreSQL",
		"postgres_databases_list",
		"failed to list PostgreSQL databases",
	)
}

// TypedDatabaseSummary interface defines common methods for database summaries.
type TypedDatabaseSummary interface {
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
	GetUpdated() string
}

// GetID returns the database ID.
func (d MySQLDatabaseSummary) GetID() int { return d.ID }

// GetLabel returns the database label.
func (d MySQLDatabaseSummary) GetLabel() string { return d.Label }

// GetEngine returns the database engine.
func (d MySQLDatabaseSummary) GetEngine() string { return d.Engine }

// GetVersion returns the database version.
func (d MySQLDatabaseSummary) GetVersion() string { return d.Version }

// GetRegion returns the database region.
func (d MySQLDatabaseSummary) GetRegion() string { return d.Region }

// GetType returns the database type.
func (d MySQLDatabaseSummary) GetType() string { return d.Type }

// GetStatus returns the database status.
func (d MySQLDatabaseSummary) GetStatus() string { return d.Status }

// GetClusterSize returns the cluster size.
func (d MySQLDatabaseSummary) GetClusterSize() int { return d.ClusterSize }

// GetHosts returns the database hosts.
func (d MySQLDatabaseSummary) GetHosts() DatabaseHosts { return d.Hosts }

// GetPort returns the database port.
func (d MySQLDatabaseSummary) GetPort() int { return d.Port }

// GetUpdated returns the update time.
func (d MySQLDatabaseSummary) GetUpdated() string { return d.Updated }

// GetID returns the database ID.
func (d PostgresDatabaseSummary) GetID() int { return d.ID }

// GetLabel returns the database label.
func (d PostgresDatabaseSummary) GetLabel() string { return d.Label }

// GetEngine returns the database engine.
func (d PostgresDatabaseSummary) GetEngine() string { return d.Engine }

// GetVersion returns the database version.
func (d PostgresDatabaseSummary) GetVersion() string { return d.Version }

// GetRegion returns the database region.
func (d PostgresDatabaseSummary) GetRegion() string { return d.Region }

// GetType returns the database type.
func (d PostgresDatabaseSummary) GetType() string { return d.Type }

// GetStatus returns the database status.
func (d PostgresDatabaseSummary) GetStatus() string { return d.Status }

// GetClusterSize returns the cluster size.
func (d PostgresDatabaseSummary) GetClusterSize() int { return d.ClusterSize }

// GetHosts returns the database hosts.
func (d PostgresDatabaseSummary) GetHosts() DatabaseHosts { return d.Hosts }

// GetPort returns the database port.
func (d PostgresDatabaseSummary) GetPort() int { return d.Port }

// GetUpdated returns the update time.
func (d PostgresDatabaseSummary) GetUpdated() string { return d.Updated }

// formatMySQLDatabaseList formats MySQL database lists.
func (s *Service) formatTypedDatabaseList(engineType string, summaries interface{}) string {
	var stringBuilder strings.Builder

	switch dbSummaries := summaries.(type) {
	case []MySQLDatabaseSummary:
		stringBuilder.WriteString(fmt.Sprintf("Found %d %s databases:\n\n", len(dbSummaries), engineType))
		for _, database := range dbSummaries {
			s.formatSingleDatabaseSummary(&stringBuilder, database)
		}
	case []PostgresDatabaseSummary:
		stringBuilder.WriteString(fmt.Sprintf("Found %d %s databases:\n\n", len(dbSummaries), engineType))
		for _, database := range dbSummaries {
			s.formatSingleDatabaseSummary(&stringBuilder, database)
		}
	}

	return stringBuilder.String()
}

// formatSingleDatabaseSummary formats a single database summary entry.
func (s *Service) formatSingleDatabaseSummary(stringBuilder *strings.Builder, database TypedDatabaseSummary) {
	fmt.Fprintf(stringBuilder, "ID: %d | %s (%s %s)\n", database.GetID(), database.GetLabel(), database.GetEngine(), database.GetVersion())
	fmt.Fprintf(stringBuilder, "  Region: %s | Type: %s | Status: %s\n", database.GetRegion(), database.GetType(), database.GetStatus())
	fmt.Fprintf(stringBuilder, "  Primary Host: %s | Port: %d\n", database.GetHosts().Primary, database.GetPort())

	if database.GetHosts().Secondary != "" {
		fmt.Fprintf(stringBuilder, "  Secondary Host: %s\n", database.GetHosts().Secondary)
	}

	fmt.Fprintf(stringBuilder, "  Cluster Size: %d nodes | Updated: %s\n", database.GetClusterSize(), database.GetUpdated())
	stringBuilder.WriteString("\n")
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
	var stringBuilder strings.Builder

	fmt.Fprintf(&stringBuilder, "%s Database Details:\n", engineType)
	fmt.Fprintf(&stringBuilder, "ID: %d\n", detail.GetID())
	fmt.Fprintf(&stringBuilder, "Label: %s\n", detail.GetLabel())
	fmt.Fprintf(&stringBuilder, "Engine: %s %s\n", detail.GetEngine(), detail.GetVersion())
	fmt.Fprintf(&stringBuilder, "Region: %s\n", detail.GetRegion())
	fmt.Fprintf(&stringBuilder, "Type: %s\n", detail.GetType())
	fmt.Fprintf(&stringBuilder, "Status: %s\n", detail.GetStatus())
	fmt.Fprintf(&stringBuilder, "Cluster Size: %d nodes\n", detail.GetClusterSize())
	fmt.Fprintf(&stringBuilder, "Primary Host: %s\n", detail.GetHosts().Primary)

	if detail.GetHosts().Secondary != "" {
		fmt.Fprintf(&stringBuilder, "Secondary Host: %s\n", detail.GetHosts().Secondary)
	}

	fmt.Fprintf(&stringBuilder, "Port: %d\n", detail.GetPort())
	fmt.Fprintf(&stringBuilder, "Created: %s\n", detail.GetCreated())
	fmt.Fprintf(&stringBuilder, "Updated: %s\n\n", detail.GetUpdated())

	if len(detail.GetAllowList()) > 0 {
		fmt.Fprintf(&stringBuilder, "Allow List (IP ranges with access):\n")

		for _, ip := range detail.GetAllowList() {
			fmt.Fprintf(&stringBuilder, "  - %s\n", ip)
		}
	} else {
		fmt.Fprintf(&stringBuilder, "Allow List: No IP restrictions (open access)\n")
	}

	return stringBuilder.String()
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
	var parameters MySQLDatabaseGetParams
	if parseErr := parseArguments(request.Params.Arguments, &parameters); parseErr != nil {
		return nil, fmt.Errorf("invalid parameters: %w", parseErr)
	}

	account, accountErr := s.accountManager.GetCurrentAccount()
	if accountErr != nil {
		return nil, fmt.Errorf("failed to get current account: %w", accountErr)
	}

	database, databaseErr := account.Client.GetMySQLDatabase(ctx, parameters.DatabaseID)
	if databaseErr != nil {
		return nil, types.NewToolError("linode", "mysql_database_get", //nolint:wrapcheck // types.NewToolError already wraps the error
			"failed to get MySQL database", databaseErr)
	}

	detail := formatMySQLDetail(database)

	return mcp.NewToolResultText(formatDatabaseDetail("MySQL", detail)), nil
}

// handlePostgresDatabaseGet gets details of a specific PostgreSQL database.
func (s *Service) handlePostgresDatabaseGet(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var parameters PostgresDatabaseGetParams
	if parseErr := parseArguments(request.Params.Arguments, &parameters); parseErr != nil {
		return nil, fmt.Errorf("invalid parameters: %w", parseErr)
	}

	account, accountErr := s.accountManager.GetCurrentAccount()
	if accountErr != nil {
		return nil, fmt.Errorf("failed to get current account: %w", accountErr)
	}

	database, databaseErr := account.Client.GetPostgresDatabase(ctx, parameters.DatabaseID)
	if databaseErr != nil {
		return nil, types.NewToolError("linode", "postgres_database_get", //nolint:wrapcheck // types.NewToolError already wraps the error
			"failed to get PostgreSQL database", databaseErr)
	}

	detail := formatPostgresDetail(database)

	return mcp.NewToolResultText(formatDatabaseDetail("PostgreSQL", detail)), nil
}

// handleMySQLDatabaseCreate creates a new MySQL database.
func (s *Service) handleMySQLDatabaseCreate(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var parameters MySQLDatabaseCreateParams
	if parseErr := parseArguments(request.Params.Arguments, &parameters); parseErr != nil {
		return nil, fmt.Errorf("invalid parameters: %w", parseErr)
	}

	account, accountErr := s.accountManager.GetCurrentAccount()
	if accountErr != nil {
		return nil, fmt.Errorf("failed to get current account: %w", accountErr)
	}

	createOptions := linodego.MySQLCreateOptions{
		Label:       parameters.Label,
		Region:      parameters.Region,
		Type:        parameters.Type,
		Engine:      parameters.Engine,
		ClusterSize: parameters.ClusterSize,
		AllowList:   parameters.AllowList,
	}

	database, databaseErr := account.Client.CreateMySQLDatabase(ctx, createOptions)
	if databaseErr != nil {
		return nil, types.NewToolError("linode", "mysql_database_create", //nolint:wrapcheck // types.NewToolError already wraps the error
			"failed to create MySQL database", databaseErr)
	}

	return mcp.NewToolResultText(s.formatDatabaseCreateResult("MySQL", database.ID, database.Label, database.Engine, database.Version, database.Region, database.Type, string(database.Status), database.Hosts.Primary, database.Port)), nil
}

// handlePostgresDatabaseCreate creates a new PostgreSQL database.
func (s *Service) handlePostgresDatabaseCreate(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var parameters PostgresDatabaseCreateParams
	if parseErr := parseArguments(request.Params.Arguments, &parameters); parseErr != nil {
		return nil, fmt.Errorf("invalid parameters: %w", parseErr)
	}

	account, accountErr := s.accountManager.GetCurrentAccount()
	if accountErr != nil {
		return nil, fmt.Errorf("failed to get current account: %w", accountErr)
	}

	createOptions := linodego.PostgresCreateOptions{
		Label:       parameters.Label,
		Region:      parameters.Region,
		Type:        parameters.Type,
		Engine:      parameters.Engine,
		ClusterSize: parameters.ClusterSize,
		AllowList:   parameters.AllowList,
	}

	database, databaseErr := account.Client.CreatePostgresDatabase(ctx, createOptions)
	if databaseErr != nil {
		return nil, types.NewToolError("linode", "postgres_database_create", //nolint:wrapcheck // types.NewToolError already wraps the error
			"failed to create PostgreSQL database", databaseErr)
	}

	return mcp.NewToolResultText(s.formatDatabaseCreateResult("PostgreSQL", database.ID, database.Label, database.Engine, database.Version, database.Region, database.Type, string(database.Status), database.Hosts.Primary, database.Port)), nil
}

// formatDatabaseCreateResult formats database creation result in a standardized way.
func (s *Service) formatDatabaseCreateResult(engineType string, id int, label, engine, version, region, dbType, status, primaryHost string, port int) string {
	return fmt.Sprintf("%s database created successfully:\nID: %d\nLabel: %s\nEngine: %s %s\nRegion: %s\nType: %s\nStatus: %s\nPrimary Host: %s\nPort: %d",
		engineType, id, label, engine, version, region, dbType, status, primaryHost, port)
}

// formatResourceDeleteSuccess formats resource deletion success messages in a standardized way.
func (s *Service) formatResourceDeleteSuccess(resourceType string, id int, label string, properties map[string]string, additionalMessage string) string {
	var stringBuilder strings.Builder

	fmt.Fprintf(&stringBuilder, "%s deleted successfully!\n\nDeleted %s:\n", resourceType, resourceType)
	fmt.Fprintf(&stringBuilder, "- ID: %d\n", id)
	fmt.Fprintf(&stringBuilder, "- Label: %s\n", label)

	for key, value := range properties {
		fmt.Fprintf(&stringBuilder, "- %s: %s\n", key, value)
	}

	fmt.Fprintf(&stringBuilder, "\n%s", additionalMessage)

	return stringBuilder.String()
}

// formatDatabaseUpdateResult formats database update result in a standardized way.
func formatDatabaseUpdateResult(engineType string, database interface {
	GetID() int
	GetLabel() string
	GetStatus() string
	GetHosts() DatabaseHosts
},
) string {
	return fmt.Sprintf("%s database updated successfully:\nID: %d\nLabel: %s\nStatus: %s\nPrimary Host: %s",
		engineType, database.GetID(), database.GetLabel(), database.GetStatus(), database.GetHosts().Primary)
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
	var parameters MySQLDatabaseUpdateParams
	if parseErr := parseArguments(request.Params.Arguments, &parameters); parseErr != nil {
		return nil, fmt.Errorf("invalid parameters: %w", parseErr)
	}

	account, accountErr := s.accountManager.GetCurrentAccount()
	if accountErr != nil {
		return nil, fmt.Errorf("failed to get current account: %w", accountErr)
	}

	updateOptions := linodego.MySQLUpdateOptions{}
	if parameters.Label != "" {
		updateOptions.Label = parameters.Label
	}
	if parameters.AllowList != nil {
		updateOptions.AllowList = &parameters.AllowList
	}

	database, databaseErr := account.Client.UpdateMySQLDatabase(ctx, parameters.DatabaseID, updateOptions)
	if databaseErr != nil {
		return nil, types.NewToolError("linode", "mysql_database_update", //nolint:wrapcheck // types.NewToolError already wraps the error
			"failed to update MySQL database", databaseErr)
	}

	result := createDatabaseResult(database)

	return mcp.NewToolResultText(formatDatabaseUpdateResult("MySQL", result)), nil
}

// handlePostgresDatabaseUpdate updates an existing PostgreSQL database.
func (s *Service) handlePostgresDatabaseUpdate(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var parameters PostgresDatabaseUpdateParams
	if parseErr := parseArguments(request.Params.Arguments, &parameters); parseErr != nil {
		return nil, fmt.Errorf("invalid parameters: %w", parseErr)
	}

	account, accountErr := s.accountManager.GetCurrentAccount()
	if accountErr != nil {
		return nil, fmt.Errorf("failed to get current account: %w", accountErr)
	}

	updateOptions := linodego.PostgresUpdateOptions{}
	if parameters.Label != "" {
		updateOptions.Label = parameters.Label
	}
	if parameters.AllowList != nil {
		updateOptions.AllowList = &parameters.AllowList
	}

	database, databaseErr := account.Client.UpdatePostgresDatabase(ctx, parameters.DatabaseID, updateOptions)
	if databaseErr != nil {
		return nil, types.NewToolError("linode", "postgres_database_update", //nolint:wrapcheck // types.NewToolError already wraps the error
			"failed to update PostgreSQL database", databaseErr)
	}

	result := createDatabaseResult(database)

	return mcp.NewToolResultText(formatDatabaseUpdateResult("PostgreSQL", result)), nil
}

// handleMySQLDatabaseDelete deletes a MySQL database.
func (s *Service) handleMySQLDatabaseDelete(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var parameters MySQLDatabaseDeleteParams
	if parseErr := parseArguments(request.Params.Arguments, &parameters); parseErr != nil {
		return nil, fmt.Errorf("invalid parameters: %w", parseErr)
	}

	account, accountErr := s.accountManager.GetCurrentAccount()
	if accountErr != nil {
		return nil, fmt.Errorf("failed to get current account: %w", accountErr)
	}

	deleteErr := account.Client.DeleteMySQLDatabase(ctx, parameters.DatabaseID)
	if deleteErr != nil {
		return nil, types.NewToolError("linode", "mysql_database_delete", //nolint:wrapcheck // types.NewToolError already wraps the error
			"failed to delete MySQL database", deleteErr)
	}

	return mcp.NewToolResultText(fmt.Sprintf("MySQL database %d deleted successfully", parameters.DatabaseID)), nil
}

// handlePostgresDatabaseDelete deletes a PostgreSQL database.
func (s *Service) handlePostgresDatabaseDelete(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var parameters PostgresDatabaseDeleteParams
	if parseErr := parseArguments(request.Params.Arguments, &parameters); parseErr != nil {
		return nil, fmt.Errorf("invalid parameters: %w", parseErr)
	}

	account, accountErr := s.accountManager.GetCurrentAccount()
	if accountErr != nil {
		return nil, fmt.Errorf("failed to get current account: %w", accountErr)
	}

	deleteErr := account.Client.DeletePostgresDatabase(ctx, parameters.DatabaseID)
	if deleteErr != nil {
		return nil, types.NewToolError("linode", "postgres_database_delete", //nolint:wrapcheck // types.NewToolError already wraps the error
			"failed to delete PostgreSQL database", deleteErr)
	}

	return mcp.NewToolResultText(fmt.Sprintf("PostgreSQL database %d deleted successfully", parameters.DatabaseID)), nil
}

// formatDatabaseCredentials formats database credentials response with common text.
func (s *Service) formatDatabaseCredentials(dbType string, databaseID int, username, password string) *mcp.CallToolResult {
	var stringBuilder strings.Builder

	fmt.Fprintf(&stringBuilder, "%s Database Credentials:\n", dbType)
	fmt.Fprintf(&stringBuilder, "Username: %s\n", username)
	fmt.Fprintf(&stringBuilder, "Password: %s\n", password)
	fmt.Fprintf(&stringBuilder, "\nConnection Details:\n")
	fmt.Fprintf(&stringBuilder, "These are the root credentials for database %d.\n", databaseID)
	fmt.Fprintf(&stringBuilder, "Use these credentials to connect to your %s database.\n", dbType)
	fmt.Fprintf(&stringBuilder, "\nSecurity Note: Store these credentials securely and limit access.")

	return mcp.NewToolResultText(stringBuilder.String())
}

// handleMySQLDatabaseCredentials gets root credentials for a MySQL database.
func (s *Service) handleMySQLDatabaseCredentials(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var parameters MySQLDatabaseCredentialsParams
	if parseErr := parseArguments(request.Params.Arguments, &parameters); parseErr != nil {
		return nil, fmt.Errorf("invalid parameters: %w", parseErr)
	}

	account, accountErr := s.accountManager.GetCurrentAccount()
	if accountErr != nil {
		return nil, fmt.Errorf("failed to get current account: %w", accountErr)
	}

	credentials, credentialsErr := account.Client.GetMySQLDatabaseCredentials(ctx, parameters.DatabaseID)
	if credentialsErr != nil {
		return nil, types.NewToolError("linode", "mysql_database_credentials", //nolint:wrapcheck // types.NewToolError already wraps the error
			"failed to get MySQL database credentials", credentialsErr)
	}

	return s.formatDatabaseCredentials("MySQL", parameters.DatabaseID, credentials.Username, credentials.Password), nil
}

// handlePostgresDatabaseCredentials gets root credentials for a PostgreSQL database.
func (s *Service) handlePostgresDatabaseCredentials(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var parameters PostgresDatabaseCredentialsParams
	if parseErr := parseArguments(request.Params.Arguments, &parameters); parseErr != nil {
		return nil, fmt.Errorf("invalid parameters: %w", parseErr)
	}

	account, accountErr := s.accountManager.GetCurrentAccount()
	if accountErr != nil {
		return nil, fmt.Errorf("failed to get current account: %w", accountErr)
	}

	credentials, credentialsErr := account.Client.GetPostgresDatabaseCredentials(ctx, parameters.DatabaseID)
	if credentialsErr != nil {
		return nil, types.NewToolError("linode", "postgres_database_credentials", //nolint:wrapcheck // types.NewToolError already wraps the error
			"failed to get PostgreSQL database credentials", credentialsErr)
	}

	return s.formatDatabaseCredentials("PostgreSQL", parameters.DatabaseID, credentials.Username, credentials.Password), nil
}

// handleMySQLDatabaseCredentialsReset resets root password for a MySQL database.
func (s *Service) handleMySQLDatabaseCredentialsReset(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var parameters MySQLDatabaseCredentialsResetParams
	if parseErr := parseArguments(request.Params.Arguments, &parameters); parseErr != nil {
		return nil, fmt.Errorf("invalid parameters: %w", parseErr)
	}

	account, accountErr := s.accountManager.GetCurrentAccount()
	if accountErr != nil {
		return nil, fmt.Errorf("failed to get current account: %w", accountErr)
	}

	resetErr := account.Client.ResetMySQLDatabaseCredentials(ctx, parameters.DatabaseID)
	if resetErr != nil {
		return nil, types.NewToolError("linode", "mysql_database_credentials_reset", //nolint:wrapcheck // types.NewToolError already wraps the error
			"failed to reset MySQL database credentials", resetErr)
	}

	return mcp.NewToolResultText(fmt.Sprintf("MySQL database %d root password reset successfully.\nRetrieve new credentials using the credentials command.", parameters.DatabaseID)), nil
}

// handlePostgresDatabaseCredentialsReset resets root password for a PostgreSQL database.
func (s *Service) handlePostgresDatabaseCredentialsReset(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var parameters PostgresDatabaseCredentialsResetParams
	if parseErr := parseArguments(request.Params.Arguments, &parameters); parseErr != nil {
		return nil, fmt.Errorf("invalid parameters: %w", parseErr)
	}

	account, accountErr := s.accountManager.GetCurrentAccount()
	if accountErr != nil {
		return nil, fmt.Errorf("failed to get current account: %w", accountErr)
	}

	resetErr := account.Client.ResetPostgresDatabaseCredentials(ctx, parameters.DatabaseID)
	if resetErr != nil {
		return nil, types.NewToolError("linode", "postgres_database_credentials_reset", //nolint:wrapcheck // types.NewToolError already wraps the error
			"failed to reset PostgreSQL database credentials", resetErr)
	}

	return mcp.NewToolResultText(fmt.Sprintf("PostgreSQL database %d root password reset successfully.\nRetrieve new credentials using the credentials command.", parameters.DatabaseID)), nil
}

// handleDatabaseEnginesList lists all available database engines.
func (s *Service) handleDatabaseEnginesList(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	account, accountErr := s.accountManager.GetCurrentAccount()
	if accountErr != nil {
		return nil, fmt.Errorf("failed to get current account: %w", accountErr)
	}

	engines, enginesErr := account.Client.ListDatabaseEngines(ctx, nil)
	if enginesErr != nil {
		return nil, types.NewToolError("linode", "database_engines_list", //nolint:wrapcheck // types.NewToolError already wraps the error
			"failed to list database engines", enginesErr)
	}

	var stringBuilder strings.Builder

	stringBuilder.WriteString(fmt.Sprintf("Found %d database engines:\n\n", len(engines)))

	for _, engine := range engines {
		fmt.Fprintf(&stringBuilder, "Engine: %s\n", engine.Engine)
		fmt.Fprintf(&stringBuilder, "  ID: %s\n", engine.ID)
		fmt.Fprintf(&stringBuilder, "  Version: %s\n", engine.Version)
		// Note: DatabaseEngine struct doesn't include status field, showing generic status.
		fmt.Fprintf(&stringBuilder, "  Status: Active\n")
		stringBuilder.WriteString("\n")
	}

	return mcp.NewToolResultText(stringBuilder.String()), nil
}

// handleDatabaseTypesList lists all available database types.
func (s *Service) handleDatabaseTypesList(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	account, accountErr := s.accountManager.GetCurrentAccount()
	if accountErr != nil {
		return nil, fmt.Errorf("failed to get current account: %w", accountErr)
	}

	databaseTypes, databaseTypesErr := account.Client.ListDatabaseTypes(ctx, nil)
	if databaseTypesErr != nil {
		return nil, types.NewToolError("linode", "database_types_list", //nolint:wrapcheck // types.NewToolError already wraps the error
			"failed to list database types", databaseTypesErr)
	}

	var stringBuilder strings.Builder

	stringBuilder.WriteString(fmt.Sprintf("Found %d database types:\n\n", len(databaseTypes)))

	for _, databaseType := range databaseTypes {
		fmt.Fprintf(&stringBuilder, "Type: %s\n", databaseType.ID)
		fmt.Fprintf(&stringBuilder, "  Label: %s\n", databaseType.Label)
		fmt.Fprintf(&stringBuilder, "  Class: %s\n", databaseType.Class)
		fmt.Fprintf(&stringBuilder, "  Disk: %d GB\n", databaseType.Disk)
		fmt.Fprintf(&stringBuilder, "  Memory: %d MB\n", databaseType.Memory)
		// Note: VCPUs and Price fields may not be available in all linodego versions.
		// Note: Engines field structure may vary in linodego versions.
		stringBuilder.WriteString("\n")
	}

	return mcp.NewToolResultText(stringBuilder.String()), nil
}
