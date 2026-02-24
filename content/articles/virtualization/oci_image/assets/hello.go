package main

import (
	"fmt"
	"os"
)

func main() {
	greeting := os.Getenv("GREETING")

	if greeting != "" {
		fmt.Println(greeting)
	} else {
		fmt.Println("HELLO WORLD!!!")
	}
}
