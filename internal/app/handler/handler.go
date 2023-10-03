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
	ApplicationJSON = "application/json; charset=utf-8"
)

type URLHandler struct {
	conf    config.Conf
	storage storage.Storage
}

func New(conf config.Conf, storage storage.Storage) URLHandler {
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
		id, err := h.storage.SaveURL(bodyURL)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
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
		url, err := h.storage.FindURL(id)
		if err != nil {
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
			return
		}
		if !validator.URL(um.URL) {
			c.String(http.StatusBadRequest, "Ошибка при валидации url")
			return
		}
		id, err := h.storage.SaveURL(um.URL)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		url := fmt.Sprintf("%s/%s", h.conf.BaseURL(), id)
		res, err := json.Marshal(NewResult(url))
		if err != nil {
			c.String(http.StatusBadRequest, "Ошибка при формировании ответного json")
			return
		}
		c.Header(ContentType, ApplicationJSON)
		c.String(http.StatusCreated, string(res))
	}
}

func (h URLHandler) NoRoute() func(c *gin.Context) {
	return func(c *gin.Context) {
		c.Data(400, TextPlain, []byte("Роут не найден"))
	}
}
