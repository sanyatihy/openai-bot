package processor

import (
	"github.com/sanyatihy/openai-bot/pkg/telegram"
	"github.com/sanyatihy/openai-go/pkg/openai"
	"go.uber.org/zap"
)

type processor struct {
	logger       *zap.Logger
	openAIClient openai.Client
	tgBotClient  telegram.BotClient
}

func NewProcessor(logger *zap.Logger, openAIClient openai.Client, tgBotClient telegram.BotClient) Processor {
	return &processor{
		logger:       logger,
		openAIClient: openAIClient,
		tgBotClient:  tgBotClient,
	}
}
