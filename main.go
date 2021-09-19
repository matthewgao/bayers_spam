package main

import (
	"bayers_spam/tokenize"
	"fmt"
)

func main() {
	tokenize.GetInstance()
	fmt.Printf("Bayers")
}
