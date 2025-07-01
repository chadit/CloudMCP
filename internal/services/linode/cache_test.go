package linode_test

import (
	"sync"
	"testing"
	"time"

	"github.com/linode/linodego"
	"github.com/stretchr/testify/require"

	"github.com/chadit/CloudMCP/internal/services/linode"
	"github.com/chadit/CloudMCP/test/mocks"
)

func TestNewCache_DefaultTTL(t *testing.T) {
	t.Parallel()

	cache := linode.NewCache(linode.CacheConfig{})
	require.NotNil(t, cache, "Cache should not be nil")
}

func TestNewCache_CustomTTL(t *testing.T) {
	t.Parallel()

	customTTL := 10 * time.Minute
	cache := linode.NewCache(linode.CacheConfig{TTL: customTTL})
	require.NotNil(t, cache, "Cache should not be nil")
}

func TestCache_GetRegions_FreshFetch(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	cache := linode.NewCache(linode.CacheConfig{TTL: 5 * time.Minute})

	// Create mock client
	mockClient := &mocks.MockClient{}
	expectedRegions := []linodego.Region{
		{ID: "us-east", Label: "Newark, NJ", Country: "us"},
		{ID: "us-west", Label: "Fremont, CA", Country: "us"},
	}

	mockClient.On("ListRegions", ctx, (*linodego.ListOptions)(nil)).Return(expectedRegions, nil)

	// Test fresh fetch
	regions, err := cache.GetRegions(ctx, mockClient)
	require.NoError(t, err, "GetRegions should not error")
	require.Equal(t, expectedRegions, regions, "Regions should match expected")

	// Verify cache was populated
	stats := cache.GetStats()
	require.True(t, stats.RegionsCached, "Regions should be cached")
	require.Equal(t, 2, stats.RegionsCount, "Cache should contain 2 regions")
	require.False(t, stats.IsExpired, "Cache should not be expired")

	mockClient.AssertExpectations(t)
}

func TestCache_GetRegions_CacheHit(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	cache := linode.NewCache(linode.CacheConfig{TTL: 5 * time.Minute})

	mockClient := &mocks.MockClient{}
	expectedRegions := []linodego.Region{
		{ID: "us-east", Label: "Newark, NJ", Country: "us"},
	}

	// First call should hit the API
	mockClient.On("ListRegions", ctx, (*linodego.ListOptions)(nil)).Return(expectedRegions, nil).Once()

	// First call
	regions1, err := cache.GetRegions(ctx, mockClient)
	require.NoError(t, err, "First GetRegions should not error")
	require.Equal(t, expectedRegions, regions1, "First call should return expected regions")

	// Second call should hit cache (no additional API call)
	regions2, err := cache.GetRegions(ctx, mockClient)
	require.NoError(t, err, "Second GetRegions should not error")
	require.Equal(t, expectedRegions, regions2, "Second call should return cached regions")

	mockClient.AssertExpectations(t)
}

func TestCache_GetRegions_CacheExpiry(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	cache := linode.NewCache(linode.CacheConfig{TTL: 100 * time.Millisecond}) // Very short TTL

	mockClient := &mocks.MockClient{}
	expectedRegions := []linodego.Region{
		{ID: "us-east", Label: "Newark, NJ", Country: "us"},
	}

	// First call
	mockClient.On("ListRegions", ctx, (*linodego.ListOptions)(nil)).Return(expectedRegions, nil).Once()
	regions1, err := cache.GetRegions(ctx, mockClient)
	require.NoError(t, err, "First GetRegions should not error")
	require.Equal(t, expectedRegions, regions1, "First call should return expected regions")

	// Wait for cache to expire
	time.Sleep(150 * time.Millisecond)

	// Second call should hit API again due to expiry
	mockClient.On("ListRegions", ctx, (*linodego.ListOptions)(nil)).Return(expectedRegions, nil).Once()
	regions2, err := cache.GetRegions(ctx, mockClient)
	require.NoError(t, err, "Second GetRegions should not error")
	require.Equal(t, expectedRegions, regions2, "Second call should return fresh regions")

	mockClient.AssertExpectations(t)
}

func TestCache_GetTypes_FreshFetch(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	cache := linode.NewCache(linode.CacheConfig{TTL: 5 * time.Minute})

	mockClient := &mocks.MockClient{}
	expectedTypes := []linodego.LinodeType{
		{ID: "g6-nanode-1", Label: "Nanode 1GB", Memory: 1024},
		{ID: "g6-standard-1", Label: "Linode 2GB", Memory: 2048},
	}

	mockClient.On("ListTypes", ctx, (*linodego.ListOptions)(nil)).Return(expectedTypes, nil)

	types, err := cache.GetTypes(ctx, mockClient)
	require.NoError(t, err, "GetTypes should not error")
	require.Equal(t, expectedTypes, types, "Types should match expected")

	stats := cache.GetStats()
	require.True(t, stats.TypesCached, "Types should be cached")
	require.Equal(t, 2, stats.TypesCount, "Cache should contain 2 types")

	mockClient.AssertExpectations(t)
}

func TestCache_GetKernels_FreshFetch(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	cache := linode.NewCache(linode.CacheConfig{TTL: 5 * time.Minute})

	mockClient := &mocks.MockClient{}
	expectedKernels := []linodego.LinodeKernel{
		{ID: "linode/latest-64bit", Label: "Latest 64 bit"},
		{ID: "linode/grub2", Label: "GRUB 2"},
	}

	mockClient.On("ListKernels", ctx, (*linodego.ListOptions)(nil)).Return(expectedKernels, nil)

	kernels, err := cache.GetKernels(ctx, mockClient)
	require.NoError(t, err, "GetKernels should not error")
	require.Equal(t, expectedKernels, kernels, "Kernels should match expected")

	stats := cache.GetStats()
	require.True(t, stats.KernelsCached, "Kernels should be cached")
	require.Equal(t, 2, stats.KernelsCount, "Cache should contain 2 kernels")

	mockClient.AssertExpectations(t)
}

func TestCache_InvalidateRegions(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	cache := linode.NewCache(linode.CacheConfig{TTL: 5 * time.Minute})

	mockClient := &mocks.MockClient{}
	expectedRegions := []linodego.Region{
		{ID: "us-east", Label: "Newark, NJ", Country: "us"},
	}

	// Populate cache
	mockClient.On("ListRegions", ctx, (*linodego.ListOptions)(nil)).Return(expectedRegions, nil).Twice()

	_, err := cache.GetRegions(ctx, mockClient)
	require.NoError(t, err, "GetRegions should not error")

	// Verify cache is populated
	stats := cache.GetStats()
	require.True(t, stats.RegionsCached, "Regions should be cached")

	// Invalidate and verify cache is cleared
	cache.InvalidateRegions()
	stats = cache.GetStats()
	require.False(t, stats.RegionsCached, "Regions should not be cached after invalidation")

	// Next call should fetch fresh data
	_, err = cache.GetRegions(ctx, mockClient)
	require.NoError(t, err, "GetRegions should not error after invalidation")

	mockClient.AssertExpectations(t)
}

func TestCache_InvalidateAll(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	cache := linode.NewCache(linode.CacheConfig{TTL: 5 * time.Minute})

	mockClient := &mocks.MockClient{}
	expectedRegions := []linodego.Region{{ID: "us-east"}}
	expectedTypes := []linodego.LinodeType{{ID: "g6-nanode-1"}}
	expectedKernels := []linodego.LinodeKernel{{ID: "linode/latest-64bit"}}

	// Populate all caches
	mockClient.On("ListRegions", ctx, (*linodego.ListOptions)(nil)).Return(expectedRegions, nil)
	mockClient.On("ListTypes", ctx, (*linodego.ListOptions)(nil)).Return(expectedTypes, nil)
	mockClient.On("ListKernels", ctx, (*linodego.ListOptions)(nil)).Return(expectedKernels, nil)

	_, err := cache.GetRegions(ctx, mockClient)
	require.NoError(t, err, "GetRegions should not error")
	_, err = cache.GetTypes(ctx, mockClient)
	require.NoError(t, err, "GetTypes should not error")
	_, err = cache.GetKernels(ctx, mockClient)
	require.NoError(t, err, "GetKernels should not error")

	// Verify all caches are populated
	stats := cache.GetStats()
	require.True(t, stats.RegionsCached, "Regions should be cached")
	require.True(t, stats.TypesCached, "Types should be cached")
	require.True(t, stats.KernelsCached, "Kernels should be cached")

	// Invalidate all
	cache.InvalidateAll()

	// Verify all caches are cleared
	stats = cache.GetStats()
	require.False(t, stats.RegionsCached, "Regions should not be cached after invalidation")
	require.False(t, stats.TypesCached, "Types should not be cached after invalidation")
	require.False(t, stats.KernelsCached, "Kernels should not be cached after invalidation")
	require.True(t, stats.IsExpired, "Cache should be marked as expired")

	mockClient.AssertExpectations(t)
}

func TestCache_ConcurrentAccess(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	cache := linode.NewCache(linode.CacheConfig{TTL: 5 * time.Minute})

	mockClient := &mocks.MockClient{}
	expectedRegions := []linodego.Region{
		{ID: "us-east", Label: "Newark, NJ", Country: "us"},
	}

	// Mock should be called only once due to caching
	mockClient.On("ListRegions", ctx, (*linodego.ListOptions)(nil)).Return(expectedRegions, nil).Once()

	// Launch multiple concurrent requests
	const numGoroutines = 10

	var waitGroup sync.WaitGroup

	results := make([][]linodego.Region, numGoroutines)
	errors := make([]error, numGoroutines)

	waitGroup.Add(numGoroutines)

	for goroutineIndex := range numGoroutines {
		go func(index int) {
			defer waitGroup.Done()

			regions, err := cache.GetRegions(ctx, mockClient)
			results[index] = regions
			errors[index] = err
		}(goroutineIndex)
	}

	waitGroup.Wait()

	// Verify all calls succeeded and returned same data
	for goroutineIndex := range numGoroutines {
		require.NoError(t, errors[goroutineIndex], "Concurrent call %d should not error", goroutineIndex)
		require.Equal(t, expectedRegions, results[goroutineIndex], "Concurrent call %d should return expected regions", goroutineIndex)
	}

	mockClient.AssertExpectations(t)
}

func TestCache_GetStats(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	cache := linode.NewCache(linode.CacheConfig{TTL: 5 * time.Minute})

	// Initial stats should show empty cache
	stats := cache.GetStats()
	require.False(t, stats.RegionsCached, "Regions should not be cached initially")
	require.False(t, stats.TypesCached, "Types should not be cached initially")
	require.False(t, stats.KernelsCached, "Kernels should not be cached initially")
	require.Equal(t, 0, stats.RegionsCount, "Regions count should be 0")
	require.Equal(t, 0, stats.TypesCount, "Types count should be 0")
	require.Equal(t, 0, stats.KernelsCount, "Kernels count should be 0")
	require.Equal(t, "5m0s", stats.TTL, "TTL should match configured value")

	// Populate cache
	mockClient := &mocks.MockClient{}
	expectedRegions := []linodego.Region{{ID: "us-east"}, {ID: "us-west"}}

	mockClient.On("ListRegions", ctx, (*linodego.ListOptions)(nil)).Return(expectedRegions, nil)

	_, err := cache.GetRegions(ctx, mockClient)
	require.NoError(t, err, "GetRegions should not error")

	// Check updated stats
	stats = cache.GetStats()
	require.True(t, stats.RegionsCached, "Regions should be cached after fetch")
	require.Equal(t, 2, stats.RegionsCount, "Regions count should be 2")
	require.False(t, stats.IsExpired, "Cache should not be expired")
	require.False(t, stats.CacheExpiry.IsZero(), "Cache expiry should be set")

	mockClient.AssertExpectations(t)
}

func TestCache_DataImmutability(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	cache := linode.NewCache(linode.CacheConfig{TTL: 5 * time.Minute})

	mockClient := &mocks.MockClient{}
	originalRegions := []linodego.Region{
		{ID: "us-east", Label: "Newark, NJ", Country: "us"},
	}

	mockClient.On("ListRegions", ctx, (*linodego.ListOptions)(nil)).Return(originalRegions, nil)

	// Get regions from cache
	regions1, err := cache.GetRegions(ctx, mockClient)
	require.NoError(t, err, "GetRegions should not error")

	// Modify returned slice
	regions1[0].Label = "Modified Label"

	// Get regions again from cache
	regions2, err := cache.GetRegions(ctx, mockClient)
	require.NoError(t, err, "Second GetRegions should not error")

	// Verify cache data was not affected by external modification
	require.Equal(t, "Newark, NJ", regions2[0].Label, "Cache data should not be modified by external changes")
	require.NotEqual(t, regions1[0].Label, regions2[0].Label, "Returned data should be independent copies")

	mockClient.AssertExpectations(t)
}
