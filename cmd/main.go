package main

import (
	"context"
	"database/sql"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"tgBotFinal/internal/domain/service"
	"tgBotFinal/internal/infrastructure/cryptoClient"
	"tgBotFinal/internal/infrastructure/database"
	"time"

	"tgBotFinal/internal/config"
	"tgBotFinal/internal/infrastructure/cryptoClient/bybit"
	"tgBotFinal/internal/infrastructure/notification/telegram"
	"tgBotFinal/internal/infrastructure/router/chi"
	"tgBotFinal/internal/logger"
	"tgBotFinal/internal/repository/postgres"

	_ "github.com/lib/pq"
)

func main() {
	migrateOnly := flag.Bool("migrate", false, "Run database migrations and exit")
	flag.Parse()

	// Load config
	cfg := config.MustLoadConfig()

	//Init logger
	appLog := logger.NewLogger(cfg.LogLevel)
	appLog.Info("Starting App")
	appLog.Info("Loading config")

	//Init DB
	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		appLog.Error("Error connecting to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(time.Minute * 5)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		appLog.Error("Error pinging database", "error", err)
		os.Exit(1)
	}
	appLog.Debug("Connected to database")

	//Auto-migrate database
	migrator := database.NewMigrator(db, appLog)
	if err := migrator.CheckAndMigrate("migrations"); err != nil {
		appLog.Error("Database migration failed", "error", err)
		os.Exit(1)
	}

	if *migrateOnly {
		appLog.Info("Migrations completed successfully, exiting...")
		return
	}

	//Init repositories
	currencyRepo := postgres.NewCurrencyRepo(db, appLog)
	userRepo := postgres.NewUserRepo(db, appLog)

	//init cryptoClient
	byBitClient := bybit.NewClient(appLog, cfg.APIUrl)

	// Wrap with cached client (TTL = 1 minute)
	cachedClient := cryptoClient.NewCachedClient(byBitClient, time.Minute, appLog)

	//init notification
	tgNotifier, err := telegram.NewNotificationTelegram(appLog, cfg.TgToken)
	if err != nil {
		appLog.Error("Error initializing Telegram", "error", err)
		os.Exit(1)
	}

	//init service
	serv := service.NewCryptService(
		currencyRepo,
		userRepo,
		cachedClient,
		tgNotifier,
		nil,
		cfg.WebhookURL,
		appLog,
		"HTTP_PORT",
	)

	//init router
	router := chi.NewChiRouter(appLog, userRepo, tgNotifier, cachedClient, serv)
	serv.ChiRouter = router
	// Graceful Shutdown
	mainCtx, stop := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT,
	)
	defer stop()

	serviceErr := make(chan error, 1)
	go func() {
		appLog.Info("Starting server")
		if err := serv.Run(mainCtx); err != nil && err != http.ErrServerClosed {
			appLog.Error("Error starting server", "error", err)
			serviceErr <- err
		}
		close(serviceErr)
	}()

	select {
	case <-mainCtx.Done():
		appLog.Info("Received shutdown signal, initiating graceful shutdown...")
	case err := <-serviceErr:
		if err != nil {
			appLog.Error("Service encountered error", "error", err)
		}
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	appLog.Info("Shutting down gracefully...")

	if err := router.Shutdown(shutdownCtx); err != nil {
		appLog.Error("Error shutting down router", "error", err)
	}

	appLog.Info("Shutting down successfully")

}
