// +build !solaris

package tailf

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"sync"
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

	var tf = Tailf{
		Config: Config{
			Subscribes: []Subscribe{
				Subscribe{
					File:          tmpfile.Name(),
					FormBeginning: true,
					Pipe:          false,
					WatchMethod:   "inotify",
					Measurement:   "tailf_measurement",
				},
			},
		},
	}

	tf.ctx, tf.cancel = context.WithCancel(context.Background())
	tf.wg = new(sync.WaitGroup)

	for _, sub := range tf.Config.Subscribes {
		tf.wg.Add(1)
		s := sub
		stream := newStream(&s, &tf)
		fmt.Println(s)
		go stream.start(tf.wg)
	}

	time.Sleep(10 * time.Second)
	stopch <- struct{}{}
	tf.Stop()
}
