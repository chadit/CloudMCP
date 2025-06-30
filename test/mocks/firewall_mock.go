// Package mocks provides mock implementations for testing CloudMCP services.
//
//nolint:wrapcheck // Mock file returns test errors without wrapping
package mocks

import (
	"context"

	"github.com/linode/linodego"
	"github.com/stretchr/testify/mock"
)

// MockFirewallService provides a mock implementation of Linode firewall operations.
// This mock implements the subset of linodego.Client methods used by firewall tools.
type MockFirewallService struct {
	mock.Mock
}

// ListFirewalls mocks the listing of all firewalls.
func (m *MockFirewallService) ListFirewalls(ctx context.Context, opts *linodego.ListOptions) ([]linodego.Firewall, error) {
	args := m.Called(ctx, opts)

	firewalls, ok := args.Get(0).([]linodego.Firewall)
	if !ok {
		return nil, args.Error(1)
	}

	return firewalls, args.Error(1)
}

// GetFirewall mocks getting details of a specific firewall.
func (m *MockFirewallService) GetFirewall(ctx context.Context, firewallID int) (*linodego.Firewall, error) {
	args := m.Called(ctx, firewallID)

	firewall, ok := args.Get(0).(*linodego.Firewall)
	if !ok {
		return nil, args.Error(1)
	}

	return firewall, args.Error(1)
}

// CreateFirewall mocks creating a new firewall.
func (m *MockFirewallService) CreateFirewall(ctx context.Context, opts linodego.FirewallCreateOptions) (*linodego.Firewall, error) {
	args := m.Called(ctx, opts)

	firewall, ok := args.Get(0).(*linodego.Firewall)
	if !ok {
		return nil, args.Error(1)
	}

	return firewall, args.Error(1)
}

// UpdateFirewall mocks updating an existing firewall.
func (m *MockFirewallService) UpdateFirewall(ctx context.Context, firewallID int, opts linodego.FirewallUpdateOptions) (*linodego.Firewall, error) {
	args := m.Called(ctx, firewallID, opts)

	firewall, ok := args.Get(0).(*linodego.Firewall)
	if !ok {
		return nil, args.Error(1)
	}

	return firewall, args.Error(1)
}

// DeleteFirewall mocks deleting a firewall.
func (m *MockFirewallService) DeleteFirewall(ctx context.Context, firewallID int) error {
	args := m.Called(ctx, firewallID)

	return args.Error(0)
}

// UpdateFirewallRules mocks updating firewall rules.
func (m *MockFirewallService) UpdateFirewallRules(ctx context.Context, firewallID int, ruleSet linodego.FirewallRuleSet) (*linodego.FirewallRuleSet, error) {
	args := m.Called(ctx, firewallID, ruleSet)

	ruleSetPtr, ok := args.Get(0).(*linodego.FirewallRuleSet)
	if !ok {
		return nil, args.Error(1)
	}

	return ruleSetPtr, args.Error(1)
}

// CreateFirewallDevice mocks assigning a device to a firewall.
func (m *MockFirewallService) CreateFirewallDevice(ctx context.Context, firewallID int, opts linodego.FirewallDeviceCreateOptions) (*linodego.FirewallDevice, error) {
	args := m.Called(ctx, firewallID, opts)

	device, ok := args.Get(0).(*linodego.FirewallDevice)
	if !ok {
		return nil, args.Error(1)
	}

	return device, args.Error(1)
}

// DeleteFirewallDevice mocks removing a device from a firewall.
func (m *MockFirewallService) DeleteFirewallDevice(ctx context.Context, firewallID int, deviceID int) error {
	args := m.Called(ctx, firewallID, deviceID)

	return args.Error(0)
}

// Example test helper function.
func (m *MockFirewallService) SetupListFirewallsSuccess(firewalls []linodego.Firewall) {
	m.On("ListFirewalls", mock.Anything, mock.Anything).Return(firewalls, nil)
}

func (m *MockFirewallService) SetupListFirewallsError(err error) {
	m.On("ListFirewalls", mock.Anything, mock.Anything).Return([]linodego.Firewall{}, err)
}

func (m *MockFirewallService) SetupGetFirewallSuccess(firewallID int, firewall *linodego.Firewall) {
	m.On("GetFirewall", mock.Anything, firewallID).Return(firewall, nil)
}

func (m *MockFirewallService) SetupGetFirewallError(firewallID int, err error) {
	m.On("GetFirewall", mock.Anything, firewallID).Return((*linodego.Firewall)(nil), err)
}

func (m *MockFirewallService) SetupCreateFirewallSuccess(opts linodego.FirewallCreateOptions, firewall *linodego.Firewall) {
	m.On("CreateFirewall", mock.Anything, opts).Return(firewall, nil)
}

func (m *MockFirewallService) SetupCreateFirewallError(opts linodego.FirewallCreateOptions, err error) {
	m.On("CreateFirewall", mock.Anything, opts).Return((*linodego.Firewall)(nil), err)
}
