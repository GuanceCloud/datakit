//go:build linux
// +build linux

package run

import "testing"

func TestNewRunCmdRegistersTraceAllProcessFlag(t *testing.T) {
	cmd := NewRunCmd()

	if cmd.Flags().Lookup("trace-allprocess") == nil {
		t.Fatal("expected trace-allprocess flag to be registered")
	}
	if cmd.Flags().Lookup("netlog-l7log-headers") == nil {
		t.Fatal("expected netlog-l7log-headers flag to be registered")
	}
	if cmd.Flags().Lookup("network-path-enabled") == nil {
		t.Fatal("expected network-path-enabled flag to be registered")
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

func TestReadEnvNetworkPathConfig(t *testing.T) {
	t.Setenv("DKE_NETWORK_PATH_ENABLED", "true")
	t.Setenv("DKE_NETWORK_PATH_API", "http://127.0.0.1:9529/v1/netpath/candidates")
	t.Setenv("DKE_NETWORK_PATH_TOKEN", "secret")
	t.Setenv("DKE_NETWORK_PATH_FLUSH_INTERVAL", "5s")
	t.Setenv("DKE_NETWORK_PATH_BATCH_SIZE", "10")
	t.Setenv("DKE_NETWORK_PATH_HTTP_TIMEOUT", "2s")
	t.Setenv("DKE_NETWORK_PATH_QUEUE_SIZE", "100")

	var flag Flag
	readEnv(&flag)

	if !flag.NetworkPath.Enabled {
		t.Fatal("expected network path to be enabled")
	}
	if flag.NetworkPath.API != "http://127.0.0.1:9529/v1/netpath/candidates" {
		t.Fatalf("unexpected api %q", flag.NetworkPath.API)
	}
	if flag.NetworkPath.Token != "secret" ||
		flag.NetworkPath.FlushInterval != "5s" ||
		flag.NetworkPath.BatchSize != 10 ||
		flag.NetworkPath.HTTPTimeout != "2s" ||
		flag.NetworkPath.QueueSize != 100 {
		t.Fatalf("unexpected network path config %#v", flag.NetworkPath)
	}
}
