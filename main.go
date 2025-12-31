package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"

	"github.com/MaxIvanyshen/local-rag/chunker"
	"github.com/MaxIvanyshen/local-rag/config"
	"github.com/MaxIvanyshen/local-rag/db"
	"github.com/MaxIvanyshen/local-rag/embedding"
	"github.com/MaxIvanyshen/local-rag/service"

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

	if cfg.Logging.LogToFile {
		logFile, err := os.OpenFile(cfg.Logging.LogFilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			slog.Error("failed to open log file", slog.String("error", err.Error()))
			logFile = nil
		}
		defer logFile.Close()

		setupLogging(logFile)
	}

	var vecVersion string
	err = sqlDB.QueryRow("select vec_version()").Scan(&vecVersion)
	if err != nil {
		slog.Error("failed to get vec extension version. sqlite-vec might not work", slog.String("error", err.Error()))
		os.Exit(1)
	}
	slog.Info("vec extension version", slog.String("version", vecVersion))

	slog.Info("application started successfully")

	mux := http.NewServeMux()

	// handle health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	contentChunker := chunker.NewParagraphChunker(cfg.Chunker.OverlapBytes)

	s := service.NewService(&service.ServiceParameters{
		DB:       db,
		Embedder: embedding.NewOllamaEmbedder(cfg.Ollama.Model),
		Chunker:  contentChunker,
		Cfg:      cfg,
	})
	s.RegisterRoutes(mux)

	port := cfg.Port

	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	slog.Info("starting HTTP server", slog.Int("port", port))
	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("HTTP server error", slog.String("error", err.Error()))
	}
}
