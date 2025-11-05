package bybit

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"tgBotFinal/internal/entity"
)

func TestByBitClient_GetPriceBySymbol(t *testing.T) {

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
				"retCode": 0,
				"retMsg": "OK",
				"result": {
						"list":[
								{
										"symbol": "BTCUSDT",
										"lastPrice": "50000.50"
								}
						]
				}
		}`))
	}))
	defer server.Close()

	client := &Client{
		httpClient: server.Client(),
		logger:     slog.Default(),
		baseURL:    server.URL,
	}

	price, err := client.GetPriceBySymbol(context.Background(), entity.BTC)
	if err != nil {
		t.Fatalf("GetPriceBySymbol failed: %v", err)
	}

	if price.Symbol != entity.BTC {
		t.Errorf("Symbol = %v, want %v", price.Symbol, entity.BTC)
	}

	if price.Price != "50000.50" {
		t.Errorf("Price = %v, want %v", price.Price, "50000.50")
	}
}

func TestByBitClient_GetPriceBySymbol_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := &Client{
		httpClient: server.Client(),
		logger:     slog.Default(),
		baseURL:    server.URL,
	}

	_, err := client.GetPriceBySymbol(context.Background(), entity.BTC)
	if err == nil {
		t.Errorf("Expected error for failed request")
	}
}

func TestByBitClient_GetPriceBySymbol_NoData(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
				"retCode": 0,
				"retMsg": "OK",
				"result": {
						"list": []
				}
		}`))
	}))
	defer server.Close()

	client := &Client{
		httpClient: server.Client(),
		logger:     slog.Default(),
		baseURL:    server.URL,
	}

	_, err := client.GetPriceBySymbol(context.Background(), entity.BTC)
	if err == nil {
		t.Errorf("Expected error for no data")
	}
}
