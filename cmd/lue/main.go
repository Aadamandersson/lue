package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/aadamandersson/lue/internal/machine"
)

func main() {
	var path string
	flag.StringVar(&path, "path", "", "Path to .lue file to interpret")
	flag.Parse()

	if path == "" {
		flag.Usage()
		os.Exit(1)
	}

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
