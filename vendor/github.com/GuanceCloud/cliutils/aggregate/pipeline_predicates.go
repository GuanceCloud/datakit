package aggregate

import (
	fp "github.com/GuanceCloud/cliutils/filter"
)

// Span-predicate compilation: statically compiles SamplingPipeline conditions into
// predicates over the DataPacket span summary (Pred* / *DurationUs fields),
// allowing decisions without decompressing the payload.
// Conditions that cannot be covered by the summary (falling back to decompressed
// walk) return ok=false at compile time.

type spanPredicateExpr func(*DataPacket) bool

// packetHasSpanPredicates reports whether the packet carries span predicates
// computed by PickTrace during grouping. Packets with all-zero predicates
// (old DataKit versions / hand-crafted / abnormal data without a duration field)
// are not trusted and decisions fall back to the decompressed walk to keep
// behavior identical to the legacy logic.
func packetHasSpanPredicates(packet *DataPacket) bool {
	if packet == nil {
		return false
	}

	return packet.PredError || packet.PredHttpError || packet.PredBizError ||
		packet.PredTraceKeep || packet.MaxSpanDurationUs > 0 ||
		packet.RootDurationUs > 0 || packet.MaxNonrootDurationUs > 0
}

const (
	predParentNone = iota
	predParentRoot
	predParentNonRoot
)

type spanPred struct {
	flag       spanPredicateExpr // direct boolean predicate (error/5xx/biz/trace_keep, etc.)
	parentID   int               // predParentNone / predParentRoot / predParentNonRoot
	durationGT int64             // duration threshold (microseconds); 0 means unconstrained
}

// compilePipelinePredicate compiles a single pipeline.
// A match-all probabilistic sampler (no condition) always matches; condition
// pipelines are analyzed one by one and fail when any of them cannot be covered.
func compilePipelinePredicate(pipeline *SamplingPipeline) (spanPredicateExpr, bool) {
	if pipeline == nil {
		return nil, false
	}

	if pipeline.isMatchAllSampling() {
		return func(*DataPacket) bool { return true }, true
	}

	if pipeline.conds == nil {
		return nil, false
	}

	var exprs []spanPredicateExpr
	for _, node := range pipeline.conds {
		wc, ok := node.(*fp.WhereCondition)
		if !ok {
			return nil, false
		}
		expr, ok2 := compileWhereCondition(wc)
		if !ok2 || expr == nil {
			return nil, false
		}
		exprs = append(exprs, expr)
	}

	// WhereConditions.Eval semantics: a hit occurs when any WhereCondition matches.
	return func(packet *DataPacket) bool {
		for _, expr := range exprs {
			if expr(packet) {
				return true
			}
		}
		return false
	}, true
}

// compileWhereCondition compiles the AND combination inside one {}.
func compileWhereCondition(wc *fp.WhereCondition) (spanPredicateExpr, bool) {
	if wc == nil {
		return nil, false
	}

	combined := spanPred{}
	for _, node := range wc.Conds() {
		be, ok := node.(*fp.BinaryExpr)
		if !ok {
			return nil, false
		}
		part, ok2 := compilePred(be)
		if !ok2 {
			return nil, false
		}
		merged, ok3 := mergePreds(combined, part)
		if !ok3 {
			return nil, false
		}
		combined = merged
	}

	return combined.eval(), true
}

func compilePred(be *fp.BinaryExpr) (spanPred, bool) {
	if be == nil {
		return spanPred{}, false
	}

	switch be.Op {
	case fp.AND:
		l, ok1 := compilePred(asBinaryExpr(be.LHS))
		r, ok2 := compilePred(asBinaryExpr(be.RHS))
		if !ok1 || !ok2 {
			return spanPred{}, false
		}
		return mergePreds(l, r)
	case fp.OR:
		l, ok1 := compilePred(asBinaryExpr(be.LHS))
		r, ok2 := compilePred(asBinaryExpr(be.RHS))
		if !ok1 || !ok2 {
			return spanPred{}, false
		}
		le, re := l.eval(), r.eval()
		if le == nil || re == nil {
			return spanPred{}, false
		}
		return spanPred{flag: func(packet *DataPacket) bool { return le(packet) || re(packet) }}, true
	default:
		return compileAtomic(be)
	}
}

func asBinaryExpr(node fp.Node) *fp.BinaryExpr {
	switch e := node.(type) {
	case *fp.BinaryExpr:
		return e
	case *fp.ParenExpr:
		return asBinaryExpr(e.Param)
	default:
		return nil
	}
}

// mergePreds merges two predicate fragments under AND semantics.
// It returns false when parent constraints conflict (root and non-root together)
// or the combination cannot be expressed.
func mergePreds(a, b spanPred) (spanPred, bool) {
	out := spanPred{}

	switch {
	case a.flag != nil && b.flag != nil:
		fa, fb := a.flag, b.flag
		out.flag = func(packet *DataPacket) bool { return fa(packet) && fb(packet) }
	case a.flag != nil:
		out.flag = a.flag
	case b.flag != nil:
		out.flag = b.flag
	}

	switch {
	case a.parentID != predParentNone && b.parentID != predParentNone && a.parentID != b.parentID:
		return spanPred{}, false
	case a.parentID != predParentNone:
		out.parentID = a.parentID
	case b.parentID != predParentNone:
		out.parentID = b.parentID
	}

	out.durationGT = max(a.durationGT, b.durationGT)

	return out, true
}

// eval produces the final decision function; nil means the combination
// cannot be covered by predicates.
func (p *spanPred) eval() spanPredicateExpr {
	var dur spanPredicateExpr
	switch {
	case p.parentID == predParentRoot && p.durationGT > 0:
		threshold := p.durationGT
		dur = func(packet *DataPacket) bool { return packet.RootDurationUs > threshold }
	case p.parentID == predParentNonRoot && p.durationGT > 0:
		threshold := p.durationGT
		dur = func(packet *DataPacket) bool { return packet.MaxNonrootDurationUs > threshold }
	case p.parentID == predParentNone && p.durationGT > 0:
		threshold := p.durationGT
		dur = func(packet *DataPacket) bool { return packet.MaxSpanDurationUs > threshold }
	}

	switch {
	case p.flag != nil && dur != nil:
		flag := p.flag
		return func(packet *DataPacket) bool { return flag(packet) && dur(packet) }
	case p.flag != nil:
		return p.flag
	case dur != nil:
		return dur
	default:
		return nil
	}
}

// compileAtomic compiles an atomic comparison condition.
func compileAtomic(be *fp.BinaryExpr) (spanPred, bool) {
	if be == nil {
		return spanPred{}, false
	}

	ident, ok := be.LHS.(*fp.Identifier)
	if !ok {
		return spanPred{}, false
	}

	switch ident.Name {
	case "status":
		if be.Op == fp.EQ {
			if s, ok := stringLiteralValue(be.RHS); ok && s == "error" {
				return spanPred{flag: func(packet *DataPacket) bool { return packet.PredError }}, true
			}
		}
	case "http_status_code":
		if be.Op == fp.MATCH && isHTTPErrorRegex(be.RHS) {
			return spanPred{flag: func(packet *DataPacket) bool { return packet.PredHttpError }}, true
		}
	case "body_code":
		if be.Op == fp.NOT_IN && isBizErrorList(be.RHS) {
			return spanPred{flag: func(packet *DataPacket) bool { return packet.PredBizError }}, true
		}
	case "trace_keep":
		if be.Op == fp.EQ {
			if b, ok := boolLiteralValue(be.RHS); ok && b {
				return spanPred{flag: func(packet *DataPacket) bool { return packet.PredTraceKeep }}, true
			}
		}
	case "parent_id":
		if s, ok := stringLiteralValue(be.RHS); ok && s == "0" {
			switch be.Op {
			case fp.EQ:
				return spanPred{parentID: predParentRoot}, true
			case fp.NEQ:
				return spanPred{parentID: predParentNonRoot}, true
			}
		}
	case "duration":
		if be.Op == fp.GT {
			if n, ok := numberLiteralInt(be.RHS); ok && n > 0 {
				return spanPred{durationGT: n}, true
			}
		}
	}

	return spanPred{}, false
}

func stringLiteralValue(node fp.Node) (string, bool) {
	s, ok := node.(*fp.StringLiteral)
	if !ok {
		return "", false
	}
	return s.Val, true
}

func boolLiteralValue(node fp.Node) (bool, bool) {
	b, ok := node.(*fp.BoolLiteral)
	if !ok {
		return false, false
	}
	return b.Val, true
}

func numberLiteralInt(node fp.Node) (int64, bool) {
	n, ok := node.(*fp.NumberLiteral)
	if !ok || !n.IsInt {
		return 0, false
	}
	return n.Int, true
}

// isHTTPErrorRegex reports whether the MATCH regex is exactly equivalent to
// a 4xx/5xx status-code match. Only known equivalent forms are recognized;
// other regexes fall back to the decompressed walk to keep semantics unchanged.
func isHTTPErrorRegex(node fp.Node) bool {
	list, ok := node.(fp.NodeList)
	if !ok || len(list) != 1 {
		return false
	}

	re, ok := list[0].(*fp.Regex)
	if !ok {
		return false
	}

	switch re.Regex {
	case "^[45][0-9][0-9]$",
		"^4[0-9][0-9]$|^5[0-9][0-9]$",
		"^[45]\\d\\d$",
		"^4\\d\\d$|^5\\d\\d$",
		"^[45][0-9]{2}$",
		"^4[0-9]{2}$|^5[0-9]{2}$":
		return true
	default:
		return false
	}
}

// isBizErrorList reports whether the NOT_IN list is exactly ["0", "200", null].
func isBizErrorList(node fp.Node) bool {
	list, ok := node.(fp.NodeList)
	if !ok || len(list) != 3 {
		return false
	}

	hasZero, has200, hasNil := false, false, false
	for _, item := range list {
		switch v := item.(type) {
		case *fp.StringLiteral:
			switch v.Val {
			case "0":
				hasZero = true
			case "200":
				has200 = true
			}
		case *fp.NilLiteral:
			hasNil = true
		}
	}

	return hasZero && has200 && hasNil
}
