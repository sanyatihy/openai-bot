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
	queueSemaphore    chan struct{}
}

func NewProcessor(logger *zap.Logger,
	openAIClient openai.Client,
	tgBotClient telegram.BotClient,
	db storage.PostgresStorage,
	queue storage.PostgresQueue,
	concurrentWorkers int,
) Processor {
	return &processor{
		logger:            logger,
		openAIClient:      openAIClient,
		tgBotClient:       tgBotClient,
		db:                db,
		queue:             queue,
		concurrentWorkers: concurrentWorkers,
		queueSemaphore:    make(chan struct{}, concurrentWorkers),
	}
}
