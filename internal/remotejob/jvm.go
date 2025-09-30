// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2024-present Guance, Inc.

package remotejob

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"time"
)

type JVM struct {
	Args       []string `json:"args"`
	Command    string   `json:"command"`
	CreatDate  int64    `json:"creat_date"`
	Host       string   `json:"host"`
	IsCustomer bool     `json:"is_customer"`
	Name       string   `json:"name"`
	PodName    string   `json:"pod_name"`
	ProcessID  int      `json:"process_id"`
	Timeout    int      `json:"timeout"`
	UUID       string   `json:"uuid"`
	Service    string   `json:"service"`

	javaHome string
}

func (j *JVM) check() error {
	// host 环境 需要的三个参数：-pid -osspath -javahome 。虚拟环境需要参数 -pid -osspath -pod_name。
	// 为了参数都有意义并传递下去，这里检查一下参数
	errStr := ""
	if j.UUID == "" {
		errStr += " uuid is nil "
	}
	if j.ProcessID == 0 {
		errStr += " process id is 0 "
	}
	if j.PodName == "" && j.javaHome == "" { // 主机环境下 java_home 不可为空。
		errStr += " can find $JAVA_HOME "
	}
	if errStr != "" {
		return fmt.Errorf(errStr)
	} else {
		return nil
	}
}

func (j *JVM) doCmd(envs []string) *jobReport {
	start := time.Now()
	jr := &jobReport{UUID: j.UUID, IsOK: true}
	err := j.check()
	if err != nil {
		jr.IsOK = false
		jr.Reason = fmt.Sprintf("jvm check err=%s", err.Error())
		return jr
	}
	path := "jvmdump"
	if j.Service == "" {
		j.Service = "default"
	}
	path = path + "/" + j.Service

	var cmd *exec.Cmd
	var name string
	status := "success"
	if j.PodName == "" {
		cmd = exec.Command("python3", //nolint
			jvmDumpHostFile,
			"-pid", strconv.Itoa(j.ProcessID),
			"-javahome", j.javaHome,
			"-osspath", path,
			"-service", j.Service)
		cmd.Env = envs
		log.Infof("cmd to string:%s", cmd.String())
		name = jvmDumpHostFile
	} else {
		host, port := os.Getenv("KUBERNETES_SERVICE_HOST"), os.Getenv("KUBERNETES_SERVICE_PORT")
		if len(host) == 0 || len(port) == 0 {
			jr.IsOK = false
			jr.Reason = fmt.Sprintf("k8s host or port not set. KUBERNETES_SERVICE_HOST = %s KUBERNETES_SERVICE_PORT=%s", host, port)
			return jr
		}

		cmd = exec.Command("python3", //nolint
			jvmDumpK8SFile,
			"-pod_name", j.PodName,
			"-pid", strconv.Itoa(j.ProcessID),
			"-osspath", path,
			"-service", j.Service)
		cmd.Env = envs
		cmd.Env = append(cmd.Env, "KUBERNETES_SERVICE_HOST="+host, "KUBERNETES_SERVICE_PORT="+port)
		log.Infof("cmd to string:%s", cmd.String())
		name = jvmDumpK8SFile
	}

	// 捕获标准输出和标准错误
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		log.Warnf("do %s err=%v bts=%s", j.Args, err, stderr.String())
		jr.IsOK = false
		jr.Reason = stderr.String()
		status = "failed"
	}
	jobRunVec.WithLabelValues(name, status).Observe(float64(time.Since(start)) / float64(time.Second))

	log.Infof("cmd command out string =%s", out.String())
	jr.Details = out.String()
	// 返回
	return jr
}
