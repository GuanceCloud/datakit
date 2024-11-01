// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package point

import (
	"fmt"

	"github.com/GuanceCloud/cliutils/logger"
)

var (
	// logger for debugging, do NOT add logging in release code.
	l = logger.DefaultSLogger("point")
)

// EncodeV2 set points to be encoded.
func (e *Encoder) EncodeV2(pts []*Point) {
	e.pts = pts
}

// Next is a iterator that return encoded data in prepared buf.
// If all points encoded done, return false.
// If any error occurred, we can get the error by call e.LastErr().
func (e *Encoder) Next(buf []byte) ([]byte, bool) {
	switch e.enc {
	case Protobuf:
		return e.doEncodeProtobuf(buf)
	case LineProtocol:
		return e.doEncodeLineProtocol(buf)
	case JSON:
		return e.doEncodeJSON(buf)
	case PBJSON:
		return e.doEncodePBJSON(buf)
	default: // TODO: json
		return nil, false
	}
}

func (e *Encoder) doEncodeProtobuf(buf []byte) ([]byte, bool) {
	var (
		need,
		curSize int
	)

	if e.lastErr != nil {
		return nil, false
	}

	defer func() {
		// clear encoding array
		e.pbpts.Arr = e.pbpts.Arr[:0]
	}()

	for _, pt := range e.pts[e.lastPtsIdx:] {
		if pt == nil {
			e.lastPtsIdx++
			continue
		}

		var ptsize int
		if !e.looseSize {
			ptsize = pt.PBSize()
		} else {
			ptsize = pt.Size()
		}

		// added points with the new-added-point will larger than buf, we should shift added point to encoding.
		if curSize+ptsize > len(buf) {
			switch len(e.pbpts.Arr) {
			case 0: // nothing added, current point(not added so far) too large
				if e.ignoreLargePoint {
					//l.Debugf("skip %d/%d point", idx, e.lastPtsIdx)
					e.lastPtsIdx++
					e.skippedPts++
					continue
				} else {
					e.lastErr = fmt.Errorf("%w: need at least %d bytes, only %d available, nothing has added, this is a huge point",
						errTooSmallBuffer, curSize+ptsize, len(buf))
					return nil, false
				}

			default:
				trimmed := 1
				for { // exists point(added) may still exceed buf, try trim tail points until size fit to @buf.
					if pbptsSize := e.pbpts.Size(); pbptsSize > len(buf) {
						if len(e.pbpts.Arr) == 1 {
							if e.ignoreLargePoint {
								// NOTE: we have added the lastPtsIdx when append the last
								// point to pbpts.Arr, so do not add here

								e.skippedPts++
								e.totalPts--                  // remove the single huge point, we haven't encode it indeed
								e.pbpts.Arr = e.pbpts.Arr[:0] // clear the last one: it's too large for buf
								break
							} else {
								e.lastErr = fmt.Errorf("%w: need at least %d bytes, only %d available, current points: %d",
									errTooSmallBuffer, curSize, len(buf), len(e.pbpts.Arr))
								return nil, false
							}
						}

						e.pbpts.Arr = e.pbpts.Arr[:len(e.pbpts.Arr)-trimmed]
						e.trimmedPts += trimmed
						e.lastPtsIdx -= trimmed
						e.totalPts -= trimmed
						trimmed *= 2 // try trim more tail points to save e.pbpts.Size() cost
					} else {
						need = pbptsSize
						goto __doEncodeDirectly // we do not need to sum size of e.pbpts
					}
				}
			}
		} else {
			curSize += ptsize
			e.pbpts.Arr = append(e.pbpts.Arr, pt.pt)
			e.totalPts++
			e.lastPtsIdx++
		}
	}

	need = e.pbpts.Size()

	if need > len(buf) {
		if e.ignoreLargePoint {
			e.trimmedPts++
			e.skippedPts++
			e.totalPts--
		} else {
			e.lastErr = fmt.Errorf("%w: need at least %d bytes, only %d available, current points: %d",
				errTooSmallBuffer, curSize, len(buf), len(e.pbpts.Arr))
		}

		// no more point in e.pts
		return nil, false
	}

__doEncodeDirectly:
	if len(e.pbpts.Arr) == 0 {
		return nil, false
	}

	if n, err := e.pbpts.MarshalToSizedBuffer(buf[:need]); err != nil {
		e.lastErr = err
		return nil, false
	} else {
		if e.fn != nil {
			if err := e.fn(len(e.pbpts.Arr), buf[:n]); err != nil {
				e.lastErr = err
				return nil, false
			}
		}

		e.parts++
		e.totalBytes += n
		return buf[:n], true
	}
}

func (e *Encoder) doEncodeLineProtocol(buf []byte) ([]byte, bool) {
	curSize := 0
	npts := 0

	for _, pt := range e.pts[e.lastPtsIdx:] {
		if pt == nil {
			continue
		}

		lppt, err := pt.LPPoint()
		if err != nil {
			e.lastErr = err
			continue
		}

		ptsize := lppt.StringSize()

		if curSize+ptsize+1 > len(buf) { // extra +1 used to store the last '\n'
			if curSize == 0 { // nothing added
				e.lastErr = errTooSmallBuffer
				return nil, false
			}

			if e.fn != nil {
				if err := e.fn(npts, buf[:curSize]); err != nil {
					e.lastErr = err
					return nil, false
				}
			}
			e.parts++
			return buf[:curSize], true
		} else {
			e.lpPointBuf = lppt.AppendString(e.lpPointBuf)

			copy(buf[curSize:], e.lpPointBuf[:ptsize])

			// Always add '\n' to the end of current point, this may
			// cause a _unneeded_ '\n' to the end of buf, it's ok for
			// line-protocol parsing.
			buf[curSize+ptsize] = '\n'
			curSize += (ptsize + 1)

			// clean buffer, next time AppendString() append from byte 0
			e.lpPointBuf = e.lpPointBuf[:0]
			e.lastPtsIdx++
			npts++
		}
	}

	if curSize > 0 {
		e.parts++
		if e.fn != nil { // NOTE: encode callback error will terminate encode
			if err := e.fn(npts, buf[:curSize]); err != nil {
				e.lastErr = err
				return nil, false
			}
		}
		return buf[:curSize], true
	} else {
		return nil, false
	}
}

func (e *Encoder) doEncodeJSON(buf []byte) ([]byte, bool) {
	return nil, false
}

func (e *Encoder) doEncodePBJSON(buf []byte) ([]byte, bool) {
	return nil, false
}
