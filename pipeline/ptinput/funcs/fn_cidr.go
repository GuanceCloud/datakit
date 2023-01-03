// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package funcs

import (
	"fmt"
	"net/netip"
	"reflect"

	"github.com/GuanceCloud/platypus/pkg/ast"
	"github.com/GuanceCloud/platypus/pkg/engine/runtime"
	"github.com/GuanceCloud/platypus/pkg/errchain"
)

func CIDRChecking(ctx *runtime.Context, funcExpr *ast.CallExpr) *errchain.PlError {
	if len(funcExpr.Param) != 2 {
		return runtime.NewRunError(ctx, fmt.Sprintf(
			"func %s expects 2 args", funcExpr.Name), funcExpr.NamePos)
	}

	param0 := funcExpr.Param[0]
	if !isPlVarbOrFunc(param0) && param0.NodeType != ast.TypeStringLiteral {
		return runtime.NewRunError(ctx, fmt.Sprintf(
			"expect StringLiteral, Identifier or AttrExpr, got %s",
			param0.NodeType), param0.StartPos())
	}

	param1 := funcExpr.Param[1]
	if !isPlVarbOrFunc(param1) && param1.NodeType != ast.TypeStringLiteral {
		return runtime.NewRunError(ctx, fmt.Sprintf(
			"expect StringLiteral, Identifier or AttrExpr, got %s",
			param1.NodeType), param1.StartPos())
	}

	return nil
}

func CIDR(ctx *runtime.Context, funcExpr *ast.CallExpr) *errchain.PlError {
	if len(funcExpr.Param) != 2 {
		err := fmt.Errorf("func %s expects 2 args", funcExpr.Name)
		l.Debug(err)
		ctx.Regs.ReturnAppend(false, ast.Bool)
		return nil
	}

	ipAddr, err := getStr(ctx, funcExpr.Param[0])
	if err != nil {
		l.Debug(err)
		ctx.Regs.ReturnAppend(false, ast.Bool)
		return nil
	}

	prefix, err := getStr(ctx, funcExpr.Param[1])
	if err != nil {
		l.Debug(err)
		ctx.Regs.ReturnAppend(false, ast.Bool)
		return nil
	}

	ok, err := CIDRContains(ipAddr, prefix)
	if err != nil {
		l.Debug(err)
	}

	ctx.Regs.ReturnAppend(ok, ast.Bool)
	return nil
}

func CIDRContains(ipAddr, prefix string) (bool, error) {
	network, err := netip.ParsePrefix(prefix)
	if err != nil {
		return false, err
	}

	ip, err := netip.ParseAddr(ipAddr)
	if err != nil {
		return false, err
	}

	return network.Contains(ip), nil
}

func getStr(ctx *runtime.Context, node *ast.Node) (string, error) {
	if node.NodeType == ast.TypeStringLiteral {
		return node.StringLiteral.Val, nil
	}

	keyName, err := getKeyName(node)
	if err != nil {
		return "", err
	}
	v, err := ctx.GetKey(keyName)
	if err != nil {
		return "", err
	}
	if v.DType != ast.String {
		return "", err
	}

	if addr, ok := v.Value.(string); ok {
		return addr, nil
	} else {
		return "", fmt.Errorf("type: %s", reflect.TypeOf(v.Value).String())
	}
}
