package config

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Logger     LoggerConf
	Server     ServerConf
	Storage    StorageConf
	Migrations MigrationsConf
	App        AppConf
	Redis      RedisConf
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
	CacheTTL      time.Duration
	LoginLimit    int
	PasswordLimit int
	IPLimit       int
	Window        int
}

type RedisConf struct {
	Address  string
	Password string
	DB       int
}

func LoadConfig(configPath string) (*Config, error) {
	viper.SetEnvPrefix("ABF")
	viper.AllowEmptyEnv(true)
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	bindEnvVariables()

	setDefaults()

	if configPath != "" {
		viper.SetConfigFile(configPath)
		if err := viper.ReadInConfig(); err != nil {
			var configFileNotFoundError viper.ConfigFileNotFoundError
			if !errors.As(err, &configFileNotFoundError) {
				return nil, fmt.Errorf("failed to read config: %w", err)
			}
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to decode config: %w", err)
	}

	return &config, nil
}

func bindEnvVariables() {
	viper.BindEnv("Server.Host", "ABF_SERVER_HOST")
	viper.BindEnv("Server.Port", "ABF_SERVER_PORT")

	viper.BindEnv("Storage.Dsn", "ABF_DATABASE_URL")
	viper.BindEnv("Migrations.Dsn", "ABF_DATABASE_URL")

	viper.BindEnv("Redis.Address", "ABF_REDIS_ADDRESS")
	viper.BindEnv("Redis.Password", "ABF_REDIS_PASSWORD")
	viper.BindEnv("Redis.DB", "ABF_REDIS_DB")

	viper.BindEnv("App.CacheTTL", "ABF_CACHE_TTL")
	viper.BindEnv("App.LoginLimit", "ABF_LOGIN_LIMIT")
	viper.BindEnv("App.PasswordLimit", "ABF_PASSWORD_LIMIT")
	viper.BindEnv("App.IPLimit", "ABF_IP_LIMIT")
	viper.BindEnv("App.Window", "ABF_WINDOW_SECONDS")

	viper.BindEnv("Logger.Level", "ABF_LOGGER_LEVEL")
	viper.BindEnv("Logger.FileName", "ABF_LOGGER_FILENAME")

	viper.BindEnv("Migrations.AutoMigrate", "ABF_MIGRATIONS_AUTOMIGRATE")
	viper.BindEnv("Migrations.Dir", "ABF_MIGRATIONS_DIR")
}

func setDefaults() {
	viper.SetDefault("Server.Host", "127.0.0.1")
	viper.SetDefault("Server.Port", "8080")
	viper.SetDefault("Redis.Address", "localhost:6379")
	viper.SetDefault("Redis.DB", 0)
	viper.SetDefault("App.LoginLimit", 10)
	viper.SetDefault("App.PasswordLimit", 100)
	viper.SetDefault("App.IPLimit", 1000)
	viper.SetDefault("App.Window", 60)
	viper.SetDefault("App.CacheTTL", "10s")
	viper.SetDefault("Logger.Level", "INFO")
	viper.SetDefault("Logger.FileName", "logs/app.log")
	viper.SetDefault("Migrations.AutoMigrate", true)
	viper.SetDefault("Migrations.Dir", "migrations")
}
