package tailer

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

const (
	dir        = "/tmp"
	pattern    = "testfile"
	prefixName = dir + "/" + pattern
	num        = 10
)

var testfiles []*os.File = func() []*os.File {
	var list []*os.File
	for i := 0; i < num; i++ {
		f, err := ioutil.TempFile(dir, pattern)
		if err == nil {
			list = append(list, f)
		}
	}
	return list
}()

func TestFileList(t *testing.T) {
	defer func() {
		for _, f := range testfiles {
			f.Close()
			os.Remove(f.Name())
		}
	}()

	var names []string
	for _, f := range testfiles {
		names = append(names, f.Name())
	}

	list := NewFileList([]string{prefixName + "*"}).List()

	if len(names) != len(list) {
		t.Error()
	}

	for index, name := range names {
		func() {
			for idx, res := range list {
				if name == res {
					t.Logf("match success [%d]%s : [%d]%s", index, name, idx, res)
					return
				}
			}
			t.Errorf("not match [%d] filename %s", index, name)
		}()
	}

	t.Logf("expect list: %v", names)
	t.Logf("result list: %v", list)
}

func TestFileListIgnore(t *testing.T) {
	defer func() {
		for _, f := range testfiles {
			f.Close()
			os.Remove(f.Name())
		}
	}()

	var names []string
	for _, f := range testfiles {
		names = append(names, f.Name())
	}

	// 屏蔽所有数字 6 开头的文件
	list := NewFileList([]string{prefixName + "*"}).Ignore([]string{prefixName + "6*"}).List()

	if len(names) != len(list) {
		t.Error()
	}

	for index, name := range names {
		if strings.Index(name, prefixName+"6") != -1 {
			t.Logf("ignore success [%d]%s", index, name)
			continue
		}
		func() {
			for idx, res := range list {
				if name == res {
					t.Logf("match success [%d]%s : [%d]%s", idx, res, index, name)
					return
				}
			}
			t.Errorf("not match [%d] filename %s", index, name)
		}()
	}

	t.Logf("expect list: %v", names)
	t.Logf("result list: %v", list)
}
