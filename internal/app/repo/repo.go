package repo

import "fmt"

type Repository interface {
	SaveURL(id, url string)
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

func (r repo) SaveURL(id, url string) {
	r[id] = url
	fmt.Println("Saved to storage with id =", id, "and value =", url)
}

func (r repo) FindURL(id string) (string, bool) {
	val, ok := r[id]
	fmt.Println("Found in storage by id =", id, "and status =", ok)
	return val, ok
}
