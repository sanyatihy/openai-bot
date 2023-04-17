package processor

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/sanyatihy/openai-go/pkg/openai"
)

const (
	createSchemaQuery            = "CREATE SCHEMA IF NOT EXISTS content;"
	createChatContextTableQuery  = "CREATE TABLE IF NOT EXISTS content.chat_context (chat_id BIGINT PRIMARY KEY, context JSONB);"
	createLastUpdateIDTableQuery = "CREATE TABLE IF NOT EXISTS content.last_update_id (id SERIAL PRIMARY KEY, update_id INT NOT NULL);"

	getChatContextQuery    = "SELECT context FROM content.chat_context WHERE chat_id = $1;"
	updateChatContextQuery = "INSERT INTO content.chat_context (chat_id, context) VALUES ($1, $2) ON CONFLICT (chat_id) DO UPDATE SET context = EXCLUDED.context;"
	deleteChatContextQuery = "DELETE FROM content.chat_context WHERE chat_id = $1;"

	getLastUpdateIDQuery  = "SELECT update_id FROM last_update_id WHERE id = 1;"
	saveLastUpdateIDQuery = "INSERT INTO content.last_update_id (id, update_id) VALUES (1, $1) ON CONFLICT (id) DO UPDATE SET update_id = $1;"
)

func (p *processor) getChatContext(ctx context.Context, chatID int) ([]openai.Message, error) {
	var contextJSON string

	err := p.db.QueryRow(ctx, getChatContextQuery, chatID).Scan(&contextJSON)
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

func (p *processor) updateChatContext(ctx context.Context, chatID int, messages []openai.Message) error {
	contextJSON, err := json.Marshal(messages)
	if err != nil {
		return err
	}

	_, err = p.db.Exec(ctx, updateChatContextQuery, chatID, string(contextJSON))
	return err
}

func (p *processor) clearChatContext(ctx context.Context, chatID int) error {
	_, err := p.db.Exec(ctx, deleteChatContextQuery, chatID)
	return err
}

func (p *processor) runInitialMigrations(ctx context.Context) error {
	_, err := p.db.Exec(ctx, createSchemaQuery)
	if err != nil {
		return fmt.Errorf("failed to create schema 'content': %w", err)
	}

	_, err = p.db.Exec(ctx, createChatContextTableQuery)
	if err != nil {
		return fmt.Errorf("failed to create table 'chat_context': %w", err)
	}

	_, err = p.db.Exec(ctx, createLastUpdateIDTableQuery)
	if err != nil {
		return fmt.Errorf("failed to create table 'last_update_id': %w", err)
	}

	return nil
}

func (p *processor) loadLastUpdateIDFromDB(ctx context.Context) (int, error) {
	var lastUpdateID int
	err := p.db.QueryRow(ctx, getLastUpdateIDQuery).Scan(&lastUpdateID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return 0, nil
		}
		return 0, err
	}

	return lastUpdateID, nil
}

func (p *processor) saveLastUpdateIDToDB(ctx context.Context, lastUpdateID int) error {
	_, err := p.db.Exec(ctx, saveLastUpdateIDQuery, lastUpdateID)
	return err
}
