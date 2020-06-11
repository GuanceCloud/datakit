//+build !windows

package hostobject

import (
	"bufio"
	"bytes"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
)

var (
	osReleaseFile = "/etc/os-release"

	rePrettyName  = regexp.MustCompile(`^PRETTY_NAME=(.*)$`)
	reDescription = regexp.MustCompile(`^Description:(.*)$`)
)

func getOSInfo() *osInfo {

	oi := &osInfo{
		OSType: runtime.GOOS,
		Arch:   runtime.GOARCH,
	}

	cmd := exec.Command("lsb_release", "-a")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err == nil {
		s := bufio.NewScanner(&out)
		for s.Scan() {
			if m := reDescription.FindStringSubmatch(s.Text()); m != nil && len(m) > 1 {
				oi.Release = strings.TrimSpace(m[1])
				break
			}
		}
		return oi
	}

	if _, err := os.Stat(osReleaseFile); os.IsNotExist(err) {
		releaseFile := `/etc/centos-release`
		if _, err := os.Stat(`/etc/centos-release`); os.IsNotExist(err) {
			releaseFile = `/etc/redhat-release`
			if _, err := os.Stat(osReleaseFile); os.IsNotExist(err) {
				releaseFile = ""
			}
		}

		if releaseFile != "" {
			data, err := ioutil.ReadFile(releaseFile)
			if err == nil {
				oi.Release = strings.TrimSpace(string(data))
			}
		}
		return oi
	}

	f, err := os.Open(osReleaseFile)
	if err != nil {
		return oi
	}
	defer f.Close()

	s := bufio.NewScanner(f)
	for s.Scan() {
		if m := rePrettyName.FindStringSubmatch(s.Text()); m != nil && len(m) > 1 {
			oi.Release = strings.Trim(m[1], `"`)
			break
		}
	}

	return oi
}
