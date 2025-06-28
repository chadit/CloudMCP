package mocks

import (
	"context"

	"github.com/linode/linodego"
	"github.com/stretchr/testify/mock"
)

// MockDatabaseService provides a mock implementation of Linode database operations.
// This mock implements the subset of linodego.Client methods used by database tools.
type MockDatabaseService struct {
	mock.Mock
}

// Generic database methods
func (m *MockDatabaseService) ListDatabases(ctx context.Context, opts *linodego.ListOptions) ([]linodego.Database, error) {
	args := m.Called(ctx, opts)
	return args.Get(0).([]linodego.Database), args.Error(1)
}

func (m *MockDatabaseService) ListDatabaseEngines(ctx context.Context, opts *linodego.ListOptions) ([]linodego.DatabaseEngine, error) {
	args := m.Called(ctx, opts)
	return args.Get(0).([]linodego.DatabaseEngine), args.Error(1)
}

func (m *MockDatabaseService) ListDatabaseTypes(ctx context.Context, opts *linodego.ListOptions) ([]linodego.DatabaseType, error) {
	args := m.Called(ctx, opts)
	return args.Get(0).([]linodego.DatabaseType), args.Error(1)
}

// MySQL-specific methods
func (m *MockDatabaseService) ListMySQLDatabases(ctx context.Context, opts *linodego.ListOptions) ([]linodego.MySQLDatabase, error) {
	args := m.Called(ctx, opts)
	return args.Get(0).([]linodego.MySQLDatabase), args.Error(1)
}

func (m *MockDatabaseService) GetMySQLDatabase(ctx context.Context, databaseID int) (*linodego.MySQLDatabase, error) {
	args := m.Called(ctx, databaseID)
	return args.Get(0).(*linodego.MySQLDatabase), args.Error(1)
}

func (m *MockDatabaseService) CreateMySQLDatabase(ctx context.Context, opts linodego.MySQLCreateOptions) (*linodego.MySQLDatabase, error) {
	args := m.Called(ctx, opts)
	return args.Get(0).(*linodego.MySQLDatabase), args.Error(1)
}

func (m *MockDatabaseService) UpdateMySQLDatabase(ctx context.Context, databaseID int, opts linodego.MySQLUpdateOptions) (*linodego.MySQLDatabase, error) {
	args := m.Called(ctx, databaseID, opts)
	return args.Get(0).(*linodego.MySQLDatabase), args.Error(1)
}

func (m *MockDatabaseService) DeleteMySQLDatabase(ctx context.Context, databaseID int) error {
	args := m.Called(ctx, databaseID)
	return args.Error(0)
}

func (m *MockDatabaseService) GetMySQLDatabaseCredentials(ctx context.Context, databaseID int) (*linodego.MySQLDatabaseCredential, error) {
	args := m.Called(ctx, databaseID)
	return args.Get(0).(*linodego.MySQLDatabaseCredential), args.Error(1)
}

func (m *MockDatabaseService) ResetMySQLDatabaseCredentials(ctx context.Context, databaseID int) error {
	args := m.Called(ctx, databaseID)
	return args.Error(0)
}

// PostgreSQL-specific methods
func (m *MockDatabaseService) ListPostgresDatabases(ctx context.Context, opts *linodego.ListOptions) ([]linodego.PostgresDatabase, error) {
	args := m.Called(ctx, opts)
	return args.Get(0).([]linodego.PostgresDatabase), args.Error(1)
}

func (m *MockDatabaseService) GetPostgresDatabase(ctx context.Context, databaseID int) (*linodego.PostgresDatabase, error) {
	args := m.Called(ctx, databaseID)
	return args.Get(0).(*linodego.PostgresDatabase), args.Error(1)
}

func (m *MockDatabaseService) CreatePostgresDatabase(ctx context.Context, opts linodego.PostgresCreateOptions) (*linodego.PostgresDatabase, error) {
	args := m.Called(ctx, opts)
	return args.Get(0).(*linodego.PostgresDatabase), args.Error(1)
}

func (m *MockDatabaseService) UpdatePostgresDatabase(ctx context.Context, databaseID int, opts linodego.PostgresUpdateOptions) (*linodego.PostgresDatabase, error) {
	args := m.Called(ctx, databaseID, opts)
	return args.Get(0).(*linodego.PostgresDatabase), args.Error(1)
}

func (m *MockDatabaseService) DeletePostgresDatabase(ctx context.Context, databaseID int) error {
	args := m.Called(ctx, databaseID)
	return args.Error(0)
}

func (m *MockDatabaseService) GetPostgresDatabaseCredentials(ctx context.Context, databaseID int) (*linodego.PostgresDatabaseCredential, error) {
	args := m.Called(ctx, databaseID)
	return args.Get(0).(*linodego.PostgresDatabaseCredential), args.Error(1)
}

func (m *MockDatabaseService) ResetPostgresDatabaseCredentials(ctx context.Context, databaseID int) error {
	args := m.Called(ctx, databaseID)
	return args.Error(0)
}

// Helper methods for setting up common test scenarios

// MySQL helper methods
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

// PostgreSQL helper methods
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
