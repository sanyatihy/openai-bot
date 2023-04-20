package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/sanyatihy/openai-go/pkg/openai"
)

const (
	schema            = "content"
	chatContextTable  = "chat_context"
	lastUpdateIDTable = "last_update_id"
)

const (
	createSchemaQuery = "CREATE SCHEMA IF NOT EXISTS %s;"

	createChatContextTableQuery = "CREATE TABLE IF NOT EXISTS %s.%s (chat_id BIGINT PRIMARY KEY, context JSONB);"
	getChatContextQuery         = "SELECT context FROM %s.%s WHERE chat_id = $1;"
	updateChatContextQuery      = "INSERT INTO %s.%s (chat_id, context) VALUES ($1, $2) ON CONFLICT (chat_id) DO UPDATE SET context = EXCLUDED.context;"
	deleteChatContextQuery      = "DELETE FROM %s.%s WHERE chat_id = $1;"
)

type postgresStorage struct {
	db DBPool
}

func NewPostgresStorage(db DBPool) PostgresStorage {
	return &postgresStorage{
		db: db,
	}
}

func (s *postgresStorage) GetChatContext(ctx context.Context, chatID int) ([]openai.Message, error) {
	var contextJSON string

	err := s.db.QueryRow(ctx, fmt.Sprintf(getChatContextQuery, schema, chatContextTable), chatID).Scan(&contextJSON)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	var messages []openai.Message
	err = json.Unmarshal([]byte(contextJSON), &messages)
	if err != nil {
		return nil, err
	}

	return messages, nil
}

func (s *postgresStorage) UpdateChatContext(ctx context.Context, chatID int, messages []openai.Message) error {
	contextJSON, err := json.Marshal(messages)
	if err != nil {
		return err
	}

	_, err = s.db.Exec(ctx, fmt.Sprintf(updateChatContextQuery, schema, chatContextTable), chatID, string(contextJSON))
	return err
}

func (s *postgresStorage) ClearChatContext(ctx context.Context, chatID int) error {
	_, err := s.db.Exec(ctx, fmt.Sprintf(deleteChatContextQuery, schema, chatContextTable), chatID)
	return err
}

func (s *postgresStorage) RunInitialMigrations(ctx context.Context) error {
	_, err := s.db.Exec(ctx, fmt.Sprintf(createSchemaQuery, schema))
	if err != nil {
		return fmt.Errorf("failed to create schema %s: %w", schema, err)
	}

	_, err = s.db.Exec(ctx, fmt.Sprintf(createChatContextTableQuery, schema, chatContextTable))
	if err != nil {
		return fmt.Errorf("failed to create table %s: %w", chatContextTable, err)
	}

	_, err = s.db.Exec(ctx, fmt.Sprintf(createChatUpdatesTableQuery, schema, chatUpdatesTable))
	if err != nil {
		return fmt.Errorf("failed to create table %s: %w", lastUpdateIDTable, err)
	}

	return nil
}
