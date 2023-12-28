//go:build linux
// +build linux

package l4log

import (
	"fmt"
	"net"
	"strings"
	"sync"

	"github.com/GuanceCloud/platypus/pkg/ast"
	"github.com/GuanceCloud/platypus/pkg/parser"
)

var _ipnetpool = struct {
	sync.RWMutex
	pool map[string]*net.IPNet
}{
	pool: make(map[string]*net.IPNet),
}

var _fnList = &fnGroup{}

var _ = _fnList.registerFuncs(&fnDesc{
	name: "ipnet_contains",
	params: []fnParam{
		{name: "ipnet", typ: ast.String},
		{name: "ip", typ: ast.String},
	},
	ret: ast.Bool,
	fn:  ipnetContains,
})

var _ = _fnList.registerFuncs(&fnDesc{
	name: "has_prefix",
	params: []fnParam{
		{name: "s", typ: ast.String},
		{name: "prefix", typ: ast.String},
	},
	ret: ast.Bool,
	fn:  hasPrefix,
})

type fnGroup struct {
	fnList []*fnDesc
}

func (fnG *fnGroup) getFnDesc(name string) *fnDesc {
	for _, fn := range fnG.fnList {
		if fn.name == name {
			return fn
		}
	}
	return nil
}

func (fnG *fnGroup) registerFuncs(fn *fnDesc) bool {
	if fn == nil {
		return false
	}

	for _, v := range fnG.fnList {
		if v.name == fn.name {
			return false
		}
	}

	fnG.fnList = append(fnG.fnList, fn)
	return true
}

type fnParam struct {
	name string
	typ  ast.DType
}

type fnDesc struct {
	name   string
	params []fnParam
	// optParams []fnParam

	ret ast.DType

	fn func(...any) any
}

var _ netElem = (*netParams)(nil)

type netParams struct {
	tcp  bool
	ipv4 bool

	// hostNet bool
	// netNS string

	ipSAddr string
	ipDAddr string

	ip6SAddr string
	ip6DAddr string

	k8sSrcPod string
	k8sDstPod string

	sPort int64
	dPort int64
}

func (p *netParams) get(key string) (ast.DType, any) {
	switch key {
	case "tcp":
		return ast.Bool, p.tcp
	case "udp":
		return ast.Bool, !p.tcp
	case "ipv4":
		return ast.Bool, p.ipv4
	case "ipv6":
		return ast.Bool, !p.ipv4

	case "src_port":
		return ast.Int, p.sPort
	case "dst_port":
		return ast.Int, p.dPort
	case "ip_saddr":
		return ast.String, p.ipSAddr
	case "ip_daddr":
		return ast.String, p.ipDAddr
	case "ip6_saddr":
		return ast.String, p.ip6SAddr
	case "ip6_daddr":
		return ast.String, p.ip6DAddr
	case "k8s_src_pod":
		return ast.String, p.k8sSrcPod
	case "k8s_dst_pod":
		return ast.String, p.k8sDstPod

	default:
		return ast.Invalid, nil
	}
}

type netElem interface {
	get(key string) (ast.DType, any)
}

func parseFilter(s string) (ast.Stmts, error) {
	a, err := parser.ParsePipeline("rule", s)
	if err != nil {
		return nil, err
	}
	return a, nil
}

type filterRuntime struct {
	fnG *fnGroup
}

func (g *filterRuntime) checkStmts(stmts ast.Stmts, elem netElem) error {
	for _, s := range stmts {
		typ, err := g._checkStmt(s, elem)
		if err != nil {
			return err
		}
		if typ != ast.Bool {
			return fmt.Errorf("rule not return bool value: %s", s.String())
		}
	}
	return nil
}

func (g *filterRuntime) runNetFilterDrop(stmts ast.Stmts, elem netElem) bool {
	if elem == nil {
		return false
	}

	for _, s := range stmts {
		if dt, val, err := g._runStmt(s, elem); err != nil {
			log.Error(err)
		} else {
			if dt != ast.Bool {
				continue
			}
			if v, ok := val.(bool); ok && v {
				return true
			}
		}
	}
	return false
}

func (g *filterRuntime) _runStmt(node *ast.Node, elem netElem) (ast.DType, any, error) {
	switch node.NodeType { //nolint:exhaustive
	case ast.TypeConditionalExpr:
		expr := node.ConditionalExpr
		typL, valL, err := g._runStmt(expr.LHS, elem)
		if err != nil {
			return ast.Invalid, nil, err
		}

		if typL == ast.Bool {
			if v, ok := valL.(bool); ok {
				if v {
					if expr.Op == ast.OR {
						return ast.Bool, true, nil
					}
				} else {
					if expr.Op == ast.AND {
						return ast.Bool, false, nil
					}
				}
			}
		}

		typR, valR, err := g._runStmt(expr.RHS, elem)
		if err != nil {
			return ast.Invalid, nil, err
		}

		switch expr.Op { //nolint:exhaustive
		case ast.AND, ast.OR:
			if typR != ast.Bool {
				return ast.Invalid, nil, fmt.Errorf("%d%d: should be bool: %s", expr.OpPos.Ln, expr.OpPos.Col, typR)
			}
			return ast.Bool, valR.(bool), nil
		case ast.EQEQ, ast.NEQ:
			if typL != typR {
				return ast.Invalid, nil, fmt.Errorf("%d%d: should be same type: %s", expr.OpPos.Ln, expr.OpPos.Col, typL)
			}

			var retVal bool
			switch typL { //nolint:exhaustive
			case ast.String:
				if valL.(string) == valR.(string) {
					retVal = true
				}
			case ast.Int:
				if valL.(int64) == valR.(int64) {
					retVal = true
				}
			case ast.Bool:
				if valL.(bool) == valR.(bool) {
					retVal = true
				}
			default:
				return ast.Invalid, nil, fmt.Errorf("%d%d: unsupport type: %s", expr.OpPos.Ln, expr.OpPos.Col, typL)
			}

			if expr.Op == ast.NEQ {
				retVal = !retVal
			}

			return ast.Bool, retVal, nil

		case ast.LT, ast.LTE,
			ast.GT, ast.GTE:
			switch expr.Op { //nolint:exhaustive
			case ast.LT:
				if valL.(int64) < valR.(int64) {
					return ast.Bool, true, nil
				}
			case ast.LTE:
				if valL.(int64) <= valR.(int64) {
					return ast.Bool, true, nil
				}
			case ast.GT:
				if valL.(int64) > valR.(int64) {
					return ast.Bool, true, nil
				}
			case ast.GTE:
				if valL.(int64) >= valR.(int64) {
					return ast.Bool, true, nil
				}
			}
		}

		return ast.Bool, false, nil
	case ast.TypeUnaryExpr:
		expr := node.UnaryExpr
		_, val, err := g._runStmt(expr.RHS, elem)
		if err != nil {
			return ast.Invalid, nil, err
		}
		return ast.Bool, !val.(bool), nil
	case ast.TypeCallExpr:
		expr := node.CallExpr

		if g.fnG == nil {
			return ast.Invalid, nil, fmt.Errorf("no function group")
		}

		fnD := g.fnG.getFnDesc(expr.Name)
		p := make([]any, 0, len(fnD.params))
		for _, v := range expr.Param {
			_, val, err := g._runStmt(v, elem)
			if err != nil {
				return ast.Invalid, nil, err
			}
			p = append(p, val)
		}
		return fnD.ret, fnD.fn(p...), nil
	case ast.TypeParenExpr:
		return g._runStmt(node.ParenExpr.Param, elem)
	case ast.TypeIdentifier:
		expr := node.Identifier
		dtyp, val := elem.get(expr.Name)
		return dtyp, val, nil
	case ast.TypeStringLiteral:
		return ast.String, node.StringLiteral.Val, nil
	case ast.TypeIntegerLiteral:
		return ast.Int, node.IntegerLiteral.Val, nil
	case ast.TypeBoolLiteral:
		return ast.Bool, node.BoolLiteral.Val, nil
	default:
		return ast.Invalid, nil, fmt.Errorf("unsupport type: %s", node.NodeType)
	}
}

func (g *filterRuntime) _checkStmt(node *ast.Node, elem netElem) (ast.DType, error) {
	switch node.NodeType { //nolint:exhaustive
	case ast.TypeConditionalExpr:
		expr := node.ConditionalExpr
		switch expr.Op { //nolint:exhaustive
		case ast.AND, ast.OR,
			ast.EQEQ, ast.NEQ,
			ast.LT, ast.LTE,
			ast.GT, ast.GTE:
		default:
			return ast.Invalid, fmt.Errorf("%d%d: unsupport operator: %s", expr.OpPos.Ln, expr.OpPos.Col, expr.Op)
		}

		typL, err := g._checkStmt(expr.LHS, elem)
		if err != nil {
			return ast.Invalid, err
		}
		typR, err := g._checkStmt(expr.RHS, elem)
		if err != nil {
			return ast.Invalid, err
		}

		switch expr.Op { //nolint:exhaustive
		case ast.AND, ast.OR:
			if typL != ast.Bool {
				return ast.Invalid, fmt.Errorf("%d%d: should be bool: %s", expr.OpPos.Ln, expr.OpPos.Col, typL)
			}
			if typR != ast.Bool {
				return ast.Invalid, fmt.Errorf("%d%d: should be bool: %s", expr.OpPos.Ln, expr.OpPos.Col, typR)
			}
		case ast.EQEQ, ast.NEQ:
			if typL != typR {
				return ast.Invalid, fmt.Errorf("%d%d: should be same type: %s", expr.OpPos.Ln, expr.OpPos.Col, typL)
			}
		case ast.LT, ast.LTE,
			ast.GT, ast.GTE:
			if typL != ast.Int {
				return ast.Invalid, fmt.Errorf("%d%d: should be int: %s", expr.OpPos.Ln, expr.OpPos.Col, typL)
			}
			if typR != ast.Int {
				return ast.Invalid, fmt.Errorf("%d%d: should be int: %s", expr.OpPos.Ln, expr.OpPos.Col, typR)
			}
		}
		return ast.Bool, nil
	case ast.TypeUnaryExpr:
		expr := node.UnaryExpr
		if expr.Op != ast.NOT {
			return ast.Invalid, fmt.Errorf("%d%d: unsupport operator: %s", expr.OpPos.Ln, expr.OpPos.Col, expr.Op)
		}
		typ, err := g._checkStmt(expr.RHS, elem)
		if err != nil {
			return ast.Invalid, err
		}
		if typ != ast.Bool {
			return ast.Invalid, fmt.Errorf("%d%d: unsupport type: %s", expr.OpPos.Ln, expr.OpPos.Col, typ)
		}
		return ast.Bool, nil
	case ast.TypeCallExpr:
		expr := node.CallExpr
		if g.fnG == nil {
			return ast.Invalid, fmt.Errorf("%d%d: unsupport function: %s", expr.NamePos.Ln, expr.NamePos.Col, expr.Name)
		}
		fnD := g.fnG.getFnDesc(expr.Name)
		if fnD == nil || fnD.fn == nil {
			return ast.Invalid, fmt.Errorf("%d%d: unsupport function: %s", expr.NamePos.Ln, expr.NamePos.Col, expr.Name)
		}

		if len(expr.Param) != len(fnD.params) {
			return ast.Invalid, fmt.Errorf("%d%d: param count not match: %s", expr.NamePos.Ln, expr.NamePos.Col, expr.Name)
		}

		for i, param := range expr.Param {
			dtyp, err := g._checkStmt(param, elem)
			if err != nil {
				return ast.Invalid, err
			}
			if dtyp != fnD.params[i].typ {
				return ast.Invalid, fmt.Errorf("%d%d: param type not match, expected: %s, actual: %s",
					expr.NamePos.Ln, expr.NamePos.Col, fnD.params[i].typ, dtyp)
			}

			switch dtyp { //nolint:exhaustive
			case ast.Int, ast.Bool, ast.String:
			default:
				return ast.Invalid, fmt.Errorf("%d%d: param %s unsupport data type: %s",
					expr.NamePos.Ln, expr.NamePos.Col, expr.Name, dtyp)
			}
		}
		return fnD.ret, nil
	case ast.TypeParenExpr:
		return g._checkStmt(node.ParenExpr.Param, elem)
	case ast.TypeIdentifier:
		expr := node.Identifier
		if v, _ := elem.get(expr.Name); v != ast.Invalid {
			switch v { //nolint:exhaustive
			case ast.Int, ast.Bool, ast.String:
			default:
				return ast.Invalid, fmt.Errorf("%d%d: param %s unsupport data type: %s",
					expr.Start.Ln, expr.Start.Col, expr.Name, v)
			}
			return v, nil
		} else {
			return ast.Invalid, fmt.Errorf("%d%d: unsupport identifier: %s", expr.Start.Ln, expr.Start.Col, expr.Name)
		}
	case ast.TypeStringLiteral:
		return ast.String, nil
	case ast.TypeIntegerLiteral:
		return ast.Int, nil
	case ast.TypeBoolLiteral:
		return ast.Bool, nil
	default:
		return ast.Invalid, fmt.Errorf("unsupport type: %s", node.NodeType)
	}
}

type DataFilters []DataFiler

type DataFiler struct {
	K8sDeployment string
	K8sService    string
	K8sPod        string
	NICName       string
	SrcIPWithMask string
	SrcPort       uint16
	DstIPWithMask string
	DstPort       uint16
}

func ipnetContains(v ...any) any {
	var ipnet, ip string
	if len(v) == 2 {
		if v, ok := v[0].(string); ok {
			ipnet = v
		}
		if v, ok := v[1].(string); ok {
			ip = v
		}
	}

	_ipnetpool.RLock()
	ipnet2, ok := _ipnetpool.pool[ipnet]
	if ok {
		_ipnetpool.RUnlock()
	} else {
		_ipnetpool.RUnlock()

		_ipnetpool.Lock()
		_, ipnet2, err := net.ParseCIDR(ipnet)
		if err != nil {
			_ipnetpool.Unlock()
			return false
		}
		_ipnetpool.pool[ipnet] = ipnet2
		_ipnetpool.Unlock()
	}

	if ipnet2 != nil {
		return ipnet2.Contains(net.ParseIP(ip))
	} else {
		return false
	}
}

func hasPrefix(v ...any) any {
	var str, prefix string
	if len(v) == 2 {
		if v, ok := v[0].(string); ok {
			str = v
		}
		if v, ok := v[1].(string); ok {
			prefix = v
		}
	}
	return strings.HasPrefix(str, prefix)
}
