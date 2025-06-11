// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package build

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/export"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/git"
)

//nolint:lll
var (
	k8sDaemonsetTemplete         = "wget https://{{.DownloadCDN}}/datakit-{{.Version}}.yaml"
	k8sDeploymentELinkerTemplete = "wget https://{{.DownloadCDN}}/datakit-elinker-{{.Version}}.yaml"
	NotifyOnly                   = false
)

var (
	NotifyToken = ""

	CIOnlineNewVersion = fmt.Sprintf(`
{
	"msg_type": "text",
	"content": {
		"text": "%s 发布了 DataKit 新版本(%s)"
	}
}`, git.Uploader, ReleaseVersion)

	CINotifyStartBuildMsg = fmt.Sprintf(`
{
  "msg_type": "text",
  "content": {
  	"text": "%s 正在执行 DataKit CI 编译，此刻请勿在分支[%s]提交代码"
  }
}`, git.Uploader, git.Branch)

	CINotifyStartPubMsg = fmt.Sprintf(`
{
  "msg_type": "text",
  "content": {
		"text": "%s 正在执行发布 DataKit:%s..."
  }
}`, git.Uploader, ReleaseVersion)

	CINotifyStartPubEBpfMsg = fmt.Sprintf(`
{
  "msg_type": "text",
  "content": {
  	"text": "%s 正在执行发布 DataKit eBPF %s..."
  }
}`, git.Uploader, ReleaseVersion)
)

type content struct {
	Text string `json:"text"`
}

type textMsg struct {
	MsgType string  `json:"msg_type"`
	Content content `json:"content"`
}

func doNotify(tkn string, body string) {
	l.Infof("robot notify message: %s", body)

	req, err := http.NewRequest("POST", "https://open.feishu.cn/open-apis/bot/v2/hook/"+tkn, bytes.NewBufferString(body))
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

	respbody, err := io.ReadAll(resp.Body)
	if err != nil {
		l.Errorf("io.ReadAll: %s", err)
		return
	}

	switch resp.StatusCode / 100 {
	case 2:
		l.Debugf("notify robot ok(%q): %q", resp.Status, respbody)
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
	doNotify(NotifyToken, CINotifyStartPubMsg)
}

func NotifyStartPubEBpf() {
	if NotifyToken == "" {
		return
	}

	l.Debugf("NotifyStartPubEBpf...")
	doNotify(NotifyToken, CINotifyStartPubEBpfMsg)
}

func NotifyStartBuild() {
	if NotifyToken == "" {
		return
	}

	l.Debugf("NotifyStartBuild...")
	doNotify(NotifyToken, CINotifyStartBuildMsg)
}

// NotifyFail send notifications and exit current process.
func NotifyFail(msg string) {
	defer func() {
		os.Exit(-1)
	}()

	if NotifyToken == "" {
		return
	}

	tm := textMsg{
		Content: content{
			Text: fmt.Sprintf("%s 触发的 DataKit CI 失败:\n%s", git.Uploader, msg),
		},
		MsgType: "text",
	}

	j, err := json.Marshal(tm)
	if err != nil {
		l.Fatal(err.Error())
	}

	l.Debugf("NotifyFail...")
	doNotify(NotifyToken, string(j))
}

func buildNotifyContent(ver, cdn, release string, archs []string) string {
	x := []string{
		fmt.Sprintf(`{{.Uploader}} 发布了 Datakit %d 个平台测试版({{.Version}})`, len(archs)),
	}

	for _, arch := range archs {
		x = append(x, "") // empty line
		x = append(x, fmt.Sprintf("### %s 安装/升级：", arch))

		p := &export.Params{}

		platform := ""
		switch arch {
		case `linux/386`,
			`linux/amd64`,
			`linux/arm`,
			`linux/arm64`,
			`darwin/amd64`:
			platform = "unix"
			x = append(x, "``` shell")

		case "windows/amd64",
			"windows/386":
			platform = "windows"
			x = append(x, "``` powershell")

		default: // other platform not support
			return ""
		}

		x = append(x, "# install")

		x = append(x, export.InstallCommand(
			p.WithPlatform(platform),
			p.WithSourceURL("https://"+cdn),
			p.WithJSON(true),
			p.WithVersion("-"+ver),
		).String())

		x = append(x, "") // empty line between install and upgrade command

		x = append(x, "# upgrade")
		x = append(x, export.InstallCommand(
			p.WithUpgrade(true),
			p.WithPlatform(platform),
			p.WithSourceURL("https://"+cdn),
			p.WithJSON(true),
			p.WithVersion("-"+ver),
		).String())

		x = append(x, "```")
	}

	// under testing release, add k8s daemonset yaml
	if release == ReleaseTesting {
		x = append(x, "") // empty line
		x = append(x, "### Kubernetes DaemonSet 安装")
		x = append(x, k8sDaemonsetTemplete)

		x = append(x, "") // empty line
		x = append(x, "### Kubernetes Datakit ELinker Deployment 安装")
		x = append(x, k8sDeploymentELinkerTemplete)
	}

	// for lambda extension
	x = append(x, "") // empty line
	x = append(x, `### AWS Lambda extension`)
	for _, arch := range archs {
		parts := strings.Split(arch, "/")
		if len(parts) != 2 {
			l.Fatalf(fmt.Sprintf("invalid arch: %s", arch))
		}

		goos, goarch := parts[0], parts[1]

		if goos == "windows" { // lambda extension not available under windows
			continue
		}

		x = append(x, fmt.Sprintf("- %s/%s 下载：%s", goos, goos,
			"https://"+filepath.Join(DownloadCDN,
				fmt.Sprintf("datakit_aws_extension-%s-%s-%s.zip", goos, goarch, ReleaseVersion))))
	}

	return strings.Join(x, "\n")
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

		tm := textMsg{
			Content: content{
				Text: buildNotifyContent(ReleaseVersion, DownloadCDN, ReleaseType, curArchs),
			},
			MsgType: "text",
		}

		j, err := json.Marshal(tm)
		if err != nil {
			l.Fatal(err.Error())
		}

		var buf bytes.Buffer
		t, err := template.New("").Parse(string(j))
		if err != nil {
			l.Errorf("template.New", err)
		}

		if err := t.Execute(&buf, x); err != nil {
			l.Fatal(err)
		}

		l.Debugf("NotifyPubDone...")
		doNotify(NotifyToken, buf.String())

	case ReleaseProduction:
		l.Debugf("NotifyPubDone for release...")
		doNotify(NotifyToken, CIOnlineNewVersion)
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
				parts := strings.Split(arch, "/")
				if len(parts) != 2 {
					l.Fatalf(fmt.Sprintf("invalid arch: %s", arch))
				}

				goos, goarch := parts[0], parts[1]

				x = append(x, "--------------------------")
				x = append(x, fmt.Sprintf("%s 下载地址：", arch))
				x = append(x, "https://"+filepath.Join(DownloadCDN, fmt.Sprintf(
					"datakit-ebpf-%s-%s-%s.tar.gz", goos, goarch, ReleaseVersion)))
			}
			return x
		}()

		CINotifyNewEBpfVersion := fmt.Sprintf(`
{
	"msg_type": "text",
	"content": {
		"text": "%s"
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
		doNotify(NotifyToken, buf.String())
	case ReleaseProduction:

		l.Debugf("NotifyPubEBpfDone for release...")
		doNotify(NotifyToken, fmt.Sprintf(`
		{
			"msg_type": "text",
			"content": {
				"text": "%s 发布了 DataKit eBPF %s 新版本(%s)"
			}
		}`, git.Uploader, strings.Join(curEBpfArchs, ", "), ReleaseVersion))
	}
}
