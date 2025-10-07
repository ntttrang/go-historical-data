package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	App      AppConfig      `mapstructure:"app"`
	Database DatabaseConfig `mapstructure:"database"`
	API      APIConfig      `mapstructure:"api"`
	Logging  LoggingConfig  `mapstructure:"logging"`
	CORS     CORSConfig     `mapstructure:"cors"`
}

type AppConfig struct {
	Env   string `mapstructure:"env"`
	Name  string `mapstructure:"name"`
	Port  int    `mapstructure:"port"`
	Debug bool   `mapstructure:"debug"`
}

type DatabaseConfig struct {
	Host            string `mapstructure:"host"`
	Port            int    `mapstructure:"port"`
	Name            string `mapstructure:"name"`
	User            string `mapstructure:"user"`
	Password        string `mapstructure:"password"`
	MaxOpenConns    int    `mapstructure:"max_open_conns"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`
	ConnMaxLifetime int    `mapstructure:"conn_max_lifetime"`
}

type APIConfig struct {
	RateLimit       int `mapstructure:"rate_limit"`
	RequestTimeout  int `mapstructure:"request_timeout"`
	ShutdownTimeout int `mapstructure:"shutdown_timeout"`
}

type LoggingConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

type CORSConfig struct {
	AllowedOrigins []string `mapstructure:"allowed_origins"`
	AllowedMethods []string `mapstructure:"allowed_methods"`
	AllowedHeaders []string `mapstructure:"allowed_headers"`
}

// Load loads configuration from file and environment variables
func Load() (*Config, error) {
	env := getEnv("APP_ENV", "dev")

	// Set config file name based on environment
	configName := fmt.Sprintf("config.%s", env)
	viper.SetConfigName(configName)
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")
	viper.AddConfigPath("../config")
	viper.AddConfigPath("../../config")

	// Enable environment variable overrides
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Read config file
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Unmarshal config
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Override with environment variables if present
	overrideFromEnv(&config)

	return &config, nil
}

func overrideFromEnv(cfg *Config) {
	if val := os.Getenv("APP_PORT"); val != "" {
		fmt.Sscanf(val, "%d", &cfg.App.Port)
	}
	if val := os.Getenv("DB_HOST"); val != "" {
		cfg.Database.Host = val
	}
	if val := os.Getenv("DB_PORT"); val != "" {
		fmt.Sscanf(val, "%d", &cfg.Database.Port)
	}
	if val := os.Getenv("DB_NAME"); val != "" {
		cfg.Database.Name = val
	}
	if val := os.Getenv("DB_USER"); val != "" {
		cfg.Database.User = val
	}
	if val := os.Getenv("DB_PASSWORD"); val != "" {
		cfg.Database.Password = val
	}
	if val := os.Getenv("LOG_LEVEL"); val != "" {
		cfg.Logging.Level = val
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
