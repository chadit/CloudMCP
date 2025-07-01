// Mock file returns test errors without wrapping
package mocks

import (
	"context"

	"github.com/linode/linodego"
	"github.com/stretchr/testify/mock"
)

// MockDomainService provides a mock implementation of Linode domain (DNS) operations.
// This mock implements the subset of linodego.Client methods used by domain tools.
type MockDomainService struct {
	mock.Mock
}

// Domain methods.
func (m *MockDomainService) ListDomains(ctx context.Context, opts *linodego.ListOptions) ([]linodego.Domain, error) {
	args := m.Called(ctx, opts)

	domains, ok := args.Get(0).([]linodego.Domain)
	if !ok {
		//nolint:wrapcheck // mock does not need to wrap errors.
		return nil, args.Error(1)
	}

	//nolint:wrapcheck // mock does not need to wrap errors.
	return domains, args.Error(1)
}

func (m *MockDomainService) GetDomain(ctx context.Context, domainID int) (*linodego.Domain, error) {
	args := m.Called(ctx, domainID)

	domain, ok := args.Get(0).(*linodego.Domain)
	if !ok {
		//nolint:wrapcheck // mock does not need to wrap errors.
		return nil, args.Error(1)
	}

	//nolint:wrapcheck // mock does not need to wrap errors.
	return domain, args.Error(1)
}

func (m *MockDomainService) CreateDomain(ctx context.Context, opts linodego.DomainCreateOptions) (*linodego.Domain, error) {
	args := m.Called(ctx, opts)

	domain, ok := args.Get(0).(*linodego.Domain)
	if !ok {
		//nolint:wrapcheck // mock does not need to wrap errors.
		return nil, args.Error(1)
	}

	//nolint:wrapcheck // mock does not need to wrap errors.
	return domain, args.Error(1)
}

func (m *MockDomainService) UpdateDomain(ctx context.Context, domainID int, opts linodego.DomainUpdateOptions) (*linodego.Domain, error) {
	args := m.Called(ctx, domainID, opts)

	domain, ok := args.Get(0).(*linodego.Domain)
	if !ok {
		//nolint:wrapcheck // mock does not need to wrap errors.
		return nil, args.Error(1)
	}

	//nolint:wrapcheck // mock does not need to wrap errors.
	return domain, args.Error(1)
}

func (m *MockDomainService) DeleteDomain(ctx context.Context, domainID int) error {
	args := m.Called(ctx, domainID)

	//nolint:wrapcheck // mock does not need to wrap errors.
	return args.Error(0)
}

// Domain record methods.
func (m *MockDomainService) ListDomainRecords(ctx context.Context, domainID int, opts *linodego.ListOptions) ([]linodego.DomainRecord, error) {
	args := m.Called(ctx, domainID, opts)

	records, ok := args.Get(0).([]linodego.DomainRecord)
	if !ok {
		//nolint:wrapcheck // mock does not need to wrap errors.
		return nil, args.Error(1)
	}

	//nolint:wrapcheck // mock does not need to wrap errors.
	return records, args.Error(1)
}

func (m *MockDomainService) GetDomainRecord(ctx context.Context, domainID int, recordID int) (*linodego.DomainRecord, error) {
	args := m.Called(ctx, domainID, recordID)

	record, ok := args.Get(0).(*linodego.DomainRecord)
	if !ok {
		//nolint:wrapcheck // mock does not need to wrap errors.
		return nil, args.Error(1)
	}

	//nolint:wrapcheck // mock does not need to wrap errors.
	return record, args.Error(1)
}

func (m *MockDomainService) CreateDomainRecord(ctx context.Context, domainID int, opts linodego.DomainRecordCreateOptions) (*linodego.DomainRecord, error) {
	args := m.Called(ctx, domainID, opts)

	record, ok := args.Get(0).(*linodego.DomainRecord)
	if !ok {
		//nolint:wrapcheck // mock does not need to wrap errors.
		return nil, args.Error(1)
	}

	//nolint:wrapcheck // mock does not need to wrap errors.
	return record, args.Error(1)
}

func (m *MockDomainService) UpdateDomainRecord(ctx context.Context, domainID int, recordID int, opts linodego.DomainRecordUpdateOptions) (*linodego.DomainRecord, error) {
	args := m.Called(ctx, domainID, recordID, opts)

	record, ok := args.Get(0).(*linodego.DomainRecord)
	if !ok {
		//nolint:wrapcheck // mock does not need to wrap errors.
		return nil, args.Error(1)
	}

	//nolint:wrapcheck // mock does not need to wrap errors.
	return record, args.Error(1)
}

func (m *MockDomainService) DeleteDomainRecord(ctx context.Context, domainID int, recordID int) error {
	args := m.Called(ctx, domainID, recordID)

	//nolint:wrapcheck // mock does not need to wrap errors.
	return args.Error(0)
}

// Helper methods for setting up common test scenarios

func (m *MockDomainService) SetupListDomainsSuccess(domains []linodego.Domain) {
	m.On("ListDomains", mock.Anything, mock.Anything).Return(domains, nil)
}

func (m *MockDomainService) SetupListDomainsError(err error) {
	m.On("ListDomains", mock.Anything, mock.Anything).Return([]linodego.Domain{}, err)
}

func (m *MockDomainService) SetupGetDomainSuccess(domainID int, domain *linodego.Domain) {
	m.On("GetDomain", mock.Anything, domainID).Return(domain, nil)
}

func (m *MockDomainService) SetupGetDomainError(domainID int, err error) {
	m.On("GetDomain", mock.Anything, domainID).Return((*linodego.Domain)(nil), err)
}

func (m *MockDomainService) SetupCreateDomainSuccess(opts linodego.DomainCreateOptions, domain *linodego.Domain) {
	m.On("CreateDomain", mock.Anything, opts).Return(domain, nil)
}

func (m *MockDomainService) SetupCreateDomainError(opts linodego.DomainCreateOptions, err error) {
	m.On("CreateDomain", mock.Anything, opts).Return((*linodego.Domain)(nil), err)
}

func (m *MockDomainService) SetupListDomainRecordsSuccess(domainID int, records []linodego.DomainRecord) {
	m.On("ListDomainRecords", mock.Anything, domainID, mock.Anything).Return(records, nil)
}

func (m *MockDomainService) SetupListDomainRecordsError(domainID int, err error) {
	m.On("ListDomainRecords", mock.Anything, domainID, mock.Anything).Return([]linodego.DomainRecord{}, err)
}

func (m *MockDomainService) SetupGetDomainRecordSuccess(domainID int, recordID int, record *linodego.DomainRecord) {
	m.On("GetDomainRecord", mock.Anything, domainID, recordID).Return(record, nil)
}

func (m *MockDomainService) SetupGetDomainRecordError(domainID int, recordID int, err error) {
	m.On("GetDomainRecord", mock.Anything, domainID, recordID).Return((*linodego.DomainRecord)(nil), err)
}

func (m *MockDomainService) SetupCreateDomainRecordSuccess(domainID int, opts linodego.DomainRecordCreateOptions, record *linodego.DomainRecord) {
	m.On("CreateDomainRecord", mock.Anything, domainID, opts).Return(record, nil)
}

func (m *MockDomainService) SetupCreateDomainRecordError(domainID int, opts linodego.DomainRecordCreateOptions, err error) {
	m.On("CreateDomainRecord", mock.Anything, domainID, opts).Return((*linodego.DomainRecord)(nil), err)
}
