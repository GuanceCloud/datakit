// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package v2

import (
	"encoding/binary"
	"net"
	"time"

	protocol "github.com/elastic/go-lumber/protocol/v2"
)

type writer struct {
	c  net.Conn
	to time.Duration
}

func newWriter(c net.Conn, to time.Duration) *writer {
	return &writer{c: c, to: to}
}

func (w *writer) ACK(n int) error {
	var buf [6]byte
	buf[0] = protocol.CodeVersion
	buf[1] = protocol.CodeACK
	binary.BigEndian.PutUint32(buf[2:], uint32(n))

	if err := w.c.SetWriteDeadline(time.Now().Add(w.to)); err != nil {
		return err
	}

	tmp := buf[:]
	for len(tmp) > 0 {
		n, err := w.c.Write(tmp)
		if err != nil {
			return err
		}
		tmp = tmp[n:]
	}
	return nil
}

func (w *writer) Keepalive(n int) error {
	return w.ACK(n)
}
