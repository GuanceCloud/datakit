// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package parsing

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/GuanceCloud/cliutils/pprofparser/domain/events"
	"github.com/GuanceCloud/cliutils/pprofparser/domain/parameter"
	"github.com/GuanceCloud/cliutils/pprofparser/domain/pprof"
	"github.com/GuanceCloud/cliutils/pprofparser/domain/quantity"
	"github.com/GuanceCloud/cliutils/pprofparser/domain/tracing"
	"github.com/GuanceCloud/cliutils/pprofparser/service/storage"
	"github.com/GuanceCloud/cliutils/pprofparser/tools/filepathtoolkit"
	"github.com/GuanceCloud/cliutils/pprofparser/tools/logtoolkit"
	"github.com/GuanceCloud/cliutils/pprofparser/tools/parsetoolkit"
)

/*
py-spy profiler output is as below:

process 95768:"/opt/homebrew/Cellar/python@3.10/3.10.5/Frameworks/Python.framework/Versions/3.10/Resources/Python.app/Contents/MacOS/Python fibobacci.py";thread (0x100850580);<module> (/Users/zy/PycharmProjects/pyroscope-demo/fibobacci.py:14);fibonacci (/Users/zy/PycharmProjects/pyroscope-demo/fibobacci.py:8);fibonacci (/Users/zy/PycharmProjects/pyroscope-demo/fibobacci.py:8);fibonacci (/Users/zy/PycharmProjects/pyroscope-demo/fibobacci.py:8);fibonacci (/Users/zy/PycharmProjects/pyroscope-demo/fibobacci.py:8);fibonacci (/Users/zy/PycharmProjects/pyroscope-demo/fibobacci.py:8);fibonacci (/Users/zy/PycharmProjects/pyroscope-demo/fibobacci.py:8);fibonacci (/Users/zy/PycharmProjects/pyroscope-demo/fibobacci.py:8);fibonacci (/Users/zy/PycharmProjects/pyroscope-demo/fibobacci.py:8);fibonacci (/Users/zy/PycharmProjects/pyroscope-demo/fibobacci.py:8);fibonacci (/Users/zy/PycharmProjects/pyroscope-demo/fibobacci.py:8);fibonacci (/Users/zy/PycharmProjects/pyroscope-demo/fibobacci.py:8);fibonacci (/Users/zy/PycharmProjects/pyroscope-demo/fibobacci.py:8);fibonacci (/Users/zy/PycharmProjects/pyroscope-demo/fibobacci.py:8);fibonacci (/Users/zy/PycharmProjects/pyroscope-demo/fibobacci.py:8);fibonacci (/Users/zy/PycharmProjects/pyroscope-demo/fibobacci.py:5) 1
process 95768:"/opt/homebrew/Cellar/python@3.10/3.10.5/Frameworks/Python.framework/Versions/3.10/Resources/Python.app/Contents/MacOS/Python fibobacci.py";thread (0x100850580);<module> (/Users/zy/PycharmProjects/pyroscope-demo/fibobacci.py:14);fibonacci (/Users/zy/PycharmProjects/pyroscope-demo/fibobacci.py:8);fibonacci (/Users/zy/PycharmProjects/pyroscope-demo/fibobacci.py:8);fibonacci (/Users/zy/PycharmProjects/pyroscope-demo/fibobacci.py:8);fibonacci (/Users/zy/PycharmProjects/pyroscope-demo/fibobacci.py:8);fibonacci (/Users/zy/PycharmProjects/pyroscope-demo/fibobacci.py:8);fibonacci (/Users/zy/PycharmProjects/pyroscope-demo/fibobacci.py:8);fibonacci (/Users/zy/PycharmProjects/pyroscope-demo/fibobacci.py:8);fibonacci (/Users/zy/PycharmProjects/pyroscope-demo/fibobacci.py:8);fibonacci (/Users/zy/PycharmProjects/pyroscope-demo/fibobacci.py:8);fibonacci (/Users/zy/PycharmProjects/pyroscope-demo/fibobacci.py:8);fibonacci (/Users/zy/PycharmProjects/pyroscope-demo/fibobacci.py:8);fibonacci (/Users/zy/PycharmProjects/pyroscope-demo/fibobacci.py:8);fibonacci (/Users/zy/PycharmProjects/pyroscope-demo/fibobacci.py:8);fibonacci (/Users/zy/PycharmProjects/pyroscope-demo/fibobacci.py:8);fibonacci (/Users/zy/PycharmProjects/pyroscope-demo/fibobacci.py:7) 1
process 95768:"/opt/homebrew/Cellar/python@3.10/3.10.5/Frameworks/Python.framework/Versions/3.10/Resources/Python.app/Contents/MacOS/Python fibobacci.py";thread (0x100850580);<module> (/Users/zy/PycharmProjects/pyroscope-demo/fibobacci.py:14);fibonacci (/Users/zy/PycharmProjects/pyroscope-demo/fibobacci.py:8);fibonacci (/Users/zy/PycharmProjects/pyroscope-demo/fibobacci.py:8);fibonacci (/Users/zy/PycharmProjects/pyroscope-demo/fibobacci.py:8);fibonacci (/Users/zy/PycharmProjects/pyroscope-demo/fibobacci.py:8);fibonacci (/Users/zy/PycharmProjects/pyroscope-demo/fibobacci.py:8);fibonacci (/Users/zy/PycharmProjects/pyroscope-demo/fibobacci.py:8);fibonacci (/Users/zy/PycharmProjects/pyroscope-demo/fibobacci.py:8);fibonacci (/Users/zy/PycharmProjects/pyroscope-demo/fibobacci.py:8);fibonacci (/Users/zy/PycharmProjects/pyroscope-demo/fibobacci.py:8);fibonacci (/Users/zy/PycharmProjects/pyroscope-demo/fibobacci.py:8);fibonacci (/Users/zy/PycharmProjects/pyroscope-demo/fibobacci.py:8);fibonacci (/Users/zy/PycharmProjects/pyroscope-demo/fibobacci.py:8);fibonacci (/Users/zy/PycharmProjects/pyroscope-demo/fibobacci.py:8);fibonacci (/Users/zy/PycharmProjects/pyroscope-demo/fibobacci.py:8);fibonacci (/Users/zy/PycharmProjects/pyroscope-demo/fibobacci.py:8);fibonacci (/Users/zy/PycharmProjects/pyroscope-demo/fibobacci.py:5) 1
process 95768:"/opt/homebrew/Cellar/python@3.10/3.10.5/Frameworks/Python.framework/Versions/3.10/Resources/Python.app/Contents/MacOS/Python fibobacci.py";thread (0x100850580);<module> (/Users/zy/PycharmProjects/pyroscope-demo/fibobacci.py:14);fibonacci (/Users/zy/PycharmProjects/pyroscope-demo/fibobacci.py:8);fibonacci (/Users/zy/PycharmProjects/pyroscope-demo/fibobacci.py:8);fibonacci (/Users/zy/PycharmProjects/pyroscope-demo/fibobacci.py:8);fibonacci (/Users/zy/PycharmProjects/pyroscope-demo/fibobacci.py:8);fibonacci (/Users/zy/PycharmProjects/pyroscope-demo/fibobacci.py:8);fibonacci (/Users/zy/PycharmProjects/pyroscope-demo/fibobacci.py:8);fibonacci (/Users/zy/PycharmProjects/pyroscope-demo/fibobacci.py:8);fibonacci (/Users/zy/PycharmProjects/pyroscope-demo/fibobacci.py:8);fibonacci (/Users/zy/PycharmProjects/pyroscope-demo/fibobacci.py:8);fibonacci (/Users/zy/PycharmProjects/pyroscope-demo/fibobacci.py:8);fibonacci (/Users/zy/PycharmProjects/pyroscope-demo/fibobacci.py:8);fibonacci (/Users/zy/PycharmProjects/pyroscope-demo/fibobacci.py:8);fibonacci (/Users/zy/PycharmProjects/pyroscope-demo/fibobacci.py:8);fibonacci (/Users/zy/PycharmProjects/pyroscope-demo/fibobacci.py:5) 1

nginx;/usr/sbin/nginx+0x24678;__libc_start_main;main;ngx_master_process_cycle;/usr/sbin/nginx+0x4f0d8;ngx_spawn_process;/usr/sbin/nginx+0x4fa44;ngx_process_events_and_timers;/usr/sbin/nginx+0x51f54;epoll_pwait 1

*/

var processRegExp = regexp.MustCompile(`^process\s+\d+:"`)
var threadRexExp = regexp.MustCompile(`^thread\s+\([a-zA-Z\d]+\)$`)
var stackTraceRegExp = regexp.MustCompile(`^(\S+)(?: +\(([^:]+):(\d+)\))?`)

type Collapse struct {
	workspaceUUID string
	profiles      []*parameter.Profile
	filterBySpan  bool
	spanIDSet     *tracing.SpanIDSet
}

func NewCollapse(workspaceUUID string, profiles []*parameter.Profile,
	filterBySpan bool, spanIDSet *tracing.SpanIDSet) *Collapse {
	return &Collapse{
		workspaceUUID: workspaceUUID,
		profiles:      profiles,
		filterBySpan:  filterBySpan,
		spanIDSet:     spanIDSet,
	}
}

func summary(filename string) (map[events.Type]*EventSummary, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("open profile file [%s] fail: %w", filename, err)
	}
	defer f.Close() //nolint:errcheck

	sampleSummary := &EventSummary{
		SummaryValueType: &SummaryValueType{
			Type: events.CPUSamples,
			Unit: quantity.CountUnit,
		},
		Value: 0,
	}

	spySummaries := map[events.Type]*EventSummary{
		events.CPUSamples: sampleSummary,
	}

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if len(line) < 2 {
			continue
		}

		blankIdx := strings.LastIndexByte(line, ' ')
		if blankIdx < 0 {
			logtoolkit.Errorf("py-spy profile doesn't contain any blank [line: %s]", line)
			continue
		}
		n, err := strconv.ParseInt(strings.TrimSpace(line[blankIdx+1:]), 10, 64)
		if err != nil {
			logtoolkit.Errorf("resolve sample count fail [line: %s]: %w", line, err)
			continue
		}
		sampleSummary.Value += n
	}
	return spySummaries, nil
}

func (p *Collapse) Summary() (map[events.Type]*EventSummary, int64, error) {

	prof := p.profiles[0]

	startNanos, err := prof.StartTime()
	if err != nil {
		return nil, 0, fmt.Errorf("resolve Profile start timestamp fail: %w", err)
	}
	endNanos, err := prof.EndTime()
	if err != nil {
		return nil, 0, fmt.Errorf("resolve Profile end timestamp fail: %w", err)
	}

	filename := storage.DefaultDiskStorage.GetProfilePath(p.workspaceUUID, prof.ProfileID, startNanos, events.DefaultProfileFilename)

	summaries, err := summary(filename)
	if err != nil {
		return nil, 0, fmt.Errorf("resolve collapse summary fail: %w", err)
	}

	return summaries, endNanos - startNanos, nil
}

func IsCollapseProfile(profiles []*parameter.Profile, workspaceUUID string) (bool, error) {
	// 当前 py-spy 一次只有一条profile数据
	if len(profiles) > 1 {
		return false, nil
	}

	metadata, err := ReadMetaData(profiles[0], workspaceUUID)
	if err != nil {
		return false, fmt.Errorf("read py-spy metadata file fail: %w", err)
	}

	return metadata.Format == RawFlameGraph || metadata.Format == Collapsed, nil
}

func (p *Collapse) ResolveFlameGraph(_ events.Type) (*pprof.Frame, AggregatorSelectSlice, error) {

	prof := p.profiles[0]

	startNanos, err := prof.StartTime()
	if err != nil {
		return nil, nil, fmt.Errorf("invalid profile start: %w", err)
	}
	file := storage.DefaultDiskStorage.GetProfilePath(p.workspaceUUID, prof.ProfileID, startNanos, events.DefaultProfileFilename)

	f, err := os.Open(file)
	if err != nil {
		return nil, nil, fmt.Errorf("open py-spy profile file fail: %w", err)
	}
	defer f.Close() //nolint:errcheck

	scanner := bufio.NewScanner(f)

	rootFrame := &pprof.Frame{
		SubFrames: make(pprof.SubFrames),
	}
	totalValue := int64(0)

	aggregatorSelects := make(AggregatorSelectSlice, 0, len(SpyAggregatorList))

	for _, aggregator := range SpyAggregatorList {
		aggregatorSelects = append(aggregatorSelects, &AggregatorSelect{
			Aggregator: aggregator,
			Mapping:    aggregator.Mapping,
			Options:    make(map[string]*AggregatorOption),
		})
	}

	for scanner.Scan() {

		line := strings.TrimSpace(scanner.Text())
		if len(line) == 0 {
			continue
		}

		stacks := strings.Split(line, ";")
		if len(stacks) == 0 {
			continue
		}

		lastStack := strings.TrimSpace(stacks[len(stacks)-1])
		if !stackTraceRegExp.MatchString(lastStack) {
			logtoolkit.Warnf("The last stacktrace not match with the regexp [%s], the stacktrace [%s]", stackTraceRegExp.String(), lastStack)
			continue
		}
		blankIdx := strings.LastIndexByte(lastStack, ' ')
		if blankIdx < 0 {
			logtoolkit.Warnf("Can not find any blank from [%s]", lastStack)
			continue
		}

		sampleCount, err := strconv.ParseInt(lastStack[blankIdx+1:], 10, 64)
		if err != nil {
			logtoolkit.Warnf("Can not resolve sample count from [%s]", lastStack)
			continue
		}
		totalValue += sampleCount

		currentFrame := rootFrame
		threadName := "<unknown>"

		for idx, stack := range stacks {
			stack = strings.TrimSpace(stack)
			matches := stackTraceRegExp.FindStringSubmatch(stack)
			if len(matches) != 4 {
				if processRegExp.MatchString(stack) {
					continue
				} else if threadRexExp.MatchString(stack) {
					threadName = stack
					continue
				} else {
					return nil, nil, fmt.Errorf("resolve stacktrace from profiling file fail")
				}
			}

			funcName, codeFile, lineNoStr := matches[1], matches[2], matches[3]

			if codeFile == "" {
				codeFile = "<unknown>"
			}

			var lineNo int64 = -1
			if lineNoStr != "" {
				lineNo, _ = strconv.ParseInt(lineNoStr, 10, 64)
			}

			funcIdentifier := fmt.Sprintf("%s###%s###%s###%d", threadName, codeFile, funcName, lineNo)

			if idx == len(stacks)-1 {

				for _, aggregatorSelect := range aggregatorSelects {

					var identifier string
					var displayStr string
					var mappingValues []string

					switch aggregatorSelect.Aggregator {
					case Function:
						identifier = fmt.Sprintf("%s###%s", codeFile, funcName)
						displayStr = GetSpyPrintStr(funcName, codeFile)
						mappingValues = []string{funcName}
					case FunctionLine:
						identifier = fmt.Sprintf("%s###%s###%d", codeFile, funcName, lineNo)
						displayStr = fmt.Sprintf("%s(%s:L#%d)", funcName, filepathtoolkit.BaseName(codeFile), lineNo)
						mappingValues = []string{funcName, fmt.Sprintf("%d", lineNo)}
					case Directory:
						identifier = filepathtoolkit.DirName(codeFile)
						displayStr = identifier
						mappingValues = []string{identifier}
					case File:
						identifier = codeFile
						displayStr = codeFile
						mappingValues = []string{codeFile}
					case ThreadName:
						identifier = threadName
						displayStr = threadName
						mappingValues = []string{threadName}
					}

					if _, ok := aggregatorSelect.Options[identifier]; ok {
						aggregatorSelect.Options[identifier].Value += sampleCount
					} else {
						aggregatorSelect.Options[identifier] = &AggregatorOption{
							Title:         displayStr,
							Value:         sampleCount,
							Unit:          quantity.CountUnit,
							MappingValues: mappingValues,
						}
					}
				}
			}

			subFrame, ok := currentFrame.SubFrames[funcIdentifier]

			if ok {
				subFrame.Value += sampleCount
			} else {
				subFrame = &pprof.Frame{
					Value:       sampleCount,
					Unit:        quantity.CountUnit,
					Function:    funcName,
					Line:        lineNo,
					File:        codeFile,
					Directory:   filepathtoolkit.DirName(codeFile),
					ThreadID:    "",
					ThreadName:  threadName,
					Package:     "",
					PrintString: GetSpyPrintStr(funcName, codeFile),
					SubFrames:   make(pprof.SubFrames),
				}
				currentFrame.SubFrames[funcIdentifier] = subFrame
			}

			currentFrame = subFrame
		}
	}

	rootFrame.Value = totalValue
	rootFrame.Unit = quantity.CountUnit

	parsetoolkit.CalcPercentAndQuantity(rootFrame, totalValue)
	aggregatorSelects.CalcPercentAndQuantity(totalValue)
	return rootFrame, aggregatorSelects, nil
}

func ParseRawFlameGraph(filename string) (*pprof.Frame, AggregatorSelectSlice, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, nil, fmt.Errorf("open py-spy profile file fail: %w", err)
	}
	defer f.Close() //nolint:errcheck

	scanner := bufio.NewScanner(f)

	rootFrame := &pprof.Frame{
		SubFrames: make(pprof.SubFrames),
	}
	totalValue := int64(0)

	aggregatorSelects := make(AggregatorSelectSlice, 0, len(SpyAggregatorList))

	for _, aggregator := range SpyAggregatorList {
		aggregatorSelects = append(aggregatorSelects, &AggregatorSelect{
			Aggregator: aggregator,
			Mapping:    aggregator.Mapping,
			Options:    make(map[string]*AggregatorOption),
		})
	}

	for scanner.Scan() {

		line := strings.TrimSpace(scanner.Text())
		if len(line) == 0 {
			continue
		}

		stacks := strings.Split(line, ";")
		if len(stacks) == 0 {
			continue
		}

		lastStack := strings.TrimSpace(stacks[len(stacks)-1])
		if !stackTraceRegExp.MatchString(lastStack) {
			logtoolkit.Warnf("The last stacktrace not match with the regexp [%s], the stacktrace [%s]", stackTraceRegExp.String(), lastStack)
			continue
		}
		blankIdx := strings.LastIndexByte(lastStack, ' ')
		if blankIdx < 0 {
			logtoolkit.Warnf("Can not find any blank from [%s]", lastStack)
			continue
		}

		sampleCount, err := strconv.ParseInt(lastStack[blankIdx+1:], 10, 64)
		if err != nil {
			logtoolkit.Warnf("Can not resolve sample count from [%s]", lastStack)
			continue
		}
		totalValue += sampleCount

		currentFrame := rootFrame
		threadName := "<unknown>"

		for idx, stack := range stacks {
			stack = strings.TrimSpace(stack)
			matches := stackTraceRegExp.FindStringSubmatch(stack)
			if len(matches) != 4 {
				if processRegExp.MatchString(stack) {
					continue
				} else if threadRexExp.MatchString(stack) {
					threadName = stack
					continue
				} else {
					return nil, nil, fmt.Errorf("resolve stacktrace from profiling file fail")
				}
			}

			funcName, codeFile, lineNoStr := matches[1], matches[2], matches[3]

			if codeFile == "" {
				codeFile = "<unknown>"
			}

			var lineNo int64 = -1
			if lineNoStr != "" {
				lineNo, _ = strconv.ParseInt(lineNoStr, 10, 64)
			}

			funcIdentifier := fmt.Sprintf("%s###%s###%s###%d", threadName, codeFile, funcName, lineNo)

			if idx == len(stacks)-1 {

				for _, aggregatorSelect := range aggregatorSelects {

					var identifier string
					var displayStr string
					var mappingValues []string

					switch aggregatorSelect.Aggregator {
					case Function:
						identifier = fmt.Sprintf("%s###%s", codeFile, funcName)
						displayStr = GetSpyPrintStr(funcName, codeFile)
						mappingValues = []string{funcName}
					case FunctionLine:
						identifier = fmt.Sprintf("%s###%s###%d", codeFile, funcName, lineNo)
						displayStr = fmt.Sprintf("%s(%s:L#%d)", funcName, filepathtoolkit.BaseName(codeFile), lineNo)
						mappingValues = []string{funcName, fmt.Sprintf("%d", lineNo)}
					case Directory:
						identifier = filepathtoolkit.DirName(codeFile)
						displayStr = identifier
						mappingValues = []string{identifier}
					case File:
						identifier = codeFile
						displayStr = codeFile
						mappingValues = []string{codeFile}
					case ThreadName:
						identifier = threadName
						displayStr = threadName
						mappingValues = []string{threadName}
					}

					if _, ok := aggregatorSelect.Options[identifier]; ok {
						aggregatorSelect.Options[identifier].Value += sampleCount
					} else {
						aggregatorSelect.Options[identifier] = &AggregatorOption{
							Title:         displayStr,
							Value:         sampleCount,
							Unit:          quantity.CountUnit,
							MappingValues: mappingValues,
						}
					}
				}
			}

			subFrame, ok := currentFrame.SubFrames[funcIdentifier]

			if ok {
				subFrame.Value += sampleCount
			} else {
				subFrame = &pprof.Frame{
					Value:       sampleCount,
					Unit:        quantity.CountUnit,
					Function:    funcName,
					Line:        lineNo,
					File:        codeFile,
					Directory:   filepathtoolkit.DirName(codeFile),
					ThreadID:    "",
					ThreadName:  threadName,
					Package:     "",
					PrintString: GetSpyPrintStr(funcName, codeFile),
					SubFrames:   make(pprof.SubFrames),
				}
				currentFrame.SubFrames[funcIdentifier] = subFrame
			}

			currentFrame = subFrame
		}
	}

	rootFrame.Value = totalValue
	rootFrame.Unit = quantity.CountUnit

	parsetoolkit.CalcPercentAndQuantity(rootFrame, totalValue)
	aggregatorSelects.CalcPercentAndQuantity(totalValue)
	return rootFrame, aggregatorSelects, nil
}
