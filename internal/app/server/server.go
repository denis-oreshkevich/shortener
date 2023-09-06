package server

import (
	"fmt"
	"github.com/denis-oreshkevich/shortener/internal/app/config"
	"github.com/denis-oreshkevich/shortener/internal/app/handler"
	"github.com/denis-oreshkevich/shortener/internal/app/repo"
)

var repository = repo.New()

func InitServer() {
	r := handler.SetupRouter(repository)
	conf := config.Get()

	err := r.Run(fmt.Sprintf("%s:%s", conf.Host, conf.Port))

	if err != nil {
		panic(err)
	}
}
