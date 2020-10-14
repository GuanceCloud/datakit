package io

import (
	"encoding/base64"
	"encoding/json"
	"net/url"
	"runtime"
	"time"

	"github.com/gorilla/websocket"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/system/rtpanic"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/git"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	tgi "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/telegraf_inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/kodo/wsmsg"
)

const (
	MTypeOnline            string = "online"
	MTypeHeartbeat         string = "heartbeat"
	MTypeGetInput          string = "get_input_config"
	MTypeGetEnableInput    string = "get_enabled_input_config"
	MTypeUpdateEnableInput string = "update_enabled_input_config"
	MTypeSetEnableInput    string = "set_enabled_input_config"
	MTypeDisableInput      string = "disable_input_config"
	MTypeReload            string = "reload"
	MTypeTestInput         string = "test_input_config"
)

var (
	cli   *wscli
	wsurl *url.URL
)

type wscli struct {
	c    *websocket.Conn
	id   string
	exit chan interface{}
}

func startWS() {
	var wsurl *url.URL
	l.Infof("ws start")
	wsurl = datakit.Cfg.MainCfg.DataWay.BuildWSURL(datakit.Cfg.MainCfg)

	l.Infof("ws server config %s", wsurl)
	cli = &wscli{
		id:   datakit.Cfg.MainCfg.UUID,
		exit: make(chan interface{}),
	}

	cli.tryConnect(wsurl.String())

	go func() {
		cli.waitMsg()
	}()

	cli.sendHeartbeat()
}

func (wc *wscli) tryConnect(wsurl string) {

	for {
		c, resp, err := websocket.DefaultDialer.Dial(wsurl, nil)
		if err != nil {
			l.Errorf("websocket.DefaultDialer.Dial(): %s", err.Error())
			time.Sleep(time.Second * 3)
			continue
		}

		_ = resp

		wc.c = c
		l.Infof("ws ready")
		break
	}
}

func (wc *wscli) reset() {
	wc.c.Close()
	wc.tryConnect(wsurl.String())
	wc.exit = make(chan interface{})

	// heartbeat worker will exit on @wc.exit, waitMsg() will not, so just restart the heartbeat worker only
	go func() {
		cli.sendHeartbeat()
	}()

}

func (wc *wscli) waitMsg() {

	var f rtpanic.RecoverCallback
	f = func(trace []byte, _ error) {
		for {

			defer rtpanic.Recover(f, nil)

			if trace != nil {
				l.Warn("recover ok: %s", string(trace))
			}

			_, resp, err := wc.c.ReadMessage()
			if err != nil {
				l.Error(err)
				wc.reset()
				continue
			}

			wm, err := wsmsg.ParseWrapMsg(resp)
			if err != nil {
				l.Error("msg.ParseWrapMsg(): %s", err.Error())
				continue
			}
			l.Infof("dk hand message %s", wm)

			wc.handle(wm)
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
		heart, err := time.ParseDuration(datakit.Cfg.MainCfg.DataWay.Heartbeat)
		l.Infof("heartbeat config %s", datakit.Cfg.MainCfg.DataWay.Heartbeat)
		l.Infof("heart,%s", heart)
		if err != nil {
			l.Error(err)
		}
		tick := time.NewTicker(heart)
		defer tick.Stop()

		for {
			select {
			case <-tick.C:
				wm, err := wsmsg.BuildMsg(m)
				if err != nil {
					l.Error(err)
				}

				_ = wc.sendText(wm)

			case <-wc.exit:
				return

			case <-datakit.Exit.Wait():
				l.Info("ws heartbeat exit")
				return
			}
		}
	}

	f(nil, nil)
}

func (wc *wscli) sendText(wm *wsmsg.WrapMsg) error {
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
	case MTypeOnline:
		wc.OnlineInfo(wm)
	case MTypeHeartbeat:
		wm.B64Data = wc.id

	}
	return wc.sendText(wm)
}

func (wc *wscli) OnlineInfo(wm *wsmsg.WrapMsg) {

	m := wsmsg.MsgDatakitOnline{
		UUID:            wc.id,
		Name:            datakit.Cfg.MainCfg.Name,
		Version:         git.Version,
		OS:              runtime.GOOS,
		Arch:            runtime.GOARCH,
		Heartbeat:       datakit.Cfg.MainCfg.DataWay.Heartbeat,
		AvailableInputs: GetAvailableInputs(),
		EnabledInputs:   GetEnableInputs(),
	}
	wm.Dest = []string{wc.id}
	wm.B64Data = ToBase64(m)
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
		//n, cfgs := inputs.InputEnabled(k)
		AvailableInputs = append(AvailableInputs, k)
	}
	return AvailableInputs
}

func GetEnableInputs() []string {
	var EnableInputs []string
	for k, _ := range inputs.Inputs {
		n, _ := inputs.InputEnabled(k)
		if n > 0 {
			EnableInputs = append(EnableInputs,k)
		}
	}

	for k, _ := range tgi.TelegrafInputs {
		n, _ := inputs.InputEnabled(k)
		if n > 0 {
			EnableInputs = append(EnableInputs,k)
		}
	}
	return EnableInputs
}