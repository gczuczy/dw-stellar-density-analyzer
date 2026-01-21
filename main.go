package main

import (
	"fmt"

	"github.com/gczuczy/dw-stellar-density-analyzer/pkg/google"
)


func main() {
	_, err := google.NewSheets()
	if err != nil {
		fmt.Printf("err: %v\n", err)
	}
}
