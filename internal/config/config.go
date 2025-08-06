package config

import (
	"log"
	"os"
	"strconv"
)

// Config holds all configuration for the application
type Config struct {
	Server ServerConfig
	Story  StoryConfig
}

// ServerConfig holds server-specific configuration
type ServerConfig struct {
	Port         string
	Host         string
	ReadTimeout  int
	WriteTimeout int
	IdleTimeout  int
}

// StoryConfig holds story-specific configuration
type StoryConfig struct {
	DataFile    string
	StaticDir   string
	TemplateDir string
}

// Load reads configuration from environment variables with sensible defaults
func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port:         getEnv("PORT", "8000"),
			Host:         getEnv("HOST", "localhost"),
			ReadTimeout:  getEnvAsInt("READ_TIMEOUT", 15),
			WriteTimeout: getEnvAsInt("WRITE_TIMEOUT", 15),
			IdleTimeout:  getEnvAsInt("IDLE_TIMEOUT", 60),
		},
		Story: StoryConfig{
			DataFile:    getEnv("STORY_DATA_FILE", "gopher.json"),
			StaticDir:   getEnv("STATIC_DIR", "./static"),
			TemplateDir: getEnv("TEMPLATE_DIR", "./templates"),
		},
	}
}

// Address returns the full server address
func (c *Config) Address() string {
	return c.Server.Host + ":" + c.Server.Port
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt gets an environment variable as integer or returns a default value
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
		log.Printf("Warning: Invalid integer value for %s: %s, using default %d", key, value, defaultValue)
	}
	return defaultValue
}
