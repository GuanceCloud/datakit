// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package election

import "testing"

func TestElectionCfgNormalizeOperatorURL(t *testing.T) {
	cfg := &ElectionCfg{OperatorURL: " datakit-operator.datakit.svc:443/ "}
	if err := cfg.Normalize(); err != nil {
		t.Fatalf("Normalize() error = %v", err)
	}
	if cfg.OperatorURL != "https://datakit-operator.datakit.svc:443" {
		t.Fatalf("OperatorURL = %q", cfg.OperatorURL)
	}

	cfg.OperatorURL = "https://user:secret@datakit-operator.datakit.svc"
	if err := cfg.Normalize(); err == nil {
		t.Fatal("expected credential-bearing URL to be rejected")
	}
}
