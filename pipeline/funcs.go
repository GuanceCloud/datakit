package pipeline

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/parser"
)

type PipelineFunc func(p *Pipeline, node parser.Node) (*Pipeline, error)

var funcsMap = map[string]PipelineFunc{
	"add_key":               Addkey,
	"add_pattern":           AddPattern,
	"cast":                  Cast,
	"datetime":              DateTime,
	"default_time":          DefaultTime,
	"default_time_with_fmt": DefaultTimeWithFmt,
	"drop_key":              Dropkey,
	"drop_origin_data":      DropOriginData,
	"expr":                  Expr,
	"geoip":                 GeoIP,
	"grok":                  Grok,
	"group_between":         Group,
	"group_in":              GroupIn,
	"json":                  JSON,
	"json_all":              JSONAll,
	"lowercase":             Lowercase,
	"nullif":                NullIf,
	"rename":                Rename,
	"strfmt":                Strfmt,
	"uppercase":             Uppercase,
	"url_decode":            URLDecode,
	"user_agent":            UserAgent,
	"parse_duration":        ParseDuration,
	"parse_date":            ParseDate,
	"cover":                 Dz,
	"replace":               Replace,
}

func fexpr(node parser.Node) *parser.FuncExpr {
	if x, ok := node.(*parser.FuncExpr); ok {
		return x
	}

	return nil
}

func arglist(fe *parser.FuncExpr, n int) parser.FuncArgList {
	if x, ok := fe.Param[n].(parser.FuncArgList); ok {
		return x
	}
	return nil
}

func JSON(p *Pipeline, node parser.Node) (*Pipeline, error) {
	funcExpr := fexpr(node)
	if len(funcExpr.Param) < 2 || len(funcExpr.Param) > 3 {
		return p, fmt.Errorf("func %s expected 2 or 3 args", funcExpr.Name)
	}

	var key, old parser.Node

	switch v := funcExpr.Param[0].(type) {
	case *parser.AttrExpr, *parser.Identifier:
		key = v
	default:
		return p, fmt.Errorf("expect AttrExpr or Identifier, got %s",
			reflect.TypeOf(funcExpr.Param[0]).String())
	}

	switch v := funcExpr.Param[1].(type) {
	case *parser.AttrExpr, *parser.Identifier:
		old = v
	default:
		return p, fmt.Errorf("expect AttrExpr or Identifier, got %s",
			reflect.TypeOf(funcExpr.Param[1]).String())
	}

	newkey := old
	if len(funcExpr.Param) == 3 {
		switch v := funcExpr.Param[2].(type) {
		case *parser.AttrExpr, *parser.Identifier, *parser.StringLiteral:
			newkey = v
		default:
			return p, fmt.Errorf("expect AttrExpr or Identifier, got %s",
				reflect.TypeOf(funcExpr.Param[2]).String())
		}
	}

	cont, err := p.getContentStr(key)
	if err != nil {
		l.Debug(err)
		return p, nil
	}

	v, err := GsonGet(cont, old)
	if err != nil {
		l.Warn(err)
		return p, nil
	}

	if err := p.setContent(newkey, v); err != nil {
		l.Warn(err)
		return p, nil
	}

	return p, nil
}

func JSONAll(p *Pipeline, node parser.Node) (*Pipeline, error) {
	l.Warnf("warning:json_all is disabled")
	return p, nil
}

func Rename(p *Pipeline, node parser.Node) (*Pipeline, error) {
	funcExpr := fexpr(node)
	if len(funcExpr.Param) != 2 {
		return p, fmt.Errorf("func %s expected 2 args", funcExpr.Name)
	}

	var from, to parser.Node

	switch v := funcExpr.Param[0].(type) {
	case *parser.AttrExpr, *parser.StringLiteral:
		to = v
	default:
		return p, fmt.Errorf("expect string or AttrExpr, got `%s'",
			reflect.TypeOf(funcExpr.Param[0]).String())
	}

	switch v := funcExpr.Param[1].(type) {
	case *parser.AttrExpr, *parser.Identifier:
		from = v
	default:
		return p, fmt.Errorf("param key expect Identifier or AttrExpr, got `%s'",
			reflect.TypeOf(funcExpr.Param[1]).String())
	}

	v, err := p.getContent(from)
	if err != nil {
		l.Debug(err)
		return p, nil
	}

	if err := p.setContent(to, v); err != nil {
		l.Warn(err)
		return p, nil
	}

	delete(p.Output, from.String())

	return p, nil
}

func UserAgent(p *Pipeline, node parser.Node) (*Pipeline, error) {
	funcExpr := fexpr(node)
	if len(funcExpr.Param) != 1 {
		return p, fmt.Errorf("func `%s' expected 1 args", funcExpr.Name)
	}

	var key parser.Node

	switch v := funcExpr.Param[0].(type) {
	case *parser.AttrExpr, *parser.Identifier:
		key = v
	default:
		return p, fmt.Errorf("param key expect AttrExpr or Identifier, got %s",
			reflect.TypeOf(funcExpr.Param[0]).String())
	}

	cont, err := p.getContentStr(key)
	if err != nil {
		l.Warnf("key `%v' not exist, ignored", key)
		return p, nil //nolint:nilerr
	}

	dic := UserAgentHandle(cont)

	for k, val := range dic {
		if err := p.setContent(k, val); err != nil {
			l.Warn(err)
			return p, nil
		}
	}

	return p, nil
}

func URLDecode(p *Pipeline, node parser.Node) (*Pipeline, error) {
	funcExpr := fexpr(node)
	if len(funcExpr.Param) != 1 {
		return p, fmt.Errorf("func `%s' expected 1 args", funcExpr.Name)
	}

	var key parser.Node
	switch v := funcExpr.Param[0].(type) {
	case *parser.AttrExpr, *parser.Identifier:
		key = v
	default:
		return p, fmt.Errorf("param key expect AttrExpr or Identifier, got %s",
			reflect.TypeOf(funcExpr.Param[0]).String())
	}

	cont, err := p.getContentStr(key)
	if err != nil {
		l.Warnf("key `%v' not exist, ignored", key)
		return p, nil //nolint:nilerr
	}

	if v, err := UrldecodeHandle(cont); err != nil {
		return p, err
	} else if err := p.setContent(key, v); err != nil {
		l.Warn(err)
		return p, nil
	}

	return p, nil
}

func GeoIP(p *Pipeline, node parser.Node) (*Pipeline, error) {
	funcExpr := fexpr(node)
	if len(funcExpr.Param) != 1 {
		return p, fmt.Errorf("func `%s' expected 1 args", funcExpr.Name)
	}

	var key parser.Node
	switch v := funcExpr.Param[0].(type) {
	case *parser.AttrExpr, *parser.Identifier:
		key = v
	default:
		return p, fmt.Errorf("param key expect AttrExpr or Identifier, got %s",
			reflect.TypeOf(funcExpr.Param[0]).String())
	}

	cont, err := p.getContentStr(key)
	if err != nil {
		l.Warnf("key `%v' not exist, ignored", key)
		return p, nil //nolint:nilerr
	}

	if dic, err := GeoIPHandle(cont); err != nil {
		l.Warnf("GeoIPHandle: %s, ignored", err)
		return p, err
	} else {
		for k, v := range dic {
			if err := p.setContent(k, v); err != nil {
				l.Warn(err)
				return p, nil
			}
		}
	}

	return p, nil
}

func DateTime(p *Pipeline, node parser.Node) (*Pipeline, error) {
	funcExpr := fexpr(node)
	if len(funcExpr.Param) != 3 {
		return p, fmt.Errorf("func %s expected 3 args", funcExpr.Name)
	}

	var key parser.Node
	var precision, fmts string
	switch v := funcExpr.Param[0].(type) {
	case *parser.AttrExpr, *parser.Identifier:
		key = v
	default:
		return p, fmt.Errorf("param `key` expect AttrExpr or Identifier, got %s",
			reflect.TypeOf(funcExpr.Param[0]).String())
	}

	switch v := funcExpr.Param[1].(type) {
	case *parser.StringLiteral:
		precision = v.Val
	default:
		return p, fmt.Errorf("param `precision` expect StringLiteral, got %s",
			reflect.TypeOf(funcExpr.Param[1]).String())
	}

	switch v := funcExpr.Param[2].(type) {
	case *parser.StringLiteral:
		fmts = v.Val
	default:
		return p, fmt.Errorf("param `fmt` expect StringLiteral, got %s",
			reflect.TypeOf(funcExpr.Param[2]).String())
	}

	cont, err := p.getContent(key)
	if err != nil {
		l.Warnf("key `%v' not exist, ignored", key)
		return p, nil //nolint:nilerr
	}

	if v, err := DateFormatHandle(cont, precision, fmts); err != nil {
		return p, err
	} else if err := p.setContent(key, v); err != nil {
		l.Warn(err)
		return p, nil
	}

	return p, nil
}

func Expr(p *Pipeline, node parser.Node) (*Pipeline, error) {
	funcExpr := fexpr(node)
	if len(funcExpr.Param) != 2 {
		return p, fmt.Errorf("func `%s' expected 2 args", funcExpr.Name)
	}

	var key parser.Node
	var expr *parser.BinaryExpr

	switch v := funcExpr.Param[0].(type) {
	case *parser.BinaryExpr:
		expr = v
	default:
		return p, fmt.Errorf("param expr expect BinaryExpr, got `%s'",
			reflect.TypeOf(funcExpr.Param[0]).String())
	}

	switch v := funcExpr.Param[1].(type) {
	case *parser.Identifier, *parser.AttrExpr:
		key = v
	default:
		return p, fmt.Errorf("param key expect Identifier or AttrExpr, got `%s'",
			reflect.TypeOf(funcExpr.Param[1]).String())
	}

	if v, err := Calc(expr, p); err != nil {
		l.Warn(err)
		return p, nil
	} else if err = p.setContent(key, v); err != nil {
		l.Warn(err)
		return p, nil
	}

	return p, nil
}

func Strfmt(p *Pipeline, node parser.Node) (*Pipeline, error) {
	outdata := make([]interface{}, 0)

	funcExpr := fexpr(node)
	if len(funcExpr.Param) < 2 {
		return p, fmt.Errorf("func `%s' expected more than 2 args", funcExpr.Name)
	}

	var key parser.Node
	var fmts string
	switch v := funcExpr.Param[0].(type) {
	case *parser.Identifier, *parser.AttrExpr:
		key = v
	default:
		return p, fmt.Errorf("param key expect Identifier or AttrExpr, got `%s'",
			reflect.TypeOf(funcExpr.Param[0]).String())
	}

	switch v := funcExpr.Param[1].(type) {
	case *parser.StringLiteral:
		fmts = v.Val
	default:
		return p, fmt.Errorf("param fmt expect StringLiteral, got `%s'",
			reflect.TypeOf(funcExpr.Param[1]).String())
	}

	for i := 2; i < len(funcExpr.Param); i++ {
		switch v := funcExpr.Param[i].(type) {
		case *parser.Identifier:
			data, _ := p.getContent(v)
			outdata = append(outdata, data)
		case *parser.AttrExpr:
			data, _ := p.getContent(v)
			outdata = append(outdata, data)
		case *parser.NumberLiteral:
			if v.IsInt {
				outdata = append(outdata, v.Int)
			} else {
				outdata = append(outdata, v.Float)
			}
		default:
			outdata = append(outdata, v)
		}
	}

	strfmt := fmt.Sprintf(fmts, outdata...)
	if err := p.setContent(key, strfmt); err != nil {
		l.Warn(err)
		return p, nil
	}

	return p, nil
}

func ParseDuration(p *Pipeline, node parser.Node) (*Pipeline, error) {
	funcExpr := fexpr(node)
	if len(funcExpr.Param) != 1 {
		l.Warn("parse_duration(): invalid param")

		return nil, fmt.Errorf("func %s expect 1 arg", funcExpr.Name)
	}

	var key parser.Node
	switch v := funcExpr.Param[0].(type) {
	case *parser.Identifier, *parser.AttrExpr:
		key = v
	default:
		err := fmt.Errorf("param expect Identifier, got `%+#v', type `%s'",
			funcExpr.Param[0], reflect.TypeOf(funcExpr.Param[0]).String())

		l.Warn("parse_duration(): %s", err)

		return p, err
	}

	cont, err := p.getContent(key)
	if err != nil {
		l.Debug(err)
		return p, nil
	}

	duStr, ok := cont.(string)
	if !ok {
		return p, fmt.Errorf("parse_duration() expect string arg")
	}

	l.Debugf("parse duration %s", duStr)
	du, err := time.ParseDuration(duStr)
	if err != nil {
		l.Debug(err)
		return p, nil
	}

	if err := p.setContent(key, int64(du)); err != nil {
		l.Warn(err)
		return p, nil
	}
	return p, nil
}

func ParseDate(p *Pipeline, node parser.Node) (*Pipeline, error) {
	funcExpr := fexpr(node)
	if len(funcExpr.Param) != 9 {
		l.Warn("parse_duration(): invalid param")

		return nil, fmt.Errorf("func %s expect 1 arg", funcExpr.Name)
	}

	var key parser.Node
	var yy, mm, dd, hh, mi, ss, ns, zone string
	switch v := funcExpr.Param[0].(type) {
	case *parser.AttrExpr, *parser.StringLiteral:
		key = v
	default:
		return p, fmt.Errorf("expect string or AttrExpr, got `%s'",
			reflect.TypeOf(funcExpr.Param[0]).String())
	}

	switch v := funcExpr.Param[1].(type) {
	case *parser.StringLiteral:
		yy = v.Val
	default:
		return p, fmt.Errorf("param `precision` expect StringLiteral, got %s",
			reflect.TypeOf(funcExpr.Param[1]).String())
	}

	switch v := funcExpr.Param[2].(type) {
	case *parser.StringLiteral:
		mm = v.Val
	default:
		return p, fmt.Errorf("param `fmt` expect StringLiteral, got %s",
			reflect.TypeOf(funcExpr.Param[2]).String())
	}

	switch v := funcExpr.Param[3].(type) {
	case *parser.StringLiteral:
		dd = v.Val
	default:
		return p, fmt.Errorf("param `fmt` expect StringLiteral, got %s",
			reflect.TypeOf(funcExpr.Param[3]).String())
	}

	switch v := funcExpr.Param[4].(type) {
	case *parser.StringLiteral:
		hh = v.Val
	default:
		return p, fmt.Errorf("param `fmt` expect StringLiteral, got %s",
			reflect.TypeOf(funcExpr.Param[2]).String())
	}

	switch v := funcExpr.Param[5].(type) {
	case *parser.StringLiteral:
		mi = v.Val
	default:
		return p, fmt.Errorf("param `fmt` expect StringLiteral, got %s",
			reflect.TypeOf(funcExpr.Param[2]).String())
	}

	switch v := funcExpr.Param[6].(type) {
	case *parser.StringLiteral:
		ss = v.Val
	default:
		return p, fmt.Errorf("param `fmt` expect StringLiteral, got %s",
			reflect.TypeOf(funcExpr.Param[2]).String())
	}

	switch v := funcExpr.Param[7].(type) {
	case *parser.StringLiteral:
		ns = v.Val
	default:
		return p, fmt.Errorf("param `fmt` expect StringLiteral, got %s",
			reflect.TypeOf(funcExpr.Param[2]).String())
	}

	switch v := funcExpr.Param[8].(type) {
	case *parser.StringLiteral:
		zone = v.Val
	default:
		return p, fmt.Errorf("param `fmt` expect StringLiteral, got %s",
			reflect.TypeOf(funcExpr.Param[2]).String())
	}
	res := parseDate(yy, mm, dd, hh, mi, ss, ns, zone)

	if err := p.setContent(key, res); err != nil {
		l.Warn(err)
		return p, nil
	}
	return p, nil
}

func Cast(p *Pipeline, node parser.Node) (*Pipeline, error) {
	funcExpr := fexpr(node)
	if len(funcExpr.Param) != 2 {
		return p, fmt.Errorf("func `%s' expected 2 args", funcExpr.Name)
	}

	var key parser.Node
	var castType string
	switch v := funcExpr.Param[0].(type) {
	case *parser.Identifier, *parser.AttrExpr:
		key = v
	default:
		return p, fmt.Errorf("param key expect Identifier or AttrExpr, got `%s'",
			reflect.TypeOf(funcExpr.Param[0]).String())
	}

	switch v := funcExpr.Param[1].(type) {
	case *parser.StringLiteral:
		castType = v.Val
	default:
		return p, fmt.Errorf("param type expect StringLiteral, got `%s'",
			reflect.TypeOf(funcExpr.Param[1]).String())
	}

	cont, err := p.getContent(key)
	if err != nil {
		l.Debug(err)
		return p, nil
	}

	val := cast(cont, castType)
	if err = p.setContent(key, val); err != nil {
		l.Warn(err)
		return p, nil
	}

	return p, nil
}

func Group(p *Pipeline, node parser.Node) (*Pipeline, error) {
	funcExpr := fexpr(node)
	if len(funcExpr.Param) < 3 || len(funcExpr.Param) > 4 {
		return p, fmt.Errorf("func `%s' expected 3 or 4 args", funcExpr.Name)
	}

	set := arglist(funcExpr, 1)
	value := funcExpr.Param[2]

	var key parser.Node
	switch v := funcExpr.Param[0].(type) {
	case *parser.AttrExpr, *parser.Identifier:
		key = v
	default:
		return p, fmt.Errorf("param key expect AttrExpr or Identifier, got %s",
			reflect.TypeOf(funcExpr.Param[0]).String())
	}

	newkey := key
	var start, end float64

	if len(funcExpr.Param) == 4 {
		switch v := funcExpr.Param[3].(type) {
		case *parser.AttrExpr, *parser.StringLiteral:
			newkey = v
		default:
			return p, fmt.Errorf("param new-key expect AttrExpr or StringLiteral, got %s",
				reflect.TypeOf(funcExpr.Param[3]).String())
		}
	}

	if len(set) != 2 {
		return p, fmt.Errorf("param between range value `%v' is not expected", set)
	}

	if v, ok := set[0].(*parser.NumberLiteral); !ok {
		return p, fmt.Errorf("range value `%v' is not expected", set)
	} else {
		if v.IsInt {
			start = float64(v.Int)
		} else {
			start = v.Float
		}
	}

	if v, ok := set[1].(*parser.NumberLiteral); !ok {
		return p, fmt.Errorf("range value `%v' is not expected", set)
	} else {
		if v.IsInt {
			end = float64(v.Int)
		} else {
			end = v.Float
		}

		if start > end {
			return p, fmt.Errorf("range value start %v must le end %v", start, end)
		}
	}

	cont, err := p.getContent(key)
	if err != nil {
		l.Warnf("key `%v' not exist, ignored", key)
		return p, nil //nolint:nilerr
	}

	if GroupHandle(cont, start, end) {
		switch v := value.(type) {
		case *parser.NumberLiteral:
			if v.IsInt {
				if err := p.setContent(newkey, v.Int); err != nil {
					l.Warn(err)
					return p, nil
				}
			} else {
				if err := p.setContent(newkey, v.Float); err != nil {
					l.Warn(err)
					return p, nil
				}
			}
		case *parser.StringLiteral:
			if err := p.setContent(newkey, v.Val); err != nil {
				l.Warn(err)
				return p, nil
			}
		case *parser.BoolLiteral:
			if err := p.setContent(newkey, v.Val); err != nil {
				l.Warn(err)
				return p, nil
			}

		default:
			l.Errorf("unknown group elements: %s", reflect.TypeOf(value).String())
			return p, fmt.Errorf("unsupported group type")
		}
	}

	return p, nil
}

func GroupIn(p *Pipeline, node parser.Node) (*Pipeline, error) {
	setdata := make([]interface{}, 0)
	funcExpr := fexpr(node)
	if len(funcExpr.Param) < 3 || len(funcExpr.Param) > 4 {
		return nil, fmt.Errorf("func %s expected 3 or 4 args", funcExpr.Name)
	}

	set := arglist(funcExpr, 1)
	value := funcExpr.Param[2]

	var key parser.Node
	switch v := funcExpr.Param[0].(type) {
	case *parser.AttrExpr, *parser.Identifier:
		key = v
	default:
		return p, fmt.Errorf("param key expect AttrExpr or Identifier, got %s",
			reflect.TypeOf(funcExpr.Param[0]).String())
	}

	newkey := key
	if len(funcExpr.Param) == 4 {
		switch v := funcExpr.Param[3].(type) {
		case *parser.AttrExpr, *parser.StringLiteral:
			newkey = v
		default:
			return p, fmt.Errorf("param new-key expect AttrExpr or StringLiteral, got %s",
				reflect.TypeOf(funcExpr.Param[3]).String())
		}
	}

	for _, node := range set {
		switch v := node.(type) {
		case *parser.Identifier:
			cont, err := p.getContent(v.Name)
			if err != nil {
				l.Warnf("key `%v' not exist, ignored", key)
				return p, nil //nolint:nilerr
			}
			setdata = append(setdata, cont)
		case *parser.NumberLiteral:
			if v.IsInt {
				setdata = append(setdata, v.Int)
			} else {
				setdata = append(setdata, v.Float)
			}
		case *parser.BoolLiteral:
			setdata = append(setdata, v.Val)
		case *parser.StringLiteral:
			setdata = append(setdata, v.Val)
		default:
			setdata = append(setdata, v)
		}
	}

	cont, err := p.getContent(key)
	if err != nil {
		l.Warnf("key `%v' not exist, ignored", key)
		return p, nil //nolint:nilerr
	}

	if GroupInHandle(cont, setdata) {
		switch v := value.(type) {
		case *parser.NumberLiteral:
			if v.IsInt {
				if err := p.setContent(newkey, v.IsInt); err != nil {
					l.Warn(err)
					return p, nil
				}
			} else if err := p.setContent(newkey, v.Float); err != nil {
				l.Warn(err)
				return p, nil
			}
		case *parser.StringLiteral:
			if err := p.setContent(newkey, v.Val); err != nil {
				l.Warn(err)
				return p, nil
			}
		case *parser.BoolLiteral:
			if err := p.setContent(newkey, v.Val); err != nil {
				l.Warn(err)
				return p, nil
			}
		}
	}

	return p, nil
}

func DefaultTime(p *Pipeline, node parser.Node) (*Pipeline, error) {
	funcExpr := fexpr(node)
	if len(funcExpr.Param) < 1 {
		return p, fmt.Errorf("func %s expected more than 1 args", funcExpr.Name)
	}

	var key parser.Node
	switch v := funcExpr.Param[0].(type) {
	case *parser.AttrExpr, *parser.Identifier:
		key = v
	default:
		return p, fmt.Errorf("param key expect AttrExpr or Identifier, got %s",
			reflect.TypeOf(funcExpr.Param[0]).String())
	}

	var tz string
	if len(funcExpr.Param) > 1 {
		switch v := funcExpr.Param[1].(type) {
		case *parser.StringLiteral:
			tz = v.Val
		default:
			return p, fmt.Errorf("param key expect StringLiteral, got %s",
				reflect.TypeOf(funcExpr.Param[1]).String())
		}
	}

	cont, err := p.getContentStr(key)
	if err != nil {
		l.Warnf("key `%v' not exist, ignored", key)
		return p, nil //nolint:nilerr
	}

	if v, err := TimestampHandle(p, cont, tz); err != nil {
		return p, fmt.Errorf("time convert fail error %w", err)
	} else if err := p.setContent(key, v); err != nil {
		l.Warn(err)
		return p, nil
	}

	return p, nil
}

func DefaultTimeWithFmt(p *Pipeline, node parser.Node) (*Pipeline, error) {
	var err error
	var goTimeFmt string
	var tz string
	var t time.Time
	timezone := time.Local

	funcExpr := fexpr(node)
	if len(funcExpr.Param) < 2 {
		return p, fmt.Errorf("func %s expected more than 2 args", funcExpr.Name)
	}

	var key parser.Node
	switch v := funcExpr.Param[0].(type) {
	case *parser.AttrExpr, *parser.Identifier:
		key = v
	default:
		return p, fmt.Errorf("param key expect AttrExpr or Identifier, got %s",
			reflect.TypeOf(funcExpr.Param[0]).String())
	}

	switch v := funcExpr.Param[1].(type) {
	case *parser.StringLiteral:
		goTimeFmt = v.Val
	default:
		return p, fmt.Errorf("param key expect StringLiteral, got %s",
			reflect.TypeOf(funcExpr.Param[1]).String())
	}

	if len(funcExpr.Param) > 2 {
		switch v := funcExpr.Param[2].(type) {
		case *parser.StringLiteral:
			tz = v.Val
		default:
			return p, fmt.Errorf("param key expect StringLiteral, got %s",
				reflect.TypeOf(funcExpr.Param[2]).String())
		}
	}

	timeStr, err := p.getContentStr(key)
	if err != nil {
		l.Warnf("key `%v' not exist, ignored", key)
		return p, nil //nolint:nilerr
	}

	if tz != "" {
		if tzCache, ok := p.timezone[tz]; ok {
			timezone = tzCache
		} else {
			timezone, err = time.LoadLocation(tz)
			if err == nil {
				p.setTimezone(tz, timezone)
			}
		}
	}

	if err == nil {
		t, err = time.ParseInLocation(goTimeFmt, timeStr, timezone)
	}

	if err != nil {
		l.Debugf("time string: %s, time format: %s, timezone: %s, error msg: %s",
			timeStr, goTimeFmt, tz, err)
		return p, err
	} else {
		if err := p.setContent(key, t.UnixNano()); err != nil {
			l.Warn(err)
			return p, nil
		}

		return p, nil
	}
}

func Uppercase(p *Pipeline, node parser.Node) (*Pipeline, error) {
	funcExpr := fexpr(node)
	if len(funcExpr.Param) != 1 {
		return p, fmt.Errorf("func %s expected 1 args", funcExpr.Name)
	}

	var key parser.Node
	switch v := funcExpr.Param[0].(type) {
	case *parser.Identifier, *parser.AttrExpr:
		key = v
	default:
		return p, fmt.Errorf("param key expect Identifier or AttrExpr, got %s",
			reflect.TypeOf(funcExpr.Param[0]).String())
	}

	cont, err := p.getContentStr(key)
	if err != nil {
		l.Debug(err)
		return p, nil
	}

	v := strings.ToUpper(cont)
	if err := p.setContent(key, v); err != nil {
		l.Warn(err)
		return p, nil
	}

	return p, nil
}

func Lowercase(p *Pipeline, node parser.Node) (*Pipeline, error) {
	funcExpr := fexpr(node)
	if len(funcExpr.Param) != 1 {
		return p, fmt.Errorf("func %s expected 1 args", funcExpr.Name)
	}

	var key parser.Node
	switch v := funcExpr.Param[0].(type) {
	case *parser.Identifier, *parser.AttrExpr:
		key = v
	default:
		return p, fmt.Errorf("param key expect Identifier or AttrExpr, got %s",
			reflect.TypeOf(funcExpr.Param[0]).String())
	}

	cont, err := p.getContentStr(key)
	if err != nil {
		l.Debug(err)
		return p, nil
	}

	v := strings.ToLower(cont)
	if err = p.setContent(key, v); err != nil {
		l.Warn(err)
		return p, nil
	}

	return p, nil
}

func NullIf(p *Pipeline, node parser.Node) (*Pipeline, error) {
	funcExpr := fexpr(node)
	if len(funcExpr.Param) != 2 {
		return p, fmt.Errorf("func %s expected 2 args", funcExpr.Name)
	}

	var key parser.Node
	switch v := funcExpr.Param[0].(type) {
	case *parser.AttrExpr, *parser.Identifier:
		key = v
	default:
		return p, fmt.Errorf("param key expect AttrExpr or Identifier, got %s",
			reflect.TypeOf(funcExpr.Param[0]).String())
	}

	var val interface{}
	switch v := funcExpr.Param[1].(type) {
	case *parser.StringLiteral:
		val = v.Val

	case *parser.NumberLiteral:
		if v.IsInt {
			val = v.Int
		} else {
			val = v.Float
		}

	case *parser.BoolLiteral:
		val = v.Val

	case *parser.NilLiteral:
		val = nil
	}

	cont, err := p.getContent(key)
	if err != nil {
		l.Warnf("key `%v' not exist, ignored", key)
		return p, nil //nolint:nilerr
	}

	// todo key string
	if reflect.DeepEqual(cont, val) {
		var k string

		switch t := key.(type) {
		case *parser.Identifier:
			k = t.String()
		case *parser.AttrExpr:
			k = t.String()
		case *parser.StringLiteral:
			k = t.Val
		default:
			l.Warnf("unsupported %v get", reflect.TypeOf(key).String())
			return p, nil
		}

		delete(p.Output, k)
	}

	return p, nil
}

func Dropkey(p *Pipeline, node parser.Node) (*Pipeline, error) {
	funcExpr := fexpr(node)
	if len(funcExpr.Param) != 1 {
		return p, fmt.Errorf("func %s expected 1 args", funcExpr.Name)
	}

	var key parser.Node
	switch v := funcExpr.Param[0].(type) {
	case *parser.Identifier, *parser.AttrExpr:
		key = v
	default:
		return p, fmt.Errorf("param key expect Identifier or AttrExpr, got %s",
			reflect.TypeOf(funcExpr.Param[0]).String())
	}

	delete(p.Output, key.String())

	return p, nil
}

func DropOriginData(p *Pipeline, node parser.Node) (*Pipeline, error) {
	delete(p.Output, "message")
	return p, nil
}

func Addkey(p *Pipeline, node parser.Node) (*Pipeline, error) {
	funcExpr := fexpr(node)
	if len(funcExpr.Param) != 2 {
		return p, fmt.Errorf("func %s expected 1 args", funcExpr.Name)
	}

	var key parser.Node
	switch v := funcExpr.Param[0].(type) {
	case *parser.Identifier, *parser.AttrExpr:
		key = v
	default:
		return p, fmt.Errorf("param key expect Identifier or AttrExpr, got %s",
			reflect.TypeOf(funcExpr.Param[0]).String())
	}

	var val interface{}
	switch v := funcExpr.Param[1].(type) {
	case *parser.StringLiteral:
		val = v.Val

	case *parser.NumberLiteral:
		if v.IsInt {
			val = v.Int
		} else {
			val = v.Float
		}

	case *parser.BoolLiteral:
		val = v.Val

	case *parser.NilLiteral:
		val = nil
	}

	if err := p.setContent(key, val); err != nil {
		l.Warn(err)
		return p, nil
	}

	return p, nil
}

func Dz(p *Pipeline, node parser.Node) (*Pipeline, error) {
	funcExpr := fexpr(node)
	if len(funcExpr.Param) != 2 {
		return p, fmt.Errorf("func %s expected 2 args", funcExpr.Name)
	}

	set := arglist(funcExpr, 1)

	var key parser.Node
	switch v := funcExpr.Param[0].(type) {
	case *parser.AttrExpr, *parser.Identifier:
		key = v
	default:
		return p, fmt.Errorf("param key expect AttrExpr or Identifier, got %s",
			reflect.TypeOf(funcExpr.Param[0]).String())
	}

	var start, end int

	if len(set) != 2 {
		return p, fmt.Errorf("param between range value `%v' is not expected", set)
	}

	if v, ok := set[0].(*parser.NumberLiteral); !ok {
		return p, fmt.Errorf("range value `%v' is not expected", set)
	} else if v.IsInt {
		start = int(v.Int)
	}

	if v, ok := set[1].(*parser.NumberLiteral); !ok {
		return p, fmt.Errorf("range value `%v' is not expected", set)
	} else {
		if v.IsInt {
			end = int(v.Int)
		}

		if start > end {
			return p, fmt.Errorf("range value start %v must le end %v", start, end)
		}
	}

	cont, err := p.getContentStr(key)
	if err != nil {
		l.Warnf("key `%v' not exist, ignored", key)
		return p, nil //nolint:nilerr
	}

	if end > utf8.RuneCountInString(cont) {
		end = utf8.RuneCountInString(cont)
	}

	if start <= 0 {
		start = 1
	}

	arrCont := []rune(cont)

	for i := 0; i < len(arrCont); i++ {
		if i+1 >= start && i < end {
			if unicode.Is(unicode.Han, arrCont[i]) {
				arrCont[i] = rune('＊')
			} else {
				arrCont[i] = rune('*')
			}
		}
	}

	if err := p.setContent(key, string(arrCont)); err != nil {
		l.Warn(err)
		return p, nil
	}

	return p, nil
}

func Replace(p *Pipeline, node parser.Node) (*Pipeline, error) {
	funcExpr := fexpr(node)

	if len(funcExpr.Param) != 3 {
		return p, fmt.Errorf("func %s expected 3 args", funcExpr.Name)
	}

	var key parser.Node
	var pattern, dz string
	switch v := funcExpr.Param[0].(type) {
	case *parser.AttrExpr, *parser.Identifier:
		key = v
	default:
		return p, fmt.Errorf("param key expect AttrExpr or Identifier, got %s",
			reflect.TypeOf(funcExpr.Param[0]).String())
	}

	switch v := funcExpr.Param[1].(type) {
	case *parser.StringLiteral:
		pattern = v.Val
	default:
		return p, fmt.Errorf("expect StringLiteral, got %s",
			reflect.TypeOf(funcExpr.Param[1]).String())
	}

	switch v := funcExpr.Param[2].(type) {
	case *parser.StringLiteral:
		dz = v.Val
	default:
		return p, fmt.Errorf("expect StringLiteral, got %s",
			reflect.TypeOf(funcExpr.Param[2]).String())
	}

	reg, err := regexp.Compile(pattern)
	if err != nil {
		return p, fmt.Errorf("regular expression %s parse err: %w",
			reflect.TypeOf(funcExpr.Param[1]).String(), err)
	}

	cont, err := p.getContentStr(key)
	if err != nil {
		l.Warnf("key `%v' not exist, ignored", key)
		return p, nil //nolint:nilerr
	}

	newCont := reg.ReplaceAllString(cont, dz)
	if err := p.setContent(key, newCont); err != nil {
		l.Warn(err)
		return p, nil
	}

	return p, nil
}
