package main

import (
	"context"
	"fmt"
	"github.com/sanyatihy/openai-go/pkg/openai"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/sanyatihy/openai-bot/pkg/processor"
	"github.com/sanyatihy/openai-bot/pkg/telegram"
	"go.uber.org/zap"
)

type App struct {
	logger *zap.Logger
}

func newApp() *App {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Failed to create logger: %v", err)
	}

	return &App{
		logger: logger,
	}
}

func main() {
	app := newApp()
	defer app.logger.Sync()

	err := godotenv.Load()
	if err != nil {
		app.logger.Error("Error loading .env file")
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
			app.logger.Error(fmt.Sprintf("Environment variable %s not found", envVar))
			os.Exit(1)
		}
		envVars[envVar] = value
	}

	connString := envVars["POSTGRES_DSN"]
	dbpool, err := pgxpool.New(context.Background(), connString)
	if err != nil {
		app.logger.Error("Failed to connect to the database", zap.Error(err))
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
	proc := processor.NewProcessor(app.logger, openAIClient, tgBotClient, dbpool)

	if err := proc.Start(); err != nil {
		app.logger.Error("Failed to start processing", zap.Error(err))
		os.Exit(1)
	}
}
