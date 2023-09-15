package storage

type Storage interface {
	SaveURL(url string) string
	FindURL(id string) (string, bool)
}
