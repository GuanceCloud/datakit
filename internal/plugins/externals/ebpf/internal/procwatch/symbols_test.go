//go:build linux
// +build linux

package procwatch

import (
	"debug/elf"
	"os"
	"testing"
)

func TestParseGoVersionMatchesProducerString(t *testing.T) {
	version, ok := parseGoVersion("Go cmd/compile go1.23.4; regabi")
	if !ok {
		t.Fatal("expected producer string version to parse")
	}
	if version != [2]int{1, 23} {
		t.Fatalf("unexpected version: %#v", version)
	}
}

func TestResolveGoRuntimeRegisterABI(t *testing.T) {
	useRegister, err := resolveGoRuntimeRegisterABI([2]int{1, 17}, true, "amd64")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !useRegister {
		t.Fatal("expected amd64 Go 1.17 to use register ABI")
	}

	useRegister, err = resolveGoRuntimeRegisterABI([2]int{1, 16}, true, "amd64")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if useRegister {
		t.Fatal("expected amd64 Go 1.16 to use stack ABI")
	}

	if _, err := resolveGoRuntimeRegisterABI([2]int{}, false, "amd64"); err == nil {
		t.Fatal("expected unknown version to fail closed")
	}
}

func TestResolveGoRuntimeGoidOffsetFromCurrentExecutable(t *testing.T) {
	path, err := os.Executable()
	if err != nil {
		t.Fatalf("executable: %v", err)
	}

	f, err := elf.Open(path)
	if err != nil {
		t.Fatalf("open elf: %v", err)
	}
	defer f.Close() //nolint:errcheck

	offset, source, err := resolveGoRuntimeGoidOffset(f)
	if err != nil {
		if source == "dwarf" {
			t.Fatalf("resolve goid offset from dwarf: %v", err)
		}
		t.Skipf("skip current executable without compatible DWARF: %v", err)
	}
	if offset == 0 {
		t.Fatal("expected non-zero goid offset")
	}
}

func TestFindGoRuntimeExecuteFromCurrentExecutable(t *testing.T) {
	path, err := os.Executable()
	if err != nil {
		t.Fatalf("executable: %v", err)
	}

	f, err := elf.Open(path)
	if err != nil {
		t.Fatalf("open elf: %v", err)
	}
	defer f.Close() //nolint:errcheck

	symbol, err := findGoRuntimeExecute(f, "runtime.execute")
	if err != nil {
		t.Fatalf("find runtime.execute: %v", err)
	}
	if symbol == nil || symbol.Start == 0 {
		t.Fatal("expected runtime.execute symbol with non-zero offset")
	}
}

func TestFindGoRuntimeExecuteWithSourceFromCurrentExecutable(t *testing.T) {
	path, err := os.Executable()
	if err != nil {
		t.Fatalf("executable: %v", err)
	}

	f, err := elf.Open(path)
	if err != nil {
		t.Fatalf("open elf: %v", err)
	}
	defer f.Close() //nolint:errcheck

	symbol, source, err := findGoRuntimeExecuteWithSource(f, "runtime.execute")
	if err != nil {
		t.Fatalf("find runtime.execute with source: %v", err)
	}
	if symbol == nil || symbol.Start == 0 {
		t.Fatal("expected runtime.execute symbol with non-zero offset")
	}
	if source == "" {
		t.Fatal("expected non-empty symbol source")
	}
}
