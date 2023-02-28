package machine

import "fmt"

type Kernel interface {
	Println(string)
}

type kernel struct{}

func NewKernel() *kernel {
	return &kernel{}
}

func (k *kernel) Println(text string) {
	fmt.Println(text)
}
