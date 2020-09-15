// +build !solaris

package tailf

import (
	"fmt"
	"os"
	"testing"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
)

var (
	logFiles = []string{
		//"/tmp/tailf_test/**/*.*",
		"/tmp/tailf_test/a123/ab.log",
		//"/tmp/tailf_test/a123/*.txt",
		//"/tmp/tailf_test/*.no",
	}

	ignore = []string{
		//"/tmp/tailf_test/a123/ab.log",
	}

	deepDir = "/tmp/tailf_test/a123/b123/"

	paths = []string{
		//"/tmp/tailf_test/a123/b123/1234.log",
		//"/tmp/tailf_test/a123/b123/5678.txt",
		"/tmp/tailf_test/a123/ab.log",
		//"/tmp/tailf_test/a123/cd.txt",
	}
)

func TestWrite(t *testing.T) {
	if err := os.MkdirAll(deepDir, os.ModePerm); err != nil {
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
		select {
		case <-datakit.Exit.Wait():
			return
		default:
			for index, file := range files {
				file.WriteString(time.Now().Format(time.RFC3339Nano) +
					fmt.Sprintf(" -- index: %d -- count: %d\n", index, count))
				time.Sleep(200 * time.Millisecond)
			}
			count++
		}
	}
}

func TestMain(t *testing.T) {
	io.TestOutput()
	go TestWrite(t)

	var tailer = Tailf{
		LogFiles: logFiles,
		Ignore:   ignore,
		Source:   "NAXXRAMAS",
		Tags:     map[string]string{"TestKey": "TestValue"},
	}

	time.Sleep(time.Second)
	go tailer.Run()

	time.Sleep(30 * time.Second)
	datakit.Exit.Close()
}

func TestFileList(t *testing.T) {
	t.Log(getFileList(logFiles, ignore))
}
