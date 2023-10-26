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
	"sync"
)

const defaultFlushFactor = 32

type MetaData struct {
	Source string `json:"source"`
	Offset int64  `json:"offset"`
}

func (m *MetaData) String() string {
	return fmt.Sprintf("source: %s, offset: %d", m.Source, m.Offset)
}

type Register interface {
	Set(string, *MetaData) error
	Get(string) *MetaData
	Flush() error
}

type register struct {
	Data map[string]*MetaData `json:"history"`

	file        *os.File
	count       int // set count
	flushFactor int

	mu sync.Mutex
}

func NewRegisterFileIfNotExist(file string) (Register, error) {
	_, err := os.Stat(file)
	if err != nil {
		if os.IsNotExist(err) {
			f, err := os.Create(filepath.Clean(file))
			if err != nil {
				return nil, err
			}
			file = f.Name()
			_ = f.Close()
		} else {
			return nil, err
		}
	}

	return New(file)
}

func New(file string) (Register, error) {
	b, err := os.ReadFile(filepath.Clean(file))
	if err != nil {
		return nil, err
	}

	r, err := parse(b)
	if err != nil {
		return nil, err
	}

	if r.Data == nil {
		r.Data = make(map[string]*MetaData)
	}
	r.flushFactor = defaultFlushFactor
	r.mu = sync.Mutex{}

	r.file, err = os.OpenFile(filepath.Clean(file), os.O_RDWR, 0o600)
	if err != nil {
		return nil, err
	}

	return r, nil
}

func (r *register) Flush() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.flush()
}

func (r *register) flush() error {
	b, err := json.MarshalIndent(r, "", "    ")
	if err != nil {
		return fmt.Errorf("unable build matedata, err %w", err)
	}

	if err := r.file.Truncate(0); err != nil {
		return fmt.Errorf("truncate logtail.history error %w", err)
	}

	// reset seek to begnning
	_, err = r.file.Seek(0, io.SeekStart)
	if err != nil {
		return fmt.Errorf("failed to reset file to begnning, err %w", err)
	}
	_, err = r.file.Write(b)
	if err != nil {
		return fmt.Errorf("failed to write file, err %w", err)
	}

	r.count = 0
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

func parse(b []byte) (*register, error) {
	r := register{}
	if len(b) != 0 {
		if err := json.Unmarshal(b, &r); err != nil {
			return nil, err
		}
	}
	return &r, nil
}
