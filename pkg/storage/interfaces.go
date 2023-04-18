package storage

import (
	"context"
	"github.com/sanyatihy/openai-go/pkg/openai"
)

type Storage interface {
	GetChatContext(ctx context.Context, chatID int) ([]openai.Message, error)
	UpdateChatContext(ctx context.Context, chatID int, messages []openai.Message) error
	ClearChatContext(ctx context.Context, chatID int) error
	LoadLastUpdateID(ctx context.Context) (int, error)
	SaveLastUpdateID(ctx context.Context, lastUpdateID int) error
	RunInitialMigrations(ctx context.Context) error
}
