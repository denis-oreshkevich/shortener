package handler

import (
	"encoding/json"
	"fmt"
	"github.com/denis-oreshkevich/shortener/internal/app/config"
	"github.com/denis-oreshkevich/shortener/internal/app/storage"
	"github.com/denis-oreshkevich/shortener/internal/app/util/validator"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
)

const (
	ContentType     = "Content-type"
	TextPlain       = "text/plain; charset=utf-8"
	ApplicationJson = "application/json; charset=utf-8"
)

type URLHandler struct {
	conf    config.ServerConf
	storage storage.Storage
}

func New(conf config.ServerConf, storage storage.Storage) URLHandler {
	return URLHandler{
		conf:    conf,
		storage: storage,
	}
}

func (h URLHandler) Post() func(c *gin.Context) {
	return func(c *gin.Context) {
		req := c.Request
		body, err := io.ReadAll(req.Body)
		if err != nil {
			c.String(http.StatusBadRequest, "Ошибка при чтении тела запроса")
			return
		}
		bodyURL := string(body)
		if !validator.URL(bodyURL) {
			c.String(http.StatusBadRequest, "Ошибка при валидации тела запроса")
			return
		}
		id := h.storage.SaveURL(bodyURL)
		url := fmt.Sprintf("%s/%s", h.conf.BaseURL(), id)
		c.String(http.StatusCreated, url)
	}
}

func (h URLHandler) Get() func(c *gin.Context) {
	return func(c *gin.Context) {
		id := c.Param("id")
		if !validator.ID(id) {
			c.String(http.StatusBadRequest, "Ошибка при валидации параметра id")
			return
		}
		url, ok := h.storage.FindURL(id)
		if !ok {
			c.String(http.StatusBadRequest, "Не найдено сохраненного URL")
			return
		}
		c.Header(ContentType, TextPlain)
		c.Redirect(http.StatusTemporaryRedirect, url)
	}
}

func (h URLHandler) ShortenPost() func(c *gin.Context) {
	return func(c *gin.Context) {
		req := c.Request
		body, err := io.ReadAll(req.Body)
		if err != nil {
			c.String(http.StatusBadRequest, "Ошибка при чтении тела запроса")
			return
		}
		var um URLModel
		if err := json.Unmarshal(body, &um); err != nil {
			c.String(http.StatusBadRequest, "Ошибка при десериализации из json")
		}
		if !validator.URL(um.URL) {
			c.String(http.StatusBadRequest, "Ошибка при валидации url")
			return
		}
		id := h.storage.SaveURL(um.URL)
		url := fmt.Sprintf("%s/%s", h.conf.BaseURL(), id)
		res, err := json.Marshal(NewResult(url))
		if err != nil {
			c.String(http.StatusBadRequest, "Ошибка при формировании ответного json")
		}
		c.Header(ContentType, ApplicationJson)
		c.String(http.StatusCreated, string(res))
	}
}

func (h URLHandler) NoRoute() func(c *gin.Context) {
	return func(c *gin.Context) {
		c.Data(400, TextPlain, []byte("Роут не найден"))
	}
}
