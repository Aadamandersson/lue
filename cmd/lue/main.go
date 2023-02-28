package main

import (
	"fmt"
	"os"

	"github.com/aadamandersson/lue/internal/machine"
)

func main() {
	filename := "examples/basic.lue"
	src, err := os.ReadFile(filename)
	if err != nil {
		fmt.Printf("could not read file `%s`: %v\n", filename, err)
		os.Exit(1)
	}
	kernel := machine.NewKernel()
	ok := machine.Interpret(filename, src, kernel)
	if !ok {
		fmt.Printf("error: could not interpret `%s` due to previous errors.\n", filename)
	}
}
