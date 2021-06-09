package parser

import (
	"fmt"
	"time"
)

type ExtraParam struct {
	// target 数量限制
	TargetsNum int

	// 追加 where conditions
	Condition *BinaryExpr

	// 最大查询时间
	MaxDuration time.Duration

	// 查询起始时间和结束时间
	// 此处如果不用指针，将无法简易判断是否为有效值
	StartTime, EndTime *time.Time

	// 聚合点数
	GroupByPointNum int64

	// 以替换值为第一优先，dafault为第二优先，max最大限制为第三优先
	// 例如:
	// 1> 参数 Limit 为 1000，且 MaxLimit 为 100 时，1000 有效
	// 2> DQL 没有原生 Limit，且 参数 DafaultLimit 值有效（正整数），以此值添加 DQL Limit
	// 2.5> 纯函数 DQL 不添加默认值，比如 show_XXX() 语句
	// 3> 参数 Limit 为空值 0，且 MaxLimit 为 100 时，限制 DQL 原生 Limit <不可大于!> 100

	Limit, MaxLimit, DefaultLimit    int64
	SLimit, MaxSLimit, DefaultSLimit int64

	Offset, SOffset int64

	// 排序，使用map的array是为了保证顺序
	OrderBy []map[string]OrderType

	// search after,深度分页使用,可以保证一次查询没有重复数据
	// "search_after": [
	//   1620873550407,
	//   "L_c2e92jqc8kgenkdf02v0"
	// ],
	SearchAfter []interface{}
	Highlight   bool //是否高亮查询结果
}

func (p *parser) addDefaultLimit() {
	if p.ExtraParam == nil || p.ExtraParam.DefaultLimit <= 0 {
		return
	}

	switch v := p.parseResult.(type) {
	case Stmts:
		for _, stmt := range v {
			query, ok := stmt.(*DFQuery)
			if !ok {
				continue
			}

			if query.Limit == nil {
				query.Limit = &Limit{Limit: p.ExtraParam.DefaultLimit}
			}
		}
	}
}

func (p *parser) addDefaultSLimit() {
	if p.ExtraParam == nil || p.ExtraParam.DefaultSLimit <= 0 {
		return
	}

	switch v := p.parseResult.(type) {
	case Stmts:
		for _, stmt := range v {
			query, ok := stmt.(*DFQuery)
			if !ok {
				continue
			}
			if query.SLimit == nil {
				query.SLimit = &SLimit{SLimit: p.ExtraParam.DefaultSLimit}
			}
		}
	}
}

// 添加search_after信息
func (p *parser) addSearchAfter() {
	if p.ExtraParam == nil || p.ExtraParam.SearchAfter == nil {
		return
	}

	switch v := p.parseResult.(type) {
	case Stmts:
		for _, stmt := range v {
			query, ok := stmt.(*DFQuery)
			if !ok {
				continue
			}
			query.SearchAfter = &SearchAfter{
				Vals: p.ExtraParam.SearchAfter,
			}
		}
	}
}

// 添加highlight信息
func (p *parser) addHighlight() {
	if p.ExtraParam == nil {
		return
	}

	switch v := p.parseResult.(type) {
	case Stmts:
		for _, stmt := range v {
			query, ok := stmt.(*DFQuery)
			if !ok {
				continue
			}
			query.Highlight = p.ExtraParam.Highlight
		}
	}
}

func (p *parser) newOrderBy(list NodeList) *OrderBy {
	if p.ExtraParam != nil && len(p.ExtraParam.OrderBy) != 0 {
		// 接收到有效的 OrderBy，原 OrderBy 将被弃用
		var orderbyList NodeList
		for _, elem := range p.ExtraParam.OrderBy {
			for column, opt := range elem {
				orderbyList = append(orderbyList, &OrderByElem{Column: column, Opt: opt})
			}
		}
		return &OrderBy{List: orderbyList}
	}

	if len(list) == 0 {
		return nil
	}

	return &OrderBy{List: list}
}

func (p *parser) newTargets(nodes []Node) []*Target {
	// 如果 target list 为空，一般需要添加默认值 '*'，此步骤由翻译层来做
	if len(nodes) == 0 {
		return nil
	}

	var targets []*Target
	for _, node := range nodes {
		targets = append(targets, node.(*Target))
	}

	// targets 数量匹配
	matchNum := func(num int) bool {
		// 如果 target 个数为 0 时，默认是 all
		if len(targets) == 0 || targets[0].Col.String() == "*" {
			return num == 0
		}
		return len(targets) == num
	}

	// FIXME:
	// target 正常字段和 all 不能共存，需要做语义检查
	// 暂时只支持数量的绝对匹配，不支持大于或小于
	if p.ExtraParam != nil && p.ExtraParam.TargetsNum != 0 {
		if p.ExtraParam.TargetsNum == 1 {
			if matchNum(0) {
				p.addParseErrf(nil, "only 1 field allowed, cannot select all fields")
				return nil
			}
		}

		if !matchNum(p.ExtraParam.TargetsNum) {
			p.addParseErrf(nil, "only %d field allowed, got: %d", p.ExtraParam.TargetsNum, len(targets))
			return nil
		}
	}

	return targets
}

func (p *parser) newWhereConditions(conditions []Node) []Node {
	if p.ExtraParam != nil && p.ExtraParam.Condition != nil {
		// 此处以 ParenExpr （括号表达式）将原 binaryExpr 封装，避免混乱
		return append(conditions, &ParenExpr{Param: p.ExtraParam.Condition})
	}
	return conditions
}

func (p *parser) newLimit(n *NumberLiteral) *Limit {
	// FIXME:
	// 如果 DQL Limit 比 ExtraParam.Limit 要小，是否还有替换该值？
	if p.ExtraParam != nil && p.ExtraParam.Limit > 0 {
		return &Limit{p.ExtraParam.Limit}
	}

	if n == nil {
		return nil
	}

	if !n.IsPositiveInteger() {
		p.addParseErrf(p.yyParser.lval.item.PositionRange(),
			"limit must be positive integer, got: %s", n)
		return nil
	}

	if p.ExtraParam != nil && p.ExtraParam.MaxLimit < n.Int {
		p.addParseErrf(p.yyParser.lval.item.PositionRange(),
			"limit should less than %d, got %d", p.ExtraParam.MaxLimit, n.Int)
		return nil
	}

	return &Limit{n.Int}
}

func (p *parser) newSLimit(n *NumberLiteral) *SLimit {
	if p.ExtraParam != nil && p.ExtraParam.SLimit > 0 {
		return &SLimit{p.ExtraParam.SLimit}
	}

	if n == nil {
		return nil
	}

	if !n.IsPositiveInteger() {
		p.addParseErrf(p.yyParser.lval.item.PositionRange(),
			"slimit must be positive integer, got: %s", n)
		return nil
	}

	if p.ExtraParam != nil && p.ExtraParam.MaxSLimit < n.Int {
		p.addParseErrf(p.yyParser.lval.item.PositionRange(),
			"slimit should less than %d, got %d", p.ExtraParam.MaxSLimit, n.Int)
		return nil
	}

	return &SLimit{n.Int}
}

func (p *parser) newOffset(n *NumberLiteral) *Offset {
	if p.ExtraParam != nil && p.ExtraParam.Offset > 0 {
		return &Offset{p.ExtraParam.Offset}
	}

	if n == nil {
		return nil
	}

	if !n.IsPositiveInteger() {
		p.addParseErrf(p.yyParser.lval.item.PositionRange(),
			"offset must be positive integer, got: %s", n)
		return nil
	}

	return &Offset{n.Int}
}

func (p *parser) newSOffset(n *NumberLiteral) *SOffset {
	if p.ExtraParam != nil && p.ExtraParam.SOffset > 0 {
		return &SOffset{p.ExtraParam.SOffset}
	}

	if n == nil {
		return nil
	}

	if !n.IsPositiveInteger() {
		p.addParseErrf(p.yyParser.lval.item.PositionRange(),
			"soffset must be positive integer, got: %s", n)
		return nil
	}

	return &SOffset{n.Int}
}

func (p *parser) newTimeRangeOpt(start, end, resolution Node, offset time.Duration) *TimeRange {
	var startTime, endTime *TimeExpr
	var resolutionTime *TimeResolution

	if start != nil {
		startTime = start.(*TimeExpr)
	}

	if end != nil {
		endTime = end.(*TimeExpr)
	}

	if resolution != nil {
		resolutionTime = resolution.(*TimeResolution)
	}

	if p.ExtraParam != nil {
		if p.ExtraParam.StartTime != nil {
			startTime = &TimeExpr{Time: *p.ExtraParam.StartTime}
		}
		if p.ExtraParam.EndTime != nil {
			endTime = &TimeExpr{Time: *p.ExtraParam.EndTime}
		}
		if p.ExtraParam.GroupByPointNum > 0 {
			resolutionTime = &TimeResolution{
				PointNum: &NumberLiteral{IsInt: true, Int: p.ExtraParam.GroupByPointNum},
			}
		}
	}

	return p.newTimeRange(startTime, endTime, resolutionTime, offset)
}

func (p *parser) newTimeRange(start, end *TimeExpr, r *TimeResolution, offset time.Duration) *TimeRange {
	if start == nil && end == nil && r == nil {
		return nil
	}

	if start == nil && end != nil {
		p.addParseErr(nil, fmt.Errorf("invalid time range, missing start time"))
		return nil
	}

	t := &TimeRange{
		Start:            start,
		End:              end,
		Resolution:       r,
		ResolutionOffset: offset,
	}

	if p.ExtraParam != nil {
		if p.ExtraParam.MaxDuration != 0 && t.TimeLength() > p.ExtraParam.MaxDuration {
			p.addParseErr(nil, fmt.Errorf("time range should less than %s", p.ExtraParam.MaxDuration))
			return nil
		}
	}

	func() {
		if t.Resolution == nil {
			return
		}

		if t.Resolution.Duration != 0 {
			return
		}

		if !t.Resolution.PointNum.IsPositiveInteger() {
			p.addParseErr(nil, fmt.Errorf("auto() param only accept positive integer"))
			return
		}
		// FIXME:
		// TimeLength() 会自动处理 end time 不存在的情况，但是因为网络传输和数据查询的延迟，可能会存在误差
		// 如果在此处，手动修改 AST，添加 end time 为 time.Now()，就不会出现上述问题
		// 不建议这样做，尽量不改动原有 AST 结构（宏替换除外）
		//
		// 问题示例：
		// start time 为 13:00:10，如果 end time 不存在，会以当前时间来计算，例如 13:00:30，得到的 time length 是 20 秒
		// pointNum 是为 20，DQL 计算 interval 是 1s
		// 但是当查询语句发送到数据库时，在数据库看来，当前时间可能是 13:00:33，有 3 秒延迟在传输和计算上
		// 根据 interval 为 1s，此时数据库可能会返回 23 个 point

		t.Resolution.Duration = time.Duration(int64(t.TimeLength()) / t.Resolution.PointNum.Int)

	}()

	return t
}

func (p *parser) newTimeResolution(n *NumberLiteral, auto bool) *TimeResolution {
	if n != nil {
		return &TimeResolution{PointNum: n}
	}

	const defaultPointNum = 360

	return &TimeResolution{
		Auto:     auto,
		PointNum: &NumberLiteral{IsInt: true, Int: defaultPointNum},
	}
}
