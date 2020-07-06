// +build !solaris

package tailf

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"
)

func TestStart(t *testing.T) {

	tmpfile, err := ioutil.TempFile("", "")
	if err != nil {
		panic(err)
	}
	defer os.Remove(tmpfile.Name())
	fmt.Println(tmpfile.Name())

	var stopch = make(chan struct{})
	go func() {
		defer tmpfile.Close()
		for i := 0; i < 10000; i++ {
			select {
			case <-stopch:
				return
			default:
				_, _ = tmpfile.Write([]byte(fmt.Sprintf("this is logger %d\n", i)))
				time.Sleep(200 * time.Millisecond)
			}
		}

	}()

	var tailer = Tailf{
		Config: struct {
			File          string `toml:"filename"`
			FormBeginning bool   `toml:"from_beginning"`
			Pipe          bool   `toml:"pipe"`
			WatchMethod   string `toml:"watch_method"`
			Measurement   string `toml:"source"`
		}{tmpfile.Name(), true, false, "inotify", "tailf_measurement"},
		offset: 0,
	}

	tailer.Run()
}
