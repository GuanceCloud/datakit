//go:build linux
// +build linux

package procwatch

import (
	"debug/buildinfo"
	"debug/dwarf"
	"debug/elf"
	"debug/gosym"
	"encoding/binary"
	"errors"
	"fmt"
	"regexp"
	"strconv"

	gosym2 "github.com/grafana/pyroscope/ebpf/symtab/gosym"
	bpfutil "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/bpfutil"
)

const defaultRuntimeGGoidOffset = uint64(152)

var goVersionPattern = regexp.MustCompile(`\bgo(\d+)\.(\d+)\b`)

type symbolLocation struct {
	Name  string
	Start uint64
	End   uint64
}

func findSymbol(elfFile *elf.File, funcName string) ([]elf.Symbol, error) {
	symbols, err := elfFile.Symbols()
	if err != nil {
		return nil, fmt.Errorf("failed to obtain symbol table: %w", err)
	}

	var matched []elf.Symbol
	for _, sym := range symbols {
		if elf.ST_TYPE(sym.Info) == elf.STT_FUNC && sym.Name == funcName {
			matched = append(matched, sym)
		}
	}
	bpfutil.SanitizeUprobeAddresses(elfFile, matched)
	return matched, nil
}

func parseGoVersion(version string) ([2]int, bool) {
	match := goVersionPattern.FindStringSubmatch(version)
	if len(match) != 3 {
		return [2]int{}, false
	}

	major, err := strconv.Atoi(match[1])
	if err != nil {
		return [2]int{}, false
	}
	minor, err := strconv.Atoi(match[2])
	if err != nil {
		return [2]int{}, false
	}

	return [2]int{major, minor}, true
}

func findGoRuntimeExecute(elfFile *elf.File, symbolName string) (*symbolLocation, error) {
	location, _, err := findGoRuntimeExecuteWithSource(elfFile, symbolName)
	return location, err
}

func findGoRuntimeExecuteWithSource(elfFile *elf.File, symbolName string) (*symbolLocation, string, error) {
	if elfFile == nil {
		return nil, "", nil
	}

	textSection := elfFile.Section(".text")
	if textSection == nil {
		return findGoRuntimeExecuteFromSymtab(elfFile, symbolName)
	}

	pclnSection := elfFile.Section(".gopclntab")
	if pclnSection == nil {
		return findGoRuntimeExecuteFromSymtab(elfFile, symbolName)
	}

	pclntab, err := pclnSection.Data()
	if err != nil {
		return findGoRuntimeExecuteFromSymtab(elfFile, symbolName)
	}

	textStart := gosym2.ParseRuntimeTextFromPclntab18(pclntab)
	if textStart == 0 {
		textStart = textSection.Addr
	}
	if textStart < textSection.Addr || textStart >= textSection.Addr+textSection.Size {
		return findGoRuntimeExecuteFromSymtab(elfFile, symbolName)
	}

	symbol, err := findGoRuntimeExecuteFromPCLN(elfFile, pclntab, textStart, symbolName)
	if err == nil {
		return symbol, "gopclntab", nil
	}

	if len(pclntab) < 4 || binary.LittleEndian.Uint32(pclntab[0:4]) != 0xFFFFFFF1 {
		return findGoRuntimeExecuteFromSymtab(elfFile, symbolName)
	}

	patched := append([]byte(nil), pclntab...)
	binary.LittleEndian.PutUint32(patched[0:4], 0xFFFFFFF0)

	patchedSymbol, patchedErr := findGoRuntimeExecuteFromPCLN(elfFile, patched, textStart, symbolName)
	if patchedErr == nil {
		return patchedSymbol, "gopclntab-patched", nil
	}
	symtabSymbol, source, symtabErr := findGoRuntimeExecuteFromSymtab(elfFile, symbolName)
	if symtabErr == nil {
		return symtabSymbol, source, nil
	}
	return nil, "", fmt.Errorf(
		"resolve %s from gopclntab failed before and after Go 1.20 magic patch: %w / %v / symtab: %v",
		symbolName, err, patchedErr, symtabErr,
	)
}

func findGoRuntimeExecuteFromSymtab(elfFile *elf.File, symbolName string) (*symbolLocation, string, error) {
	if elfFile == nil {
		return nil, "", errors.New("nil elf file")
	}

	symbols, err := findSymbol(elfFile, symbolName)
	if err != nil {
		return nil, "", err
	}
	if len(symbols) == 0 {
		return nil, "", fmt.Errorf("symbol %s not found in symtab", symbolName)
	}

	sym := symbols[0]
	location := normalizeSymbolLocation(elfFile, &symbolLocation{
		Name:  sym.Name,
		Start: sym.Value,
		End:   sym.Value + sym.Size,
	})
	return location, "symtab", nil
}

func findGoRuntimeExecuteFromPCLN(elfFile *elf.File, pclntab []byte, textStart uint64, symbolName string) (*symbolLocation, error) {
	lineTable := gosym.NewLineTable(pclntab, textStart)
	table, err := gosym.NewTable(nil, lineTable)
	if err != nil {
		return nil, err
	}
	if len(table.Funcs) == 0 {
		return nil, errors.New("gosymtab: no symbols found")
	}

	for _, fn := range table.Funcs {
		if fn.Name == symbolName {
			return normalizeSymbolLocation(elfFile, &symbolLocation{
				Name:  fn.Name,
				Start: fn.Entry,
				End:   fn.End,
			}), nil
		}
	}

	return nil, fmt.Errorf("symbol %s not found", symbolName)
}

func resolveGoVersion(binPath string, elfFile *elf.File) ([2]int, bool) {
	if binPath != "" {
		if buildInfo, err := buildinfo.ReadFile(binPath); err == nil {
			if version, ok := parseGoVersion(buildInfo.GoVersion); ok {
				return version, true
			}
		}
	}

	if elfFile == nil {
		return [2]int{}, false
	}

	dw, err := elfFile.DWARF()
	if err != nil {
		return [2]int{}, false
	}

	return resolveGoVersionFromDWARF(dw)
}

func resolveGoVersionFromDWARF(data *dwarf.Data) ([2]int, bool) {
	if data == nil {
		return [2]int{}, false
	}

	r := data.Reader()
	for {
		ent, err := r.Next()
		if err != nil || ent == nil {
			return [2]int{}, false
		}

		if ent.Tag == dwarf.TagCompileUnit {
			if producer, _ := ent.Val(dwarf.AttrProducer).(string); producer != "" {
				if version, ok := parseGoVersion(producer); ok {
					return version, true
				}
			}
		}

		if ent.Children {
			r.SkipChildren()
		}
	}
}

func resolveGoRuntimeGoidOffset(elfFile *elf.File) (uint64, string, error) {
	if elfFile == nil {
		return 0, "", errors.New("nil elf file")
	}

	dw, err := elfFile.DWARF()
	if err != nil {
		return defaultRuntimeGGoidOffset, "fallback", nil
	}

	offset, err := getStructFieldOffsetFromDWARF(dw, []string{"runtime.g", "g"}, "goid")
	if err != nil {
		// Some Go toolchains emit DWARF that our lightweight struct walker
		// cannot reliably resolve even though the runtime layout is otherwise
		// compatible. Fall back to the long-standing default offset instead of
		// disabling tracing for the whole process.
		return defaultRuntimeGGoidOffset, "fallback", nil
	}
	return offset, "dwarf", nil
}

func resolveGoRuntimeRegisterABI(goVersion [2]int, known bool, arch string) (bool, error) {
	if !known {
		return false, fmt.Errorf("unknown Go version")
	}

	if goVersion[0] != 1 {
		return false, fmt.Errorf("unsupported Go major version %d", goVersion[0])
	}

	switch arch {
	case "amd64":
		return goVersion[1] >= 17, nil
	case "arm64":
		return goVersion[1] >= 18, nil
	default:
		return false, fmt.Errorf("unsupported arch %q", arch)
	}
}

func getStructFieldOffsetFromDWARF(data *dwarf.Data, structNames []string, fieldName string) (uint64, error) {
	r := data.Reader()
	for {
		ent, err := r.Next()
		if err != nil {
			return 0, err
		}
		if ent == nil {
			break
		}

		if ent.Tag == dwarf.TagStructType {
			if name, _ := ent.Val(dwarf.AttrName).(string); name != "" {
				for _, candidate := range structNames {
					if name == candidate {
						return getStructFieldOffsetFromDWARFEntry(r, ent, fieldName)
					}
				}
			}
		}

		if ent.Children {
			r.SkipChildren()
		}
	}

	return 0, fmt.Errorf("struct %q not found", structNames)
}

func getStructFieldOffsetFromDWARFEntry(r *dwarf.Reader, ent *dwarf.Entry, fieldName string) (uint64, error) {
	if !ent.Children {
		return 0, fmt.Errorf("struct has no children")
	}

	for {
		child, err := r.Next()
		if err != nil {
			return 0, err
		}
		if child == nil || child.Tag == 0 {
			break
		}

		if child.Tag == dwarf.TagMember {
			if name, _ := child.Val(dwarf.AttrName).(string); name == fieldName {
				return dwarfFieldOffset(child)
			}
		}

		if child.Children {
			r.SkipChildren()
		}
	}

	return 0, fmt.Errorf("field %q not found", fieldName)
}

func dwarfFieldOffset(ent *dwarf.Entry) (uint64, error) {
	loc := ent.Val(dwarf.AttrDataMemberLoc)
	switch v := loc.(type) {
	case int64:
		if v < 0 {
			return 0, fmt.Errorf("negative offset %d", v)
		}
		return uint64(v), nil
	case uint64:
		return v, nil
	default:
		return 0, fmt.Errorf("unsupported DWARF member location type %T", loc)
	}
}

func normalizeSymbolLocation(elfFile *elf.File, sym *symbolLocation) *symbolLocation {
	if elfFile.Type == elf.ET_EXEC || elfFile.Type == elf.ET_DYN {
		for _, prog := range elfFile.Progs {
			if prog.Type == elf.PT_LOAD && sym.Start >= prog.Vaddr && sym.Start < prog.Vaddr+prog.Memsz {
				sym.Start = sym.Start - prog.Vaddr + prog.Off
				sym.End = sym.End - prog.Vaddr + prog.Off
			}
		}
	}
	return sym
}
