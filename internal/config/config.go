package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Logger     LoggerConf
	Server     ServerConf
	Storage    StorageConf
	Migrations MigrationsConf
	App        AppConf
}

type LoggerConf struct {
	Level    string
	FileName string
}

type ServerConf struct {
	Host string
	Port string
}

type StorageConf struct {
	Dsn string
}

type MigrationsConf struct {
	AutoMigrate bool
	Dir         string
	Dsn         string
}

type AppConf struct {
	CacheTTL time.Duration
}

func LoadConfig(configPath string) (*Config, error) {
	viper.SetConfigFile(configPath)

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to decode config: %w", err)
	}

	return &config, nil
}
