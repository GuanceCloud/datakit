package quantity

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

var (
	MemoryDisplayName     = "digital"
	MemoryDefaultUnitFlag = "B"

	DurationDisplayName     = "time"
	DurationDefaultUnitFlag = "μs"
)

var UnknownUnit = &Unit{Kind: UnknownKind, Name: "unknown"}

var (
	NanoSecond  = &Unit{Kind: Duration, Base: int64(time.Nanosecond), Name: "ns"}
	MicroSecond = &Unit{Kind: Duration, Base: int64(time.Microsecond), Name: "us"}
	MilliSecond = &Unit{Kind: Duration, Base: int64(time.Millisecond), Name: "ms"}
	Second      = &Unit{Kind: Duration, Base: int64(time.Second), Name: "s"}
	Minute      = &Unit{Kind: Duration, Base: int64(time.Minute), Name: "m"}
	Hour        = &Unit{Kind: Duration, Base: int64(time.Hour), Name: "h"}
)

var (
	CountUnit = &Unit{Kind: Count, Base: 1, Name: ""}
)

var (
	Byte     = &Unit{Kind: Memory, Base: 1, Name: "Bytes"}
	KiloByte = &Unit{Kind: Memory, Base: 1 << 10, Name: "KB"}
	MegaByte = &Unit{Kind: Memory, Base: 1 << 20, Name: "MB"}
	GigaByte = &Unit{Kind: Memory, Base: 1 << 30, Name: "GB"}
	TeraByte = &Unit{Kind: Memory, Base: 1 << 40, Name: "TB"}
	PetaByte = &Unit{Kind: Memory, Base: 1 << 50, Name: "PB"}
)

var (
	memoryUnitMaps = map[string]*Unit{
		"bytes": Byte,
		"byte":  Byte,
		"b":     Byte,

		"kilobytes": KiloByte,
		"kilobyte":  KiloByte, // 1000
		"kibibytes": KiloByte,
		"kibibyte":  KiloByte, // 1024
		"kib":       KiloByte,
		"kb":        KiloByte,

		"megabytes": MegaByte,
		"megabyte":  MegaByte,
		"megibytes": MegaByte,
		"megibyte":  MegaByte,
		"mib":       MegaByte,
		"mb":        MegaByte,

		"gigabytes": GigaByte,
		"gigabyte":  GigaByte,
		"gigibytes": GigaByte,
		"gigibyte":  GigaByte,
		"gib":       GigaByte,
		"gb":        GigaByte,

		"terabytes":     TeraByte,
		"terabyte":      TeraByte,
		"tebibytes":     TeraByte,
		"tebibyte":      TeraByte,
		"trillionbytes": TeraByte,
		"trillionbyte":  TeraByte,
		"tib":           TeraByte,
		"tb":            TeraByte,

		"petabytes": PetaByte,
		"petabyte":  PetaByte,
		"pebibytes": PetaByte,
		"pebibyte":  PetaByte,
		"pib":       PetaByte,
		"pb":        PetaByte,
	}

	durationUnitMaps = map[string]*Unit{
		"nanoseconds": NanoSecond,
		"nanosecond":  NanoSecond,
		"ns":          NanoSecond,

		"microseconds": MicroSecond,
		"microsecond":  MicroSecond,
		"us":           MicroSecond,
		"µs":           MicroSecond,

		"milliseconds": MilliSecond,
		"millisecond":  MilliSecond,
		"ms":           MilliSecond,

		"seconds": Second,
		"second":  Second,
		"s":       Second,

		"minutes": Minute,
		"minute":  Minute,
		"m":       Minute,

		"hours": Hour,
		"hour":  Hour,
		"h":     Hour,
	}
)

type Unit struct {
	Kind *Kind
	Base int64
	Name string
}

func (u *Unit) String() string {
	return u.Name
}

// ConvertTo convert value for Unit u to Unit target
func (u *Unit) ConvertTo(target *Unit, value int64) (int64, error) {
	if u.Kind != target.Kind {
		return 0, fmt.Errorf("unit kinds are not compatiable")
	}
	return value * u.Base / target.Base, nil
}

// ConvertToDefaultUnit convert value in Unit u to Unit Kind's default unit
func (u *Unit) ConvertToDefaultUnit(value int64) int64 {
	v, _ := u.ConvertTo(u.Kind.DefaultUnit, value)
	return v
}

func (u *Unit) Quantity(v int64) *Quantity {
	return &Quantity{
		Unit:  u,
		Value: v,
	}
}

func (u *Unit) MarshalJSON() ([]byte, error) {
	var flags []string

	switch u.Kind {
	case Memory:
		flags = []string{MemoryDisplayName, MemoryDefaultUnitFlag}
	case Duration:
		flags = []string{DurationDisplayName, DurationDefaultUnitFlag}
	default:
		flags = make([]string, 0)
	}
	return json.Marshal(flags)
}

func parseDuration(s string) (*Unit, error) {
	unit := strings.ToLower(strings.TrimSpace(s))
	if u, ok := durationUnitMaps[unit]; ok {
		return u, nil
	}
	switch {
	case strings.HasPrefix(unit, "na"): // nano
		return NanoSecond, nil
	case strings.HasPrefix(unit, "mic"): // micro
		return MicroSecond, nil
	case strings.HasPrefix(unit, "mil"): // milli
		return MilliSecond, nil
	case strings.HasPrefix(unit, "min"): // minute
		return Minute, nil
	case strings.HasPrefix(unit, "hour"): // hour
		return Hour, nil
	case strings.HasPrefix(unit, "sec"): // second
		return Second, nil
	}
	return nil, fmt.Errorf("can not resolve duration unit from [%s]", s)
}

func parseMemoryUnit(s string) (*Unit, error) {
	unit := strings.ToLower(strings.TrimSpace(s))
	if u, ok := memoryUnitMaps[unit]; ok {
		return u, nil
	}
	switch {
	case strings.HasPrefix(unit, "k"):
		return KiloByte, nil
	case strings.HasPrefix(unit, "m"):
		return MegaByte, nil
	case strings.HasPrefix(unit, "g"):
		return GigaByte, nil
	case strings.HasPrefix(unit, "t"):
		return TeraByte, nil
	case strings.HasPrefix(unit, "p"):
		return PetaByte, nil
	case strings.HasPrefix(unit, "byte"):
		return Byte, nil
	}
	return nil, fmt.Errorf("can not resolve memory unit from [%s]", s)
}

func ParseUnit(kind *Kind, s string) (*Unit, error) {
	switch kind {
	case Count:
		return CountUnit, nil
	case Memory:
		return parseMemoryUnit(s)
	case Duration:
		return parseDuration(s)
	}
	return nil, fmt.Errorf("unknown quantity kind")
}
