package parser

//
// NOTE: we move all Node/Expr's InfluxQL() implements here
//

import (
	"fmt"
	"strings"
	"time"
)

type showTransfer func(*Show) (string, error)

var (
	//  retention policy list
	influxdbRPList = map[string]string{
		"df_metering": "autogen",
		"df_bill":     "autogen",
		// FIXME:
		// show函数的from是作为函数参数存在，类型为 StringLiteral，翻译时两边会加引号
		// 此问题会在show函数重构之后解决
		`"df_metering"`: "autogen",
		`"df_bill"`:     "autogen",
	}

	influxdbShowFunctions = map[string]showTransfer{
		`show_measurement`: transferShowMeasurement,
		`show_field_key`:   transferShowFieldKey,
		`show_tag_key`:     transferShowTagKey,   //`SHOW TAG KEYS`
		`show_tag_value`:   transferShowTagValue, //`SHOW TAG VALUES`
		// TODO: add more
	}

	influxdbFunctions = map[string]string{
		"avg":                     "MEAN",
		"bottom":                  "BOTTOM",
		"count":                   "COUNT",
		"derivative":              "DERIVATIVE",
		"difference":              "DIFFERENCE",
		"distinct":                "DISTINCT",
		"first":                   "FIRST",
		"last":                    "LAST",
		"log":                     "LOG",
		"max":                     "MAX",
		"min":                     "MIN",
		"moving_average":          "MOVING_AVERAGE",
		"non_negative_derivative": "NON_NEGATIVE_DERIVATIVE",
		"percentile":              "PERCENTILE",
		"sum":                     "SUM",
		"top":                     "TOP",

		// combination functions
		"count_distinct": "count_distinct",

		// // influxdb aggregation functions
		// "integral": "INTEGRAL",
		// "median":   "MEDIAN",
		// "mode":     "MODE",
		// "spread":   "SPREAD",
		// "stddev":   "STDDEV",

		// // influxdb selector functions
		// "percentile": "PERCENTILE",
		// "sample":     "SAMPLE",

		// // influxdb transformation funcitons
		// "abs":     "ABS",
		// "acos":    "ACOS",
		// "asin":    "ASIN",
		// "atan":    "ATAN",
		// "atan2":   "ATAN2",
		// "ceil":    "CEIL",
		// "cos":     "COS",
		// "csum":    "CUMULATIVE_SUM",
		// "elapsed": "ELAPSED",
		// "exp":     "EXP",
		// "floor":   "FLOOR",

		// // not support: see: https://docs.influxdata.com/influxdb/v1.7/query_language/functions/#histogra
		// "Histogram": "HISTOGRAM",

		// "ln":      "LN",
		// "log2":    "LOG2",
		// "log10":   "LOG10",
		// "nndiff":  "NON_NEGATIVE_DIFFERENCE",
		// "pow":     "POW",
		// "round":   "ROUND",
		// "sin":     "SIN",
		// "sqrt":    "SQRT",
		// "tan":     "TAN",
		// "predict": "HOLT_WINTERS",
		// "cmo":     "CHANDE_MOMENTUM_OSCILLATOR",
		// "ema":     "EXPONENTIAL_MOVING_AVERAGE",
		// "dema":    "DOUBLE_EXPONENTIAL_MOVING_AVERAGE",
		// "ker":     "KAUFMANS_EFFICIENCY_RATIO",
		// "kama":    "KAUFMANS_ADAPTIVE_MOVING_AVERAGE",
		// "tema":    "TRIPLE_EXPONENTIAL_MOVING_AVERAGE",
		// "ted":     "TRIPLE_EXPONENTIAL_DERIVATIVE",
		// "rsi":     "RELATIVE_STRENGTH_INDEX",
	}
)

func (s *Show) InfluxQL() (string, error) {
	transfer, ok := influxdbShowFunctions[strings.ToLower(s.Func.Name)]
	if !ok {
		return "", fmt.Errorf("unknown show function %s", s.Func.Name)
	}
	s.Namespace = "metric"

	return transfer(s)
}

func (s *Show) transfer() (string, error) {
	var err error
	var arr []string

	if s.TimeRange != nil && s.TimeRange.Resolution != nil && s.TimeRange.Resolution.Duration != 0 {
		return "", fmt.Errorf("show clause not accept time range interval")
	}

	x, err := s.transferFilter()
	if err != nil {
		return "", err
	}
	if x != "" {
		arr = append(arr, x)
	}

	x, err = s.transferLimit()
	if err != nil {
		return "", err
	}
	if x != "" {
		arr = append(arr, x)
	}

	x, err = s.transferOffset()
	if err != nil {
		return "", err
	}
	if x != "" {
		arr = append(arr, x)
	}

	return strings.Join(arr, " "), nil
}

func (s *Show) transferFilter() (string, error) {
	timeFilters := s.TimeRange.TimeRangeFilter(time.UTC) // default use UTC time
	s.WhereCondition = append(s.WhereCondition, timeFilters...)

	if s.WhereCondition == nil {
		return "", nil
	}

	var arr []string
	for _, x := range s.WhereCondition {
		ql, err := x.InfluxQL()
		if err != nil {
			return "", nil
		}

		arr = append(arr, ql)
	}

	return "WHERE " + strings.Join(arr, " AND "), nil
}

func (s *Show) transferLimit() (string, error) {
	if s.Limit != nil {
		return s.Limit.InfluxQL()
	}
	return "", nil
}

func (s *Show) transferOffset() (string, error) {
	if s.Offset != nil {
		return s.Offset.InfluxQL()
	}
	return "", nil
}

func transferShowTagValue(s *Show) (string, error) {
	// i.e.,
	//  show_tag_values(from=["cpu"], keyin=["total"])

	parts := []string{"SHOW TAG VALUES"}

	var from, keyin []string

	for _, p := range s.Func.Param {
		arg, ok := p.(*FuncArg)
		if !ok {
			return "", fmt.Errorf("show_tag_values() only accept from=[...] and keyin=[...]")
		}

		switch arg.ArgName {
		case "from":
			argList, ok := arg.ArgVal.(FuncArgList)
			if !ok {
				log.Debugf("get arg val: %+#v", arg.ArgVal)
				return "", fmt.Errorf("from only accept list values: from=[...]")
			}

			list, err := argList.InfluxQLStringLiteralArray()
			if err != nil {
				return "", fmt.Errorf("from %s", err)
			}

			from = append(from, list...)

		case "keyin":
			argList, ok := arg.ArgVal.(FuncArgList)
			if !ok {
				log.Debugf("get arg val: %+#v", arg.ArgVal)
				return "", fmt.Errorf("keyin only accept list values: keyin=[...]")
			}

			list, err := argList.InfluxQLStringLiteral()
			if err != nil {
				return "", fmt.Errorf("keyin %s", err)
			}

			keyin = append(keyin, list)

		default:
			return "", fmt.Errorf("show_tag_values() only accept from=[...] and keyin=[...]")
		}
	}

	if len(from) > 0 {
		for idx, meas := range from {
			if rp, ok := influxdbRPList[meas]; ok {
				from[idx] = fmt.Sprintf(`"%s".%s`, rp, meas)
			}
		}
		parts = append(parts, "FROM "+strings.Join(from, ", "))
	}

	if len(keyin) == 0 {
		return "", fmt.Errorf("show_tag_values() keyin params is required")
	} else {
		parts = append(parts, fmt.Sprintf("WITH KEY IN (%s)", strings.Join(keyin, ", ")))
	}

	// has WHERE LIMIT_clause and OFFSET_clause
	clause, err := s.transfer()
	if err != nil {
		return "", err
	}
	if clause != "" {
		parts = append(parts, clause)
	}

	return strings.Join(parts, " "), nil
}

func transferShowTagKey(s *Show) (string, error) {
	// i.e.,
	//  show_tag_keys("cpu")
	//  show_tag_keys()

	parts := []string{"SHOW TAG KEYS"}

	var from []string

	for _, p := range s.Func.Param {
		arg, ok := p.(*FuncArg)
		if !ok {
			return "", fmt.Errorf("show_tag_keys() only accept from=[...]")
		}

		switch arg.ArgName {
		case "from":
			argList, ok := arg.ArgVal.(FuncArgList)
			if !ok {
				log.Debugf("get arg val: %+#v", arg.ArgVal)
				return "", fmt.Errorf("from only accept list values: from=[...]")
			}

			list, err := argList.InfluxQLStringLiteralArray()
			if err != nil {
				return "", fmt.Errorf("from %s", err)
			}

			from = append(from, list...)
		default:
			return "", fmt.Errorf("show_tag_keys() only accept from=[...]")
		}
	}

	if len(from) > 0 {
		for idx, meas := range from {
			if rp, ok := influxdbRPList[meas]; ok {
				from[idx] = fmt.Sprintf(`"%s".%s`, rp, meas)
			}
		}
		parts = append(parts, "FROM "+strings.Join(from, ", "))
	}

	// has WHERE LIMIT_clause and OFFSET_clause
	clause, err := s.transfer()
	if err != nil {
		return "", err
	}
	if clause != "" {
		parts = append(parts, clause)
	}

	return strings.Join(parts, " "), nil
}

func transferShowFieldKey(s *Show) (string, error) {
	// i.e.,
	//  show_field_keys("cpu")
	//  show_field_keys()

	if len(s.WhereCondition) > 0 ||
		s.TimeRange != nil ||
		s.Limit != nil ||
		s.Offset != nil {
		return "", fmt.Errorf("show_field_keys() not accept query clause")
	}

	parts := []string{"SHOW FIELD KEYS"}

	var from []string

	for _, p := range s.Func.Param {
		arg, ok := p.(*FuncArg)
		if !ok {
			return "", fmt.Errorf("show_field_keys() only accept from=[...]")
		}

		switch arg.ArgName {
		case "from":
			argList, ok := arg.ArgVal.(FuncArgList)
			if !ok {
				log.Debugf("get arg val: %+#v", arg.ArgVal)
				return "", fmt.Errorf("from only accept list values: from=[...]")
			}

			list, err := argList.InfluxQLStringLiteralArray()
			if err != nil {
				return "", fmt.Errorf("from %s", err)
			}

			from = append(from, list...)
		default:
			return "", fmt.Errorf("show_field_keys() only accept from=[...]")
		}
	}

	if len(from) > 0 {
		for idx, meas := range from {
			if rp, ok := influxdbRPList[meas]; ok {
				from[idx] = fmt.Sprintf(`"%s".%s`, rp, meas)
			}
		}
		parts = append(parts, "FROM "+strings.Join(from, ", "))
	}

	return strings.Join(parts, " "), nil
}

func transferShowMeasurement(s *Show) (string, error) {
	// i.e.,
	//  show_measurements(re("cpu*"))
	//  show_measurements()

	parts := []string{"SHOW MEASUREMENTS"}

	for _, p := range s.Func.Param {
		switch v := p.(type) {
		case *Regex:
			x, err := v.InfluxQL()
			if err != nil {
				return "", err
			}
			parts = append(parts, fmt.Sprintf("WITH MEASUREMENT =~ %s", x))
		default:
			return "", fmt.Errorf("show_measurements() only accept regex args")
		}
	}

	clause, err := s.transfer()
	if err != nil {
		return "", err
	}
	if clause != "" {
		parts = append(parts, clause)
	}

	return strings.Join(parts, " "), nil
}

/////////////////////////////////////////////  metric  /////////////////////////////////////////////

func (m *DFQuery) transferFrom() (string, error) {
	if m.Subquery != nil {
		ql, err := m.Subquery.InfluxQL()
		if err != nil {
			return "", err
		}

		return "FROM " + fmt.Sprintf("(%s)", ql), nil
	}

	arr := []string{}
	for _, x := range m.Names {
		if rp, ok := influxdbRPList[x]; ok {
			x = fmt.Sprintf(`"%s"."%s"`, rp, x)
		} else {
			x = fmt.Sprintf(`"%s"`, x)
		}
		arr = append(arr, x)
	}

	for _, x := range m.RegexNames {
		s, err := x.InfluxQL()
		if err != nil {
			return "", err
		}

		arr = append(arr, s)
	}

	return "FROM " + strings.Join(arr, ", "), nil
}

func (m *DFQuery) transferTarget() (string, error) {
	if len(m.Targets) == 0 {
		return fmt.Sprintf("*"), nil
	}

	// target semantic-checking
	{
		aggrFuncFound := false
		identifierFound := false

		var found func(Node)
		found = func(node Node) {
			switch v := node.(type) {
			case *FuncExpr:
				aggrFuncFound = true
			case *Identifier:
				identifierFound = true
			case *BinaryExpr:
				found(v.LHS)
				found(v.RHS)
			}
		}

		for _, t := range m.Targets {
			found(t.Col)
		}
		if aggrFuncFound && identifierFound {
			return "", fmt.Errorf("Metric mixing aggregate and non-aggregate queries is not supported")
		}
	}

	arr := []string{}
	for _, x := range m.Targets {
		ql, err := x.InfluxQL()
		if err != nil {
			return "", err
		}

		arr = append(arr, ql)
	}

	return fmt.Sprintf("%s", strings.Join(arr, ", ")), nil
}

func (m *DFQuery) transferFilter() (string, error) {
	var tzLoc *time.Location

	if m.TimeZone != nil {
		tzLoc, _ = time.LoadLocation(m.TimeZone.TimeZone)
	} else {
		tzLoc = time.UTC
	}

	timeFilters := m.TimeRange.TimeRangeFilter(tzLoc)

	m.WhereCondition = append(m.WhereCondition, timeFilters...)

	if len(m.WhereCondition) == 0 {
		return "", nil
	}

	var arr []string
	for _, x := range m.WhereCondition {
		ql, err := x.InfluxQL()
		if err != nil {
			return "", err
		}

		arr = append(arr, ql)
	}

	return "WHERE " + strings.Join(arr, " AND "), nil
}

func (m *DFQuery) transferOrderBy() (string, error) {
	if m.OrderBy == nil {
		return "", nil
	}

	// influxql only ORDER BY time supported at this time

	supportErr := fmt.Errorf("Metric only ORDER BY time supported")

	if len(m.OrderBy.List) != 1 {
		return "", supportErr
	}

	orderbyElem, ok := m.OrderBy.List[0].(*OrderByElem)
	if !ok {
		return "", fmt.Errorf("unreachable: orderby param is not orderbyElem")
	}

	switch orderbyElem.Column {
	case `'time'`, `time`:
	default:
		return "", supportErr
	}

	if orderbyElem.Opt == OrderDesc {
		return `ORDER BY "time" DESC`, nil
	}

	return `ORDER BY "time"`, nil
}

func (m *DFQuery) transferLimit() (string, error) {
	if m.Limit != nil {
		return m.Limit.InfluxQL()
	}
	return "", nil
}

func (m *DFQuery) transferSLimit() (string, error) {
	if m.SLimit != nil {
		return m.SLimit.InfluxQL()
	}
	return "", nil
}

func (m *DFQuery) transferOffset() (string, error) {
	if m.Offset != nil {
		return m.Offset.InfluxQL()
	}
	return "", nil
}

func (m *DFQuery) transferSOffset() (string, error) {
	if m.SOffset != nil {
		return m.SOffset.InfluxQL()
	}
	return "", nil
}

func (m *DFQuery) transferTimezone() (string, error) {
	if m.TimeZone != nil {
		return m.TimeZone.InfluxQL()
	}
	return "", nil
}

func (m *DFQuery) transferGroupBy() (string, error) {
	// InfluxQL, Advenced group by time syntax
	// https://docs.influxdata.com/influxdb/v1.8/query_language/explore-data/#advanced-group-by-time-syntax

	// group by time(5m), host fill("abc")

	if m.GroupBy == nil {
		// has TimeRange, add GROUP BY time(xxx)
		if m.TimeRange != nil && m.TimeRange.Resolution != nil {
			m.GroupBy = &GroupBy{}
		} else {
			return "", nil
		}
	}

	// log.Debugf("groupby: %+#v", m.GroupBy)

	var res []string

	if m.GroupBy.List != nil {
		groupbyList, err := m.GroupBy.List.InfluxQL()
		if err != nil {
			return "", err
		}

		res = append(res, groupbyList)
	}

	if m.TimeRange != nil {
		// check if target-list any aggr functions

		r := m.TimeRange.Resolution
		rs := m.TimeRange.ResolutionOffset

		// XXX: current influxdb do not support float duration like `3.5m`
		// `4.7h`, we convert all duration to ms

		// group by time(xxx) must prefix ahead other group by items

		timeAdd := false
		if r != nil && rs != 0 {
			timeStr := fmt.Sprintf("time(%dms, %dms)", r.Duration.Milliseconds(), rs.Milliseconds())
			res = append([]string{timeStr}, res...)
			timeAdd = true
		} else if r != nil {
			timeStr := fmt.Sprintf("time(%dms)", r.Duration.Milliseconds())
			res = append([]string{timeStr}, res...)
			timeAdd = true
		}

		if timeAdd {
			aggrFuncFound := false
			identifierFound := false

			var found func(Node)
			found = func(node Node) {
				switch v := node.(type) {
				case *FuncExpr:
					aggrFuncFound = true
				case *Identifier:
					identifierFound = true
				case *BinaryExpr:
					found(v.LHS)
					found(v.RHS)
				}
			}

			for _, t := range m.Targets {
				found(t.Col)
			}
			if aggrFuncFound && identifierFound {
				return "", fmt.Errorf("Metric mixing aggregate and non-aggregate queries is not supported")
			}
			if !aggrFuncFound {
				return "", fmt.Errorf("Metric GROUP BY interval require at least a aggregate function in target list")
			}
		}
	}

	if len(res) > 0 {
		return "GROUP BY " + strings.Join(res, ", "), nil
	}

	return "", nil
}

func (x *TimeRange) TimeRangeFilter(tzLoc *time.Location) []Node {
	// time-range -> where-clause
	if x == nil {
		return nil
	}

	var res []Node

	if x.Start != nil {
		res = append(res, &BinaryExpr{
			Op:  GTE,
			LHS: &Identifier{Name: "time"},
			RHS: &StringLiteral{Val: x.Start.Time.In(tzLoc).Format(DateTimeFormat)},
		})
	}

	if x.End != nil {
		res = append(res, &BinaryExpr{
			Op:  LT,
			LHS: &Identifier{Name: "time"},
			RHS: &StringLiteral{Val: x.End.Time.In(tzLoc).Format(DateTimeFormat)},
		})
	}

	return res
}

func (m *DFQuery) InfluxQL() (string, error) {
	var err error

	switch m.Namespace {
	case "metric", "M", "":
	default:
		return "", fmt.Errorf("invalid namespace `%s'", m.Namespace)
	}

	arr := []string{
		"SELECT",
	}

	x, err := m.transferTarget()
	if err != nil {
		return "", err
	}
	if x != "" {
		arr = append(arr, x)
	}

	x, err = m.transferFrom()
	if err != nil {
		return "", err
	}

	if x != "" {
		arr = append(arr, x)
	}

	x, err = m.transferFilter()
	if err != nil {
		return "", err
	}
	if x != "" {
		arr = append(arr, x)
	}

	x, err = m.transferGroupBy()
	if err != nil {
		return "", err
	}
	if x != "" {
		arr = append(arr, x)
	}

	x, err = m.transferOrderBy()
	if err != nil {
		return "", err
	}
	if x != "" {
		arr = append(arr, x)
	}

	x, err = m.transferLimit()
	if err != nil {
		return "", err
	}
	if x != "" {
		arr = append(arr, x)
	}

	x, err = m.transferOffset()
	if err != nil {
		return "", err
	}
	if x != "" {
		arr = append(arr, x)
	}

	x, err = m.transferSLimit()
	if err != nil {
		return "", err
	}
	if x != "" {
		arr = append(arr, x)
	}

	x, err = m.transferSOffset()
	if err != nil {
		return "", err
	}
	if x != "" {
		arr = append(arr, x)
	}

	x, err = m.transferTimezone()
	if err != nil {
		return "", err
	}
	if x != "" {
		arr = append(arr, x)
	}

	return strings.Join(arr, " "), nil
}

///////////////////////////////////////////// end of metric /////////////////////////////////////////////

func (x *Target) InfluxQL() (string, error) {
	s, err := x.Col.InfluxQL()
	if err != nil {
		return "", err
	}

	if x.Alias != "" {
		return fmt.Sprintf(`%s AS "%s"`, s, x.Alias), nil
	} else {
		return fmt.Sprintf(`%s`, s), nil
	}
}

func (x *Identifier) InfluxQL() (string, error) {
	return fmt.Sprintf(`"%s"`, strings.Replace(x.String(), `"`, `\"`, -1)), nil
}

func (x *FuncExpr) InfluxQL() (string, error) {
	op, ok := influxdbFunctions[strings.ToLower(x.Name)]
	if !ok {
		return "", fmt.Errorf("this funciton not supported, check the funciton name")
	}

	param := func() (string, error) {
		var arr []string

		for _, p := range x.Param {
			ql, err := p.InfluxQL()
			if err != nil {
				return "", err
			}
			arr = append(arr, ql)
		}
		if len(arr) == 0 {
			arr = append(arr, "*")
		}

		return strings.Join(arr, ", "), nil
	}

	var paramStr string
	var err error

	switch op {
	case "count_distinct":
		paramStr, err = param()
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("COUNT(DISTINCT(%s))", paramStr), nil

	default:
		paramStr, err = param()
		if err != nil {
			return "", err
		}
	}

	return fmt.Sprintf("%s(%s)", op, paramStr), nil
}

func (x *BinaryExpr) InfluxQL() (string, error) {
	rhsIsRegex := false
	op := x.Op.String()

	switch x.RHS.(type) {
	case *Regex:
		rhsIsRegex = true
	}

	if rhsIsRegex {
		switch x.Op {
		case EQ:
			op = "=~"
		case NEQ:
			op = "!~"
		default: // PASS
			return "", fmt.Errorf("should never been here")
		}
	}

	l, err := x.LHS.InfluxQL()
	if err != nil {
		return "", err
	}

	var r string

	switch v := x.RHS.(type) {
	case NodeList:
		var res []string
		for _, node := range v {
			s, err := (&BinaryExpr{EQ, x.LHS, node, true}).InfluxQL()
			if err != nil {
				return "", err
			}
			res = append(res, s)
		}
		return "(" + strings.Join(res, " OR ") + ")", nil

	case *StringLiteral:
		// influxql: WHERE "host" = 'desktop'
		// right value use single quote
		r, err = v.InfluxQL2()
		if err != nil {
			return "", err
		}

	default:
		r, err = x.RHS.InfluxQL()
		if err != nil {
			return "", err
		}
	}

	return fmt.Sprintf(`%s %s %s`, l, op, r), nil
}

func (x *ParenExpr) InfluxQL() (string, error) {
	q, err := x.Param.InfluxQL()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("(%s)", q), nil
}

func (x *StringLiteral) InfluxQL() (string, error) {
	return fmt.Sprintf(`"%s"`, x.Val), nil
}

func (x *StringLiteral) InfluxQL2() (string, error) {
	return fmt.Sprintf("'%s'", x.Val), nil
}

func (x *NumberLiteral) InfluxQL() (string, error) {
	return x.String(), nil
}

func (x *FuncArg) InfluxQL() (string, error) {
	return "", fmt.Errorf("not impl")
}

func (x *Fill) InfluxQL() (string, error) {
	return x.StringInfluxql(), nil
}

func (x *Regex) InfluxQL() (string, error) {
	return fmt.Sprintf(`/%s/`, x.Regex), nil
}

func (x NodeList) InfluxQL() (string, error) {
	arr := []string{}
	for _, n := range x {
		ql, err := n.InfluxQL()
		if err != nil {
			return "", err
		}

		arr = append(arr, ql)
	}

	return strings.Join(arr, ", "), nil
}

func (x *TimeRange) InfluxQL() (string, error) {
	return "", nil
}

func (x *TimeZone) InfluxQL() (string, error) {
	return fmt.Sprintf("tz('%s')", x.TimeZone), nil
}

func (x *OrderBy) InfluxQL() (string, error) {
	return "", fmt.Errorf("nil no support")
}

func (x *OrderByElem) InfluxQL() (string, error) {
	return "", fmt.Errorf("nil no support")
}

func (x *TimeExpr) InfluxQL() (string, error) {
	return "", fmt.Errorf("not impl")
}

func (x *Limit) InfluxQL() (string, error) {
	return x.String(), nil
}

func (x *SLimit) InfluxQL() (string, error) {
	return x.String(), nil
}

func (x *Offset) InfluxQL() (string, error) {
	return x.String(), nil
}

func (x *SOffset) InfluxQL() (string, error) {
	return x.String(), nil
}

func (x *NilLiteral) InfluxQL() (string, error) {
	return "", fmt.Errorf("nil no support")
}

func (x *BoolLiteral) InfluxQL() (string, error) {
	return fmt.Sprintf("%v", x.Val), nil
}

func (x Stmts) InfluxQL() (string, error) {
	arr := []string{}
	for _, s := range x {
		ql, err := s.InfluxQL()
		if err != nil {
			return "", err
		}

		arr = append(arr, ql)
	}

	return strings.Join(arr, "; "), nil
}

func (x *GroupBy) InfluxQL() (string, error) {
	return "", fmt.Errorf("should no been here")
}

func (x FuncArgList) InfluxQL() (string, error) {
	arr := []string{}
	for _, arg := range x {
		ql, err := arg.InfluxQL()
		if err != nil {
			return "", err
		}

		arr = append(arr, ql)
	}

	return strings.Join(arr, ", "), nil
}

func (x FuncArgList) InfluxQLStringLiteral() (string, error) {
	arr, err := x.InfluxQLStringLiteralArray()
	if err != nil {
		return "", err
	}
	return strings.Join(arr, ", "), nil
}

func (x FuncArgList) InfluxQLStringLiteralArray() ([]string, error) {
	arr := []string{}
	for _, arg := range x {
		n, ok := arg.(*StringLiteral)
		if !ok {
			return nil, fmt.Errorf(`only accept string list values, got %v`, n)
		}
		ql, err := n.InfluxQL()
		if err != nil {
			return nil, err
		}

		arr = append(arr, ql)
	}

	return arr, nil
}

func (x *Star) InfluxQL() (string, error) {
	return "*", nil
}

func (x *AttrExpr) InfluxQL() (string, error) {
	return "", fmt.Errorf("no impl")
}

func (x *CascadeFunctions) InfluxQL() (string, error) {
	return "", fmt.Errorf("no impl")
}

func (x *TimeResolution) InfluxQL() (string, error) {
	return "", fmt.Errorf("not impl")
}

func (x *Lambda) InfluxQL() (string, error) {
	return "", fmt.Errorf("not impl")
}

func (x *StaticCast) InfluxQL() (string, error) {
	v, err := x.Val.InfluxQL()
	if err != nil {
		return "", err
	}

	if x.IsInt {
		return fmt.Sprintf("%s::integer", v), nil
	}

	if x.IsFloat {
		return fmt.Sprintf("%s::float", v), nil
	}

	return "", fmt.Errorf("unreachable, invalid cast operations")
}
