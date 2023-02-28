package machine

import "fmt"

type Kernel interface {
	println(string)
}

type kernel struct{}

func NewKernel() *kernel {
	return &kernel{}
}

func (k *kernel) println(text string) {
	fmt.Println(text)
}
