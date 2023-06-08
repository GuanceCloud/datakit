// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package register

import (
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRegisterFileIfNotExist(t *testing.T) {
	t.Run("create file", func(t *testing.T) {
		filename := "nonexisting"
		defer os.Remove(filename)

		err := Init(filename)
		assert.NoError(t, err)
		assert.NotNil(t, globalRegister)
	})
}

func TestNew(t *testing.T) {
	// reset globalRegister
	globalRegister = nil

	t.Run("nonexisting", func(t *testing.T) {
		_, err := New("nonexisting")
		assert.Error(t, err)
	})

	t.Run("parse err", func(t *testing.T) {
		file, err := ioutil.TempFile("", "")
		assert.NoError(t, err)
		defer os.Remove(file.Name())

		_, err = file.WriteString("NO JSON")
		assert.NoError(t, err)

		_, err = New(file.Name())
		assert.Error(t, err)
	})

	t.Run("ok", func(t *testing.T) {
		file, err := ioutil.TempFile("", "")
		assert.NoError(t, err)
		defer os.Remove(file.Name())

		content := `{"history":{}}`
		_, err = file.WriteString(content)
		assert.NoError(t, err)

		_, err = New(file.Name())
		assert.NoError(t, err)
	})
}

func TestSetAndGet(t *testing.T) {
	// reset globalRegister
	globalRegister = nil

	t.Run("set err", func(t *testing.T) {
		inKey := "key"
		inValue := &MetaData{Source: "source", Offset: 100}

		err := Set(inKey, inValue)
		assert.Error(t, err)
	})

	t.Run("get err", func(t *testing.T) {
		in := "key"
		value := Get(in)
		assert.Nil(t, value)
	})

	globalRegister = &register{
		Data:        map[string]*MetaData{},
		flushFactor: 2,
	}

	t.Run("global set", func(t *testing.T) {
		inKey := "key"
		inValue := &MetaData{Source: "source", Offset: 100}

		err := Set(inKey, inValue)
		assert.NoError(t, err)
	})

	t.Run("global get", func(t *testing.T) {
		in := "key"
		out := &MetaData{Source: "source", Offset: 100}

		value := Get(in)
		assert.Equal(t, value, out)
	})
}

func TestFlush(t *testing.T) {
	// write file
	file, err := ioutil.TempFile("", "")
	assert.NoError(t, err)
	defer os.Remove(file.Name())

	content := `{"history":{}}`
	_, err = file.WriteString(content)
	assert.NoError(t, err)

	// new register
	r, err := New(file.Name())
	assert.NoError(t, err)

	err = r.Set("key", &MetaData{Source: "source", Offset: 100})
	assert.NoError(t, err)

	// flush
	err = r.(*register).Flush()
	assert.NoError(t, err)

	// verification
	data, err := ioutil.ReadFile(file.Name())
	assert.NoError(t, err)

	out := strings.ReplaceAll(string(data), " ", "")
	out = strings.ReplaceAll(out, "\n", "")
	tc := `{"history":{"key":{"source":"source","offset":100}}}`

	assert.Equal(t, tc, out)
}

func TestParse(t *testing.T) {
	cases := []struct {
		in   string
		out  *register
		fail bool
	}{
		{
			in: `{ 
				"history": { 
					"key1": {
						"source": "source01",
						"offset": 100
				        },
					"key2": {
						"source": "source02",
						"offset": 200
				        }
				} 
			}`,
			out: &register{
				Data: map[string]*MetaData{
					"key1": {
						Source: "source01",
						Offset: 100,
					},
					"key2": {
						Source: "source02",
						Offset: 200,
					},
				},
			},
		},
		{
			in:  ``,
			out: &register{},
		},
		{
			in:   `NO JSON`,
			out:  nil,
			fail: true,
		},
	}

	for idx, tc := range cases {
		t.Run(strconv.Itoa(idx), func(t *testing.T) {
			res, err := parse([]byte(tc.in))
			if !tc.fail {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
			assert.Equal(t, tc.out, res)
		})
	}
}
