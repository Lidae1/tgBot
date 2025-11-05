package service

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"tgBotFinal/internal/entity"
)

func TestCryptService_SendNotifications(t *testing.T) {
	mockCrypto := NewMockCryptoClient()
	mockNotifier := NewMockNotification()
	mockUserRepo := NewMockUserRepository()
	mockCurrencyRepo := NewMockCurrencyRepository()

	var sentPrices *entity.PriceResponse

	mockNotifier.SendAllPricesFunc = func(ctx context.Context, chatID int64, prices *entity.PriceResponse) error {
		sentPrices = prices
		return nil
	}

	service := &CryptService{
		CurrencyRepo: mockCurrencyRepo,
		UserRepo:     mockUserRepo,
		CryptClient:  mockCrypto,
		Notification: mockNotifier,
		ChiRouter:    nil,
		logger:       slog.Default(),
		webhookURL:   "",
		port:         "",
	}

	err := service.sendNotificationsToActive(context.Background())
	if err != nil {
		t.Fatalf("sendNotificationsToActive failed: %v", err)
	}

	if sentPrices == nil {
		t.Errorf("Prices should be sent to notification")
	}
}

func TestCryptService_SendNotifications_Error(t *testing.T) {
	mockCrypto := NewMockCryptoClient()
	mockNotifier := NewMockNotification()
	mockUserRepo := NewMockUserRepository()
	mockCurrencyRepo := NewMockCurrencyRepository()

	mockCrypto.GetAllPricesFunc = func(ctx context.Context) (*entity.PriceResponse, error) {
		return nil, errors.New("API error")
	}

	service := &CryptService{
		CurrencyRepo: mockCurrencyRepo,
		UserRepo:     mockUserRepo,
		CryptClient:  mockCrypto,
		Notification: mockNotifier,
		ChiRouter:    nil,
		logger:       slog.Default(),
		webhookURL:   "",
		port:         "",
	}

	err := service.sendNotificationsToActive(context.Background())
	if err == nil {
		t.Errorf("Expected error when API fails")
	}
}

func TestCryptService_GetPricesWithRetry(t *testing.T) {
	callCount := 0
	mockCrypto := &MockCryptoClient{
		GetAllPricesFunc: func(ctx context.Context) (*entity.PriceResponse, error) {
			callCount++
			if callCount < 2 {
				return nil, errors.New("temporary error")
			}
			return &entity.PriceResponse{
				BTC: &entity.Price{Symbol: entity.BTC, Price: "50000.00"},
			}, nil
		},
	}

	service := &CryptService{
		CryptClient: mockCrypto,
		logger:      slog.Default(),
	}

	prices, err := service.getPricesWithRetry(context.Background())
	if err != nil {
		t.Fatalf("getPricesWithRetry failed: %v", err)
	}

	if prices.BTC.Price != "50000.00" {
		t.Errorf("Price = %v, want %v", prices.BTC.Price, "50000.00")
	}

	if callCount != 2 {
		t.Errorf("Retry count = %v, want %v", callCount, 2)
	}
}
