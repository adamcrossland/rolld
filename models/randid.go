package models

import (
	"fmt"
	"os"
)

func GetRandomID(byteCount int) string {
	f, _ := os.Open("/dev/urandom")
	defer f.Close()
	b := make([]byte, byteCount)
	f.Read(b)
	return fmt.Sprintf("%x", b)
}
