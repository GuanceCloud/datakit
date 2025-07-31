// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2024-present Guance, Inc.

// Package remotejob is running GuanCe remote job.
package remotejob

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/dataway"

	"github.com/GuanceCloud/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
)

var (
	jvmDumpHostScriptFile = "jvm_dump_host_script.py"
	jvmDumpK8SScriptFile  = "jvm_dump_k8s_script.py"
	scriptTempPath        = filepath.Join(datakit.TemplateDir, "service-task")
	jvmDumpHostFile       = filepath.Join(scriptTempPath, jvmDumpHostScriptFile)
	jvmDumpK8SFile        = filepath.Join(scriptTempPath, jvmDumpK8SScriptFile)
	jvmMemorySnapshot     = "jvm_memory_snapshot"
	dwReportURL           = "/v1/write/remote_job"
)

var log = logger.DefaultSLogger("io_remote_job")

// RemoteJob  UnMarshal from dw.
type RemoteJob struct {
	JvmDumpJob *JVM `json:"jvm_dump_job"`
	// other jobs ...

	PullInterval int64 `json:"pull_interval"`
}

type jobReport struct {
	UUID    string `json:"uuid"`    // UUID
	IsOK    bool   `json:"isOK"`    // 是否成功
	Reason  string `json:"reason"`  // 任务执行失败原因
	Details string `json:"details"` // 命令行输出
}

type Manager struct {
	// 唯一作用就是将执行结果返回到 dw.
	DWURL    *dataway.Dataway
	Envs     []string
	Internal time.Duration
	PullFunc func(args string) ([]byte, error)
	JavaHome string
}

func (m *Manager) writeToFile() {
	_ = os.MkdirAll(scriptTempPath, 0o750) //nolint
	err := os.WriteFile(jvmDumpHostFile, []byte(jvmDumpHostScript), 0o600)
	if err != nil {
		log.Errorf("write host script to file err=%v", err)
		return
	}
	err = os.WriteFile(jvmDumpK8SFile, []byte(jvmDumpK8sScript), 0o600)
	if err != nil {
		log.Errorf("write k8s script to file err=%v", err)
		return
	}
}

func (m *Manager) Start() {
	log = logger.SLogger("remote_job")
	log.Infof("remote_job start, internal =%s", m.Internal.String())
	ticker := time.NewTicker(m.Internal)
	m.writeToFile()
	for {
		select {
		case <-ticker.C:
			log.Debugf("-----------job-----start")
			args := fmt.Sprintf("%s=true&host=%s", jvmMemorySnapshot, datakit.DKHost)
			body, err := m.PullFunc(args)
			if err != nil {
				log.Warnf("request remote err=%v", err)
				continue
			}
			if len(body) == 0 {
				continue
			}

			dumpJob := &RemoteJob{}
			err = json.Unmarshal(body, dumpJob)
			if err != nil {
				log.Errorf("json unmarshal err=%v", err)
				continue
			}
			if dumpJob.JvmDumpJob != nil {
				dumpJob.JvmDumpJob.javaHome = m.JavaHome
				report := dumpJob.JvmDumpJob.doCmd(m.Envs)
				m.returnToDW(report)
			}
		case <-datakit.Exit.Wait():
			return
		}
	}
}

func (m *Manager) returnToDW(jr *jobReport) {
	log.Infof("return job report is :%+v", jr)
	bts, _ := json.Marshal(jr)
	resp, err := m.DWURL.RemoteJob(bts)
	if err != nil {
		log.Errorf("err=%v", err)
		return
	}
	log.Infof("status code=%d", resp.StatusCode)
	if resp.StatusCode != 200 {
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Errorf("read err=%v", err)
			return
		}
		log.Debugf("read from %s body is:%s", dwReportURL, string(respBody))
	}
}
