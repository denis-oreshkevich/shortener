package storage

import "fmt"

var storage = make(map[string]string)

func SaveURL(id, url string) {
	storage[id] = url
	fmt.Println("Saved to storage with id =", id, "and value =", url)
}

func FindURL(id string) (string, bool) {
	val, ok := storage[id]
	fmt.Println("Found in storage by id =", id, "and status =", ok)
	return val, ok
}
