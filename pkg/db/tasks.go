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

func (ts *TaskStore) MarkComplete(ctx context.Context, tx pgx.Tx, taskID int64) error {
	defer tx.Rollback(ctx)

	query := `
		UPDATE tasks
		SET status = 'completed', updated_at = $1 
		WHERE id = $2;
	`

	_, err := tx.Exec(ctx, query, time.Now(), taskID)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (ts *TaskStore) MarkFailed(
	ctx context.Context,
	tx pgx.Tx,
	taskID int64,
	errMsg string,
	attempts int,
	maxAttempts int,
) error {
	defer tx.Rollback(ctx)

	nextStatus := "pending"
	if attempts+1 >= maxAttempts {
		nextStatus = "failed"
	}

	query := `
		UPDATE tasks
		SET 
			status = $1,
			attempts = attempts + 1,
			last_error = $2,
			updated_at = $3
		WHERE id = $4;
	`

	_, err := tx.Exec(ctx, query, nextStatus, errMsg, time.Now(), taskID)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}
