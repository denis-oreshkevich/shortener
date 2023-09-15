package handler

import (
	"fmt"
	"github.com/denis-oreshkevich/shortener/internal/app/constant"
	"github.com/denis-oreshkevich/shortener/internal/app/generator"
	"github.com/denis-oreshkevich/shortener/internal/app/storage"
	"io"
	"net/http"
)

const ()

func Handle(res http.ResponseWriter, req *http.Request) {
	fmt.Println("New request received with method", req.Method)
	switch req.Method {
	case "POST":
		body, err := io.ReadAll(req.Body)
		if err != nil {
			res.WriteHeader(http.StatusBadRequest)
			res.Write([]byte("Ошибка при чтении тела запроса"))
			return
		}
		id := saveURL(body)
		result := constant.ServerURL + id
		res.WriteHeader(http.StatusCreated)
		res.Write([]byte(result))
	case "GET":
		id := req.URL.Path
		result, ok := storage.FindURL(id)
		if !ok {
			res.WriteHeader(http.StatusBadRequest)
			res.Write([]byte("Не найдено сохраненного запроса"))
			return
		}
		fmt.Println("Response Location", result)
		res.Header().Set("Location", result)
		res.WriteHeader(http.StatusTemporaryRedirect)
		return
	default:
		res.WriteHeader(http.StatusBadRequest)
		res.Write([]byte("Meтод не поддерживается, я 405"))
	}

}

func saveURL(body []byte) string {
	id := generator.RandString(8)
	id = "/" + id
	storage.SaveURL(id, string(body))
	return id
}
