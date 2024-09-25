package parser

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
)

const (
	MetadataEventType     = 0
	ConstantPoolEventType = 1

	EventSuperType = "jdk.jfr.Event"
)

var magic = []byte{'F', 'L', 'R', 0}

type Version struct {
	Major uint16
	Minor uint16
}

type CPool struct {
	Pool     map[int64]ParseResolvable
	resolved bool
}
type ClassMap map[int64]*ClassMetadata
type PoolMap map[int64]*CPool
type ChunkEvents map[string]*EventCollection

func (c ChunkEvents) Apply(filter EventFilter) []*GenericEvent {
	var filtered []*GenericEvent

	for _, collection := range c {
		predicate := filter.GetPredicate(collection.ClassMetadata)
		if IsAlwaysFalse(predicate) {
			continue
		} else if IsAlwaysTrue(predicate) {
			filtered = append(filtered, collection.Events...)
		} else {
			for _, event := range collection.Events {
				if predicate.Test(event) {
					filtered = append(filtered, event)
				}
			}
		}
	}
	return filtered
}

type Chunk struct {
	Header   Header
	Metadata ChunkMetadata
	ChunkEvents
}

type EventCollection struct {
	ClassMetadata *ClassMetadata
	Events        []*GenericEvent
}

func (c *EventCollection) Add(e *GenericEvent) {
	c.Events = append(c.Events, e)
}

type ChunkParseOptions struct {
	CPoolProcessor func(meta *ClassMetadata, cpool *CPool)
}

func (c *Chunk) addEvent(e *GenericEvent) {
	if c.ChunkEvents == nil {
		c.ChunkEvents = make(ChunkEvents)
	}

	classMeta := c.Metadata.ClassMap[e.ClassID]

	ec, ok := c.ChunkEvents[classMeta.Name]
	if !ok {
		ec = &EventCollection{
			ClassMetadata: classMeta,
		}
		c.ChunkEvents[classMeta.Name] = ec
	}
	ec.Add(e)
}

func (c *Chunk) ShowClassMeta(name string) {
	for _, classMeta := range c.Metadata.ClassMap {
		if classMeta.Name == name {
			fmt.Printf("simple type: %t, super type: %s\n",
				classMeta.SimpleType, classMeta.SuperType)

			for _, field := range classMeta.Fields {
				fmt.Printf("field name: %s, field label: %s, field class: %s, field description: %s, field constant pool: %t, field is array: %t, field unsigned: %t, field unit: %+#v\n",
					field.Name, field.Label(c.Metadata.ClassMap), c.Metadata.ClassMap[field.ClassID].Name, field.Description(c.Metadata.ClassMap),
					field.ConstantPool, field.IsArray(), field.Unsigned(c.Metadata.ClassMap), field.Unit(c.Metadata.ClassMap))
			}
			break
		}
	}
}

func (c *Chunk) Parse(r io.Reader, options *ChunkParseOptions) (err error) {
	bufR, ok := r.(*bufio.Reader)
	if !ok {
		bufR = bufio.NewReader(r)
	}

	buf := make([]byte, len(magic))
	if _, err = io.ReadFull(bufR, buf); err != nil {
		if err == io.EOF {
			return err
		}
		return fmt.Errorf("unable to read chunk's header: %w", err)
	}

	// TODO magic header
	for i, r := range magic {
		if r != buf[i] {
			return fmt.Errorf("unexpected magic header %v expected, %v found", magic, buf)
		}
	}
	if _, err = io.ReadFull(bufR, buf); err != nil {
		return fmt.Errorf("unable to read format version: %w", err)
	}

	// TODO Check supported major / minor

	buf = make([]byte, headerSize)
	if _, err = io.ReadFull(bufR, buf); err != nil {
		return fmt.Errorf("unable to read chunk header: %w", err)
	}
	if err = c.Header.Parse(NewReader(bytes.NewReader(buf), false)); err != nil {
		return fmt.Errorf("unable to parse chunk header: %w", err)
	}
	c.Header.ChunkSize -= headerSize + 8
	c.Header.MetadataOffset -= headerSize + 8
	c.Header.ConstantPoolOffset -= headerSize + 8
	useCompression := c.Header.Features&1 == 1
	// TODO: assert c.Header.ChunkSize is small enough
	buf = make([]byte, c.Header.ChunkSize)
	if _, err := io.ReadFull(r, buf); err != nil {
		return fmt.Errorf("unable to read chunk contents: %w", err)
	}

	br := bytes.NewReader(buf)
	rd := NewReader(br, useCompression)
	pointer := int64(0)
	eventsOffset := make(map[int64]int32)

	// Parse metadata
	if _, err := br.Seek(c.Header.MetadataOffset, io.SeekStart); err != nil {
		return fmt.Errorf("unable to seek reader: %w", err)
	}
	metadataSize, err := rd.VarInt()
	if err != nil {
		return fmt.Errorf("unable to parse chunk metadata size: %w", err)
	}
	eventsOffset[c.Header.MetadataOffset] = metadataSize

	c.Metadata.Header = &c.Header // for resolving class fields unit ticks

	if err := c.Metadata.Parse(rd); err != nil {
		return fmt.Errorf("unable to parse chunk metadata: %w", err)
	}

	// Parse checkpoint event(s)
	if _, err := br.Seek(c.Header.ConstantPoolOffset, io.SeekStart); err != nil {
		return fmt.Errorf("unable to seek reader: %w", err)
	}
	constantPoolSize := int32(0)
	cpools := make(PoolMap)
	delta := int64(0)
	cp := new(ConstantPoolEvent)
	for {
		size, err := rd.VarInt()
		if err != nil {
			return fmt.Errorf("unable to parse checkpoint event size: %w", err)
		}
		eventsOffset[c.Header.ConstantPoolOffset+delta] = size
		constantPoolSize += size
		if err := cp.Parse(rd, c.Metadata.ClassMap, cpools); err != nil {
			return fmt.Errorf("unable to parse constant pool event: %w", err)
		}
		if cp.Delta == 0 {
			break
		}
		delta += cp.Delta
		if _, err := br.Seek(c.Header.ConstantPoolOffset+delta, io.SeekStart); err != nil {
			return fmt.Errorf("unable to seek reader: %w", err)
		}
	}

	if options.CPoolProcessor != nil {
		for classID, pool := range cpools {
			options.CPoolProcessor(c.Metadata.ClassMap[classID], pool)
		}
	}

	// Second pass over constant pools: resolve constants
	if err = ResolveConstants(c.Metadata.ClassMap, cpools); err != nil {
		return err
	}

	// Parse the rest of events
	if _, err := br.Seek(pointer, io.SeekStart); err != nil {
		return fmt.Errorf("unable to seek reader: %w", err)
	}
	for pointer != c.Header.ChunkSize {
		if size, ok := eventsOffset[pointer]; ok {
			pointer += int64(size)
		} else {
			if _, err := br.Seek(pointer, io.SeekStart); err != nil {
				return fmt.Errorf("unable to seek to position %d: %w", pointer, err)
			}
			size, err := rd.VarInt()
			if err != nil {
				return fmt.Errorf("unable to parse event size: %w", err)
			}
			if size == 0 {
				return fmt.Errorf("found event with invalid size (0)")
			}
			eventsOffset[pointer] = size
			ge, err := ParseEvent(rd, c.Metadata.ClassMap, cpools)
			if err != nil {
				return fmt.Errorf("unable to parse event: %w", err)
			}
			if ge != nil {
				c.addEvent(ge)
			}
			pointer += int64(size)
		}
	}
	return nil
}

func ResolveConstants(classes ClassMap, poolMap PoolMap) (err error) {

	for classID, pool := range poolMap {
		if pool.resolved {
			continue
		}
		if classes[classID] != nil && classes[classID].SuperType == EventSuperType {
			continue
		}

		for _, t := range pool.Pool {
			if err = t.Resolve(classes, poolMap); err != nil {
				return fmt.Errorf("unable to resolve constants: %w", err)
			}
		}
		pool.resolved = true
	}

	return nil
}
