package generator

import (
	"crypto/sha256"
	"fmt"
	"math/rand"
	"strings"
)

const base62 = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

// Generator generates a base62 string from an integer
func Generator(id int64) string {
	var result strings.Builder

	if id == 0 {
		return base62[0:1]
	}

	for id > 0 {
		result.WriteByte(base62[id%62])
		id /= 62
	}

	return reverse(result.String())
}

func reverse(str string) string {
	result := []rune(str)

	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}

	return string(result)
}

func GeneratorSha256(url string, salt string) string {
	if salt != "" {
		url += salt
	}
	// Convert to byte slice
	data := []byte(url)

	shaSum := sha256.Sum256(data)
	shaSumStr := fmt.Sprintf("%x", shaSum)

	var short strings.Builder

	for i := 0; i < 8; i += 1 {
		pos := rand.Intn(31)
		short.WriteByte(shaSumStr[pos])
	}

	return short.String()
}

// NewSalt generates a random 10 character string
func NewSalt() string {
	var salt strings.Builder

	for i := 0; i < 10; i += 1 {
		pos := rand.Intn(61)
		salt.WriteByte(base62[pos])
	}
	return salt.String()
}
