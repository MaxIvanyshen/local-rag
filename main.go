package main

import (
	"io"
	"log/slog"
	"os"

	"local_rag/config"
	"local_rag/db"

	_ "github.com/mattn/go-sqlite3"
)

func setupLogging(cfg *config.Config) {
	file, err := os.OpenFile(cfg.LogFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		slog.Error("Failed to open log file", "error", err)
		os.Exit(1)
	}
	defer file.Close()

	multi := io.MultiWriter(os.Stdout, file)
	handler := slog.NewTextHandler(multi, nil)
	logger := slog.New(handler)
	slog.SetDefault(logger)
}

func main() {
	cfg := config.LoadConfig()
	db := db.Init(cfg)
	defer db.Close()

	setupLogging(cfg)

	var vecVersion string
	err := db.QueryRow("select vec_version()").Scan(&vecVersion)
	if err != nil {
		slog.Error("Failed to query vec_version", "error", err)
		os.Exit(1)
	}
	slog.Info("Vec version retrieved", "vec_version", vecVersion)
}
