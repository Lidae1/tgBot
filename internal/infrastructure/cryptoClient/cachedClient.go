package cryptoClient

import (
	"context"
	"log/slog"
	"tgBotFinal/internal/domain/service"
	"tgBotFinal/internal/entity"
	"tgBotFinal/internal/infrastructure/cache"
	"time"
)

type CachedClient struct {
	client service.CryptoClient
	cache  *cache.PriceCache
	logger *slog.Logger
	ttl    time.Duration
}

func NewCachedClient(client service.CryptoClient, ttl time.Duration, logger *slog.Logger) *CachedClient {
	return &CachedClient{
		client: client,
		cache:  cache.NewPriceCache(ttl),
		logger: logger.With(slog.String("component", "CachedClient")),
		ttl:    ttl,
	}
}

func (c *CachedClient) GetPriceBySymbol(ctx context.Context, symbol entity.CurrencyName) (*entity.Price, error) {
	return c.client.GetPriceBySymbol(ctx, symbol)
}

func (c *CachedClient) GetAllPrices(ctx context.Context) (*entity.PriceResponse, error) {
	if cached := c.cache.Get(); cached != nil {
		c.logger.Debug("Return price from cache",
			"cache_age", c.cache.GetCacheAge().Round(time.Second))
		return cached, nil
	}

	// Если кэш пустой или просроченный, и мы не обновляем уже - начинаем обновление
	if c.cache.StartRefresh() {
		defer c.cache.EndRefresh()

		c.logger.Debug("Cache miss, fetching prices from API")
		prices, err := c.client.GetAllPrices(ctx)
		if err != nil {
			c.logger.Error("Failed to fetch prices from API", "error", err)

			// Если есть старые данные в кэше, вернем их даже если просрочены
			if stale := c.cache.Get(); stale != nil {
				c.logger.Warn("Returning stale cache data due to API error")

				return stale, nil
			}

			return nil, err
		}

		c.cache.Set(prices)
		c.logger.Debug("Prices cached successfully",
			"ttl", c.ttl,
			"BTC", prices.BTC != nil,
			"ETH", prices.ETH != nil)

		return prices, nil
	}

	// Если уже обновляется, ждем и возвращаем то что есть (даже просроченное)
	c.logger.Debug("No cache available, waiting for refresh")
	return c.client.GetAllPrices(ctx)
}
