package service

import (
	"context"

	"tgBotFinal/internal/entity"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Notification interface {
	SendAllPrices(ctx context.Context, chatID int64, prices *entity.PriceResponse) error
	ActivateUser(ctx context.Context, chatID int64) error
	DeactivateUser(ctx context.Context, chatID int64) error
	SendInfoMessage(ctx context.Context, chatID int64, text string) error
	CheckAPI(ctx context.Context) error
	GetBotAPI() *tgbotapi.BotAPI
}

type CryptoClient interface {
	GetPriceBySymbol(ctx context.Context, symbol entity.CurrencyName) (*entity.Price, error)
	GetAllPrices(ctx context.Context) (*entity.PriceResponse, error)
}

type CurrencyRepository interface {
	SaveOrUpdate(ctx context.Context, currency *entity.Price) error
	GetBySymbol(ctx context.Context, symbol entity.CurrencyName) (*entity.Price, error)
	GetAll(ctx context.Context) ([]*entity.Price, error)
}

type UserRepository interface {
	SaveOrUpdate(ctx context.Context, user *entity.User) error
	GetByChatID(ctx context.Context, chatID int64) (*entity.User, error)
	GetAll(ctx context.Context) ([]*entity.User, error)
	GetAllActive(ctx context.Context) ([]*entity.User, error)
}

type BotHandler interface {
	HandleWelcome() string
	HandleStart() string
	HandleStop() string
	HandleHelp() string
}

type Router interface {
	SetupMiddleware()
	SetupRoutes()
	Start(string) error
	Shutdown(context.Context) error
}

type HealthChecker interface {
	CheckDB(ctx context.Context) error
	CheckByBitAPI(ctx context.Context) error
	CheckTelegramAPI(ctx context.Context) error
}
