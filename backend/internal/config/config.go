package config

import "os"

type Config struct {
	HTTPAddr string
	GRPCAddr string
	DBPath   string
	APIKey   string
}

func Load() Config {
	return Config{
		HTTPAddr: getenv("HTTP_ADDR", ":8081"),
		GRPCAddr: getenv("GRPC_ADDR", ":9091"),
		DBPath:   getenv("DB_PATH", "app.db"),
		APIKey:   getenv("API_KEY", "dev-key"),
	}
}

func getenv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
