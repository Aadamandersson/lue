package main

import (
	"bufio"
	"flag"
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
	var blessed bool
	flag.BoolVar(
		&blessed,
		"bless",
		false,
		"Set to true to generate the expected output from the machine.",
	)
	flag.Parse()

	fn := func(filename string, f fs.DirEntry, err error) error {
		return testFile(blessed, filename, f, err)
	}
	err := filepath.WalkDir("tests/", fn)
	if err != nil {
		fmt.Println(err)
	}
}

func testFile(blessed bool, filename string, f fs.DirEntry, err error) error {
	if err != nil {
		return err
	}

	if f.IsDir() {
		return nil
	}

	if filepath.Ext(filename) != ".lue" {
		if filepath.Ext(filename) != ".stdout" {
			fmt.Printf("skipping `%s`\n", filename)
		}
		return nil
	}

	src, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("could not read test file `%s`: %v", filename, err)
	}

	kernel := &kernel{}
	machine.Interpret(filename, src, kernel)
	if blessed {
		err := blessFile(filename, kernel.Buf)
		if err != nil {
			return err
		}
	}

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

	got := kernel.Buf
	want := outBuilder.String()
	if got != want {
		fmt.Printf("test `%s` failed\n\n", filename)
		fmt.Printf("got:\n%s\n", got)
		fmt.Printf("want:\n%s\n", want)
	} else {
		fmt.Printf("test `%s` passed\n", filename)
	}

	return nil
}

/*
func expectedOutput(blessed bool, filename string, kernel *kernel) (string, error) {
	testDir := strings.TrimSuffix(filename, filepath.Ext(filename))
	stdoutFile := fmt.Sprintf(
		"%s/%s.stdout",
		testDir,
		filepath.Base(testDir),
	)

	if blessed {
		return blessFile(testDir, stdoutFile, kernel.Buf)
	}

	stdoutBytes, err := os.ReadFile(stdoutFile)
	if err != nil {
		return "", fmt.Errorf("could not read stdout file `%s`: %v", stdoutFile, err)
	}
	return string(stdoutBytes), nil
}*/

func blessFile(filename string, with string) error {
	testDir := strings.TrimSuffix(filename, filepath.Ext(filename))
	stdoutFile := fmt.Sprintf(
		"%s/%s.stdout",
		testDir,
		filepath.Base(testDir),
	)
	if _, err := os.Stat(testDir); err == os.ErrNotExist {
		err := os.Mkdir(testDir, os.ModeDir)
		if err != nil {
			return err
		}
	}

	f, err := os.Create(stdoutFile)
	if err != nil {
		return err
	}
	defer f.Close()

	f.WriteString(with)
	return nil
}
