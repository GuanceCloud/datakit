// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package build

import (
	"testing"
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
