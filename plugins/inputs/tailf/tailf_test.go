// +build !solaris

package tailf

import (
	"fmt"
	"os"
	"testing"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
)

var (
	logFiles = []string{
		"/tmp/tailf_test/a*/b*/*.log",
		"/tmp/tailf_test/a*/b*/*.txt",
	}

	ignore = []string{
		"/tmp/tailf_test/a*/b*/12345.log",
	}

	deepDir = "/tmp/tailf_test/a123/b123/"

	paths = []string{
		"/tmp/tailf_test/a123/b123/1234.log",
		"/tmp/tailf_test/a123/b123/5678.txt",
	}
)

func __init() {
	logger.SetGlobalRootLogger("", logger.DEBUG, logger.OPT_DEFAULT)
	l = logger.SLogger(inputName)
}

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
		if count > 1000 {
			return
		}
		for index, file := range files {
			file.WriteString(time.Now().Format(time.RFC3339Nano) +
				fmt.Sprintf(" -- index: %d -- count: %d\n", index, count))
			time.Sleep(200 * time.Millisecond)
		}
		count++
	}
}

func TestMain(t *testing.T) {
	__init()
	testAssert = true

	var tailer = Tailf{
		LogFiles: logFiles,
		Ignore:   ignore,
		Source:   "NAXXRAMAS",
	}

	go tailer.Run()

	time.Sleep(90 * time.Second)
}

func TestFileList(t *testing.T) {

	t.Log(getFileList(logFiles, ignore))
}
