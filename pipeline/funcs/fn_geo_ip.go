package funcs

import (
	"fmt"
	"reflect"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/ipdb"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/parser"
)

var ipdbInstance ipdb.IPdb

var geoDefaultVal = "unknown"

func Geo(ip string) (*ipdb.IPdbRecord, error) {
	if ipdbInstance != nil {
		return ipdbInstance.Geo(ip)
	} else {
		return &ipdb.IPdbRecord{}, nil
	}
}

func InitIPdb(instance ipdb.IPdb) {
	ipdbInstance = instance
}

func GeoIPChecking(node parser.Node) error {
	funcExpr := fexpr(node)
	if len(funcExpr.Param) != 1 {
		return fmt.Errorf("func `%s' expected 1 args", funcExpr.Name)
	}

	switch funcExpr.Param[0].(type) {
	case *parser.AttrExpr, *parser.Identifier:
	default:
		return fmt.Errorf("param key expect AttrExpr or Identifier, got %s",
			reflect.TypeOf(funcExpr.Param[0]).String())
	}
	return nil
}

func GeoIP(ng *parser.Engine, node parser.Node) error {
	funcExpr := fexpr(node)
	if len(funcExpr.Param) != 1 {
		return fmt.Errorf("func `%s' expected 1 args", funcExpr.Name)
	}

	var key parser.Node
	switch v := funcExpr.Param[0].(type) {
	case *parser.AttrExpr, *parser.Identifier:
		key = v
	default:
		return fmt.Errorf("param key expect AttrExpr or Identifier, got %s",
			reflect.TypeOf(funcExpr.Param[0]).String())
	}

	cont, err := ng.GetContentStr(key)
	if err != nil {
		l.Debugf("key `%v' not exist, ignored", key)
		return nil //nolint:nilerr
	}

	if dic, err := GeoIPHandle(cont); err != nil {
		l.Debugf("GeoIPHandle: %s, ignored", err)
		return err
	} else {
		for k, v := range dic {
			if err := ng.SetContent(k, v); err != nil {
				l.Warn(err)
				return nil
			}
		}
	}

	return nil
}
