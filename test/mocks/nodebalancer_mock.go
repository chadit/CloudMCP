package mocks

import (
	"context"

	"github.com/linode/linodego"
	"github.com/stretchr/testify/mock"
)

// MockNodeBalancerService provides a mock implementation of Linode NodeBalancer operations.
// This mock implements the subset of linodego.Client methods used by NodeBalancer tools.
type MockNodeBalancerService struct {
	mock.Mock
}

// ListNodeBalancers mocks listing all NodeBalancers.
func (m *MockNodeBalancerService) ListNodeBalancers(ctx context.Context, opts *linodego.ListOptions) ([]linodego.NodeBalancer, error) {
	args := m.Called(ctx, opts)
	return args.Get(0).([]linodego.NodeBalancer), args.Error(1)
}

// GetNodeBalancer mocks getting details of a specific NodeBalancer.
func (m *MockNodeBalancerService) GetNodeBalancer(ctx context.Context, nodebalancerID int) (*linodego.NodeBalancer, error) {
	args := m.Called(ctx, nodebalancerID)
	return args.Get(0).(*linodego.NodeBalancer), args.Error(1)
}

// CreateNodeBalancer mocks creating a new NodeBalancer.
func (m *MockNodeBalancerService) CreateNodeBalancer(ctx context.Context, opts linodego.NodeBalancerCreateOptions) (*linodego.NodeBalancer, error) {
	args := m.Called(ctx, opts)
	return args.Get(0).(*linodego.NodeBalancer), args.Error(1)
}

// UpdateNodeBalancer mocks updating an existing NodeBalancer.
func (m *MockNodeBalancerService) UpdateNodeBalancer(ctx context.Context, nodebalancerID int, opts linodego.NodeBalancerUpdateOptions) (*linodego.NodeBalancer, error) {
	args := m.Called(ctx, nodebalancerID, opts)
	return args.Get(0).(*linodego.NodeBalancer), args.Error(1)
}

// DeleteNodeBalancer mocks deleting a NodeBalancer.
func (m *MockNodeBalancerService) DeleteNodeBalancer(ctx context.Context, nodebalancerID int) error {
	args := m.Called(ctx, nodebalancerID)
	return args.Error(0)
}

// NodeBalancer configuration methods
func (m *MockNodeBalancerService) ListNodeBalancerConfigs(ctx context.Context, nodebalancerID int, opts *linodego.ListOptions) ([]linodego.NodeBalancerConfig, error) {
	args := m.Called(ctx, nodebalancerID, opts)
	return args.Get(0).([]linodego.NodeBalancerConfig), args.Error(1)
}

func (m *MockNodeBalancerService) CreateNodeBalancerConfig(ctx context.Context, nodebalancerID int, opts linodego.NodeBalancerConfigCreateOptions) (*linodego.NodeBalancerConfig, error) {
	args := m.Called(ctx, nodebalancerID, opts)
	return args.Get(0).(*linodego.NodeBalancerConfig), args.Error(1)
}

func (m *MockNodeBalancerService) UpdateNodeBalancerConfig(ctx context.Context, nodebalancerID int, configID int, opts linodego.NodeBalancerConfigUpdateOptions) (*linodego.NodeBalancerConfig, error) {
	args := m.Called(ctx, nodebalancerID, configID, opts)
	return args.Get(0).(*linodego.NodeBalancerConfig), args.Error(1)
}

func (m *MockNodeBalancerService) DeleteNodeBalancerConfig(ctx context.Context, nodebalancerID int, configID int) error {
	args := m.Called(ctx, nodebalancerID, configID)
	return args.Error(0)
}

// Helper methods for setting up common test scenarios

func (m *MockNodeBalancerService) SetupListNodeBalancersSuccess(nodebalancers []linodego.NodeBalancer) {
	m.On("ListNodeBalancers", mock.Anything, mock.Anything).Return(nodebalancers, nil)
}

func (m *MockNodeBalancerService) SetupListNodeBalancersError(err error) {
	m.On("ListNodeBalancers", mock.Anything, mock.Anything).Return([]linodego.NodeBalancer{}, err)
}

func (m *MockNodeBalancerService) SetupGetNodeBalancerSuccess(nodebalancerID int, nodebalancer *linodego.NodeBalancer) {
	m.On("GetNodeBalancer", mock.Anything, nodebalancerID).Return(nodebalancer, nil)
}

func (m *MockNodeBalancerService) SetupGetNodeBalancerError(nodebalancerID int, err error) {
	m.On("GetNodeBalancer", mock.Anything, nodebalancerID).Return((*linodego.NodeBalancer)(nil), err)
}

func (m *MockNodeBalancerService) SetupCreateNodeBalancerSuccess(opts linodego.NodeBalancerCreateOptions, nodebalancer *linodego.NodeBalancer) {
	m.On("CreateNodeBalancer", mock.Anything, opts).Return(nodebalancer, nil)
}

func (m *MockNodeBalancerService) SetupCreateNodeBalancerError(opts linodego.NodeBalancerCreateOptions, err error) {
	m.On("CreateNodeBalancer", mock.Anything, opts).Return((*linodego.NodeBalancer)(nil), err)
}

func (m *MockNodeBalancerService) SetupListNodeBalancerConfigsSuccess(nodebalancerID int, configs []linodego.NodeBalancerConfig) {
	m.On("ListNodeBalancerConfigs", mock.Anything, nodebalancerID, mock.Anything).Return(configs, nil)
}

func (m *MockNodeBalancerService) SetupListNodeBalancerConfigsError(nodebalancerID int, err error) {
	m.On("ListNodeBalancerConfigs", mock.Anything, nodebalancerID, mock.Anything).Return([]linodego.NodeBalancerConfig{}, err)
}
