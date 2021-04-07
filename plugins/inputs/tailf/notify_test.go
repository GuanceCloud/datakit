package tailf

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

func TestWatching(t *testing.T) {
	tempFile, err := ioutil.TempFile("", "")
	if err != nil {
		t.Fatal(err)
	}
	file := tempFile.Name()

	watcher, err := NewWatcher()
	if err != nil {
		t.Fatal(err)
	}
	defer watcher.Close()

	notifyChan := make(chan notifyType)

	watcher.Add(file, notifyChan)
	t.Logf("watcher add file %s", file)

	go watcher.Watching(datakit.Exit.Wait())

	go func() {
		for {
			select {
			case notify := <-notifyChan:
				switch notify {
				case renameNotify:
					t.Log("receive rename notify")
				default:
					t.Log("panic")
				}
			}
		}
	}()

	time.Sleep(time.Second * 2)
	if err := os.Rename(file, file+".bak"); err != nil {
		t.Error(err)
	}
	defer os.Remove(file + ".bak")

	time.Sleep(time.Second * 2)
	datakit.Exit.Close()
}
