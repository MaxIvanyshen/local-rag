package db

import (
	"embed"
	"local_rag/config"
	"log/slog"
	"time"

	sqlite_vec "github.com/asg017/sqlite-vec-go-bindings/cgo"
	"github.com/pressly/goose/v3"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

func Init(cfg *config.Config) *gorm.DB {
	sqlite_vec.Auto()
	db, err := gorm.Open(sqlite.Open(cfg.DBPath), &gorm.Config{})
	if err != nil {
		slog.Error("failed to open database", slog.String("error", err.Error()))
	}

	goose.SetBaseFS(embedMigrations)
	if err := goose.SetDialect("sqlite3"); err != nil {
		slog.Error("failed to set goose dialect", slog.String("error", err.Error()))
	}

	sqlDB, _ := db.DB()
	if err := goose.Up(sqlDB, "migrations"); err != nil {
		slog.Error("failed to run migrations", slog.String("error", err.Error()))
	}

	slog.Info("database initialized successfully")

	return db
}

type Document struct {
	ID        string    `gorm:"primaryKey"`
	Name      string    `gorm:"not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

type Chunk struct {
	ID         string `gorm:"primaryKey"`
	DocumentID string
	ChunkIndex int       `gorm:"not null"`
	Data       []byte    `gorm:"not null"`
	CreatedAt  time.Time `gorm:"autoCreateTime"`
}
