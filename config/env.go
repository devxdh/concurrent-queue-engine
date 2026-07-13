package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type EnvConfig struct {
	DatabaseURL string
}

func Env() (EnvConfig, error) {
	_ = godotenv.Load(".env")

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return EnvConfig{}, fmt.Errorf("DATABASE_URL env is missing or empty")
	}

	return EnvConfig{
		DatabaseURL: dbURL,
	}, nil
}
