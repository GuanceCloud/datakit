package wsmsg

import (
	"encoding/base64"
	"encoding/json"
	"net"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/kodo/config"
	"gitlab.jiagouyun.com/cloudcare-tools/kodo/models"
)

var (
	l         = logger.DefaultSLogger("kodows/msg")
	Msgch     = make(chan *WrapMsg)
	Hbch      = make(chan string)
	Clich     = make(chan *DatakitClient)
	Onlinedks = map[string]*DatakitClient{}
)

type DatakitClient struct {
	UUID     string
	Version  string
	OS       string
	Arch     string
	Docker   bool
	Token    string
	HostName string
	Ip       string
	Conn     net.Conn

	HeartbeatConf string
	Heartbeat     time.Time
	Status        string
}

type WrapMsg struct {
	Type    string   `json:"type"`
	ID      string   `json:"id"`
	Dest    []string `json:"dest,omitempty"`
	B64Data string   `json:"b64data,omitempty"`

	Code string `json:"code"`
	Raw  []byte
}

func SendOnline(dk *DatakitClient) {
	var m MsgDatakitOnline
	wm, err := BuildMsg(m, dk.UUID)
	if err != nil {
		l.Error(err)
	}
	b, err := json.Marshal(wm)
	if err != nil {
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

	l.Debugf("handle msg %+#v", wm)
	switch wm.Type {
	case MTypeHeartbeat:
		var m MsgDatakitHeartbeat
		if err := json.Unmarshal(raw, &m); err != nil {
			return err
		}
		return m.Handle(wm)
	case MTypeOnline:
		wm.SetRedis()
		SetDatakitOnline(wm)
		return nil
	case MtypeModifyDKStatus:
		Onlinedks[wm.Dest[0]].Status = "reload"
		return nil
	default:
		wm.SetRedis()
		return nil
	}
}

func SetDatakitOnline(wm *WrapMsg) {
	if wm.Code != "ok" {
		l.Errorf("online err:%s", wm)
		return
	}
	data, err := base64.StdEncoding.DecodeString(wm.B64Data)
	if err != nil {
		l.Errorf("ws return err:%s", err)
		return
	}
	var dk MsgDatakitOnline

	err = json.Unmarshal(data, &dk)
	if err != nil {
		l.Errorf("ws online parse err : %s", err)
		return
	}
	now := time.Now().Unix()
	info, _ := json.Marshal(dk.InputInfo)

	var count = 0
	err = models.Stmts["existDK"].QueryRow(dk.UUID).Scan(&count)
	if count == 1 {
		_, err = models.Stmts["updateDKOnline"].Exec(dk.Name, Onlinedks[dk.UUID].Token, Onlinedks[dk.UUID].HostName, Onlinedks[dk.UUID].Ip, dk.Version, dk.OS, dk.Arch, info, now, now, 0, dk.UUID)

	} else {
		uuid := cliutils.XID("dkol_")
		_, err = models.Stmts[`setDKOnline`].Exec(uuid, dk.Name, Onlinedks[dk.UUID].Token, Onlinedks[dk.UUID].HostName, Onlinedks[dk.UUID].Ip, dk.UUID, dk.Version, dk.OS, dk.Arch, info, now, now, models.StatusOK, now, now)
	}

	if err != nil {
		l.Errorf("set online run mysql err:%s", err)
	}
}

func (wm *WrapMsg) SetRedis() {
	//dkId := wm.Dest[0]
	//token := Onlinedks[dkId].Token
	b, err := json.Marshal(&wm)
	if err != nil {
		l.Errorf("set redis parse wm err:%s", err)
	}
	if err := config.Redis.Publish(wm.Type, b).Err(); err != nil {
		l.Errorf("redis publish err:%s", err.Error())
	}
	l.Debugf("set redis ok")

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

// datakit heartbeat
type MsgDatakitHeartbeat struct {
	UUID string `json:"id"`
}

func (m *MsgDatakitHeartbeat) Handle(_ *WrapMsg) error {
	Hbch <- m.UUID
	return nil
}

type MsgDatakitOnline struct {
	UUID      string
	Version   string
	OS        string
	Arch      string
	Name      string
	Heartbeat string
	InputInfo map[string]interface{}
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

	return json.Unmarshal(data, &m.Names)
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

	return json.Unmarshal(data, &m.Configs)
}

const (
	MTypeOnline         string = "online"
	MTypeHeartbeat      string = "heartbeat"
	MTypeGetInput       string = "get_input_config"
	MTypeGetEnableInput string = "get_enabled_input_config"
	MTypeSetInput       string = "set_input_config"
	MTypeDisableInput   string = "disable_input_config"
	MTypeReload         string = "reload"
	MTypeTestInput      string = "test_input_config"
	MTypeEnableInput    string = "enable_input_config"
	MTypeCMD            string = "cmd"
	MtypeCsharkCmd      string = "csharkCmd"
	MtypeModifyDKStatus string = "modifyStatus"
)
