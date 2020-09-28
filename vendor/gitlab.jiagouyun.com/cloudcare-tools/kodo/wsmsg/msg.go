package wsmsg

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
)

var (
	l = logger.DefaultSLogger("kodows/msg")
)

type WrapMsg struct {
	Type    int      `json:"type"`
	ID      string   `json:"id"`
	Dest    []string `json:"dest,omitempty"`
	B64Data string   `json:"b64data,omitempty"`

	raw []byte
}

type Msg interface {
	Handle(*WrapMsg) error
}

func ParseWrapMsg(data []byte) (*WrapMsg, error) {

	var wm WrapMsg
	if err := json.Unmarshal(data, &wm); err != nil {
		return nil, err
	}

	return &wm, nil
}

func (wm *WrapMsg) Invalid() bool {
	return (wm.Type == 0 || wm.ID == "")
}

func (wm *WrapMsg) Send() {
	msgch <- wm
}

func SendToDatakit(rawmsg []byte) {

	wm, err := ParseWrapMsg([]byte(rawmsg))
	if err != nil {
		l.Errorf("json.Unmarshal(): %s", err.Error())
		return
	}

	wm.raw = rawmsg
	msgch <- wm
}

func (wm *WrapMsg) Handle() error {
	raw, err := base64.StdEncoding.DecodeString(wm.B64Data)
	if err != nil {
		return err
	}

	l.Debugf("handle msg %+#v", wm)

	switch wm.Type {
	case MTypeHeartbeat:
		var m MsgDatakitHeartbeat
		if err := json.Unmarshal(raw, &m); err != nil {
			return err
		}

		return m.Handle(wm)

	case MTypeInputConfig:
		var m MsgInputConfig
		if err := json.Unmarshal(raw, &m); err != nil {
			return err
		}

		return m.Handle(wm)

	case MTypeGetInputConfig, MTypeSetInputConfig: // should not been here

		// case: xxx:
		// TODO: echo resp to redis key, some HTTP request looping on it
		// ...
		return fmt.Errorf("should not been here")

	default:
		return fmt.Errorf("unknown msg type: %d", wm.Type)
	}
}

func BuildMsg(m interface{}, dest ...string) (*WrapMsg, error) {
	j, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	wm := &WrapMsg{
		ID:      cliutils.XID("wmsg_"),
		Dest:    dest,
		B64Data: base64.StdEncoding.EncodeToString(j),
	}

	switch m.(type) {
	case MsgDatakitHeartbeat:
		wm.Type = MTypeHeartbeat
	case MsgGetInputConfig:
		wm.Type = MTypeGetInputConfig
	case MsgSetInputConfig:
		wm.Type = MTypeSetInputConfig
	case MsgInputConfig:
		wm.Type = MTypeInputConfig

		// TODO: add more type

	default:
		return nil, fmt.Errorf("unknown msg type")
	}

	return wm, nil
}

//
// These msg will wrapped within ws.Msg as the field B64Data
//

// datakit heartbeat
type MsgDatakitHeartbeat struct {
	UUID string `json:"id"`
}

func (m *MsgDatakitHeartbeat) Handle(_ *WrapMsg) error {
	hbch <- m.UUID
	return nil
}

// get datakit input config
type MsgGetInputConfig struct {
	Names []string `json:"names"`
}

func (m *MsgGetInputConfig) Handle(wm *WrapMsg) error {
	// TODO
	return nil
}

type MsgInputConfig struct {
	Configs map[string][]string `json:"configs"`
}

func (m *MsgInputConfig) Handle(wm *WrapMsg) error {
	// TODO
	return nil
}

type MsgSetInputConfig struct {
	Configs map[string][]string `json:"configs"`
}

func (m *MsgSetInputConfig) Handle(wm *WrapMsg) error {
	// TODO
	return nil
}

const (
	// msgs upload from datakit
	MTypeHeartbeat int = iota + 1

	MTypeInputConfig

	// msgs send to datakit
	MTypeGetInputConfig
	MTypeSetInputConfig

	// ...
)
