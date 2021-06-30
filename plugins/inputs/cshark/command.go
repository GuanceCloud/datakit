package cshark

import (
	"bytes"
	"fmt"
	"github.com/blang/semver"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"io"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"sync"
	"syscall"
	"time"
)

var TSharkVersionUnknown = fmt.Errorf("Could not determine version of tshark")

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

func TSharkVersionFromOutput(output string) (semver.Version, error) {
	var ver = regexp.MustCompile(`^TShark .*?(\d+\.\d+\.\d+)`)
	res := ver.FindStringSubmatch(output)

	if len(res) > 0 {
		if v, err := semver.Make(res[1]); err == nil {
			return v, nil
		} else {
			return semver.Version{}, err
		}
	}

	return semver.Version{}, TSharkVersionUnknown
}

func TSharkVersion(tsharkBin string) (semver.Version, error) {
	cmd := exec.Command(tsharkBin, "-v")

	cmdOutput := &bytes.Buffer{}
	cmd.Stdout = cmdOutput
	cmd.Run() // don't check error - older versions return error code 1. Just search output.
	output := cmdOutput.Bytes()

	return TSharkVersionFromOutput(string(output))
}
