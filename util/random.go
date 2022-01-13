package util

import (
	"math/rand"
	"strings"
	"time"
)

const (
	alphabet = "abcdefghijklmnopqrstuvwxyz"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// RandomInt generate random int
func RandomInt(min, max int64) int64 {
	return min + rand.Int63n(max-min+1)
}

// RandomString generate random string
func RandomString(n int) string {
	var sb strings.Builder
	k := len(alphabet)
	for i := 0; i < n; i++ {
		c := alphabet[rand.Intn(k)]
		sb.WriteByte(c)
	}

	return sb.String()
}

// RandomOwner generate random owner
func RandomOwner() string {
	return RandomString(6)
}

// RandomMoney generate a random money
func RandomMoney() int64 {
	return RandomInt(0, 1000)
}

// RandomCurrency generate a random currency
func RandomCurrency() string {
	currencyList := []string{"USD", "EUR"}
	n := len(currencyList)
	return currencyList[rand.Intn(n)]
}
