package main

import (
	"fmt"
	"os"

	"github.com/aadamandersson/lue/internal/evaluator"
)

func main() {
	inputFile := "examples/basic.lue"
	src, err := os.ReadFile(inputFile)
	if err != nil {
		fmt.Printf("could not read file `%s`: %v\n", inputFile, err)
		os.Exit(1)
	}
	result := evaluator.Evaluate(src)
	fmt.Println(result)
}
