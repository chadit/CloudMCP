package mocks

import (
	"context"

	"github.com/linode/linodego"
	"github.com/stretchr/testify/mock"
)

// MockLKEService provides a mock implementation of Linode Kubernetes Engine (LKE) operations.
// This mock implements the subset of linodego.Client methods used by LKE tools.
type MockLKEService struct {
	mock.Mock
}

// LKE Cluster methods
func (m *MockLKEService) ListLKEClusters(ctx context.Context, opts *linodego.ListOptions) ([]linodego.LKECluster, error) {
	args := m.Called(ctx, opts)
	return args.Get(0).([]linodego.LKECluster), args.Error(1)
}

func (m *MockLKEService) GetLKECluster(ctx context.Context, clusterID int) (*linodego.LKECluster, error) {
	args := m.Called(ctx, clusterID)
	return args.Get(0).(*linodego.LKECluster), args.Error(1)
}

func (m *MockLKEService) CreateLKECluster(ctx context.Context, opts linodego.LKEClusterCreateOptions) (*linodego.LKECluster, error) {
	args := m.Called(ctx, opts)
	return args.Get(0).(*linodego.LKECluster), args.Error(1)
}

func (m *MockLKEService) UpdateLKECluster(ctx context.Context, clusterID int, opts linodego.LKEClusterUpdateOptions) (*linodego.LKECluster, error) {
	args := m.Called(ctx, clusterID, opts)
	return args.Get(0).(*linodego.LKECluster), args.Error(1)
}

func (m *MockLKEService) DeleteLKECluster(ctx context.Context, clusterID int) error {
	args := m.Called(ctx, clusterID)
	return args.Error(0)
}

// LKE Node Pool methods
func (m *MockLKEService) ListLKEClusterPools(ctx context.Context, clusterID int, opts *linodego.ListOptions) ([]linodego.LKEClusterPool, error) {
	args := m.Called(ctx, clusterID, opts)
	return args.Get(0).([]linodego.LKEClusterPool), args.Error(1)
}

func (m *MockLKEService) CreateLKEClusterPool(ctx context.Context, clusterID int, opts linodego.LKEClusterPoolCreateOptions) (*linodego.LKEClusterPool, error) {
	args := m.Called(ctx, clusterID, opts)
	return args.Get(0).(*linodego.LKEClusterPool), args.Error(1)
}

func (m *MockLKEService) UpdateLKEClusterPool(ctx context.Context, clusterID int, poolID int, opts linodego.LKEClusterPoolUpdateOptions) (*linodego.LKEClusterPool, error) {
	args := m.Called(ctx, clusterID, poolID, opts)
	return args.Get(0).(*linodego.LKEClusterPool), args.Error(1)
}

func (m *MockLKEService) DeleteLKEClusterPool(ctx context.Context, clusterID int, poolID int) error {
	args := m.Called(ctx, clusterID, poolID)
	return args.Error(0)
}

// Kubeconfig method
func (m *MockLKEService) GetLKEClusterKubeconfig(ctx context.Context, clusterID int) (*linodego.LKEClusterKubeconfig, error) {
	args := m.Called(ctx, clusterID)
	return args.Get(0).(*linodego.LKEClusterKubeconfig), args.Error(1)
}

// Helper methods for setting up common test scenarios

func (m *MockLKEService) SetupListLKEClustersSuccess(clusters []linodego.LKECluster) {
	m.On("ListLKEClusters", mock.Anything, mock.Anything).Return(clusters, nil)
}

func (m *MockLKEService) SetupListLKEClustersError(err error) {
	m.On("ListLKEClusters", mock.Anything, mock.Anything).Return([]linodego.LKECluster{}, err)
}

func (m *MockLKEService) SetupGetLKEClusterSuccess(clusterID int, cluster *linodego.LKECluster) {
	m.On("GetLKECluster", mock.Anything, clusterID).Return(cluster, nil)
}

func (m *MockLKEService) SetupGetLKEClusterError(clusterID int, err error) {
	m.On("GetLKECluster", mock.Anything, clusterID).Return((*linodego.LKECluster)(nil), err)
}

func (m *MockLKEService) SetupCreateLKEClusterSuccess(opts linodego.LKEClusterCreateOptions, cluster *linodego.LKECluster) {
	m.On("CreateLKECluster", mock.Anything, opts).Return(cluster, nil)
}

func (m *MockLKEService) SetupCreateLKEClusterError(opts linodego.LKEClusterCreateOptions, err error) {
	m.On("CreateLKECluster", mock.Anything, opts).Return((*linodego.LKECluster)(nil), err)
}

func (m *MockLKEService) SetupListLKEClusterPoolsSuccess(clusterID int, pools []linodego.LKEClusterPool) {
	m.On("ListLKEClusterPools", mock.Anything, clusterID, mock.Anything).Return(pools, nil)
}

func (m *MockLKEService) SetupListLKEClusterPoolsError(clusterID int, err error) {
	m.On("ListLKEClusterPools", mock.Anything, clusterID, mock.Anything).Return([]linodego.LKEClusterPool{}, err)
}

func (m *MockLKEService) SetupGetLKEClusterKubeconfigSuccess(clusterID int, kubeconfig *linodego.LKEClusterKubeconfig) {
	m.On("GetLKEClusterKubeconfig", mock.Anything, clusterID).Return(kubeconfig, nil)
}

func (m *MockLKEService) SetupGetLKEClusterKubeconfigError(clusterID int, err error) {
	m.On("GetLKEClusterKubeconfig", mock.Anything, clusterID).Return((*linodego.LKEClusterKubeconfig)(nil), err)
}
