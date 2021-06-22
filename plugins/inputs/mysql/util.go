package mysql

import (
	"database/sql"
	"github.com/spf13/cast"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type ConversionFunc func(value sql.RawBytes) (interface{}, error)

func Conv(val interface{}, Datatype string) (interface{}, error) {
	var (
		res interface{}
		err error
	)

	switch Datatype {
	case inputs.Float:
		res, err = cast.ToFloat64E(val)
	case inputs.Int:
		res, err = cast.ToInt64E(val)
	case inputs.Bool:
		res, err = cast.ToBoolE(val)
	case inputs.String:
		res, err = cast.ToStringE(val)
	}

	return res, err
}
