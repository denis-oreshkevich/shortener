package server

import (
	"github.com/denis-oreshkevich/shortener/internal/app/constant"
	"github.com/denis-oreshkevich/shortener/internal/app/handler"
	"github.com/denis-oreshkevich/shortener/internal/app/repo"
	"net/http"
)

var repository = repo.Get()

func InitServer() {
	mux := http.NewServeMux()
	mux.HandleFunc(`/`, handler.URL(repository))

	err := http.ListenAndServe(constant.ServerHost+":"+constant.ServerPort, mux)

	if err != nil {
		panic(err)
	}
}
