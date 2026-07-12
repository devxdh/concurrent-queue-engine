package db

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Task struct {
	ID          int64
	URL         string
	Status      string
	Attempts    int
	MaxAttempts int
	LastError   *string
}

type TaskStore struct {
	pool *pgxpool.Pool
}

func NewTaskStore(pool *pgxpool.Pool) *TaskStore {
	return &TaskStore{pool: pool}
}

func (ts *TaskStore) FetchNextTask(ctx context.Context) (pgx.Tx, *Task, error) {
	tx, err := ts.pool.Begin(ctx)
	if err != nil {
		return nil, nil, err
	}

	query := `
		SELECT id, url, status, attempts, max_attempts
		FROM tasks
		WHERE status = 'pending' AND attempts < max_attempts
		ORDER BY id ASC
		LIMIT 1
		FOR UPDATE SKIP LOCKED;
	`

	var task Task
	err = tx.QueryRow(ctx, query).Scan(
		&task.ID,
		&task.URL,
		&task.Status,
		&task.Attempts,
		&task.MaxAttempts,
	)

	if err != nil {
		_ = tx.Rollback(ctx)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil, nil
		}
		return nil, nil, err
	}

	updateQuery := `
		UPDATE tasks
		SET status = 'processing', updated_at = $1
		WHERE id = $2;
	`

	_, err = tx.Exec(ctx, updateQuery, time.Now(), task.ID)
	if err != nil {
		return nil, nil, err
	}

	return tx, &task, nil
}
