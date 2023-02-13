// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package point

import (
	"bytes"
	"fmt"
	"hash/fnv"
	"sort"
	"strings"
)

// Key is the key-name and it's type composite.
type Key struct {
	key []byte // key-name + key-type
	def any
}

// NewKey create Key.
func NewKey(k []byte, t KeyType, defaultVal ...any) *Key {
	var def any
	if len(defaultVal) > 0 {
		def = defaultVal[0]
	}

	return &Key{
		key: append(k, uint8(t)),
		def: def,
	}
}

// NewTagKey create tag key with type []byte.
func NewTagKey(k []byte, defaultVal []byte) *Key {
	return NewKey(k, KeyType_D, defaultVal)
}

// Key get key-name.
func (k *Key) Key() []byte {
	switch len(k.key) {
	case 0, 1:
		return nil
	default:
		return k.key[:len(k.key)-1]
	}
}

// Type get key-type.
func (k *Key) Type() KeyType {
	switch len(k.key) {
	case 0:
		return KeyType_X

	case 1:
		return KeyType(k.key[0])

	default:
		return KeyType(k.key[len(k.key)-1])
	}
}

// Default get key's default value.
func (k *Key) Default() any {
	return k.def
}

// Keys is sorted Keys.
type Keys struct {
	hashCount,
	hash uint64
	arr []*Key
}

func (x *Keys) Len() int { return len(x.arr) }

func (x *Keys) Swap(i, j int) {
	arr := x.arr
	arr[i], arr[j] = arr[j], arr[i]
}

func (x *Keys) Less(i, j int) bool {
	return bytes.Compare(x.arr[i].key, x.arr[j].key) < 0
}

// Has test if k exist.
func (x *Keys) Has(k *Key) bool {
	// TODO: should replaced by sort.Search()
	for _, item := range x.arr {
		if bytes.Equal(k.key, item.key) {
			return true
		}
	}

	return false
}

// Add add specific k. if key & type exist, do nothing.
func (x *Keys) Add(k *Key) {
	if x.Has(k) {
		return
	}

	x.arr = append(x.arr, k)
	sort.Sort(x)
	x.hash = 0 // reset hash
}

// Del remove specific k.
func (x *Keys) Del(k *Key) {
	i := 0
	for _, key := range x.arr {
		if !bytes.Equal(key.key, k.key) {
			x.arr[i] = key
			i++
		}
	}

	// len changed, reset hash.
	if len(x.arr) != i {
		x.hash = 0
		x.arr = x.arr[:i]
		sort.Sort(x)
	}
}

// Pretty get pretty showing of all keys.
func (x *Keys) Pretty() string {
	arr := []string{}
	for _, k := range x.arr {
		arr = append(arr, fmt.Sprintf("% 4s: %q", KeyType(k.key[len(k.key)-1]), k.key[:len(k.key)-1]))
	}

	arr = append(arr, fmt.Sprintf("-----\nhashed: %d", x.hashCount))

	return strings.Join(arr, "\n")
}

// Hash calculate x's hash.
func (x *Keys) Hash() uint64 {
	if x.hash == 0 {
		h := fnv.New64a()
		if _, err := h.Write(func() []byte {
			var arr []byte
			for _, k := range x.arr {
				arr = append(arr, k.key...)
			}
			return arr
		}()); err == nil {
			x.hash = h.Sum64()
			x.hashCount++
		}
	}
	return x.hash
}
