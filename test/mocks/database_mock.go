// Mock file returns test errors without wrapping
package mocks

import (
	"context"
	"fmt"

	"github.com/linode/linodego"
	"github.com/stretchr/testify/mock"
)

// MockDatabaseService provides a mock implementation of Linode database operations.
// This mock implements the subset of linodego.Client methods used by database tools.
type MockDatabaseService struct {
	mock.Mock
}

// Generic database methods.
func (m *MockDatabaseService) ListDatabases(ctx context.Context, opts *linodego.ListOptions) ([]linodego.Database, error) {
	args := m.Called(ctx, opts)

	if result := args.Get(0); result != nil {
		databases, ok := result.([]linodego.Database)
		if !ok {
			//nolint:wrapcheck // mock does not need to wrap errors.
			return nil, args.Error(1)
		}

		//nolint:wrapcheck // mock does not need to wrap errors.
		return databases, args.Error(1)
	}

	if err := args.Error(1); err != nil {
		return nil, fmt.Errorf("mock database error: %w", err)
	}

	return nil, nil
}

func (m *MockDatabaseService) ListDatabaseEngines(ctx context.Context, opts *linodego.ListOptions) ([]linodego.DatabaseEngine, error) {
	args := m.Called(ctx, opts)

	if result := args.Get(0); result != nil {
		engines, ok := result.([]linodego.DatabaseEngine)
		if !ok {
			//nolint:wrapcheck // mock does not need to wrap errors.
			return nil, args.Error(1)
		}

		//nolint:wrapcheck // mock does not need to wrap errors.
		return engines, args.Error(1)
	}

	if err := args.Error(1); err != nil {
		return nil, fmt.Errorf("mock database error: %w", err)
	}

	return nil, nil
}

func (m *MockDatabaseService) ListDatabaseTypes(ctx context.Context, opts *linodego.ListOptions) ([]linodego.DatabaseType, error) {
	args := m.Called(ctx, opts)

	if result := args.Get(0); result != nil {
		types, ok := result.([]linodego.DatabaseType)
		if !ok {
			//nolint:wrapcheck // mock does not need to wrap errors.
			return nil, args.Error(1)
		}

		//nolint:wrapcheck // mock does not need to wrap errors.
		return types, args.Error(1)
	}

	if err := args.Error(1); err != nil {
		return nil, fmt.Errorf("mock database error: %w", err)
	}

	return nil, nil
}

// MySQL-specific methods.
func (m *MockDatabaseService) ListMySQLDatabases(ctx context.Context, opts *linodego.ListOptions) ([]linodego.MySQLDatabase, error) {
	args := m.Called(ctx, opts)

	mysqlDatabases, ok := args.Get(0).([]linodego.MySQLDatabase)
	if !ok {
		//nolint:wrapcheck // mock does not need to wrap errors.
		return nil, args.Error(1)
	}

	//nolint:wrapcheck // mock does not need to wrap errors.
	return mysqlDatabases, args.Error(1)
}

func (m *MockDatabaseService) GetMySQLDatabase(ctx context.Context, databaseID int) (*linodego.MySQLDatabase, error) {
	args := m.Called(ctx, databaseID)

	mysqlDB, ok := args.Get(0).(*linodego.MySQLDatabase)
	if !ok {
		//nolint:wrapcheck // mock does not need to wrap errors.
		return nil, args.Error(1)
	}

	//nolint:wrapcheck // mock does not need to wrap errors.
	return mysqlDB, args.Error(1)
}

func (m *MockDatabaseService) CreateMySQLDatabase(ctx context.Context, opts linodego.MySQLCreateOptions) (*linodego.MySQLDatabase, error) {
	args := m.Called(ctx, opts)

	mysqlDB, ok := args.Get(0).(*linodego.MySQLDatabase)
	if !ok {
		//nolint:wrapcheck // mock does not need to wrap errors.
		return nil, args.Error(1)
	}

	//nolint:wrapcheck // mock does not need to wrap errors.
	return mysqlDB, args.Error(1)
}

func (m *MockDatabaseService) UpdateMySQLDatabase(ctx context.Context, databaseID int, opts linodego.MySQLUpdateOptions) (*linodego.MySQLDatabase, error) {
	args := m.Called(ctx, databaseID, opts)

	mysqlDB, ok := args.Get(0).(*linodego.MySQLDatabase)
	if !ok {
		//nolint:wrapcheck // mock does not need to wrap errors.
		return nil, args.Error(1)
	}

	//nolint:wrapcheck // mock does not need to wrap errors.
	return mysqlDB, args.Error(1)
}

func (m *MockDatabaseService) DeleteMySQLDatabase(ctx context.Context, databaseID int) error {
	args := m.Called(ctx, databaseID)

	//nolint:wrapcheck // mock does not need to wrap errors.
	return args.Error(0)
}

func (m *MockDatabaseService) GetMySQLDatabaseCredentials(ctx context.Context, databaseID int) (*linodego.MySQLDatabaseCredential, error) {
	args := m.Called(ctx, databaseID)

	creds, ok := args.Get(0).(*linodego.MySQLDatabaseCredential)
	if !ok {
		//nolint:wrapcheck // mock does not need to wrap errors.
		return nil, args.Error(1)
	}

	//nolint:wrapcheck // mock does not need to wrap errors.
	return creds, args.Error(1)
}

func (m *MockDatabaseService) ResetMySQLDatabaseCredentials(ctx context.Context, databaseID int) error {
	args := m.Called(ctx, databaseID)

	//nolint:wrapcheck // mock does not need to wrap errors.
	return args.Error(0)
}

// PostgreSQL-specific methods.
func (m *MockDatabaseService) ListPostgresDatabases(ctx context.Context, opts *linodego.ListOptions) ([]linodego.PostgresDatabase, error) {
	args := m.Called(ctx, opts)

	postgresDatabases, ok := args.Get(0).([]linodego.PostgresDatabase)
	if !ok {
		//nolint:wrapcheck // mock does not need to wrap errors.
		return nil, args.Error(1)
	}

	//nolint:wrapcheck // mock does not need to wrap errors.
	return postgresDatabases, args.Error(1)
}

func (m *MockDatabaseService) GetPostgresDatabase(ctx context.Context, databaseID int) (*linodego.PostgresDatabase, error) {
	args := m.Called(ctx, databaseID)

	postgresDB, ok := args.Get(0).(*linodego.PostgresDatabase)
	if !ok {
		//nolint:wrapcheck // mock does not need to wrap errors.
		return nil, args.Error(1)
	}

	//nolint:wrapcheck // mock does not need to wrap errors.
	return postgresDB, args.Error(1)
}

func (m *MockDatabaseService) CreatePostgresDatabase(ctx context.Context, opts linodego.PostgresCreateOptions) (*linodego.PostgresDatabase, error) {
	args := m.Called(ctx, opts)

	postgresDB, ok := args.Get(0).(*linodego.PostgresDatabase)
	if !ok {
		//nolint:wrapcheck // mock does not need to wrap errors.
		return nil, args.Error(1)
	}

	//nolint:wrapcheck // mock does not need to wrap errors.
	return postgresDB, args.Error(1)
}

func (m *MockDatabaseService) UpdatePostgresDatabase(ctx context.Context, databaseID int, opts linodego.PostgresUpdateOptions) (*linodego.PostgresDatabase, error) {
	args := m.Called(ctx, databaseID, opts)

	postgresDB, ok := args.Get(0).(*linodego.PostgresDatabase)
	if !ok {
		//nolint:wrapcheck // mock does not need to wrap errors.
		return nil, args.Error(1)
	}

	//nolint:wrapcheck // mock does not need to wrap errors.
	return postgresDB, args.Error(1)
}

func (m *MockDatabaseService) DeletePostgresDatabase(ctx context.Context, databaseID int) error {
	args := m.Called(ctx, databaseID)

	//nolint:wrapcheck // mock does not need to wrap errors.
	return args.Error(0)
}

func (m *MockDatabaseService) GetPostgresDatabaseCredentials(ctx context.Context, databaseID int) (*linodego.PostgresDatabaseCredential, error) {
	args := m.Called(ctx, databaseID)

	creds, ok := args.Get(0).(*linodego.PostgresDatabaseCredential)
	if !ok {
		//nolint:wrapcheck // mock does not need to wrap errors.
		return nil, args.Error(1)
	}

	//nolint:wrapcheck // mock does not need to wrap errors.
	return creds, args.Error(1)
}

func (m *MockDatabaseService) ResetPostgresDatabaseCredentials(ctx context.Context, databaseID int) error {
	args := m.Called(ctx, databaseID)

	//nolint:wrapcheck // mock does not need to wrap errors.
	return args.Error(0)
}

// Helper methods for setting up common test scenarios

// MySQL helper methods.
func (m *MockDatabaseService) SetupListMySQLDatabasesSuccess(databases []linodego.MySQLDatabase) {
	m.On("ListMySQLDatabases", mock.Anything, mock.Anything).Return(databases, nil)
}

func (m *MockDatabaseService) SetupListMySQLDatabasesError(err error) {
	m.On("ListMySQLDatabases", mock.Anything, mock.Anything).Return([]linodego.MySQLDatabase{}, err)
}

func (m *MockDatabaseService) SetupGetMySQLDatabaseSuccess(databaseID int, database *linodego.MySQLDatabase) {
	m.On("GetMySQLDatabase", mock.Anything, databaseID).Return(database, nil)
}

func (m *MockDatabaseService) SetupGetMySQLDatabaseError(databaseID int, err error) {
	m.On("GetMySQLDatabase", mock.Anything, databaseID).Return((*linodego.MySQLDatabase)(nil), err)
}

func (m *MockDatabaseService) SetupCreateMySQLDatabaseSuccess(opts linodego.MySQLCreateOptions, database *linodego.MySQLDatabase) {
	m.On("CreateMySQLDatabase", mock.Anything, opts).Return(database, nil)
}

func (m *MockDatabaseService) SetupCreateMySQLDatabaseError(opts linodego.MySQLCreateOptions, err error) {
	m.On("CreateMySQLDatabase", mock.Anything, opts).Return((*linodego.MySQLDatabase)(nil), err)
}

// PostgreSQL helper methods.
func (m *MockDatabaseService) SetupListPostgresDatabasesSuccess(databases []linodego.PostgresDatabase) {
	m.On("ListPostgresDatabases", mock.Anything, mock.Anything).Return(databases, nil)
}

func (m *MockDatabaseService) SetupListPostgresDatabasesError(err error) {
	m.On("ListPostgresDatabases", mock.Anything, mock.Anything).Return([]linodego.PostgresDatabase{}, err)
}

func (m *MockDatabaseService) SetupGetPostgresDatabaseSuccess(databaseID int, database *linodego.PostgresDatabase) {
	m.On("GetPostgresDatabase", mock.Anything, databaseID).Return(database, nil)
}

func (m *MockDatabaseService) SetupGetPostgresDatabaseError(databaseID int, err error) {
	m.On("GetPostgresDatabase", mock.Anything, databaseID).Return((*linodego.PostgresDatabase)(nil), err)
}

func (m *MockDatabaseService) SetupCreatePostgresDatabaseSuccess(opts linodego.PostgresCreateOptions, database *linodego.PostgresDatabase) {
	m.On("CreatePostgresDatabase", mock.Anything, opts).Return(database, nil)
}

func (m *MockDatabaseService) SetupCreatePostgresDatabaseError(opts linodego.PostgresCreateOptions, err error) {
	m.On("CreatePostgresDatabase", mock.Anything, opts).Return((*linodego.PostgresDatabase)(nil), err)
}
