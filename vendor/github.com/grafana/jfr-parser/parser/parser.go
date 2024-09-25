package parser

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"unsafe"

	types2 "github.com/grafana/jfr-parser/parser/types"
	"github.com/grafana/jfr-parser/parser/types/def"
)

func ParseFile(p string) ([]*Chunk, error) {
	f, err := os.Open(p)
	if err != nil {
		return nil, fmt.Errorf("unable to open file [%s]: %w", p, err)
	}
	defer f.Close()
	return Parse(f)
}

func Parse(r io.Reader) ([]*Chunk, error) {
	rc, err := Decompress(r)
	if err != nil {
		return nil, fmt.Errorf("unable to decompress input stream: %w", err)
	}
	defer rc.Close()
	return ParseWithOptions(bufio.NewReader(rc), &ChunkParseOptions{})
}

func ParseWithOptions(r io.Reader, options *ChunkParseOptions) ([]*Chunk, error) {
	var chunks []*Chunk
	for {
		chunk := new(Chunk)
		err := chunk.Parse(r, options)
		if err == io.EOF {
			return chunks, nil
		}
		if err != nil {
			return chunks, fmt.Errorf("unable to parse chunk: %w", err)
		}
		chunks = append(chunks, chunk)
	}
}

const chunkHeaderSize = 68
const bufferSize = 1024 * 1024
const chunkMagic = 0x464c5200

type ChunkHeader struct {
	Magic              uint32
	Version            uint32
	Size               int
	OffsetConstantPool int
	OffsetMeta         int
	StartNanos         uint64
	DurationNanos      uint64
	StartTicks         uint64
	TicksPerSecond     uint64
	Features           uint32
}

func (c *ChunkHeader) String() string {
	return fmt.Sprintf("ChunkHeader{Magic: %x, Version: %x, Size: %d, OffsetConstantPool: %d, OffsetMeta: %d, StartNanos: %d, DurationNanos: %d, StartTicks: %d, TicksPerSecond: %d, Features: %d}", c.Magic, c.Version, c.Size, c.OffsetConstantPool, c.OffsetMeta, c.StartNanos, c.DurationNanos, c.StartTicks, c.TicksPerSecond, c.Features)
}

type SymbolProcessor func(ref *types2.SymbolList)

type Options struct {
	ChunkSizeLimit  int
	SymbolProcessor SymbolProcessor
}

type Parser struct {
	FrameTypes   types2.FrameTypeList
	ThreadStates types2.ThreadStateList
	Threads      types2.ThreadList
	Classes      types2.ClassList
	Methods      types2.MethodList
	Packages     types2.PackageList
	Symbols      types2.SymbolList
	LogLevels    types2.LogLevelList
	Stacktrace   types2.StackTraceList

	ExecutionSample             types2.ExecutionSample
	ObjectAllocationInNewTLAB   types2.ObjectAllocationInNewTLAB
	ObjectAllocationOutsideTLAB types2.ObjectAllocationOutsideTLAB
	JavaMonitorEnter            types2.JavaMonitorEnter
	ThreadPark                  types2.ThreadPark
	LiveObject                  types2.LiveObject
	ActiveSetting               types2.ActiveSetting

	header   ChunkHeader
	options  Options
	buf      []byte
	pos      int
	metaSize uint32
	chunkEnd int

	TypeMap def.TypeMap

	bindFrameType   *types2.BindFrameType
	bindThreadState *types2.BindThreadState
	bindThread      *types2.BindThread
	bindClass       *types2.BindClass
	bindMethod      *types2.BindMethod
	bindPackage     *types2.BindPackage
	bindSymbol      *types2.BindSymbol
	bindLogLevel    *types2.BindLogLevel
	bindStackFrame  *types2.BindStackFrame
	bindStackTrace  *types2.BindStackTrace

	bindExecutionSample *types2.BindExecutionSample

	bindAllocInNewTLAB   *types2.BindObjectAllocationInNewTLAB
	bindAllocOutsideTLAB *types2.BindObjectAllocationOutsideTLAB
	bindMonitorEnter     *types2.BindJavaMonitorEnter
	bindThreadPark       *types2.BindThreadPark
	bindLiveObject       *types2.BindLiveObject
	bindActiveSetting    *types2.BindActiveSetting
}

func NewParser(buf []byte, options Options) *Parser {
	p := &Parser{
		options: options,
		buf:     buf,
	}
	return p
}

func (p *Parser) ParseEvent() (def.TypeID, error) {
	for {
		if p.pos == p.chunkEnd {
			if p.pos == len(p.buf) {
				return 0, io.EOF
			}
			if err := p.readChunk(p.pos); err != nil {
				return 0, err
			}
		}
		pp := p.pos
		size, err := p.varLong()
		if err != nil {
			return 0, err
		}
		if size == 0 {
			return 0, def.ErrIntOverflow
		}
		typ, err := p.varLong()
		if err != nil {
			return 0, err
		}
		_ = size

		ttyp := def.TypeID(typ)
		switch ttyp {
		case p.TypeMap.T_EXECUTION_SAMPLE:
			if p.bindExecutionSample == nil {
				p.pos = pp + int(size) // skip
				continue
			}
			_, err := p.ExecutionSample.Parse(p.buf[p.pos:], p.bindExecutionSample, &p.TypeMap)
			if err != nil {
				return 0, err
			}
			p.pos = pp + int(size)
			return ttyp, nil
		case p.TypeMap.T_ALLOC_IN_NEW_TLAB:
			if p.bindAllocInNewTLAB == nil {
				p.pos = pp + int(size) // skip
				continue
			}
			_, err := p.ObjectAllocationInNewTLAB.Parse(p.buf[p.pos:], p.bindAllocInNewTLAB, &p.TypeMap)
			if err != nil {
				return 0, err
			}
			p.pos = pp + int(size)
			return ttyp, nil
		case p.TypeMap.T_ALLOC_OUTSIDE_TLAB:
			if p.bindAllocOutsideTLAB == nil {
				p.pos = pp + int(size) // skip
				continue
			}
			_, err := p.ObjectAllocationOutsideTLAB.Parse(p.buf[p.pos:], p.bindAllocOutsideTLAB, &p.TypeMap)
			if err != nil {
				return 0, err
			}
			p.pos = pp + int(size)
			return ttyp, nil
		case p.TypeMap.T_LIVE_OBJECT:
			if p.bindLiveObject == nil {
				p.pos = pp + int(size) // skip
				continue
			}
			_, err := p.LiveObject.Parse(p.buf[p.pos:], p.bindLiveObject, &p.TypeMap)
			if err != nil {
				return 0, err
			}
			p.pos = pp + int(size)
			return ttyp, nil
		case p.TypeMap.T_MONITOR_ENTER:
			if p.bindMonitorEnter == nil {
				p.pos = pp + int(size) // skip
				continue
			}
			_, err := p.JavaMonitorEnter.Parse(p.buf[p.pos:], p.bindMonitorEnter, &p.TypeMap)
			if err != nil {
				return 0, err
			}
			p.pos = pp + int(size)
			return ttyp, nil
		case p.TypeMap.T_THREAD_PARK:
			if p.bindThreadPark == nil {
				p.pos = pp + int(size) // skip
				continue
			}
			_, err := p.ThreadPark.Parse(p.buf[p.pos:], p.bindThreadPark, &p.TypeMap)
			if err != nil {
				return 0, err
			}
			p.pos = pp + int(size)
			return ttyp, nil

		case p.TypeMap.T_ACTIVE_SETTING:
			if p.bindActiveSetting == nil {
				p.pos = pp + int(size) // skip
				continue
			}
			_, err := p.ActiveSetting.Parse(p.buf[p.pos:], p.bindActiveSetting, &p.TypeMap)
			if err != nil {
				return 0, err
			}
			p.pos = pp + int(size)
			return ttyp, nil
		default:
			//fmt.Printf("skipping %s %v\n", def.TypeID2Sym(ttyp), ttyp)
			p.pos = pp + int(size)
		}
	}
}

func (p *Parser) ChunkHeader() ChunkHeader {
	return p.header
}

func (p *Parser) GetStacktrace(stID types2.StackTraceRef) *types2.StackTrace {
	idx, ok := p.Stacktrace.IDMap[stID]
	if !ok {
		return nil
	}
	return &p.Stacktrace.StackTrace[idx]
}

func (p *Parser) GetThreadState(ref types2.ThreadStateRef) *types2.ThreadState {
	idx, ok := p.ThreadStates.IDMap[ref]
	if !ok {
		return nil
	}
	return &p.ThreadStates.ThreadState[idx]
}

func (p *Parser) GetMethod(mID types2.MethodRef) *types2.Method {
	if mID == 0 {
		return nil
	}
	var idx int

	refIDX := int(mID)
	if refIDX < len(p.Methods.IDMap.Slice) {
		idx = int(p.Methods.IDMap.Slice[mID])
	} else {
		idx = p.Methods.IDMap.Get(mID)
	}

	if idx == -1 {
		return nil
	}
	return &p.Methods.Method[idx]
}

func (p *Parser) GetClass(cID types2.ClassRef) *types2.Class {
	idx, ok := p.Classes.IDMap[cID]
	if !ok {
		return nil
	}
	return &p.Classes.Class[idx]
}

func (p *Parser) GetSymbol(sID types2.SymbolRef) *types2.Symbol {
	idx, ok := p.Symbols.IDMap[sID]
	if !ok {
		return nil
	}
	return &p.Symbols.Symbol[idx]
}

func (p *Parser) GetSymbolString(sID types2.SymbolRef) string {
	idx, ok := p.Symbols.IDMap[sID]
	if !ok {
		return ""
	}
	return p.Symbols.Symbol[idx].String
}

func (p *Parser) readChunk(pos int) error {
	if err := p.readChunkHeader(pos); err != nil {
		return fmt.Errorf("error reading chunk header: %w", err)
	}

	if err := p.readMeta(pos + p.header.OffsetMeta); err != nil {
		return fmt.Errorf("error reading metadata: %w", err)
	}
	if err := p.readConstantPool(pos + p.header.OffsetConstantPool); err != nil {
		return fmt.Errorf("error reading CP: %w @ %d", err, pos+p.header.OffsetConstantPool)
	}
	pp := p.options.SymbolProcessor
	if pp != nil {
		pp(&p.Symbols)
	}
	p.pos = pos + chunkHeaderSize
	return nil
}

func (p *Parser) seek(pos int) error {
	if pos < len(p.buf) {
		p.pos = pos
		return nil
	}
	return io.ErrUnexpectedEOF
}

func (p *Parser) byte() (byte, error) {
	if p.pos >= len(p.buf) {
		return 0, io.ErrUnexpectedEOF
	}
	b := p.buf[p.pos]
	p.pos++
	return b, nil
}
func (p *Parser) varInt() (uint32, error) {
	v := uint32(0)
	for shift := uint(0); ; shift += 7 {
		if shift >= 32 {
			return 0, def.ErrIntOverflow
		}
		if p.pos >= len(p.buf) {
			return 0, io.ErrUnexpectedEOF
		}
		b := p.buf[p.pos]
		p.pos++
		v |= uint32(b&0x7F) << shift
		if b < 0x80 {
			break
		}
	}
	return v, nil
}

func (p *Parser) varLong() (uint64, error) {
	v64_ := uint64(0)
	for shift := uint(0); shift <= 56; shift += 7 {
		if p.pos >= len(p.buf) {
			return 0, io.ErrUnexpectedEOF
		}
		b_ := p.buf[p.pos]
		p.pos++
		if shift == 56 {
			v64_ |= uint64(b_&0xFF) << shift
			break
		} else {
			v64_ |= uint64(b_&0x7F) << shift
			if b_ < 0x80 {
				break
			}
		}
	}
	return v64_, nil
}

func (p *Parser) string() (string, error) {
	if p.pos >= len(p.buf) {
		return "", io.ErrUnexpectedEOF
	}
	b := p.buf[p.pos]
	p.pos++
	switch b { //todo implement 2
	case 0:
		return "", nil //todo this should be nil
	case 1:
		return "", nil
	case 3:
		bs, err := p.bytes()
		if err != nil {
			return "", err
		}
		str := *(*string)(unsafe.Pointer(&bs))
		return str, nil
	case 4:
		return p.charArrayString()
	default:
		return "", fmt.Errorf("unknown string type %d", b)
	}

}

func (p *Parser) charArrayString() (string, error) {
	l, err := p.varInt()
	if err != nil {
		return "", err
	}
	if l < 0 {
		return "", def.ErrIntOverflow
	}
	buf := make([]rune, int(l))
	for i := 0; i < int(l); i++ {
		c, err := p.varInt()
		if err != nil {
			return "", err
		}
		buf[i] = rune(c)
	}

	res := string(buf)
	return res, nil
}

func (p *Parser) bytes() ([]byte, error) {
	l, err := p.varInt()
	if err != nil {
		return nil, err
	}
	if l < 0 {
		return nil, def.ErrIntOverflow
	}
	if p.pos+int(l) > len(p.buf) {
		return nil, io.ErrUnexpectedEOF
	}
	bs := p.buf[p.pos : p.pos+int(l)]
	p.pos += int(l)
	return bs, nil
}

func (p *Parser) checkTypes() error {

	tint := p.TypeMap.NameMap["int"]
	tlong := p.TypeMap.NameMap["long"]
	tfloat := p.TypeMap.NameMap["float"]
	tboolean := p.TypeMap.NameMap["boolean"]
	tstring := p.TypeMap.NameMap["java.lang.String"]

	if tint == nil {
		return fmt.Errorf("missing \"int\"")
	}
	if tlong == nil {
		return fmt.Errorf("missing \"long\"")
	}
	if tfloat == nil {
		return fmt.Errorf("missing \"float\"")
	}
	if tboolean == nil {
		return fmt.Errorf("missing \"boolean\"")
	}
	if tstring == nil {
		return fmt.Errorf("missing \"java.lang.String\"")
	}
	p.TypeMap.T_INT = tint.ID
	p.TypeMap.T_LONG = tlong.ID
	p.TypeMap.T_FLOAT = tfloat.ID
	p.TypeMap.T_BOOLEAN = tboolean.ID
	p.TypeMap.T_STRING = tstring.ID

	typeCPFrameType := p.TypeMap.NameMap["jdk.types.FrameType"]
	typeCPThreadState := p.TypeMap.NameMap["jdk.types.ThreadState"]
	typeCPThread := p.TypeMap.NameMap["java.lang.Thread"]
	typeCPClass := p.TypeMap.NameMap["java.lang.Class"]
	typeCPMethod := p.TypeMap.NameMap["jdk.types.Method"]
	typeCPPackage := p.TypeMap.NameMap["jdk.types.Package"]
	typeCPSymbol := p.TypeMap.NameMap["jdk.types.Symbol"]
	typeCPLogLevel := p.TypeMap.NameMap["profiler.types.LogLevel"]
	typeCPStackTrace := p.TypeMap.NameMap["jdk.types.StackTrace"]
	typeCPClassLoader := p.TypeMap.NameMap["jdk.types.ClassLoader"]

	if typeCPFrameType == nil {
		return fmt.Errorf("missing \"jdk.types.FrameType\"")
	}
	if typeCPThreadState == nil {
		return fmt.Errorf("missing \"jdk.types.ThreadState\"")
	}
	if typeCPThread == nil {
		return fmt.Errorf("missing \"java.lang.Thread\"")
	}
	if typeCPClass == nil {
		return fmt.Errorf("missing \"java.lang.Class\"")
	}
	if typeCPMethod == nil {
		return fmt.Errorf("missing \"jdk.types.Method\"")
	}
	if typeCPPackage == nil {
		return fmt.Errorf("missing \"jdk.types.Package\"")
	}
	if typeCPSymbol == nil {
		return fmt.Errorf("missing \"jdk.types.Symbol\"")
	}
	if typeCPStackTrace == nil {
		return fmt.Errorf("missing \"jdk.types.StackTrace\"")
	}
	if typeCPClassLoader == nil {
		return fmt.Errorf("missing \"jdk.types.ClassLoader\"")
	}
	p.TypeMap.T_FRAME_TYPE = typeCPFrameType.ID
	p.TypeMap.T_THREAD_STATE = typeCPThreadState.ID
	p.TypeMap.T_THREAD = typeCPThread.ID
	p.TypeMap.T_CLASS = typeCPClass.ID
	p.TypeMap.T_METHOD = typeCPMethod.ID
	p.TypeMap.T_PACKAGE = typeCPPackage.ID
	p.TypeMap.T_SYMBOL = typeCPSymbol.ID
	if typeCPLogLevel != nil {
		p.TypeMap.T_LOG_LEVEL = typeCPLogLevel.ID
	} else {
		p.TypeMap.T_LOG_LEVEL = 0
	}
	p.TypeMap.T_STACK_TRACE = typeCPStackTrace.ID
	p.TypeMap.T_CLASS_LOADER = typeCPClassLoader.ID

	typeStackFrame := p.TypeMap.NameMap["jdk.types.StackFrame"]

	if typeStackFrame == nil {
		return fmt.Errorf("missing \"jdk.types.StackFrame\"")
	}
	p.TypeMap.T_STACK_FRAME = typeStackFrame.ID

	p.bindFrameType = types2.NewBindFrameType(typeCPFrameType, &p.TypeMap)
	p.bindThreadState = types2.NewBindThreadState(typeCPThreadState, &p.TypeMap)
	p.bindThread = types2.NewBindThread(typeCPThread, &p.TypeMap)
	p.bindClass = types2.NewBindClass(typeCPClass, &p.TypeMap)
	p.bindMethod = types2.NewBindMethod(typeCPMethod, &p.TypeMap)
	p.bindPackage = types2.NewBindPackage(typeCPPackage, &p.TypeMap)
	p.bindSymbol = types2.NewBindSymbol(typeCPSymbol, &p.TypeMap)
	if typeCPLogLevel != nil {
		p.bindLogLevel = types2.NewBindLogLevel(typeCPLogLevel, &p.TypeMap)
	} else {
		p.bindLogLevel = nil
	}
	p.bindStackTrace = types2.NewBindStackTrace(typeCPStackTrace, &p.TypeMap)
	p.bindStackFrame = types2.NewBindStackFrame(typeStackFrame, &p.TypeMap)

	typeExecutionSample := p.TypeMap.NameMap["jdk.ExecutionSample"]
	typeAllocInNewTLAB := p.TypeMap.NameMap["jdk.ObjectAllocationInNewTLAB"]
	typeALlocOutsideTLAB := p.TypeMap.NameMap["jdk.ObjectAllocationOutsideTLAB"]
	typeMonitorEnter := p.TypeMap.NameMap["jdk.JavaMonitorEnter"]
	typeThreadPark := p.TypeMap.NameMap["jdk.ThreadPark"]
	typeLiveObject := p.TypeMap.NameMap["profiler.LiveObject"]
	typeActiveSetting := p.TypeMap.NameMap["jdk.ActiveSetting"]

	if typeExecutionSample != nil {
		p.TypeMap.T_EXECUTION_SAMPLE = typeExecutionSample.ID
		p.bindExecutionSample = types2.NewBindExecutionSample(typeExecutionSample, &p.TypeMap)
	}
	if typeAllocInNewTLAB != nil {
		p.TypeMap.T_ALLOC_IN_NEW_TLAB = typeAllocInNewTLAB.ID
		p.bindAllocInNewTLAB = types2.NewBindObjectAllocationInNewTLAB(typeAllocInNewTLAB, &p.TypeMap)
	}
	if typeALlocOutsideTLAB != nil {
		p.TypeMap.T_ALLOC_OUTSIDE_TLAB = typeALlocOutsideTLAB.ID
		p.bindAllocOutsideTLAB = types2.NewBindObjectAllocationOutsideTLAB(typeALlocOutsideTLAB, &p.TypeMap)
	}
	if typeMonitorEnter != nil {
		p.TypeMap.T_MONITOR_ENTER = typeMonitorEnter.ID
		p.bindMonitorEnter = types2.NewBindJavaMonitorEnter(typeMonitorEnter, &p.TypeMap)
	}
	if typeThreadPark != nil {
		p.TypeMap.T_THREAD_PARK = typeThreadPark.ID
		p.bindThreadPark = types2.NewBindThreadPark(typeThreadPark, &p.TypeMap)
	}
	if typeLiveObject != nil {
		p.TypeMap.T_LIVE_OBJECT = typeLiveObject.ID
		p.bindLiveObject = types2.NewBindLiveObject(typeLiveObject, &p.TypeMap)
	}
	if typeActiveSetting != nil {
		p.TypeMap.T_ACTIVE_SETTING = typeActiveSetting.ID
		p.bindActiveSetting = types2.NewBindActiveSetting(typeActiveSetting, &p.TypeMap)
	}

	p.FrameTypes.IDMap = nil
	p.ThreadStates.IDMap = nil
	p.Threads.IDMap = nil
	p.Classes.IDMap = nil
	p.Methods.IDMap.Slice = nil
	p.Packages.IDMap = nil
	p.Symbols.IDMap = nil
	p.LogLevels.IDMap = nil
	p.Stacktrace.IDMap = nil
	return nil
}
