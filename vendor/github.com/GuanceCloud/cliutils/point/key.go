// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package point

import (
	"fmt"
	"hash/fnv"
	"sort"
	"strings"
)

// Key is the key-name and it's type composite.
type Key struct {
	key string // key-name + key-type
	t   KeyType
	def any
}

// NewKey create Key.
func NewKey(k string, t KeyType, defaultVal ...any) *Key {
	var def any
	if len(defaultVal) > 0 {
		def = defaultVal[0]
	}

	return &Key{
		key: k,
		t:   t,
		def: def,
	}
}

// NewTagKey create tag key with type string.
func NewTagKey(k string, defaultVal string) *Key {
	return NewKey(k, S, defaultVal)
}

// Key get key-name.
func (k *Key) Key() string {
	return k.key
}

// Type get key-type.
func (k *Key) Type() KeyType {
	switch len(k.key) {
	case 0:
		return X

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
	hashed bool
	hash   uint64
	arr    []*Key
}

func (x *Keys) Len() int { return len(x.arr) }

func (x *Keys) Swap(i, j int) {
	arr := x.arr
	arr[i], arr[j] = arr[j], arr[i]
}

func (x *Keys) Less(i, j int) bool {
	return strings.Compare(x.arr[i].key, x.arr[j].key) < 0
}

// Has test if k exist.
func (x *Keys) Has(k *Key) bool {
	// TODO: should replaced by sort.Search()
	for _, item := range x.arr {
		if k.key == item.key {
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
	x.hashed = false
}

// Del remove specific k.
func (x *Keys) Del(k *Key) {
	i := 0
	for _, key := range x.arr {
		if key.key != k.key {
			x.arr[i] = key
			i++
		}
	}

	if i != len(x.arr) {
		x.hashed = false
		x.arr = x.arr[:i]
	}
}

// Pretty get pretty showing of all keys.
func (x *Keys) Pretty() string {
	arr := []string{}
	for _, k := range x.arr {
		arr = append(arr, fmt.Sprintf("% 4s: %q", KeyType(k.key[len(k.key)-1]), k.key[:len(k.key)-1]))
	}

	arr = append(arr, fmt.Sprintf("-----\nhashed: %v", x.hashed))

	return strings.Join(arr, "\n")
}

// Hash calculate x's hash.
func (x *Keys) Hash() uint64 {
	if !x.hashed {
		sort.Sort(x)
		h := fnv.New64a()
		if _, err := h.Write(func() []byte {
			var arr []byte
			for _, k := range x.arr {
				arr = append(arr, k.key...)
			}
			return arr
		}()); err == nil {
			x.hash = h.Sum64()
			x.hashed = true
		}
	}

	return x.hash
}
