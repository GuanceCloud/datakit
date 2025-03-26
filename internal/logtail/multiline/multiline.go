// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package multiline wraps text multiline match functions
package multiline

import (
	"bytes"
	"unicode"
	"unicode/utf8"
)

const defaultMaxLength = 1 * 1024 * 1024 // 1MB

type option struct {
	// 限制一段多行数据的最大长度，避免出现超级长的多行数据，超出限制会执行 flush
	maxLength int
}

type Option func(*option)

func WithMaxLength(max int) Option {
	return func(opt *option) {
		if max > 0 {
			opt.maxLength = max
		}
	}
}

func defaultOption() *option {
	return &option{maxLength: defaultMaxLength}
}

type Multiline struct {
	*Matcher
	buff bytes.Buffer
	opt  *option

	// 记录上一次 flush 的原因
	lastFlushState State
}

func New(patterns []string, opts ...Option) (*Multiline, error) {
	c := defaultOption()
	for _, opt := range opts {
		opt(c)
	}

	match, err := NewMatcher(patterns)
	if err != nil {
		return nil, err
	}

	return &Multiline{
		Matcher: match,
		opt:     c,
	}, err
}

var newLine = []byte{'\n'}

func (m *Multiline) ProcessLineString(text string) (string, State) {
	if m.MatchString(text) {
		finishedText := m.FlushString()
		m.buff.WriteString(text)
		m.lastFlushState = NewMultiline
		return finishedText, m.lastFlushState
	}

	if m.buff.Len() == 0 && m.lastFlushState != FlushPartial {
		m.lastFlushState = NoContext
		return text, m.lastFlushState
	}

	if m.buff.Len() != 0 {
		m.buff.Write(newLine)
	}
	m.buff.WriteString(text)

	if m.buff.Len() > m.opt.maxLength {
		m.lastFlushState = FlushPartial
		return m.FlushString(), m.lastFlushState
	}

	return "", Written
}

func (m *Multiline) ProcessLine(text []byte) ([]byte, State) {
	// --匹配成功--
	// 清空 buff 并写入新的文本，符合多行模式
	if m.Match(text) {
		finishedText := m.Flush()
		m.buff.Write(text)
		m.lastFlushState = NewMultiline
		return finishedText, NewMultiline
	}

	// --匹配失败--

	// 这一条文本匹配失败，要判断是写入到 buff，还是直接 return
	// 如果此时 buff 的长度为 0，说明刚执行过 flush。flush 的原因有 2 种，即 “正常多行结束” 或 “超出长度被迫 flush”
	// - 正常的多行结束，说明这条数据没有行首，按照 “僵尸数据” 处理，直接 return
	// - 超出长度被迫 flush，说明多行实际没有结束，要继续写入到 buff
	if m.buff.Len() == 0 && m.lastFlushState != FlushPartial {
		m.lastFlushState = NoContext
		return text, m.lastFlushState
	}

	// buff 不为空，说明 buff 中存在匹配成功的文本
	// 将本次匹配失败的文本写入到 buff，追加到末尾，符合多行行为
	if m.buff.Len() != 0 {
		// 换行符只添加到有效数据后面
		m.buff.Write(newLine)
	}
	m.buff.Write(text)

	// buff 长度超过限制
	if m.buff.Len() > m.opt.maxLength {
		m.lastFlushState = FlushPartial
		return m.Flush(), m.lastFlushState
	}

	// 不符合 flush 条件，文本数据都在 buff 中
	return nil, Written
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

func (m *Multiline) FlushString() string {
	if m.buff.Len() == 0 {
		return ""
	}
	text := m.buff.String()
	m.buff.Reset()
	return text
}

var asciiSpace = [256]uint8{'\t': 1, '\n': 1, '\v': 1, '\f': 1, '\r': 1, ' ': 1}

func TrimRightSpace(s []byte) []byte {
	end := len(s)
	for ; end > 0; end-- {
		c := s[end-1]
		if c >= utf8.RuneSelf {
			return bytes.TrimFunc(s[:end], unicode.IsSpace)
		}
		if asciiSpace[c] == 0 {
			break
		}
	}
	return s[:end]
}
