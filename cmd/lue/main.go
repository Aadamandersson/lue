package main

import (
	"fmt"
	"os"

	"github.com/aadamandersson/lue/internal/evaluator"
)

func main() {
	filename := "examples/basic.lue"
	src, err := os.ReadFile(filename)
	if err != nil {
		fmt.Printf("could not read file `%s`: %v\n", filename, err)
		os.Exit(1)
	}
	ok := evaluator.Evaluate(filename, src)
	if !ok {
		fmt.Printf("error: could not interpret `%s` due to previous errors.\n", filename)
	}
}
