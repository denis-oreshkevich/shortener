package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/denis-oreshkevich/shortener/internal/app/config"
	"github.com/denis-oreshkevich/shortener/internal/app/model"
	"github.com/denis-oreshkevich/shortener/internal/app/shortener"
	"github.com/denis-oreshkevich/shortener/internal/app/storage"
	"github.com/denis-oreshkevich/shortener/internal/app/util/logger"
	"github.com/denis-oreshkevich/shortener/internal/app/util/validator"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"io"
	"net/http"
)

const (
	ContentType     = "Content-type"
	TextPlain       = "text/plain; charset=utf-8"
	ApplicationJSON = "application/json; charset=utf-8"
)

type Server struct {
	conf       config.Conf
	sh         *shortener.Shortener
	delChannel chan model.BatchDeleteEntry
}

func New(conf config.Conf, sh *shortener.Shortener,
	delChannel chan model.BatchDeleteEntry) *Server {
	inst := &Server{
		conf:       conf,
		sh:         sh,
		delChannel: delChannel,
	}
	return inst
}

func (s Server) Post(c *gin.Context) {
	req := c.Request
	body, err := io.ReadAll(req.Body)
	if err != nil {
		logger.Log.Error("readAll", zap.Error(err))
		c.String(http.StatusBadRequest, "Ошибка при чтении тела запроса")
		return
	}
	bodyURL := string(body)
	if !validator.URL(bodyURL) {
		logger.Log.Warn(fmt.Sprintf("validate URL %s", bodyURL))
		c.String(http.StatusBadRequest, "Ошибка при валидации тела запроса")
		return
	}
	ctx := c.Request.Context()
	id, err := s.sh.SaveURL(ctx, bodyURL)
	if err != nil {
		if errors.Is(err, storage.ErrDBConflict) {
			logger.Log.Info(fmt.Sprintf("saveURL conflict on original url = %s", bodyURL))
			url := fmt.Sprintf("%s/%s", s.conf.BaseURL(), id)
			c.String(http.StatusConflict, url)
			return
		}
		logger.Log.Error("saveURL", zap.Error(err))
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	url := fmt.Sprintf("%s/%s", s.conf.BaseURL(), id)
	c.String(http.StatusCreated, url)
}

func (s Server) Get(c *gin.Context) {
	id := c.Param("id")
	log := logger.Log.With(zap.String("id", id))
	if !validator.ID(id) {
		log.Warn(fmt.Sprintf("validate ID %s", id))
		c.String(http.StatusBadRequest, "Ошибка при валидации параметра id")
		return
	}
	ctx := c.Request.Context()
	url, err := s.sh.FindURL(ctx, id)
	if err != nil {
		if errors.Is(err, storage.ErrResultIsDeleted) {
			log.Debug("record is already deleted", zap.Error(err))
			c.AbortWithStatus(http.StatusGone)
			return
		}
		log.Error("findURL", zap.Error(err))
		c.String(http.StatusBadRequest, "Не найдено сохраненного URL")
		return
	}
	c.Header(ContentType, TextPlain)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

func (s Server) GetUsersURLs(c *gin.Context) {
	ctx := c.Request.Context()
	urls, err := s.sh.FindUserURLs(ctx)
	if err != nil {
		if errors.Is(err, shortener.ErrUserIsNew) {
			logger.Log.Debug("user is new")
			c.AbortWithStatus(http.StatusUnauthorized)
			return

		}
		if errors.Is(err, shortener.ErrUserItemsNotFound) {
			logger.Log.Debug("findUserURLs items not found")
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		logger.Log.Error("findUserURLs", zap.Error(err))
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	resp, err := json.Marshal(urls)
	if err != nil {
		logger.Log.Error("marshal response", zap.Error(err))
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.Data(http.StatusOK, ApplicationJSON, resp)
}

func (s Server) ShortenPost(c *gin.Context) {
	req := c.Request
	body, err := io.ReadAll(req.Body)
	if err != nil {
		logger.Log.Error("readAll", zap.Error(err))
		c.String(http.StatusBadRequest, "Ошибка при чтении тела запроса")
		return
	}
	var um URLModel
	if err := json.Unmarshal(body, &um); err != nil {
		logger.Log.Error("unmarshal", zap.Error(err))
		c.String(http.StatusBadRequest, "Ошибка при десериализации из json")
		return
	}
	if !validator.URL(um.URL) {
		logger.Log.Warn(fmt.Sprintf("validate URL %s", um.URL))
		c.String(http.StatusBadRequest, "Ошибка при валидации url")
		return
	}
	id, err := s.sh.SaveURL(req.Context(), um.URL)
	if err != nil {
		if errors.Is(err, storage.ErrDBConflict) {
			logger.Log.Info(fmt.Sprintf("saveURL conflict on original url = %s", um.URL))
			s.sendJSONResultResp(c, id, http.StatusConflict)
			return
		}
		logger.Log.Error("saveURL", zap.Error(err))
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	s.sendJSONResultResp(c, id, http.StatusCreated)
}

func (s Server) Ping(c *gin.Context) {
	ctx := c.Request.Context()
	err := s.sh.Ping(ctx)
	if err != nil {
		logger.Log.Error("shortener ping", zap.Error(err))
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.AbortWithStatus(http.StatusOK)
}

func (s Server) NoRoute(c *gin.Context) {
	c.Data(http.StatusBadRequest, TextPlain, []byte("Роут не найден"))
}

func (s Server) ShortenBatch(c *gin.Context) {
	req := c.Request
	body, err := io.ReadAll(req.Body)
	if err != nil {
		logger.Log.Error("readAll", zap.Error(err))
		c.String(http.StatusBadRequest, "Ошибка при чтении тела запроса")
		return
	}
	var batch []model.BatchReqEntry
	if err := json.Unmarshal(body, &batch); err != nil {
		logger.Log.Error("unmarshal", zap.Error(err))
		c.String(http.StatusBadRequest, "Ошибка при десериализации из json")
		return
	}
	if len(batch) == 0 {
		logger.Log.Warn("batch len = 0")
		c.String(http.StatusBadRequest, "Длина батча равна 0")
		return
	}
	respEntries, err := s.sh.SaveURLBatch(req.Context(), batch)
	if err != nil {
		logger.Log.Error("saveURLBatch", zap.Error(err))
		c.String(http.StatusBadRequest, "Ошибка при сохранении данных")
		return
	}
	resp, err := json.Marshal(respEntries)
	if err != nil {
		logger.Log.Error("marshal response", zap.Error(err))
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.Data(http.StatusCreated, ApplicationJSON, resp)
}

func (s Server) DeleteURLs(c *gin.Context) {
	req := c.Request
	ctx := c.Request.Context()
	userID, userErr := s.sh.GetUserID(ctx)
	body, bodyErr := io.ReadAll(req.Body)
	in := make(chan model.BatchDeleteEntry)
	f := func() {
		defer close(in)
		if userErr != nil {
			logger.Log.Error("get userID", zap.Error(userErr))
			return
		}

		if bodyErr != nil {
			logger.Log.Error("readAll", zap.Error(bodyErr))
			return
		}
		var batch []string
		if err := json.Unmarshal(body, &batch); err != nil {
			logger.Log.Error("unmarshal", zap.Error(err))
			return
		}
		if len(batch) == 0 {
			logger.Log.Warn("batch len = 0")
			return
		}
		entry := model.NewBatchDeleteEntry(userID, batch)
		logger.Log.Debug("send to channel in")
		in <- entry
	}
	go f()
	go func() {
		entry := <-in
		logger.Log.Debug("send to delChannel")
		s.delChannel <- entry
	}()

	c.AbortWithStatus(http.StatusAccepted)
}

func (s Server) sendJSONResultResp(c *gin.Context, id string, status int) {
	url := fmt.Sprintf("%s/%s", s.conf.BaseURL(), id)
	resp, err := json.Marshal(NewResult(url))
	if err != nil {
		logger.Log.Error("buildJSONResp", zap.Error(err))
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.Data(status, ApplicationJSON, resp)
}
