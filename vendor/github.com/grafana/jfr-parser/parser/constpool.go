package parser

import (
	"fmt"
)

type ConstantPoolEvent struct {
	StartTime int64
	Duration  int64
	Delta     int64
	TypeMask  int8
}

func (c *ConstantPoolEvent) Parse(r Reader, classes ClassMap, cpools PoolMap) (err error) {
	eventType, err := r.VarLong()
	if err != nil {
		return fmt.Errorf("unable to retrieve event type: %w", err)
	}
	if eventType != ConstantPoolEventType {
		return fmt.Errorf("unexpected checkpoint event type: %d", eventType)
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
		cPool, ok := cpools[classID]
		if !ok {
			cPool = &CPool{Pool: make(map[int64]ParseResolvable)}
			cpools[classID] = cPool
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
			cPool.Pool[idx] = v
		}
	}
	return nil
}
