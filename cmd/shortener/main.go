package main

import (
	"context"
	"errors"
	"fmt"
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
	"log"
	"net/http"
	"os/signal"
	"sync"
	"syscall"
)

var buildVersion = "N/A"

var buildDate = "N/A"

var buildCommit = "N/A"

func main() {
	err := logger.Initialize(zapcore.DebugLevel.String())
	if err != nil {
		log.Fatal("logger.Initialize", err)
	}
	defer logger.Log.Sync()

	logger.Log.Info(fmt.Sprintf("Build version: %s\n", buildVersion))
	logger.Log.Info(fmt.Sprintf("Build date: %s\n", buildDate))
	logger.Log.Info(fmt.Sprintf("Build commit: %s\n", buildCommit))

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
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	defer stop()

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

	var wg sync.WaitGroup

	//delete worker
	wg.Add(1)
	go func() {
		defer wg.Done()
		sh.DeleteUserURLs(ctx, delChannel)
	}()

	uh := server.New(conf, sh, delChannel)
	r := setUpRouter(conf, uh)

	addr := fmt.Sprintf("%s:%s", conf.Host(), conf.Port())

	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	wg.Add(1)
	go func() {
		defer close(delChannel)
		defer wg.Done()
		<-ctx.Done()

		if err := srv.Shutdown(context.Background()); err != nil {
			logger.Log.Error("HTTP server Shutdown", zap.Error(err))
		}
	}()

	var err error
	if conf.EnableHTTPS() {
		manager, errHTTPS := server.NewCertManager("./certs/cert.pem", "./certs/key.pem")
		if errHTTPS != nil {
			return fmt.Errorf("server.NewCertManager: %w", errHTTPS)
		}
		err = srv.ListenAndServeTLS(manager.CertPath, manager.KeyPath)
	} else {
		err = srv.ListenAndServe()
	}

	if err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			logger.Log.Info("Server closed")
		} else {
			return fmt.Errorf("router run %w", err)
		}
	}

	wg.Wait()
	logger.Log.Info("Server Shutdown gracefully")
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
	r.GET(`/api/internal/stats`, uh.GetAPIInternalStats)
	r.NoRoute(uh.NoRoute)

	return r
}
