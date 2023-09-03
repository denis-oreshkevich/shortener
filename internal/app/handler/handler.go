package handler

import (
	"fmt"
	"github.com/denis-oreshkevich/shortener/internal/app/constant"
	"github.com/denis-oreshkevich/shortener/internal/app/generator"
	"github.com/denis-oreshkevich/shortener/internal/app/repo"
	"io"
	"net/http"
	"regexp"
)

const (
	HeaderLocation    = "Location"
	HeaderContentType = "Content-Type"
	URLRegex          = "(?:http|https):\\/\\/(www\\\\.)?[-a-zA-Z0-9@:%._\\+~#=]{1,256}\\.[a-zA-Z0-9()]{1,6}\\b([-a-zA-Z0-9()@:%_\\+.~#?&//=]*)"

	IDRegex = "^\\/[A-Za-z]{8}"
)

var postValidator = regexp.MustCompile(URLRegex)
var getValidator = regexp.MustCompile(IDRegex)

func URL(rep repo.Repository) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		fmt.Println("New request received with url =", req.URL.Path, ",method =", req.Method)
		res.Header().Set(HeaderContentType, "text/plain")
		switch req.Method {
		case "POST":
			handlePost(rep, res, req)
			return
		case "GET":
			handleGet(rep, res, req)
			return
		default:
			res.WriteHeader(http.StatusBadRequest)
			res.Write([]byte("Meтод не поддерживается, я 405"))
		}

	}
}

func handlePost(rep repo.Repository, res http.ResponseWriter, req *http.Request) {
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
	bodyUrl := string(body)
	if !postValidator.MatchString(bodyUrl) {
		res.WriteHeader(http.StatusBadRequest)
		res.Write([]byte("Ошибка при валидации тела запроса"))
		return
	}
	id := saveURL(rep, bodyUrl)
	url := constant.ServerURL + id
	res.WriteHeader(http.StatusCreated)
	res.Write([]byte(url))
	return
}

func handleGet(rep repo.Repository, res http.ResponseWriter, req *http.Request) {
	id := req.URL.Path
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
	return
}

func saveURL(rep repo.Repository, url string) string {
	id := generator.RandString(8)
	id = "/" + id
	rep.SaveURL(id, url)
	return id
}
