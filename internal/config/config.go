package config

import "os"

type Config struct {
	DBHost             string
	DBPort             string
	DBUser             string
	DBPassword         string
	DBName             string
	RedisAddr          string
	RedisPassword      string
	JWTKey             string
	ServerPort         string
	OpenSearchURL      string
	OpenSearchUser     string
	OpenSearchPassword string
}

func Load() (*Config, error) {
	return &Config{
		DBHost:             getEnv("DB_HOST", "localhost"),
		DBPort:             getEnv("DB_PORT", "5433"),
		DBUser:             getEnv("DB_USER", "postgres"),
		DBPassword:         getEnv("DB_PASSWORD", "password"),
		DBName:             getEnv("DB_NAME", "forum_db"),
		RedisAddr:          getEnv("REDIS_ADDR", "127.0.0.1:6380"),
		RedisPassword:      getEnv("REDIS_PASSWORD", ""),
		JWTKey:             getEnv("JWT_KEY", "conch2147"),
		ServerPort:         getEnv("SERVER_PORT", "8080"),
		OpenSearchURL:      getEnv("OPENSEARCH_URL", ""),
		OpenSearchUser:     getEnv("OPENSEARCH_USER", ""),
		OpenSearchPassword: getEnv("OPENSEARCH_PASSWORD", ""),
	}, nil
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
