package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/devxdh/concurrent-queue-engine/config"
	"github.com/devxdh/concurrent-queue-engine/pkg/db"
	"github.com/devxdh/concurrent-queue-engine/pkg/worker"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	pool := startup()
	defer pool.Close()

	taskStore := db.NewTaskStore(pool)

	workerPool := worker.NewPool(5, 50)
	workerPool.Start()

	daemonCtx, daemonCancel := context.WithCancel(context.Background())

	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-daemonCtx.Done():
				return
			case <-ticker.C:
				for range 5 {
					workerPool.Submit(worker.JobPayload{
						Store: taskStore,
					})
				}
			}
		}
	}()

	shutdownSig := make(chan os.Signal, 1)
	signal.Notify(shutdownSig, os.Interrupt, syscall.SIGTERM)

	<-shutdownSig

	fmt.Println("\n[DAEMON] Shutdown signal trapped. Initializing clean exit...")

	// 5. Execute Graceful Unwinding Lifecycle Sequence
	daemonCancel()        // Instantly stops the database ticker loop from issuing new submits
	workerPool.Shutdown() // Closes channels, drains work, and blocks until running workers finish

	fmt.Println("[DAEMON] Engine stopped cleanly.")
}

func startup() *pgxpool.Pool {
	fmt.Println("[DAEMON] Booting Task Queue Engine...")

	cfg, err := config.Env()
	if err != nil {
		log.Fatalf("Configuration Boot Failed: %v", err)
	}

	pool, err := db.Init(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Database Connection Failed: %v", err)
	}

	db.InjectDDL(pool)

	fmt.Println("[DAEMON] Boot sequence completed successfully. DB is live.")

	return pool
}
