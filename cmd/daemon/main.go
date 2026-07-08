package daemon

import (
	"fmt"
	"log"

	"github.com/devxdh/concurrent-queue-engine/config"
	"github.com/devxdh/concurrent-queue-engine/pkg/db"
)

func main() {
	fmt.Println("[DAEMON] Booting Task Queue Engine...")

	cfg, err := config.Env()
	if err != nil {
		log.Fatalf("Configuration Boot Failed: %v", err)
	}

	pool, err := db.Init(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Database Connection Failed: %v", err)
	}
	defer pool.Close()

	db.InjectDDL(pool)

	fmt.Println("[DAEMON] Boot sequence completed successfully. DB is live.")
}
