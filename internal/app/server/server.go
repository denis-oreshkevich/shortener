package server

import (
	"github.com/denis-oreshkevich/shortener/internal/app/constant"
	"github.com/denis-oreshkevich/shortener/internal/app/handler"
	"net/http"
)

func InitServer() {
	mux := http.NewServeMux()
	mux.HandleFunc(`/`, handler.Handle)

	err := http.ListenAndServe(constant.ServerHost+":"+constant.ServerPort, mux)

	if err != nil {
		panic(err)
	}
}
