package parsing

import (
	"fmt"
	"github.com/GuanceCloud/cliutils/pprofparser/domain/events"
	"github.com/GuanceCloud/cliutils/pprofparser/domain/languages"
	"github.com/GuanceCloud/cliutils/pprofparser/domain/parameter"
	"github.com/GuanceCloud/cliutils/pprofparser/domain/pprof"
	"github.com/GuanceCloud/cliutils/pprofparser/domain/tracing"
	"github.com/GuanceCloud/cliutils/pprofparser/tools/logtoolkit"
)

type Parser interface {
	Summary() (map[events.Type]*EventSummary, int64, error)
	ResolveFlameGraph(eventType events.Type) (*pprof.Frame, AggregatorSelectSlice, error)
}

type SummaryValueType = pprof.SummaryValueType
type EventSummary = pprof.EventSummary
type SummaryCollection = pprof.SummaryCollection

func GetSummary(param parameter.SummaryParam, filterBySpan bool, spanIDSet *tracing.SpanIDSet) (map[events.Type]*EventSummary, int64, error) {

	isCollapseProfile := false
	meta, err := ReadMetaData(param.Profiles[0], param.WorkspaceUUID)
	if err != nil {
		logtoolkit.Warnf("unable to read profile metadata: %s", err)
	} else {
		if meta.Format == RawFlameGraph || meta.Format == Collapsed {
			isCollapseProfile = true
		}
	}

	if isCollapseProfile {
		return NewCollapse(param.WorkspaceUUID, param.Profiles, filterBySpan, spanIDSet).Summary()
	}

	lang, err := parameter.VerifyLanguage(param.Profiles)
	if err != nil {
		return nil, 0, fmt.Errorf("invalid language: %w", err)
	}
	var ctl DisplayCtl
	if meta != nil && meta.Profiler == Pyroscope && lang == languages.NodeJS {
		ctl = new(PyroscopeNodejs)
	} else {
		ctl = new(DDTrace)
	}

	parser := NewPProfParser(param.From, param.WorkspaceUUID, param.Profiles,
		filterBySpan, &param.Span, spanIDSet, ctl)
	return parser.Summary()
}
