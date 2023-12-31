package main

import (
	"context"
	"fmt"
	"os"

	"github.com/denis-oreshkevich/shortener/internal/app/config"
	"github.com/denis-oreshkevich/shortener/internal/app/model"
	"github.com/denis-oreshkevich/shortener/internal/app/server"
	"github.com/denis-oreshkevich/shortener/internal/app/shortener"
	"github.com/denis-oreshkevich/shortener/internal/app/storage"
	"github.com/denis-oreshkevich/shortener/internal/app/util/logger"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
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
	ctx := context.Background()

	var s storage.Storage
	if conf.DatabaseDSN() != "" {
		dbStorage, err := storage.NewDBStorage(conf.DatabaseDSN())
		if err != nil {
			return fmt.Errorf("initializing db storage %w", err)
		}
		defer dbStorage.Close()
		s = dbStorage
		logger.Log.Info("using dbStorage as storage")
	} else if conf.FsPath() != "" {
		fileStorage, err := storage.NewFileStorage(conf.FsPath())
		if err != nil {
			return fmt.Errorf("initializing file storage %w", err)
		}
		defer fileStorage.Close()
		s = fileStorage
		logger.Log.Info("using fileStorage as storage")
	} else {
		mapStorage := storage.NewMapStorage()
		s = mapStorage
		logger.Log.Info("using mapStorage as storage")
	}

	delChannel := make(chan model.BatchDeleteEntry, 3)
	sh := shortener.New(s)

	//delete worker
	go func() {
		sh.DeleteUserURLs(ctx, delChannel)
	}()

	uh := server.New(conf, sh, delChannel)
	r := setUpRouter(conf, uh)

	err := r.Run(fmt.Sprintf("%s:%s", conf.Host(), conf.Port()))
	if err != nil {
		return fmt.Errorf("router run %w", err)
	}
	return nil
}

func setUpRouter(conf config.Conf, uh *server.Server) *gin.Engine {
	r := gin.New()
	pprof.Register(r)

	r.Use(gin.Recovery(), server.JWTAuth, server.Gzip, server.Logging)

	r.POST(`/`, uh.Post)
	r.GET(conf.BasePath()+`/:id`, uh.Get)
	r.GET(`/api/user/urls`, uh.GetUsersURLs)
	r.POST(`/api/shorten`, uh.ShortenPost)
	r.POST(`/api/shorten/batch`, uh.ShortenBatch)
	r.GET(`/ping`, uh.Ping)
	r.DELETE(`/api/user/urls`, uh.DeleteURLs)
	r.NoRoute(uh.NoRoute)

	return r
}
