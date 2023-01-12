package parser

import (
	"fmt"

	"github.com/pyroscope-io/jfr-parser/reader"
)

type CheckpointEvent struct {
	StartTime int64
	Duration  int64
	Delta     int64
	TypeMask  int8
}

func (c *CheckpointEvent) Parse(r reader.Reader, classes ClassMap, cpools PoolMap) (err error) {
	if kind, err := r.VarLong(); err != nil {
		return fmt.Errorf("unable to retrieve event type: %w", err)
	} else if kind != 1 {
		return fmt.Errorf("unexpected checkpoint event type: %d", kind)
	}
	if c.StartTime, err = r.VarLong(); err != nil {
		return fmt.Errorf("unable to parse checkpoint event's start time: %w", err)
	}
	if c.Duration, err = r.VarLong(); err != nil {
		return fmt.Errorf("unable to parse checkpoint event's duration: %w", err)
	}
	if c.Delta, err = r.VarLong(); err != nil {
		return fmt.Errorf("unable to parse checkpoint event's delta: %w", err)
	}
	c.TypeMask, _ = r.Byte()
	n, err := r.VarInt()
	if err != nil {
		return fmt.Errorf("unable to parse checkpoint event's number of constant pools: %w", err)
	}
	// TODO: assert n is small enough
	for i := 0; i < int(n); i++ {
		classID, err := r.VarLong()
		if err != nil {
			return fmt.Errorf("unable to parse constant pool class: %w", err)
		}
		cm, ok := cpools[int(classID)]
		if !ok {
			cpools[int(classID)] = &CPool{Pool: make(map[int]ParseResolvable)}
			cm = cpools[int(classID)]
		}
		m, err := r.VarInt()
		if err != nil {
			return fmt.Errorf("unable to parse constant pool's number of constants: %w", err)
		}
		// TODO: assert m is small enough
		for j := 0; j < int(m); j++ {
			idx, err := r.VarLong()
			if err != nil {
				return fmt.Errorf("unable to parse contant's index: %w", err)
			}
			v, err := ParseClass(r, classes, cpools, classID)
			if err != nil {
				return fmt.Errorf("unable to parse constant type %d: %w", classID, err)
			}
			cm.Pool[int(idx)] = v
		}
	}
	return nil
}
