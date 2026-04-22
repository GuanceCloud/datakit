//go:build linux
// +build linux

package netflow

import "testing"

func TestFailedNetflowLoadProgram(t *testing.T) {
	program, ok := failedNetflowLoadProgram(
		assertErr(`load collection: program kretprobe__do_sendfile: load program: permission denied`),
	)
	if !ok || program != "kretprobe__do_sendfile" {
		t.Fatalf("unexpected result: ok=%v program=%q", ok, program)
	}
}

func TestFailedNetflowAttachProgram(t *testing.T) {
	program, ok := failedNetflowAttachProgram(
		assertErr(`attach probe "kprobe__sockfd_lookup_light" (kprobe/sockfd_lookup_light): creating tracefs event`),
	)
	if !ok || program != "kprobe__sockfd_lookup_light" {
		t.Fatalf("unexpected result: ok=%v program=%q", ok, program)
	}
}

type stringError string

func (e stringError) Error() string { return string(e) }

func assertErr(msg string) error { return stringError(msg) }
