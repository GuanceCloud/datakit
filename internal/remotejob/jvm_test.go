// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2024-present Guance, Inc.

// Package remotejob is running GuanCe remote job.
package remotejob

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"testing"
)

func GetRemotePullMock(args string) (bts []byte, err error) {
	// 测试专用。
	req, err := http.NewRequest("get", "http://localhost:9876/v1/datakit/pull?token=xxxxxxxxxxxxxx", nil)
	if err != nil {
		log.Errorf("new request err=%v", err)
		return
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Errorf("response err=%v", err)
		return
	}
	bts, err = io.ReadAll(resp.Body)
	return bts, err
}

func GetRemotePullMockForK8S() (bts []byte) {
	// 测试专用。
	job := &RemoteJob{
		JvmDumpJob: &JVM{
			Args:       []string{"-pid", "1", "-osspath", "dd-image", "-pod_name", "tomcat-74dd8575c7-ht6b7"},
			Command:    "jvm.py",
			CreatDate:  0,
			Host:       "localhost",
			IsCustomer: false,
			Name:       "xxxx",
			PodName:    "tomcat-74dd8575c7-ht6b7",
			ProcessID:  0,
			Timeout:    10,
			UUID:       "00000000111111111111",
		},
		PullInterval: 10,
	}

	bts, _ = json.Marshal(job)

	return bts
}

func TestJVM_doCmd(t *testing.T) {
	host := os.Getenv("OSS_BUCKET_HOST")
	key := os.Getenv("OSS_ACCESS_KEY_ID")
	secret := os.Getenv("OSS_ACCESS_KEY_SECRET")
	name := os.Getenv("OSS_BUCKET_NAME")

	envs := []string{
		"OSS_BUCKET_HOST=" + host,
		"OSS_ACCESS_KEY_ID=", key,
		"OSS_ACCESS_KEY_SECRET=" + secret,
		"OSS_BUCKET_NAME=" + name,
	}
	j := &JVM{
		Args:       []string{"-pid", "333776", "-osspath", "dd-image", "-filename", "/tmp/jvmdump1"},
		Command:    "jvm.py",
		CreatDate:  0,
		Host:       "localhost",
		IsCustomer: false,
		Name:       "xxxx",
		PodName:    "",
		ProcessID:  0,
		Timeout:    10,
		UUID:       "00000000111111111111",
	}
	if jr := j.doCmd(envs); jr != nil {
		t.Logf("jvm string=%+v", jr)
	}
}

func TestJVM_check(t *testing.T) {
	type fields struct {
		Args       []string
		Command    string
		CreatDate  int64
		Host       string
		IsCustomer bool
		Name       string
		PodName    string
		ProcessID  int
		Timeout    int
		UUID       string
		javaHome   string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "case args",
			fields: fields{
				Args:       nil,
				Command:    "",
				CreatDate:  0,
				Host:       "",
				IsCustomer: false,
				Name:       "",
				PodName:    "",
				ProcessID:  123,
				Timeout:    0,
				UUID:       "UUID",
				javaHome:   "/usr/local/java_1.8",
			},
			wantErr: false,
		},
		{
			name: "case UUID",
			fields: fields{
				Args:       nil,
				Command:    "",
				CreatDate:  0,
				Host:       "",
				IsCustomer: false,
				Name:       "",
				PodName:    "",
				ProcessID:  0,
				Timeout:    0,
				UUID:       "",
				javaHome:   "",
			},
			wantErr: true,
		},
		{
			name: "case pid",
			fields: fields{
				Args:       nil,
				Command:    "",
				CreatDate:  0,
				Host:       "",
				IsCustomer: false,
				Name:       "",
				PodName:    "",
				ProcessID:  0,
				Timeout:    0,
				UUID:       "",
				javaHome:   "",
			},
			wantErr: true,
		},
		{
			name: "case pod_name",
			fields: fields{
				Args:       []string{"pid", "123"},
				Command:    "",
				CreatDate:  0,
				Host:       "",
				IsCustomer: false,
				Name:       "",
				PodName:    "pod_xx",
				ProcessID:  33333,
				Timeout:    0,
				UUID:       "uuid",
				javaHome:   "",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j := &JVM{
				Args:       tt.fields.Args,
				Command:    tt.fields.Command,
				CreatDate:  tt.fields.CreatDate,
				Host:       tt.fields.Host,
				IsCustomer: tt.fields.IsCustomer,
				Name:       tt.fields.Name,
				PodName:    tt.fields.PodName,
				ProcessID:  tt.fields.ProcessID,
				Timeout:    tt.fields.Timeout,
				UUID:       tt.fields.UUID,
				javaHome:   tt.fields.javaHome,
			}
			if err := j.check(); (err != nil) != tt.wantErr {
				t.Errorf("check() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
