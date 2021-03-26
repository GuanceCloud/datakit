package http

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"runtime"
	"time"

	"github.com/gorilla/websocket"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/system/rtpanic"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/git"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/cshark"
	tgi "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/telegraf_inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/kodo/wsmsg"
)

var (
	cli     *wscli
	wsurl   *url.URL
	wsReset = make(chan interface{})
)

type wscli struct {
	c  *websocket.Conn
	id string
}

func (c *wscli) setup() {

	if err := cli.tryConnect(wsurl.String()); err != nil {
		return
	}

	l.Info("ws connect ok")

	go func() {
		//defer datakit.WG.Done()
		cli.waitMsg() // blocking reader, do not add to datakit.WG
	}()
	datakit.WG.Add(1)
	go func() {
		defer datakit.WG.Done()
		cli.sendHeartbeat()
	}()
}

func StartWS() {
	l.Infof("ws start")
	wsurl = datakit.Cfg.MainCfg.DataWay.BuildWSURL(datakit.Cfg.MainCfg)
	l.Infof(wsurl.String())

	cli = &wscli{
		id: datakit.Cfg.MainCfg.UUID,
	}

	cli.setup()

	for {
		select {
		case <-datakit.Exit.Wait():
			l.Info("start ws exit")
			if cli.c != nil {
				l.Info("ws closed")
				cli.c.Close()
			}
			return
		case <-wsReset:
			l.Info("start ws on reset")
			if cli.c != nil {
				l.Info("ws closed")

				cli.c.Close()
			}
			cli.setup()

		}
	}
}

// blocking wait ws connect ok
func (wc *wscli) tryConnect(wsurl string) error {

	for {
		select {
		case <-datakit.Exit.Wait():
			return fmt.Errorf("ws not ready:wsurl:%s", wsurl)

		default:
			c, resp, err := websocket.DefaultDialer.Dial(wsurl, nil)
			if err != nil {
				l.Errorf("websocket.DefaultDialer.Dial(): %s", err.Error())
				time.Sleep(time.Second * 3)
				continue
			}

			l.Debugf("ws connect ok, resp: %+#v", resp)

			wc.c = c
			return nil
		}
	}
}

func (wc *wscli) waitMsg() {

	var f rtpanic.RecoverCallback
	f = func(trace []byte, err error) {
		for {

			// TODO: if panic, fire a key event to dataflux

			defer rtpanic.Recover(f, nil)

			if trace != nil {
				l.Warnf("recover ok: %s err:", string(trace), err.Error())
			}

			select {
			case <-datakit.Exit.Wait():
				l.Info("wait message exit on global exit")
				return
			default:
			}

			l.Debug("waiting message...")
			_, resp, err := wc.c.ReadMessage()
			if err != nil {
				l.Errorf("ws read message error: %s", err)
				select {
				case wsReset <- nil:
				default:
					l.Info("wait message exit.")
					return
				}
			}
			wm, err := wsmsg.ParseWrapMsg(resp)
			if err != nil {
				l.Errorf("msg.ParseWrapMsg(): %s", err.Error())
				continue
			}

			l.Debugf("ws hand message %s", wm)

			if err := wc.handle(wm); err != nil {
				select {
				case wsReset <- nil:
				default:
					return
				}
			}
		}
	}

	f(nil, nil)
}

func (wc *wscli) sendHeartbeat() {
	m := wsmsg.MsgDatakitHeartbeat{UUID: wc.id}

	var f rtpanic.RecoverCallback
	f = func(trace []byte, _ error) {

		defer rtpanic.Recover(f, nil)
		if trace != nil {
			l.Warn("recover ok: %s", string(trace))
		}
		heartbeatTime := datakit.Cfg.MainCfg.DataWay.Heartbeat
		if heartbeatTime == "" {
			heartbeatTime = "30s"
		}
		heart, err := time.ParseDuration(heartbeatTime)

		if err != nil {
			l.Error(err)
		}
		tick := time.NewTicker(heart)
		defer tick.Stop()

		for {
			wm, err := wsmsg.BuildMsg(m)
			if err != nil {
				l.Error(err)
			}

			err = wc.sendText(wm)
			if err != nil {
				select {
				case wsReset <- nil:
				default:
					return
				}
			}
			select {
			case <-tick.C:
			case <-datakit.Exit.Wait():
				l.Info("ws heartbeat exit")
				return
			}
		}
	}

	f(nil, nil)
}

func (wc *wscli) sendText(wm *wsmsg.WrapMsg) error {
	wm.Dest = []string{datakit.Cfg.MainCfg.UUID}
	j, err := json.Marshal(wm)
	if err != nil {
		l.Error(err)
		return err
	}

	if err := wc.c.WriteMessage(websocket.TextMessage, j); err != nil {
		l.Errorf("WriteMessage(): %s", err.Error())

		return err
	}

	return nil
}

func (wc *wscli) handle(wm *wsmsg.WrapMsg) error {
	switch wm.Type {
	case wsmsg.MTypeOnline:
		wc.OnlineInfo(wm)
	case wsmsg.MtypeCsharkCmd:
		wc.CsharkCmd(wm)
	default:
		wc.SetMessage(wm, "error", fmt.Errorf("unknow type %s ", wm.Type).Error())
	}
	return wc.sendText(wm)
}

func (wc *wscli) CsharkCmd(wm *wsmsg.WrapMsg) {
	data, _ := base64.StdEncoding.DecodeString(wm.B64Data)
	_, ok := inputs.InputsInfo["cshark"]
	if !ok {
		wc.SetMessage(wm, "error", "input cshark not enable")
		return
	}
	err := cshark.SendCmdOpt(string(data))
	if err != nil {
		wc.SetMessage(wm, "error", err.Error())
		return
	}
	wc.SetMessage(wm, "ok", "")

}

func (wc *wscli) SetMessage(wm *wsmsg.WrapMsg, code string, Message interface{}) {
	wm.Code = code
	if Message != "" {
		wm.B64Data = ToBase64(Message)
	} else {
		wm.B64Data = ""
	}

}

func (wc *wscli) OnlineInfo(wm *wsmsg.WrapMsg) {
	m := wsmsg.MsgDatakitOnline{
		UUID:      wc.id,
		Name:      datakit.Cfg.MainCfg.Name,
		Version:   git.Version,
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
		Heartbeat: datakit.Cfg.MainCfg.DataWay.Heartbeat,
		InputInfo: map[string]interface{}{},
	}
	m.InputInfo["availableInputs"] = GetAvailableInputs()
	m.InputInfo["enabledInputs"] = GetEnableInputs()
	state, err := io.GetStats()
	if err != nil {
		l.Errorf("get state err:%s", err.Error())
		state = []*io.InputsStat{}
	}
	m.InputInfo["state"] = state

	wc.SetMessage(wm, "ok", m)
}

func ToBase64(wm interface{}) string {
	body, err := json.Marshal(wm)
	if err != nil {
		l.Errorf("%s toBase64 err:%s", wm, err)
	}
	return base64.StdEncoding.EncodeToString(body)
}

func GetAvailableInputs() []string {
	var AvailableInputs []string
	for k, _ := range inputs.Inputs {
		AvailableInputs = append(AvailableInputs, k)
	}
	for k, _ := range tgi.TelegrafInputs {
		AvailableInputs = append(AvailableInputs, k)
	}
	return AvailableInputs
}

func GetEnableInputs() (Enable []string) {
	for k, _ := range inputs.Inputs {
		n, _ := inputs.InputEnabled(k)
		if n > 0 {
			Enable = append(Enable, k)
		}
	}

	for k, _ := range tgi.TelegrafInputs {
		n, _ := inputs.InputEnabled(k)
		if n > 0 {
			Enable = append(Enable, k)
		}
	}
	return Enable
}
