//+build windows

package telegrafwrap

import (
	"fmt"
	"os"
)

func KillProcess(pid int) error {

	prs, err := os.FindProcess(pid)
	if err != nil {
		return err
	}

	if prs != nil {
		if err = prs.Kill(); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("cannot find process %d", pid)
	}

	return nil
}
