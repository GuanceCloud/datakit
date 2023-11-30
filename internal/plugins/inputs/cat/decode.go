// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package cat decode
package cat

import (
	"bytes"
	"encoding/binary"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"

	"github.com/GuanceCloud/cliutils/point"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	itrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/trace"
)

const (
	maxRead = 1024 * 200 // 200KB
)

// CATKVS is cat client http config.
type CATKVS struct {
	Kvs *Kvs `json:"kvs"`
}

type Kvs struct {
	StartTransactionTypes string `json:"startTransactionTypes"`
	Block                 string `json:"block"`
	Routers               string `json:"routers"`
	Sample                string `json:"sample"`
	MatchTransactionTypes string `json:"matchTransactionTypes"`
}

func defaultKVS() *CATKVS {
	return &CATKVS{Kvs: &Kvs{
		StartTransactionTypes: "Cache.;Squirrel.",
		Block:                 "false",
		Routers:               "127.0.0.1",
		Sample:                "1.0",
		MatchTransactionTypes: "SQL",
	}}
}

func (ipt *Input) dotcp(port string) {
	listen, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Errorf("Listen tcp error=%v", err)
		return
	}
	ipt.listener = listen
	for {
		conn, err := listen.Accept()
		if err != nil {
			if strings.Contains(err.Error(), "use of closed network connection") ||
				errors.Is(err, net.ErrClosed) {
				log.Error("conn is close by Input")
				break
			}
			log.Errorf("accept err = %v", err)
			continue
		}
		go func() {
			defer conn.Close() //nolint:errcheck

			// 持续读取请求并处理
			for {
				buf, err := readPacket(conn)
				if err != nil {
					log.Error("failed to read packet: %v", err)
					break
				}

				// 处理数据包
				ipt.handleMsg(buf)
			}
		}()
	}
	log.Infof("tcp Listen exit")
}

// readPacket: see readme.md:78.
func readPacket(r io.Reader) (*bytes.Buffer, error) {
	header := make([]byte, 4)
	if _, err := io.ReadFull(r, header); err != nil {
		return nil, err
	}
	length := binary.BigEndian.Uint32(header)

	if length > uint32(maxRead) {
		return nil, fmt.Errorf("packet too large: %d", length)
	}

	body := make([]byte, length)
	if _, err := io.ReadFull(r, body); err != nil {
		return nil, err
	}

	return bytes.NewBuffer(body), nil
}

func (ipt *Input) handleMsg(buf *bytes.Buffer) {
	version := readVersion(buf) // do not delete this line !!!
	log.Debugf("version=%s ", version)
	msgTree := &MessageTree{
		domain:          readString(buf),
		hostName:        readString(buf),
		addr:            readString(buf),
		threadGroupName: readString(buf),
		ThreadID:        readString(buf),
		threadName:      readString(buf),
		MessageID:       readString(buf),
		parentMessageID: readString(buf),
		RootMessageID:   readString(buf),
		SessionToken:    readString(buf),
	}

	ctx := &Context{mTree: msgTree}
	// data t
	dt := ctx.decodeMessage(buf)
	if dt == nil {
		log.Debugf("dt is nil  return")
		return
	}

	if dt.Name == "Reboot" {
		log.Debugf("name is reboot  return")
		return
	}
	if dt.Name == "Status" {
		pts := parseMetrics(ctx.mTree.heartbeats, ctx.mTree.domain, ctx.mTree.hostName)
		if len(pts) > 0 {
			err := ipt.feeder.Feed("cat", point.Metric, pts, &dkio.Option{})
			if err != nil {
				log.Error("io feed err=%v", err)
			}
			log.Infof("feed %d metric pts", len(pts))
		}

		return
	}
	dktraces := parseResourceSpans(ctx, dt)
	if afterGatherRun != nil {
		afterGatherRun.Run(inputName, dktraces)
	}
}

func parseMetrics(heartbeats []*Heartbeat, domain, hostname string) []*point.Point {
	if len(heartbeats) == 0 {
		return nil
	}
	pts := make([]*point.Point, 0)
	for _, heartbeat := range heartbeats {
		status := &Status{}
		err := xml.Unmarshal(heartbeat.data.Bytes(), status)
		if err != nil {
			log.Warnf("err=%v", err)
			continue
		}
		spts := status.toPoint(domain, hostname)
		if len(spts) != 0 {
			pts = append(pts, spts...)
		}
	}

	return pts
}

var traceOpts = []point.Option{}

func parseResourceSpans(ctx *Context, dt *Transaction) itrace.DatakitTraces {
	var dktraces itrace.DatakitTraces
	var dktrace itrace.DatakitTrace

	tree := ctx.mTree
	root := tree.MessageID
	if tree.RootMessageID != "" {
		root = tree.RootMessageID
	}
	parent := tree.parentMessageID
	spanType := itrace.SpanTypeLocal
	if parent == "" {
		spanType = itrace.SpanTypeEntry
		parent = "0"
	}
	status := SUCCESS
	if dt.Status != "0" {
		status = FAIL
	}

	spanKV := point.KVs{}
	spanKV = spanKV.Add(itrace.FieldTraceID, root, false, false).
		Add(itrace.FieldParentID, parent, false, false).
		Add(itrace.FieldSpanid, tree.MessageID, false, false).
		AddTag(itrace.TagService, tree.domain).
		Add(itrace.FieldResource, dt.Name, false, false).
		AddTag(itrace.TagOperation, dt.Name).
		AddTag(itrace.TagSource, inputName).
		AddTag(itrace.TagSourceType, dt.Type).
		AddTag(itrace.TagSpanType, spanType).
		Add(itrace.FieldStart, dt.durationStartInNano/1000, false, false).
		Add(itrace.FieldDuration, dt.durationInNano/1000, false, false).
		AddTag(itrace.TagSpanStatus, status).
		AddTag(itrace.TagHost, tree.hostName).
		AddTag("address", tree.addr).
		AddTag("thread_group_name", tree.threadGroupName).
		AddTag("thread_id", tree.ThreadID).
		AddTag("thread_name", tree.threadName)

	for k, v := range globalTags {
		spanKV = spanKV.AddTag(k, v)
	}

	pt := point.NewPointV2(inputName, spanKV, traceOpts...)
	dktrace = append(dktrace, &itrace.DkSpan{Point: pt})

	dktraces = append(dktraces, dktrace)

	return dktraces
}

func (c *Context) decodeMessage(buf *bytes.Buffer) *Transaction {
	var msg *Transaction
	for buf.Len() > 0 {
		b := readInt(buf)
		switch b {
		case 't':
			timestamp := readInt64(buf)
			t := readString(buf)
			name := readString(buf)
			if t == "System" && strings.HasPrefix(name, "UploadMetric") {
				name = "UploadMetric"
			}
			dt := &Transaction{}
			dt.Type = t
			dt.Name = name
			dt.durationStartInNano = timestamp * 1e6
			c.pushTransaction(dt)
		case 'T': // 结束
			status := readString(buf)
			data := readString(buf)
			duration := readInt64(buf)

			dt := c.popTransaction()
			dt.SetStatus(status)
			dt.durationInNano = duration * 1e3
			dt.setData([]byte(data))
			c.mTree.transactions = append(c.mTree.transactions, dt)
			msg = dt
		case 'E':
			timestamp := readInt64(buf)
			t := readString(buf)
			name := readString(buf)
			status := readString(buf)
			data := readString(buf)
			e := &Event{
				Message{
					Type:            t,
					Name:            name,
					Status:          status,
					timestampInNano: timestamp,
					data:            bytes.NewBuffer([]byte(data)),
				},
			}

			c.mTree.addEvent(e)
		case 'M':
			timestamp := readInt64(buf)
			t := readString(buf)
			name := readString(buf)
			status := readString(buf)
			data := readString(buf)
			m := &metric{Message{
				Type:            t,
				Name:            name,
				Status:          status,
				timestampInNano: timestamp,
				data:            bytes.NewBuffer([]byte(data)),
			}}

			c.mTree.addMetric(m)
			c.addChild(&m.Message)
		case 'L':
			timestamp := readInt64(buf)
			t := readString(buf)
			name := readString(buf)
			status := readString(buf)
			data := readString(buf)
			h := &Message{
				Type:            t,
				Name:            name,
				Status:          status,
				timestampInNano: timestamp,
				data:            bytes.NewBuffer([]byte(data)),
			}

			c.addChild(h)
		case 'H':
			timestamp := readInt64(buf)
			t := readString(buf)
			name := readString(buf)
			status := readString(buf)
			data := readString(buf)
			h := &Heartbeat{
				Message{
					Type:            t,
					Name:            name,
					Status:          status,
					timestampInNano: timestamp,
					data:            bytes.NewBuffer([]byte(data)),
				},
			}

			if c.mTree.heartbeats == nil {
				c.mTree.heartbeats = make([]*Heartbeat, 0)
			}
			c.mTree.heartbeats = append(c.mTree.heartbeats, h)
		default:
			log.Debugf("default case")
		}
	}
	if msg == nil {
		return nil
	}

	return msg
}

func readString(buf *bytes.Buffer) string {
	mData := make([]byte, 256)
	length := int(readVarInt(buf, 32))
	if length == 0 {
		return ""
	} else if length > len(mData) {
		mData = make([]byte, length)
	}
	_, err := buf.Read(mData[0:length])
	if err != nil {
		return ""
	}

	return string(mData[0:length])
}

func readInt64(buf *bytes.Buffer) int64 {
	return readVarInt(buf, 64)
}

func readVarInt(buf *bytes.Buffer, readLen int) int64 {
	shift, result := 0, int64(0)
	for shift < readLen {
		b, err := buf.ReadByte()
		if err != nil {
			log.Error(err)
		}
		result |= (int64)(b&0x7F) << shift
		if (b & 0x80) == 0 {
			return result
		}
		shift += 7
	}

	return 0
}

func readInt(buf *bytes.Buffer) byte {
	b, err := buf.ReadByte()
	if err != nil {
		log.Error(err)
	}
	return b
}

func readVersion(buf *bytes.Buffer) string {
	bts := make([]byte, 3)
	if _, err := buf.Read(bts); err != nil {
		return ""
	}

	return string(bts)
}
