// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package dataway

import (
	"bytes"
	"fmt"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
)

type bodyPayload int

const (
	payloadLineProtocol bodyPayload = iota
	// TODO: used to cache protobuf point.
)

type body struct {
	buf     []byte
	rawLen  int
	gzon    bool
	npts    int
	payload bodyPayload
}

func (b *body) String() string {
	return fmt.Sprintf("gzon: %v, pts: %d, buf bytes: %d", b.gzon, b.npts, len(b.buf))
}

// getBody buidl a body instance.
func getBody(lines [][]byte, idxBegin, idxEnd, curPartSize int) (*body, error) {
	out := &body{
		buf:     bytes.Join(lines, seprator),
		payload: payloadLineProtocol,
		npts:    idxEnd - idxBegin,
	}

	out.rawLen = len(out.buf)
	gzbuf, err := datakit.GZip(out.buf)
	if err != nil {
		log.Errorf("gz: %s", err.Error())
		return nil, err
	} else {
		out.buf = gzbuf
		out.gzon = true
	}

	return out, nil
}

// buildBody convert pts to lineprotocol body.
func buildBody(pts []*point.Point, max int) ([]*body, error) {
	lines := [][]byte{}
	curPartSize := 0

	var bodies []*body

	idxBegin := 0
	for idx, pt := range pts {
		ptbytes := []byte(pt.String())

		// 此处必须提前预判包是否会大于上限值，当新进来的 ptbytes 可能
		// 会超过上限时，就应该及时将已有数据（肯定没超限）打包一下。
		if curPartSize+len(lines)+len(ptbytes) >= max {
			log.Debugf("merge %d points as body", len(lines))

			if body, err := getBody(lines, idxBegin, idx, curPartSize); err != nil {
				return nil, err
			} else {
				idxBegin = idx
				bodies = append(bodies, body)
				lines = lines[:0]
				curPartSize = 0
			}
		}

		// 如果上面有打包，这里将是一个新的包，否则 ptbytes 还是追加到
		// 已有数据上。
		lines = append(lines, ptbytes)
		curPartSize += len(ptbytes)
	}

	if len(lines) > 0 { // 尾部 lines 单独打包一下
		if body, err := getBody(lines, idxBegin, len(pts), curPartSize); err != nil {
			return nil, err
		} else {
			return append(bodies, body), nil
		}
	} else {
		return bodies, nil
	}
}
