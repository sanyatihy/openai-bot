package storage

import (
	"context"
	"github.com/sanyatihy/openai-bot/pkg/telegram"
	"github.com/sanyatihy/openai-go/pkg/openai"
)

type PostgresStorage interface {
	GetChatContext(ctx context.Context, chatID int) ([]openai.Message, error)
	UpdateChatContext(ctx context.Context, chatID int, messages []openai.Message) error
	ClearChatContext(ctx context.Context, chatID int) error
	RunInitialMigrations(ctx context.Context) error
}

type PostgresQueue interface {
	InsertChatUpdate(ctx context.Context, update telegram.Update) error
	GetNextChatUpdate(ctx context.Context, status string) (int, telegram.Update, error)
	GetLastChatUpdate(ctx context.Context) (int, error)
	SetChatUpdateStatus(ctx context.Context, updateID int, status string) error
}
