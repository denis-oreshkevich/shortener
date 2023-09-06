package repo

import (
	"fmt"
	"github.com/denis-oreshkevich/shortener/internal/app/util/generator"
)

type Repository interface {
	SaveURL(url string) string
	FindURL(id string) (string, bool)
}

type repo map[string]string

var repositoryMap repo

func init() {
	repositoryMap = make(map[string]string)
}
func New() Repository {
	return repositoryMap
}

func (r repo) SaveURL(url string) string {
	id := generator.RandString(8)
	r[id] = url
	fmt.Println("Saved to storage with id =", id, "and value =", url)
	return id
}

func (r repo) FindURL(id string) (string, bool) {
	val, ok := r[id]
	fmt.Println("Found in storage by id =", id, "and status =", ok)
	return val, ok
}
