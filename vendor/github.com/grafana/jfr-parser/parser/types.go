package parser

import (
	"errors"
	"fmt"
	"reflect"

	types2 "github.com/grafana/jfr-parser/common/types"
)

var types = map[types2.FieldClass]func() ParseResolvable{
	types2.Boolean:              func() ParseResolvable { return SetPfFunc(new(Boolean)) },
	types2.Byte:                 func() ParseResolvable { return SetPfFunc(new(Byte)) },
	types2.Char:                 func() ParseResolvable { return SetPfFunc(new(Char)) },
	types2.Double:               func() ParseResolvable { return SetPfFunc(new(Double)) },
	types2.Float:                func() ParseResolvable { return SetPfFunc(new(Float)) },
	types2.Int:                  func() ParseResolvable { return SetPfFunc(new(Int)) },
	types2.Long:                 func() ParseResolvable { return SetPfFunc(new(Long)) },
	types2.Short:                func() ParseResolvable { return SetPfFunc(new(Short)) },
	types2.Class:                func() ParseResolvable { return SetPfFunc(new(Class)) },
	types2.String:               func() ParseResolvable { return SetPfFunc(new(String)) },
	types2.Thread:               func() ParseResolvable { return SetPfFunc(new(Thread)) },
	types2.ClassLoader:          func() ParseResolvable { return SetPfFunc(new(ClassLoader)) },
	types2.CodeBlobType:         func() ParseResolvable { return SetPfFunc(new(CodeBlobType)) },
	types2.FlagValueOrigin:      func() ParseResolvable { return SetPfFunc(new(FlagValueOrigin)) },
	types2.FrameType:            func() ParseResolvable { return SetPfFunc(new(FrameType)) },
	types2.G1YCType:             func() ParseResolvable { return SetPfFunc(new(G1YCType)) },
	types2.GCName:               func() ParseResolvable { return SetPfFunc(new(GCName)) },
	types2.Method:               func() ParseResolvable { return SetPfFunc(new(Method)) },
	types2.Module:               func() ParseResolvable { return SetPfFunc(new(Module)) },
	types2.NarrowOopMode:        func() ParseResolvable { return SetPfFunc(new(NarrowOopMode)) },
	types2.NetworkInterfaceName: func() ParseResolvable { return SetPfFunc(new(NetworkInterfaceName)) },
	types2.Package:              func() ParseResolvable { return SetPfFunc(new(Package)) },
	types2.StackFrame:           func() ParseResolvable { return SetPfFunc(new(StackFrame)) },
	types2.StackTrace:           func() ParseResolvable { return SetPfFunc(new(StackTrace)) },
	types2.Symbol:               func() ParseResolvable { return SetPfFunc(new(Symbol)) },
	types2.ThreadState:          func() ParseResolvable { return SetPfFunc(new(ThreadState)) },
	types2.InflateCause:         func() ParseResolvable { return SetPfFunc(new(InflateCause)) },
	types2.GCCause:              func() ParseResolvable { return SetPfFunc(new(GCCause)) },
	types2.CompilerPhaseType:    func() ParseResolvable { return SetPfFunc(new(CompilerPhaseType)) },
	types2.ThreadGroup:          func() ParseResolvable { return SetPfFunc(new(ThreadGroup)) },
	types2.GCThresholdUpdater:   func() ParseResolvable { return SetPfFunc(new(GCThresholdUpdater)) },
	types2.MetaspaceObjectType:  func() ParseResolvable { return SetPfFunc(new(MetaspaceObjectType)) },
	types2.ExecutionMode:        func() ParseResolvable { return SetPfFunc(new(ExecutionMode)) },
	types2.VMOperationType:      func() ParseResolvable { return SetPfFunc(new(VMOperationType)) },
	types2.G1HeapRegionType:     func() ParseResolvable { return SetPfFunc(new(G1HeapRegionType)) },
	types2.GCWhen:               func() ParseResolvable { return SetPfFunc(new(GCWhen)) },
	types2.ReferenceType:        func() ParseResolvable { return SetPfFunc(new(ReferenceType)) },
	types2.MetadataType:         func() ParseResolvable { return SetPfFunc(new(MetadataType)) },
	types2.LogLevel:             func() ParseResolvable { return SetPfFunc(new(LogLevel)) },
	types2.AttributeValue:       func() ParseResolvable { return SetPfFunc(new(AttributeValue)) },
}

var (
	_ ParseResolveFielder = (*Class)(nil)
	_ ParseResolveFielder = (*Thread)(nil)
	_ ParseResolveFielder = (*ClassLoader)(nil)
	_ ParseResolveFielder = (*CodeBlobType)(nil)
	_ ParseResolveFielder = (*FrameType)(nil)
	_ ParseResolveFielder = (*G1YCType)(nil)
	_ ParseResolveFielder = (*GCName)(nil)
	_ ParseResolveFielder = (*Method)(nil)
	_ ParseResolveFielder = (*Module)(nil)
	_ ParseResolveFielder = (*NarrowOopMode)(nil)
	_ ParseResolveFielder = (*NetworkInterfaceName)(nil)
	_ ParseResolveFielder = (*Package)(nil)
	_ ParseResolveFielder = (*Symbol)(nil)
	_ ParseResolveFielder = (*StackTrace)(nil)
	_ ParseResolveFielder = (*ThreadState)(nil)
	_ ParseResolveFielder = (*InflateCause)(nil)
	_ ParseResolveFielder = (*GCCause)(nil)
	_ ParseResolveFielder = (*CompilerPhaseType)(nil)
	_ ParseResolveFielder = (*ThreadGroup)(nil)
	_ ParseResolveFielder = (*GCThresholdUpdater)(nil)
	_ ParseResolveFielder = (*MetaspaceObjectType)(nil)
	_ ParseResolveFielder = (*ExecutionMode)(nil)
	_ ParseResolveFielder = (*VMOperationType)(nil)
	_ ParseResolveFielder = (*G1HeapRegionType)(nil)
	_ ParseResolveFielder = (*GCWhen)(nil)
	_ ParseResolveFielder = (*ReferenceType)(nil)
	_ ParseResolveFielder = (*MetadataType)(nil)
	_ ParseResolveFielder = (*LogLevel)(nil)
	_ ParseResolveFielder = (*AttributeValue)(nil)
	_ ParseResolveFielder = (*InflateCause)(nil)
)

func ParseClass(r Reader, classes ClassMap, cpools PoolMap, classID int64) (ParseResolvable, error) {
	class, ok := classes[classID]
	if !ok {
		return nil, fmt.Errorf("unexpected class %d", classID)
	}
	var v ParseResolvable
	if typeFn, ok := types[types2.FieldClass(class.Name)]; ok {
		v = typeFn()
	} else {
		v = NewParseResolvable[*DefaultStructType]()
		if vx, ok := v.(*DefaultStructType); ok {
			classMeta := classes[classID]
			classMeta.buildFieldMap()
			vx.className = classMeta.Name
			vx.fieldsDict = classMeta.fieldsDict
		}
	}
	if err := v.Parse(r, classes, cpools, class); err != nil {
		return nil, err
	}
	return v, nil
}

type Event interface {
	Parseable
}

type Parseable interface {
	Parse(Reader, ClassMap, PoolMap, *ClassMetadata) error
}

type Resolvable interface {
	Resolve(ClassMap, PoolMap) error
}

type ParseFieldFunc func(name string, p ParseResolvable) error

type setFielder interface {
	setField(name string, p ParseResolvable) error
}

type parseFieldFuncSetter interface {
	setParseFieldFunc(fn ParseFieldFunc)
}

type ParseResolvable interface {
	Parseable
	Resolvable
}

type ParseResolveFielder interface {
	ParseResolvable
	setFielder
}

type constantReference struct {
	classID int64
	field   string
	index   int64
}

type BaseStructType struct {
	constants   []constantReference
	resolved    bool
	fieldAssign ParseFieldFunc
	unresolved  []ParseResolvable
}

func SetPfFunc(p ParseResolvable) ParseResolvable {
	if setter, ok := p.(parseFieldFuncSetter); ok {
		if pf, yes := p.(setFielder); yes {
			setter.setParseFieldFunc(pf.setField)
		}
	}
	return p
}

func NewParseResolvable[T ParseResolvable]() ParseResolvable {
	var t T
	rt := reflect.TypeOf(t)
	if rt.Kind() != reflect.Pointer {
		panic(fmt.Sprintf("generic parameter should be a pointer type: %v", rt.Kind()))
	}

	instance := reflect.New(rt.Elem()).Interface()

	if setter, ok := instance.(parseFieldFuncSetter); ok {
		if pf, ok := instance.(setFielder); ok {
			setter.setParseFieldFunc(pf.setField)
		}
	}
	return instance.(ParseResolvable)
}

func (b *BaseStructType) setParseFieldFunc(fn ParseFieldFunc) {
	b.fieldAssign = fn
}

func (b *BaseStructType) Parse(r Reader, classMap ClassMap, poolMap PoolMap, metadata *ClassMetadata) error {
	return b.parseFields(r, classMap, poolMap, metadata, b.resolved)
}

func (b *BaseStructType) parseFields(r Reader, classes ClassMap, cpools PoolMap, class *ClassMetadata, resolved bool) error {
	for _, f := range class.Fields {
		if f.ConstantPool {
			if !resolved {
				if err := b.appendConstant(r, f.Name, f.ClassID); err != nil {
					return fmt.Errorf("failed to parse %s: unable to append constant: %w", class.Name, err)
				}
			} else {
				cPool, ok := cpools[f.ClassID]
				if !ok {
					continue
					//return fmt.Errorf("constant pool for class [%s] doesn't exists", class.Name)
				}
				idx, err := r.VarLong()
				if err != nil {
					return fmt.Errorf("unable to read constant index")
				}
				p, ok := cPool.Pool[idx]
				if !ok {
					continue
					//return fmt.Errorf("constant value of index [%d] doesn't exists", idx)
				}
				if err := b.fieldAssign(f.Name, p); err != nil {
					return fmt.Errorf("unable to parse constant field %s: %w", f.Name, err)
				}
			}
		} else if f.Dimension == 1 {
			n, err := r.VarInt()
			if err != nil {
				return fmt.Errorf("failed to parse %s: unable to read array length: %w", class.Name, err)
			}
			// done: assert n is small enough
			for i := int64(0); i < int64(n); i++ {
				p, err := ParseClass(r, classes, cpools, f.ClassID)
				if err != nil {
					return fmt.Errorf("failed to parse %s: unable to read an array element: %w", class.Name, err)
				}
				if err := b.fieldAssign(f.Name, p); err != nil {
					return fmt.Errorf("failed to parse %s: unable to parse an array element: %w", class.Name, err)
				}

				b.unresolved = append(b.unresolved, p) // cache fields need to resolve
			}
		} else {
			p, err := ParseClass(r, classes, cpools, f.ClassID)
			if err != nil {
				return fmt.Errorf("failed to parse %s: unable to read a field: %w", class.Name, err)
			}
			if err := b.fieldAssign(f.Name, p); err != nil {
				return fmt.Errorf("failed to parse %s: unable to parse a field: %w", class.Name, err)
			}
			b.unresolved = append(b.unresolved, p) // cache fields need to resolve
		}
	}
	return nil
}

func (b *BaseStructType) Resolve(classMap ClassMap, poolMap PoolMap) error {
	if !b.resolved {
		b.resolved = true
		for _, c := range b.constants {
			p, ok := poolMap[c.classID]
			if !ok {
				// Non-existent constant pool references seem to be used to mark no s
				continue
			}
			it, ok := p.Pool[c.index]
			if !ok {
				// Non-existent constant pool references seem to be used to mark no s
				continue
			}
			if b.fieldAssign != nil {
				if err := b.fieldAssign(c.field, it); err != nil {
					return fmt.Errorf("unable to resolve constants for field %s: %w", c.field, err)
				}
			}
		}
		b.constants = nil
	}

	if len(b.unresolved) > 0 {
		for _, needResolve := range b.unresolved {
			if err := needResolve.Resolve(classMap, poolMap); err != nil {
				return fmt.Errorf("unable to resolve field value: %v", needResolve)
			}
		}
		b.unresolved = nil
	}

	return nil
}

func appendConstant(r Reader, constants *[]constantReference, name string, class int64) error {
	i, err := r.VarLong()
	if err != nil {
		return fmt.Errorf("unable to read constant index")
	}
	*constants = append(*constants, constantReference{field: name, index: i, classID: class})
	return nil
}

func (b *BaseStructType) appendConstant(r Reader, name string, class int64) error {
	i, err := r.VarLong()
	if err != nil {
		return fmt.Errorf("unable to read constant index")
	}
	b.constants = append(b.constants, constantReference{field: name, index: i, classID: class})
	return nil
}

func parseFields(r Reader, classes ClassMap, cpools PoolMap, class *ClassMetadata, constants *[]constantReference, resolved bool, cb func(string, ParseResolvable) error) error {
	for _, f := range class.Fields {
		if f.ConstantPool {
			if constants != nil && !resolved {
				if err := appendConstant(r, constants, f.Name, f.ClassID); err != nil {
					return fmt.Errorf("failed to parse %s: unable to append constant: %w", class.Name, err)
				}
			} else {
				cPool, ok := cpools[f.ClassID]
				if !ok {
					continue
					//return fmt.Errorf("constant pool for class [%s] doesn't exists", class.Name)
				}
				idx, err := r.VarLong()
				if err != nil {
					return fmt.Errorf("unable to read constant index")
				}
				p, ok := cPool.Pool[idx]
				if !ok {
					continue
					//return fmt.Errorf("constant value of index [%d] doesn't exists", idx)
				}
				if err := cb(f.Name, p); err != nil {
					return fmt.Errorf("unable to parse constant field %s: %w", f.Name, err)
				}
			}
		} else if f.Dimension == 1 {
			n, err := r.VarInt()
			if err != nil {
				return fmt.Errorf("failed to parse %s: unable to read array length: %w", class.Name, err)
			}
			// done: assert n is small enough
			for i := int64(0); i < int64(n); i++ {
				p, err := ParseClass(r, classes, cpools, f.ClassID)
				if err != nil {
					return fmt.Errorf("failed to parse %s: unable to read an array element: %w", class.Name, err)
				}
				if err := cb(f.Name, p); err != nil {
					return fmt.Errorf("failed to parse %s: unable to parse an array element: %w", class.Name, err)
				}
			}
		} else {
			p, err := ParseClass(r, classes, cpools, f.ClassID)
			if err != nil {
				return fmt.Errorf("failed to parse %s: unable to read a field: %w", class.Name, err)
			}
			if err := cb(f.Name, p); err != nil {
				return fmt.Errorf("failed to parse %s: unable to parse a field: %w", class.Name, err)
			}
		}
	}
	return nil
}

func resolveConstants(classes ClassMap, cpools PoolMap, constants *[]constantReference, resolved *bool, cb func(string, ParseResolvable) error) error {
	if *resolved {
		return nil
	}
	*resolved = true
	for _, c := range *constants {
		if err := ResolveConstants(classes, cpools); err != nil {
			return fmt.Errorf("unable to resolve contants: %w", err)
		}
		p, ok := cpools[c.classID]
		if !ok {
			// Non-existent constant pool references seem to be used to mark no value
			continue
		}
		it, ok := p.Pool[c.index]
		if !ok {
			// Non-existent constant pool references seem to be used to mark no s
			continue
		}
		if err := it.Resolve(classes, cpools); err != nil {
			return err
		}
		if err := cb(c.field, it); err != nil {
			return fmt.Errorf("unable to resolve constants for field %s: %w", c.field, err)
		}
	}
	*constants = nil
	return nil
}

type Boolean bool

func (b *Boolean) Parse(r Reader, _ ClassMap, _ PoolMap, _ *ClassMetadata) error {
	// TODO: Assert simpletype, no fields, etc.
	x, err := r.Boolean()
	*b = Boolean(x)
	return err
}

func (*Boolean) Resolve(ClassMap, PoolMap) error { return nil }

func toBoolean(p Parseable) (bool, error) {
	x, ok := p.(*Boolean)
	if !ok {
		return false, errors.New("not a Boolean")
	}
	return bool(*x), nil
}

type Byte int8

func (b *Byte) Parse(r Reader, _ ClassMap, _ PoolMap, _ *ClassMetadata) error {
	x, err := r.Byte()
	*b = Byte(x)
	return err
}

func (*Byte) Resolve(ClassMap, PoolMap) error { return nil }

func toByte(p Parseable) (int8, error) {
	x, ok := p.(*Byte)
	if !ok {
		return 0, errors.New("not a Byte")
	}
	return int8(*x), nil
}

type Char uint16

func (c *Char) Parse(r Reader, _ ClassMap, _ PoolMap, _ *ClassMetadata) error {
	x, err := r.Char()
	if err != nil {
		return fmt.Errorf("unable to resolve char: %w", err)
	}
	*c = Char(x)
	return nil
}

func (*Char) Resolve(ClassMap, PoolMap) error {
	return nil
}

type Double float64

func (d *Double) Parse(r Reader, _ ClassMap, _ PoolMap, _ *ClassMetadata) error {
	x, err := r.Double()
	*d = Double(x)
	return err
}

func (*Double) Resolve(ClassMap, PoolMap) error { return nil }

func toDouble(p Parseable) (float64, error) {
	x, ok := p.(*Double)
	if !ok {
		return 0, errors.New("not a Double")
	}
	return float64(*x), nil
}

type Float float32

func (f *Float) Parse(r Reader, _ ClassMap, _ PoolMap, _ *ClassMetadata) error {
	x, err := r.Float()
	*f = Float(x)
	return err
}

func (*Float) Resolve(ClassMap, PoolMap) error { return nil }

func toFloat(p Parseable) (float32, error) {
	x, ok := p.(*Float)
	if !ok {
		return 0, errors.New("not a Float")
	}
	return float32(*x), nil
}

type Int int32

func (i *Int) Parse(r Reader, _ ClassMap, _ PoolMap, _ *ClassMetadata) error {
	x, err := r.VarInt()
	*i = Int(x)
	return err
}

func (*Int) Resolve(ClassMap, PoolMap) error { return nil }

func toInt(p Parseable) (int32, error) {
	x, ok := p.(*Int)
	if !ok {
		return 0, errors.New("not an Int")
	}
	return int32(*x), nil
}

type Long int64

func (l *Long) Parse(r Reader, _ ClassMap, _ PoolMap, _ *ClassMetadata) error {
	x, err := r.VarLong()
	*l = Long(x)
	return err
}

func (*Long) Resolve(ClassMap, PoolMap) error { return nil }

func toLong(p Parseable) (int64, error) {
	x, ok := p.(*Long)
	if !ok {
		return 0, errors.New("not a Long")
	}
	return int64(*x), nil
}

type Short int16

func (s *Short) Parse(r Reader, _ ClassMap, _ PoolMap, _ *ClassMetadata) error {
	x, err := r.VarShort()
	*s = Short(x)
	return err
}

func (*Short) Resolve(ClassMap, PoolMap) error { return nil }

type UShort uint16

func (u *UShort) Parse(r Reader, _ ClassMap, _ PoolMap, _ *ClassMetadata) error {
	x, err := r.VarShort()
	if err != nil {
		return fmt.Errorf("unable to resolve unsigned short: %w", err)
	}
	*u = UShort(x)
	return nil
}

func (*UShort) Resolve(ClassMap, PoolMap) error { return nil }

type Class struct {
	ClassLoader *ClassLoader
	Name        *Symbol
	Package     *Package
	Modifiers   int64
	BaseStructType
}

func (c *Class) setField(name string, p ParseResolvable) (err error) {
	switch name {
	case "classLoader":
		c.ClassLoader, err = toClassLoader(p)
	case "name":
		c.Name, err = toSymbol(p)
	case "package":
		c.Package, err = toPackage(p)
	case "modifers":
		c.Modifiers, err = toLong(p)
	}
	return err
}

func toClass(p ParseResolvable) (*Class, error) {
	c, ok := p.(*Class)
	if !ok {
		// TODO
		return nil, errors.New("")
	}
	return c, nil
}

type String struct {
	s           string
	constantRef *constantReference
}

func (s *String) Parse(r Reader, classMap ClassMap, pools PoolMap, classMetadata *ClassMetadata) error {
	if classMap[classMetadata.ID].Name != "java.lang.String" {
		return fmt.Errorf("expect type of java.lang.String, got type %s", classMap[classMetadata.ID].Name)
	}

	x, err := r.String()
	if err != nil {
		return fmt.Errorf("unable to parse string: %w", err)
	}

	if x.constantRef != nil {
		x.constantRef.classID = classMetadata.ID
	}

	*s = *x
	return nil
}

func (s *String) Resolve(_ ClassMap, poolMap PoolMap) error {
	if s.constantRef != nil {
		cPool := poolMap[s.constantRef.classID]
		if cPool == nil {
			return errors.New("the string constant pool is nil")
		}
		v, ok := cPool.Pool[s.constantRef.index]
		if !ok {
			return fmt.Errorf("string not found in the pool")
		}
		str, ok := v.(*String)
		if !ok {
			return fmt.Errorf("not type of parser.String")
		}
		*s = *str
	}
	return nil
}

func ToString(p Parseable) (string, error) {
	s, ok := p.(*String)
	if !ok {
		return "", errors.New("not a String")
	}
	return s.s, nil
}

type Thread struct {
	BaseStructType
	OsName       string
	OsThreadID   int64
	JavaName     string
	JavaThreadID int64
}

func (t *Thread) setField(name string, p ParseResolvable) (err error) {
	switch name {
	case "osName":
		t.OsName, err = ToString(p)
	case "osThreadId":
		t.OsThreadID, err = toLong(p)
	case "javaName":
		t.JavaName, err = ToString(p)
	case "javaThreadId":
		t.JavaThreadID, err = toLong(p)
	}
	return err
}

func toThread(p ParseResolvable) (*Thread, error) {
	t, ok := p.(*Thread)
	if !ok {
		return nil, errors.New("not a Thread")
	}
	return t, nil
}

type ClassLoader struct {
	Type *Class
	Name *Symbol
	BaseStructType
}

func (cl *ClassLoader) setField(name string, p ParseResolvable) (err error) {
	switch name {
	case "type":
		cl.Type, err = toClass(p)
	case "name":
		cl.Name, err = toSymbol(p)
	}
	return err
}

func toClassLoader(p ParseResolvable) (*ClassLoader, error) {
	c, ok := p.(*ClassLoader)
	if !ok {
		// TODO
		return nil, errors.New("")
	}
	return c, nil
}

type CodeBlobType struct {
	String string
	BaseStructType
}

func (cbt *CodeBlobType) setField(name string, p ParseResolvable) (err error) {
	switch name {
	case "string":
		cbt.String, err = ToString(p)
	}
	return err
}

func toCodeBlobType(p ParseResolvable) (*CodeBlobType, error) {
	cbt, ok := p.(*CodeBlobType)
	if !ok {
		return nil, errors.New("not a CodeBlobType")
	}
	return cbt, nil
}

type FlagValueOrigin struct {
	String string
	BaseStructType
}

func (fvo *FlagValueOrigin) setField(name string, p ParseResolvable) (err error) {
	switch name {
	case "description":
		fvo.String, err = ToString(p)
	}
	return err
}

func toFlagValueOrigin(p Parseable) (*FlagValueOrigin, error) {
	fvo, ok := p.(*FlagValueOrigin)
	if !ok {
		return nil, errors.New("not a FlagValueOrigin")
	}
	return fvo, nil
}

type FrameType struct {
	BaseStructType
	Description string
}

func (ft *FrameType) setField(name string, p ParseResolvable) (err error) {
	switch name {
	case "description":
		ft.Description, err = ToString(p)
	}
	return err
}

func toFrameType(p Parseable) (*FrameType, error) {
	ft, ok := p.(*FrameType)
	if !ok {
		return nil, errors.New("not a FrameType")
	}
	return ft, nil
}

type G1YCType struct {
	String string
	BaseStructType
}

func (gyt *G1YCType) setField(name string, p ParseResolvable) (err error) {
	switch name {
	case "string":
		gyt.String, err = ToString(p)
	}
	return err
}

func toG1YCType(p Parseable) (*G1YCType, error) {
	gyt, ok := p.(*G1YCType)
	if !ok {
		return nil, errors.New("not a G1YCType")
	}
	return gyt, nil
}

type GCName struct {
	String string
	BaseStructType
}

func (gn *GCName) setField(name string, p ParseResolvable) (err error) {
	switch name {
	case "string":
		gn.String, err = ToString(p)
	}
	return err
}

func toGCName(p Parseable) (*GCName, error) {
	gn, ok := p.(*GCName)
	if !ok {
		return nil, errors.New("not a GCName")
	}
	return gn, nil
}

type Method struct {
	Type       *Class
	Name       *Symbol
	Descriptor *Symbol
	Modifiers  int32
	Hidden     bool
	BaseStructType
}

func (m *Method) setField(name string, p ParseResolvable) (err error) {
	switch name {
	case "type":
		m.Type, err = toClass(p)
	case "name":
		m.Name, err = toSymbol(p)
	case "descriptor":
		m.Descriptor, err = toSymbol(p)
	case "modifiers":
		m.Modifiers, err = toInt(p)
	case "hidden":
		m.Hidden, err = toBoolean(p)
	}
	return err
}

func toMethod(p ParseResolvable) (*Method, error) {
	m, ok := p.(*Method)
	if !ok {
		return nil, errors.New("not a Method")
	}
	return m, nil
}

type Module struct {
	Name        *Symbol
	Version     *Symbol
	Location    *Symbol
	ClassLoader *ClassLoader
	BaseStructType
}

func (m *Module) setField(name string, p ParseResolvable) (err error) {
	switch name {
	case "name":
		m.Name, err = toSymbol(p)
	case "version":
		m.Version, err = toSymbol(p)
	case "location":
		m.Location, err = toSymbol(p)
	case "classLoader":
		m.ClassLoader, err = toClassLoader(p)
	}
	return err
}

func toModule(p ParseResolvable) (*Module, error) {
	m, ok := p.(*Module)
	if !ok {
		return nil, errors.New("not a Module")
	}
	return m, nil
}

type NarrowOopMode struct {
	String string
	BaseStructType
}

func (nom *NarrowOopMode) setField(name string, p ParseResolvable) (err error) {
	switch name {
	case "string":
		nom.String, err = ToString(p)
	}
	return err
}

func toNarrowOopMode(p Parseable) (*NarrowOopMode, error) {
	nom, ok := p.(*NarrowOopMode)
	if !ok {
		return nil, errors.New("not a NarrowOopMode")
	}
	return nom, nil
}

type NetworkInterfaceName struct {
	NetworkInterface string
	BaseStructType
}

func (nim *NetworkInterfaceName) setField(name string, p ParseResolvable) (err error) {
	switch name {
	case "networkInterface":
		nim.NetworkInterface, err = ToString(p)
	}
	return err
}

func toNetworkInterfaceName(p Parseable) (*NetworkInterfaceName, error) {
	nim, ok := p.(*NetworkInterfaceName)
	if !ok {
		return nil, errors.New("not a NetworkInterfaceName")
	}
	return nim, nil
}

type Package struct {
	Name *Symbol
	BaseStructType
}

func (pkg *Package) setField(name string, p ParseResolvable) (err error) {
	switch name {
	case "name":
		pkg.Name, err = toSymbol(p)
	}
	return err
}

func toPackage(p ParseResolvable) (*Package, error) {
	pkg, ok := p.(*Package)
	if !ok {
		// TODO
		return nil, errors.New("")
	}
	return pkg, nil
}

type StackFrame struct {
	Method        *Method
	LineNumber    int32
	ByteCodeIndex int32
	Type          *FrameType
	BaseStructType
}

func (sf *StackFrame) setField(name string, p ParseResolvable) (err error) {
	switch name {
	case "method":
		sf.Method, err = toMethod(p)
	case "lineNumber":
		sf.LineNumber, err = toInt(p)
	case "byteCodeIndex":
		sf.ByteCodeIndex, err = toInt(p)
	case "type":
		sf.Type, err = toFrameType(p)
	}
	return err
}

func toStackFrame(p ParseResolvable) (*StackFrame, error) {
	sf, ok := p.(*StackFrame)
	if !ok {
		return nil, errors.New("not a StackFrame")
	}
	return sf, nil
}

type StackTrace struct {
	Truncated bool
	Frames    []*StackFrame
	BaseStructType
}

func (st *StackTrace) setField(name string, p ParseResolvable) (err error) {
	switch name {
	case "truncated":
		st.Truncated, err = toBoolean(p)
	case "frames":
		var sf *StackFrame
		sf, err := toStackFrame(p)
		if err != nil {
			return err
		}
		st.Frames = append(st.Frames, sf)
	}
	return err
}

func toStackTrace(p ParseResolvable) (*StackTrace, error) {
	st, ok := p.(*StackTrace)
	if !ok {
		return nil, errors.New("not a StackTrace")
	}
	return st, nil
}

type Symbol struct {
	String string
	BaseStructType
}

func (s *Symbol) setField(name string, p ParseResolvable) (err error) {
	switch name {
	case "string":
		s.String, err = ToString(p)
	}
	return err
}

func toSymbol(p ParseResolvable) (*Symbol, error) {
	s, ok := p.(*Symbol)
	if !ok {
		// TODO
		return nil, errors.New("")
	}
	return s, nil
}

type ThreadState struct {
	Name string
	BaseStructType
}

func (ts *ThreadState) setField(name string, p ParseResolvable) (err error) {
	switch name {
	case "name":
		ts.Name, err = ToString(p)
	}
	return err
}

func toThreadState(p ParseResolvable) (*ThreadState, error) {
	ts, ok := p.(*ThreadState)
	if !ok {
		return nil, errors.New("not a ThreadState")
	}
	return ts, nil
}

type InflateCause struct {
	BaseStructType
	Cause string
}

func (i *InflateCause) setField(name string, p ParseResolvable) error {
	return setStringField(name, "cause", p, &i.Cause)
}

type GCCause struct {
	BaseStructType
	Cause string
}

func (g *GCCause) setField(name string, p ParseResolvable) error {
	return setStringField(name, "cause", p, &g.Cause)
}

func setStringField(name, expectedFieldName string, p ParseResolvable, ptr *string) (err error) {
	if name != expectedFieldName {
		return
	}
	*ptr, err = ToString(p)
	if err != nil {
		return fmt.Errorf("unable to resolve string from %v(type: %T)", p, p)
	}
	return
}

type CompilerPhaseType struct {
	BaseStructType
	Phase string
}

func (c *CompilerPhaseType) setField(name string, p ParseResolvable) error {
	return setStringField(name, "phase", p, &c.Phase)
}

type ThreadGroup struct {
	BaseStructType
	Name   string
	Parent *ThreadGroup
}

func (t *ThreadGroup) setField(name string, p ParseResolvable) (err error) {
	switch name {
	case "name":
		t.Name, err = ToString(p)
	case "parent":
		t.Parent, err = ToThreadGroup(p)
	}
	return
}

func ToThreadGroup(p ParseResolvable) (*ThreadGroup, error) {
	t, ok := p.(*ThreadGroup)
	if !ok {
		return nil, fmt.Errorf("type *ThreadGroup expected, got %T", p)
	}

	return t, nil
}

type GCThresholdUpdater struct {
	BaseStructType
	Updater string
}

func (g *GCThresholdUpdater) setField(name string, p ParseResolvable) (err error) {
	return setStringField(name, "updater", p, &g.Updater)
}

// MetaspaceObjectType jdk.types.MetaspaceObjectType
type MetaspaceObjectType struct {
	BaseStructType
	Type string
}

func (m *MetaspaceObjectType) setField(name string, p ParseResolvable) (err error) {
	return setStringField(name, "type", p, &m.Type)
}

// ExecutionMode datadog.types.ExecutionMode
type ExecutionMode struct {
	BaseStructType
	Name string
}

func (e *ExecutionMode) setField(name string, p ParseResolvable) (err error) {
	return setStringField(name, "name", p, &e.Name)
}

// VMOperationType jdk.types.VMOperationType
type VMOperationType struct {
	BaseStructType
	Type string
}

func (v *VMOperationType) setField(name string, p ParseResolvable) (err error) {
	return setStringField(name, "type", p, &v.Type)
}

// G1HeapRegionType jdk.types.G1HeapRegionType
type G1HeapRegionType struct {
	BaseStructType
	Type string
}

func (g *G1HeapRegionType) setField(name string, p ParseResolvable) (err error) {
	return setStringField(name, "type", p, &g.Type)
}

// GCWhen jdk.types.GCWhen
type GCWhen struct {
	BaseStructType
	When string
}

func (g *GCWhen) setField(name string, p ParseResolvable) error {
	return setStringField(name, "when", p, &g.When)
}

// ReferenceType jdk.types.ReferenceType
type ReferenceType struct {
	BaseStructType
	Type string
}

func (r *ReferenceType) setField(name string, p ParseResolvable) error {
	return setStringField(name, "type", p, &r.Type)
}

// MetadataType jdk.types.MetadataType
type MetadataType struct {
	BaseStructType
	Type string
}

func (m *MetadataType) setField(name string, p ParseResolvable) error {
	return setStringField(name, "type", p, &m.Type)
}

// LogLevel profiler.types.LogLevel
type LogLevel struct {
	BaseStructType
	Name string
}

func (l *LogLevel) setField(name string, p ParseResolvable) error {
	return setStringField(name, "name", p, &l.Name)
}

// AttributeValue profiler.types.AttributeValue
type AttributeValue struct {
	BaseStructType
	Value string
}

func (a *AttributeValue) setField(name string, p ParseResolvable) error {
	return setStringField(name, "value", p, &a.Value)
}

type ParseResolvableArray []ParseResolvable

func (a ParseResolvableArray) Parse(r Reader, classes ClassMap, cpools PoolMap, class *ClassMetadata) error {
	return nil
}

func (a ParseResolvableArray) Resolve(classes ClassMap, cpools PoolMap) error {
	for _, resolvable := range a {
		if err := resolvable.Resolve(classes, cpools); err != nil {
			return err
		}
	}
	return nil
}

// DefaultStructType represents any type that is not supported by the parser.
// This will allow to still read the unsupported type instead of returning an error.
type DefaultStructType struct {
	BaseStructType
	className  string
	fieldsDict map[string]*FieldMetadata
	fields     map[string]ParseResolvable
}

func (d *DefaultStructType) setField(name string, p ParseResolvable) error {
	if d.fields == nil {
		d.fields = make(map[string]ParseResolvable)
	}

	if d.fieldsDict[name].IsArray() {
		if d.fields[name] == nil {
			d.fields[name] = make(ParseResolvableArray, 0, 1)
		}
		d.fields[name] = append(d.fields[name].(ParseResolvableArray), p)
	} else {
		d.fields[name] = p
	}

	return nil
}
