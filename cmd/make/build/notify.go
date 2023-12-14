// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package build

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/export"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/git"
)

//nolint:lll
var (
	k8sDaemonsetTemplete = "wget https://{{.DownloadCDN}}/datakit.yaml"
	NotifyOnly           = false
)

var (
	NotifyToken = ""

	CIOnlineNewVersion = fmt.Sprintf(`
{
	"msgtype": "text",
	"text": {
		"content": "%s 发布了 DataKit 新版本(%s)"
	}
}`, git.Uploader, ReleaseVersion)

	CIPassNotifyMsg = fmt.Sprintf(`
{
	"msgtype": "text",
	"text": {
		"content": "%s 触发的 DataKit CI 通过"
	}
}`, git.Uploader)

	CINotifyStartBuildMsg = fmt.Sprintf(`
{
  "msgtype": "text",
  "text": {
  	"content": "%s 正在执行 DataKit CI 编译，此刻请勿在分支[%s]提交代码"
  }
}`, git.Uploader, git.Branch)

	CINotifyStartPubMsg = fmt.Sprintf(`
{
  "msgtype": "text",
  "text": {
  	"content": "%s 正在执行发布 %s..."
  }
}`, git.Uploader, ReleaseVersion)

	CINotifyStartPubEBpfMsg = fmt.Sprintf(`
{
  "msgtype": "text",
  "text": {
  	"content": "%s 正在执行发布 DataKit eBPF %s..."
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

	// TODO 校验 respbody 中的 errcode，如 310000
	respbody, err := io.ReadAll(resp.Body)
	if err != nil {
		l.Errorf("io.ReadAll: %s", err)
		return
	}

	switch resp.StatusCode / 100 {
	case 2:
		l.Debugf("notify dingding ok(%q): %q", resp.Status, respbody)
		return
	default:
		l.Error(fmt.Errorf(string(respbody)))
	}
}

func NotifyStartPub() {
	if NotifyToken == "" {
		return
	}

	l.Debugf("NotifyStartPub...")
	notify(NotifyToken, bytes.NewBufferString(CINotifyStartPubMsg))
}

func NotifyStartPubEBpf() {
	if NotifyToken == "" {
		return
	}

	l.Debugf("NotifyStartPubEBpf...")
	notify(NotifyToken, bytes.NewBufferString(CINotifyStartPubEBpfMsg))
}

func NotifyStartBuild() {
	if NotifyToken == "" {
		return
	}

	l.Debugf("NotifyStartBuild...")
	notify(NotifyToken, bytes.NewBufferString(CINotifyStartBuildMsg))
}

// NotifyFail send notifications and exit current process.
func NotifyFail(msg string) {
	defer func() {
		os.Exit(-1)
	}()

	if NotifyToken == "" {
		return
	}

	failNotify := fmt.Sprintf(`
{
	"msgtype": "text",
	"text": {
		"content": "%s 触发的 DataKit CI 失败:\n%s"
	}
}`, git.Uploader, msg)

	l.Debugf("NotifyFail...")
	notify(NotifyToken, bytes.NewBufferString(failNotify))
}

func NotifyBuildDone() {
	if NotifyToken == "" {
		return
	}

	l.Debugf("NotifyBuildDone...")
	notify(NotifyToken, bytes.NewBufferString(CIPassNotifyMsg))
}

func NotifyPubDone() {
	if NotifyToken == "" {
		return
	}

	x := struct {
		Uploader, Version, DownloadCDN string
	}{
		Uploader:    git.Uploader,
		Version:     ReleaseVersion,
		DownloadCDN: DownloadCDN,
	}

	switch ReleaseType {
	case ReleaseLocal, ReleaseTesting:

		content := func() []string {
			x := []string{
				fmt.Sprintf(`{{.Uploader}} 发布了 Datakit %d 个平台测试版({{.Version}})`, len(curArchs)),
			}
			for _, arch := range curArchs {
				x = append(x, "--------------------------")
				x = append(x, fmt.Sprintf("%s 安装/升级：", arch))

				p := &export.Params{}

				platform := ""
				switch arch {
				case `linux/386`,
					`linux/amd64`,
					`linux/arm`,
					`linux/arm64`,
					`darwin/amd64`:
					platform = "unix"

				case "windows/amd64",
					"windows/386":
					platform = "windows"
				}

				x = append(x, export.InstallCommand(
					p.WithPlatform(platform),
					p.WithSourceURL("https://"+DownloadCDN),
					p.WithJSON(true),
					p.WithVersion("-"+ReleaseVersion),
				).String())

				x = append(x, "\n")

				x = append(x, export.InstallCommand(
					p.WithUpgrade(true),
					p.WithPlatform(platform),
					p.WithSourceURL("https://"+DownloadCDN),
					p.WithJSON(true),
					p.WithVersion("-"+ReleaseVersion),
				).String())
			}

			// under testing release, add k8s daemonset yaml
			if ReleaseType == ReleaseTesting {
				x = append(x, "--------------------------")
				x = append(x, "Kubernetes DaemonSet 安装")
				x = append(x, k8sDaemonsetTemplete)
			}

			return x
		}()

		notifyNewVersion := fmt.Sprintf(`
{
	"msgtype": "text",
	"text": {
		"content": "%s"
	}
}`, strings.Join(content, "\n"))

		var buf bytes.Buffer
		t, err := template.New("").Parse(notifyNewVersion)
		if err != nil {
			l.Errorf("template.New", err)
		}

		if err := t.Execute(&buf, x); err != nil {
			l.Fatal(err)
		}

		l.Debugf("NotifyPubDone...")
		notify(NotifyToken, &buf)

	case ReleaseProduction:

		l.Debugf("NotifyPubDone for release...")
		notify(NotifyToken, bytes.NewBufferString(CIOnlineNewVersion))
	}
}

func NotifyPubEBpfDone() {
	if NotifyToken == "" {
		return
	}

	x := struct {
		Uploader, Version, DownloadCDN string
	}{
		Uploader:    git.Uploader,
		Version:     ReleaseVersion,
		DownloadCDN: DownloadCDN,
	}

	switch ReleaseType {
	case ReleaseLocal, ReleaseTesting:

		content := func() []string {
			x := []string{
				fmt.Sprintf(`{{.Uploader}} 发布了 DataKit eBPF %d 个平台测试版({{.Version}})。`, len(curEBpfArchs)),
			}
			for _, arch := range curEBpfArchs {
				x = append(x, "--------------------------")
				x = append(x, fmt.Sprintf("%s 下载地址：", arch))
				x = append(x, "https://"+filepath.Join(DownloadCDN, fmt.Sprintf(
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

		l.Debugf("NotifyPubEBpfDone...")
		notify(NotifyToken, &buf)
	case ReleaseProduction:

		l.Debugf("NotifyPubEBpfDone for release...")
		notify(NotifyToken, bytes.NewBufferString(fmt.Sprintf(`
		{
			"msgtype": "text",
			"text": {
				"content": "%s 发布了 DataKit eBPF %s 新版本(%s)"
			}
		}`, git.Uploader, strings.Join(curEBpfArchs, ", "), ReleaseVersion)))
	}
}
