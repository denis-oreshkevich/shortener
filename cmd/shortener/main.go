package main

import (
	"fmt"
	"github.com/denis-oreshkevich/shortener/internal/app/config"
	"github.com/denis-oreshkevich/shortener/internal/app/handler"
	"github.com/denis-oreshkevich/shortener/internal/app/server"
	"github.com/denis-oreshkevich/shortener/internal/app/storage"
	"github.com/denis-oreshkevich/shortener/internal/app/util/logger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
)

func main() {
	err := logger.Initialize(zapcore.DebugLevel.String())
	defer logger.Log.Sync()
	if err != nil {
		fmt.Fprintln(os.Stderr, "logger initialize", err)
		os.Exit(1)
	}
	err = config.Parse()
	if err != nil {
		logger.Log.Fatal("parse config", zap.Error(err))
	}

	if err := run(); err != nil {
		logger.Log.Fatal("main error", zap.Error(err))
	}
}

func run() error {
	conf := config.Get()

	//uh := handler.New(conf, storage.NewMapStorage(make(map[string]string)))
	fileStorage, err := storage.NewFileStorage(conf.FsPath())
	if err != nil {
		return fmt.Errorf("initializing file storage %w", err)
	}
	defer fileStorage.Close()

	uh := handler.New(conf, fileStorage)

	dbStorage, err := storage.NewDBStorage(conf.DbDSN())
	if err != nil {
		return fmt.Errorf("initializing db storage %w", err)
	}
	defer dbStorage.Close()

	ph := handler.NewPingHandler(dbStorage)
	srv := server.New(conf, uh)

	srv.AddPing(ph)

	err = srv.Start()
	if err != nil {
		return fmt.Errorf("start server %w", err)
	}
	return nil
}
