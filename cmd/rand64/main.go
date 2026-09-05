package main

import (
	"crypto/rand"
	"encoding/base64"
	"flag"
	"fmt"
)

var (
	size int
)

func main() {
	flag.IntVar(&size, "size", 32, "amount of random bytes to generate")
	flag.Parse()

	if size <= 0 {
		fmt.Println("Error: size must be a positive number.")
		return
	}

	b := make([]byte, size)
	rand.Read(b)

	enc := base64.StdEncoding.EncodeToString(b)
	fmt.Print(enc)
}
