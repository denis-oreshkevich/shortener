package handler

import (
	"github.com/denis-oreshkevich/shortener/internal/app/storage"
	"github.com/denis-oreshkevich/shortener/internal/app/util/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
)

type PingHandler struct {
	dbStorage *storage.DBStorage
}

func NewPingHandler(dbStorage *storage.DBStorage) PingHandler {
	return PingHandler{dbStorage: dbStorage}
}

func (ph PingHandler) Ping() func(c *gin.Context) {
	return func(c *gin.Context) {
		err := ph.dbStorage.Ping()
		if err != nil {
			logger.Log.Error("pingHandler ping() dbStorage.Ping()", zap.Error(err))
			c.AbortWithError(500, err)
			return
		}
		c.AbortWithStatus(http.StatusOK)
	}
}
