package randutil

import (
	"math/rand"
	"time"
)

func Random(min int, max int) int {
	rand.Seed(time.Now().Unix())
	return min + rand.Intn(max+1-min)
}

func RandomPercent(percent int) bool {
	return Random(0, 10000) <= percent*100
}
