// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package build

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/git"
)

//nolint:lll
var (
	installNotifyTemplate = map[string]string{
		`linux/386`:   `DK_DATAWAY=\"https://openway.guance.com?token=<TOKEN>\" bash -c \"$(curl -L https://{{.DownloadAddr}}/install-{{.Version}}.sh)\"`,
		`linux/amd64`: `DK_DATAWAY=\"https://openway.guance.com?token=<TOKEN>\" bash -c \"$(curl -L https://{{.DownloadAddr}}/install-{{.Version}}.sh)\"`,
		`linux/arm`:   `DK_DATAWAY=\"https://openway.guance.com?token=<TOKEN>\" bash -c \"$(curl -L https://{{.DownloadAddr}}/install-{{.Version}}.sh)\"`,
		`linux/arm64`: `DK_DATAWAY=\"https://openway.guance.com?token=<TOKEN>\" bash -c \"$(curl -L https://{{.DownloadAddr}}/install-{{.Version}}.sh)\"`,

		`darwin/amd64`: `DK_DATAWAY=\"https://openway.guance.com?token=<TOKEN>\" bash -c \"$(curl -L https://{{.DownloadAddr}}/install-{{.Version}}.sh)\"`,

		`windows/amd64`: `$env:DK_DATAWAY=\"https://openway.guance.com?token=<TOKEN>\";Set-ExecutionPolicy Bypass -scope Process -Force; Import-Module bitstransfer; start-bitstransfer -source https://{{.DownloadAddr}}/install-{{.Version}}.ps1 -destination .install.ps1; powershell .install.ps1;`,
		`windows/386`:   `$env:DK_DATAWAY=\"https://openway.guance.com?token=<TOKEN>\";Set-ExecutionPolicy Bypass -scope Process -Force; Import-Module bitstransfer; start-bitstransfer -source https://{{.DownloadAddr}}/install-{{.Version}}.ps1 -destination .install.ps1; powershell .install.ps1;`,
	}

	upgradeNotifyTemplate = map[string]string{
		`linux/386`:   `DK_UPGRADE=1 bash -c \"$(curl -L https://{{.DownloadAddr}}/install-{{.Version}}.sh)\"`,
		`linux/amd64`: `DK_UPGRADE=1 bash -c \"$(curl -L https://{{.DownloadAddr}}/install-{{.Version}}.sh)\"`,
		`linux/arm`:   `DK_UPGRADE=1 bash -c \"$(curl -L https://{{.DownloadAddr}}/install-{{.Version}}.sh)\"`,
		`linux/arm64`: `DK_UPGRADE=1 bash -c \"$(curl -L https://{{.DownloadAddr}}/install-{{.Version}}.sh)\"`,

		`darwin/amd64`: `DK_UPGRADE=1 bash -c \"$(curl -L https://{{.DownloadAddr}}/install-{{.Version}}.sh)\"`,

		`windows/amd64`: `$env:DK_UPGRADE=\"1\"; Set-ExecutionPolicy Bypass -scope Process -Force; Import-Module bitstransfer; start-bitstransfer -source https://{{.DownloadAddr}}/install-{{.Version}}.ps1 -destination .install.ps1; powershell .install.ps1;`,
		`windows/386`:   `$env:DK_UPGRADE=\"1\"; Set-ExecutionPolicy Bypass -scope Process -Force; Import-Module bitstransfer; start-bitstransfer -source https://{{.DownloadAddr}}/install-{{.Version}}.ps1 -destination .install.ps1; powershell .install.ps1;`,
	}

	k8sDaemonsetTemplete = "wget https://{{.DownloadAddr}}/datakit.yaml"
)

var (
	NotifyToken = ""

	CIOnlineNewVersion = fmt.Sprintf(`
{
	"msgtype": "text",
	"text": {
		"content": "%s ????????? DataKit ?????????(%s)"
	}
}`, git.Uploader, ReleaseVersion)

	CIPassNotifyMsg = fmt.Sprintf(`
{
	"msgtype": "text",
	"text": {
		"content": "%s ????????? DataKit CI ??????"
	}
}`, git.Uploader)

	CINotifyStartBuildMsg = fmt.Sprintf(`
{
  "msgtype": "text",
  "text": {
  	"content": "%s ???????????? DataKit CI ??????????????????????????????[%s]????????????"
  }
}`, git.Uploader, git.Branch)

	CINotifyStartPubMsg = fmt.Sprintf(`
{
  "msgtype": "text",
  "text": {
  	"content": "%s ?????????????????? %s..."
  }
}`, git.Uploader, ReleaseVersion)

	CINotifyStartPubEBpfMsg = fmt.Sprintf(`
{
  "msgtype": "text",
  "text": {
  	"content": "%s ?????????????????? DataKit eBPF %s..."
  }
}`, git.Uploader, ReleaseVersion)
)

func notify(tkn string, body io.Reader) {
	req, err := http.NewRequest("POST", "https://oapi.dingtalk.com/robot/send?access_token="+tkn, body)
	if err != nil {
		l.Errorf("NewRequest: %s", err.Error())
		return
	}

	req.Header.Set("Content-Type", "application/json")
	cli := http.Client{}

	resp, err := cli.Do(req)
	if err != nil {
		l.Errorf("notify: %s", err)
		return
	}

	defer resp.Body.Close() //nolint:errcheck

	// TODO ?????? respbody ?????? errcode?????? 310000
	respbody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		l.Errorf("ioutil.ReadAll: %s", err)
		return
	}

	switch resp.StatusCode / 100 {
	case 2:
		return
	default:
		l.Error(fmt.Errorf(string(respbody)))
	}
}

func NotifyStartPub() {
	if NotifyToken == "" {
		return
	}
	notify(NotifyToken, bytes.NewBufferString(CINotifyStartPubMsg))
}

func NotifyStartPubEBpf() {
	if NotifyToken == "" {
		return
	}
	notify(NotifyToken, bytes.NewBufferString(CINotifyStartPubEBpfMsg))
}

func NotifyStartBuild() {
	if NotifyToken == "" {
		return
	}
	notify(NotifyToken, bytes.NewBufferString(CINotifyStartBuildMsg))
}

func NotifyFail(msg string) {
	if NotifyToken == "" {
		return
	}

	failNotify := fmt.Sprintf(`
{
	"msgtype": "text",
	"text": {
		"content": "%s ????????? DataKit CI ??????:\n%s"
	}
}`, git.Uploader, msg)
	notify(NotifyToken, bytes.NewBufferString(failNotify))
}

func NotifyBuildDone() {
	if NotifyToken == "" {
		return
	}
	notify(NotifyToken, bytes.NewBufferString(CIPassNotifyMsg))
}

func NotifyPubDone() {
	if NotifyToken == "" {
		return
	}

	x := struct {
		Uploader, Version, DownloadAddr string
	}{
		Uploader:     git.Uploader,
		Version:      ReleaseVersion,
		DownloadAddr: DownloadAddr,
	}

	switch ReleaseType {
	case ReleaseLocal, ReleaseTesting:

		content := func() []string {
			x := []string{
				fmt.Sprintf(`{{.Uploader}} ????????????????????? DataKit %d ??????????????????({{.Version}})???`, len(curArchs)),
			}
			for _, arch := range curArchs {
				x = append(x, "--------------------------")
				x = append(x, fmt.Sprintf("%s ??????/?????????", arch))
				x = append(x, installNotifyTemplate[arch])
				x = append(x, "\n")
				x = append(x, upgradeNotifyTemplate[arch])
			}

			// under testing release, add k8s daemonset yaml
			if ReleaseType == ReleaseTesting {
				x = append(x, "--------------------------")
				x = append(x, "Kubernetes DaemonSet ??????")
				x = append(x, k8sDaemonsetTemplete)
			}

			return x
		}()

		CINotifyNewVersion := fmt.Sprintf(`
{
	"msgtype": "text",
	"text": {
		"content": "%s"
		}
}`, strings.Join(content, "\n"))

		var buf bytes.Buffer
		t, err := template.New("").Parse(CINotifyNewVersion)
		if err != nil {
			l.Fatal(err)
		}

		if err := t.Execute(&buf, x); err != nil {
			l.Fatal(err)
		}
		notify(NotifyToken, &buf)
	case ReleaseProduction:
		notify(NotifyToken, bytes.NewBufferString(CIOnlineNewVersion))
	}
}

func NotifyPubEBpfDone() {
	if NotifyToken == "" {
		return
	}

	x := struct {
		Uploader, Version, DownloadAddr string
	}{
		Uploader:     git.Uploader,
		Version:      ReleaseVersion,
		DownloadAddr: DownloadAddr,
	}

	switch ReleaseType {
	case ReleaseLocal, ReleaseTesting:

		content := func() []string {
			x := []string{
				fmt.Sprintf(`{{.Uploader}} ????????????????????? DataKit eBPF %d ??????????????????({{.Version}})???`, len(curEBpfArchs)),
			}
			for _, arch := range curEBpfArchs {
				x = append(x, "--------------------------")
				x = append(x, fmt.Sprintf("%s ???????????????", arch))
				x = append(x, "https://"+filepath.Join(DownloadAddr, fmt.Sprintf(
					"datakit-ebpf-%s-%s-%s.tar.gz", runtime.GOOS, runtime.GOARCH, ReleaseVersion)))
			}
			return x
		}()

		CINotifyNewEBpfVersion := fmt.Sprintf(`
{
	"msgtype": "text",
	"text": {
		"content": "%s"
		}
}`, strings.Join(content, "\n"))

		var buf bytes.Buffer
		t, err := template.New("").Parse(CINotifyNewEBpfVersion)
		if err != nil {
			l.Fatal(err)
		}

		if err := t.Execute(&buf, x); err != nil {
			l.Fatal(err)
		}
		notify(NotifyToken, &buf)
	case ReleaseProduction:
		notify(NotifyToken, bytes.NewBufferString(fmt.Sprintf(`
		{
			"msgtype": "text",
			"text": {
				"content": "%s ????????? DataKit eBPF %s ?????????(%s)"
			}
		}`, git.Uploader, strings.Join(curEBpfArchs, ", "), ReleaseVersion)))
	}
}
