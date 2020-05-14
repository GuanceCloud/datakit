package inputs

import (
	"github.com/influxdata/telegraf"
)

type DataKitInput interface {
	telegraf.Input

	CollectionName() string
	Status() string

	/* TotalBytes() int64 */

	// add more...
}

type Creator func() telegraf.Input

var Inputs = map[string]Creator{}

var InternalInputsData = map[string][]byte{}

func Add(name string, creator Creator) {
	Inputs[name] = creator
}
