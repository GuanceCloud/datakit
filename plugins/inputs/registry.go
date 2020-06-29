package inputs

type Input interface {
	Catalog() string
	Run()
	SampleConfig() string

	// Status() string
	// TotalBytes() int64

	// add more...
}

type Creator func() Input

var (
	Inputs             = map[string]Creator{}
	InternalInputsData = map[string][]byte{}
)

func Add(name string, creator Creator) {
	Inputs[name] = creator
}
