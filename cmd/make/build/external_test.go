// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package build

import (
	"testing"
	"time"
)

func Test_getProjectPrefix(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "normal",
			args: args{str: "/root/gopath/src/gitlab.jiagouyun.com/cloudcare-tools/datakit"},
			want: "/root/gopath/src/",
		},

		{
			name: "empty",
			args: args{str: "/root/gopath/src/cloudcare-tools/datakit"},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getProjectPrefix(tt.args.str); got != tt.want {
				t.Errorf("getProjectPrefix() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_ebpfCollectorLocalBuildCommand(t *testing.T) {
	t.Setenv("GO_MODULE_MODE", "")

	buildAt := time.Date(2026, 5, 7, 1, 2, 3, 0, time.UTC)
	args, envs := ebpfCollectorLocalBuildCommand("/tmp/dist/externals", "datakit-ebpf", "linux", "arm64", buildAt)

	wantArgs := []string{
		"go",
		"build",
		"-tags", "ebpf netgo",
		"-buildvcs=false",
		"-o", "/tmp/dist/externals/datakit-ebpf",
		"-ldflags", "-w -s -X 'main.Arch=linux/arm64' -X 'main.Date=2026-05-07T01:02:03Z'",
		"internal/plugins/externals/ebpf/cmd/datakit-ebpf/datakit-ebpf.go",
	}
	if len(args) != len(wantArgs) {
		t.Fatalf("args len = %d, want %d: %#v", len(args), len(wantArgs), args)
	}
	for i := range wantArgs {
		if args[i] != wantArgs[i] {
			t.Fatalf("args[%d] = %q, want %q", i, args[i], wantArgs[i])
		}
	}

	assertEnvValue(t, envs, "GOOS", "linux")
	assertEnvValue(t, envs, "GOARCH", "arm64")
	assertEnvValue(t, envs, "CGO_ENABLED", "0")
	assertEnvValue(t, envs, "GOFLAGS", "-mod=vendor")
}

func Test_ebpfTargetArg(t *testing.T) {
	if got := ebpfTargetArg("ebpf"); got != " EXTERNAL_EBPF_TARGET='bpfobjs'" {
		t.Fatalf("ebpfTargetArg(ebpf) = %q", got)
	}
	if got := ebpfTargetArg("oracle"); got != "" {
		t.Fatalf("ebpfTargetArg(oracle) = %q", got)
	}
}

func assertEnvValue(t *testing.T, envs []string, key, want string) {
	t.Helper()
	if got := envValue(envs, key); got != want {
		t.Fatalf("env %s = %q, want %q", key, got, want)
	}
}

func Test_getProjectSuffix(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "normal",
			args: args{str: "/root/gopath/src/gitlab.jiagouyun.com/cloudcare-tools/datakit/dist/datakit-linux-amd64/externals/oceanbase"},
			want: "dist/datakit-linux-amd64/externals/oceanbase",
		},

		{
			name: "empty",
			args: args{str: "/root/gopath/src/gitlab.jiagouyun.com/cloudcare-tools/datakit1/dist/datakit-linux-amd64/externals/oceanbase"},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getProjectSuffix(tt.args.str); got != tt.want {
				t.Errorf("getProjectSuffix() = %v, want %v", got, tt.want)
			}
		})
	}
}
