package service

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"

	"time"

	"tgBotFinal/internal/entity"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"golang.org/x/sync/errgroup"
)

type CryptService struct {
	CurrencyRepo CurrencyRepository
	UserRepo     UserRepository
	CryptClient  CryptoClient
	Notification Notification
	ChiRouter    Router
	webhookURL   string
	logger       *slog.Logger
	port         string
}

func NewCryptService(CurrencyRepo CurrencyRepository,
	UserRepo UserRepository,
	CryptoClient CryptoClient,
	Notification Notification,
	ChiRouter Router,
	webhookURL string,
	logger *slog.Logger,
	port string,
) *CryptService {

	return &CryptService{
		CurrencyRepo: CurrencyRepo,
		UserRepo:     UserRepo,
		CryptClient:  CryptoClient,
		Notification: Notification,
		ChiRouter:    ChiRouter,
		webhookURL:   webhookURL,
		logger:       logger.With(slog.String("component", "CryptoService")),
		port:         port,
	}
}

func (s *CryptService) Run(ctx context.Context) error {
	s.logger.Debug("Run CryptoService")

	if err := s.setupTelegramWebhook(ctx, s.webhookURL); err != nil {
		s.logger.Error("Failed to setup Telegram webhook", "error", err)
	}

	prices, err := s.getPricesWithRetry(ctx)
	if err != nil {
		return fmt.Errorf("failed to get prices: %w", err)
	}

	if err := s.savePrices(ctx, prices); err != nil {
		s.logger.Warn("Failed to save prices", "err", err)
	}

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		return s.runNotificationWorker(ctx)
	})

	g.Go(func() error {
		return s.runCacheRefreshWorker(ctx)
	})

	g.Go(func() error {
		s.ChiRouter.SetupMiddleware()
		s.ChiRouter.SetupRoutes()

		s.logger.Info("Starting HTTP server on port 8080")
		if err := s.ChiRouter.Start("8080"); err != nil {
			return fmt.Errorf("HTTP server failed: %w", err)
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		return fmt.Errorf("service stopped with error: %w", err)
	}

	return nil
}

func (s *CryptService) runCacheRefreshWorker(ctx context.Context) error {
	defer func() {
		if r := recover(); r != nil {
			s.logger.Error("panic in cache refresh worker", "recover", r)
		}
	}()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("Stopping cache refresh worker")
			return nil
		case <-ticker.C:
			if err := s.refreshCache(ctx); err != nil {
				s.logger.Error("Failed to refresh cache", "err", err)
			}
		}
	}
}

func (s *CryptService) refreshCache(ctx context.Context) error {
	prices, err := s.CryptClient.GetAllPrices(ctx)
	if err != nil {
		return fmt.Errorf("failed to get prices: %w", err)
	}

	if err := s.savePrices(ctx, prices); err != nil {
		s.logger.Warn("Failed to save prices during cache refresh", "err", err)
	}

	s.logger.Debug("Cache refresh successfully")
	return nil
}

func (s *CryptService) getPricesWithRetry(ctx context.Context) (*entity.PriceResponse, error) {
	maxRetries := 5
	initialDelay := 100 * time.Millisecond
	maxDelay := 10 * time.Minute

	delay := initialDelay

	for i := 0; i < maxRetries; i++ {
		prices, err := s.CryptClient.GetAllPrices(ctx)
		if err == nil {
			return prices, nil
		}

		s.logger.Warn("Failed to get prices, retrying", "attemp", i+1, "maxRetries", maxRetries, "error", err)

		if i == maxRetries-1 {
			return nil, fmt.Errorf("failed to get prices after %d attemps: %w", maxRetries, err)
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(delay):
		}

		delay *= 2
		if delay > maxDelay {
			delay = maxDelay
		}
	}

	return nil, fmt.Errorf("unexpected error in getPriceWithRetry")
}

func (s *CryptService) savePrices(ctx context.Context, prices *entity.PriceResponse) error {

	var errs []error

	if prices.BTC != nil {
		if err := s.CurrencyRepo.SaveOrUpdate(ctx, prices.BTC); err != nil {
			errs = append(errs, fmt.Errorf("failed to save BTC price: %w", err))
		}
	}

	if prices.ETH != nil {
		if err := s.CurrencyRepo.SaveOrUpdate(ctx, prices.ETH); err != nil {
			errs = append(errs, fmt.Errorf("failed to save ETH price: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("failed to save some prices: %v", errs)
	}

	return nil
}

func (s *CryptService) runNotificationWorker(ctx context.Context) error {

	defer func() {
		if r := recover(); r != nil {
			s.logger.Error("panic in notification worker", "recover", r)
		}
	}()

	ticker := time.NewTicker(15 * time.Minute)
	defer ticker.Stop()
	s.logger.Debug("Run Notification Worker")

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("Notification worker stopped")
			return nil
		case <-ticker.C:
			if err := s.sendNotificationsToActive(ctx); err != nil {
				s.logger.Error("failed to send notifications", "error", err)
			}
		}
	}
}

func (s *CryptService) sendNotificationsToActive(ctx context.Context) error {
	users, err := s.UserRepo.GetAllActive(ctx)
	if err != nil {
		s.logger.Error("failed to get all active users", "err", err)
		return fmt.Errorf("get active users: %w", err)
	}

	prices, err := s.CryptClient.GetAllPrices(ctx)
	if err != nil {
		s.logger.Error("failed to get all prices", "err", err)
		return fmt.Errorf("get prices: %w", err)
	}

	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(5)

	for _, user := range users {
		g.Go(func() error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				if err := s.Notification.SendAllPrices(ctx, user.ChatID, prices); err != nil {
					s.logger.Warn("send prices failed", "chatID", user.ChatID, "error", err)
				}
				return nil
			}
		})

	}
	return g.Wait()
}

func (s *CryptService) CheckDB(ctx context.Context) error {
	s.logger.Info("start CheckDB")
	_, err := s.UserRepo.GetAll(ctx)
	if err != nil {
		s.logger.Error("failed to get all users", "err", err)
		return fmt.Errorf("DB connection failed: %w", err)
	}

	return nil
}

func (s *CryptService) CheckByBitAPI(ctx context.Context) error {
	s.logger.Info("start CheckByBitAPI")

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := s.CryptClient.GetPriceBySymbol(ctx, entity.BTC)
	if err != nil {
		s.logger.Error("failed to get price by bit api", "err", err)
		return fmt.Errorf("bybit API unavailable failed: %w", err)
	}

	return nil
}

func (s *CryptService) CheckTelegramAPI(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return s.Notification.CheckAPI(ctx)
}

func (s *CryptService) setupTelegramWebhook(ctx context.Context, webhookURL string) error {
	if webhookURL == "" {
		s.logger.Warn("WEBHOOK_URL not set, Telegram webhook disabled")
		return nil
	}

	botAPI := s.Notification.GetBotAPI()
	if botAPI == nil {
		return fmt.Errorf("Telegram API is not available")
	}

	fullWebhookURL := webhookURL + "/webhook/telegram"

	parsedURL, err := url.Parse(fullWebhookURL)
	if err != nil {
		return fmt.Errorf("failed to parse webhook URL: %w", err)
	}

	s.logger.Info("Setting up Telegram webhook", "url", webhookURL)

	_, err = botAPI.Request(tgbotapi.WebhookConfig{
		URL: parsedURL,
	})

	if err != nil {
		return fmt.Errorf("failed to set webhook: %w", err)
	}

	s.logger.Info("Telegram webhook setup successfully")
	return nil
}
