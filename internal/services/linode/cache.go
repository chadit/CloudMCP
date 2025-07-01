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

// Client defines the interface for Linode API operations used by the cache.
type Client interface {
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

// getCachedData provides generic caching logic for any slice type.
func getCachedData[T any](
	ctx context.Context,
	cache *Cache,
	getData func() []T,
	setData func([]T),
	fetchData func(context.Context, *linodego.ListOptions) ([]T, error),
) ([]T, error) {
	cache.mu.RLock()
	currentData := getData()

	if !cache.isExpired() && len(currentData) > 0 {
		result := make([]T, len(currentData))
		copy(result, currentData)
		cache.mu.RUnlock()

		return result, nil
	}
	cache.mu.RUnlock()

	// Cache is expired or empty, fetch fresh data
	cache.mu.Lock()
	defer cache.mu.Unlock()

	// Double-check pattern: another goroutine might have updated the cache
	currentData = getData()
	if !cache.isExpired() && len(currentData) > 0 {
		result := make([]T, len(currentData))
		copy(result, currentData)

		return result, nil
	}

	data, err := fetchData(ctx, nil)
	if err != nil {
		return nil, err
	}

	setData(data)

	cache.expiry = time.Now().Add(cache.ttl)

	// Return a copy to prevent external modifications
	result := make([]T, len(data))
	copy(result, data)

	return result, nil
}

func (c *Cache) getCachedRegions(ctx context.Context, client Client) ([]linodego.Region, error) {
	return getCachedData(
		ctx,
		c,
		func() []linodego.Region { return c.regions },
		func(data []linodego.Region) { c.regions = data },
		client.ListRegions,
	)
}

// getCachedTypes provides caching logic for types.
func (c *Cache) getCachedTypes(ctx context.Context, client Client) ([]linodego.LinodeType, error) {
	return getCachedData(
		ctx,
		c,
		func() []linodego.LinodeType { return c.types },
		func(data []linodego.LinodeType) { c.types = data },
		client.ListTypes,
	)
}

// getCachedKernels provides caching logic for kernels.
func (c *Cache) getCachedKernels(ctx context.Context, client Client) ([]linodego.LinodeKernel, error) {
	return getCachedData(
		ctx,
		c,
		func() []linodego.LinodeKernel { return c.kernels },
		func(data []linodego.LinodeKernel) { c.kernels = data },
		client.ListKernels,
	)
}

// GetRegions returns cached regions or fetches them from the API if cache is expired.
// This method is thread-safe and automatically refreshes stale data.
func (c *Cache) GetRegions(ctx context.Context, client Client) ([]linodego.Region, error) {
	return c.getCachedRegions(ctx, client)
}

// GetTypes returns cached Linode types or fetches them from the API if cache is expired.
// This method is thread-safe and automatically refreshes stale data.
func (c *Cache) GetTypes(ctx context.Context, client Client) ([]linodego.LinodeType, error) {
	return c.getCachedTypes(ctx, client)
}

// GetKernels returns cached kernels or fetches them from the API if cache is expired.
// This method is thread-safe and automatically refreshes stale data.
func (c *Cache) GetKernels(ctx context.Context, client Client) ([]linodego.LinodeKernel, error) {
	return c.getCachedKernels(ctx, client)
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
