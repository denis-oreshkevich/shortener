package server

import (
	"fmt"
	"github.com/denis-oreshkevich/shortener/internal/app/config"
	"github.com/denis-oreshkevich/shortener/internal/app/handler"
	"github.com/gin-gonic/gin"
)

type Server struct {
	conf   config.Conf
	router *gin.Engine
}

func New(conf config.Conf, uh handler.URLHandler) Server {
	r := gin.New()

	r.Use(gin.Recovery(), Gzip, Logging)

	r.POST(`/`, uh.Post())
	r.GET(conf.BasePath()+`/:id`, uh.Get())
	r.POST(`/api/shorten`, uh.ShortenPost())
	r.NoRoute(uh.NoRoute())

	return Server{
		conf:   conf,
		router: r,
	}
}

func (s Server) Start() error {
	err := s.router.Run(fmt.Sprintf("%s:%s", s.conf.Host(), s.conf.Port()))
	return err
}
