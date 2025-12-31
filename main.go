package main

import (
	"context"
	"io"
	"log/slog"
	"os"

	"github.com/MaxIvanyshen/local-rag/config"
	"github.com/MaxIvanyshen/local-rag/db"

	_ "github.com/mattn/go-sqlite3"
)

func setupLogging(file *os.File) {
	multi := io.MultiWriter(os.Stdout, file)
	handler := slog.NewTextHandler(multi, nil)
	logger := slog.New(handler)
	slog.SetDefault(logger)
}

func main() {
	ctx := context.Background()
	cfg := config.GetConfig(ctx)

	db := db.Init(cfg)
	sqlDB, err := db.DB()
	if err != nil {
		slog.Error("failed to get sql.DB", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer sqlDB.Close()

	logFile, err := os.OpenFile(cfg.LogFilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		slog.Error("failed to open log file", slog.String("error", err.Error()))
		logFile = nil
	}
	defer logFile.Close()

	setupLogging(logFile)

	var vecVersion string
	err = sqlDB.QueryRow("select vec_version()").Scan(&vecVersion)
	if err != nil {
		slog.Error("failed to get vec extension version", slog.String("error", err.Error()))
		os.Exit(1)
	}
	slog.Info("vec extension version", slog.String("version", vecVersion))

	slog.Info("application started successfully")
}
