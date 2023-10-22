package server

import (
	"encoding/json"
	"fmt"
	"github.com/denis-oreshkevich/shortener/internal/app/config"
	"github.com/denis-oreshkevich/shortener/internal/app/model"
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
	conf    config.Conf
	storage storage.Storage
}

func New(conf config.Conf, st storage.Storage) Server {
	return Server{
		conf:    conf,
		storage: st,
	}
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
	id, err := s.storage.SaveURL(c.Request.Context(), bodyURL)
	if err != nil {
		logger.Log.Error("saveURL", zap.Error(err))
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	url := fmt.Sprintf("%s/%s", s.conf.BaseURL(), id)
	c.String(http.StatusCreated, url)
}

func (s Server) Get(c *gin.Context) {
	id := c.Param("id")
	if !validator.ID(id) {
		logger.Log.Warn(fmt.Sprintf("validate ID %s", id))
		c.String(http.StatusBadRequest, "Ошибка при валидации параметра id")
		return
	}
	url, err := s.storage.FindURL(c.Request.Context(), id)
	if err != nil {
		logger.Log.Error("findURL", zap.Error(err))
		c.String(http.StatusBadRequest, "Не найдено сохраненного URL")
		return
	}
	c.Header(ContentType, TextPlain)
	c.Redirect(http.StatusTemporaryRedirect, url)
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
	id, err := s.storage.SaveURL(req.Context(), um.URL)
	if err != nil {
		logger.Log.Error("saveURL", zap.Error(err))
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	url := fmt.Sprintf("%s/%s", s.conf.BaseURL(), id)
	res, err := json.Marshal(NewResult(url))
	if err != nil {
		logger.Log.Error("marshal response", zap.Error(err))
		c.String(http.StatusBadRequest, "Ошибка при формировании ответного json")
		return
	}
	c.Header(ContentType, ApplicationJSON)
	c.String(http.StatusCreated, string(res))
}

func (s Server) Ping(c *gin.Context) {
	st, ok := s.storage.(*storage.DBStorage)
	if !ok {
		logger.Log.Error("ping() is not DB storage")
		c.AbortWithStatus(500)
		return
	}
	err := st.Ping(c.Request.Context())
	if err != nil {
		logger.Log.Error("ping() dbStorage.Ping()", zap.Error(err))
		c.AbortWithError(500, err)
		return
	}
	c.AbortWithStatus(http.StatusOK)
}

func (s Server) NoRoute(c *gin.Context) {
	c.Data(400, TextPlain, []byte("Роут не найден"))
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
	respEntries, err := s.storage.SaveURLBatch(req.Context(), batch)
	if err != nil {
		logger.Log.Error("saveURLBatch", zap.Error(err))
		c.String(http.StatusBadRequest, "Ошибка при сохранении данных")
		return
	}
	resp, err := json.Marshal(respEntries)
	if err != nil {
		logger.Log.Error("marshal response", zap.Error(err))
		c.String(http.StatusBadRequest, "Ошибка при формировании ответного json")
		return
	}
	c.Header(ContentType, ApplicationJSON)
	c.String(http.StatusCreated, string(resp))
}
