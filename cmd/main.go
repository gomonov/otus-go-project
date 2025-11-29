package main

import (
	"context"
	"flag"
	"os/signal"
	"syscall"
	"time"

	"github.com/gomonov/otus-go-project/internal/app"
	"github.com/gomonov/otus-go-project/internal/config"
	"github.com/gomonov/otus-go-project/internal/logger"
	migrations "github.com/gomonov/otus-go-project/internal/migration"
	"github.com/gomonov/otus-go-project/internal/ratelimit"
	"github.com/gomonov/otus-go-project/internal/server"
	"github.com/gomonov/otus-go-project/internal/storage/sqlstorage"
	"github.com/redis/go-redis/v9"
)

var configFile string

func init() {
	flag.StringVar(&configFile, "config", "/etc/application/config.toml", "Path to configuration file")
}

func main() {
	flag.Parse()

	if flag.Arg(0) == "version" {
		printVersion()
		return
	}

	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		panic(err)
	}

	logg, err := logger.New(logger.Conf(cfg.Logger))
	if err != nil {
		panic(err)
	}

	if err = migrations.AutoMigrate(logg, migrations.Conf(cfg.Migrations)); err != nil {
		panic(err)
	}

	store, err := sqlstorage.NewStorage(cfg.Storage.Dsn)
	if err != nil {
		panic(err)
	}
	defer store.Close()

	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Address,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		store.Close()
		panic(err)
	}
	defer redisClient.Close()

	rateLimiter := ratelimit.NewRateLimiter(redisClient, ratelimit.Config{
		LoginLimit:    cfg.App.LoginLimit,
		PasswordLimit: cfg.App.PasswordLimit,
		IPLimit:       cfg.App.IPLimit,
		Window:        cfg.App.Window,
	})

	application := app.New(logg, store, cfg.App.CacheTTL, rateLimiter)

	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	context.AfterFunc(ctx, func() {
		logg.Info("application is stopping...")
	})

	httpServer := server.NewServer(logg, application, server.Conf(cfg.Server))
	go func() {
		logg.Info("HTTP server starting...")
		if err := httpServer.Start(ctx); err != nil {
			logg.Error("failed to start HTTP server: " + err.Error())
			cancel()
		}
	}()

	defer func() {
		stopCtx, stopCancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer stopCancel()

		logg.Info("Shutting down HTTP server...")
		if err := httpServer.Stop(stopCtx); err != nil {
			logg.Error("Failed to stop HTTP server: " + err.Error())
		} else {
			logg.Info("HTTP server stopped successfully")
		}
	}()

	logg.Info("application is running...")

	<-ctx.Done()
	logg.Info("Application stopped")
}
