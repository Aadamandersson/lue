package main

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/aadamandersson/lue/internal/machine"
)

type kernel struct {
	Buf string
}

func (k *kernel) Println(text string) {
	k.Buf += text + "\n"
}

func main() {
	err := filepath.WalkDir("tests/", testFile)
	if err != nil {
		fmt.Println(err)
	}
}

func testFile(filename string, f fs.DirEntry, err error) error {
	if err != nil {
		return err
	}

	if f.IsDir() {
		return nil
	}

	if filepath.Ext(filename) != ".lue" {
		fmt.Printf("skipping `%s`\n", filename)
		return nil
	}

	src, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("could not read test file `%s`: %v", filename, err)
	}

	kernel := &kernel{}
	machine.Interpret(filename, src, kernel)

	actual := kernel.Buf
	expected := expectedOutput(src, filename)
	if actual != expected {
		fmt.Printf("test `%s` failed\n\n", filename)
		fmt.Printf("actual:\n%s\n", actual)
		fmt.Printf("expected:\n%s\n", expected)
	} else {
		fmt.Printf("test `%s` passed\n", filename)
	}

	return nil
}

func expectedOutput(src []byte, filename string) string {
	outputRe := regexp.MustCompile(`^//[\s]*Output:`)
	expectRe := regexp.MustCompile(`^//[\s]*(?P<value>.*)`)

	var outBuilder strings.Builder
	s := bufio.NewScanner(strings.NewReader(string(src)))
	firstLine := true
	for s.Scan() {
		line := []byte(s.Text())
		if firstLine && !outputRe.Match(line) {
			fmt.Printf("test `%s` is missing `Output` tag\n", filename)
			break
		}
		valueI := expectRe.SubexpIndex("value")
		matches := expectRe.FindSubmatch(line)

		if !firstLine && len(matches) > 0 {
			outBuilder.WriteString(string(matches[valueI]))
			outBuilder.WriteByte('\n')
		}

		firstLine = false
	}
	return outBuilder.String()
}
