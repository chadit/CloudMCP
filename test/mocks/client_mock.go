//nolint:wrapcheck // Mock file returns test errors without wrapping
package mocks

import (
	"context"

	"github.com/linode/linodego"
	"github.com/stretchr/testify/mock"
)

// MockLinodeClient provides a mock implementation of Linode client for testing.
type MockLinodeClient struct {
	mock.Mock
}

// ListRegions mocks the ListRegions method.
func (m *MockLinodeClient) ListRegions(ctx context.Context, opts *linodego.ListOptions) ([]linodego.Region, error) {
	args := m.Called(ctx, opts)

	regions, ok := args.Get(0).([]linodego.Region)
	if !ok {
		return nil, args.Error(1)
	}

	return regions, args.Error(1)
}

// ListTypes mocks the ListTypes method.
func (m *MockLinodeClient) ListTypes(ctx context.Context, opts *linodego.ListOptions) ([]linodego.LinodeType, error) {
	args := m.Called(ctx, opts)

	types, ok := args.Get(0).([]linodego.LinodeType)
	if !ok {
		return nil, args.Error(1)
	}

	return types, args.Error(1)
}

// ListKernels mocks the ListKernels method.
func (m *MockLinodeClient) ListKernels(ctx context.Context, opts *linodego.ListOptions) ([]linodego.LinodeKernel, error) {
	args := m.Called(ctx, opts)

	kernels, ok := args.Get(0).([]linodego.LinodeKernel)
	if !ok {
		return nil, args.Error(1)
	}

	return kernels, args.Error(1)
}
