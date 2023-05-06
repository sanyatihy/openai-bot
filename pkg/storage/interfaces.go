package storage

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/sanyatihy/openai-bot/pkg/telegram"
	"github.com/sanyatihy/openai-go/pkg/openai"
)

type PostgresStorage interface {
	GetChatContext(ctx context.Context, chatID int) (string, []openai.Message, error)
	UpdateChatContext(ctx context.Context, chatID int, messages []openai.Message, model string) error
	ClearChatContext(ctx context.Context, chatID int) error
	UpdateChatModel(ctx context.Context, chatID int, gptModel string) error
	RunInitialMigrations(ctx context.Context) error
}

type PostgresQueue interface {
	InsertChatUpdate(ctx context.Context, update telegram.Update) error
	GetNextChatUpdate(ctx context.Context, status string) (int, telegram.Update, error)
	GetLastChatUpdateID(ctx context.Context) (int, error)
	SetChatUpdateStatus(ctx context.Context, updateID int, status string) error
	ResetChatUpdatesStatus(ctx context.Context) error
}

type DBPool interface {
	Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error)
	Begin(ctx context.Context) (pgx.Tx, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
}

type DBRow interface {
	Scan(dest ...interface{}) error
}
