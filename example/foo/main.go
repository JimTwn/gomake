package main

import (
	"log"
	"os"
)

func main() {
	data, err := os.ReadFile("file.txt") // shuold be in bin directory with this program.
	if err != nil {
		log.Fatal(err)
	}
	log.Println(string(data))
}
