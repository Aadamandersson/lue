ifeq ($(OS), Windows_NT)
	RM = del
	OUT = bin\lue.exe
else
	RM = rm
	OUT = bin/lue
endif

all: test build

run: build
	./${OUT}

build:
	go build -o ${OUT} main.go

test:
	go test ./...

.PHONY: all test clean
clean:
	${RM} ${OUT}
