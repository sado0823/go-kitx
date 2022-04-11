package stringx

import (
	"math/rand"
	"time"
)

var (
	letterBytes = []byte("0123456789abcdefghijklmnpqrstuvwxyzABCDEFGHIJKLMNPQRSTUVWXYZ")
	randSource  = rand.New(rand.NewSource(time.Now().UnixNano()))
)

// Rand return a random string with incoming count
func Rand(count int) (randomStr string) {

	for i := 0; i < 6; i++ {
		var ranStr []byte
		for i := 0; i < count; i++ {
			ranStr = append(ranStr, letterBytes[randSource.Intn(len(letterBytes))])
		}
		randomStr = string(ranStr)
	}
	return
}
