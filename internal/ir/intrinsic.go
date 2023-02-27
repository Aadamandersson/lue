package ir

import "strconv"

type Intrinsic int

const (
	IntrPrintln Intrinsic = iota
)

func Intrinsics() []Intrinsic {
	return []Intrinsic{IntrPrintln}
}

var intrinsics = [...]string{
	IntrPrintln: "println",
}

func (i Intrinsic) String() string {
	if i < 0 || i > Intrinsic(len(intrinsics)) {
		return "Intrinsic(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return intrinsics[IntrPrintln]
}
