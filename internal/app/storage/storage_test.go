package storage

import (
	"context"
	"github.com/denis-oreshkevich/shortener/internal/app/util/generator"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func BenchmarkStorageSave(b *testing.B) {
	fn := "./test"
	fs, err := NewFileStorage(fn)
	require.NoError(b, err)
	defer func() {
		err := os.Remove(fn)
		require.NoError(b, err)
	}()
	ms := NewMapStorage()

	ctx := context.Background()
	userID := generator.UUIDString()
	baseURL := "http://localhost:8080/"

	b.ResetTimer()
	b.Run("fileStorage", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			fs.SaveURL(ctx, userID, baseURL+generator.UUIDString())
		}
	})

	b.Run("mapStorage", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ms.SaveURL(ctx, userID, baseURL+generator.UUIDString())
		}
	})
}
