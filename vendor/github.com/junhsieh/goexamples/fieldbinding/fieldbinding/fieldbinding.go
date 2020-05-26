package fieldbinding

import (
	"sync"
)

// NewFieldBinding ...
func NewFieldBinding() *FieldBinding {
	return &FieldBinding{}
}

// FieldBinding is deisgned for SQL rows.Scan() query.
type FieldBinding struct {
	sync.RWMutex // embedded.  see http://golang.org/ref/spec#Struct_types
	FieldArr     []interface{}
	FieldPtrArr  []interface{}
	FieldCount   int64
	MapFieldToID map[string]int64
}

func (fb *FieldBinding) put(k string, v int64) {
	fb.Lock()
	defer fb.Unlock()
	fb.MapFieldToID[k] = v
}

// Get ...
func (fb *FieldBinding) Get(k string) interface{} {
	fb.RLock()
	defer fb.RUnlock()
	// TODO: check map key exist and fb.FieldArr boundary.
	return fb.FieldArr[fb.MapFieldToID[k]]
}

// PutFields ...
func (fb *FieldBinding) PutFields(fArr []string) {
	fCount := len(fArr)
	fb.FieldArr = make([]interface{}, fCount)
	fb.FieldPtrArr = make([]interface{}, fCount)
	fb.MapFieldToID = make(map[string]int64, fCount)

	for k, v := range fArr {
		fb.FieldPtrArr[k] = &fb.FieldArr[k]
		fb.put(v, int64(k))
	}
}

// GetFieldPtrArr ...
func (fb *FieldBinding) GetFieldPtrArr() []interface{} {
	return fb.FieldPtrArr
}

// GetFieldArr ...
func (fb *FieldBinding) GetFieldArr() map[string]interface{} {
	m := make(map[string]interface{}, fb.FieldCount)

	for k, v := range fb.MapFieldToID {
		m[k] = fb.FieldArr[v]
	}

	return m
}
