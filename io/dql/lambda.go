package dql

import (
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/kodo/dql/parser"
)

type lq struct {
	wsid      string
	lambdaAST *parser.Lambda
	qw        *queryWorker
	start     time.Time
	explain   bool
}

func (qw *queryWorker) lambdaQuery(wsid string, q *ASTResult, explain bool) (*QueryResult, error) {
	lq := &lq{
		wsid:      wsid,
		lambdaAST: q.AST.(*parser.Lambda),
		qw:        qw,
		start:     time.Now(),
		explain:   explain,
	}

	if err := lq.checkLambdaAST(); err != nil {
		return nil, err
	}

	return lq.runLambdaAST()
}

func (lq *lq) checkLambdaAST() error {
	switch lq.lambdaAST.Opt {
	case parser.LambdaFilter:
		if len(lq.lambdaAST.Right) != 1 {
			return fmt.Errorf("filter right value only accept 1 DQL")
		}
		if !(lq.lambdaAST.Right[0].Namespace == NSObject || lq.lambdaAST.Right[0].Namespace == NSObjectAbbr) {
			return fmt.Errorf("filter right only accept object namesapce")
		}

	case parser.LambdaLink:
		if !(lq.lambdaAST.Left.Namespace == NSObject || lq.lambdaAST.Left.Namespace == NSObjectAbbr) {
			return fmt.Errorf("link left only accept object namesapce")
		}
		if len(lq.lambdaAST.WhereCondition) != 1 {
			return fmt.Errorf("link with condition only accept 1 param")
		}
	default:
		return fmt.Errorf("unknown lambdaQuery type")
	}

	return lq.checkWithExpr()
}

func (lq *lq) checkWithExpr() error {
	for idx, expr := range lq.lambdaAST.WhereCondition {
		binaryExpr, ok := expr.(*parser.BinaryExpr)
		if !ok {
			return fmt.Errorf("with only accept binaryExpr param")
		}

		_, leftOk := binaryExpr.LHS.(*parser.Identifier)
		_, rightOk := binaryExpr.RHS.(*parser.Identifier)

		if !(leftOk && rightOk) {
			return fmt.Errorf("with conditions only accept identifier param")
		}

		switch binaryExpr.Op {
		case parser.EQ, parser.NEQ, parser.LTE, parser.LT, parser.GTE, parser.GT:
			// = , != , <= , < , >= , >
			// pass
		default:
			return fmt.Errorf("found unexpected operator of WITH[%d]: %s, only accept =,!=,<=,<,>=,>",
				idx, binaryExpr.Op.String())
		}
	}

	return nil
}

func (lq *lq) runLambdaAST() (*QueryResult, error) {
	switch lq.lambdaAST.Opt {
	case parser.LambdaFilter:
		return lq.runFilterQuery()
	case parser.LambdaLink:
		return lq.runLinkQuery()
	}
	return nil, fmt.Errorf("unknown lambdaQuery type")
}

func (lq *lq) runQuery(m *parser.DFQuery) (*QueryResult, error) {
	ast, err := newASTResult(m)
	if err != nil {
		return nil, err
	}

	return lq.qw.runQuery(lq.wsid, ast, lq.explain)
}

func (lq *lq) runFilterQuery() (*QueryResult, error) {
	// filter right only accept 1 DQL
	res, err := lq.runQuery(lq.lambdaAST.Right[0])
	if err != nil {
		return nil, err
	}

	var conditions []*parser.BinaryExpr
	for _, expr := range lq.lambdaAST.WhereCondition {
		binExpr := expr.(*parser.BinaryExpr)

		condition := lq.newCondition(binExpr.LHS.String(), binExpr.Op, lq.findValues(res, binExpr.RHS.String()).unique())
		if condition == nil {
			return nil, fmt.Errorf("result not found column of %s", binExpr.RHS.String())
		}

		conditions = append(conditions, condition)
	}

	for _, condition := range conditions {
		lq.lambdaAST.Left.WhereCondition = append(lq.lambdaAST.Left.WhereCondition,
			&parser.ParenExpr{Param: condition})
	}

	return lq.runQuery(lq.lambdaAST.Left)
}

func (lq *lq) runLinkQuery() (*QueryResult, error) {
	leftRes, err := lq.runQuery(lq.lambdaAST.Left)
	if err != nil {
		return nil, err
	}

	binExpr := lq.lambdaAST.WhereCondition[0].(*parser.BinaryExpr)
	condition := lq.newCondition(binExpr.RHS.String(), binExpr.Op, lq.findValues(leftRes, binExpr.LHS.String()).unique())
	if condition == nil {
		return leftRes, nil
	}

	var rightRes []*QueryResult

	for idx := range lq.lambdaAST.Right {
		lq.lambdaAST.Right[idx].WhereCondition = append(lq.lambdaAST.Right[idx].WhereCondition,
			&parser.ParenExpr{Param: condition})

		res, err := lq.runQuery(lq.lambdaAST.Right[idx])
		if err != nil {
			return nil, err
		}

		rightRes = append(rightRes, res)
	}

	return lq.mergeResult(leftRes, rightRes), nil
}

// mergeResult
//     default has 1 time lines
func (lq *lq) mergeResult(leftRes *QueryResult, rightRes []*QueryResult) *QueryResult {
	if len(leftRes.Series) == 0 {
		return leftRes
	}

	binExpr := lq.lambdaAST.WhereCondition[0].(*parser.BinaryExpr)

	lHitIdx := findField(leftRes, binExpr.LHS.String())
	if lHitIdx == -1 {
		return leftRes
	}

	for idx, res := range rightRes {

		var rlines [][]interface{}
		var rColumns []string

		lr := leftRes.Series
		operator := binExpr.Op.String()

		rHitIdx := findField(res, binExpr.RHS.String())

		if rHitIdx == -1 {
			for _, target := range lq.lambdaAST.Right[idx].Targets {
				rColumns = append(rColumns, target.String2())
			}
			rlines = append(rlines, make([]interface{}, len(lq.lambdaAST.Right[idx].Targets)))
		} else {

			rlines = mergeLines(res)
			if len(rlines) == 0 {
				l.Warnf("right DQL [%d] right result is empty", idx)
				return leftRes
			}
			// add nil line
			rlines = append(rlines, make([]interface{}, len(rlines[0])))

			rColumns = getColumns(res)
		}

		for sidx, row := range lr {
			lr[sidx].Columns = append(lr[sidx].Columns, rColumns...)

			for vidx, values := range row.Values {
				added := false

				for _, line := range rlines[:len(rlines)-1] {
					if contrast(values[lHitIdx], operator, line[rHitIdx]) {
						lr[sidx].Values[vidx] = append(lr[sidx].Values[vidx], line...)
						added = true
						break
					}
				}

				if !added {
					lr[sidx].Values[vidx] = append(lr[sidx].Values[vidx], rlines[len(rlines)-1]...)
				}
			}
		}
	}

	return leftRes
}

func (lq *lq) newCondition(leftName string, operator parser.ItemType, values []interface{}) *parser.BinaryExpr {
	var arr []*parser.BinaryExpr

	for _, v := range values {
		arr = append(arr, &parser.BinaryExpr{
			LHS: &parser.Identifier{Name: leftName},
			Op:  operator,
			RHS: &parser.StringLiteral{Val: fmt.Sprintf("%s", v)},
		})
	}

	var res *parser.BinaryExpr

	switch len(arr) {
	case 0:
		return nil
	case 1:
		res = arr[0]
	default:
		res = arr[0]
		for idx := 1; idx < len(arr); idx++ {
			res = &parser.BinaryExpr{
				LHS: res,
				Op:  parser.OR,
				RHS: arr[idx],
			}
		}
	}

	return res
}

type valueList []interface{}

func (list valueList) unique() valueList {
	var keys = make(map[interface{}]interface{})
	res := []interface{}{}

	for _, v := range list {
		if v == nil {
			continue
		}

		if _, ok := keys[v]; !ok {
			keys[v] = nil
			res = append(res, v)
		}
	}

	return res
}

func (lq *lq) findValues(res *QueryResult, fieldName string) valueList {
	if res == nil || len(res.Series) == 0 {
		return nil
	}

	var list []interface{}

	for _, row := range res.Series {
		for idx, column := range row.Columns {
			if column != fieldName {
				continue
			}

			for _, values := range row.Values {
				if len(values) > idx {
					list = append(list, values[idx])
				}
			}
			break
		}
	}

	return list
}

// contrast
// FIXME: It's loooooong! float==float is undefined
func contrast(x interface{}, op string, y interface{}) (b bool) {
	var (
		float  []float64
		str    []string
		booler []bool
	)
	var err error

	const typeErr = "mismatch of type: %s(%v) %s %s(%v)"

	switch x.(type) {
	case json.Number:
		var xx float64
		xx, err = x.(json.Number).Float64()
		if err != nil {
			l.Warn(err)
			return
		}

		float = append(float, xx)

		switch y.(type) {
		case json.Number:
			var yy float64
			yy, err = y.(json.Number).Float64()
			if err != nil {
				return
			}
			float = append(float, yy)
		case float64:
			float = append(float, y.(float64))
		case int64:
			float = append(float, float64(y.(int64)))
		default:
			l.Warnf(typeErr, reflect.TypeOf(x), x, op, reflect.TypeOf(y), y)
			return
		}

	case int64:
		float = append(float, float64(x.(int64)))

		switch y.(type) {
		case json.Number:
			var yy float64
			yy, err = y.(json.Number).Float64()
			if err != nil {
				l.Warn(err)
				return
			}
			float = append(float, yy)
		case float64:
			float = append(float, y.(float64))
		case int64:
			float = append(float, float64(y.(int64)))
		default:
			l.Warnf(typeErr, reflect.TypeOf(x), x, op, reflect.TypeOf(y), y)
			return
		}

	case float64:
		float = append(float, x.(float64))

		switch y.(type) {
		case json.Number:
			var yy float64
			yy, err = y.(json.Number).Float64()
			if err != nil {
				l.Warn(err)
				return
			}
			float = append(float, yy)
		case float64:
			float = append(float, y.(float64))
		case int64:
			float = append(float, float64(y.(int64)))
		default:
			l.Warnf(typeErr, reflect.TypeOf(x), x, op, reflect.TypeOf(y), y)
			return
		}

	case string:
		yy, ok := y.(string)
		if !ok {
			l.Warnf(typeErr, reflect.TypeOf(x), x, op, reflect.TypeOf(y), y)
			return

		}
		str = append(str, x.(string))
		str = append(str, yy)
	case bool:
		yy, ok := y.(bool)
		if !ok {
			l.Warnf(typeErr, reflect.TypeOf(x), x, op, reflect.TypeOf(y), y)
			return
		}
		booler = append(booler, x.(bool))
		booler = append(booler, yy)

	case nil:
		booler = append(booler, true)
		if y == nil {
			booler = append(booler, true)
		} else {
			booler = append(booler, false)
		}

	default:
		l.Warnf("mismatch of type: %s(%v)", reflect.TypeOf(x), x)
		return
	}

	switch op {
	case "=":
		b = reflect.DeepEqual(x, y)
		return
	case "!=":
		if len(float) != 0 {
			b = float[0] != float[1]
			return
		}
		if len(str) != 0 {
			b = str[0] != str[1]
			return
		}
		if len(booler) != 0 {
			b = booler[0] != booler[1]
			return
		}
	case "<=":
		if len(float) != 0 {
			b = float[0] <= float[1]
			return
		}
	case "<":
		if len(float) != 0 {
			b = float[0] < float[1]
			return
		}
	case ">=":
		if len(float) != 0 {
			b = float[0] >= float[1]
			return
		}
	case ">":
		if len(float) != 0 {
			b = float[0] > float[1]
			return
		}
	default:
		l.Warn("unexpected operator")
		return
	}

	l.Warn("the operator is not available for this type")
	return
}

func findField(res *QueryResult, field string) int {
	for _, row := range res.Series {
		for idx, column := range row.Columns {
			if column == field {
				return idx
			}
		}
	}

	return -1
}

func mergeLines(res *QueryResult) [][]interface{} {
	var values [][]interface{}

	for _, row := range res.Series {
		for _, value := range row.Values {
			values = append(values, value)
		}
	}

	return values
}

func getColumns(res *QueryResult) []string {
	for _, row := range res.Series {
		return row.Columns
	}
	return nil
}
