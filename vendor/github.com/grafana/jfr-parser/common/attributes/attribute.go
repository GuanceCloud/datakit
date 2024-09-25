package attributes

import (
	"fmt"
	"github.com/grafana/jfr-parser/common/types"
	"github.com/grafana/jfr-parser/common/units"
	"github.com/grafana/jfr-parser/parser"
	"reflect"
)

var (
	Blocking     = Attr[bool]("blocking", "Blocking", types.Boolean, "Whether the thread calling the vm operation was blocked or not")
	Safepoint    = Attr[bool]("safepoint", "Safepoint", types.Boolean, "Whether the vm operation occured at a safepoint or not")
	EventThread  = Attr[*parser.Thread]("eventThread", "Thread", types.Thread, "The thread in which the event occurred")
	BytesRead    = Attr[units.IQuantity]("bytesRead", "Bytes Read", types.Long, "Number of bytes read from the file (possibly 0)")
	BytesWritten = Attr[units.IQuantity]("bytesWritten", "Bytes Written", types.Long, "Number of bytes written to the file")
	SwitchRate   = Attr[float64]("switchRate", "Switch Rate", types.Float, "Number of context switches per second")

	StartTime           = AttrNoDesc[units.IQuantity]("startTime", "Start Time", types.Long)
	GcWhen              = AttrNoDesc[string]("when", "When", types.String)
	EventStacktrace     = AttrNoDesc[*parser.StackTrace]("stackTrace", "Stack Trace", types.StackTrace)
	ThreadStat          = AttrNoDesc[string]("state", "Thread State", types.ThreadState)
	CpuSamplingInterval = AttrNoDesc[units.IQuantity]("cpuInterval", "CPU Sampling Interval", types.Long)
	SettingName         = AttrNoDesc[string]("name", "Setting Name", types.String)
	SettingValue        = AttrNoDesc[string]("value", "Setting Value", types.String)
	SettingUnit         = AttrNoDesc[string]("unit", "Setting Unit", types.String)
	DatadogEndpoint     = AttrNoDesc[string]("endpoint", "Endpoint", types.String)
	Duration            = AttrNoDesc[units.IQuantity]("duration", "Duration", types.Long)

	JVMStartTime       = AttrSimple[units.IQuantity]("jvmStartTime", types.Long)
	SampleWeight       = AttrSimple[int64]("weight", types.Long)
	AllocWeight        = AttrSimple[float64]("weight", types.Float)
	AllocSize          = AttrSimple[units.IQuantity]("size", types.Long)
	WallSampleInterval = AttrSimple[units.IQuantity]("wallInterval", types.Long)
	Allocated          = AttrSimple[units.IQuantity]("allocated", types.Long)
	Size               = AttrSimple[units.IQuantity]("size", types.Long)
	HeapWeight         = AttrSimple[float64]("weight", types.Long)
)

type Attribute[T any] struct {
	Name        string // unique identifier for attribute
	Label       string // human-readable name
	ClassName   types.FieldClass
	Description string
}

func Attr[T any](name, label string, className types.FieldClass, description string) *Attribute[T] {
	return &Attribute[T]{
		Name:        name,
		Label:       label,
		ClassName:   className,
		Description: description,
	}
}

func AttrSimple[T any](name string, className types.FieldClass) *Attribute[T] {
	return &Attribute[T]{
		Name:      name,
		ClassName: className,
	}
}

func AttrNoDesc[T any](name, label string, className types.FieldClass) *Attribute[T] {
	return &Attribute[T]{
		Name:      name,
		Label:     label,
		ClassName: className,
	}
}

func (a *Attribute[T]) GetValue(event *parser.GenericEvent) (T, error) {
	var t T
	attr, ok := event.Attributes[a.Name]
	if !ok {
		return t, fmt.Errorf("attribute name [%s] is not found in the event", a.Name)
	}

	if x, ok := attr.(T); ok {
		return x, nil
	}

	attrValue := reflect.ValueOf(attr)
	attrType := attrValue.Type()
	tValue := reflect.ValueOf(&t).Elem()
	tType := tValue.Type()

	if attrType.ConvertibleTo(tType) {
		// t = t(attr)
		tValue.Set(attrValue.Convert(tType))
		return t, nil
	} else if attrValue.Kind() == reflect.Pointer && attrValue.Elem().Type().ConvertibleTo(tType) {
		// t = t(*attr)
		tValue.Set(attrValue.Elem().Convert(tType))
		return t, nil
	} else if tType.Kind() == reflect.Pointer && attrType.ConvertibleTo(tType.Elem()) {
		// t = t(&attr)
		ap := reflect.New(attrType)
		ap.Elem().Set(attrValue)
		if ap.Type().ConvertibleTo(tType) {
			tValue.Set(ap.Convert(tType))
			return t, nil
		}
	}

	fieldMeta := event.ClassMetadata.GetField(a.Name)
	fieldUnit := fieldMeta.Unit(event.ClassMetadata.ClassMap)

	if fieldUnit != nil || fieldMeta.TickTimestamp(event.ClassMetadata.ClassMap) {
		var (
			num      units.Number
			quantity units.IQuantity
		)

		switch attr.(type) {
		case *parser.Byte, *parser.Short, *parser.Int, *parser.Long:
			if fieldMeta.Unsigned(event.ClassMetadata.ClassMap) {
				var x any
				switch ax := attr.(type) {
				case *parser.Byte:
					x = uint8(*ax)
				case *parser.Short:
					x = uint16(*ax)
				case *parser.Int:
					x = uint32(*ax)
				case *parser.Long:
					x = uint64(*ax)
				}
				num = units.I64(reflect.ValueOf(x).Uint())
			} else {
				num = units.I64(reflect.ValueOf(attr).Elem().Int())
			}
		case *parser.Float, *parser.Double:
			num = units.F64(reflect.ValueOf(attr).Elem().Float())
		}

		if fieldMeta.TickTimestamp(event.ClassMetadata.ClassMap) {
			ts := fieldMeta.ChunkHeader.StartTimeNanos + ((num.Int64() - fieldMeta.ChunkHeader.StartTicks) * 1e9 / fieldMeta.ChunkHeader.TicksPerSecond)
			quantity = units.NewIntQuantity(ts, units.UnixNano)
		} else {
			if num.Float() {
				quantity = units.NewFloatQuantity(num.Float64(), fieldUnit)
			} else {
				quantity = units.NewIntQuantity(num.Int64(), fieldUnit)
			}
		}

		if q, ok := quantity.(T); ok {
			return q, nil
		}
	}

	switch any(t).(type) {
	case string:
		s, err := parser.ToString(attr)
		if err != nil {
			return t, fmt.Errorf("unable to resolve string: %w", err)
		}
		reflect.ValueOf(&t).Elem().SetString(s)
		return t, nil
	}

	return t, fmt.Errorf("attribute is not type of %T", t)
}
