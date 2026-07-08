package config

import (
	"log"
	"testing"
)

func TestEnv(t *testing.T) {
	t.Run("should get DATABASE_URL from env", func(t *testing.T) {
		cfg, err := Env()

		if err != nil {
			log.Fatalf("Configuration failed: %v", err)
		}

		if cfg.DatabaseURL == "" {
			t.Errorf("expected DatabaseURL to not be empty")
		}
	})
}
