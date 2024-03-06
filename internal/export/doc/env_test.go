// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package doc

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

func TestSetENVDoc(t *testing.T) {
	type args struct {
		prefix string
		infos  []*inputs.ENVInfo
	}
	tests := []struct {
		name string
		args args
		want []*inputs.ENVInfo
	}{
		{
			name: "string",
			args: args{
				prefix: "ENV_INPUT_NEO4J_",
				infos:  mockENVInfos,
			},
			want: wantENVInfos,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SetENVDoc(tt.args.prefix, tt.args.infos)
			assert.NoErrorf(t, checkENVName(tt.want, got), "checkENVName error")
			assert.NoErrorf(t, checkToml(tt.want, got), "checkToml error")
		})
	}
}

func checkENVName(want, got []*inputs.ENVInfo) error {
	for i := range got {
		if want[i].ENVName != got[i].ENVName {
			return fmt.Errorf("checkENVName error, want: %s, got: %s", want[i].ENVName, got[i].ENVName)
		}
	}
	return nil
}

func checkToml(want, got []*inputs.ENVInfo) error {
	for i := range got {
		if want[i].ConfField != got[i].ConfField {
			return fmt.Errorf("checkToml error, want: %s, got: %s", want[i].ConfField, got[i].ConfField)
		}
	}
	return nil
}

var mockENVInfos = []*inputs.ENVInfo{
	{FieldName: "Interval", Type: TimeDuration, Desc: "Collect interval"},
	{FieldName: "URLs", Type: List, Example: `'["http://127.0.0.1:2004/metrics","http://127.0.0.2:2004/metrics"]'`, Desc: "Exporter URLs"},
	{FieldName: "TLSOpen", Type: Boolean, Desc: "TLS configuration"},
	{FieldName: "CacertFile", ENVName: "TLS_CA", ConfField: "tls_ca", Type: String, Example: `"/tmp/ca.crt"`},
	{FieldName: "CertFile", ENVName: "TLS_CERT", ConfField: "tls_cert", Type: String, Example: `"/tmp/peer.crt"`},
	{FieldName: "KeyFile", ENVName: "TLS_KEY", ConfField: "tls_key", Type: String, Example: `"/tmp/peer.key"`},
	{FieldName: "Election", Type: Boolean, Desc: "Set to 'true' to enable election"},
	{FieldName: "DisableHostTag", Type: Boolean, Desc: "Disable setting host tag for this input"},
	{FieldName: "DisableInstanceTag", Type: Boolean, Desc: "Disable setting instance tag for this input"},
	{FieldName: "Tags", Type: Map, Desc: "Customize tags"},
}

var wantENVInfos = []*inputs.ENVInfo{
	{FieldName: "Interval", ENVName: "ENV_INPUT_NEO4J_INTERVAL", ConfField: "interval", Type: "TimeDuration", Example: "", Default: "", Desc: "Collect interval"},
	{FieldName: "URLs", ENVName: "ENV_INPUT_NEO4J_URLS", ConfField: "urls", Type: "List", Example: "'[\"http://127.0.0.1:2004/metrics\",\"http://127.0.0.2:2004/metrics\"]'", Default: "", Desc: "Exporter URLs"},
	{FieldName: "TLSOpen", ENVName: "ENV_INPUT_NEO4J_TLS_OPEN", ConfField: "tls_open", Type: "Boolean", Example: "", Default: "", Desc: "TLS configuration"},
	{FieldName: "CacertFile", ENVName: "ENV_INPUT_NEO4J_TLS_CA", ConfField: "tls_ca", Type: "String", Example: "\"/tmp/ca.crt\"", Default: "", Desc: ""},
	{FieldName: "CertFile", ENVName: "ENV_INPUT_NEO4J_TLS_CERT", ConfField: "tls_cert", Type: "String", Example: "\"/tmp/peer.crt\"", Default: "", Desc: ""},
	{FieldName: "KeyFile", ENVName: "ENV_INPUT_NEO4J_TLS_KEY", ConfField: "tls_key", Type: "String", Example: "\"/tmp/peer.key\"", Default: "", Desc: ""},
	{FieldName: "Election", ENVName: "ENV_INPUT_NEO4J_ELECTION", ConfField: "election", Type: "Boolean", Example: "", Default: "", Desc: "Set to 'true' to enable election"},
	{FieldName: "DisableHostTag", ENVName: "ENV_INPUT_NEO4J_DISABLE_HOST_TAG", ConfField: "disable_host_tag", Type: "Boolean", Example: "", Default: "", Desc: "Disable setting host tag for this input"},
	{FieldName: "DisableInstanceTag", ENVName: "ENV_INPUT_NEO4J_DISABLE_INSTANCE_TAG", ConfField: "disable_instance_tag", Type: "Boolean", Example: "", Default: "", Desc: "Disable setting instance tag for this input"},
	{FieldName: "Tags", ENVName: "ENV_INPUT_NEO4J_TAGS", ConfField: "tags", Type: "Map", Example: "", Default: "", Desc: "Customize tags"},
}
