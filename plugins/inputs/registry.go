package inputs

import (
	"github.com/influxdata/telegraf"
)

type Input interface {
	telegraf.Input

	Catalog() string
	//Status() string

	/* TotalBytes() int64 */

	// add more...
}

//type Creator func() telegraf.Input
type Creator func() Input

var Inputs = map[string]Creator{}

var InternalInputsData = map[string][]byte{}

func Add(name string, creator Creator) {
	Inputs[name] = creator
}
