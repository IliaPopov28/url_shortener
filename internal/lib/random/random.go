package random

import (
	"math/rand"
	"time"
)

var globalRand = rand.New(rand.NewSource(time.Now().UnixNano()))

func NewRandomString(size int) string {
	chars := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	b := make([]rune, size)
	for i := range b {
		b[i] = chars[globalRand.Intn(len(chars))]
	}
	return string(b)
}
