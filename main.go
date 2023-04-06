package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/sanyatihy/openai-bot/pkg/processor"
	"github.com/sanyatihy/openai-bot/pkg/telegram"
	"github.com/sanyatihy/openai-go/pkg/openai"
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

	openAIApiKey := os.Getenv("OPENAI_API_KEY")
	openAIOrgID := os.Getenv("OPENAI_ORG_ID")
	telegramToken := os.Getenv("TELEGRAM_BOT_TOKEN")

	transport := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: false,
	}
	httpClient := &http.Client{
		Transport: transport,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	openAIClient := openai.NewClient(httpClient, app.logger, openAIApiKey, openAIOrgID)
	tgBotClient := telegram.NewBotClient(httpClient, telegramToken)
	proc := processor.NewProcessor(app.logger, openAIClient, tgBotClient)

	if err := proc.Start(ctx); err != nil {
		app.logger.Error("Failed to start processing", zap.Error(err))
		os.Exit(1)
	}
}
