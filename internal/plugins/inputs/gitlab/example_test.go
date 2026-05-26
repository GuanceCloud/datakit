// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package gitlab

import (
	"testing"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

func ExampleInput_SampleConfig() {
	// ipt := &Input{}
	// fmt.Println(ipt.SampleConfig())
	// Output:
	// [[inputs.gitlab]]
	//     ## set true if you need to collect metric from url below
	//     enable_collect = true
	//
	//     ## param type: string - default: http://127.0.0.1:80/-/metrics
	//     prometheus_url = "http://127.0.0.1:80/-/metrics"
	//
	//     ## param type: string - optional: time units are "ms", "s", "m", "h" - default: 10s
	//     interval = "10s"
	//
	//     ## datakit can listen to gitlab ci data at /v1/gitlab when enabled
	//     enable_ci_visibility = true
	//
	//     ## Set true to enable election
	//     election = true
	//
	//     ## Bearer token file for authentication
	//     bearer_token_file = ""
	//
	//     ## HTTP headers
	//     [inputs.gitlab.http_headers]
	//     # X-Custom-Header = "custom-value"
	//
	//     ## TLS configuration
	//     # tls_ca = "/path/to/ca.pem"
	//     # tls_cert = "/path/to/cert.pem"
	//     # tls_key = "/path/to/key.pem"
	//     # insecure_skip_verify = false
	//
	//     ## extra tags for gitlab-ci data.
	//     ## these tags will not overwrite existing tags.
	//     [inputs.gitlab.ci_extra_tags]
	//     # some_tag = "some_value"
	//     # more_tag = "some_other_value"
	//
	//     ## extra tags for gitlab metrics
	//     [inputs.gitlab.tags]
	//     # some_tag = "some_value"
	//     # more_tag = "some_other_value"
}

func TestInputInterface(t *testing.T) {
	// Test that Input implements required interfaces
	var _ interface {
		ElectionEnabled() bool
		RegHTTPHandler()
		Run()
		Terminate()
		Pause() error
		Resume() error
		SampleConfig() string
		Catalog() string
		SampleMeasurement() []inputs.Measurement
		AvailableArchs() []string
	} = &Input{}

	// The test passes if the above compiles without errors
}

func TestCatalog(t *testing.T) {
	ipt := &Input{}
	if ipt.Catalog() != "gitlab" {
		t.Errorf("Expected catalog 'gitlab', got '%s'", ipt.Catalog())
	}
}
