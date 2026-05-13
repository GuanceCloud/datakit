//go:build linux
// +build linux

package netflow

import (
	"testing"
)

func TestDisabledNetflowProgramsDisableUDPWhenNotEnabled(t *testing.T) {
	SetEnableUDP(false)
	t.Cleanup(func() { SetEnableUDP(false) })

	disabled := disabledNetflowPrograms(0x0005000000000000, false, false)
	want := map[string]struct{}{
		"kprobe__ip_make_skb":      {},
		"kprobe__ip6_make_skb":     {},
		"kprobe__udp_recvmsg":      {},
		"kretprobe__udp_recvmsg":   {},
		"kprobe__inet_bind":        {},
		"kretprobe__inet_bind":     {},
		"kprobe__inet6_bind":       {},
		"kretprobe__inet6_bind":    {},
		"kprobe__udp_destroy_sock": {},
	}
	got := make(map[string]struct{}, len(disabled))
	for _, program := range disabled {
		got[program] = struct{}{}
	}
	for program := range want {
		if _, ok := got[program]; !ok {
			t.Fatalf("expected %s to be disabled, got %v", program, disabled)
		}
	}
}

func TestDisabledNetflowProgramsKeepsUDPWhenEnabled(t *testing.T) {
	SetEnableUDP(true)
	t.Cleanup(func() { SetEnableUDP(false) })

	disabled := disabledNetflowPrograms(0x0005000000000000, false, false)
	for _, program := range disabled {
		if program == "kprobe__udp_recvmsg" || program == "kprobe__ip_make_skb" {
			t.Fatalf("unexpected UDP program disabled when UDP is enabled: %v", disabled)
		}
	}
}

func TestTCPBindPortProbesAreOptional(t *testing.T) {
	for _, program := range []string{
		"kretprobe__inet_csk_accept",
		"kprobe__inet_csk_listen_stop",
	} {
		if isCriticalNetflowProgram(program) {
			t.Fatalf("expected TCP bindport probe %s to be optional", program)
		}
	}
}

func TestNetflowMapMaxEntriesEnv(t *testing.T) {
	t.Setenv(netflowMapMaxEntriesEnv, "")
	if got := netflowMapMaxEntries(); got != defaultNetflowMapMaxEntries {
		t.Fatalf("unexpected default map entries %d", got)
	}

	t.Setenv(netflowMapMaxEntriesEnv, "2048")
	if got := netflowMapMaxEntries(); got != 2048 {
		t.Fatalf("unexpected map entries %d", got)
	}

	t.Setenv(netflowMapMaxEntriesEnv, "1")
	if got := netflowMapMaxEntries(); got != minNetflowMapMaxEntries {
		t.Fatalf("unexpected clamped min map entries %d", got)
	}

	t.Setenv(netflowMapMaxEntriesEnv, "99999999")
	if got := netflowMapMaxEntries(); got != maxNetflowMapMaxEntries {
		t.Fatalf("unexpected clamped max map entries %d", got)
	}
}
