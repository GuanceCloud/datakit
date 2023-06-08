// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package funcs

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/GuanceCloud/platypus/pkg/ast"
	conv "github.com/spf13/cast"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/ptinput"
)

func getKeyName(node *ast.Node) (string, error) {
	var key string

	switch node.NodeType { //nolint:exhaustive
	case ast.TypeIdentifier:
		key = node.Identifier.Name
	case ast.TypeAttrExpr:
		key = node.AttrExpr.String()
	case ast.TypeStringLiteral:
		key = node.StringLiteral.Val
	default:
		return "", fmt.Errorf("expect StringLiteral or Identifier or AttrExpr, got %s",
			node.NodeType)
	}
	return key, nil
}

func isPlVarbOrFunc(node *ast.Node) bool {
	if node == nil {
		return false
	}
	switch node.NodeType { //nolint:exhaustive
	case ast.TypeIdentifier, ast.TypeAttrExpr, ast.TypeCallExpr:
		return true
	default:
		return false
	}
}

const (
	InvalidInt   = math.MinInt32
	DefaultInt   = int64(0xdeadbeef)
	DefaultStr   = ""
	InvalidStr   = "deadbeaf"
	InvalidFloat = math.SmallestNonzeroFloat64
	DefaultFloat = float64(0.0)
)

func fixYear(now time.Time, x int64) (int, error) {
	if x == DefaultInt {
		return now.Year(), nil
	}

	if x < 0 {
		return -1, fmt.Errorf("year should larger than 0")
	}

	// year like 02 -> 2002, 21 -> 2021
	if x < int64(100) { //nolint:gomnd
		x += int64(now.Year() / 100 * 100) //nolint:gomnd
		return int(x), nil
	}

	return int(x), nil
}

func fixMonth(now time.Time, x string) (time.Month, error) {
	if x == DefaultStr {
		return now.Month(), nil
	} else {
		if v, err := strconv.ParseInt(x, 10, 64); err == nil { // digital month: i.e., 01, 11
			if v < 1 || v > 12 {
				return time.Month(-1), fmt.Errorf("month should between [1,12], got %x,", x)
			}
			return time.Month(v), nil
		} else { // month like aug/august, january/jan
			v, ok := monthMaps[strings.ToLower(x)]
			if !ok {
				return InvalidInt, fmt.Errorf("unknown month %s", x)
			}
			return v, nil
		}
	}
}

func fixDay(now time.Time, x int64) (int, error) {
	if x == DefaultInt {
		return now.Day(), nil
	}

	if x < 1 || x > 31 {
		return -1, fmt.Errorf("month day should between [1,31], got %d", x)
	}

	return int(x), nil
}

func fixHour(now time.Time, x int64) (int, error) {
	if x == DefaultInt {
		return now.Hour(), nil
	}

	if x < 0 || x > 23 {
		return -1, fmt.Errorf("hour should between [0,24], got %d", x)
	}

	return int(x), nil
}

func fixMinute(now time.Time, x int64) (int, error) {
	if x == DefaultInt {
		return now.Minute(), nil
	}

	if x < 0 || x > 59 {
		return -1, fmt.Errorf("minute should between [0,59], got %d", x)
	}

	return int(x), nil
}

func fixSecond(x int64) (int, error) {
	if x == DefaultInt {
		return 0, nil
	}

	if x < 0 || x > 59 {
		return -1, fmt.Errorf("second should between [0,59], got %d", x)
	}

	return int(x), nil
}

func tz(s string) (z *time.Location, err error) {
	z, err = time.LoadLocation(s)
	if err != nil {
		if _, ok := timezoneList[s]; !ok {
			return nil, fmt.Errorf("unknown timezone %s", s)
		}

		z, err = time.LoadLocation(timezoneList[s])
		if err != nil {
			return nil, err
		}
	}

	return z, nil
}

func doCast(result interface{}, tInfo string) (interface{}, ast.DType) {
	switch strings.ToLower(tInfo) {
	case "bool":
		return conv.ToBool(result), ast.Bool

	case "int":
		return conv.ToInt64(conv.ToFloat64(result)), ast.Int

	case "float":
		return conv.ToFloat64(result), ast.Float

	case "str":
		return conv.ToString(result), ast.String
	}

	return nil, ast.Nil
}

func reindexFuncArgs(fnStmt *ast.CallExpr, keyList []string, reqParm int) error {
	// reqParm >= 1, if < 0, no optional args
	args := fnStmt.Param

	if reqParm < 0 || reqParm > len(keyList) {
		reqParm = len(keyList)
	}

	if len(args) > len(keyList) {
		return fmt.Errorf("the number of parameters does not match")
	}

	beforPosArg := true

	kMap := map[string]int{}
	for k, v := range keyList {
		kMap[v] = k
	}

	ret := make([]*ast.Node, len(keyList))

	for idx, arg := range args {
		if arg.NodeType == ast.TypeAssignmentExpr {
			if beforPosArg {
				beforPosArg = false
			}
			kname, err := getKeyName(arg.AssignmentExpr.LHS)
			if err != nil {
				return err
			}
			kIndex, ok := kMap[kname]
			if !ok {
				return fmt.Errorf("argument %s does not exist", kname)
			}
			ret[kIndex] = arg.AssignmentExpr.RHS
		} else {
			if !beforPosArg {
				return fmt.Errorf("positional arguments cannot follow keyword arguments")
			}
			ret[idx] = arg
		}
	}

	for i := 0; i < reqParm; i++ {
		if v := ret[i]; v == nil {
			return fmt.Errorf("parameter %s is required", keyList[i])
		}
	}

	fnStmt.Param = ret
	return nil
}

func getPoint(in any) (*ptinput.Point, error) {
	if in == nil {
		return nil, fmt.Errorf("nil ptr: input")
	}

	pt, ok := in.(*ptinput.Point)

	if !ok {
		return nil, fmt.Errorf("typeof input is not Point")
	}

	if pt == nil {
		return nil, fmt.Errorf("nil ptr: input")
	}
	return pt, nil
}

func getPtKey(in any, key string) (any, ast.DType, error) {
	pt, err := getPoint(in)
	if err != nil {
		return nil, ast.Invalid, err
	}

	if key == "_" {
		key = ptinput.Originkey
	}
	return pt.Get(key)
}

func deletePtKey(in any, key string) {
	pt, err := getPoint(in)
	if err != nil {
		return
	}

	if key == "_" {
		key = ptinput.Originkey
	}

	pt.Delete(key)
}

func pointTime(in any) int64 {
	pt, err := getPoint(in)
	if err != nil {
		return time.Now().UnixNano()
	}
	if pt.Time.IsZero() {
		return time.Now().UnixNano()
	} else {
		return pt.Time.UnixNano()
	}
}

func addKey2PtWithVal(in any, key string, value any, dtype ast.DType, kind ptinput.KeyKind) error {
	pt, err := getPoint(in)
	if err != nil {
		return err
	}

	if key == "_" {
		key = ptinput.Originkey
	}

	switch kind { //nolint:exhaustive
	case ptinput.KindPtTag:
		return pt.SetTag(key, value, dtype)
	default:
		return pt.Set(key, value, dtype)
	}
}

func renamePtKey(in any, to, from string) error {
	if to == "_" {
		to = ptinput.Originkey
	}

	if from == "_" {
		from = ptinput.Originkey
	}

	if to == from {
		return nil
	}

	if in == nil {
		return fmt.Errorf("nil ptr: input")
	}

	pt, err := getPoint(in)
	if err != nil {
		return err
	}

	if v, ok := pt.Fields[from]; ok {
		pt.Fields[to] = v
		delete(pt.Fields, from)
	} else if v, ok := pt.Tags[from]; ok {
		pt.Tags[to] = v
		delete(pt.Tags, from)
	} else {
		return fmt.Errorf("key(from) %s not found", from)
	}

	return nil
}

func setMeasurement(in any, val string) error {
	pt, err := getPoint(in)
	if err != nil {
		return err
	}
	pt.Name = val
	return nil
}

func markPtDrop(in any) error {
	pt, err := getPoint(in)
	if err != nil {
		return err
	}

	pt.Drop = true

	return nil
}
