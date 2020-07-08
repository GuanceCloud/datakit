// +build !solaris

package tailf

import (
	"fmt"
	"os"
	"testing"
	"time"
)

var (
	root = "/tmp/tailf_test"

	dir = "/tmp/tailf_test/1/2"

	paths = []string{
		"/tmp/tailf_test/zero.txt",
		"/tmp/tailf_test/1/one.txt",
		"/tmp/tailf_test/1/2/two.txt",
	}
)

func TestWrite(t *testing.T) {
	defer func() {
		os.RemoveAll(root)
	}()

	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		panic(err)
	}

	var files []*os.File
	for _, path := range paths {
		f, err := os.Create(path)
		if err != nil {
			panic(err)
		}
		files = append(files, f)
	}
	defer func() {
		for _, f := range files {
			f.Close()
		}
	}()

	count := 0
	for {
		if count > 1000 {
			return
		}
		for index, file := range files {
			file.WriteString(time.Now().Format(time.RFC3339Nano) +
				fmt.Sprintf(" -- index: %d -- count: %d\n", index, count))
			time.Sleep(100 * time.Millisecond)
		}
		count++
	}
}

func TestMain(t *testing.T) {
	var tailer = Tailf{
		Paths:         paths,
		FormBeginning: false,
	}

	go tailer.Run()

	time.Sleep(60 * time.Second)
}

func TestFilterPath(t *testing.T) {
	t.Logf("%s\n", filterPath([]string{root}))
}
