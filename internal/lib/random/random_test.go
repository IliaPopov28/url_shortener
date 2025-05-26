package random

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewRandomString(t *testing.T) {
	tests := []struct {
		name string
		size int
	}{
		{
			name: "size 1",
			size: 1,
		},
		{
			name: "size 2",
			size: 2,
		},
		{
			name: "size 5",
			size: 5,
		},
		{
			name: "size 7",
			size: 7,
		},
		{
			name: "size 15",
			size: 15,
		},
		{
			name: "size 30",
			size: 30,
		},
		{
			name: "size 50",
			size: 50,
		},
		{
			name: "size 100",
			size: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			str1 := NewRandomString(tt.size)
			str2 := NewRandomString(tt.size)
			str3 := NewRandomString(tt.size)

			assert.Len(t, str1, tt.size)
			assert.Len(t, str2, tt.size)
			assert.Len(t, str3, tt.size)

			assert.NotEqual(t, str1, str2)
			assert.NotEqual(t, str1, str3)
			assert.NotEqual(t, str2, str3)
		})
	}
}
