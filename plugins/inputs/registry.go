package inputs

import (
	"fmt"
)

type Input interface {
	Catalog() string
	Run()
	SampleConfig() string
	//Stop() error

	// add more...
}

type Creator func() Input

var (
	Inputs = map[string]Creator{}
)

func Add(name string, creator Creator) {
	if _, ok := Inputs[name]; ok {
		panic(fmt.Sprintf("inputs %s exist", name))
	}

	Inputs[name] = creator
}
