package chi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"tgBotFinal/internal/domain/service"
	"time"

	"tgBotFinal/internal/entity"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type ChiRouter struct {
	router        *chi.Mux
	server        *http.Server
	logger        *slog.Logger
	userRepo      service.UserRepository
	notification  service.Notification
	cryptClient   service.CryptoClient
	healthChecker service.HealthChecker
}

func NewChiRouter(
	logger *slog.Logger,
	userRepo service.UserRepository,
	notification service.Notification,
	cryptClient service.CryptoClient,
	healthChecker service.HealthChecker,
) service.Router {
	return &ChiRouter{
		router:        chi.NewRouter(),
		logger:        logger.With(slog.String("component", "chi.Router")),
		userRepo:      userRepo,
		notification:  notification,
		cryptClient:   cryptClient,
		healthChecker: healthChecker,
	}
}

func (c *ChiRouter) SetupMiddleware() {
	c.logger.Debug("Setting up middleware")

	c.router.Use(middleware.RequestID)
	c.router.Use(middleware.RealIP)
	c.router.Use(middleware.Logger)
	c.router.Use(middleware.Recoverer)
	c.router.Use(middleware.Timeout(60 * time.Second))

	c.router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	})
}

func (c *ChiRouter) SetupRoutes() {
	c.logger.Debug("Setting up routes")

	// Health check
	c.router.Get("/health", c.healthHandler)
	c.router.Get("/health/live", c.livenessHandler)
	c.router.Get("/health/ready", c.readinessHandler)
	c.router.Get("/health/detalied", c.detailedHealthHandler)

	// API routes
	c.router.Post("/webhook/telegram", c.telegramWebhookHandler)
	c.router.Get("/users/active", c.getActiveUsersHandler)
	c.router.Get("/currensies", c.getCurrenciesHandler)

	c.router.NotFound(c.notFoundHandler)
	c.router.MethodNotAllowed(c.methodNotAllowedHandler)
}

func (c *ChiRouter) Start(port string) error {
	c.logger.Debug("Starting HTTP server", "port", port)

	c.server = &http.Server{
		Addr:         ":" + port,
		Handler:      c.router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	if err := c.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server faied to start: %w", err)
	}

	return nil
}

func (c *ChiRouter) telegramWebhookHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		c.logger.Error("failed to read request body", "error", err)
		http.Error(w, `{"error": "Bad request"}`, http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var update entity.TelegramUpdate

	if err := json.Unmarshal(body, &update); err != nil {
		c.logger.Error("failed to parse telegram update", "error", err)
		http.Error(w, `{"error": "Invalid JSON"}`, http.StatusBadRequest)
		return
	}

	c.logger.Debug("received telegram update", "update", update.UpdateID)

	go c.handleTelegramUpdate(context.Background(), update)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "ok"}`))
}

func (c *ChiRouter) handleTelegramUpdate(ctx context.Context, update entity.TelegramUpdate) {

	defer func() {
		if r := recover(); r != nil {
			c.logger.Error("panic in telegram update handler", "recover", r)
		}
	}()

	if update.Message == nil {
		c.logger.Debug("Update doesn`t contain message")
		return
	}

	message := update.Message
	chatID := message.ChatID
	text := message.Text

	c.logger.Info("Processing telegram message",
		"chat_id", chatID,
		"username", message.From.Username,
		"text", text)

	user := &entity.User{
		ChatID:   chatID.ID,
		Username: message.From.Username,
		Active:   true,
	}

	if err := c.userRepo.SaveOrUpdate(ctx, user); err != nil {
		c.logger.Error("failed to save user", "chatID", chatID, "error", err)
		return
	}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				c.logger.Error("panic in command handler", "chat_id", chatID, "recover", r)
			}
		}()

		switch text {
		case "/start":
			c.handleStartCommand(ctx, chatID.ID, user)
		case "/stop":
			c.handleStopCommand(ctx, chatID.ID, user)
		case "/price", "/prices":
			c.handleHelpCommand(ctx, chatID.ID)
		case "/help":
			c.handleHelpCommand(ctx, chatID.ID)
		default:
			c.handleUnknowCommand(ctx, chatID.ID)
		}
	}()
}

func (c *ChiRouter) handleStartCommand(ctx context.Context, chatID int64, user *entity.User) {
	c.logger.Debug("Handling start command", "chatId", chatID)

	user.Active = true
	if err := c.userRepo.SaveOrUpdate(ctx, user); err != nil {
		c.logger.Error("failed to activate user", "chatID", chatID, "error", err)
		c.notification.SendInfoMessage(ctx, chatID, "Failed to activate user")
		return
	}

	message := `Crypto Price  Bot Activated

Теперь вы будете получать обновления курсов каждые 15 минут.

*Доступные команды:*
/price - Текущие цены
/stop - Отписаться от рассылки
/help - Помощь

*Отслеживаемые монеты:*
• BTC (Bitcoin)
• ETH (Ethereum)`

	if err := c.notification.SendInfoMessage(ctx, chatID, message); err != nil {
		c.logger.Warn("failed to activation message", "chatID", chatID, "error", err)
	}

	c.handlePriceCommand(ctx, chatID)
}

func (c *ChiRouter) handleStopCommand(ctx context.Context, chatID int64, user *entity.User) {
	c.logger.Debug("Handling stop command", "chatId", chatID)

	user.Active = false
	if err := c.userRepo.SaveOrUpdate(ctx, user); err != nil {
		c.logger.Error("failed to deactivated user", "chatID", chatID, "error", err)
		c.notification.SendInfoMessage(ctx, chatID, "Failed to deactivated user")
		return
	}

	if err := c.notification.DeactivateUser(ctx, chatID); err != nil {
		c.logger.Warn("failed to deactivated user", "chatID", chatID, "error", err)
	}
}

func (c *ChiRouter) handlePriceCommand(ctx context.Context, chatID int64) {
	c.logger.Debug("Handling price command", "chatId", chatID)

	prices, err := c.cryptClient.GetAllPrices(ctx)
	if err != nil {
		c.logger.Error("failed to get prices for command", "chatId", chatID, "error", err)
		c.notification.SendInfoMessage(ctx, chatID, "Failed to get prices. Please try again later.")
		return
	}

	if err := c.notification.SendAllPrices(ctx, chatID, prices); err != nil {
		c.logger.Warn("failed to send all prices", "chatId", chatID, "error", err)
	}
}

func (c *ChiRouter) handleHelpCommand(ctx context.Context, chatID int64) {
	c.logger.Debug("Handling help command", "chatId", chatID)

	helpText := `Crypto Price Bot Help

Команды:
/start - Подписаться на рассылку
/stop - Отписаться от рассылки  
/price - Текущие цены монет
/help - Это сообщение

*Функции:
• Автоматическая рассылка каждые 15 минут
• Отслеживание BTC и ETH
• Точные цены с Bybit API

Для начала работы используйте /start`

	if err := c.notification.SendInfoMessage(ctx, chatID, helpText); err != nil {
		c.logger.Warn("failed to send help message", "chatId", chatID, "error", err)
	}
}

func (c *ChiRouter) handleUnknowCommand(ctx context.Context, chatID int64) {
	c.logger.Debug("Handling unknow command", "chatId", chatID)

	message := `Неизвестная команда

Используйте /help для просмотра доступных команд`

	if err := c.notification.SendInfoMessage(ctx, chatID, message); err != nil {
		c.logger.Warn("failed to send unknow message", "chatId", chatID, "error", err)
	}
}

func (c *ChiRouter) getActiveUsersHandler(w http.ResponseWriter, r *http.Request) {
	users, err := c.userRepo.GetAllActive(r.Context())
	if err != nil {
		c.logger.Error("Error getting active users", "error", err)
		http.Error(w, `{"error": "Internal server error"}`, http.StatusInternalServerError)

		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`{"active_users":%d}`, len(users))))
}

func (c *ChiRouter) getCurrenciesHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: Get currencies from repository and return as JSON

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"currencies":["BTC", "ETH"]}`))
}

func (c *ChiRouter) notFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte(`{"error": "Not found"}`))
}

func (c *ChiRouter) methodNotAllowedHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusMethodNotAllowed)
	w.Write([]byte(`{"error": "Method not allowed"}`))
}

func (c *ChiRouter) Shutdown(ctx context.Context) error {
	c.logger.Info("Shutting down HTTP server")

	if c.server != nil {
		if err := c.server.Shutdown(ctx); err != nil {
			c.logger.Error("Error shutting down HTTP server", "error", err)
			return err
		}
		c.logger.Info("HTTP server shut down successfully")
	}
	return nil
}

func (c *ChiRouter) CheckTelegramAPI(ctx context.Context) error {

	return c.healthChecker.CheckTelegramAPI(ctx)
}

// Basic health check
func (c *ChiRouter) healthHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "ok", "timestamp": "` + time.Now().Format(time.RFC3339) + `"}`))
}

// Liveness probe
func (c *ChiRouter) livenessHandler(w http.ResponseWriter, r *http.Request) {
	response := entity.NewHealthResponse()
	response.Status = entity.HealStatusOk

	c.writeHealthResponse(w, response, http.StatusOK)
}

func (c *ChiRouter) readinessHandler(w http.ResponseWriter, r *http.Request) {
	response := entity.NewHealthResponse()

	if err := c.checkDatabase(r.Context()); err != nil {
		response.Checks["database"] = entity.HealCheck{
			Status:    entity.HealStatusError,
			Message:   err.Error(),
			Timestamp: time.Now(),
		}
	}

	if response.Status == entity.HealStatusError {
		c.writeHealthResponse(w, response, http.StatusServiceUnavailable)
	} else {
		c.writeHealthResponse(w, response, http.StatusOK)
	}
}

func (c *ChiRouter) detailedHealthHandler(w http.ResponseWriter, r *http.Request) {
	response := entity.NewHealthResponse()
	allHealthy := true

	// Check DB
	if err := c.checkDatabase(r.Context()); err != nil {
		response.Checks["database"] = entity.HealCheck{
			Status:    entity.HealStatusError,
			Message:   err.Error(),
			Timestamp: time.Now(),
		}
		allHealthy = false
	} else {
		response.Checks["database"] = entity.HealCheck{
			Status:    entity.HealStatusOk,
			Message:   "Database is healthy",
			Timestamp: time.Now(),
		}
	}

	// check BB API
	if err := c.checkByBitAPI(r.Context()); err != nil {
		response.Checks["bybit"] = entity.HealCheck{
			Status:    entity.HealStatusError,
			Message:   err.Error(),
			Timestamp: time.Now(),
		}
		allHealthy = false
	} else {
		response.Checks["bybit"] = entity.HealCheck{
			Status:    entity.HealStatusOk,
			Message:   "ByBit API is healthy",
			Timestamp: time.Now(),
		}
	}

	// check TG
	if err := c.CheckTelegramAPI(r.Context()); err != nil {
		response.Checks["telegram"] = entity.HealCheck{
			Status:    entity.HealStatusError,
			Message:   err.Error(),
			Timestamp: time.Now(),
		}
	} else {
		response.Checks["telegram"] = entity.HealCheck{
			Status:    entity.HealStatusOk,
			Message:   "Telegram API is healthy",
			Timestamp: time.Now(),
		}
	}

	if !allHealthy {
		response.Status = entity.HealStatusError
		c.writeHealthResponse(w, response, http.StatusServiceUnavailable)
	} else if response.Status != entity.HealStatusOk {
		c.writeHealthResponse(w, response, http.StatusOK)
	} else {
		response.Status = entity.HealStatusDegraded
		c.writeHealthResponse(w, response, http.StatusOK)
	}

}

func (c *ChiRouter) checkDatabase(ctx context.Context) error {

	return c.healthChecker.CheckDB(ctx)
}

func (c *ChiRouter) checkByBitAPI(ctx context.Context) error {

	return c.healthChecker.CheckByBitAPI(ctx)
}

func (c *ChiRouter) writeHealthResponse(w http.ResponseWriter, response *entity.HealthResponse, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		c.logger.Error("failed to encode health response", "error", err)
		http.Error(w, `{"error": "Internal server error"}`, http.StatusInternalServerError)
	}
}
