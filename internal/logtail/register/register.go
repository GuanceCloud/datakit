// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package register wraps history cache functions
package register

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/GuanceCloud/cliutils/logger"
)

const defaultFlushFactor = 32

type MetaData struct {
	Source string `json:"source"`
	Offset int64  `json:"offset"`
}

func (m *MetaData) String() string {
	return fmt.Sprintf("source: %s, offset: %d", m.Source, m.Offset)
}

func (m *MetaData) DeepCopy() MetaData {
	return MetaData{
		Source: m.Source,
		Offset: m.Offset,
	}
}

type Register interface {
	Set(string, *MetaData) error
	Get(string) *MetaData
	Clean()
	Flush() error
}

type register struct {
	Data map[string]*MetaData `json:"history"`

	file        string
	count       int // set count
	flushFactor int

	mu sync.Mutex
}

func MustNewRegisterFile(file string) (Register, error) {
	return newRegister(file, true)
}

var l = logger.DefaultSLogger("register")

func newRegister(file string, force bool) (*register, error) {
	l = logger.SLogger("register")

	var r *register
	var err error

	func() {
		b, readErr := os.ReadFile(filepath.Clean(file))
		if readErr != nil {
			err = readErr
			return
		}
		r, err = parse(b)
	}()

	if err != nil {
		l.Warnf("init register error: %s", err)
		if force {
			l.Info("create new register file")
			r = &register{}
		} else {
			return nil, err
		}
	}

	if r.Data == nil {
		r.Data = make(map[string]*MetaData)
	}
	r.file = filepath.Clean(file)
	r.flushFactor = defaultFlushFactor
	r.mu = sync.Mutex{}

	return r, nil
}

func (r *register) Flush() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.flush()
}

func (r *register) flush() error {
	if r.count != 0 {
		b, err := json.Marshal(r)
		if err != nil {
			return fmt.Errorf("unable build matedata, err %w", err)
		}

		if err := os.WriteFile(r.file, b, 0o600); err != nil {
			return fmt.Errorf("failed to write file, err %w", err)
		}

		r.count = 0
	}
	return nil
}

func (r *register) Set(key string, value *MetaData) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.Data[key] = value
	r.count++
	if r.count%r.flushFactor == 0 {
		return r.flush()
	}

	return nil
}

func (r *register) Get(key string) *MetaData {
	r.mu.Lock()
	defer r.mu.Unlock()

	v := r.Data[key]
	return v
}

func (r *register) Clean() {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(r.Data) == 0 {
		l.Info("position is empty")
		return
	}

	l.Infof("already existing position len(%d)", len(r.Data))
	tmp := make(map[string]*MetaData)

	for key, data := range r.Data {
		filename := getFilename(key)
		if filename == "" {
			continue
		}
		_, err := os.Stat(filename)
		if err == nil {
			newdata := data.DeepCopy()
			tmp[key] = &newdata
		}
	}

	r.Data = tmp
	l.Infof("now existing posistion len(%d)", len(r.Data))
}

func parse(b []byte) (*register, error) {
	r := register{}
	if len(b) != 0 {
		if err := json.Unmarshal(b, &r); err != nil {
			return nil, err
		}
	}
	return &r, nil
}

func getFilename(key string) string {
	if res := strings.Split(key, "::"); len(res) > 0 {
		return res[0]
	}
	return ""
}
