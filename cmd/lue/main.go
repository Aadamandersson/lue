package main

import (
	"fmt"

	"github.com/aadamandersson/lue/internal/evaluator"
)

func main() {
	src := "2 + 3 * 4 + 5"
	result := evaluator.Evaluate([]byte(src))
	fmt.Println(result)
}
