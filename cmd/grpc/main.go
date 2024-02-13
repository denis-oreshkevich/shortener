package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/denis-oreshkevich/shortener/internal/app/config"
	"github.com/denis-oreshkevich/shortener/internal/app/model"
	"github.com/denis-oreshkevich/shortener/internal/app/server"
	pb "github.com/denis-oreshkevich/shortener/internal/app/server/proto"
	"github.com/denis-oreshkevich/shortener/internal/app/shortener"
	"github.com/denis-oreshkevich/shortener/internal/app/storage"
	"github.com/denis-oreshkevich/shortener/internal/app/util/logger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
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

	srv := grpc.NewServer()
	wg.Add(1)
	go func() {
		defer close(delChannel)
		defer wg.Done()
		<-ctx.Done()
		srv.GracefulStop()
	}()

	lis, err := net.Listen("tcp", ":3200")
	if err != nil {
		return fmt.Errorf("net.Listen: %w", err)
	}

	gs := server.NewGRPCServer(sh, conf, delChannel)

	reflection.Register(srv)

	pb.RegisterShortenerServer(srv, gs)
	if err = srv.Serve(lis); err != nil {
		if errors.Is(err, grpc.ErrServerStopped) {
			logger.Log.Info("Server closed")
		}
		return fmt.Errorf("router run %w", err)
	}

	wg.Wait()
	logger.Log.Info("Server Shutdown gracefully")
	return nil
}
