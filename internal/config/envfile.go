package config

import (
	"os"

	"github.com/joho/godotenv"
)

func loadEnvFiles() error {
	if fileExists(".env.local") {
		return godotenv.Load(".env.local")
	}
	if fileExists(".env") {
		return godotenv.Load(".env")
	}
	return nil
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}
