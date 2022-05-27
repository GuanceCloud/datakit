// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package funcs

import (
	"fmt"
	"reflect"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/obfuscate"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/parser"
)

func SQLCoverChecking(ng *parser.EngineData, node parser.Node) error {
	funcExpr := fexpr(node)
	if len(funcExpr.Param) != 1 {
		return fmt.Errorf("func %s expects 1 args", funcExpr.Name)
	}

	switch funcExpr.Param[0].(type) {
	case *parser.AttrExpr, *parser.Identifier:
	default:
		return fmt.Errorf("param key expects AttrExpr or Identifier, got %s",
			reflect.TypeOf(funcExpr.Param[0]).String())
	}
	return nil
}

func SQLCover(ng *parser.EngineData, node parser.Node) interface{} {
	o := obfuscate.NewObfuscator(&obfuscate.Config{})
	funcExpr := fexpr(node)
	if len(funcExpr.Param) != 1 {
		return fmt.Errorf("func %s expects 1 args", funcExpr.Name)
	}

	var key parser.Node
	switch v := funcExpr.Param[0].(type) {
	case *parser.AttrExpr, *parser.Identifier:
		key = v
	default:
		return fmt.Errorf("param key expects AttrExpr or Identifier, got %s",
			reflect.TypeOf(funcExpr.Param[0]).String())
	}

	cont, err := ng.GetContentStr(key)
	if err != nil {
		l.Debugf("key `%v' not exist, ignored", key)
		return nil //nolint:nilerr
	}

	Type := "sql"

	v, err := obfuscatedResource(o, Type, cont)
	if err != nil {
		l.Warn(err)
		return nil
	}

	if err := ng.SetContent(key, v); err != nil {
		l.Warn(err)
		return nil
	}

	return nil
}

func obfuscatedResource(o *obfuscate.Obfuscator, typ, resource string) (string, error) {
	if typ != "sql" {
		return resource, nil
	}
	oq, err := o.ObfuscateSQLString(resource)
	if err != nil {
		l.Errorf("Error obfuscating stats group resource %q: %w", resource, err)
		return "", err
	}
	return oq.Query, nil
}
