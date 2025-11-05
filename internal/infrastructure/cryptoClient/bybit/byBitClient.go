package bybit

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"tgBotFinal/internal/domain/service"
	"time"

	"tgBotFinal/internal/entity"

	"golang.org/x/sync/errgroup"
)

type Client struct {
	httpClient *http.Client
	logger     *slog.Logger
	baseURL    string
}

type byBitResponse struct {
	RetCode int    `json:"retCode"`
	RetMsg  string `json:"retMsg"`
	Result  struct {
		List []TickerInfo `json:"list"`
	} `json:"result"`
}

type TickerInfo struct {
	Symbol    string `json:"symbol"`
	LastPrice string `json:"lastPrice"`
}

func NewClient(logger *slog.Logger, api string) service.CryptoClient {
	client := &Client{
		httpClient: &http.Client{Timeout: 10 * time.Second},
		logger:     logger.With(slog.String("component", "byBitClient")),
		baseURL:    api,
	}

	return client
}

func (c *Client) GetPriceBySymbol(ctx context.Context, symbol entity.CurrencyName) (*entity.Price, error) {
	c.logger.Debug("Get price by symbol", "symbol", symbol)

	url := fmt.Sprintf("%s?category=spot&symbol=%sUSDT", c.baseURL, symbol)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		c.logger.Error("error creating request", "err", err)
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error("error executing request", "err", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.logger.Error("error executing request", "err", resp.Status)
		return nil, fmt.Errorf("bad status: %s", resp.Status)
	}

	var result byBitResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		c.logger.Error("error parsing response", "err", err)
		return nil, err
	}

	if result.RetCode != 0 {
		c.logger.Error("error executing request", "err", result.RetMsg)
		return nil, fmt.Errorf("bybit API error: %s (code: %d)", result.RetMsg, result.RetCode)
	}

	if len(result.Result.List) == 0 {
		c.logger.Error("error executing request", "err", result.RetMsg)
		return nil, fmt.Errorf("no data for symbol: %s", symbol)
	}

	c.logger.Debug("got price by symbol", "symbol", symbol)
	return &entity.Price{
		Symbol:  symbol,
		Price:   result.Result.List[0].LastPrice,
		Updated: time.Now().Format("2006-01-02 15:04:05"),
	}, nil

}
func (c *Client) GetAllPrices(ctx context.Context) (*entity.PriceResponse, error) {
	c.logger.Debug("Get all prices")

	g, ctx := errgroup.WithContext(ctx)

	var mu sync.Mutex

	prices := make(map[entity.CurrencyName]*entity.Price)

	for _, symbol := range entity.TokenList {
		g.Go(func() error {
			res, err := c.GetPriceBySymbol(ctx, symbol)
			if err != nil {
				c.logger.Error("error getting price by symbol", "symbol", symbol, "err", err)
				return nil
			}

			mu.Lock()
			prices[symbol] = res
			mu.Unlock()

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	result := &entity.PriceResponse{}
	for symbol, price := range prices {
		switch symbol {
		case entity.BTC:
			result.BTC = price
		case entity.ETH:
			result.ETH = price
		}
	}

	return result, nil
}
