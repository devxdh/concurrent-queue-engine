// Package worker
package worker

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/devxdh/concurrent-queue-engine/pkg/db"
)

type JobPayload struct {
	Store *db.TaskStore
}

type Pool struct {
	taskQueue   chan JobPayload
	workerCount int
	wg          sync.WaitGroup
	ctx         context.Context
	cancle      context.CancelFunc
}

func NewPool(workerCount, queueCapacity int) *Pool {
	ctx, cancle := context.WithCancel(context.Background())

	return &Pool{
		taskQueue:   make(chan JobPayload, queueCapacity),
		workerCount: workerCount,
		ctx:         ctx,
		cancle:      cancle,
	}
}

func (p *Pool) worker(workerID int) {
	defer p.wg.Done()

	client := &http.Client{
		Timeout: 7 * time.Second,
	}

	for job := range p.taskQueue {
		tx, task, err := job.Store.FetchNextTask(p.ctx)
		if err != nil {
			fmt.Printf("[WORKER #%d] Error acquiring database row: %v\n", workerID, err)
			continue
		}

		if task == nil {
			continue
		}

		fmt.Printf("[WORKER %d] Claimed Task ID %d - Validating URL: %s\n", workerID, task.ID, task.URL)

		ioCtx, ioCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer ioCancel()

	}
}
