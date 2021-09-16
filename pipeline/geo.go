package pipeline

import (
	"fmt"

	"github.com/ip2location/ip2location-go"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

var (
	defaultGEO ipgeo
)

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
