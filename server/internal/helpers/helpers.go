package helpers

import (
	"crypto/rand"
	"fmt"
)

const tokenLength = 10

//GenerateToken - создает токен
func GenerateToken() string {
	b := make([]byte, tokenLength)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}
