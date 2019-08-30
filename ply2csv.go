package main

import (
	"fmt"
	"os"
)

type ply_header struct{}

func main() {
	var header ply_header

	if infile, err := os.Open(os.Args[1]); err != nil {
		panic(err)
	} else {
		header := parse_header()
	}
	fmt.Println(header)
}
