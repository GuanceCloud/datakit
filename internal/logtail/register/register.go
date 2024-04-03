// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package register wraps history cache functions
package register

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/GuanceCloud/cliutils/logger"
)

var defaultFlushFactor = 512

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

	encoder *json.Encoder
	file    *os.File
	count   int // set count

	mu sync.Mutex
}

func NewRegisterFile(file string) (Register, error) {
	return newRegister(file)
}

var l = logger.DefaultSLogger("register")

func newRegister(file string) (*register, error) {
	l = logger.SLogger("register")

	f, err := os.OpenFile(filepath.Clean(file), os.O_RDWR|os.O_CREATE, 0o600)
	if err != nil {
		l.Warnf("failed of open register file: %s", err)
		return nil, err
	}

	data, err := io.ReadAll(f)
	if err != nil {
		l.Warnf("read register error: %s", err)
		return nil, err
	}

	r, err := parse(data)
	if err != nil {
		l.Infof("parse err %s, create new register", err)
		r = &register{}
	}

	if r.Data == nil {
		r.Data = make(map[string]*MetaData)
	}
	r.file = f
	r.encoder = json.NewEncoder(f)
	r.encoder.SetEscapeHTML(false)
	r.mu = sync.Mutex{}

	return r, nil
}

func (r *register) Flush() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.flush()
}

func (r *register) flush() error {
	if r.count != 0 && r.encoder != nil {
		_, err := r.file.Seek(0, io.SeekStart)
		if err != nil {
			return fmt.Errorf("failed of reset file, err %w", err)
		}
		if err := r.encoder.Encode(r); err != nil {
			return fmt.Errorf("failed of encode, err %w", err)
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
	if r.count%defaultFlushFactor == 0 {
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
