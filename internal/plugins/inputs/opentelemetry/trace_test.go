// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package opentelemetry

import (
	"encoding/hex"
	"testing"
)

func Test_convert(t *testing.T) {
	convertToDD = true
	id, _ := hex.DecodeString("818616084f850520843d19e3936e4720")
	t.Logf("id len=%d", len(id))
	type args struct {
		id []byte
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "case128", args: args{id: id}, want: "9528800851807586080"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := convert(tt.args.id); got != tt.want {
				t.Errorf("convert() = %v, want %v", got, tt.want)
			}
		})
	}
	convertToDD = false
}
