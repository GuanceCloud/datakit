// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package parser

import (
	"math"
	"reflect"
	"regexp"
	"strings"
)

type KVs interface {
	Get(k string) (v any, ok bool)
}

type tfData struct {
	tags   map[string]string
	fields map[string]any
}

func (d *tfData) Get(name string) (any, bool) {
	if v, ok := d.tags[name]; ok {
		return v, true
	}

	if v, ok := d.fields[name]; ok {
		return v, true
	}

	return nil, false
}

func newtf(tags map[string]string, fields map[string]any) *tfData {
	return &tfData{
		tags:   tags,
		fields: fields,
	}
}

func (p *ParenExpr) Eval(data KVs) bool {
	if p.Param == nil {
		return false
	}

	switch expr := p.Param.(type) {
	case Evaluable:
		return expr.Eval(data)
	default:
		log.Errorf("ParenExpr's Param should be Evaluable")
	}

	return false
}

func (e *BinaryExpr) Eval(data KVs) bool {
	switch e.Op {
	case AND:
		for _, expr := range []Node{e.LHS, e.RHS} {
			switch expr.(type) {
			case Evaluable:
			default:
				log.Errorf("LHS and RHS should be BinaryExpr or ParenExpr")
				return false
			}
		}

		return e.LHS.(Evaluable).Eval(data) && e.RHS.(Evaluable).Eval(data)

	case OR: // LHS/RHS should be BinaryExpr

		for _, expr := range []Node{e.LHS, e.RHS} {
			switch expr.(type) {
			case Evaluable:
			default:
				log.Errorf("LHS and RHS should be BinaryExpr or ParenExpr")
				return false
			}
		}

		return e.LHS.(Evaluable).Eval(data) || e.RHS.(Evaluable).Eval(data)

	default:
		return e.doEval(data)
	}
}

func (e *BinaryExpr) doEval(data KVs) bool {
	switch e.Op {
	case GTE, GT, LT, LTE, NEQ, EQ, IN, NOT_IN, MATCH, NOT_MATCH:
	default:
		log.Errorf("unsupported OP %s", e.Op.String())
		return false
	}

	return e.singleEval(data)
}

const float64EqualityThreshold = 1e-9

// see: https://stackoverflow.com/a/47969546/342348
func almostEqual(a, b float64) bool {
	return math.Abs(a-b) <= float64EqualityThreshold
}

func toFloat64(f interface{}) float64 {
	switch v := f.(type) {
	case float32:
		return float64(v)
	case float64:
		return v
	default:
		log.Error("should not been here")
		return 0.0
	}
}

func toInt64(i interface{}) int64 {
	switch v := i.(type) {
	case int:
		return int64(v)
	case int8:
		return int64(v)
	case int16:
		return int64(v)
	case int32:
		return int64(v)
	case int64:
		return v
	case uint:
		return int64(v)
	case uint8:
		return int64(v)
	case uint16:
		return int64(v)
	case uint32:
		return int64(v)
	default:
		log.Error("should not been here")
		return 0
	}
}

func binEval(op ItemType, lhs, rhs interface{}) bool {
	if _, ok := rhs.(*Regex); !ok {
		tl := reflect.TypeOf(lhs).String()
		tr := reflect.TypeOf(rhs).String()
		switch op {
		case GTE, GT, LT, LTE, EQ, NEQ: // type conflict detecting on comparison expr
			if tl != tr {
				log.Warnf("type conflict %+#v(%s) <> %+#v(%s)", lhs, reflect.TypeOf(lhs), rhs, reflect.TypeOf(rhs))
				return false
			}

		default:
			log.Warnf("op %s should not been here", op.String())
			return false
		}
	}

	switch op {
	case EQ:
		switch lv := lhs.(type) {
		case float64:
			if f, ok := rhs.(float64); !ok {
				return false
			} else {
				return almostEqual(lv, f)
			}

		default: // NOTE: interface{} EQ/NEQ, see: https://stackoverflow.com/a/34246225/342348
			switch reg := rhs.(type) {
			case *Regex:
				ok, err := regexp.MatchString(reg.Regex, lhs.(string))
				if err != nil {
					log.Error(err)
				}

				return ok
			default:
				return lhs == rhs
			}
		}

	case MATCH:
		return rhs.(*Regex).Re.MatchString(lhs.(string))

	case NOT_MATCH:
		return !rhs.(*Regex).Re.MatchString(lhs.(string))

	case NEQ:
		return lhs != rhs

	case GTE, GT, LT, LTE: // rhs/lhs should be number or string
		switch lv := lhs.(type) {
		case int, int8, int16, int32, int64,
			uint, uint8, uint16, uint32:

			if i, ok := rhs.(int64); !ok {
				log.Warnf("rhs not int64")
				return false
			} else {
				return cmpint(op, toInt64(lv), i)
			}

		case bool: // bool not support >/>=/</<=
			return false
		case string:
			if s, ok := rhs.(string); !ok {
				return false
			} else {
				return cmpstr(op, lv, s)
			}
		case float32, float64:
			if f, ok := rhs.(float64); !ok {
				return false
			} else {
				return cmpfloat(op, toFloat64(lv), f)
			}
		}
	}

	return false
}

func cmpstr(op ItemType, l, r string) bool {
	switch op {
	case GTE:
		return strings.Compare(l, r) >= 0
	case LTE:
		return strings.Compare(l, r) <= 0
	case LT:
		return strings.Compare(l, r) < 0
	case GT:
		return strings.Compare(l, r) > 0
	default:
		log.Warn("should not been here, %s %s %s", l, op.String(), r)
	}
	return false
}

func cmpint(op ItemType, l, r int64) bool {
	switch op {
	case GTE:
		return l >= r
	case LTE:
		return l <= r
	case LT:
		return l < r
	case GT:
		return l > r
	default:
		log.Warn("should not been here, %d %s %d", l, op.String(), r)
	}
	return false
}

func cmpfloat(op ItemType, l, r float64) bool {
	switch op {
	case GTE:
		return l >= r
	case LTE:
		return l <= r
	case LT:
		return l < r
	case GT:
		return l > r
	default:
		log.Warn("should not been here, %f %s %f", l, op.String(), r)
	}
	return false
}

func (e *BinaryExpr) singleEval(data KVs) bool {
	if e.LHS == nil || e.RHS == nil {
		log.Warn("LHS or RHS nil, should not been here")
		return false
	}

	// first: fetch right-handle-symbol
	var lit interface{}
	var arr []interface{}
	switch rhs := e.RHS.(type) {
	case *StringLiteral:
		lit = rhs.Val

	case *NumberLiteral:
		if rhs.IsInt {
			lit = rhs.Int
		} else {
			lit = rhs.Float
		}

	case NodeList:
		for _, elem := range rhs {
			switch x := elem.(type) {
			case *StringLiteral:
				arr = append(arr, x.Val)
			case *NumberLiteral:
				if x.IsInt {
					arr = append(arr, x.Int)
				} else {
					arr = append(arr, x.Float)
				}
			case *Regex:
				arr = append(arr, x)
			default:
				log.Warnf("unsupported node list with type `%s'", reflect.TypeOf(elem).String())
			}
		}

	case *Regex:
		lit = rhs

	default:
		log.Errorf("invalid RHS, got type `%s'", reflect.TypeOf(e.RHS).String())
		return false
	}

	// first: fetch left-handle-symbol and OP on right-handle-symbol
	switch lhs := e.LHS.(type) {
	case *Identifier:
		name := lhs.Name

		switch e.Op {
		case MATCH, NOT_MATCH:
			for _, item := range e.RHS.(NodeList) {
				if v, ok := data.Get(name); ok {
					switch x := v.(type) {
					case string:
						if binEval(e.Op, x, item) {
							return true
						}
					default:
						continue
					}
				}
			}
			return false

		case IN:
			for _, item := range arr {
				if v, ok := data.Get(name); ok {
					if binEval(EQ, v, item) {
						return true
					}
				}
			}
			return false

		case NOT_IN:
			for _, item := range arr {
				if v, ok := data.Get(name); ok {
					if binEval(EQ, v, item) {
						return false
					}
				}
			}

			return true

		case GTE, GT, LT, LTE, NEQ, EQ:

			if v, ok := data.Get(name); ok {
				if binEval(e.Op, v, lit) {
					return true
				}
			}
		}

	default:
		log.Errorf("unknown LHS type, expect Identifier, got `%s'", reflect.TypeOf(e.LHS).String())
	}
	return false
}
