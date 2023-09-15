package main

import (
	"flag"
	"github.com/denis-oreshkevich/shortener/internal/app/server"
)

func main() {
	flag.Parse()
	server.InitServer()
}
