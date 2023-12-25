package generator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRandString(t *testing.T) {
	tests := []struct {
		name   string
		length int
		assert func(result string)
	}{
		{
			name:   "simple generate #1",
			length: 8,
			assert: func(result string) {
				randStr := RandString(8)
				assert.NotEqual(t, result, randStr)
				assert.Len(t, result, 8)
			},
		},
		{
			name:   "zero length generate #2",
			length: 0,
			assert: func(result string) {
				assert.Empty(t, result)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := RandString(tt.length)
			tt.assert(res)
		})
	}
}

func TestUUIDString(t *testing.T) {
	tests := []struct {
		name   string
		assert func(result string)
	}{
		{
			name: "simple UUIDString #1",
			assert: func(result string) {
				randStr := UUIDString()
				assert.NotEqual(t, result, randStr)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := UUIDString()
			tt.assert(res)
		})
	}
}
