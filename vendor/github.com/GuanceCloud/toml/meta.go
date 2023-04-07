package toml

import (
	"strconv"
	"strings"
)

// MetaData allows access to meta information about TOML data that's not
// accessible otherwise.
//
// It allows checking if a key is defined in the TOML data, whether any keys
// were undecoded, and the TOML type of a key.
type MetaData struct {
	context Key // Used only during decoding.

	keyInfo  map[string]keyInfo
	mapping  map[string]interface{}
	keys     []Key
	comments KeyDict
	decoded  map[string]struct{}
	data     []byte // Input file; for errors.
}

// IsDefined reports if the key exists in the TOML data.
//
// The key should be specified hierarchically, for example to access the TOML
// key "a.b.c" you would use IsDefined("a", "b", "c"). Keys are case sensitive.
//
// Returns false for an empty key.
func (md *MetaData) IsDefined(key ...string) bool {
	if len(key) == 0 {
		return false
	}

	var (
		hash      map[string]interface{}
		ok        bool
		hashOrVal interface{} = md.mapping
	)
	for _, k := range key {
		if hash, ok = hashOrVal.(map[string]interface{}); !ok {
			return false
		}
		if hashOrVal, ok = hash[k]; !ok {
			return false
		}
	}
	return true
}

// Type returns a string representation of the type of the key specified.
//
// Type will return the empty string if given an empty key or a key that does
// not exist. Keys are case sensitive.
func (md *MetaData) Type(key ...string) string {
	if ki, ok := md.keyInfo[Key(key).String()]; ok {
		return ki.tomlType.typeString()
	}
	return ""
}

// Keys returns a slice of every key in the TOML data, including key groups.
//
// Each key is itself a slice, where the first element is the top of the
// hierarchy and the last is the most specific. The list will have the same
// order as the keys appeared in the TOML data.
//
// All keys returned are non-empty.
func (md *MetaData) Keys() []Key {
	return md.keys
}

// Undecoded returns all keys that have not been decoded in the order in which
// they appear in the original TOML document.
//
// This includes keys that haven't been decoded because of a [Primitive] value.
// Once the Primitive value is decoded, the keys will be considered decoded.
//
// Also note that decoding into an empty interface will result in no decoding,
// and so no keys will be considered decoded.
//
// In this sense, the Undecoded keys correspond to keys in the TOML document
// that do not have a concrete type in your representation.
func (md *MetaData) Undecoded() []Key {
	undecoded := make([]Key, 0, len(md.keys))
	for _, key := range md.keys {
		if _, ok := md.decoded[key.String()]; !ok {
			undecoded = append(undecoded, key)
		}
	}
	return undecoded
}

// Key represents any TOML key, including key groups. Use [MetaData.Keys] to get
// values of this type.
type Key []string

func (k Key) String() string {
	ss := make([]string, len(k))
	for i := range k {
		ss[i] = k.maybeQuoted(i)
	}
	return strings.Join(ss, ".")
}

func (k Key) maybeQuoted(i int) string {
	if k[i] == "" {
		return `""`
	}
	for _, c := range k[i] {
		if !isBareKeyChar(c) {
			return `"` + dblQuotedReplacer.Replace(k[i]) + `"`
		}
	}
	return k[i]
}

func (k Key) add(piece string) Key {
	newKey := make(Key, len(k)+1)
	copy(newKey, k)
	newKey[len(k)] = piece
	return newKey
}

type EntryType int
type keyPhase int
type segmentType int

const (
	EntryComment EntryType = iota + 1
	EntryKeySegments
)

const (
	atHead keyPhase = iota + 1
	atTail
)

const (
	typeString segmentType = iota + 1
	typeInt
)

type segment struct {
	typ   segmentType
	field string
	idx   int
}

func newSegment(x interface{}) segment {
	switch x := x.(type) {
	case string:
		return newField(x)
	case int:
		return newIndex(x)
	}
	panic("only support string or int type")
}

func newField(field string) segment {
	return segment{
		typ:   typeString,
		field: field,
		idx:   0,
	}
}

func newIndex(idx int) segment {
	return segment{
		typ:   typeInt,
		field: "",
		idx:   idx,
	}
}

type Entry interface {
	entryType() EntryType
}

type KeyDict map[segment]*KeySegments

func (k KeyDict) find(ks *KeySegments) *KeySegments {
	var (
		target *KeySegments
		ok     bool
	)
	childDict := k
	for _, piece := range ks.segments {
		seg := newSegment(piece)
		target, ok = childDict[seg]
		if !ok {
			return nil
		}
		childDict = target.children
	}
	return target
}

type KeySegments struct {
	segments        []interface{}
	phase           keyPhase
	documentComment *Comment
	lineTailComment *Comment
	children        KeyDict
}

func newKeySegments() *KeySegments {
	return &KeySegments{
		phase: atHead,
	}
}

func findParentKey(arr []*KeySegments, ks *KeySegments) *KeySegments {
	if len(ks.segments) <= 1 {
		return nil
	}

	ks = ks.copy()
	ks.segments = ks.segments[:len(ks.segments)-1]

	for _, v := range arr {
		if ks.sameAs(v) {
			return v
		}
	}
	return nil
}

func (k *KeySegments) toKey() Key {
	newKey := make(Key, len(k.segments)/2)
	for _, seg := range k.segments {
		if s, ok := seg.(string); ok {
			newKey = append(newKey, s)
		}
	}
	return newKey
}

func (k *KeySegments) entryType() EntryType {
	return EntryKeySegments
}

func (k *KeySegments) String() string {
	var sb strings.Builder
	for i, key := range k.segments {
		switch key := key.(type) {
		case string:
			if i > 0 {
				sb.WriteString(".")
			}
			sb.WriteString(key)
		case int:
			sb.WriteString("[" + strconv.Itoa(key) + "]")
		}
	}
	return sb.String()
}

func (k *KeySegments) atHead() bool {
	return k.phase == atHead
}

func (k *KeySegments) atTail() bool {
	return k.phase == atTail
}

func (k *KeySegments) sameAs(other *KeySegments) bool {
	if len(k.segments) != len(other.segments) {
		return false
	}
	for i, s := range k.segments {
		if s != other.segments[i] {
			return false
		}
	}
	return true
}

func (k *KeySegments) copy() *KeySegments {
	cp := *k
	cp.segments = append([]interface{}(nil), cp.segments...)
	return &cp
}

func (k *KeySegments) start() *KeySegments {
	cp := k.copy()
	cp.phase = atHead
	return cp
}

func (k *KeySegments) finish() *KeySegments {
	cp := k.copy()
	cp.phase = atTail
	return cp
}

func (k *KeySegments) appendIndex(idx ...int) *KeySegments {
	for _, i := range idx {
		k.segments = append(k.segments, i)
	}
	return k
}

func (k *KeySegments) appendFields(keys ...string) *KeySegments {
	for _, key := range keys {
		k.segments = append(k.segments, key)
	}
	return k
}

func (k *KeySegments) popIndex() int {
	if len(k.segments) == 0 {
		panic("empty segments")
	}
	v := k.segments[len(k.segments)-1]
	if _, ok := v.(int); !ok {
		panic("current segment is not an index")
	}
	k.segments = k.segments[:len(k.segments)-1]
	return v.(int)
}

func (k *KeySegments) popField() string {
	if len(k.segments) == 0 {
		panic("empty segments")
	}
	v := k.segments[len(k.segments)-1]
	if _, ok := v.(string); !ok {
		panic("current segment is not a field")
	}
	k.segments = k.segments[:len(k.segments)-1]
	return v.(string)
}

type Comment struct {
	lines      []string
	IsDocument bool
}

func newComment(text string, isDoc bool) *Comment {
	return &Comment{
		lines:      []string{text},
		IsDocument: isDoc,
	}
}

func (c *Comment) entryType() EntryType {
	return EntryComment
}

func (c *Comment) merge(o *Comment) {
	if c.IsDocument != o.IsDocument {
		panic("only same type comment can be merged")
	}

	for i, line := range o.lines {
		if !strings.HasPrefix(strings.TrimLeftFunc(line, isWhitespace), "#") {
			o.lines[i] = "#" + line
		}
	}
	c.lines = append(c.lines, o.lines...)
}

func (c *Comment) String() string {
	return strings.Join(c.lines, "\n")
}

type keyStack struct {
	buckets []*KeySegments
}

func (k *keyStack) push(ks *KeySegments) {
	k.buckets = append(k.buckets, ks)
}

func (k *keyStack) pop() *KeySegments {
	if len(k.buckets) == 0 {
		panic("empty stack")
	}
	v := k.buckets[len(k.buckets)-1]
	k.buckets = k.buckets[:len(k.buckets)-1]
	return v
}
