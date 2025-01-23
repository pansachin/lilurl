package generator

import "strings"

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
