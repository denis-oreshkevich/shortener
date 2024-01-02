package generator

import (
	"math/rand"
	"time"

	"github.com/denis-oreshkevich/shortener/internal/app/util/logger"
	"github.com/google/uuid"
)

func RandString(length int) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	characters := `ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789`
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		result[i] = characters[r.Intn(len(characters))]
	}
	logger.Log.Debug("generated ID is " + string(result))
	return string(result)
}

func UUIDString() string {
	id := uuid.New()
	return id.String()
}
