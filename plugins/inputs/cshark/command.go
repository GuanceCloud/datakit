package cshark

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
	"time"
	"runtime"
	"syscall"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

type Command struct {
	sync.Mutex
	*exec.Cmd
}

func (c *Command) String() string {
	c.Lock()
	defer c.Unlock()
	return fmt.Sprintf("%v %v", c.Cmd.Path, c.Cmd.Args)
}

func (c *Command) Start() error {
	c.Lock()
	defer c.Unlock()
	res := c.Cmd.Start()
	return res
}

func (c *Command) Wait() error {
	return c.Cmd.Wait()
}

func (c *Command) StdoutReader() (io.ReadCloser, error) {
	c.Lock()
	defer c.Unlock()
	return c.Cmd.StdoutPipe()
}

func (c *Command) SetStdout(w io.Writer) {
	c.Lock()
	defer c.Unlock()
	c.Cmd.Stdout = w
}

// If stdout supports Close(), call it. If stdout is a pipe, for example,
// this can be used to have EOF appear on the reading side (e.g. tshark -T psml)
func (c *Command) Close() error {
	c.Lock()
	defer c.Unlock()
	if cl, ok := c.Cmd.Stdout.(io.Closer); ok {
		return cl.Close()
	}
	return nil
}

func (c *Command) Pid() int {
	c.Lock()
	defer c.Unlock()
	if c.Cmd.Process == nil {
		return -1
	}
	return c.Cmd.Process.Pid
}

func RunCommand(commandStr string) *Command {
	cmd := exec.Command("sh", "-c", commandStr)

	res := &Command{
		Cmd: cmd,
	}

	return res
}

func (c *Command) MonitProc() error {
	tick := time.NewTicker(time.Second)
	defer tick.Stop()

	for {
		select {
		case <-tick.C:
			p, err := os.FindProcess(c.Pid())
			if err != nil {
				continue
			}

			switch runtime.GOOS {
			case "windows":

			default:
				if err := p.Signal(syscall.Signal(0)); err != nil {
					return err
				}
			}

		case <-datakit.Exit.Wait():
			if err := c.Cmd.Process.Kill(); err != nil { // XXX: should we wait here?
				return err
			}

			return nil
		}
	}
}

