package models

import (
	"fmt"
	"os"
)

// GetRandomID returns a string representing the requested number of random bytes
// in Base16. Depends on the existence of /dev/urandom which makes it Linux-specific.
func GetRandomID(byteCount int) string {
	f, _ := os.Open("/dev/urandom")
	defer f.Close()
	b := make([]byte, byteCount)
	f.Read(b)
	return fmt.Sprintf("%x", b)
}
