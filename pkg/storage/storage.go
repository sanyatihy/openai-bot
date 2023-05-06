package storage

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/sanyatihy/openai-go/pkg/openai"
)

const (
	chatContextTable = "chat_context"
)

const (
	createChatContextTableQuery = "CREATE TABLE IF NOT EXISTS %s.%s (chat_id BIGINT PRIMARY KEY, model_id VARCHAR(20), context JSONB);"
	getChatContextQuery         = "SELECT model_id, context FROM %s.%s WHERE chat_id = $1;"
	updateChatContextQuery      = "INSERT INTO %s.%s (chat_id, context, model_id) VALUES ($1, $2, $3) ON CONFLICT (chat_id) DO UPDATE SET context = EXCLUDED.context, model_id = EXCLUDED.model_id;"
	deleteChatContextQuery      = "UPDATE %s.%s SET context = '[{\"role\": \"system\", \"content\": \"\"}]' WHERE chat_id = $1;"
	updateGPTModelQuery         = "UPDATE %s.%s SET model_id = $2 WHERE chat_id = $1;"
)

type postgresStorage struct {
	db DBPool
}

func NewPostgresStorage(db DBPool) PostgresStorage {
	return &postgresStorage{
		db: db,
	}
}

func (s *postgresStorage) GetChatContext(ctx context.Context, chatID int) (string, []openai.Message, error) {
	var modelID string
	var contextJSON string

	err := s.db.QueryRow(ctx, fmt.Sprintf(getChatContextQuery, schema, chatContextTable), chatID).Scan(&modelID, &contextJSON)
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", nil, nil
		}
		return "", nil, err
	}

	var messages []openai.Message
	err = json.Unmarshal([]byte(contextJSON), &messages)
	if err != nil {
		return "", nil, err
	}

	return modelID, messages, nil
}

func (s *postgresStorage) UpdateChatContext(ctx context.Context, chatID int, messages []openai.Message, modelID string) error {
	contextJSON, err := json.Marshal(messages)
	if err != nil {
		return err
	}

	_, err = s.db.Exec(ctx, fmt.Sprintf(updateChatContextQuery, schema, chatContextTable), chatID, string(contextJSON), modelID)
	return err
}

func (s *postgresStorage) ClearChatContext(ctx context.Context, chatID int) error {
	_, err := s.db.Exec(ctx, fmt.Sprintf(deleteChatContextQuery, schema, chatContextTable), chatID)
	return err
}

func (s *postgresStorage) UpdateChatModel(ctx context.Context, chatID int, modelID string) error {
	_, err := s.db.Exec(ctx, fmt.Sprintf(updateGPTModelQuery, schema, chatContextTable), chatID, modelID)
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
		return fmt.Errorf("failed to create table %s: %w", chatUpdatesTable, err)
	}

	return nil
}
