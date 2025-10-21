package algorithms

import (
	"time"

	"math/rand"
)

const (
	letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	length  = len(letters)
)

func GetRandom09() int {
	r := rand.New(rand.NewSource(time.Now().UnixNano())) // 使用新的随机源
	return r.Intn(10)                                    // 10 是上限,不包含在内
}

func GetRandomaZ() rune {
	r := rand.New(rand.NewSource(time.Now().UnixNano())) // 使用新的随机源
	return rune(letters[r.Intn(length)])
}
