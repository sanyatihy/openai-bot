package processor

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sanyatihy/openai-bot/pkg/telegram"
	"github.com/sanyatihy/openai-go/pkg/openai"
	"go.uber.org/zap"
)

type processor struct {
	logger       *zap.Logger
	openAIClient openai.Client
	tgBotClient  telegram.BotClient
	db           *pgxpool.Pool
}

func NewProcessor(logger *zap.Logger,
	openAIClient openai.Client,
	tgBotClient telegram.BotClient,
	db *pgxpool.Pool,
) Processor {
	return &processor{
		logger:       logger,
		openAIClient: openAIClient,
		tgBotClient:  tgBotClient,
		db:           db,
	}
}
