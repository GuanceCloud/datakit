package wsmsg

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/kodo/config"
)

var (
	l     = logger.DefaultSLogger("kodows/msg")
	Msgch = make(chan *WrapMsg)
	Hbch  = make(chan string)
	Clich = make(chan *DatakitClient)
	Onlinedks = map[string]*DatakitClient{}


)


type DatakitClient struct {
	UUID    string
	Version string
	OS      string
	Arch    string
	Docker  bool
	Token   string
	Conn    net.Conn

	Heartbeat time.Time
}

type WrapMsg struct {
	Type    string   `json:"type"`
	ID      string   `json:"id"`
	Dest    []string `json:"dest,omitempty"`
	B64Data string   `json:"b64data,omitempty"`

	Code string
	Raw  []byte
}



func SendOnline(dk *DatakitClient){
	var m MsgDatakitOnline
	wm,err:= BuildMsg(m,dk.UUID)
	if err !=nil{
		l.Error(err)
	}


	b,err:= json.Marshal(wm)
	if err !=nil{
		l.Error(err)
	}
	SendToDatakit(b)
}


func ParseWrapMsg(data []byte) (*WrapMsg, error) {

	var wm WrapMsg
	if err := json.Unmarshal(data, &wm); err != nil {
		return nil, err
	}

	return &wm, nil
}

func (wm *WrapMsg) Invalid() bool {
	return wm.ID == ""
}

func (dc *DatakitClient) Online() {
	Clich <- dc
	SendOnline(dc)
}

func (wm *WrapMsg) Send() {
	Msgch <- wm
}

func SendToDatakit(rawmsg []byte) {

	wm, err := ParseWrapMsg(rawmsg)
	if err != nil {
		l.Errorf("json.Unmarshal(): %s", err.Error())
		return
	}
	wm.Raw = rawmsg
	Msgch <- wm
}

func (wm *WrapMsg) Handle() error {
	raw, err := base64.StdEncoding.DecodeString(wm.B64Data)
	if err != nil {
		return err
	}

	l.Infof("handle msg %+#v", wm)

	switch wm.Type {
	case MTypeHeartbeat:
		var m MsgDatakitHeartbeat
		if err := json.Unmarshal(raw, &m); err != nil {
			return err
		}
		return m.Handle(wm)
	case MTypeOnline:
		wm.SetRedis()
		//TODO set mysql
		return nil
	case MTypeGetInput,MTypeDisableInput,MTypeGetEnableInput,MTypeSetInput,MTypeUpdateEnableInput,MTypeReload,MTypeTestInput:
		wm.SetRedis()
		return nil


	default:
		return fmt.Errorf("unknown msg type: %s", wm.Type)
	}
}

func (wm *WrapMsg)SetRedis(){
	dkId := wm.Dest[0]
	token := Onlinedks[dkId].Token
	b,err := json.Marshal(&wm)
	if err != nil{
		l.Errorf("set redis parse wm err:%s",err)
	}
	config.RedisCli.Set(fmt.Sprintf("%s-%s-%s",token,dkId,wm.ID),b,time.Second*30)
	l.Info("set redis ok")
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
	case MsgDatakitOnline:
		wm.Type = MTypeOnline
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
	Hbch <- m.UUID
	return nil
}

type MsgDatakitOnline struct {
	UUID            string
	Version         string
	OS              string
	Arch            string
	Name            string
	Heartbeat       string
	EnabledInputs   []string
	AvailableInputs []string
}

// get datakit input config
type MsgGetInputConfig struct {
	Names []string `json:"names"`
}

func (m *MsgGetInputConfig) Handle(wm *WrapMsg) error {
	data, err := base64.StdEncoding.DecodeString(wm.B64Data)
	if err != nil {
		l.Errorf("get inputs config err %s", err)
		return err
	}

	return json.Unmarshal(data,&m.Names)
}


type MsgSetInputConfig struct {
	Configs map[string]map[string]string `json:"configs"`
}

func (m *MsgSetInputConfig) Handle(wm *WrapMsg) error {
	data, err := base64.StdEncoding.DecodeString(wm.B64Data)
	if err != nil {
		l.Errorf("get inputs config err %s", err)
		return err
	}

	return json.Unmarshal(data,&m.Configs)
}

const (
	MTypeOnline            string = "online"
	MTypeHeartbeat         string = "heartbeat"
	MTypeGetInput          string = "get_input_config"
	MTypeGetEnableInput    string = "get_enabled_input_config"
	MTypeUpdateEnableInput string = "update_enabled_input_config"
	MTypeSetInput          string = "set_input_config"
	MTypeDisableInput      string = "disable_input_config"
	MTypeReload            string = "reload"
	MTypeTestInput         string = "test_input_config"
)
