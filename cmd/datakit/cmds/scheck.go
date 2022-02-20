package cmds

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os/exec"
	"runtime"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

const (
	BaseURL = "https://zhuyun-static-files-production.oss-cn-hangzhou.aliyuncs.com/security-checker/"
)

var ScheckOsArch = map[string]bool{
	datakit.OSArchLinuxArm:   true,
	datakit.OSArchLinuxArm64: true,
	datakit.OSArchLinuxAmd64: true,
	datakit.OSArchLinux386:   true,
}

type SecCheckVersion struct {
	Version string
}

func installScheck() error {
	osArch := runtime.GOOS + "/" + runtime.GOARCH
	if _, ok := ScheckOsArch[osArch]; !ok {
		return fmt.Errorf("security checker not support in %v", osArch)
	}

	infof("Start downloading install script...\n")

	verURL := BaseURL + "install.sh"
	cli := getcli()

	req, err := http.NewRequest("GET", verURL, nil)
	if err != nil {
		return err
	}

	resp, err := cli.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status code %v", resp.StatusCode)
	}

	infof("Download install script successfully.\n")

	defer resp.Body.Close() //nolint:errcheck
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response body %w", err)
	}

	// TODO: add network proxy option
	cmd := exec.Command("/bin/bash", "-c", string(body)) //nolint:gosec
	if _, err = cmd.CombinedOutput(); err != nil {
		return err
	}

	infof("Install Security Checker successfully.\n")

	return nil
}
