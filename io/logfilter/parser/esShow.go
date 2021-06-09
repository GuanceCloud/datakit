package parser

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

// ShowFunc show函数结构
type ShowFunc struct {
	funcName  string // 函数名称
	funcType  string // 函数类型，例如: class, tag等
	className string // 分类字段名称
	classRes  string // 分类查询返回值
	timeField string // time字段
}

// 全局变量，只读
var (
	// show 指标集合
	objectShowClassFunc = ShowFunc{
		funcName:  "show_object_class",
		funcType:  "class",
		className: OCLASS,
	}

	objectShowClassFunc1 = ShowFunc{
		funcName:  "show_object_source",
		funcType:  "class",
		className: OCLASS,
	}

	objectShowFieldsFunc = ShowFunc{
		funcName:  "show_object_field",
		funcType:  "fields",
		className: OCLASS,
		timeField: OTIME,
	}

	loggingShowClassFunc = ShowFunc{
		funcName:  "show_logging_source",
		funcType:  "class",
		className: LCLASS,
	}

	loggingShowFieldsFunc = ShowFunc{
		funcName:  "show_logging_field",
		funcType:  "fields",
		className: LCLASS,
		timeField: LTIME,
	}

	eventShowClassFunc = ShowFunc{
		funcName:  "show_event_source",
		funcType:  "class",
		className: ECLASS,
	}

	eventShowFieldsFunc = ShowFunc{
		funcName:  "show_event_field",
		funcType:  "fields",
		className: ECLASS,
		timeField: ETIME,
	}

	tracingShowClassFunc = ShowFunc{
		funcName:  "show_tracing_service",
		funcType:  "class",
		className: TCLASS,
	}

	tracingShowClassFunc1 = ShowFunc{
		funcName:  "show_tracing_source",
		funcType:  "class",
		className: TCLASS,
	}

	tracingShowFieldsFunc = ShowFunc{
		funcName:  "show_tracing_field",
		funcType:  "fields",
		className: TCLASS,
		timeField: TTIME,
	}

	rumShowClassFunc = ShowFunc{
		funcName:  "show_rum_type",
		funcType:  "class",
		className: RCLASS,
	}

	rumShowClassFunc1 = ShowFunc{
		funcName:  "show_rum_source",
		funcType:  "class",
		className: RCLASS,
	}

	rumShowFieldsFunc = ShowFunc{
		funcName:  "show_rum_field",
		funcType:  "fields",
		className: RCLASS,
		timeField: RTIME,
	}

	securityShowClassFunc = ShowFunc{
		funcName:  "show_security_category",
		funcType:  "class",
		className: SCLASS,
	}
	securityShowClassFunc1 = ShowFunc{
		funcName:  "show_security_source",
		funcType:  "class",
		className: SCLASS,
	}

	securityShowFieldsFunc = ShowFunc{
		funcName:  "show_security_field",
		funcType:  "fields",
		className: SCLASS,
		timeField: STIME,
	}

	objectFuncs = []*ShowFunc{
		&objectShowClassFunc,
		&objectShowClassFunc1,
		&objectShowFieldsFunc,
	}

	loggingFuncs = []*ShowFunc{
		&loggingShowClassFunc,
		&loggingShowFieldsFunc,
	}

	eventFuncs = []*ShowFunc{
		&eventShowClassFunc,
		&eventShowFieldsFunc,
	}

	tracingFuncs = []*ShowFunc{
		&tracingShowClassFunc,
		&tracingShowClassFunc1,
		&tracingShowFieldsFunc,
	}

	rumFuncs = []*ShowFunc{
		&rumShowClassFunc,
		&rumShowClassFunc1,
		&rumShowFieldsFunc,
	}

	securityFuncs = []*ShowFunc{
		&securityShowClassFunc,
		&securityShowClassFunc1,
		&securityShowFieldsFunc,
	}

	// ShowFuncs show funcs
	ShowFuncs = map[string][]*ShowFunc{
		"object":   objectFuncs,
		"logging":  loggingFuncs,
		"event":    eventFuncs,
		"tracing":  tracingFuncs,
		"rum":      rumFuncs,
		"security": securityFuncs,
	}
)

// checkAstFunc, check ast attribute
func (s *Show) checkAstFunc() error {
	// show 只有func属性
	// if (s.WhereCondition != nil) || (s.TimeRange != nil) || (s.Limit != nil) || (s.Offset != nil) {
	// 	return fmt.Errorf("can not translate show")
	// }
	if s.Func == nil {
		return fmt.Errorf("show need func")
	}
	return nil
}

// checkFuncExist, 检查函数是否存在
func (s *Show) checkFuncExist() (*ShowFunc, error) {

	var funcPtr *ShowFunc
	funcName := s.Func.Name
	if fns, ok := ShowFuncs[s.Namespace]; ok {
		for _, v := range fns {
			if funcName == v.funcName {
				funcPtr = v
				break
			}
		}
		if funcPtr == nil {
			return nil, fmt.Errorf("no such show func")
		}
	} else {
		return nil, fmt.Errorf("unsupport show namespace %s", s.Namespace)
	}
	return funcPtr, nil
}

// checkShowFields, 如果是show fields函数，可以指定一个具体类型值
func (s *Show) checkShowFields(funcPtr *ShowFunc) error {
	if funcPtr.funcType == "fields" {
		len := len(s.Func.Param)
		if len > 1 {
			return fmt.Errorf("show fileds func, only support one class name") // 只能有一个class值
		}
		if len == 1 {
			if IsStringParam(s.Func.Param[0]) == false {
				return fmt.Errorf("show fields func, class name should be string") // 只支持字符串
			}
		}

	}
	return nil
}

// check_valid 检查
func (s *Show) checkValid() (*ShowFunc, error) {
	var err error
	var funcPtr *ShowFunc
	// show 只有func属性
	err = s.checkAstFunc()
	if err != nil {
		return nil, err
	}
	// 函数是否存在
	funcPtr, err = s.checkFuncExist()
	if err != nil {
		return nil, err
	}
	// 如果是show fields函数，可以指定一个具体类型值
	err = s.checkShowFields(funcPtr)
	if err != nil {
		return nil, err
	}
	return funcPtr, nil
}

func (f *ShowFunc) transport() (ESMetric, error) {
	var em ESMetric
	var err error
	em, err = f.transportClass()
	if err != nil {
		return em, err
	}
	return em, nil
}

// 指标集，聚合查询
func (f *ShowFunc) transportClass() (ESMetric, error) {
	var em ESMetric
	iinner := map[string]string{
		"field": f.className,
		"size":  strconv.Itoa(MaxLimit),
	}
	inner := SIMAP{"terms": iinner}
	outer := SIMAP{"aggs1": inner}
	em.Size = ZeroLimit
	em.Aggs = outer
	return em, nil
}

// ESQL show
func (s *Show) ESQL() (interface{}, error) {
	// check vaild
	var (
		err     error
		funcPtr *ShowFunc
		em      ESMetric
		sRes    string
	)
	funcPtr, err = s.checkValid()
	if err != nil {
		return "", err
	}
	// 将解析信息添加到AST上，用于结果解析
	estResPtr := &ESTRes{
		Alias:      map[string]string{},
		ClassNames: "",
		SortFields: []string{},
		Show:       true,
	}
	helper := &Helper{
		ESTResPtr: estResPtr,
	}
	// translate 函数
	switch funcPtr.funcType {
	case "class":
		em, err = funcPtr.transport()
		if err != nil {
			return "", err
		}
		bRes, err := json.Marshal(em)

		if err != nil {
			return "", fmt.Errorf("json marshal error, %s", err)
		}
		sRes = string(bRes)
	case "fields": // show fields函数
		// sRes = funcPtr.funcName
		sRes = ""
		if len(s.Func.Param) > 0 {
			classValue := GetStringParam(s.Func.Param[0])
			sRes = fmt.Sprintf(`{"term":{"%s":{"value":"%s"}}},`, funcPtr.className, classValue)
		}
		// 添加时间范围
		timeRangeStr, start, end, err := s.getTimeRange()
		if err != nil {
			return "", fmt.Errorf("invalid body, %s", err.Error())
		}
		sRes = sRes + timeRangeStr
		helper.ESTResPtr.TimeField = funcPtr.timeField
		helper.ESTResPtr.ShowFields = true
		helper.ESTResPtr.StartTime = start
		helper.ESTResPtr.EndTime = end
	}
	s.Helper = helper

	return sRes, nil
}

// getTimeRange 时间范围过滤
func (s *Show) getTimeRange() (string, int64, int64, error) {
	var (
		res          string
		rStart, rEnd int64
	)

	if s.TimeRange == nil || s.TimeRange.Start == nil || s.TimeRange.End == nil {
		return res, rStart, rEnd, nil
	}
	start := s.TimeRange.Start
	rStart = int64(start.Time.UnixNano() / int64(time.Millisecond))
	end := s.TimeRange.Start
	rEnd = int64(end.Time.UnixNano() / int64(time.Millisecond))
	res = fmt.Sprintf(`{"range":{"date":{"gte":%d,"lte":%d}}},`, rStart, rEnd)
	return "", rStart, rEnd, nil
}
