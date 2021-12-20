package funcs

import (
	"fmt"
	"reflect"

	"github.com/ip2location/ip2location-go"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/parser"
)

var defaultGEO ipgeo

const (
	ErrInvalidIP = "Invalid IP address"
	ErrInvalidDB = "Invalid database file"

	geoDefaultVal = "unknown"
)

type ipdb struct {
	db *ip2location.DB
}

func (g *ipdb) Get(ip string) (*ip2location.IP2Locationrecord, error) {
	if g.db == nil {
		return nil, fmt.Errorf("GEO DB not set")
	}

	r, err := g.db.Get_all(ip)
	if err != nil {
		return nil, err
	}

	return &r, nil
}

type ipgeo interface {
	Get(ip string) (*ip2location.IP2Locationrecord, error)
}

func Geo(ip string) (*ip2location.IP2Locationrecord, error) {
	r, err := defaultGEO.Get(ip)
	if err != nil {
		return nil, err
	}

	switch r.City {
	case ErrInvalidIP, ErrInvalidDB:
		r.City = geoDefaultVal
	}

	switch r.Region {
	case ErrInvalidIP, ErrInvalidDB:
		r.Region = geoDefaultVal
	}

	switch r.Country_short {
	case ErrInvalidIP, ErrInvalidDB:
		r.Country_short = geoDefaultVal
	}

	return r, err
}

func LoadIPLib(f string) error {
	if !datakit.FileExist(f) {
		l.Warnf("%v not found", f)
		return nil
	}

	db, err := ip2location.OpenDB(f)
	if err != nil {
		return err
	}

	defaultGEO = &ipdb{db: db}

	return nil
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
		l.Warnf("key `%v' not exist, ignored", key)
		return nil //nolint:nilerr
	}

	if dic, err := GeoIPHandle(cont); err != nil {
		l.Warnf("GeoIPHandle: %s, ignored", err)
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
