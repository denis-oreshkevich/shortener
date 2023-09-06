package handler

import (
	"fmt"
	"github.com/denis-oreshkevich/shortener/internal/app/constant"
	"github.com/denis-oreshkevich/shortener/internal/app/repo"
	"github.com/denis-oreshkevich/shortener/internal/app/util/generator"
	"github.com/denis-oreshkevich/shortener/internal/app/util/validator"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
)

const (
	ContentType = "Content-type"
	TextPlain   = "text/plain; charset=utf-8"
)

func SetupRouter(repository repo.Repository) *gin.Engine {
	r := gin.Default()

	r.POST(`/`, post(repository))
	//r.POST(`/`, gin.WrapF(post(repository)))
	r.GET(`/:id`, get(repository))

	r.NoRoute(func(c *gin.Context) {
		c.Data(400, TextPlain, []byte("Роут не найден"))
	})
	return r
}

func post(repository repo.Repository) func(c *gin.Context) {
	return func(c *gin.Context) {
		req := c.Request
		body, err := io.ReadAll(req.Body)
		if err != nil {
			c.String(http.StatusBadRequest, "Ошибка при чтении тела запроса")
		}
		bodyURL := string(body)
		if !validator.URL(bodyURL) {
			c.String(http.StatusBadRequest, "Ошибка при валидации тела запроса")
		}
		id := saveURL(repository, bodyURL)
		url := fmt.Sprintf("%s/%s", constant.ServerURL, id)
		c.String(http.StatusCreated, url)
	}
}

func get(repository repo.Repository) func(c *gin.Context) {
	return func(c *gin.Context) {
		id := c.Param("id")
		if !validator.ID(id) {
			c.String(http.StatusBadRequest, "Ошибка при валидации параметра id")
		}
		url, ok := repository.FindURL(id)
		if !ok {
			c.String(http.StatusBadRequest, "Не найдено сохраненного URL")
		}
		c.Header(ContentType, TextPlain)
		c.Redirect(http.StatusTemporaryRedirect, url)
	}
}

func saveURL(rep repo.Repository, url string) string {
	id := generator.RandString(8)
	rep.SaveURL(id, url)
	return id
}
