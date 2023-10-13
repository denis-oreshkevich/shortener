package generator

import (
	"github.com/denis-oreshkevich/shortener/internal/app/util/logger"
	"math/rand"
	"time"
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
