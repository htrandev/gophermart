package strutil

import (
	"math/rand/v2"
	"time"
)

func Random(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	rnd := rand.New(rand.NewPCG(uint64(time.Now().UnixNano()), 0))
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rnd.IntN(len(letters))]
	}
	return string(b)
}
