package config

import "context"

type Config struct {
	DBPath      string
	LogFilePath string
}

var cfg *Config

// TODO: Load configuration from file or environment variables
func LoadConfig() *Config {
	return &Config{
		DBPath:      "./local_rag.db",
		LogFilePath: "app.log",
	}
}

func GetConfig(_ctx context.Context) *Config {
	if cfg == nil {
		cfg = LoadConfig()
	}
	return cfg
}
