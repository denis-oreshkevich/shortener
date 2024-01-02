package generator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRandString(t *testing.T) {
	tests := []struct {
		name   string
		assert func(result string)
	}{
		{
			name: "simple generate #1",
			assert: func(result string) {
				randStr := RandString()
				assert.NotEqual(t, result, randStr)
				assert.Len(t, result, 8)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := RandString()
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
