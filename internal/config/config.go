package config

import "os"

type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	RedisAddr  string
	JWTKey     string
	ServerPort string
}

func Load() (*Config, error) {
	return &Config{
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5433"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "password"),
		DBName:     getEnv("DB_NAME", "forum_db"),
		RedisAddr:  getEnv("REDIS_ADDR", "localhost:6379"),
		JWTKey:     getEnv("JWT_KEY", "conch2147"),
		ServerPort: getEnv("SERVER_PORT", "8080"),
	}, nil
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
