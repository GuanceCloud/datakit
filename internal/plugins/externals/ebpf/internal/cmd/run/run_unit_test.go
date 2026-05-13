//go:build linux
// +build linux

package run

import "testing"

func TestNewRunCmdRegistersTraceAllProcessFlag(t *testing.T) {
	cmd := NewRunCmd()

	if cmd.Flags().Lookup("trace-allprocess") == nil {
		t.Fatal("expected trace-allprocess flag to be registered")
	}
}

func TestParseFlagsResetsFeatureGlobals(t *testing.T) {
	_, _, err := parseFlags(&Flag{
		Enabled: []string{
			inputNameNet,
			inputNameNetlog,
			pluginNameConntrack,
			pluginNameTracing,
		},
		EBPFNet: FlagNet{
			L7NetEnabled: []string{"httpflow", "httpflow-tls"},
			IPv6Disabled: true,
		},
		EBPFTrace: FlagTrace{
			ConvTraceToDD: true,
		},
	})
	if err != nil {
		t.Fatalf("parse flags first pass: %v", err)
	}

	if !enableEbpfNet || !enableBpfNetlog || !enableEbpfConntrack || !enableTrace {
		t.Fatal("expected feature flags to be enabled after first parse")
	}
	if !enableHTTPFlow || !enableHTTPFlowTLS || !ipv6Disabled || !conv2ddID {
		t.Fatal("expected protocol and trace flags to be enabled after first parse")
	}

	_, _, err = parseFlags(&Flag{
		Enabled: []string{inputNameBash},
	})
	if err != nil {
		t.Fatalf("parse flags second pass: %v", err)
	}

	if !enableEbpfBash {
		t.Fatal("expected bash feature to remain enabled on second parse")
	}
	if enableEbpfNet || enableBpfNetlog || enableEbpfConntrack || enableTrace {
		t.Fatal("expected unrelated feature flags to be reset on second parse")
	}
	if enableHTTPFlow || enableHTTPFlowTLS || ipv6Disabled || conv2ddID {
		t.Fatal("expected protocol and trace flags to be reset on second parse")
	}
}
