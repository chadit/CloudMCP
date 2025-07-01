// Mock file returns test errors without wrapping
package mocks

import (
	"context"

	"github.com/linode/linodego"
	"github.com/stretchr/testify/mock"
)

// MockClient provides a mock implementation of Linode client for testing.
type MockClient struct {
	mock.Mock
}

// ListRegions mocks the ListRegions method.
func (m *MockClient) ListRegions(ctx context.Context, opts *linodego.ListOptions) ([]linodego.Region, error) {
	args := m.Called(ctx, opts)

	regions, ok := args.Get(0).([]linodego.Region)
	if !ok {
		//nolint:wrapcheck // mock does not need to wrap errors.
		return nil, args.Error(1)
	}

	//nolint:wrapcheck // mock does not need to wrap errors.
	return regions, args.Error(1)
}

// ListTypes mocks the ListTypes method.
func (m *MockClient) ListTypes(ctx context.Context, opts *linodego.ListOptions) ([]linodego.LinodeType, error) {
	args := m.Called(ctx, opts)

	types, ok := args.Get(0).([]linodego.LinodeType)
	if !ok {
		//nolint:wrapcheck // mock does not need to wrap errors.
		return nil, args.Error(1)
	}

	//nolint:wrapcheck // mock does not need to wrap errors.
	return types, args.Error(1)
}

// ListKernels mocks the ListKernels method.
func (m *MockClient) ListKernels(ctx context.Context, opts *linodego.ListOptions) ([]linodego.LinodeKernel, error) {
	args := m.Called(ctx, opts)

	kernels, ok := args.Get(0).([]linodego.LinodeKernel)
	if !ok {
		//nolint:wrapcheck // mock does not need to wrap errors.
		return nil, args.Error(1)
	}

	//nolint:wrapcheck // mock does not need to wrap errors.
	return kernels, args.Error(1)
}
