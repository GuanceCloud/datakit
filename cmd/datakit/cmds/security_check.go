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
	BaseUrl = "https://zhuyun-static-files-production.oss-cn-hangzhou.aliyuncs.com/security-checker/"
)

var (
	SecCheckOsArch = map[string]bool{
		datakit.OSArchLinuxArm:   true,
		datakit.OSArchLinuxArm64: true,
		datakit.OSArchLinuxAmd64: true,
		datakit.OSArchLinux386:   true,
	}
)

type SecCheckVersion struct {
	Version string
}

func InstallSecCheck(installDir string) error {
	osArch := runtime.GOOS + "/" + runtime.GOARCH
	if _, ok := SecCheckOsArch[osArch]; !ok {
		return fmt.Errorf("Security Checker not support in %v\n", osArch)
	}

	fmt.Printf("Start downloading install script...\n")

	verUrl := BaseUrl + "install.sh"
	resp, err := http.Get(verUrl)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("status code %v", resp.StatusCode)
	}

	fmt.Printf("Download install script successfully.\n")

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response body %v", err)
	}

	cmd := exec.Command("/bin/bash", "-c", string(body))
	x, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Security Checker install failed: %s \n", string(x))
		return err
	}
	fmt.Printf("Install Security Checker successfully.\n")

	return nil
}
