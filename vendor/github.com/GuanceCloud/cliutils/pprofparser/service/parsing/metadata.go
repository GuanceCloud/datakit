// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package parsing

import (
	"encoding/json"
	"fmt"
	"github.com/GuanceCloud/cliutils/pprofparser/domain/events"
	"github.com/GuanceCloud/cliutils/pprofparser/domain/parameter"
	"github.com/GuanceCloud/cliutils/pprofparser/service/storage"
	"github.com/GuanceCloud/cliutils/pprofparser/tools/filepathtoolkit"
	"os"
	"strings"
)

const (
	FlameGraph Format = "flamegraph" // see https://github.com/brendangregg/FlameGraph
	//Deprecated use Collapsed instead
	RawFlameGraph Format = "rawflamegraph" // flamegraph collapse
	Collapsed     Format = "collapse"      // flamegraph collapse format see https://github.com/brendangregg/FlameGraph/blob/master/stackcollapse.pl
	SpeedScope    Format = "speedscope"    // see https://github.com/jlfwong/speedscope
	JFR           Format = "jfr"           // see https://github.com/openjdk/jmc#core-api-example
	PPROF         Format = "pprof"         // see https://github.com/google/pprof/blob/main/proto/profile.proto
)

type Profiler string

const (
	Unknown       Profiler = "unknown"
	DDtrace                = "ddtrace"
	AsyncProfiler          = "async-profiler"
	PySpy
	Pyroscope = "pyroscope"
)

type Format string

type MetaData struct {
	Format      Format   `json:"format"`
	Profiler    Profiler `json:"profiler"`
	Attachments []string `json:"attachments"`
}

func ReadMetaDataFile(f string) (*MetaData, error) {
	if !filepathtoolkit.FileExists(f) && !strings.HasSuffix(f, ".json") {
		f += ".json"
	}
	data, err := os.ReadFile(f)
	if err != nil {
		return nil, fmt.Errorf("open metadata file fail: %w", err)
	}
	m := new(MetaData)
	if err := json.Unmarshal(data, m); err != nil {
		return nil, fmt.Errorf("metadata json unmarshal fail: %w", err)
	}
	return m, nil
}

func ReadMetaData(prof *parameter.Profile, workspaceUUID string) (*MetaData, error) {
	startNanos, err := prof.StartTime()
	if err != nil {
		return nil, fmt.Errorf("resolve Profile start timestamp fail: %w", err)
	}

	file := storage.DefaultDiskStorage.GetProfilePath(workspaceUUID, prof.ProfileID, startNanos, events.DefaultMetaFileName)

	return ReadMetaDataFile(file)
}
