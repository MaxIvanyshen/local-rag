package config

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/ilyakaznacheev/cleanenv"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Port int `yaml:"port" env:"LOCAL_RAG_PORT" env-default:"8080"`

	DBPath string `yaml:"db_path" env:"DB_PATH" env-default:"~/.local_rag/local_rag.db"`

	Search   SearchConfig   `yaml:"search"`
	Embedder EmbedderConfig `yaml:"embedder"`

	Logging LoggingConfig `yaml:"logging"`

	Chunker ChunkerConfig `yaml:"chunker"`

	BatchProcessing BatchProcessingConfig `yaml:"batch_processing"`

	Extensions ExtensionsConfig `yaml:"extensions"`
}

type ExtensionsConfig struct {
	Host string `yaml:"host" env:"LOCAL_RAG_HOST" env-default:"http://localhost"`
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
	LogFilePath string `yaml:"log_file_path" env:"LOG_FILE_PATH" env-default:"~/.local_rag/local_rag.log"`
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

func expandHome(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, path[2:])
	}
	return path
}

func LoadConfig() *Config {
	cfg := &Config{}

	cleanenv.ReadEnv(cfg)

	// Expand ~ in paths
	cfg.DBPath = expandHome(cfg.DBPath)
	cfg.Logging.LogFilePath = expandHome(cfg.Logging.LogFilePath)

	// Create directories for db and log if they don't exist
	dbDir := filepath.Dir(cfg.DBPath)
	if dbDir != "." {
		if err := os.MkdirAll(dbDir, 0755); err != nil {
			slog.Error("failed to create database directory", slog.String("error", err.Error()))
		}
	}
	logDir := filepath.Dir(cfg.Logging.LogFilePath)
	if logDir != "." && logDir != dbDir {
		if err := os.MkdirAll(logDir, 0755); err != nil {
			slog.Error("failed to create log directory", slog.String("error", err.Error()))
		}
	}

	home, err := os.UserHomeDir()
	if err != nil {
		slog.Error("failed to get user home directory", slog.String("error", err.Error()))
		return cfg
	}
	configPath := filepath.Join(home, ".config", "local_rag", "config.yml")

	// Check if config file exists, if not, create it with defaults
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Create the directory
		dir := filepath.Dir(configPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			slog.Error("failed to create config directory", slog.String("error", err.Error()))
		} else {
			// Marshal the config to YAML
			data, err := yaml.Marshal(cfg)
			if err != nil {
				slog.Error("failed to marshal config to YAML", slog.String("error", err.Error()))
			} else {
				// Write to file
				if err := os.WriteFile(configPath, data, 0644); err != nil {
					slog.Error("failed to write default config file", slog.String("error", err.Error()))
				} else {
					slog.Info("created default config file", slog.String("path", configPath))
				}
			}
		}
	}

	// Always attempt to read the config file to override defaults
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
