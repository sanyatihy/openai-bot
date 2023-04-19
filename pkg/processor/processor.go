package processor

import (
	"github.com/sanyatihy/openai-bot/pkg/storage"
	"github.com/sanyatihy/openai-bot/pkg/telegram"
	"github.com/sanyatihy/openai-go/pkg/openai"
	"go.uber.org/zap"
)

type processor struct {
	logger            *zap.Logger
	openAIClient      openai.Client
	tgBotClient       telegram.BotClient
	db                storage.PostgresStorage
	queue             storage.PostgresQueue
	concurrentWorkers int
	queueUpdates      chan updateWithID
	queueBufferSize   int
}

func NewProcessor(logger *zap.Logger,
	openAIClient openai.Client,
	tgBotClient telegram.BotClient,
	db storage.PostgresStorage,
	queue storage.PostgresQueue,
	concurrentWorkers int,
	queueBufferSize int,
) Processor {
	return &processor{
		logger:            logger,
		openAIClient:      openAIClient,
		tgBotClient:       tgBotClient,
		db:                db,
		queue:             queue,
		concurrentWorkers: concurrentWorkers,
		queueUpdates:      make(chan updateWithID, queueBufferSize),
	}
}
