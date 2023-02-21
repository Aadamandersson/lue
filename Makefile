ifeq ($(OS), Windows_NT)
	RM = del
	OUT = bin\lue.exe
	CMD = cmd\lue\main.go
else
	RM = rm
	OUT = bin/lue
	CMD = cmd/lue/main.go
endif

all: test build

run: build
	./${OUT}

build:
	go build -o ${OUT} ${CMD}

test:
	go test ./...

.PHONY: all test clean
clean:
	${RM} ${OUT}
