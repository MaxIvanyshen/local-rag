package config

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Port int `yaml:"port" env:"LOCAL_RAG_PORT" env-default:"8080"`

	DBPath string `yaml:"db_path" env:"DB_PATH" env-default:"./local_rag.db"`

	Search   SearchConfig   `yaml:"search"`
	Embedder EmbedderConfig `yaml:"embedder"`

	Logging LoggingConfig `yaml:"logging"`

	Chunker ChunkerConfig `yaml:"chunker"`

	BatchProcessing BatchProcessingConfig `yaml:"batch_processing"`

	CLI CLIConfig `yaml:"cli"`
}

type CLIConfig struct {
	Host string `yaml:"host" env:"CLI_HOST" env-default:"http://localhost"`
}

type BatchProcessingConfig struct {
	WorkerCount int `yaml:"worker_count" env:"BATCH_WORKER_COUNT" env-default:"4"`
}

type ChunkerConfig struct {
	Type         string `yaml:"type" env:"CHUNKER_TYPE" env-default:"paragraph"`
	OverlapBytes int    `yaml:"overlap_bytes" env:"CHUNKER_OVERLAP_BYTES" env-default:"0"`
	ChunkSize    int    `yaml:"chunk_size" env:"CHUNKER_CHUNK_SIZE" env-default:"1000"`
}

type LoggingConfig struct {
	LogToFile   bool   `yaml:"log_to_file" env:"LOG_TO_FILE" env-default:"true"`
	LogFilePath string `yaml:"log_file_path" env:"LOG_FILE_PATH" env-default:"./local_rag.log"`
}

type SearchConfig struct {
	TopK int `yaml:"top_k" env:"SEARCH_TOP_K" env-default:"5"`
}

type EmbedderConfig struct {
	Type    string `yaml:"type" env:"EMBEDDER_TYPE" env-default:"ollama"`
	BaseURL string `yaml:"base_url" env:"EMBEDDER_BASE_URL" env-default:"http://localhost:11434"`
	Model   string `yaml:"model" env:"EMBEDDER_MODEL" env-default:"nomic-embed-text"`
}

var cfg *Config

func LoadConfig() *Config {
	cfg := &Config{}

	cleanenv.ReadEnv(cfg)

	home, err := os.UserHomeDir()
	if err != nil {
		slog.Error("failed to get user home directory", slog.String("error", err.Error()))
		return cfg
	}
	configPath := filepath.Join(home, ".config", "local_rag", "config.yml")

	err = cleanenv.ReadConfig(configPath, cfg)
	if err != nil {
		if strings.Contains(err.Error(), "no such file or directory") {
			slog.Info("config file not found, using defaults and environment variables")
		} else {
			slog.Error("failed to read config file", slog.String("error", err.Error()))
		}
	}
	return cfg
}

func GetConfig(_ctx context.Context) *Config {
	if cfg == nil {
		cfg = LoadConfig()
	}
	return cfg
}
