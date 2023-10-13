package generator

import (
	"github.com/denis-oreshkevich/shortener/internal/app/util/logger"
	"math/rand"
	"time"
)

func RandString(length int) string {
	rand.New(rand.NewSource(time.Now().UnixNano()))
	characters := `ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789`
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		result[i] = characters[rand.Intn(len(characters))]
	}
	logger.Log.Info("generated ID is " + string(result))
	return string(result)
}
