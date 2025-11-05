package telegram

import (
	"context"
	"fmt"
	"log/slog"
	"tgBotFinal/internal/domain/service"
	"time"

	"tgBotFinal/internal/entity"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type NotificationTelegram struct {
	api    *tgbotapi.BotAPI
	logger *slog.Logger
}

func NewNotificationTelegram(logger *slog.Logger, token string) (service.Notification, error) {

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		logger.Error("error creating telegram bot ", "error", err)
		return nil, err
	}
	notificationTelegram := &NotificationTelegram{
		api:    bot,
		logger: logger.With(slog.String("component", "NotificationTelegram")),
	}
	return notificationTelegram, nil
}

func (n *NotificationTelegram) SendAllPrices(ctx context.Context, chatID int64, prices *entity.PriceResponse) error {
	n.logger.Debug("Starting sendAllPrices")

	message := "Current Crypto Prices: \n"
	if prices.BTC != nil {
		message += fmt.Sprintf("BTC: %s\n", prices.BTC.Price)
	}
	if prices.ETH != nil {
		message += fmt.Sprintf("ETH: %s\n", prices.ETH.Price)
	}

	message += fmt.Sprintf("Last update: %s", time.Now().Format("2006-01-02 15:04:05"))

	if err := n.sendMessage(chatID, message); err != nil {
		n.logger.Error("error sending message ", "error", err)
		return err
	}

	return nil
}

func (n *NotificationTelegram) ActivateUser(ctx context.Context, chatID int64) error {
	n.logger.Debug("Starting activateUser")
	message := "You have been successfully activated"
	if err := n.sendMessage(chatID, message); err != nil {
		n.logger.Error("error sending message ", "error", err)
		return err
	}

	return nil
}

func (n *NotificationTelegram) DeactivateUser(ctx context.Context, chatID int64) error {
	n.logger.Debug("Starting deactivateUser")

	message := "You have been successfully deactivated"

	if err := n.sendMessage(chatID, message); err != nil {
		n.logger.Error("error sending message ", "error", err)
		return err
	}

	return nil
}

func (n *NotificationTelegram) SendInfoMessage(ctx context.Context, chatID int64, text string) error {
	n.logger.Debug("Starting sendInfoMessage")

	if err := n.sendMessage(chatID, text); err != nil {
		n.logger.Error("error sending message ", "error", err)
		return err
	}

	return nil
}

func (n *NotificationTelegram) sendMessage(chatID int64, text string) error {
	msg := tgbotapi.NewMessage(chatID, text)
	_, err := n.api.Send(msg)
	return err
}

func (n *NotificationTelegram) CheckAPI(ctx context.Context) error {
	n.logger.Debug("Starting checkAPI")

	_, err := n.api.GetMe()
	if err != nil {
		n.logger.Error("telegram API check failed", "error", err)
		return fmt.Errorf("telegram API unavailable: %w", err)
	}

	n.logger.Debug("Telegram API check passed")
	return nil
}

func (n *NotificationTelegram) GetBotAPI() *tgbotapi.BotAPI {
	return n.api
}
