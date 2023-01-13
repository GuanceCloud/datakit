package parser

import (
	"bytes"
	"fmt"
	"io"

	"github.com/pyroscope-io/jfr-parser/reader"
)

var magic = []byte{'F', 'L', 'R', 0}

type CPool struct {
	Pool     map[int]ParseResolvable
	resolved bool
}
type ClassMap map[int]ClassMetadata
type PoolMap map[int]*CPool

type Chunk struct {
	Header      Header
	Metadata    MetadataEvent
	Checkpoints []CheckpointEvent
	Events      []Parseable
}

type ChunkParseOptions struct {
	CPoolProcessor func(meta ClassMetadata, cpool *CPool)
}

func (c *Chunk) Parse(r io.Reader, options *ChunkParseOptions) (err error) {
	buf := make([]byte, len(magic))
	if _, err = io.ReadFull(r, buf); err != nil {
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
	if _, err = io.ReadFull(r, buf); err != nil {
		return fmt.Errorf("unable to read format version: %w", err)
	}
	// TODO Check supported major / minor

	buf = make([]byte, headerSize)
	if _, err := io.ReadFull(r, buf); err != nil {
		return fmt.Errorf("unable to read chunk header: %w", err)
	}
	if err := c.Header.Parse(reader.NewReader(bytes.NewReader(buf), false)); err != nil {
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
	rd := reader.NewReader(br, useCompression)
	pointer := int64(0)
	events := make(map[int64]int32)

	// Parse metadata
	br.Seek(c.Header.MetadataOffset, io.SeekStart)
	metadataSize, err := rd.VarInt()
	if err != nil {
		return fmt.Errorf("unable to parse chunk metadata size: %w", err)
	}
	events[c.Header.MetadataOffset] = metadataSize
	var metadata MetadataEvent
	if err := metadata.Parse(rd); err != nil {
		return fmt.Errorf("unable to parse chunk metadata: %w", err)
	}
	classes := buildClasses(metadata)

	// Parse checkpoint event(s)
	br.Seek(c.Header.ConstantPoolOffset, io.SeekStart)
	checkpointsSize := int32(0)
	cpools := make(PoolMap)
	delta := int64(0)
	for {
		size, err := rd.VarInt()
		if err != nil {
			return fmt.Errorf("unable to parse checkpoint event size: %w", err)
		}
		events[c.Header.ConstantPoolOffset+delta] = size
		checkpointsSize += size
		var cp CheckpointEvent
		if err := cp.Parse(rd, classes, cpools); err != nil {
			return fmt.Errorf("unable to parse checkpoint event: %w", err)
		}
		c.Checkpoints = append(c.Checkpoints, cp)
		if cp.Delta == 0 {
			break
		}
		delta += cp.Delta
		br.Seek(c.Header.ConstantPoolOffset+delta, io.SeekStart)
	}

	if options.CPoolProcessor != nil {
		for classID, pool := range cpools {
			options.CPoolProcessor(classes[classID], pool)
		}
	}

	// Second pass over constant pools: resolve constants
	for classID := range cpools {
		if err := ResolveConstants(classes, cpools, classID); err != nil {
			return err
		}
	}

	// Parse the rest of events
	br.Seek(pointer, io.SeekStart)
	for pointer != c.Header.ChunkSize {
		if size, ok := events[pointer]; ok {
			pointer += int64(size)
		} else {
			if _, err := br.Seek(pointer, io.SeekStart); err != nil {
				return fmt.Errorf("unable to seek to position %d: %w", pointer, err)
			}
			size, err := rd.VarInt()
			if err != nil {
				return fmt.Errorf("unable to parse event size: %w", err)
			}
			events[pointer] = size
			e, err := ParseEvent(rd, classes, cpools)
			if err != nil {
				return fmt.Errorf("unable to parse event: %w", err)
			}
			c.Events = append(c.Events, e)
			pointer += int64(size)
		}
	}
	return nil
}

func buildClasses(metadata MetadataEvent) ClassMap {
	classes := make(map[int]ClassMetadata)
	for _, class := range metadata.Root.Metadata.Classes {
		classes[int(class.ID)] = class
	}
	return classes
}

func ResolveConstants(classes ClassMap, cpools PoolMap, classID int) (err error) {
	cpool, ok := cpools[classID]
	if !ok {
		// Non-existent constant pool references seem to be used to mark no value
		return nil
	}
	if cpool.resolved {
		return nil
	}
	cpool.resolved = true
	for _, t := range cpool.Pool {
		if err := t.Resolve(classes, cpools); err != nil {
			return fmt.Errorf("unable to resolve constants: %w", err)
		}
	}
	return nil
}
