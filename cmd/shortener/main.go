package main

import (
	"fmt"
	"github.com/denis-oreshkevich/shortener/internal/app/config"
	"github.com/denis-oreshkevich/shortener/internal/app/handler"
	"github.com/denis-oreshkevich/shortener/internal/app/server"
	"github.com/denis-oreshkevich/shortener/internal/app/storage"
	"os"
)

func main() {
	err := config.Parse()
	if err != nil {
		fmt.Fprintln(os.Stderr, "parse config", err)
		os.Exit(1)
	}

	conf := config.Get()

	uh := handler.New(conf, storage.NewMapStorage())
	srv := server.New(conf, uh)

	err = srv.Start()
	if err != nil {
		fmt.Fprintln(os.Stderr, "start server", err)
		os.Exit(1)
	}
}
