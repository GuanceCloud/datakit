// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package parsing

import (
	"fmt"

	"github.com/GuanceCloud/cliutils/pprofparser/domain/events"
	"github.com/GuanceCloud/cliutils/pprofparser/domain/languages"
)

type SampleFile struct {
	Filename   string
	SampleType string
}

var typeSampleFileMap = generateDictByEvent(pprofTypeMaps)

func generateDictByEvent(pprofTypeMaps map[languages.Lang]fileSampleTypesMap) map[languages.Lang]map[events.Type]*SampleFile {
	result := make(map[languages.Lang]map[events.Type]*SampleFile, len(pprofTypeMaps))

	for lang, sampleTypes := range pprofTypeMaps {
		if _, ok := result[lang]; !ok {
			result[lang] = make(map[events.Type]*SampleFile, len(sampleTypes))
		}

		for filename, eventsMap := range sampleTypes {
			for typeName, typeID := range eventsMap {
				result[lang][typeID] = &SampleFile{Filename: filename, SampleType: typeName}
			}
		}
	}

	return result
}

var pprofTypeMaps = map[languages.Lang]fileSampleTypesMap{
	languages.Python: pyPprofTypeMaps,
	languages.GoLang: goPprofTypeMaps,
	languages.NodeJS: nodejsEventMaps,
	languages.DotNet: dotnetPProfEventMaps,
	languages.PHP:    phpEventMaps,
}

type fileSampleTypesMap map[string]map[string]events.Type

var pyPprofTypeMaps = fileSampleTypesMap{
	"prof|auto|*.pprof": {
		//		"cpu-samples":       events.CPUSamples,
		"cpu-time":          events.CPUTime,
		"wall-time":         events.WallTime,
		"exception-samples": events.ThrownExceptions,
		"lock-acquire":      events.LockAcquires,
		"lock-acquire-wait": events.LockWaitTime,
		"lock-release":      events.LockReleases,
		"lock-release-hold": events.LockedTime,
		"alloc-samples":     events.Allocations,
		"alloc-space":       events.AllocatedMemory,
		"heap-space":        events.HeapLiveSize,
	},
}
var dotnetPProfEventMaps = fileSampleTypesMap{
	"prof|auto|*.pprof": {
		"cpu":           events.CPUTime,
		"exception":     events.ThrownExceptions,
		"alloc-samples": events.Allocations,
		"alloc-size":    events.AllocatedMemory,
		"inuse-objects": events.HeapLiveObjects,
		"inuse-space":   events.HeapLiveSize,
		"wall":          events.WallTime,
		"lock-count":    events.LockAcquires,
		"lock-time":     events.LockWaitTime,
	},
}

var phpEventMaps = fileSampleTypesMap{
	"prof|auto|*.pprof": {
		"cpu-time":      events.CPUTime,
		"sample":        events.CPUSamples,
		"wall-time":     events.WallTime,
		"alloc-samples": events.Allocations,
		"alloc-size":    events.AllocatedMemory,
	},
}

var nodejsEventMaps = fileSampleTypesMap{
	"cpu.pprof": {
		"": events.CPUSamples,
	},

	"inuse_objects.pprof": {
		"": events.HeapLiveObjects,
	},

	"inuse_space.pprof": {
		"": events.HeapLiveSize,
	},
}

var goPprofTypeMaps = fileSampleTypesMap{
	"cpu.pprof|*cpu.pprof*": {
		"cpu": events.CPUTime,
	},
	"delta-heap.pprof|*delta-heap.pprof*": {
		"alloc_objects": events.Allocations,
		"alloc_space":   events.AllocatedMemory,
		"inuse_objects": events.HeapLiveObjects,
		"inuse_space":   events.HeapLiveSize,
	},
	"delta-mutex.pprof|*delta-mutex.pprof*": {
		"delay": events.Mutex,
	},
	"delta-block.pprof|*delta-block.pprof*": {
		"delay": events.Block,
	},
	"goroutines.pprof|*goroutines.pprof*": {
		"goroutine": events.Goroutines,
	},
}

func getFileSampleTypes(lang languages.Lang) fileSampleTypesMap {
	if typesMap, ok := pprofTypeMaps[lang]; ok {
		return typesMap
	}
	return nil
}

func GetFileByEvent(lang languages.Lang, typ events.Type) (*SampleFile, error) {
	typeFileMap, ok := typeSampleFileMap[lang]

	if !ok {
		return nil, fmt.Errorf("not supported lang: %s", lang)
	}

	sampleFile, ok := typeFileMap[typ]
	if !ok {
		return nil, fmt.Errorf("not supported event type: %s", typ)
	}

	return sampleFile, nil
}
