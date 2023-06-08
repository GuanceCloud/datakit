// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package cat struct
package cat

import (
	"bytes"
	"fmt"
)

const (
	SUCCESS = "ok"
	FAIL    = "error"
)

type Message struct {
	Type   string
	Name   string
	Status string

	timestampInNano int64

	data *bytes.Buffer
}

func (m *Message) SetStatus(status string) {
	m.Status = status
}

func (m *Message) toString() string {
	return fmt.Sprintf("Message Type:%s, Name: %s, Status: %s, timeNamo:%d data lens :%d",
		m.Type, m.Name, m.Status, m.timestampInNano/1e6, m.data.Len())
}

type Event struct {
	Message
}

func (e *Event) ToString() string {
	return e.Message.toString()
}

type metric struct {
	Message
}

func (m *metric) ToString() string {
	return m.Message.toString()
}

type Heartbeat struct {
	Message
}

func (h Heartbeat) ToString() string {
	return fmt.Sprintf("Heartbeat Type:%s, Name: %s, Status: %s,data.len=%d, timeNamo:%d ",
		h.Type, h.Name, h.Status, h.data.Len(), h.timestampInNano/1e6)
}

type Transaction struct {
	Message

	durationInNano      int64
	durationStartInNano int64
	children            []*Message
}

func (t *Transaction) addChild(message *Message) {
	if t.children == nil {
		t.children = make([]*Message, 0)
	}
	t.children = append(t.children, message)
}

func (t *Transaction) setData(bts []byte) {
	if t.data == nil {
		t.data = bytes.NewBuffer(bts)
	} else {
		_, _ = t.data.Write(bts)
	}
}

func (t *Transaction) ToString() string {
	return t.toString()
}

type Context struct {
	mTree    *MessageTree
	mParents []*Transaction
}

func (mt *MessageTree) ToString() string {
	str := "domain: " + mt.domain + "\n" +
		"host name: " + mt.hostName + "\n" +
		"address: " + mt.addr + "\n" +
		"thread group name: " + mt.threadGroupName + "\n" +
		"thread ID: " + mt.ThreadID + "\n" +
		"thread name: " + mt.threadName + "\n" +
		"message ID: " + mt.MessageID + "\n" +
		"parent message ID: " + mt.parentMessageID + "\n" +
		"root message ID: " + mt.RootMessageID + "\n" +
		"session token: " + mt.SessionToken + "\n"

	if mt.message != nil {
		str += "message: " + mt.message.toString() + "\n"
	}

	str += "events: [\n"
	for _, event := range mt.events {
		str += "\t" + event.ToString() + "\n"
	}
	str += "]\n"

	str += "transactions: [\n"
	for _, transaction := range mt.transactions {
		str += "\t" + transaction.ToString() + "\n"
	}
	str += "]\n"

	str += "heartbeats: [\n"
	for _, heartbeat := range mt.heartbeats {
		str += "\t" + heartbeat.ToString() + "\n"
	}
	str += "]\n"

	str += "metrics: [\n"
	for _, metric := range mt.metrics {
		str += "\t" + metric.ToString() + "\n"
	}
	str += "]\n"

	return str
}

type MessageTree struct {
	domain          string
	hostName        string
	addr            string
	threadGroupName string
	ThreadID        string
	threadName      string
	MessageID       string
	parentMessageID string
	RootMessageID   string
	SessionToken    string
	message         *Message
	events          []*Event
	transactions    []*Transaction
	heartbeats      []*Heartbeat
	metrics         []*metric
}

func (mt *MessageTree) addEvent(e *Event) {
	if mt.events == nil {
		mt.events = make([]*Event, 0)
	}
	mt.events = append(mt.events, e)
}

func (mt *MessageTree) addMetric(m *metric) {
	if mt.metrics == nil {
		mt.metrics = make([]*metric, 0)
	}
	mt.metrics = append(mt.metrics, m)
}

func (c *Context) addChild(msg *Message) {
	if len(c.mParents) != 0 {
		c.mParents[0].addChild(msg)
	} else {
		c.mTree.message = msg
	}
}

func (c *Context) popTransaction() *Transaction {
	t := c.mParents[len(c.mParents)-1]
	c.mParents = c.mParents[:len(c.mParents)-1]
	return t
}

func (c *Context) pushTransaction(t *Transaction) {
	if len(c.mParents) > 0 {
		c.mParents[len(c.mParents)-1].addChild(&t.Message)
	}
	c.mParents = append(c.mParents, t)
}
