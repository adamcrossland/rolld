package models

import (
	"fmt"
	"math/rand"
)

// GetRandomID returns a string representing the requested number of random bytes
// in Base16. Depends on the existence of /dev/urandom which makes it Linux-specific.
func GetRandomID(byteCount int) string {
	b := make([]byte, byteCount)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}
