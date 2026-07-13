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
	cancel      context.CancelFunc
}

func NewPool(workerCount, queueCapacity int) *Pool {
	ctx, cancel := context.WithCancel(context.Background())

	return &Pool{
		taskQueue:   make(chan JobPayload, queueCapacity),
		workerCount: workerCount,
		ctx:         ctx,
		cancel:      cancel,
	}
}

func (p *Pool) Shutdown() {
	p.cancel()
	close(p.taskQueue)
	p.wg.Wait()
	fmt.Println("[POOL] All background worker goroutines terminated")
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

		fmt.Printf("[WORKER #%d] Claimed Task ID %d - Validating URL: %s\n", workerID, task.ID, task.URL)

		ioCtx, ioCancel := context.WithTimeout(context.Background(), 5*time.Second)

		err = checkConnectivity(ioCtx, client, task.URL)
		ioCancel()

		if err != nil {
			fmt.Printf("[WORKER #%d] Task ID %d Verification Failed: %v\n", workerID, task.ID, err)

			updateErr := job.Store.MarkFailed(p.ctx, tx, task.ID, err.Error(), task.Attempts, task.MaxAttempts)
			if updateErr != nil {
				fmt.Printf("[CRITICAL] State sync failure for Task ID %d: %v\n", task.ID, updateErr)
			}
			continue
		}

		fmt.Printf("[WORKER #%d] Task ID %d Verification Succeeded\n", workerID, task.ID)
		if updateErr := job.Store.MarkComplete(p.ctx, tx, task.ID); updateErr != nil {
			fmt.Printf("[CRITICAL] State sync completion failure for Task ID %d: %v\n", task.ID, updateErr)
		}
	}
}

func checkConnectivity(ctx context.Context, clint *http.Client, url string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("request initialization failure: %w", err)
	}

	resp, err := clint.Do(req)
	if err != nil {
		return fmt.Errorf("network I/O boundary error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("target returned non-2xx status code: %d", resp.StatusCode)
	}

	return nil
}

func (p *Pool) Start() {
	for i := 1; i <= p.workerCount; i++ {
		p.wg.Add(1)
		go p.worker(i)
	}
}

func (p *Pool) Submit(job JobPayload) {
	select {
	case p.taskQueue <- job:
	case <-p.ctx.Done():
	}
}
