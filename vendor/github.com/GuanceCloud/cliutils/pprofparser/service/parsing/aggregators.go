package parsing

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/GuanceCloud/cliutils/pprofparser/domain/languages"
	"github.com/GuanceCloud/cliutils/pprofparser/domain/pprof"
	"github.com/GuanceCloud/cliutils/pprofparser/domain/quantity"
	"github.com/GuanceCloud/cliutils/pprofparser/tools/filepathtoolkit"
	"github.com/GuanceCloud/cliutils/pprofparser/tools/logtoolkit"
	"github.com/GuanceCloud/cliutils/pprofparser/tools/parsetoolkit"
	"github.com/google/pprof/profile"
)

// [|lm:System.Private.CoreLib;|ns:System.Diagnostics.Tracing;|ct:EventSource;|fn:DebugCheckEvent]

type DDFieldTag string

const UnknownInfo = "<unknown>"

const (
	AssemblyTag  DDFieldTag = "|lm:"
	NamespaceTag DDFieldTag = "|ns:"
	ClassTag     DDFieldTag = "|ct:"
	MethodTag    DDFieldTag = "|fn:"
)

var ddDotnetFieldIdx = map[DDFieldTag]int{
	AssemblyTag:  0,
	NamespaceTag: 1,
	ClassTag:     2,
	MethodTag:    3,
}

type GetPropertyByLine func(lang languages.Lang, line profile.Line) string

var (
	getFuncName       = getPropertyCallable(getFuncNameByLine)
	getMethod         = getPropertyCallable(getMethodByLine)
	getClass          = getPropertyCallable(getClassByLine)
	getNamespace      = getPropertyCallable(getNamespaceByLine)
	getAssembly       = getPropertyCallable(getAssemblyByLine)
	getFuncDisplayStr = getPropertyCallable(GetPrintStrByLine)
	getDirectory      = getPropertyCallable(getDirectoryByLine)
	getFile           = getPropertyCallable(getFileByLine)
	getPackageName    = getPropertyCallable(getPackageNameByLine)
	getLine           = getPropertyCallable(func(lang languages.Lang, line profile.Line) string {
		return strconv.FormatInt(getLineByLine(lang, line), 10)
	})
)

func getFuncNameByLine(lang languages.Lang, line profile.Line) string {
	switch lang {
	case languages.NodeJS:
		segments := strings.Split(line.Function.Name, ":")
		if len(segments) > 1 {
			return segments[len(segments)-2]
		}
		return UnknownInfo
	case languages.DotNet:
		return getDDDotnetMethodName(line.Function.Name)
	case languages.PHP:
		return getPHPBaseFuncName(line.Function.Name)
	}
	return line.Function.Name
}

func getDDDotnetMethodName(funcName string) string {
	pieces := strings.Split(funcName, " ")

	var className, methodName string

	classIdx, fnIdx := ddDotnetFieldIdx[ClassTag], ddDotnetFieldIdx[MethodTag]
	if classIdx < len(pieces) && strings.HasPrefix(pieces[classIdx], string(ClassTag)) {
		className = strings.TrimPrefix(pieces[classIdx], string(ClassTag))
	}
	if fnIdx < len(pieces) && strings.HasPrefix(pieces[fnIdx], string(MethodTag)) {
		methodName = strings.TrimPrefix(pieces[fnIdx], string(MethodTag))
	}

	if className == "" || methodName == "" {
		for _, piece := range pieces {
			piece = strings.TrimSpace(piece)
			if className == "" && strings.HasPrefix(piece, string(ClassTag)) {
				className = strings.TrimPrefix(piece, string(ClassTag))
			}
			if methodName == "" && strings.HasPrefix(piece, string(MethodTag)) {
				methodName = strings.TrimPrefix(piece, string(MethodTag))
			}
		}
	}
	if methodName == "" {
		return "unknown"
	}

	if className != "" {
		return className + "." + methodName
	}

	return methodName
}

func getDDDotnetField(funcName string, tag DDFieldTag) string {
	pieces := strings.Split(funcName, " ")
	idx := ddDotnetFieldIdx[tag]

	tagStr := string(tag)
	if idx < len(pieces) && strings.HasPrefix(pieces[idx], tagStr) {
		return strings.TrimPrefix(pieces[idx], tagStr)
	}
	for _, piece := range pieces {
		piece = strings.TrimSpace(piece)
		if strings.HasPrefix(piece, tagStr) {
			return strings.TrimPrefix(piece, tagStr)
		}
	}
	return "<unknown>"
}

func getFuncIdentifier(lang languages.Lang, smp *profile.Sample, reverse bool) string {
	i := 0
	if reverse {
		i = len(smp.Location) - 1
	}
	if len(smp.Location) > 0 {
		loc := smp.Location[i]
		if len(loc.Line) > 0 {
			return strconv.FormatUint(loc.Line[len(loc.Line)-1].Function.ID, 10)
		}
	}
	return UnknownInfo
}

func getMethodByLine(lang languages.Lang, line profile.Line) string {
	return getDDDotnetMethodName(line.Function.Name)
}

func getPropertyCallable(getPropertyByLine GetPropertyByLine) GetPropertyFunc {
	return func(lang languages.Lang, sample *profile.Sample, reverse bool) string {
		i := 0
		if reverse {
			i = len(sample.Location) - 1
		}

		if len(sample.Location) > 0 {
			loc := sample.Location[i]
			if len(loc.Line) > 0 {
				return getPropertyByLine(lang, loc.Line[len(loc.Line)-1])
			}
		}
		return UnknownInfo
	}
}

func getClassByLine(lang languages.Lang, line profile.Line) string {
	funcName := line.Function.Name
	switch lang {
	case languages.DotNet:
		return getDDDotnetField(funcName, ClassTag)
	case languages.PHP:
		funcName = getPHPBaseFuncName(funcName)
		if pos := strings.LastIndex(funcName, "::"); pos >= 0 {
			return funcName[:pos]
		}
		if pos := strings.Index(funcName, "|"); pos >= 0 {
			return funcName[:pos]
		}
		filename := strings.ReplaceAll(line.Function.Filename, "\\", "/")

		if pos := strings.Index(filename, "/vendor/"); pos >= 0 {
			filename = filename[pos+len("/vendor/"):]
			if idx := strings.Index(filename, "/src/"); idx >= 0 {
				filename = filename[:idx]
			} else if idx := strings.LastIndexByte(filename, '/'); idx >= 0 {
				filename = filename[:idx]
			}
			return filename
		}

		return "standard"
	}

	return UnknownInfo
}

func getNamespaceByLine(lang languages.Lang, line profile.Line) string {
	switch lang {
	case languages.DotNet:
		if namespace := getDDDotnetField(line.Function.Name, NamespaceTag); namespace != "" {
			return namespace
		}
	}
	return UnknownInfo
}

func getAssemblyByLine(lang languages.Lang, line profile.Line) string {
	switch lang {
	case languages.DotNet:
		if assembly := getDDDotnetField(line.Function.Name, AssemblyTag); assembly != "" {
			return assembly
		}
	}
	return UnknownInfo
}

func getLineByLine(lang languages.Lang, line profile.Line) int64 {
	switch lang {
	case languages.NodeJS:
		segments := strings.Split(line.Function.Name, ":")
		if len(segments) > 0 {
			lineNo := segments[len(segments)-1]
			if lineNoRegExp.MatchString(lineNo) {
				lineNum, _ := strconv.ParseInt(lineNo, 10, 64)
				return lineNum
			}
		}
	}
	return line.Line
}

func getFileByLine(lang languages.Lang, line profile.Line) string {
	switch lang {
	case languages.NodeJS:
		funcName := line.Function.Name
		segments := strings.Split(funcName, ":")
		if len(segments) >= 3 {
			filename := strings.TrimSpace(strings.Join(segments[:len(segments)-2], ":"))
			if filename != "" {
				return filename
			}
		}
		return UnknownInfo
	case languages.PHP:
		filename := strings.TrimSpace(line.Function.Filename)
		if filename == "" {
			filename = "standard"
		}
		return filename
	}
	return line.Function.Filename
}

func getDirectoryByLine(lang languages.Lang, line profile.Line) string {
	return filepathtoolkit.DirName(getFileByLine(lang, line))
}

func getThreadID(lang languages.Lang, smp *profile.Sample, reverse bool) string {
	return getThreadIDBySample(smp)
}

func getThreadIDBySample(smp *profile.Sample) string {
	if tid := parsetoolkit.GetLabel(smp, LabelThreadID); tid != "" {
		return tid
	}
	return UnknownInfo
}

func getThreadName(lang languages.Lang, smp *profile.Sample, reverse bool) string {
	return getThreadNameBySample(smp)
}

func getThreadNameBySample(smp *profile.Sample) string {
	if tName := parsetoolkit.GetLabel(smp, LabelThreadName); tName != "" {
		return tName
	}
	return UnknownInfo
}

func getPackageNameByLine(lang languages.Lang, line profile.Line) string {
	switch lang {
	case languages.GoLang:
		packageName, _ := cutGoFuncName(line.Function.Name)
		return packageName
	}
	return UnknownInfo
}

// cutGoFuncName 切割pprof go func 为 package 和 func name
// return package name 和 func name
func cutGoFuncName(funcName string) (string, string) {
	pos := strings.LastIndexByte(funcName, '/')
	packageName := ""
	if pos > -1 {
		packageName, funcName = funcName[:pos+1], funcName[pos+1:]
	}
	cuts := strings.SplitN(funcName, ".", 2)
	if len(cuts) < 2 {
		logtoolkit.Errorf(`func name not contains ".": %s`, funcName)
		return packageName, cuts[0]
	}
	return packageName + cuts[0], cuts[1]
}

func GetPrintStrByLine(lang languages.Lang, line profile.Line) string {
	switch lang {
	case languages.GoLang:
		_, funcName := cutGoFuncName(line.Function.Name)
		return fmt.Sprintf("%s(%s)", funcName, filepathtoolkit.BaseName(line.Function.Filename))
	case languages.NodeJS:
		// node:internal/timers:listOnTimeout:569
		// ./node_modules/@pyroscope/nodejs/dist/cjs/index.js:(anonymous):313
		// :(idle):0
		segments := strings.Split(line.Function.Name, ":")
		funcName := "<unknown>"
		filename := ""
		if len(segments) == 1 {
			funcName = segments[0]
		} else if len(segments) > 1 {
			funcName = segments[len(segments)-2]
			filename = strings.TrimSpace(strings.Join(segments[:len(segments)-2], ":"))
		}
		baseName := filepathtoolkit.BaseName(filename)
		if baseName == "" || baseName == "." {
			return funcName
		}
		return fmt.Sprintf("%s(%s)", funcName, baseName)

	case languages.DotNet:
		return getDDDotnetMethodName(line.Function.Name)
	case languages.PHP:
		filename := line.Function.Filename
		if filename != "" {
			filename = filepathtoolkit.BaseName(filename)
		}
		funcName := getPHPBaseFuncName(line.Function.Name)
		if filename != "" {
			return fmt.Sprintf("%s(%s)", funcName, filename)
		}
		return funcName
	default:
		return fmt.Sprintf("%s(%s)", line.Function.Name, filepathtoolkit.BaseName(line.Function.Filename))
	}
}

func getPHPBaseFuncName(funcName string) string {
	if funcName == "" {
		return UnknownInfo
	}
	pos := strings.LastIndexByte(funcName, '\\')
	if pos >= 0 && pos < len(funcName)-1 {
		return funcName[pos+1:]
	}
	return funcName
}

func GetSpyPrintStr(funcName, fileName string) string {
	return fmt.Sprintf("%s(%s)", funcName, filepathtoolkit.BaseName(fileName))
}

func GetFuncAndLineDisplay(lang languages.Lang, smp *profile.Sample, reverse bool) string {
	i := 0
	if reverse {
		i = len(smp.Location) - 1
	}
	if len(smp.Location) > 0 {
		loc := smp.Location[i]
		if len(loc.Line) > 0 {
			line := loc.Line[len(loc.Line)-1]
			switch lang {
			case languages.PHP:
				funcName := getPHPBaseFuncName(line.Function.Name)
				filename := line.Function.Filename
				if filename != "" {
					filename = filepathtoolkit.BaseName(filename)
				}
				if filename != "" {
					return fmt.Sprintf("%s(%s:L#%d)", funcName, filename, line.Line)
				}
				return funcName
			case languages.GoLang:
				_, funcName := cutGoFuncName(line.Function.Name)
				return fmt.Sprintf("%s(%s:L#%d)",
					funcName, filepathtoolkit.BaseName(line.Function.Filename), line.Line)
			default:
				return fmt.Sprintf("%s(%s:L#%d)",
					line.Function.Name, filepathtoolkit.BaseName(line.Function.Filename), line.Line)
			}
		}
	}
	return "<unknown>"
}

var (
	Function = &Aggregator{
		Name:            "Function",
		Mapping:         []string{pprof.FieldFunctionName},
		ShowLanguages:   languages.PythonID | languages.GolangID,
		GetIdentifier:   getFuncIdentifier,
		GetDisplayStr:   getFuncDisplayStr,
		GetMappingFuncs: []GetPropertyFunc{getFuncName},
	}

	PHPFunction = &Aggregator{
		Name:            "Function",
		Mapping:         []string{pprof.FieldFunctionName},
		GetIdentifier:   getFuncIdentifier,
		GetDisplayStr:   getFuncDisplayStr,
		GetMappingFuncs: []GetPropertyFunc{getFuncName},
	}

	Method = &Aggregator{
		Name:            "Method",
		Mapping:         []string{pprof.FieldFunctionName},
		ShowLanguages:   languages.JavaID | languages.DotNetID,
		GetIdentifier:   getMethod,
		GetDisplayStr:   getMethod,
		GetMappingFuncs: []GetPropertyFunc{getMethod},
	}

	Class = &Aggregator{
		Name:            "Class",
		Mapping:         []string{pprof.FieldClass},
		ShowLanguages:   languages.DotNetID,
		GetIdentifier:   getClass,
		GetDisplayStr:   getClass,
		GetMappingFuncs: []GetPropertyFunc{getClass},
	}

	Namespace = &Aggregator{
		Name:            "Namespace",
		Mapping:         []string{pprof.FieldNamespace},
		ShowLanguages:   languages.DotNetID,
		GetIdentifier:   getNamespace,
		GetDisplayStr:   getNamespace,
		GetMappingFuncs: []GetPropertyFunc{getNamespace},
	}

	Assembly = &Aggregator{
		Name:            "Assembly",
		Mapping:         []string{pprof.FieldAssembly},
		ShowLanguages:   languages.DotNetID,
		GetIdentifier:   getAssembly,
		GetDisplayStr:   getAssembly,
		GetMappingFuncs: []GetPropertyFunc{getAssembly},
	}

	PyroNodeFunction = &Aggregator{
		Name:            "Function",
		Mapping:         []string{pprof.FieldFunctionName},
		ShowLanguages:   languages.NodeJSID,
		GetIdentifier:   getFuncIdentifier,
		GetDisplayStr:   getFuncDisplayStr,
		GetMappingFuncs: []GetPropertyFunc{getFuncName},
	}

	FunctionLine = &Aggregator{
		Name:          "Function + Line",
		Mapping:       []string{pprof.FieldFunctionName, pprof.FieldLine},
		ShowLanguages: languages.PythonID | languages.GolangID,
		GetIdentifier: func(lang languages.Lang, smp *profile.Sample, reverse bool) string {
			i := 0
			if reverse {
				i = len(smp.Location) - 1
			}
			if len(smp.Location) > 0 {
				loc := smp.Location[i]
				if len(loc.Line) > 0 {
					return fmt.Sprintf("%s###%d###%d",
						loc.Line[len(loc.Line)-1].Function.Filename, loc.Line[len(loc.Line)-1].Function.ID, loc.Line[len(loc.Line)-1].Line)
				}
			}
			return "<unknown>"
		},
		GetDisplayStr:   GetFuncAndLineDisplay,
		GetMappingFuncs: []GetPropertyFunc{getFuncName, getLine},
	}

	Directory = &Aggregator{
		Name:            "Directory",
		Mapping:         []string{pprof.FieldDirectory},
		ShowLanguages:   languages.PythonID | languages.GolangID,
		GetIdentifier:   getDirectory,
		GetDisplayStr:   getDirectory,
		GetMappingFuncs: []GetPropertyFunc{getDirectory},
	}

	File = &Aggregator{
		Name:            "File",
		Mapping:         []string{pprof.FieldFile},
		ShowLanguages:   languages.PythonID | languages.GolangID,
		GetIdentifier:   getFile,
		GetDisplayStr:   getFile,
		GetMappingFuncs: []GetPropertyFunc{getFile},
	}

	PyroNodeFile = &Aggregator{
		Name:            "File",
		Mapping:         []string{pprof.FieldFile},
		ShowLanguages:   languages.NodeJSID,
		GetIdentifier:   getFile,
		GetDisplayStr:   getFile,
		GetMappingFuncs: []GetPropertyFunc{getFile},
	}

	ThreadID = &Aggregator{
		Name:            "Thread ID",
		Mapping:         []string{pprof.FieldThreadID},
		ShowLanguages:   languages.PythonID | languages.DotNetID,
		GetIdentifier:   getThreadID,
		GetDisplayStr:   getThreadID,
		GetMappingFuncs: []GetPropertyFunc{getThreadID},
	}

	ThreadName = &Aggregator{
		Name:            "Thread Name",
		Mapping:         []string{pprof.FieldThreadName},
		ShowLanguages:   languages.PythonID | languages.DotNetID,
		GetIdentifier:   getThreadName,
		GetDisplayStr:   getThreadName,
		GetMappingFuncs: []GetPropertyFunc{getThreadName},
	}

	Package = &Aggregator{
		Name:            "Package",
		Mapping:         []string{pprof.FieldPackage},
		ShowLanguages:   languages.GolangID,
		GetIdentifier:   getPackageName,
		GetDisplayStr:   getPackageName,
		GetMappingFuncs: []GetPropertyFunc{getPackageName},
	}
)

type GetPropertyFunc func(lang languages.Lang, smp *profile.Sample, reverse bool) string

type Aggregator struct {
	Name    string
	Mapping []string

	ShowLanguages languages.LangID

	// GetIdentifier 获取维度的唯一标识
	GetIdentifier GetPropertyFunc

	// GetDisplayStr 获取维度的显示字符
	GetDisplayStr GetPropertyFunc

	// GetMappingFuncs, 获取与Mapping字段对应值Func
	GetMappingFuncs []GetPropertyFunc
}

type AggregatorSelectSlice []*AggregatorSelect

func (asm AggregatorSelectSlice) CalcPercentAndQuantity(total int64) {
	for _, aggregatorSelect := range asm {
		for _, opt := range aggregatorSelect.Options {
			opt.CalcPercentAndQuantity(total)
		}
	}
}

func (asm AggregatorSelectSlice) MarshalJSON() ([]byte, error) {

	JSONMap := make([]*AggregatorSelectForJSON, 0, len(asm))

	for _, aggregatorSelect := range asm {

		selectForJSON := &AggregatorSelectForJSON{
			Dimension: aggregatorSelect.Aggregator.Name,
			Mapping:   aggregatorSelect.Mapping,
		}

		for _, opt := range aggregatorSelect.Options {
			selectForJSON.Options = append(selectForJSON.Options, opt)
		}

		sort.Sort(selectForJSON.Options)

		JSONMap = append(JSONMap, selectForJSON)
	}
	return json.Marshal(JSONMap)
}

type OptionSlice []*AggregatorOption

func (os OptionSlice) Len() int {
	return len(os)
}

func (os OptionSlice) Less(i, j int) bool {
	return os[i].Value > os[j].Value
}

func (os OptionSlice) Swap(i, j int) {
	os[i], os[j] = os[j], os[i]
}

type OptionMap map[string]*AggregatorOption

type AggregatorSelect struct {
	Aggregator *Aggregator
	Mapping    []string
	Options    OptionMap
}

type AggregatorSelectForJSON struct {
	Dimension string      `json:"dimension"`
	Mapping   []string    `json:"mapping"`
	Options   OptionSlice `json:"data"`
}

type AggregatorOption struct {
	Title         string             `json:"title"`
	Quantity      *quantity.Quantity `json:"quantity"`
	Value         int64              `json:"value"`
	Unit          *quantity.Unit     `json:"unit"`
	Percent       string             `json:"percent"`
	MappingValues []string           `json:"mappingValues"`
}

func (ao *AggregatorOption) CalcPercentAndQuantity(total int64) {

	if total <= 0 {
		ao.Percent = "100"
	} else {
		ao.Percent = fmt.Sprintf("%.2f", float64(ao.Value)/float64(total)*100)
	}

	if ao.Unit != nil {
		ao.Quantity = ao.Unit.Quantity(ao.Value)

		// 转成默认单位
		ao.Quantity.SwitchToDefaultUnit()
		ao.Value = ao.Quantity.Value
		ao.Unit = ao.Quantity.Unit
	}
}

var PythonAggregatorList = []*Aggregator{
	Function,
	FunctionLine,
	Directory,
	File,
	ThreadID,
	ThreadName,
}

var GoAggregatorList = []*Aggregator{
	Function,
	FunctionLine,
	Directory,
	File,
	Package,
}

var SpyAggregatorList = []*Aggregator{
	Function,
	FunctionLine,
	Directory,
	File,
	ThreadName,
}

var PyroscopeNodeJSAggregatorList = []*Aggregator{
	PyroNodeFunction,
	PyroNodeFile,
}

var DDTraceDotnetAggregatorList = []*Aggregator{
	Method,
	Class,
	Namespace,
	Assembly,
	ThreadID,
	ThreadName,
}

var DDTracePHPAggregatorList = []*Aggregator{
	PHPFunction,
	FunctionLine,
	Class,
	File,
	Directory,
}
