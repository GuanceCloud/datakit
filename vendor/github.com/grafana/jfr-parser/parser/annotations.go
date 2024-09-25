package parser

import (
	"github.com/grafana/jfr-parser/common/units"
	"strconv"

	"github.com/grafana/jfr-parser/internal/utils"
)

const (
	valueProperty = "value"

	annotationLabel         = "jdk.jfr.Label"
	annotationDescription   = "jdk.jfr.Description"
	annotationExperimental  = "jdk.jfr.Experimental"
	annotationCategory      = "jdk.jfr.Category"
	annotationTimestamp     = "jdk.jfr.Timestamp"
	annotationTimespan      = "jdk.jfr.Timespan"
	annotationMemoryAddress = "jdk.jfr.MemoryAddress"
	annotationPercentage    = "jdk.jfr.Percentage"
	annotationMemoryAmount  = "jdk.jfr.MemoryAmount"
	annotationDataAmount    = "jdk.jfr.DataAmount"
	annotationFrequency     = "jdk.jfr.Frequency"
	annotationUnsigned      = "jdk.jfr.Unsigned"
)

const (
	unitS            = "SECONDS"
	unitMS           = "MILLISECONDS"
	unitNS           = "NANOSECONDS"
	unitTicks        = "TICKS"
	unitSSinceEpoch  = "SECONDS_SINCE_EPOCH"
	unitMSSinceEpoch = "MILLISECONDS_SINCE_EPOCH"
	unitNSSinceEpoch = "NANOSECONDS_SINCE_EPOCH"
)

type AnnotationMetadata struct {
	ClassID int64
	Values  map[string]string
}

func (a *AnnotationMetadata) SetAttribute(key, value string) (err error) {
	switch key {
	case "class":
		a.ClassID, err = strconv.ParseInt(value, 10, 64)
	default:
		if a.Values == nil {
			a.Values = make(map[string]string)
		}
		a.Values[key] = value
	}
	return err
}

func (a *AnnotationMetadata) AppendChild(string) Element { return nil }

type BaseAnnotation struct {
	label        *string
	description  *string
	experimental *bool
	Annotations  []*AnnotationMetadata
}

func (b *BaseAnnotation) Label(classMap ClassMap) string {
	if b.label == nil {
		for _, annotation := range b.Annotations {
			if classMap[annotation.ClassID].Name == annotationLabel {
				b.label = utils.NewPointer(annotation.Values[valueProperty])
				break
			}
		}
		if b.label == nil {
			b.label = utils.NewPointer("")
		}
	}
	return *b.label
}

func (b *BaseAnnotation) Description(classMap ClassMap) string {
	if b.description == nil {
		for _, annotation := range b.Annotations {
			if classMap[annotation.ClassID].Name == annotationDescription {
				b.description = utils.NewPointer(annotation.Values[valueProperty])
				break
			}
		}
		if b.description == nil {
			b.description = utils.NewPointer("")
		}
	}
	return *b.description
}

func (b *BaseAnnotation) Experimental(classMap ClassMap) bool {
	if b.experimental == nil {
		if b.experimental == nil {
			for _, annotation := range b.Annotations {
				if classMap[annotation.ClassID].Name == annotationExperimental {
					b.experimental = utils.NewPointer(true)
					break
				}
			}
			b.experimental = utils.NewPointer(false)
		}
	}
	return *b.experimental
}

type ClassAnnotation struct {
	categories []string
	BaseAnnotation
}

type FieldAnnotation struct {
	unsigned      *bool
	tickTimestamp *bool
	unit          *units.Unit
	BaseAnnotation
}
