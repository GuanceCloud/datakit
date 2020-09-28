package io

import (
	"encoding/json"
	"net/url"
	"time"

	"github.com/gorilla/websocket"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/system/rtpanic"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/kodo/wsmsg"
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

	wsurl = datakit.Cfg.MainCfg.DataWay.BuildWSURL(datakit.Cfg.MainCfg)
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

			mtype, resp, err := wc.c.ReadMessage()
			if err != nil {
				l.Error(err)
				wc.reset()
				continue
			}

			_ = mtype // TODO:

			wm, err := wsmsg.ParseWrapMsg(resp)
			if err != nil {
				l.Error("msg.ParseWrapMsg(): %s", err.Error())
				continue
			}

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

		tick := time.NewTicker(time.Minute) // FIXME: interval should be configurable
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
	}
	return nil
	// TODO
}
