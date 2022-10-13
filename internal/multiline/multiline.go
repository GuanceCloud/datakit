// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package multiline wrap regexp/match functions
package multiline

import (
	"bytes"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

const (
	maxMutilineBytes = 32 * 1024 * 1024
	flushDuration    = time.Second * 3
)

type Multiline struct {
	automult *AutoMultiline
	buff     bytes.Buffer

	// prefixSpace 用以标记 pattern 为空的情况，属于默认行为，
	// 如果一行数据，它的首字符是 WhiteSpace，那它就是多行
	// WhiteSpace 定义为 '\t', '\n', '\v', '\f', '\r', ' ', 0x85, 0xA0
	prefixSpace bool

	lastWriteTime time.Time // 记录最后一次匹配成功并写入到 buff 的时间
}

func New(patterns []string) (*Multiline, error) {
	m := &Multiline{}
	var err error

	if len(patterns) == 0 {
		m.prefixSpace = true
	} else {
		m.automult, err = NewAutoMultiline(patterns)
	}

	return m, err
}

func (m *Multiline) ProcessLineString(text string) string {
	// --匹配成功--
	// 清空 buff 并写入新的文本，符合多行行为
	// 记录当前时间
	// ex buff:
	// "2022-10-10 15:13:11 INFO"
	if m.matchString(text) {
		previousText := m.FlushString()
		m.buff.WriteString(text)
		m.lastWriteTime = time.Now()
		return previousText
	}

	// --匹配失败--

	// 这一条文本匹配失败，原应该追加写入到 buff 中，但是此时 buff 为空，说明这条文本没有头，是一条“僵尸多行文本”
	// 为了避免匹配失败的文本堆积在 buff 中，需要在此处直接 return
	if m.buff.Len() == 0 {
		return text
	}

	// buff 不为空，说明 buff 中存在匹配成功的文本
	// 将本次匹配失败的文本写入到 buff，追加到末尾，符合多行行为
	// ex buff:
	// "2022-10-10 15:13:11 INFO\nTEXT-ABCD"
	m.buff.WriteString("\n")
	m.buff.WriteString(text)

	// flush 规则一：
	// 为了确保不会出现一条匹配成功，N 条匹配失败然后追加写入到 buff，导致数据全部堆积的情况，约定多行数据的采集周期不超过 3 秒
	// 即，如果上次匹配成功的时间，距离现在超过 3 秒，则强行 flush 并清空 buff
	// 注意，这个 3 秒包括文件读取和函数调用，如果上层调用采集的间隔过长，需要酌情调整，推荐时间 10 秒、30 秒
	if time.Since(m.lastWriteTime) > flushDuration {
		previousText := m.FlushString()
		return previousText
	}

	// flush 规则二：
	// buff 长度超过限制，强行 flush
	if m.buff.Len() > maxMutilineBytes {
		return m.FlushString()
	}

	// 没有 flush，文本数据都在 buff 中，只能返回一个空字符串
	return ""
}

func (m *Multiline) ProcessLine(text []byte) []byte {
	tmp := m.ProcessLineString(string(text))
	return []byte(tmp)
}

func (m *Multiline) BuffLength() int {
	return m.buff.Len()
}

func (m *Multiline) Flush() []byte {
	if m.buff.Len() == 0 {
		return nil
	}

	text := make([]byte, m.buff.Len())
	copy(text, m.buff.Bytes())

	m.buff.Reset()
	return text
}

func (m *Multiline) match(text []byte) bool {
	if m.prefixSpace {
		return m.matchOfPrefixSpace(text)
	}
	return m.automult.Match(text)
}

func (m *Multiline) matchString(text string) bool {
	return m.match([]byte(text))
}

func (m *Multiline) matchOfPrefixSpace(text []byte) bool {
	if len(text) == 0 {
		return true
	}
	return !unicode.IsSpace(rune(text[0]))
}

func (m *Multiline) FlushString() string {
	if m.buff.Len() == 0 {
		return ""
	}
	text := m.buff.String()
	m.buff.Reset()
	return text
}

func (m *Multiline) BuffString() string {
	return m.buff.String()
}

var asciiSpace = [256]uint8{'\t': 1, '\n': 1, '\v': 1, '\f': 1, '\r': 1, ' ': 1}

func TrimRightSpace(s string) string {
	end := len(s)
	for ; end > 0; end-- {
		c := s[end-1]
		if c >= utf8.RuneSelf {
			return strings.TrimFunc(s[:end], unicode.IsSpace)
		}
		if asciiSpace[c] == 0 {
			break
		}
	}
	return s[:end]
}
