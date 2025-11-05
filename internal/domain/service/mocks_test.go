package service

import (
	"context"
	"tgBotFinal/internal/entity"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type MockCryptoClient struct {
	GetPriceBySymbolFunc func(ctx context.Context, symbol entity.CurrencyName) (*entity.Price, error)
	GetAllPricesFunc     func(ctx context.Context) (*entity.PriceResponse, error)
}

func (m *MockCryptoClient) GetPriceBySymbol(ctx context.Context, symbol entity.CurrencyName) (*entity.Price, error) {
	return m.GetPriceBySymbolFunc(ctx, symbol)
}

func (m *MockCryptoClient) GetAllPrices(ctx context.Context) (*entity.PriceResponse, error) {
	return m.GetAllPricesFunc(ctx)
}

type MockNotification struct {
	SendAllPricesFunc   func(ctx context.Context, chatID int64, prices *entity.PriceResponse) error
	ActivateUserFunc    func(ctx context.Context, chatID int64) error
	DeactivateUserFunc  func(ctx context.Context, chatID int64) error
	SendInfoMessageFunc func(ctx context.Context, chatID int64, text string) error
	CheckAPIFunc        func(ctx context.Context) error
	GetBotAPIFunc       func() *tgbotapi.BotAPI
}

func (m *MockNotification) SendAllPrices(ctx context.Context, chatID int64, prices *entity.PriceResponse) error {
	return m.SendAllPricesFunc(ctx, chatID, prices)
}

func (m *MockNotification) ActivateUser(ctx context.Context, chatID int64) error {
	return m.ActivateUserFunc(ctx, chatID)
}

func (m *MockNotification) DeactivateUser(ctx context.Context, chatID int64) error {
	return m.DeactivateUserFunc(ctx, chatID)
}

func (m *MockNotification) SendInfoMessage(ctx context.Context, chatID int64, text string) error {
	return m.SendInfoMessageFunc(ctx, chatID, text)
}

func (m *MockNotification) CheckAPI(ctx context.Context) error {
	return m.CheckAPIFunc(ctx)
}

func (m *MockNotification) GetBotAPI() *tgbotapi.BotAPI {
	return m.GetBotAPIFunc()
}

type MockUserRepository struct {
	SaveOrUpdateUserFunc func(ctx context.Context, user *entity.User) error
	GetByChatIDFunc      func(ctx context.Context, chatID int64) (*entity.User, error)
	GetAllFunc           func(ctx context.Context) ([]*entity.User, error)
	GetAllActiveFunc     func(ctx context.Context) ([]*entity.User, error)
}

func (m *MockUserRepository) SaveOrUpdate(ctx context.Context, user *entity.User) error {
	return m.SaveOrUpdateUserFunc(ctx, user)
}

func (m *MockUserRepository) GetByChatID(ctx context.Context, chatID int64) (*entity.User, error) {
	return m.GetByChatIDFunc(ctx, chatID)
}

func (m *MockUserRepository) GetAll(ctx context.Context) ([]*entity.User, error) {
	return m.GetAllFunc(ctx)
}

func (m *MockUserRepository) GetAllActive(ctx context.Context) ([]*entity.User, error) {
	return m.GetAllActiveFunc(ctx)
}

type MockCurrencyRepository struct {
	SaveOrUpdateFunc func(ctx context.Context, currency *entity.Price) error
	GetBySymbolFunc  func(ctx context.Context, symbol entity.CurrencyName) (*entity.Price, error)
	GetAllFunc       func(ctx context.Context) ([]*entity.Price, error)
}

func (m *MockCurrencyRepository) SaveOrUpdate(ctx context.Context, currency *entity.Price) error {
	return m.SaveOrUpdateFunc(ctx, currency)
}

func (m *MockCurrencyRepository) GetBySymbol(ctx context.Context, symbol entity.CurrencyName) (*entity.Price, error) {
	return m.GetBySymbolFunc(ctx, symbol)
}

func (m *MockCurrencyRepository) GetAll(ctx context.Context) ([]*entity.Price, error) {
	return m.GetAllFunc(ctx)
}

func NewMockCryptoClient() *MockCryptoClient {
	return &MockCryptoClient{
		GetPriceBySymbolFunc: func(ctx context.Context, symbol entity.CurrencyName) (*entity.Price, error) {
			return &entity.Price{
				Symbol: symbol,
				Price:  "50000.00",
			}, nil
		},
		GetAllPricesFunc: func(ctx context.Context) (*entity.PriceResponse, error) {
			return &entity.PriceResponse{
				BTC: &entity.Price{Symbol: entity.BTC, Price: "50000.00"},
				ETH: &entity.Price{Symbol: entity.ETH, Price: "3000.00"},
			}, nil
		},
	}
}

func NewMockNotification() *MockNotification {
	return &MockNotification{
		SendAllPricesFunc: func(ctx context.Context, chatID int64, prices *entity.PriceResponse) error {
			return nil
		},
		ActivateUserFunc: func(ctx context.Context, chatID int64) error {
			return nil
		},
		DeactivateUserFunc: func(ctx context.Context, chatID int64) error {
			return nil
		},
		SendInfoMessageFunc: func(ctx context.Context, chatID int64, text string) error {
			return nil
		},
		CheckAPIFunc: func(ctx context.Context) error {
			return nil
		},
		GetBotAPIFunc: func() *tgbotapi.BotAPI {
			return nil
		},
	}
}

func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		GetAllActiveFunc: func(ctx context.Context) ([]*entity.User, error) {

			return []*entity.User{
				{ChatID: 12345, Username: "testuser1", Active: true},
				{ChatID: 67890, Username: "testuser2", Active: true},
			}, nil
		},
		SaveOrUpdateUserFunc: func(ctx context.Context, user *entity.User) error {
			return nil
		},
		GetByChatIDFunc: func(ctx context.Context, chatID int64) (*entity.User, error) {
			return &entity.User{ChatID: chatID, Username: "testuser", Active: true}, nil
		},
		GetAllFunc: func(ctx context.Context) ([]*entity.User, error) {
			return []*entity.User{}, nil
		},
	}
}

func NewMockCurrencyRepository() *MockCurrencyRepository {
	return &MockCurrencyRepository{
		SaveOrUpdateFunc: func(ctx context.Context, currency *entity.Price) error {
			return nil
		},
		GetBySymbolFunc: func(ctx context.Context, symbol entity.CurrencyName) (*entity.Price, error) {
			return &entity.Price{Symbol: symbol, Price: "50000.00"}, nil
		},
		GetAllFunc: func(ctx context.Context) ([]*entity.Price, error) {
			return []*entity.Price{}, nil
		},
	}
}
