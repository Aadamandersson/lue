package main

import (
	"fmt"
	"os"

	"github.com/aadamandersson/lue/internal/machine"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: lue <path>")
		os.Exit(1)
	}

	path := os.Args[1]
	src, err := os.ReadFile(path)
	if err != nil {
		fmt.Printf("could not read file `%s`: %v\n", path, err)
		os.Exit(1)
	}

	kernel := machine.NewKernel()
	ok := machine.Interpret(path, src, kernel)
	if !ok {
		fmt.Printf("error: could not interpret `%s` due to previous errors.\n", path)
	}
}
