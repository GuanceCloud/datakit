// +build !solaris

package tailf

import (
	"os"
	"testing"
	"time"
)

func TestWrite(t *testing.T) {
	defer func() {
		os.RemoveAll("/tmp/tailf_test")
	}()

	if err := os.MkdirAll("/tmp/tailf_test/1/2", os.ModePerm); err != nil {
		panic(err)
	}

	var paths = []string{"/tmp/tailf_test/zero.txt", "/tmp/tailf_test/1/one.txt", "/tmp/tailf_test/1/2/two.txt"}
	// var paths = []string{"/tmp/tailf_test/zero.txt"}

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

	for {
		// for _, f := range files {
		files[0].WriteString(time.Now().Format(time.RFC3339Nano) + " -- 1111111111\n")
		files[1].WriteString(time.Now().Format(time.RFC3339Nano) + " -- 2222222222\n")
		files[2].WriteString(time.Now().Format(time.RFC3339Nano) + " -- 3333333333\n")
		// }
		time.Sleep(100 * time.Millisecond)
	}
}

func TestStart(t *testing.T) {
	var paths = []string{"/tmp/tailf_test/zero.txt", "/tmp/tailf_test/1/one.txt", "/tmp/tailf_test/1/2/two.txt"}
	// var paths = []string{"/tmp/tailf_test/zero.txt"}

	var tailer = Tailf{
		Paths:         paths,
		FormBeginning: false,
	}

	time.Sleep(2 * time.Second)
	go tailer.Run()

	time.Sleep(100 * time.Second)
	time.Sleep(2 * time.Second)
}

func TestCheckPaths(t *testing.T) {
	paths := []string{"/tmp/tailf_test"}
	t.Logf("%v\n", paths)

	for _, p := range filterPath(paths) {
		t.Logf("%s\n", p)
	}
}
