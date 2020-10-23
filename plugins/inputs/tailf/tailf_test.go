// +build !solaris

package tailf

import (
	"fmt"
	"io/ioutil"
	//"os"
	"testing"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
)

func writeFile() (filepath string) {
	file, err := ioutil.TempFile("", "example")
	if err != nil {
		panic(err)
	}

	count := 0
	go func() {
		for {
			select {
			case <-datakit.Exit.Wait():
				// os.Remove(file.Name())
				return
			default:
				str := fmt.Sprintf("%s -- %d\n%s\n%s\n%s\n",
					time.Now().Format(time.RFC3339Nano), count,
					"   SPACE", "\rRETURN", "\tTAB")
				file.WriteString(str)
				time.Sleep(200 * time.Millisecond)
				count++
			}
		}
	}()

	return file.Name()
}

func TestMain(t *testing.T) {
	io.TestOutput()

	w1 := writeFile()
	time.Sleep(time.Second)

	var tailer = Tailf{
		LogFiles:      []string{w1},
		FromBeginning: true,
		Source:        "Testing",
	}

	go tailer.Run()

	time.Sleep(time.Second * 15)
	datakit.Exit.Close()
	time.Sleep(time.Second)
}
