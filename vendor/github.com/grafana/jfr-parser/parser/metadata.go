package parser

import (
	"fmt"
	"github.com/grafana/jfr-parser/common/units"
	"github.com/grafana/jfr-parser/internal/utils"
	"strconv"

	"github.com/grafana/jfr-parser/parser/types/def"
)

type Element interface {
	SetAttribute(key, value string) error
	AppendChild(name string) Element
}

type ElementWithHeader interface {
	Element
	SetHeader(header *Header)
}

// SettingMetadata TODO: Proper attribute support for SettingMetadata
type SettingMetadata struct {
	Values map[string]string
}

func (s *SettingMetadata) SetAttribute(key, value string) error {
	if s.Values == nil {
		s.Values = make(map[string]string)
	}
	s.Values[key] = value
	return nil
}

func (s *SettingMetadata) AppendChild(string) Element { return nil }

type FieldMetadata struct {
	ClassID      int64
	Name         string
	ConstantPool bool
	Dimension    int32
	ChunkHeader  *Header
	FieldAnnotation
}

func (f *FieldMetadata) SetAttribute(key, value string) (err error) {
	switch key {
	case "name":
		f.Name = value
	case "class":
		f.ClassID, err = strconv.ParseInt(value, 10, 64)
	case "constantPool":
		f.ConstantPool, err = parseBool(value)
	case "dimension":
		var n int64
		n, err = strconv.ParseInt(value, 10, 32)
		f.Dimension = int32(n)
	}
	return nil
}

func (f *FieldMetadata) AppendChild(name string) Element {
	switch name {
	case "annotation":
		am := &AnnotationMetadata{}
		f.Annotations = append(f.Annotations, am)
		return am
	}
	return nil
}

func (f *FieldMetadata) IsArray() bool {
	switch f.Dimension {
	case 0:
		return false
	case 1:
		return true
	default:
		panic(fmt.Sprintf("dimension value [%d] is not supported", f.Dimension))
	}
}

func (f *FieldMetadata) SetHeader(header *Header) {
	f.ChunkHeader = header
}

func (f *FieldMetadata) Unsigned(classMap ClassMap) bool {
	if f.unsigned == nil {
		f.resolve(classMap)
	}
	return *f.unsigned
}

func (f *FieldMetadata) Unit(classMap ClassMap) *units.Unit {
	if f.unit == nil {
		f.resolve(classMap)
	}
	if f.unit == units.Unknown {
		return nil
	}
	return f.unit
}

func (f *FieldMetadata) TickTimestamp(classMap ClassMap) bool {
	if f.tickTimestamp == nil {
		f.resolve(classMap)
	}
	return *f.tickTimestamp
}

func (f *FieldMetadata) resolve(classMap ClassMap) {
	for _, annotation := range f.Annotations {
		switch classMap[annotation.ClassID].Name {
		case annotationUnsigned:
			f.unsigned = utils.NewPointer(true)
		case annotationMemoryAmount, annotationDataAmount:
			f.unit = units.Byte
		case annotationPercentage:
			f.unit = units.Multiple
		case annotationTimespan:
			switch annotation.Values[valueProperty] {
			case unitTicks:
				f.unit = units.Nanosecond.Derived("tick", units.F64(1e9/float64(f.ChunkHeader.TicksPerSecond)))
			case unitNS:
				f.unit = units.Nanosecond
			case unitMS:
				f.unit = units.Millisecond
			case unitS:
				f.unit = units.Second
			}
		case annotationFrequency:
			f.unit = units.Hertz
		case annotationTimestamp:
			switch annotation.Values[valueProperty] {
			case unitTicks:
				f.tickTimestamp = utils.NewPointer(true)
			case unitSSinceEpoch:
				f.unit = units.UnixSecond
			case unitMSSinceEpoch:
				f.unit = units.UnixMilli
			case unitNSSinceEpoch:
				f.unit = units.UnixNano
			}
		}
	}
	if f.unsigned == nil {
		f.unsigned = utils.NewPointer(false)
	}
	if f.tickTimestamp == nil {
		f.tickTimestamp = utils.NewPointer(false)
	}
	if f.unit == nil {
		f.unit = units.Unknown
	}
}

type ClassMetadata struct {
	ID         int64
	Name       string
	SuperType  string
	SimpleType bool
	Fields     []*FieldMetadata
	fieldsDict map[string]*FieldMetadata
	Settings   []*SettingMetadata
	ClassMap   ClassMap // redundant ClassMap here is for getting field class more easily
	ClassAnnotation
}

func (c *ClassMetadata) Category() []string {
	if c.categories == nil {
		for _, annotation := range c.Annotations {
			if c.ClassMap[annotation.ClassID].Name == annotationCategory {
				categories := make([]string, 0, len(annotation.Values))
				idx := 0
				for {
					cat, ok := annotation.Values[fmt.Sprintf("%s-%d", valueProperty, idx)]
					if !ok {
						break
					}
					categories = append(categories, cat)
					idx++
				}
				c.categories = categories
			}
			if c.categories == nil {
				c.categories = []string{}
			}
		}
	}
	return c.categories
}

func (c *ClassMetadata) Label() string {
	return c.BaseAnnotation.Label(c.ClassMap)
}

func (c *ClassMetadata) Unit(fieldName string) *units.Unit {
	fieldMeta := c.GetField(fieldName)
	if fieldMeta == nil {
		return nil
	}
	return fieldMeta.Unit(c.ClassMap)
}

func (c *ClassMetadata) Unsigned(fieldName string) bool {
	fieldMeta := c.GetField(fieldName)
	if fieldMeta == nil {
		return false
	}
	return fieldMeta.Unsigned(c.ClassMap)
}

func (c *ClassMetadata) buildFieldMap() {
	if c.fieldsDict == nil {
		c.fieldsDict = make(map[string]*FieldMetadata, len(c.Fields))
		for _, field := range c.Fields {
			c.fieldsDict[field.Name] = field
		}
	}
}

func (c *ClassMetadata) SetAttribute(key, value string) (err error) {
	switch key {
	case "id":
		c.ID, err = strconv.ParseInt(value, 10, 64)
	case "name":
		c.Name = value
	case "superType":
		c.SuperType = value
	case "simpleType":
		c.SimpleType, err = parseBool(value)
	}
	return err
}

func (c *ClassMetadata) ContainsField(fieldName, fieldClass string) bool {
	md := c.GetField(fieldName)
	if md == nil {
		return false
	}

	if c.ClassMap[md.ClassID].Name == fieldClass {
		return true
	}
	return false
}

func (c *ClassMetadata) GetField(fieldName string) *FieldMetadata {
	if c.fieldsDict == nil {
		c.buildFieldMap()
	}
	return c.fieldsDict[fieldName]
}

func (c *ClassMetadata) AppendChild(name string) Element {
	switch name {
	case "field":
		fm := &FieldMetadata{}
		c.Fields = append(c.Fields, fm)
		return fm
	case "setting":
		sm := &SettingMetadata{}
		c.Settings = append(c.Settings, sm)
		return sm
	case "annotation":
		am := &AnnotationMetadata{}
		c.Annotations = append(c.Annotations, am)
		return am
	}
	return nil
}

type Metadata struct {
	Classes []*ClassMetadata
}

func (m *Metadata) SetAttribute(string, string) error { return nil }

func (m *Metadata) AppendChild(name string) Element {
	switch name {
	case "class":
		cm := &ClassMetadata{}
		m.Classes = append(m.Classes, cm)
		return cm
	default:
	}
	return nil
}

type Region struct {
	Locale        string
	GMTOffset     string
	TicksToMillis string
}

func (m *Region) SetAttribute(key, value string) error {
	switch key {
	case "locale":
		m.Locale = value
	case "gmtOffset":
		// TODO int?
		m.GMTOffset = value
	case "ticksToMillis":
		// TODO int?
		m.TicksToMillis = value
	}
	return nil
}

func (m *Region) AppendChild(string) Element { return nil }

type Root struct {
	Metadata *Metadata
	Region   Region
}

func (r *Root) SetAttribute(string, string) error { return nil }

func (r *Root) AppendChild(name string) Element {
	switch name {
	case "metadata":
		r.Metadata = &Metadata{}
		return r.Metadata
	case "region":
		r.Region = Region{}
		return &r.Region
	}
	return nil
}

type ChunkMetadata struct {
	StartTime int64
	Duration  int64
	ID        int64
	Root      *Root
	Header    *Header
	ClassMap  ClassMap
}

func (m *ChunkMetadata) buildClassMap() {
	classMap := make(ClassMap, len(m.Root.Metadata.Classes))
	for _, class := range m.Root.Metadata.Classes {
		class.ClassMap = classMap // assign all class to every class metadata
		classMap[class.ID] = class
	}

	m.ClassMap = classMap
	m.Root.Metadata.Classes = nil
}

func (m *ChunkMetadata) Parse(r Reader) (err error) {
	if kind, err := r.VarLong(); err != nil {
		return fmt.Errorf("unable to retrieve event type: %w", err)
	} else if kind != 0 {
		return fmt.Errorf("unexpected metadata event type: %d", kind)
	}

	if m.StartTime, err = r.VarLong(); err != nil {
		return fmt.Errorf("unable to parse metadata event's start time: %w", err)
	}
	if m.Duration, err = r.VarLong(); err != nil {
		return fmt.Errorf("unable to parse metadata event's duration: %w", err)
	}
	if m.ID, err = r.VarLong(); err != nil {
		return fmt.Errorf("unable to parse metadata event's ID: %w", err)
	}
	n, err := r.VarInt()
	if err != nil {
		return fmt.Errorf("unable to parse metadata event's number of strings: %w", err)
	}
	// TODO: assert n is small enough
	strings := make([]string, n)
	for i := 0; i < int(n); i++ {
		if x, err := r.String(); err != nil {
			return fmt.Errorf("unable to parse metadata event's string: %w", err)
		} else {
			strings[i] = x.s
		}
	}

	name, err := parseName(r, strings)
	if err != nil {
		return err
	}
	if name != "root" {
		return fmt.Errorf("invalid root element name: %s", name)
	}

	m.Root = &Root{}
	if err = parseElement(r, strings, m.Header, m.Root); err != nil {
		return fmt.Errorf("unable to parse metadata element tree: %w", err)
	}

	m.buildClassMap()

	return nil
}

func (p *Parser) readMeta(pos int) error {
	p.TypeMap.IDMap = make(map[def.TypeID]*def.Class, 43+5)
	p.TypeMap.NameMap = make(map[string]*def.Class, 43+5)

	if err := p.seek(pos); err != nil {
		return err
	}
	sz, err := p.varInt()
	if err != nil {
		return err
	}
	p.metaSize = sz
	_, err = p.varInt()
	if err != nil {
		return err
	}
	_, err = p.varLong()
	if err != nil {
		return err
	}
	_, err = p.varLong()
	if err != nil {
		return err
	}
	_, err = p.varLong()
	if err != nil {
		return err
	}
	nstr, err := p.varInt()
	if err != nil {
		return err
	}
	strings := make([]string, nstr)
	for i := 0; i < int(nstr); i++ {
		strings[i], err = p.string()
		if err != nil {
			return err
		}
	}

	e, err := p.readElement(strings, false)
	if err != nil {
		return err
	}
	if e.name != "root" {
		return fmt.Errorf("expected root element, got %s", e.name)
	}
	for i := 0; i < e.childCount; i++ {
		meta, err := p.readElement(strings, false)
		if err != nil {
			return err
		}
		//fmt.Println(meta.name)
		switch meta.name {
		case "metadata":
			for j := 0; j < meta.childCount; j++ {
				classElement, err := p.readElement(strings, true)

				if err != nil {
					return err
				}
				cls, err := def.NewClass(classElement.attr, classElement.childCount)
				if err != nil {
					return err
				}

				for k := 0; k < classElement.childCount; k++ {
					field, err := p.readElement(strings, true)
					if err != nil {
						return err
					}
					if field.name == "field" {
						f, err := def.NewField(field.attr)
						if err != nil {
							return err
						}
						cls.Fields = append(cls.Fields, f)
					}
					for l := 0; l < field.childCount; l++ {
						_, err := p.readElement(strings, false)
						if err != nil {
							return err
						}
					}

				}
				//fmt.Println(cls.String())
				p.TypeMap.IDMap[cls.ID] = cls
				p.TypeMap.NameMap[cls.Name] = cls

			}
		case "region":
			break
		default:
			return fmt.Errorf("unexpected element %s", meta.name)
		}
	}
	if err := p.checkTypes(); err != nil {
		return err
	}
	return nil
}
func parseElement(r Reader, s []string, chunkHeader *Header, e Element) error {
	n, err := r.VarInt()
	if err != nil {
		return fmt.Errorf("unable to parse attribute count: %w", err)
	}

	if ex, ok := e.(ElementWithHeader); ok {
		ex.SetHeader(chunkHeader)
	}

	for i := int64(0); i < int64(n); i++ {
		k, err := parseName(r, s)
		if err != nil {
			return fmt.Errorf("unable to parse attribute key: %w", err)
		}
		v, err := parseName(r, s)
		if err != nil {
			return fmt.Errorf("unable to parse attribute value: %w", err)
		}
		if err := e.SetAttribute(k, v); err != nil {
			return fmt.Errorf("unable to set element attribute: %w", err)
		}
	}
	n, err = r.VarInt()
	if err != nil {
		return fmt.Errorf("unable to parse element count: %w", err)
	}
	// TODO: assert n is small enough
	for i := 0; i < int(n); i++ {
		name, err := parseName(r, s)
		if err != nil {
			return fmt.Errorf("unable to parse element name: %w", err)
		}
		child := e.AppendChild(name)
		if child == nil {
			return fmt.Errorf("unexpected child in metadata event: %s", name)
		}
		if err = parseElement(r, s, chunkHeader, child); err != nil {
			return fmt.Errorf("unable to parse child element: %w", err)
		}
	}
	return nil
}

func (p *Parser) readElement(strings []string, needAttributes bool) (element, error) {
	iname, err := p.varInt()
	if err != nil {
		return element{}, err
	}
	if iname < 0 || int(iname) >= len(strings) {
		return element{}, def.ErrIntOverflow
	}
	name := strings[iname]
	attributeCount, err := p.varInt()
	if err != nil {
		return element{}, err
	}
	var attributes map[string]string
	if needAttributes {
		attributes = make(map[string]string, attributeCount)
	}
	for i := 0; i < int(attributeCount); i++ {
		attributeName, err := p.varInt()
		if err != nil {
			return element{}, err
		}
		if attributeName < 0 || int(attributeName) >= len(strings) {
			return element{}, def.ErrIntOverflow
		}
		attributeValue, err := p.varInt()
		if err != nil {
			return element{}, err
		}
		if attributeValue < 0 || int(attributeValue) >= len(strings) {
			return element{}, def.ErrIntOverflow
		}
		if needAttributes {
			attributes[strings[attributeName]] = strings[attributeValue]
		} else {
			//fmt.Printf("                              >>> skipping attribute %s=%s\n", strings[attributeName], strings[attributeValue])
		}
	}

	childCount, err := p.varInt()
	if err != nil {
		return element{}, err
	}
	return element{
		name:       name,
		attr:       attributes,
		childCount: int(childCount),
	}, nil

}

func parseName(r Reader, s []string) (string, error) {
	n, err := r.VarInt()
	if err != nil {
		return "", fmt.Errorf("unable to parse string name index: %w", err)
	}
	if int(n) >= len(s) {
		return "", fmt.Errorf("invalid name index %d, only %d names available", n, len(s))
	}
	return s[int(n)], nil
}

func parseBool(s string) (bool, error) {
	if s == "true" {
		return true, nil
	}
	if s == "false" {
		return false, nil
	}
	return false, fmt.Errorf("unable to parse '%s' as boolean", s)
}

type element struct {
	name       string
	attr       map[string]string
	childCount int
}
