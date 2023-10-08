package storage

type Storage interface {
	SaveURL(url string) (string, error)
	FindURL(id string) (string, error)
}
