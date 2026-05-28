//go:build linux
// +build linux

package conntrack

import (
	"os"
	"path/filepath"
	"testing"
)

func TestKernelSymbolsAvailable(t *testing.T) {
	text := "" +
		"0000000000000000 t __nf_conntrack_hash_insert\t[nf_conntrack]\n" +
		"0000000000000000 t nf_conntrack_hash_check_insert\t[nf_conntrack]\n" +
		"0000000000000000 t nf_ct_delete\t[nf_conntrack]\n"

	if !kernelSymbolsAvailable(text, []string{"__nf_conntrack_hash_insert"}) {
		t.Fatal("expected __nf_conntrack_hash_insert to be detected")
	}
	if !kernelSymbolsAvailable(text, []string{"nf_ct_delete"}) {
		t.Fatal("expected nf_ct_delete to be detected")
	}
	if kernelSymbolsAvailable(text, []string{"nf_conntrack_confirm"}) {
		t.Fatal("did not expect unrelated symbol to be detected")
	}
}

func TestKernelSymbolDetectionIgnoresDerivedSymbols(t *testing.T) {
	text := "" +
		"0000000000000000 t nf_conntrack_hash_check_insert.cold\t[nf_conntrack]\n" +
		"0000000000000000 t __pfx_nf_conntrack_hash_check_insert\t[nf_conntrack]\n" +
		"0000000000000000 r __kstrtab_nf_conntrack_hash_check_insert\t[nf_conntrack]\n"

	if kernelSymbolsAvailable(text, []string{"nf_conntrack_hash_check_insert"}) {
		t.Fatal("did not expect derived symbols to satisfy exact hook detection")
	}
	if got := firstAvailableKernelSymbol(text, []string{"nf_conntrack_hash_check_insert"}); got != "" {
		t.Fatalf("unexpected derived symbol match: %q", got)
	}
}

func TestResolveHookSelectionPrefersHashCheckInsert(t *testing.T) {
	text := "" +
		"0000000000000000 t __nf_conntrack_hash_insert\t[nf_conntrack]\n" +
		"0000000000000000 t nf_conntrack_hash_check_insert\t[nf_conntrack]\n" +
		"0000000000000000 t nf_ct_delete\t[nf_conntrack]\n"

	got := resolveHookSelectionWithKernelVersion(text, 0)
	if got.InsertSymbol != "nf_conntrack_hash_check_insert" {
		t.Fatalf("unexpected insert symbol: %q", got.InsertSymbol)
	}
	if len(got.InsertSymbols) != 2 ||
		got.InsertSymbols[0] != "nf_conntrack_hash_check_insert" ||
		got.InsertSymbols[1] != "__nf_conntrack_hash_insert" {
		t.Fatalf("unexpected insert symbols: %#v", got.InsertSymbols)
	}
	if got.DeleteSymbol != "nf_ct_delete" {
		t.Fatalf("unexpected delete symbol: %q", got.DeleteSymbol)
	}
}

func TestResolveHookSelectionPrefersConfirmOn415(t *testing.T) {
	text := "" +
		"0000000000000000 t __nf_conntrack_confirm\t[nf_conntrack]\n" +
		"0000000000000000 t __nf_conntrack_hash_insert\t[nf_conntrack]\n" +
		"0000000000000000 t nf_conntrack_hash_check_insert\t[nf_conntrack]\n" +
		"0000000000000000 t nf_ct_delete\t[nf_conntrack]\n"

	got := resolveHookSelectionWithKernelVersion(text, 0x0004000f00000000)
	if got.InsertSymbol != "__nf_conntrack_confirm" {
		t.Fatalf("unexpected insert symbol: %q", got.InsertSymbol)
	}
	if len(got.InsertSymbols) != 3 ||
		got.InsertSymbols[0] != "__nf_conntrack_confirm" ||
		got.InsertSymbols[1] != "nf_conntrack_hash_check_insert" ||
		got.InsertSymbols[2] != "__nf_conntrack_hash_insert" {
		t.Fatalf("unexpected insert symbols: %#v", got.InsertSymbols)
	}
}

func TestResolveHookSelectionPrefersConfirmWhenAvailableOn310(t *testing.T) {
	text := "" +
		"0000000000000000 t __nf_conntrack_confirm\t[nf_conntrack]\n" +
		"0000000000000000 t __nf_conntrack_hash_insert\t[nf_conntrack]\n" +
		"0000000000000000 t nf_ct_delete\t[nf_conntrack]\n"

	got := resolveHookSelectionWithKernelVersion(text, 0x0003000a00000000)
	if got.InsertSymbol != "__nf_conntrack_confirm" {
		t.Fatalf("unexpected insert symbol: %q", got.InsertSymbol)
	}
	if len(got.InsertSymbols) != 2 ||
		got.InsertSymbols[0] != "__nf_conntrack_confirm" ||
		got.InsertSymbols[1] != "__nf_conntrack_hash_insert" {
		t.Fatalf("unexpected insert symbols: %#v", got.InsertSymbols)
	}
}

func TestConntrackInsertProgramNames(t *testing.T) {
	got := ConntrackInsertProgramNames([]string{
		"__nf_conntrack_confirm",
		"nf_conntrack_hash_check_insert",
		"__nf_conntrack_hash_insert",
		"nf_conntrack_hash_check_insert",
		"missing",
	})

	want := []string{
		"kprobe___nf_conntrack_confirm",
		"kprobe__nf_conntrack_hash_check_insert",
		"kprobe___nf_conntrack_hash_insert",
	}

	if len(got) != len(want) {
		t.Fatalf("unexpected program count: got=%d want=%d (%#v)", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("unexpected program at %d: got=%q want=%q", i, got[i], want[i])
		}
	}
}

func TestConntrackKprobeInterfaceAvailableAt(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "kprobe_events")
	if err := os.WriteFile(path, []byte(""), 0o644); err != nil {
		t.Fatalf("write probe marker: %v", err)
	}

	if !conntrackKprobeInterfaceAvailableAt([]string{filepath.Join(dir, "missing"), path}) {
		t.Fatal("expected synthetic kprobe interface to be detected")
	}
	if conntrackKprobeInterfaceAvailableAt([]string{filepath.Join(dir, "missing")}) {
		t.Fatal("did not expect missing path to be treated as available")
	}
}

func TestConntrackTupleMapMaxEntriesEnv(t *testing.T) {
	t.Setenv(conntrackTupleMapMaxEntriesEnv, "")
	if got := conntrackTupleMapMaxEntries(); got != defaultConntrackTupleMapMaxEntries {
		t.Fatalf("default tuple map entries = %d, want %d", got, defaultConntrackTupleMapMaxEntries)
	}

	t.Setenv(conntrackTupleMapMaxEntriesEnv, "2048")
	if got := conntrackTupleMapMaxEntries(); got != 2048 {
		t.Fatalf("configured tuple map entries = %d, want 2048", got)
	}

	t.Setenv(conntrackTupleMapMaxEntriesEnv, "1")
	if got := conntrackTupleMapMaxEntries(); got != minConntrackTupleMapMaxEntries {
		t.Fatalf("min-clamped tuple map entries = %d, want %d", got, minConntrackTupleMapMaxEntries)
	}

	t.Setenv(conntrackTupleMapMaxEntriesEnv, "99999999")
	if got := conntrackTupleMapMaxEntries(); got != maxConntrackTupleMapMaxEntries {
		t.Fatalf("max-clamped tuple map entries = %d, want %d", got, maxConntrackTupleMapMaxEntries)
	}
}
