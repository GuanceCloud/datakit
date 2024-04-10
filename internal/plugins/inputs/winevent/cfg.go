// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.
// Some code modified from project Beats (https://github.com/elastic/beats).

//go:build windows
// +build windows

package winevent

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"strconv"
	"sync"
	"time"
	"unicode/utf16"
	"unicode/utf8"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
)

const defaultEventFetchSize = 100

var (
	l = logger.DefaultSLogger(inputName)
	// nolint:lll
	query = `<QueryList>
    <Query Id="0" Path="Security">
      <Select Path="Security">*</Select>
      <Suppress Path="Security">*[System[( (EventID &gt;= 5152 and EventID &lt;= 5158) or EventID=5379 or EventID=4672)]]</Suppress>
    </Query>
    <Query Id="1" Path="Application">
      <Select Path="Application">*[System[(Level &lt;= 4)]]</Select>
    </Query>
    <Query Id="2" Path="Windows PowerShell">
      <Select Path="Windows PowerShell">*[System[(Level &lt; 4)]]</Select>
    </Query>
    <Query Id="3" Path="System">
      <Select Path="System">*</Select>
    </Query>
    <Query Id="4" Path="Setup">
      <Select Path="Setup">*</Select>
    </Query>
  </QueryList>`

	keywordsMap = map[int64]string{
		0:                "AnyKeyword",
		0x1000000000000:  "Response Time",
		0x4000000000000:  "WDI Diag",
		0x8000000000000:  "SQM",
		0x10000000000000: "Audit Failure",
		0x20000000000000: "Audit Success",
		0x40000000000000: "Correlation Hint",
		0x80000000000000: "Classic",
	}

	opcodesMap = map[uint8]string{
		0: "Info",
		1: "Start",
		2: "Stop",
		3: "DCStart",
		4: "DCStop",
		5: "Extension",
		6: "Reply",
		7: "Resume",
		8: "Suspend",
		9: "Send",
	}

	levelsMap = map[uint8]string{
		0: "Information", // "Log Always", but Event Viewer shows Information.
		1: "Critical",
		2: "Error",
		3: "Warning",
		4: "Information",
		5: "Verbose",
	}
	tasksMap = map[uint16]string{
		0: "None",
	}
)

type Input struct {
	winMetaCache
	Query          string            `toml:"xpath_query"`
	EventFetchSize uint32            `toml:"event_fetch_size"`
	Tags           map[string]string `toml:"tags,omitempty"`

	subscription EvtHandle
	buf          []byte
	collectCache []*point.Point

	handleCache *handleCache

	mergedTags map[string]string

	semStop       *cliutils.Sem // start stop signal
	feeder        dkio.Feeder
	Tagger        datakit.GlobalTagger
	subscribeFlag EvtSubscribeFlag
}
type HexInt64 uint64

func (v *HexInt64) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var s string
	if err := d.DecodeElement(&s, &start); err != nil {
		return err
	}

	num, err := strconv.ParseUint(s, 0, 64)
	if err != nil {
		// Ignore invalid version values.
		return err
	}

	*v = HexInt64(num)
	return nil
}

type WinMeta struct {
	Keywords map[int64]string  // Keyword bit mask to keyword name.
	Opcodes  map[uint8]string  // Opcode value to name.
	Levels   map[uint8]string  // Level value to name.
	Tasks    map[uint16]string // Task value to name.
}

type winMetaCache struct {
	ttl    time.Duration
	logger *logger.Logger

	mu    sync.RWMutex
	cache map[string]winMetaCacheEntry
}

func (c *winMetaCache) winMeta(provider string) *WinMeta {
	c.mu.RLock()
	e, ok := c.cache[provider]
	c.mu.RUnlock()
	if ok && time.Until(e.expire) > 0 {
		return e.WinMeta
	}

	// Upgrade lock.
	defer c.mu.Unlock()
	c.mu.Lock()

	// Did the cache get updated during lock upgrade?
	// No need to check expiry here since we must have a new entry
	// if there is an entry at all.
	if e, ok := c.cache[provider]; ok {
		return e.WinMeta
	}

	s, err := NewPublisherMetadataStore(NilHandle, provider, c.logger)
	if err != nil {
		// Return an empty store on error (can happen in cases where the
		// log was forwarded and the provider doesn't exist on collector).
		s = NewEmptyPublisherMetadataStore(provider, c.logger)
		l.Warn("failed to load publisher metadata for %s (returning an empty metadata store): %s", provider, err.Error())
	}
	s.Close()
	c.cache[provider] = winMetaCacheEntry{expire: time.Now().Add(c.ttl), WinMeta: &s.WinMeta}
	return &s.WinMeta
}

type winMetaCacheEntry struct {
	expire time.Time
	*WinMeta
}

// EventIdentifier is the identifier that the provider uses to identify a
// specific event type.
type EventIdentifier struct {
	Qualifiers uint16 `xml:"Qualifiers,attr"`
	ID         uint32 `xml:",chardata"`
}

type Event struct {
	Source        Provider        `xml:"System>Provider"`
	EventID       EventIdentifier `xml:"System>EventID"`
	Version       int             `xml:"System>Version"`
	LevelRaw      uint8           `xml:"System>Level"`
	TaskRaw       uint16          `xml:"System>Task"`
	OpcodeRaw     *uint8          `xml:"System>Opcode"`
	KeywordsRaw   HexInt64        `xml:"System>Keywords"`
	TimeCreated   TimeCreated     `xml:"System>TimeCreated"`
	EventRecordID int             `xml:"System>EventRecordID"`
	Correlation   Correlation     `xml:"System>Correlation"`
	Execution     Execution       `xml:"System>Execution"`
	Channel       string          `xml:"System>Channel"`
	Computer      string          `xml:"System>Computer"`
	Security      Security        `xml:"System>Security"`
	UserData      UserData        `xml:"UserData"`
	EventData     EventData       `xml:"EventData"`

	Message  string   `xml:"RenderingInfo>Message"`
	Level    string   `xml:"RenderingInfo>Level"`
	Task     string   `xml:"RenderingInfo>Task"`
	Opcode   string   `xml:"RenderingInfo>Opcode"`
	Keywords []string `xml:"RenderingInfo>Keywords>Keyword"`

	KeywordsText string
}

// KeyValue is a key value pair of strings.
type KeyValue struct {
	Key   string
	Value string
}

// EventData Application-provided XML data.
type EventData struct {
	Pairs []KeyValue `xml:",any"`
}

// UserData contains the event data.
type UserData struct {
	Name  xml.Name
	Pairs []KeyValue
}

// UnmarshalXML unmarshals UserData XML.
func (u *UserData) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	// Assume that UserData has the same general key-value structure as
	// EventData does.
	in := struct {
		Pairs []KeyValue `xml:",any"`
	}{}

	// Read tokens until we find the first StartElement then unmarshal it.
	for {
		t, err := d.Token()
		if err != nil {
			return err
		}

		if se, ok := t.(xml.StartElement); ok {
			err = d.DecodeElement(&in, &se)
			if err != nil {
				return err
			}

			u.Name = se.Name
			u.Pairs = in.Pairs
			err = d.Skip()
			if err != nil {
				return err
			}
			break
		}
	}

	return nil
}

// Provider is the Event provider information.
type Provider struct {
	Name string `xml:"Name,attr"`
}

// Correlation is used for the event grouping.
type Correlation struct {
	ActivityID        string `xml:"ActivityID,attr"`
	RelatedActivityID string `xml:"RelatedActivityID,attr"`
}

// Execution Info for Event.
type Execution struct {
	ProcessID   uint32 `xml:"ProcessID,attr"`
	ThreadID    uint32 `xml:"ThreadID,attr"`
	ProcessName string
}

// Security Data for Event.
type Security struct {
	UserID string `xml:"UserID,attr"`
}

// TimeCreated field for Event.
type TimeCreated struct {
	SystemTime string `xml:"SystemTime,attr"`
}

func DecodeUTF16(b []byte) ([]byte, error) {
	if len(b)%2 != 0 {
		return nil, fmt.Errorf("must have even length byte slice")
	}

	u16s := make([]uint16, 1)

	ret := &bytes.Buffer{}

	b8buf := make([]byte, 4)

	lb := len(b)
	for i := 0; i < lb; i += 2 {
		u16s[0] = uint16(b[i]) + (uint16(b[i+1]) << 8)
		r := utf16.Decode(u16s)
		n := utf8.EncodeRune(b8buf, r[0])
		ret.Write(b8buf[:n])
	}

	return ret.Bytes(), nil
}
