//+build windows

package winevent

import (
	"bytes"
	"fmt"
	"unicode/utf16"
	"unicode/utf8"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	l     = logger.DefaultSLogger(inputName)
	query = `<QueryList>
    <Query Id="0" Path="Security">
      <Select Path="Security">*</Select>
      <Suppress Path="Security">*[System[( (EventID &gt;= 5152 and EventID &lt;= 5158) or EventID=5379 or EventID=4672)]]</Suppress>
    </Query>
    <Query Id="1" Path="Application">
      <Select Path="Application">*[System[(Level &lt; 4)]]</Select>
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
)

type Input struct {
	Query string            `toml:"xpath_query"`
	Tags  map[string]string `toml:"tags,omitempty"`

	subscription EvtHandle
	buf          []byte
	collectCache []inputs.Measurement
}

type Event struct {
	Source        Provider    `xml:"System>Provider"`
	EventID       int         `xml:"System>EventID"`
	Version       int         `xml:"System>Version"`
	Level         int         `xml:"System>Level"`
	Task          int         `xml:"System>Task"`
	Opcode        int         `xml:"System>Opcode"`
	Keywords      string      `xml:"System>Keywords"`
	TimeCreated   TimeCreated `xml:"System>TimeCreated"`
	EventRecordID int         `xml:"System>EventRecordID"`
	Correlation   Correlation `xml:"System>Correlation"`
	Execution     Execution   `xml:"System>Execution"`
	Channel       string      `xml:"System>Channel"`
	Computer      string      `xml:"System>Computer"`
	Security      Security    `xml:"System>Security"`
	UserData      UserData    `xml:"UserData"`
	EventData     EventData   `xml:"EventData"`
	Message       string
	LevelText     string
	TaskText      string
	OpcodeText    string
}

// UserData Application-provided XML data
type UserData struct {
	InnerXML []byte `xml:",innerxml"`
}

// EventData Application-provided XML data
type EventData struct {
	InnerXML []byte `xml:",innerxml"`
}

// Provider is the Event provider information
type Provider struct {
	Name string `xml:"Name,attr"`
}

// Correlation is used for the event grouping
type Correlation struct {
	ActivityID        string `xml:"ActivityID,attr"`
	RelatedActivityID string `xml:"RelatedActivityID,attr"`
}

// Execution Info for Event
type Execution struct {
	ProcessID   uint32 `xml:"ProcessID,attr"`
	ThreadID    uint32 `xml:"ThreadID,attr"`
	ProcessName string
}

// Security Data for Event
type Security struct {
	UserID string `xml:"UserID,attr"`
}

// TimeCreated field for Event
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
