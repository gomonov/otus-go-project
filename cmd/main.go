package main

import (
	"context"
	"flag"
	"log"
	"os/signal"
	"syscall"
	"time"

	"github.com/gomonov/otus-go-project/internal/app"
	"github.com/gomonov/otus-go-project/internal/config"
	"github.com/gomonov/otus-go-project/internal/logger"
	migrations "github.com/gomonov/otus-go-project/internal/migration"
	"github.com/gomonov/otus-go-project/internal/server"
	"github.com/gomonov/otus-go-project/internal/storage"
	"github.com/gomonov/otus-go-project/internal/storage/sqlstorage"
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

	var store storage.Storage

	store, err = sqlstorage.NewStorage(cfg.Storage.Dsn)
	if err != nil {
		log.Fatal("Failed to connect to PostgreSQL:", err)
	}

	defer func() {
		if err := store.Close(); err != nil {
			logg.Error("Failed to close storage: " + err.Error())
		}
	}()

	application := app.New(logg, store, cfg.App.CacheTTL)

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
