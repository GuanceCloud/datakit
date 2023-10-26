//go:build linux
// +build linux

// Package sysmonitor monitor system status
package sysmonitor

import (
	"debug/dwarf"
	"debug/elf"
	"debug/gosym"
	"encoding/binary"
	"errors"
	"fmt"
	"regexp"
	"strconv"

	manager "github.com/DataDog/ebpf-manager"
	gosym2 "github.com/grafana/pyroscope/ebpf/symtab/gosym"
)

var goVersionRe = regexp.MustCompile(`^go(\d+)\.(\d+)`)

func FindSymbol(elfFile *elf.File, fnName string) ([]elf.Symbol, error) {
	if syms, err := elfFile.Symbols(); err != nil {
		return nil, fmt.Errorf("failed to obtain symbol table: %w", err)
	} else {
		var matched []elf.Symbol
		for _, s := range syms {
			if elf.ST_TYPE(s.Info) == elf.STT_FUNC && s.Name == fnName {
				matched = append(matched, s)
			}
		}
		manager.SanitizeUprobeAddresses(elfFile, matched)
		return matched, nil
	}
}

func FindDynamicSymbol(elfFile *elf.File, fnName string) ([]elf.Symbol, error) {
	if syms, err := elfFile.DynamicSymbols(); err != nil {
		return nil, fmt.Errorf("failed to obtain symbol table: %w", err)
	} else {
		var matched []elf.Symbol
		for _, s := range syms {
			if elf.ST_TYPE(s.Info) == elf.STT_FUNC&0xF && s.Name == fnName {
				matched = append(matched, s)
			}
		}
		manager.SanitizeUprobeAddresses(elfFile, matched)
		return matched, nil
	}
}

func FindStructMemberOffset(data *dwarf.Data, typeName, memberName string) (int64, error) {
	reader := data.Reader()

	var findType bool
	for !findType {
		entry, err := reader.Next()
		if err != nil {
			break
		}
		if entry == nil {
			break
		}
		if entry.Tag != dwarf.TagStructType {
			continue
		}

		for _, f := range entry.Field {
			if f.Attr == dwarf.AttrName {
				if name, ok := f.Val.(string); ok &&
					name == typeName {
					findType = true
					break
				}
			}
		}
	}

	if !findType {
		return 0, fmt.Errorf("struct not found")
	}

	var findMemb bool
	var offset int64
	for !findMemb {
		entry, err := reader.Next()
		if err != nil {
			break
		}
		if entry == nil {
			break
		}

		if entry.Tag == dwarf.TagMember {
			for _, f := range entry.Field {
				switch f.Attr { //nolint:exhaustive
				case dwarf.AttrDataMemberLoc:
					if v, ok := f.Val.(int64); ok {
						offset = v
					}
				case dwarf.AttrName:
					if v, ok := f.Val.(string); ok && v == memberName {
						findMemb = true
					}
				}
			}
		}
	}

	if findMemb {
		return offset, nil
	}

	return 0, fmt.Errorf("field not found")
}

func FindMemberOffsetFromFile(fp string, typeName, memberName string) (int64, error) {
	f, err := elf.Open(fp)
	if err != nil {
		return 0, err
	}

	dw, err := f.DWARF()
	if err != nil {
		return 0, err
	}

	return FindStructMemberOffset(dw, typeName, memberName)
}

func parseGoVersion(v string) ([2]int, bool) {
	r := [2]int{}
	ver := goVersionRe.FindAllStringSubmatch(v, 2)
	if len(ver) != 1 || len(ver[0]) != 3 {
		return [2]int{}, false
	}
	var err error
	r[0], err = strconv.Atoi(ver[0][1])
	if err != nil {
		return [2]int{}, false
	}
	r[1], err = strconv.Atoi(ver[0][2])
	if err != nil {
		return [2]int{}, false
	}

	return r, true
}

type SymLoc struct {
	Name  string
	Start uint64
	End   uint64
}

// Copyright 2020-2023 Grafana Labs
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
func getGoUprobeSymbolFromPCLN(elfFile *elf.File, patchGo20Magic bool, symName string) (*SymLoc, error) {
	var err error
	var pclntab []byte

	if elfFile == nil {
		return nil, nil
	}

	text := elfFile.Section(".text")
	if text == nil {
		return nil, errors.New("empty .text")
	}
	if sect := elfFile.Section(".gopclntab"); sect != nil {
		if pclntab, err = sect.Data(); err != nil {
			return nil, err
		}
	} else {
		return nil, errors.New("empty .gopclntab")
	}

	textStart := gosym2.ParseRuntimeTextFromPclntab18(pclntab)

	if textStart == 0 {
		// for older versions text.Addr is enough
		// https://github.com/golang/go/commit/b38ab0ac5f78ac03a38052018ff629c03e36b864
		textStart = text.Addr
	}
	if textStart < text.Addr || textStart >= text.Addr+text.Size {
		return nil, fmt.Errorf(" runtime.text out of .text bounds %d %d %d", textStart, text.Addr, text.Size)
	}

	if patchGo20Magic {
		magic := pclntab[0:4]
		if binary.LittleEndian.Uint32(magic) == 0xFFFFFFF1 {
			binary.LittleEndian.PutUint32(magic, 0xFFFFFFF0)
		}
	}
	pcln := gosym.NewLineTable(pclntab, textStart)
	table, err := gosym.NewTable(nil, pcln)
	if err != nil {
		return nil, err
	}
	if len(table.Funcs) == 0 {
		return nil, errors.New("gosymtab: no symbols found")
	}

	for _, fun := range table.Funcs {
		if fun.Name == symName {
			loc := sanitizeUprobeAddresses(elfFile, &SymLoc{
				Start: fun.Entry,
				End:   fun.End,
				Name:  fun.Name,
			})
			return loc, nil
		}
	}

	return nil, fmt.Errorf("symbol %s not found", symName)
}

// modified from github.com/DataDog/ebpf/manager/utils.go/SanitizeUprobeAddresses
//
// Copyright 2016-present Datadog, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
func sanitizeUprobeAddresses(f *elf.File, sym *SymLoc) *SymLoc {
	// If the binary is a non-PIE executable, addr must be a virtual address, otherwise it must be an offset relative to
	// the file load address. For executable (ET_EXEC) binaries and shared objects (ET_DYN), translate the virtual
	// address to physical address in the binary file.
	if f.Type == elf.ET_EXEC || f.Type == elf.ET_DYN {
		for _, prog := range f.Progs {
			if prog.Type == elf.PT_LOAD {
				if sym.Start >= prog.Vaddr && sym.Start < (prog.Vaddr+prog.Memsz) {
					sym.Start = sym.Start - prog.Vaddr + prog.Off
					sym.End = sym.End - prog.Vaddr + prog.Off
				}
			}
		}
	}
	return sym
}
