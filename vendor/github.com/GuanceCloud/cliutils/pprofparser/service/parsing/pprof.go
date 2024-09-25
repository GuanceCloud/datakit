package parsing

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/GuanceCloud/cliutils/pprofparser/domain/events"
	"github.com/GuanceCloud/cliutils/pprofparser/domain/languages"
	"github.com/GuanceCloud/cliutils/pprofparser/domain/parameter"
	"github.com/GuanceCloud/cliutils/pprofparser/domain/pprof"
	"github.com/GuanceCloud/cliutils/pprofparser/domain/quantity"
	"github.com/GuanceCloud/cliutils/pprofparser/domain/tracing"
	"github.com/GuanceCloud/cliutils/pprofparser/service/storage"
	"github.com/GuanceCloud/cliutils/pprofparser/tools/logtoolkit"
	"github.com/GuanceCloud/cliutils/pprofparser/tools/parsetoolkit"
	"github.com/google/pprof/profile"
	"github.com/pierrec/lz4/v4"
)

const (
	LabelExceptionType   = "exception type"
	LabelThreadID        = "thread id"
	LabelThreadNativeID  = "thread native id"
	LabelThreadName      = "thread name"
	LabelSpanID          = "span id"
	LabelLocalRootSpanID = "local root span id"
)

var (
	ZIPMagic  = []byte{0x50, 0x4b, 3, 4}
	LZ4Magic  = []byte{4, 34, 77, 24}
	GZIPMagic = []byte{31, 139}
)

var lineNoRegExp = regexp.MustCompile(`^\d+$`)

type PProf struct {
	from          string
	workspaceUUID string
	profiles      []*parameter.Profile
	filterBySpan  bool
	span          *parameter.Span
	spanIDSet     *tracing.SpanIDSet
	DisplayCtl
}

type Decompressor struct {
	io.Reader
	r io.Reader
}

func NewDecompressor(r io.Reader) io.ReadCloser {
	bufReader := bufio.NewReader(r)

	magics, err := bufReader.Peek(4)
	if err != nil {
		return &Decompressor{
			r:      r,
			Reader: bufReader,
		}
	}

	if bytes.Compare(LZ4Magic, magics) == 0 {
		return &Decompressor{
			r:      r,
			Reader: lz4.NewReader(bufReader),
		}
	}

	return &Decompressor{
		r:      r,
		Reader: bufReader,
	}
}

func (d *Decompressor) Close() error {
	var err error
	if rc, ok := d.Reader.(io.Closer); ok {
		if e := rc.Close(); e != nil {
			err = e
		}
	}

	if rc, ok := d.r.(io.Closer); ok {
		if e := rc.Close(); e != nil {
			err = e
		}
	}
	return err
}

func NewPProfParser(
	from string,
	workspaceUUID string,
	profiles []*parameter.Profile,
	filterBySpan bool,
	span *parameter.Span,
	spanIDSet *tracing.SpanIDSet,
	ctl DisplayCtl,
) *PProf {
	return &PProf{
		from:          from,
		workspaceUUID: workspaceUUID,
		profiles:      profiles,
		filterBySpan:  filterBySpan,
		span:          span,
		spanIDSet:     spanIDSet,
		DisplayCtl:    ctl,
	}
}

func isGlobPattern(pattern string) bool {
	return strings.ContainsAny(pattern, "?*")
}

func (p *PProf) mergePProf(filename string) (*profile.Profile, error) {
	if len(p.profiles) == 0 {
		return nil, fmt.Errorf("empty profiles")
	}

	client, err := storage.GetStorage(storage.LocalDisk)
	if err != nil {
		return nil, fmt.Errorf("init oss client err: %w", err)
	}

	filenames := []string{filename}

	if strings.ContainsRune(filename, '|') {
		filenames = strings.Split(filename, "|")
	}

	profSrc := make([]*profile.Profile, 0, len(p.profiles))

	for _, prof := range p.profiles {
		startTime, err := prof.StartTime()
		if err != nil {
			return nil, fmt.Errorf("cast ProfileStart to int64 fail: %w", err)
		}

		profilePath := ""

	FilenameLoop:
		for _, name := range filenames {
			if name == "" {
				continue
			}

			pattern := client.GetProfilePath(p.workspaceUUID, prof.ProfileID, startTime, name)
			if ok, err := client.IsFileExists(pattern); ok && err == nil {
				profilePath = pattern
				break
			}

			// check whether the filename is a glob pattern
			if isGlobPattern(name) {
				matches, err := filepath.Glob(pattern)
				if err != nil {
					return nil, fmt.Errorf("illegal glob pattern [%s]; %w", pattern, err)
				}

				for _, match := range matches {
					baseName := filepath.Base(match)
					if baseName != events.DefaultMetaFileName && baseName != events.DefaultMetaFileNameWithExt {
						profilePath = match
						break FilenameLoop
					}
				}
			}
		}

		if profilePath == "" {
			return nil, fmt.Errorf("no available profile file found: [%s]: %w", filename, fs.ErrNotExist)
		}

		reader, err := client.ReadFile(profilePath)

		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				return nil, fmt.Errorf("profile file [%s] not exists: %w", profilePath, err)
			}
			if ok, err := client.IsFileExists(profilePath); err == nil && !ok {
				return nil, fmt.Errorf("profile file [%s] not exists：%w", profilePath, fs.ErrNotExist)
			}
			return nil, fmt.Errorf("unable to read profile file [%s]: %w", profilePath, err)
		}

		parsedPProf, err := parseAndClose(NewDecompressor(reader))
		if err != nil {
			logtoolkit.Errorf("parse profile [path:%s] fail: %s", profilePath, err)
			continue
		}

		profSrc = append(profSrc, parsedPProf)
	}

	if len(profSrc) == 0 {
		return nil, fmt.Errorf("no available profile")
	}

	mergedPProf, err := profile.Merge(profSrc)
	if err != nil {
		return nil, fmt.Errorf("merge profile fail: %w", err)
	}
	if err := mergedPProf.CheckValid(); err != nil {
		return nil, fmt.Errorf("invalid merged profile file: %w", err)
	}
	return mergedPProf, nil
}

func (p *PProf) Summary() (map[events.Type]*EventSummary, int64, error) {
	lang, err := parameter.VerifyLanguage(p.profiles)
	if err != nil {
		return nil, 0, fmt.Errorf("unable to resolve language: %w", err)
		//=======
		//		return nil, 0, fmt.Errorf("GetSummary VerifyLanguage err: %w", err)
		//	}
		//
		//	ok, err := IsPySpyProfile(param.Profiles, param.WorkspaceUUID)
		//	if ok && err == nil {
		//		prof := param.Profiles[0]
		//		startNanos, err := jsontoolkit.IFaceCast2Int64(prof.ProfileStart)
		//		if err != nil {
		//			return nil, 0, fmt.Errorf("resolve Profile start timestamp fail: %w", err)
		//		}
		//		endNanos, err := jsontoolkit.IFaceCast2Int64(prof.ProfileEnd)
		//		if err != nil {
		//			return nil, 0, fmt.Errorf("resolve Profile end timestamp fail: %w", err)
		//		}
		//		profileFile := storage.DefaultDiskStorage.GetProfilePath(param.WorkspaceUUID, prof.ProfileID, startNanos, events.DefaultProfileFilename)
		//
		//		summaries, err := GetPySpySummary(profileFile)
		//		if err != nil {
		//			return nil, 0, fmt.Errorf("get py-spy profile summary fail: %w", err)
		//		}
		//		return summaries, endNanos - startNanos, nil
		//	} else if err != nil {
		//		logtoolkit.Warnf("judge if profile is from py-spy err: %s", err)
		//>>>>>>> 66994b8f59cd601e6e7fbac181122f708d77ef3a:service/parsing/multiparser.go
	}

	fileSampleTypes := getFileSampleTypes(lang)
	if len(fileSampleTypes) == 0 {
		return nil, 0, fmt.Errorf("getFileSampleTypes: not supported language [%s]", lang)
	}

	summariesTypedMap := make(map[events.Type]*EventSummary)
	var totalDurationNanos int64 = 0

	filesCount := 0
	for filename, sampleTypes := range fileSampleTypes {

		mergedPProf, err := p.mergePProf(filename)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				continue
			}
			return nil, 0, fmt.Errorf("merge pprof: %w", err)
		}
		filesCount++

		if mergedPProf.DurationNanos > totalDurationNanos {
			totalDurationNanos = mergedPProf.DurationNanos
		}

		// pprof.SampleType 和 pprof.Sample[xx].Value 一一对应
		summaryMap := make(map[int]*EventSummary)

		for i, st := range mergedPProf.SampleType {

			if et, ok := sampleTypes[st.Type]; ok {

				if p.from == parameter.FromTrace && !p.ShowInTrace(et) {
					continue
				}

				if p.from == parameter.FromProfile && !p.ShowInProfile(et) {
					continue
				}

				unit, err := quantity.ParseUnit(et.GetQuantityKind(), st.Unit)
				if err != nil {
					return nil, 0, fmt.Errorf("parseUnit error: %w", err)
				}

				summaryMap[i] = &EventSummary{
					SummaryValueType: &SummaryValueType{
						Type: et,
						Unit: unit,
					},
					Value: 0,
				}
			}
		}

		for _, sample := range mergedPProf.Sample {
			// 需要进行span过滤
			if p.filterBySpan {
				spanID := parsetoolkit.GetStringLabel(sample, LabelSpanID)
				rootSpanId := parsetoolkit.GetStringLabel(sample, LabelLocalRootSpanID)
				// 没有spanID的数据去掉
				if spanID == "" {
					continue
				}
				if p.spanIDSet != nil {
					if p.spanIDSet == tracing.AllTraceSpanSet {
						if rootSpanId != p.span.SpanID {
							continue
						}
					} else if !p.spanIDSet.Contains(spanID) {
						continue
					}
				}
			}
			for i, v := range sample.Value {
				if _, ok := summaryMap[i]; ok {
					summaryMap[i].Value += v
				}
			}
		}

		for _, summary := range summaryMap {
			summariesTypedMap[summary.Type] = summary
		}
	}

	if filesCount == 0 {
		sb := &strings.Builder{}
		for i, pro := range p.profiles {
			if i > 0 {
				sb.WriteByte(';')
			}
			sb.WriteString(pro.ProfileID)
		}
		return nil, 0, fmt.Errorf("no corresponding profiling file exists, workspaceUUID [%s], profileID [%s]", p.workspaceUUID, sb.String())
	}

	return summariesTypedMap, totalDurationNanos, nil
}

// parseAndClose parse profile from a readable stream, and try to close the reader when end
func parseAndClose(r io.Reader) (*profile.Profile, error) {
	if r == nil {
		return nil, fmt.Errorf("nil reader")
	}

	if closable, ok := r.(io.Closer); ok {
		defer closable.Close()
	}

	goPprof, err := profile.Parse(r)

	if err != nil {
		return nil, fmt.Errorf("parse pprof err: %w", err)
	}

	return goPprof, nil
}

// ResolveFlameGraph (lang languages.Lang, eType events.Type, pprofSampleType string, filterBySpan bool, span *parameter.Span, spanIDSet *dql.SpanIDSet)
func (p *PProf) ResolveFlameGraph(eventType events.Type) (*pprof.Frame, AggregatorSelectSlice, error) {

	lang, err := parameter.VerifyLanguage(p.profiles)
	if err != nil {
		return nil, nil, fmt.Errorf("VerifyLanguage fail: %s", err)
	}

	sampleFile, err := GetFileByEvent(lang, eventType)
	if err != nil {
		return nil, nil, fmt.Errorf("GetFileByEvent: %s", err)
	}

	mergedPProf, err := p.mergePProf(sampleFile.Filename)
	if err != nil {
		return nil, nil, fmt.Errorf("merge pprof: %w", err)
	}

	valueIdx, valueUnit, err := p.getIdxOfTypeAndUnit(sampleFile.SampleType, mergedPProf)
	if err != nil {
		return nil, nil, fmt.Errorf("render frame: %w", err)
	}

	unit, err := quantity.ParseUnit(eventType.GetQuantityKind(), valueUnit)
	if err != nil {
		return nil, nil, fmt.Errorf("ParseUnit fail: %w", err)
	}

	rootFrame := &pprof.Frame{
		SubFrames: make(pprof.SubFrames),
	}

	aggregatorList := PythonAggregatorList
	switch lang {
	case languages.GoLang:
		aggregatorList = GoAggregatorList
	case languages.NodeJS:
		aggregatorList = PyroscopeNodeJSAggregatorList
	case languages.DotNet:
		aggregatorList = DDTraceDotnetAggregatorList
	case languages.PHP:
		aggregatorList = DDTracePHPAggregatorList
	}

	aggregatorSelectMap := make(AggregatorSelectSlice, 0, len(aggregatorList))

	for _, aggregator := range aggregatorList {
		aggregatorSelectMap = append(aggregatorSelectMap, &AggregatorSelect{
			Aggregator: aggregator,
			Mapping:    aggregator.Mapping,
			Options:    make(map[string]*AggregatorOption),
		})
	}

	totalValue := int64(0)
	for _, smp := range mergedPProf.Sample {
		if smp.Value[valueIdx] == 0 {
			// 过滤值为0的采样数据
			continue
		}

		// span 过滤，必须有spanID的才显示
		if p.filterBySpan {
			spanID := parsetoolkit.GetStringLabel(smp, LabelSpanID)
			rootSpanId := parsetoolkit.GetStringLabel(smp, LabelLocalRootSpanID)
			if spanID == "" {
				continue
			}
			if p.spanIDSet == tracing.AllTraceSpanSet {
				if rootSpanId != p.span.SpanID {
					continue
				}
			} else if p.spanIDSet != nil {
				if !p.spanIDSet.Contains(spanID) {
					continue
				}
			}
		}

		currentFrame := rootFrame

		totalValue += smp.Value[valueIdx]

		for _, aggregatorSelect := range aggregatorSelectMap {
			aggregator := aggregatorSelect.Aggregator
			identifier := aggregator.GetIdentifier(lang, smp, false)

			if _, ok := aggregatorSelect.Options[identifier]; ok {
				aggregatorSelect.Options[identifier].Value += smp.Value[valueIdx]
			} else {
				mappingValues := make([]string, 0, len(aggregator.GetMappingFuncs))
				for _, mFunc := range aggregator.GetMappingFuncs {
					mappingValues = append(mappingValues, mFunc(lang, smp, false))
				}
				aggregatorSelect.Options[identifier] = &AggregatorOption{
					Title:         aggregator.GetDisplayStr(lang, smp, false),
					Value:         smp.Value[valueIdx],
					Unit:          unit,
					MappingValues: mappingValues,
				}
			}
		}

		for i := len(smp.Location) - 1; i >= 0; i-- {
			location := smp.Location[i]
			line := location.Line[len(location.Line)-1]

			var funcIdentifier string
			//if i == 0 {
			// 最后一层必须严格相同, 不是最后一层行号不相同也允许合并
			//funcIdentifier = fmt.Sprintf("%s###%s###%d", line.Function.Filename, line.Function.Name, line.Line)
			funcIdentifier = strconv.FormatUint(location.ID, 10)
			//} else {
			//	funcIdentifier = fmt.Sprintf("%s###%s###%s", parsetoolkit.GetLabel(smp, LabelThreadID), line.Function.Filename, line.Function.Name)
			//}

			subFrame, ok := currentFrame.SubFrames[funcIdentifier]

			if ok {
				subFrame.Value += smp.Value[valueIdx]
			} else {
				subFrame = &pprof.Frame{
					Value:       smp.Value[valueIdx],
					Unit:        unit,
					Function:    getFuncNameByLine(lang, line),
					Line:        getLineByLine(lang, line),
					File:        getFileByLine(lang, line),
					Directory:   getDirectoryByLine(lang, line),
					ThreadID:    getThreadIDBySample(smp),
					ThreadName:  getThreadNameBySample(smp),
					Class:       getClassByLine(lang, line),
					Namespace:   getNamespaceByLine(lang, line),
					Assembly:    getAssemblyByLine(lang, line),
					Package:     getPackageNameByLine(lang, line),
					PrintString: GetPrintStrByLine(lang, line),
					SubFrames:   make(pprof.SubFrames),
				}
				currentFrame.SubFrames[funcIdentifier] = subFrame
			}

			currentFrame = subFrame
		}
	}

	rootFrame.Value = totalValue
	rootFrame.Unit = unit

	parsetoolkit.CalcPercentAndQuantity(rootFrame, totalValue)
	aggregatorSelectMap.CalcPercentAndQuantity(totalValue)

	return rootFrame, aggregatorSelectMap, nil
}

func (p *PProf) getIdxOfTypeAndUnit(typeName string, pprof *profile.Profile) (int, string, error) {
	for idx, st := range pprof.SampleType {
		if st.Type == typeName {
			return idx, st.Unit, nil
		}
	}
	return 0, "", fmt.Errorf("the pprof does not contain the event type: %s", typeName)
}
