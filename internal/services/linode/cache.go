package linode

import (
	"context"
	"sync"
	"time"

	"github.com/linode/linodego"
)

const (
	// Default cache TTL in minutes.
	defaultCacheTTLMinutes = 30
)

// LinodeClient defines the interface for Linode API operations used by the cache.
type LinodeClient interface {
	ListRegions(ctx context.Context, opts *linodego.ListOptions) ([]linodego.Region, error)
	ListTypes(ctx context.Context, opts *linodego.ListOptions) ([]linodego.LinodeType, error)
	ListKernels(ctx context.Context, opts *linodego.ListOptions) ([]linodego.LinodeKernel, error)
}

// Cache provides thread-safe caching for static Linode data with TTL expiration.
// It caches frequently accessed, rarely changing data like regions, types, and kernels
// to reduce API calls and improve performance.
type Cache struct {
	regions []linodego.Region
	types   []linodego.LinodeType
	kernels []linodego.LinodeKernel
	mu      sync.RWMutex
	expiry  time.Time
	ttl     time.Duration
}

// CacheConfig contains configuration for the cache behavior.
type CacheConfig struct {
	TTL time.Duration // Cache time-to-live duration
}

// NewCache creates a new cache instance with the specified configuration.
func NewCache(config CacheConfig) *Cache {
	if config.TTL == 0 {
		config.TTL = defaultCacheTTLMinutes * time.Minute // Default TTL
	}

	return &Cache{
		ttl: config.TTL,
	}
}

// isExpired checks if the cache has expired based on the TTL.
func (c *Cache) isExpired() bool {
	return time.Now().After(c.expiry)
}

// GetRegions returns cached regions or fetches them from the API if cache is expired.
// This method is thread-safe and automatically refreshes stale data.
func (c *Cache) GetRegions(ctx context.Context, client LinodeClient) ([]linodego.Region, error) { //nolint:dupl // Similar caching pattern needed for type safety
	c.mu.RLock()
	if !c.isExpired() && len(c.regions) > 0 {
		regions := make([]linodego.Region, len(c.regions))
		copy(regions, c.regions)
		c.mu.RUnlock()

		return regions, nil
	}
	c.mu.RUnlock()

	// Cache is expired or empty, fetch fresh data
	c.mu.Lock()
	defer c.mu.Unlock()

	// Double-check pattern: another goroutine might have updated the cache
	if !c.isExpired() && len(c.regions) > 0 {
		regions := make([]linodego.Region, len(c.regions))
		copy(regions, c.regions)

		return regions, nil
	}

	regions, err := client.ListRegions(ctx, nil)
	if err != nil {
		return nil, err
	}

	c.regions = regions
	c.expiry = time.Now().Add(c.ttl)

	// Return a copy to prevent external modifications
	result := make([]linodego.Region, len(regions))
	copy(result, regions)

	return result, nil
}

// GetTypes returns cached Linode types or fetches them from the API if cache is expired.
// This method is thread-safe and automatically refreshes stale data.
func (c *Cache) GetTypes(ctx context.Context, client LinodeClient) ([]linodego.LinodeType, error) { //nolint:dupl // Similar caching pattern needed for type safety
	c.mu.RLock()
	if !c.isExpired() && len(c.types) > 0 {
		types := make([]linodego.LinodeType, len(c.types))
		copy(types, c.types)
		c.mu.RUnlock()

		return types, nil
	}
	c.mu.RUnlock()

	// Cache is expired or empty, fetch fresh data
	c.mu.Lock()
	defer c.mu.Unlock()

	// Double-check pattern: another goroutine might have updated the cache
	if !c.isExpired() && len(c.types) > 0 {
		types := make([]linodego.LinodeType, len(c.types))
		copy(types, c.types)

		return types, nil
	}

	types, err := client.ListTypes(ctx, nil)
	if err != nil {
		return nil, err
	}

	c.types = types
	c.expiry = time.Now().Add(c.ttl)

	// Return a copy to prevent external modifications
	result := make([]linodego.LinodeType, len(types))
	copy(result, types)

	return result, nil
}

// GetKernels returns cached kernels or fetches them from the API if cache is expired.
// This method is thread-safe and automatically refreshes stale data.
func (c *Cache) GetKernels(ctx context.Context, client LinodeClient) ([]linodego.LinodeKernel, error) {
	c.mu.RLock()
	if !c.isExpired() && len(c.kernels) > 0 {
		kernels := make([]linodego.LinodeKernel, len(c.kernels))
		copy(kernels, c.kernels)
		c.mu.RUnlock()

		return kernels, nil
	}
	c.mu.RUnlock()

	// Cache is expired or empty, fetch fresh data
	c.mu.Lock()
	defer c.mu.Unlock()

	// Double-check pattern: another goroutine might have updated the cache
	if !c.isExpired() && len(c.kernels) > 0 {
		kernels := make([]linodego.LinodeKernel, len(c.kernels))
		copy(kernels, c.kernels)

		return kernels, nil
	}

	kernels, err := client.ListKernels(ctx, nil)
	if err != nil {
		return nil, err
	}

	c.kernels = kernels
	c.expiry = time.Now().Add(c.ttl)

	// Return a copy to prevent external modifications
	result := make([]linodego.LinodeKernel, len(kernels))
	copy(result, kernels)

	return result, nil
}

// InvalidateRegions forcefully removes cached region data, causing the next
// GetRegions call to fetch fresh data from the API.
func (c *Cache) InvalidateRegions() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.regions = nil
}

// InvalidateTypes forcefully removes cached type data, causing the next
// GetTypes call to fetch fresh data from the API.
func (c *Cache) InvalidateTypes() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.types = nil
}

// InvalidateKernels forcefully removes cached kernel data, causing the next
// GetKernels call to fetch fresh data from the API.
func (c *Cache) InvalidateKernels() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.kernels = nil
}

// InvalidateAll clears all cached data, forcing fresh API calls on the next access.
// This is useful for testing or when you need to ensure completely fresh data.
func (c *Cache) InvalidateAll() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.regions = nil
	c.types = nil
	c.kernels = nil
	c.expiry = time.Time{} // Reset expiry to force refresh
}

// Stats returns cache statistics for monitoring and debugging.
type CacheStats struct {
	RegionsCached bool      `json:"regionsCached"`
	TypesCached   bool      `json:"typesCached"`
	KernelsCached bool      `json:"kernelsCached"`
	CacheExpiry   time.Time `json:"cacheExpiry"`
	IsExpired     bool      `json:"isExpired"`
	TTL           string    `json:"ttl"`
	RegionsCount  int       `json:"regionsCount"`
	TypesCount    int       `json:"typesCount"`
	KernelsCount  int       `json:"kernelsCount"`
}

// GetStats returns current cache statistics for monitoring purposes.
func (c *Cache) GetStats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return CacheStats{
		RegionsCached: len(c.regions) > 0,
		TypesCached:   len(c.types) > 0,
		KernelsCached: len(c.kernels) > 0,
		CacheExpiry:   c.expiry,
		IsExpired:     c.isExpired(),
		TTL:           c.ttl.String(),
		RegionsCount:  len(c.regions),
		TypesCount:    len(c.types),
		KernelsCount:  len(c.kernels),
	}
}
