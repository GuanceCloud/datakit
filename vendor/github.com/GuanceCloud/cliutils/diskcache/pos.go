// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package diskcache

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
)

type pos struct {
	Seek int64  `json:"seek"`
	Name []byte `json:"name"`

	fname string        // where to dump the binary data
	buf   *bytes.Buffer // reused buffer to build the binary data
}

func (p *pos) String() string {
	return fmt.Sprintf("%s:%d", string(p.Name), p.Seek)
}

func posFromFile(fname string) (*pos, error) {
	bin, err := ioutil.ReadFile(filepath.Clean(fname))
	if err != nil {
		return nil, err
	}

	var p pos
	if err := p.UnmarshalBinary(bin); err != nil {
		return nil, err
	}
	return &p, nil
}

func (p *pos) MarshalBinary() ([]byte, error) {
	if p.buf == nil {
		p.buf = new(bytes.Buffer)
	}

	p.buf.Reset()

	if err := binary.Write(p.buf, binary.LittleEndian, p.Seek); err != nil {
		return nil, err
	}

	if _, err := p.buf.Write(p.Name); err != nil {
		return nil, err
	}

	return p.buf.Bytes(), nil
}

func (p *pos) UnmarshalBinary(bin []byte) error {
	p.buf = bytes.NewBuffer(bin)

	if err := binary.Read(p.buf, binary.LittleEndian, &p.Seek); err != nil {
		return err
	}

	p.Name = p.buf.Bytes()
	return nil
}

func (p *pos) reset() error {
	p.Seek = -1
	p.Name = nil
	if p.buf != nil {
		p.buf.Reset()
	}

	return p.dumpFile()
}

func (p *pos) dumpFile() error {
	if data, err := p.MarshalBinary(); err != nil {
		return err
	} else {
		return ioutil.WriteFile(p.fname, data, 0o600)
	}
}

// for benchmark.
func (p *pos) dumpJSON() ([]byte, error) {
	j, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}
	return j, nil
}
