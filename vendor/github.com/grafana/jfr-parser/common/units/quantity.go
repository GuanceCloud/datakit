package units

import (
	"fmt"
)

type IQuantity interface {
	Unit() *Unit
	In(unit *Unit) (IQuantity, error)
	FloatValue() float64
	IntValue() int64
	Add(q IQuantity) (IQuantity, error)
	String() string
}

type IntQuantity struct {
	num  int64
	unit *Unit
}

func (i *IntQuantity) String() string {
	return fmt.Sprintf("%d%s", i.num, i.unit.Name)
}

func (i *IntQuantity) Add(q IQuantity) (IQuantity, error) {
	//TODO implement me
	panic("implement me")
}

func (i *IntQuantity) IntValue() int64 {
	return i.num
}

func (i *IntQuantity) FloatValue() float64 {
	return float64(i.num)
}

func (i *IntQuantity) Unit() *Unit {
	return i.unit
}

func (i *IntQuantity) In(u *Unit) (IQuantity, error) {
	if i.unit.Kind != u.Kind {
		return nil, fmt.Errorf("imcompatible unit kinds between source [%q] and target [%q]", i.unit.Kind, u.Kind)
	}
	if i.num == 0 {
		return NewIntQuantity(0, u), nil
	}
	if !i.unit.Base.Float() && !u.Base.Float() {
		if iBase, uBase := i.unit.Base.Int64(), u.Base.Int64(); (i.num*iBase)%uBase == 0 {
			return NewIntQuantity(i.num*iBase/uBase, u), nil
		}
	}
	return NewFloatQuantity(i.FloatValue()*i.unit.Base.Float64()/u.Base.Float64(), u), nil
}

func NewIntQuantity(number int64, unit *Unit) IQuantity {
	return &IntQuantity{
		num:  number,
		unit: unit,
	}
}

type FloatQuantity struct {
	num  float64
	unit *Unit
}

func (f *FloatQuantity) String() string {
	return fmt.Sprintf("%f%s", f.num, f.unit.Name)
}

func (f *FloatQuantity) Add(q IQuantity) (IQuantity, error) {
	//TODO implement me
	panic("implement me")
}

func (f *FloatQuantity) Unit() *Unit {
	return f.unit
}

func (f *FloatQuantity) In(u *Unit) (IQuantity, error) {
	if f.unit.Kind != u.Kind {
		return nil, fmt.Errorf("imcompatible unit kinds between source [%q] and target [%q]", f.unit.Kind, u.Kind)
	}
	return NewFloatQuantity(f.num*f.unit.Base.Float64()/u.Base.Float64(), u), nil
}

func (f *FloatQuantity) FloatValue() float64 {
	return f.num
}

func (f *FloatQuantity) IntValue() int64 {
	return int64(f.num)
}

func NewFloatQuantity(number float64, unit *Unit) IQuantity {
	return &FloatQuantity{
		num:  number,
		unit: unit,
	}
}
