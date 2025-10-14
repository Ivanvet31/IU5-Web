package config

import (
	"strings"
	"time"

	"github.com/spf13/viper"
)

type JWTConfig struct {
	Secret    string
	ExpiresIn time.Duration
}

type RedisConfig struct {
	Host     string
	Port     int
	Password string
	User     string
}

type Config struct {
	ServiceHost string
	ServicePort int
	JWT         JWTConfig
	Redis       RedisConfig
}

func NewConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.AddConfigPath("config")
	viper.AddConfigPath(".")

	// --- НАСТРОЙКА VIPER ДЛЯ ЧТЕНИЯ ИЗ .ENV ---
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv() // Автоматически читать переменные окружения

	if err := viper.ReadInConfig(); err != nil {
		// Игнорируем ошибку, если файл config.toml не найден, т.к. .env может быть достаточно
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	cfg := &Config{}
	if err := viper.Unmarshal(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
