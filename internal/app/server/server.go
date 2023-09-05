package server

import (
	"fmt"
	"github.com/denis-oreshkevich/shortener/internal/app/constant"
	"github.com/denis-oreshkevich/shortener/internal/app/handler"
	"github.com/denis-oreshkevich/shortener/internal/app/repo"
)

var repository = repo.New()

func Init() {
	r := handler.SetupRouter(repository)

	host := fmt.Sprintf("%s:%s", constant.ServerHost, constant.ServerPort)
	err := r.Run(host)

	if err != nil {
		panic(err)
	}
}
