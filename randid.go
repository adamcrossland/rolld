package main

import (
	"fmt"
	"os"
)

func GetRandomID() string {
	f, _ := os.Open("/dev/urandom")
	defer f.Close()
	b := make([]byte, 16)
	f.Read(b)
	return fmt.Sprintf("%x", b)
}
