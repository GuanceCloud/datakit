// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package languages define the supported languages.
package languages

import "strings"

const (
	Any     Lang = "any" // 代表全部语言
	Java    Lang = "java"
	Python  Lang = "python"
	GoLang  Lang = "golang"
	Ruby    Lang = "ruby"
	Ebpf    Lang = "ebpf"
	NodeJS  Lang = "nodejs"
	DotNet  Lang = "dotnet"
	PHP     Lang = "php"
	Unknown Lang = "unknown"
)

const (
	AnyID LangID = 1 << iota
	PythonID
	GolangID
	JavaID
	RubyID
	EbpfID
	NodeJSID
	DotNetID
	PhpID
)

type Lang string
type LangID int

func (l Lang) String() string {
	return strings.ToLower(string(l))
}

func (l Lang) Is(target Lang) bool {
	return l.String() == target.String()
}
