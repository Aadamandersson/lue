ifeq ($(OS), Windows_NT)
	RM = del
	OUT_LUE = bin\lue.exe
	CMD_LUE = cmd\lue\main.go
	OUT_LUET = bin\luet.exe
	CMD_LUET = cmd\luet\main.go
else
	RM = rm
	OUT_LUE = bin/lue
	CMD_LUE = cmd/lue/main.go
	OUT_LUET = bin/luet
	CMD_LUET = cmd/luet/main.go
endif

all: test build

build:
	go build -o ${OUT_LUE} ${CMD_LUE}
	go build -o ${OUT_LUET} ${CMD_LUET}

test:
	go test ./...

.PHONY: all test clean
clean:
	${RM} ${OUT_LUE}
	${RM} ${OUT_LUET}
