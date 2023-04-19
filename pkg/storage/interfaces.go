package storage

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
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

type DBPool interface {
	Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error)
	Begin(ctx context.Context) (pgx.Tx, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
}

type DBRow interface {
	Scan(dest ...interface{}) error
}
