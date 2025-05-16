package security

import (
	"crypto/rand"
	"fmt"
)

func NewSecureToken(len int) string {
	buf := make([]byte, len)
	rand.Read(buf)
	return fmt.Sprintf("%x", buf)
}
