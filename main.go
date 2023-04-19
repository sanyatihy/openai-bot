package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/sanyatihy/openai-bot/pkg/processor"
	storage "github.com/sanyatihy/openai-bot/pkg/storage"
	"github.com/sanyatihy/openai-bot/pkg/telegram"
	"github.com/sanyatihy/openai-go/pkg/openai"
	"go.uber.org/zap"
)

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Sync()

	if err = godotenv.Load(); err != nil {
		logger.Error("Error loading .env file")
	}

	envVars := map[string]string{
		"OPENAI_API_KEY":     "",
		"OPENAI_ORG_ID":      "",
		"TELEGRAM_BOT_TOKEN": "",
		"POSTGRES_DSN":       "",
	}

	for envVar := range envVars {
		value, exists := os.LookupEnv(envVar)
		if !exists {
			logger.Error(fmt.Sprintf("Environment variable %s not found", envVar))
			os.Exit(1)
		}
		envVars[envVar] = value
	}

	connString := envVars["POSTGRES_DSN"]
	dbpool, err := pgxpool.New(context.Background(), connString)
	if err != nil {
		logger.Error("Failed to connect to the database", zap.Error(err))
	}
	defer dbpool.Close()

	transport := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: false,
	}
	httpClient := &http.Client{
		Transport: transport,
	}

	openAIClient := openai.NewClient(httpClient, envVars["OPENAI_API_KEY"], envVars["OPENAI_ORG_ID"])
	tgBotClient := telegram.NewBotClient(httpClient, envVars["TELEGRAM_BOT_TOKEN"])
	db := storage.NewPostgresStorage(dbpool)
	queue := storage.NewPostgresQueue(dbpool)
	proc := processor.NewProcessor(logger, openAIClient, tgBotClient, db, queue, 5, 16)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	if err := proc.Start(); err != nil {
		logger.Error("Failed to start processing", zap.Error(err))
		os.Exit(1)
	}

	<-sigChan
	logger.Info("Shutting down...")
}
