package dql

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	ifxModel "github.com/influxdata/influxdb1-client/models"
	"github.com/sgreben/piecewiselinear"

	"gitlab.jiagouyun.com/cloudcare-tools/kodo/config"
	"gitlab.jiagouyun.com/cloudcare-tools/kodo/dql/parser"
	"gitlab.jiagouyun.com/cloudcare-tools/kodo/utils"
)

const notSLimit, defaultSLimit, maxSLimit = -1, 20 /* disable SLIMIT */, 100

func newDQLParamWithQuery(sq *singleQuery) (*parser.ExtraParam, error) {
	p := &parser.ExtraParam{
		GroupByPointNum: sq.MaxPoint,
		Limit:           sq.Limit,
		MaxLimit:        config.C.DQL.MaxLimit,
		MaxSLimit:       maxSLimit,
		DefaultLimit:    config.C.DQL.DefaultLimit,
		SearchAfter:     sq.SearchAfter,
		DefaultSLimit: func() int64 {
			if sq.DisableSlimit {
				return notSLimit
			}
			return defaultSLimit
		}(),
		Offset:    sq.Offset,
		Highlight: sq.Highlight,
	}

	if len(sq.TimeRange) == 1 {
		st := time.Unix(0, sq.TimeRange[0]*int64(time.Millisecond))
		p.StartTime = &st
	}

	if len(sq.TimeRange) >= 2 {
		st := time.Unix(0, sq.TimeRange[0]*int64(time.Millisecond))
		et := time.Unix(0, sq.TimeRange[1]*int64(time.Millisecond))
		p.StartTime = &st
		p.EndTime = &et
	}

	if sq.Conditions != "" {
		if binaryExpr, err := parser.ParseBinaryExpr(sq.Conditions); err != nil {
			return nil, err
		} else {
			p.Condition = binaryExpr

		}
	}

	if sq.DisableMultipleField {
		p.TargetsNum = 1 // singleQuery
	}

	for _, elem := range sq.OrderBy {
		for column, opt := range elem {
			switch opt {
			case "asc":
				p.OrderBy = append(p.OrderBy,
					map[string]parser.OrderType{column: parser.OrderAsc})
			case "desc":
				p.OrderBy = append(p.OrderBy,
					map[string]parser.OrderType{column: parser.OrderDesc})
			default:
				return nil, fmt.Errorf("invalid orderby param, only accept asc and desc, got: %s", opt)
			}
		}
	}

	if sq.MaxDuration != "" {
		if maxDuration, err := utils.ParseDuration(sq.MaxDuration); err != nil {
			return nil, fmt.Errorf("parse rewrite maxDuration failed: %s", err.Error())
		} else {
			p.MaxDuration = maxDuration
		}
	} else {
		p.MaxDuration = config.C.DQL.MaxDuration
	}

	return p, nil
}

func addAggrOnDFQuery(ast *parser.DFQuery) {
	if ast == nil {
		return
	}

	if ast.TimeRange == nil {
		return
	}

	if ast.TimeRange.Resolution == nil {
		return
	}

	if ast.TimeRange.Resolution.Duration > 0 {
		const defaultAggrName = `last`

		if !hasAggrOnTargets(ast.Targets) {
			doAddAggrOnTargets(ast.Targets, defaultAggrName)
		}
	}
}

func hasAggrOnTargets(targets []*parser.Target) bool {
	exist := false
	for _, target := range targets {
		switch t := target.Col.(type) {
		case *parser.FuncExpr:
			exist = hasAggr(target.Col)

			if exist {
				return true
			}

		case *parser.BinaryExpr, *parser.ParenExpr:
			exist = walkTargetAggr(target.Col, func(expr parser.Node) bool {
				return hasAggr(expr)
			})

			if exist {
				return true
			}

		default:
			l.Debugf("skip field type %v", t)
		}
	}
	return false
}

type walkf func(parser.Node) bool

func walkTargetAggr(expr parser.Node, f walkf) bool {
	exist := false
	switch t := expr.(type) {

	case *parser.FuncExpr:
		exist = f(expr)

	case *parser.BinaryExpr:
		exist = walkTargetAggr(expr.(*parser.BinaryExpr).LHS, f)
		if exist {
			return true
		}

		exist = walkTargetAggr(expr.(*parser.BinaryExpr).RHS, f)

	case *parser.ParenExpr:
		exist = walkTargetAggr(expr.(*parser.ParenExpr).Param, f)

	default:
		l.Debugf("skip field type %v", t)
	}

	return exist
}

func hasAggr(expr parser.Node) bool {
	switch t := expr.(type) {
	case *parser.FuncExpr:
		callName := strings.ToUpper(expr.(*parser.FuncExpr).Name)
		switch callName {
		case `COUNT`, `COUNT_DISTINCT`, `DISTINCT`, `AVG`, `SUM`, `LAST`, `TOP`, `FIRST`,
			`BOTTOM`, `MAX`, `MIN`, `PERCENTILE`:
			return true
		}

	default:
		l.Debugf("skip expr type %v", t)
	}
	return false
}

func aggrVarRef(vref parser.Node, funcName string) parser.Node {
	return &parser.FuncExpr{
		Name:  funcName,
		Param: []parser.Node{vref},
	}
}

func doAddAggrOnTargets(tgs []*parser.Target, fname string) {
	for idx, tg := range tgs {

		switch tg.Col.(type) {
		case *parser.Identifier, *parser.StringLiteral:
			aggrf := aggrVarRef(tg.Col, fname)
			tgs[idx].Col = aggrf

		case *parser.BinaryExpr, *parser.ParenExpr, *parser.FuncExpr:
			expr := walkNodeVarRef(tg.Col, func(vref parser.Node) parser.Node {
				return aggrVarRef(vref, fname)
			})
			tgs[idx].Col = expr

		default:

		}
	}
}

type walkVarRef func(parser.Node) parser.Node

// walk througth @bexpr, geting all varref handled by @f
func walkNodeVarRef(bexpr parser.Node, f walkVarRef) parser.Node {

	switch t := bexpr.(type) {
	case *parser.Identifier, *parser.StringLiteral:
		bexpr = f(bexpr)

	case *parser.BinaryExpr:
		bexpr.(*parser.BinaryExpr).LHS = walkNodeVarRef(bexpr.(*parser.BinaryExpr).LHS, f)
		bexpr.(*parser.BinaryExpr).RHS = walkNodeVarRef(bexpr.(*parser.BinaryExpr).RHS, f)

	case *parser.ParenExpr:
		bexpr.(*parser.ParenExpr).Param = walkNodeVarRef(bexpr.(*parser.ParenExpr).Param, f)

	case *parser.FuncExpr:
		for i, x := range bexpr.(*parser.FuncExpr).Param {
			_ = x
			bexpr.(*parser.FuncExpr).Param[i] = walkNodeVarRef(bexpr.(*parser.FuncExpr).Param[i], f)
		}
	default:
		l.Warnf("unexpected bexpr type %v", t)

	}
	return bexpr

}

/*
 *  RewriteResults: Fill() RenameColumns()
 *
 */

type RewriteData struct {
	m   *parser.DFQuery
	res []ifxModel.Row
}

func (r *RewriteData) RenameColumns() error {
	// no targets
	if r.m.IsAllTargets() {
		return nil
	}

	for _, row := range r.res {
		for idx := 0; idx < len(r.m.Targets); idx++ {
			// row.Columns[0] == "time"
			row.Columns[idx+1] = r.m.Targets[idx].String2()
		}
	}

	return nil
}

func (r *RewriteData) FillResult() error {
	var hasFill bool
	for _, target := range r.m.Targets {
		if target.Fill != nil {
			hasFill = true
		}
	}
	if !hasFill {
		return nil
	}

	for _, row := range r.res {
		type line struct {
			begin int
			val   []float64
			end   int
		}

		var linearCache, previousCache = make(map[int]*line), make(map[int]interface{})
		var linearContext = make([][]float64, len(row.Columns))

		for _, value := range row.Values {
			for idx, targetIdx := 0, 0; idx < len(value); idx++ {
				if row.Columns[idx] == "time" {
					continue
				}
				if !isFillLinear(r.m.Targets[targetIdx].Fill) {
					continue
				}
				targetIdx++

				// i.e.,
				// column values like this: [0, 0, 12.3, 5.5, 0, 8.6, 0]
				// FIXME: xxx
				if value[idx] == nil {
					linearContext[idx] = append(linearContext[idx], 0)
				} else {
					switch v := value[idx].(type) {
					case json.Number:
						f, err := v.Float64()
						if err != nil {
							return fmt.Errorf("unreachable: result has invalid json.Number, %s", err)
						}
						linearContext[idx] = append(linearContext[idx], f)

					case float64:
						linearContext[idx] = append(linearContext[idx], v)
					}
				}
			}
		}

		for idx, lc := range linearContext {
			// i.e.,
			// val: [12.3, 5.5, 0, 8.6]
			begin, end, val := trimZero(lc)

			// i.e.,
			// val: [12.3, 5.5, XX, 8.6]
			fillna := linearInterpolateFillna(val)

			linearCache[idx] = &line{begin: begin, end: end, val: fillna}
		}

		for vidx, value := range row.Values {
			for cidx, targetIdx := 0, 0; cidx < len(row.Columns); cidx++ {
				if row.Columns[cidx] == "time" {
					continue
				}

				if value[cidx] != nil && isFillPrevious(r.m.Targets[targetIdx].Fill) {
					previousCache[cidx] = value[cidx]
					continue
				}

				if value[cidx] != nil {
					continue
				}

				f := r.m.Targets[targetIdx].Fill
				if f == nil {
					continue
				}
				targetIdx++

				var x interface{}

				switch f.FillType {
				case parser.FillNil:
					x = nil
				case parser.FillInt:
					x = f.Int
				case parser.FillFloat:
					x = f.Float
				case parser.FillStr:
					x = f.Str
				case parser.FillLinear:
					lc := linearCache[cidx]
					if lc == nil {
						continue
					}
					if lc.begin < vidx && vidx < lc.end {
						l.Debugf("fill lc.val[%d] to x, len(lc.val)=%d",
							vidx-lc.begin, len(lc.val))
						x = lc.val[vidx-lc.begin]
					}
				case parser.FillPrevious:
					x = previousCache[cidx]
				}

				value[cidx] = x
			}
		}
	}

	return nil
}

// FillConstant 填充定值
// ES 进行时间聚合查询时，返回的点数可能和聚合点数不匹配，在此进行填充
// Influxdb 数据不存在此问题
func (r *RewriteData) FillConstant() {
	if r.m.Namespace == NSMetric || r.m.Namespace == NSMetricAbbr {
		return
	}

	target := r.m.Targets
	// check：单个target，且存在fill
	if len(target) != 1 || target[0].Fill == nil {
		l.Debug("only accept 1 target and has fill()")
		return
	}

	var fillValue interface{}

	// fill类型为定值
	switch target[0].Fill.FillType {
	case parser.FillNil:
		fillValue = nil
	case parser.FillInt:
		fillValue = target[0].Fill.Int
	case parser.FillFloat:
		fillValue = target[0].Fill.Float
	case parser.FillStr:
		fillValue = target[0].Fill.Str
	default:
		l.Debug("invalid filltype, only accept NIL,INT,FLOAT,STRING")
		return
	}

	// 使用时间聚合
	timeRange := r.m.TimeRange
	if timeRange == nil || timeRange.TimeLength() == 0 {
		l.Debug("timeRange should not be nil")
		return
	}
	if timeRange.Resolution == nil || timeRange.Resolution.Duration == 0 {
		l.Debug("invalid aggregate time, interval cannot be zero")
		return
	}

	// 时间聚合间隔
	interval := timeRange.Resolution.Duration.Milliseconds()
	// 应该存在的point数量
	totalPointNum := timeRange.TimeLength().Milliseconds() / interval

	for idx, row := range r.res {
		// 当前point数量
		pointNum := int64(len(row.Values))

		// 数据集point数量符合，不需要fill
		if pointNum == totalPointNum {
			continue
		}

		// 临时变量，新的values
		var x = make([][]interface{}, 0, totalPointNum)

		// 空集情况
		if len(row.Values) == 0 {
			// start time时间戳转毫秒
			startTime := timeRange.Start.Time.Unix() * 1000

			for i := int64(0); i < totalPointNum; i++ {
				// 时间戳 + N * 聚合间隔
				x = append(x, []interface{}{startTime + i*interval, fillValue})
			}

			r.res[idx].Values = x
			continue
		}

		// 行协议结构首字段为time，类型为float64
		// 取第一个和最后一个point的时间戳
		sTime := int64(row.Values[0][0].(float64))
		eTime := int64(row.Values[pointNum-1][0].(float64))

		// i.e.,
		//  查询起始时间     sTime       eTime         查询结束时间
		//  |_______________,[v1, v2, v3, v4],____________|

		// 右侧point数量
		// -1 为减一纳秒，因为查询的时间区间为左闭右开
		// 否则会在特殊情况下，多一个point
		// 特殊情况为，只有一个point，且point的时间是查询起始时间
		rightPointNum := ((timeRange.End.Time.Unix()*1000 - 1) - eTime) / interval

		for i := int64(1); i <= rightPointNum; i++ {
			row.Values = append(row.Values, []interface{}{eTime + i*interval, fillValue})
		}

		// 左侧point数量 = 总量 - 右侧数量 - 已由数量
		leftPointNum := totalPointNum - rightPointNum - pointNum

		for i := leftPointNum; i > 0; i-- {
			x = append(x, []interface{}{sTime - i*interval, fillValue})
		}

		x = append(x, row.Values...)

		r.res[idx].Values = x
	}

	return
}

func RewriteResults(m *parser.DFQuery, results []ifxModel.Row) error {
	if m == nil || len(results) == 0 {
		return nil
	}

	const matchErr = "undefined behavior, invalid columns number: expected %d, got %d"

	var targetsLen = len(m.Targets)

	if !m.IsAllTargets() {
		for _, res := range results {
			if targetsLen != len(res.Columns)-1 {
				return fmt.Errorf(matchErr, targetsLen, len(res.Columns)-1)
			}

			for _, values := range res.Values {
				if targetsLen != len(values)-1 {
					return fmt.Errorf(matchErr, targetsLen, len(values)-1)
				}
			}
		}
	}

	var err error
	var r = RewriteData{m, results}

	if err = r.RenameColumns(); err != nil {
		return err
	}

	if err = r.FillResult(); err != nil {
		return err
	}

	r.FillConstant()

	return nil
}

func isFillLinear(f *parser.Fill) bool {
	if f != nil && f.FillType == parser.FillLinear {
		return true
	}
	return false
}

func isFillPrevious(f *parser.Fill) bool {
	if f != nil && f.FillType == parser.FillPrevious {
		return true
	}
	return false
}

func linearInterpolateFillna(vertor []float64) []float64 {
	if len(vertor) == 0 {
		return nil
	}

	var Y []float64
	var zero = make([]int, len(vertor))

	// i.e.,
	// [12.3, 0, 0, 5.5, 8.6]
	// zero[1] = 2
	area := 0
	for _, ver := range vertor {
		if ver != 0 {
			Y = append(Y, ver)
			area++
		} else {
			zero[area]++
		}
	}

	interpolate := piecewiselinear.Function{
		Y: Y,
		X: piecewiselinear.Span(1, float64(len(Y)), len(Y)),
	}

	var fillna []float64

	for idx, y := range Y {
		if n := zero[idx]; n > 0 {
			for i := 0; i < n; i++ {
				x := float64(idx) + float64(i+1)/float64(n+1)
				fillna = append(fillna, interpolate.At(x))
			}
		}
		fillna = append(fillna, y)
	}

	return fillna
}

func trimZero(val []float64) (left int, right int, res []float64) {
	if len(val) == 0 {
		return
	}

	var lstop, rstop bool

	for left, right = 0, len(val)-1; left < right; {
		if val[left] == 0 {
			left++
		} else {
			lstop = true
		}

		if val[right] == 0 {
			right--
		} else {
			rstop = true
		}

		if lstop && rstop {
			res = make([]float64, right+1-left)
			copy(res, val[left:right+1])
			break
		}
	}

	l.Debugf("input: %+#v, val: %+#v, b: %d, e: %d", val, res, left, right)
	return
}
