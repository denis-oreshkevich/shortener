package handler

import (
	"fmt"
	"github.com/denis-oreshkevich/shortener/internal/app/constant"
	"github.com/denis-oreshkevich/shortener/internal/app/generator"
	"github.com/denis-oreshkevich/shortener/internal/app/repo"
	"io"
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"
)

const (
	HeaderLocation    = "Location"
	HeaderContentType = "Content-Type"
	URLRegex          = "(?:http|https):\\/\\/(www\\.)?[-a-zA-Z0-9@:%._\\+~#=]{1,256}\\.[a-zA-Z0-9()]{1,6}\\b([-a-zA-Z0-9()@:%_\\+.~#?&//=]*)"

	IDRegex = "^[A-Za-z]{8}"
)

var postValidator = regexp.MustCompile(URLRegex)
var getValidator = regexp.MustCompile(IDRegex)

func SetupRouter(repository repo.Repository) *gin.Engine {
	r := gin.Default()

	r.POST(`/`, gin.WrapF(Post(repository)))
	r.GET(`/:id`, func(c *gin.Context) {
		id := c.Param("id")
		fmt.Println("id", id)
		get := Get(id, repository)
		get.ServeHTTP(c.Writer, c.Request)
	})

	r.NoRoute(func(c *gin.Context) {
		c.Data(400, "text/plain", []byte("Роут не найден"))
	})
	return r
}

func Post(rep repo.Repository) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		res.Header().Set(HeaderContentType, "text/plain")
		if req.URL.Path != "/" {
			res.WriteHeader(http.StatusBadRequest)
			res.Write([]byte("Неверный адрес запроса"))
			return
		}
		body, err := io.ReadAll(req.Body)
		if err != nil {
			res.WriteHeader(http.StatusBadRequest)
			res.Write([]byte("Ошибка при чтении тела запроса"))
			return
		}
		bodyURL := string(body)
		if !postValidator.MatchString(bodyURL) {
			res.WriteHeader(http.StatusBadRequest)
			res.Write([]byte("Ошибка при валидации тела запроса"))
			return
		}
		id := saveURL(rep, bodyURL)
		url := fmt.Sprintf("%s/%s", constant.ServerURL, id)
		res.WriteHeader(http.StatusCreated)
		res.Write([]byte(url))
	}
}

func Get(id string, rep repo.Repository) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		res.Header().Set(HeaderContentType, "text/plain")
		if !getValidator.MatchString(id) {
			res.WriteHeader(http.StatusBadRequest)
			res.Write([]byte("Ошибка при валидации параметра id"))
			return
		}
		url, ok := rep.FindURL(id)
		if !ok {
			res.WriteHeader(http.StatusBadRequest)
			res.Write([]byte("Не найдено сохраненного запроса"))
			return
		}
		fmt.Println("Response Location", url)
		res.Header().Set(HeaderLocation, url)
		res.WriteHeader(http.StatusTemporaryRedirect)
	}
}

func saveURL(rep repo.Repository, url string) string {
	id := generator.RandString(8)
	rep.SaveURL(id, url)
	return id
}
