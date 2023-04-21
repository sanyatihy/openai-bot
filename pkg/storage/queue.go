package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/sanyatihy/openai-bot/pkg/telegram"
	"time"
)

const (
	chatUpdatesTable = "chat_updates"

	UpdateStatusPending    = "pending"
	UpdateStatusProcessing = "processing"
	UpdateStatusProcessed  = "processed"
	UpdateStatusError      = "error"
)

const (
	createChatUpdatesTableQuery = "CREATE TABLE IF NOT EXISTS %s.%s (id SERIAL PRIMARY KEY, update_id INTEGER NOT NULL, chat_id INTEGER NOT NULL, update_data JSONB NOT NULL, status VARCHAR(20) NOT NULL, created_at TIMESTAMP NOT NULL);"
	insertChatUpdatesQuery      = "INSERT INTO %s.%s (update_id, chat_id, update_data, status, created_at) VALUES ($1, $2, $3, $4, $5);"
	getNextChatUpdateQuery      = "SELECT id, update_data FROM %s.%s WHERE status = '%s' AND chat_id NOT IN (SELECT chat_id FROM %s.%s WHERE status = '%s') ORDER BY update_id FOR UPDATE SKIP LOCKED LIMIT 1;"
	getLastChatUpdateQuery      = "SELECT update_data FROM %s.%s ORDER BY update_id DESC LIMIT 1;"
	setChatUpdateStatusQuery    = "UPDATE %s.%s SET status = $1 WHERE id = $2;"
	resetChatUpdatesStatusQuery = "UPDATE %s.%s SET status = '%s' WHERE status = '%s' AND (NOW() - created_at) > INTERVAL '60 seconds';"
)

type postgresQueue struct {
	db DBPool
}

func NewPostgresQueue(db DBPool) PostgresQueue {
	return &postgresQueue{
		db: db,
	}
}

func (q *postgresQueue) InsertChatUpdate(ctx context.Context, update telegram.Update) error {
	updateJSON, err := json.Marshal(update)
	if err != nil {
		return err
	}

	_, err = q.db.Exec(ctx, fmt.Sprintf(insertChatUpdatesQuery, schema, chatUpdatesTable), update.UpdateID, update.Message.Chat.ID, string(updateJSON), "pending", time.Now().UTC())
	return err
}

func (q *postgresQueue) GetNextChatUpdate(ctx context.Context, status string) (int, telegram.Update, error) {
	var updateID int
	var updateJSON string

	tx, err := q.db.Begin(ctx)
	if err != nil {
		return 0, telegram.Update{}, err
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx, fmt.Sprintf(getNextChatUpdateQuery, schema, chatUpdatesTable, UpdateStatusPending, schema, chatUpdatesTable, UpdateStatusProcessing)).Scan(&updateID, &updateJSON)
	if err != nil {
		if err == pgx.ErrNoRows {
			return 0, telegram.Update{}, nil
		}
		return 0, telegram.Update{}, err
	}

	var update telegram.Update
	err = json.Unmarshal([]byte(updateJSON), &update)
	if err != nil {
		return 0, telegram.Update{}, err
	}

	_, err = tx.Exec(ctx, fmt.Sprintf(setChatUpdateStatusQuery, schema, chatUpdatesTable), status, updateID)
	if err != nil {
		return 0, telegram.Update{}, err
	}

	return updateID, update, tx.Commit(ctx)
}

func (q *postgresQueue) GetLastChatUpdateID(ctx context.Context) (int, error) {
	var updateJSON string

	err := q.db.QueryRow(ctx, fmt.Sprintf(getLastChatUpdateQuery, schema, chatUpdatesTable)).Scan(&updateJSON)
	if err != nil {
		if err == pgx.ErrNoRows {
			return 0, nil
		}
		return 0, err
	}

	var update telegram.Update
	err = json.Unmarshal([]byte(updateJSON), &update)
	if err != nil {
		return 0, err
	}

	return update.UpdateID, nil
}

func (q *postgresQueue) SetChatUpdateStatus(ctx context.Context, updateID int, status string) error {
	_, err := q.db.Exec(ctx, fmt.Sprintf(setChatUpdateStatusQuery, schema, chatUpdatesTable), status, updateID)
	return err
}

func (q *postgresQueue) ResetChatUpdatesStatus(ctx context.Context) error {
	_, err := q.db.Exec(ctx, fmt.Sprintf(resetChatUpdatesStatusQuery, schema, chatUpdatesTable, UpdateStatusPending, UpdateStatusProcessing))
	return err
}
