package quantity

var (
	UnknownKind = &Kind{}
	Count       = &Kind{}
	Duration    = &Kind{}
	Memory      = &Kind{}
)

func init() {
	// avoid loop reference
	UnknownKind.DefaultUnit = UnknownUnit
	Count.DefaultUnit = CountUnit
	Duration.DefaultUnit = MicroSecond
	Memory.DefaultUnit = Byte
}

// Kind 度量类型，数量，时间，内存...
type Kind struct {
	DefaultUnit *Unit
}
