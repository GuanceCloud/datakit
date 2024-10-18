// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package quantity

import (
	"fmt"
	"time"

	"github.com/GuanceCloud/cliutils/pprofparser/tools/mathtoolkit"
)

type Quantity struct {
	Value int64
	Unit  *Unit
}

func (q *Quantity) MarshalJSON() ([]byte, error) {
	if q == nil {
		return []byte("null"), nil
	}
	return []byte(fmt.Sprintf("\"%s\"", q.String())), nil
}

func (q *Quantity) SwitchToDefaultUnit() {
	if q == nil || q.Unit == nil {
		return
	}
	if q.Unit != q.Unit.Kind.DefaultUnit {
		q.Value, _ = q.IntValueIn(q.Unit.Kind.DefaultUnit)
		q.Unit = q.Unit.Kind.DefaultUnit
	}
}

func (q *Quantity) String() string {
	if q.Value == 0 {
		return "No Data"
	}
	switch q.Unit.Kind {
	case Count:
		return fmt.Sprintf("%d", q.Value)
	case Memory:
		byt, _ := q.IntValueIn(Byte)
		if byt >= GigaByte.Base {
			gb, _ := q.DoubleValueIn(GigaByte)
			return fmt.Sprintf("%.2f %s", gb, GigaByte)
		} else if byt >= MegaByte.Base {
			mb, _ := q.DoubleValueIn(MegaByte)
			return fmt.Sprintf("%.2f %s", mb, MegaByte)
		} else if byt >= KiloByte.Base {
			kb, _ := q.DoubleValueIn(KiloByte)
			return fmt.Sprintf("%.2f %s", kb, KiloByte)
		}
		return fmt.Sprintf("%d %s", byt, Byte)
	case Duration:
		td := q.toTimeDuration()
		return td.String()
	}
	return fmt.Sprintf("%d %s", q.Value, q.Unit)
}

func (q *Quantity) DoubleValueIn(target *Unit) (float64, error) {
	if q.Unit == target {
		return float64(q.Value), nil
	}
	if q.Unit.Kind != target.Kind {
		return 0, fmt.Errorf("can not convert [%s] to [%s], the kinds of the unit should be same", q.Unit, target)
	}

	return float64(q.Value) * (float64(q.Unit.Base) / float64(target.Base)), nil
}

func (q *Quantity) IntValueIn(target *Unit) (int64, error) {
	if q.Unit == target {
		return q.Value, nil
	}
	v, err := q.DoubleValueIn(target)
	if err != nil {
		return 0, err
	}
	return mathtoolkit.Trunc(v), nil
}

func (q *Quantity) Sub(sub *Quantity) *Quantity {
	if q.Unit.Kind != sub.Unit.Kind {
		panic("arithmetic operation between not matched unit kind")
	}

	m, n := q.Value, sub.Value

	toUnit := q.Unit
	if q.Unit.Base > sub.Unit.Base {
		m, _ = q.IntValueIn(sub.Unit)
		toUnit = sub.Unit
	} else if q.Unit.Base < sub.Unit.Base {
		n, _ = sub.IntValueIn(q.Unit)
	}

	return toUnit.Quantity(m - n)
}

func (q *Quantity) toTimeDuration() time.Duration {
	if q.Unit.Kind != Duration {
		panic("not kind of duration, can not convert")
	}

	num := time.Duration(q.Value)

	switch q.Unit {
	case NanoSecond:
		return time.Nanosecond * num
	case MicroSecond:
		return time.Microsecond * num
	case MilliSecond:
		return time.Millisecond * num
	case Second:
		return time.Second * num
	case Minute:
		return time.Minute * num
	case Hour:
		return time.Hour * num
	}
	panic(fmt.Sprintf("not resolved duration unit: [%s]", q.Unit))
}
