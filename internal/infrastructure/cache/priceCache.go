package cache

import (
	"sync"
	"tgBotFinal/internal/entity"
	"time"
)

type PriceCache struct {
	mu           sync.RWMutex
	prices       *entity.PriceResponse
	cacheAt      time.Time
	ttl          time.Duration
	isRefreshing bool
	refreshMux   sync.Mutex
}

func NewPriceCache(ttl time.Duration) *PriceCache {
	return &PriceCache{
		ttl:    ttl,
		prices: &entity.PriceResponse{},
	}
}

func (c *PriceCache) Get() *entity.PriceResponse {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.isExpired() {
		return nil
	}

	return c.prices
}

func (c *PriceCache) Set(prices *entity.PriceResponse) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.prices = prices
	c.cacheAt = time.Now()
}

func (c *PriceCache) IsExpired() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.isExpired()
}

func (c *PriceCache) isExpired() bool {
	return time.Since(c.cacheAt) > c.ttl || c.prices == nil
}

func (c *PriceCache) StartRefresh() bool {
	c.refreshMux.Lock()
	defer c.refreshMux.Unlock()

	if c.isRefreshing {
		return false
	}

	c.isRefreshing = true
	return true
}

func (c *PriceCache) EndRefresh() {
	c.refreshMux.Lock()
	defer c.refreshMux.Unlock()
	c.isRefreshing = false
}

func (c *PriceCache) GetCacheAge() time.Duration {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.cacheAt.IsZero() {
		return 0
	}

	return time.Since(c.cacheAt)
}
