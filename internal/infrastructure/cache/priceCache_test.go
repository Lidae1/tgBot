package cache

import (
	"testing"
	"tgBotFinal/internal/entity"
	"time"
)

func TestPriceCache(t *testing.T) {
	cache := NewPriceCache(1 * time.Minute)

	if cached := cache.Get(); cached != nil {
		t.Errorf("Empty cache should return nil")
	}

	prices := &entity.PriceResponse{
		BTC: &entity.Price{
			Symbol:  entity.BTC,
			Price:   "50000.00",
			Updated: time.Now().Format("2006-01-02 15:04:05"),
		},
	}

	cache.Set(prices)
	cached := cache.Get()
	if cached == nil {
		t.Errorf("Cache should return prices after setting")
	}

	if cached.BTC.Price != "50000.00" {
		t.Errorf("Cached price = %v, want %v", cached.BTC.Price, "50000.00")
	}
}

func TestPriceCacheExpiration(t *testing.T) {
	cache := NewPriceCache(1 * time.Millisecond)

	prices := &entity.PriceResponse{
		BTC: &entity.Price{
			Symbol: entity.BTC,
			Price:  "50000.00",
		},
	}

	cache.Set(prices)

	if cache.Get() == nil {
		t.Errorf("Prices should be available before TTL")
	}

	time.Sleep(150 * time.Millisecond)

	if cache.Get() != nil {
		t.Errorf("Prices should be expired after TTL")
	}
}

func TestPriceCacheConccurrentAccess(t *testing.T) {
	cache := NewPriceCache(1 * time.Minute)
	done := make(chan bool)

	for i := 0; i < 10; i++ {
		go func() {
			prices := &entity.PriceResponse{
				BTC: &entity.Price{
					Symbol: entity.BTC,
					Price:  string(rune(50000 + i)),
				},
			}

			cache.Set(prices)
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		go func() {
			cache.Get()
			done <- true
		}()
	}

	for i := 0; i < 20; i++ {
		<-done
	}

	if cache.Get() == nil {
		t.Errorf("Cache should have a value after concurrent access")
	}
}
