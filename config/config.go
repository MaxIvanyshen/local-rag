package config

type Config struct {
	DBPath      string
	LogFilePath string
}

// TODO: Load configuration from file or environment variables
func LoadConfig() *Config {
	return &Config{
		DBPath:      "./local_rag.db",
		LogFilePath: "app.log",
	}
}
