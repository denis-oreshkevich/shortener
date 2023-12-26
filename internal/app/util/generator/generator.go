package generator

import (
	"math/rand"
	"time"

	"github.com/denis-oreshkevich/shortener/internal/app/util/logger"
	"github.com/google/uuid"
)

// RandString generates new random string of specified length.
func RandString() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	characters := `ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789`
	result := make([]byte, 8)
	for i := 0; i < 8; i++ {
		result[i] = characters[r.Intn(len(characters))]
	}
	logger.Log.Debug("generated ID is " + string(result))
	return string(result)
}

// UUIDString generates string UUID.
func UUIDString() string {
	id := uuid.New()
	return id.String()
}
