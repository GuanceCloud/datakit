package wsmsg

import (
	"net"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/network/ws"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/system/rtpanic"
	"gitlab.jiagouyun.com/cloudcare-tools/kodo"
)

type DatakitClient struct {
	UUID    string
	Version string
	OS      string
	Arch    string
	Docker  bool
	Token   string
	Conn    net.Conn

	heartbeat time.Time
}

func (dc *DatakitClient) Online() {
	clich <- dc
}

var (
	clich = make(chan *DatakitClient)
	hbch  = make(chan string)
	msgch = make(chan *WrapMsg)
)

func DatakitManager() {

	l = logger.SLogger("kodows/msg")

	tick := time.NewTicker(time.Minute)
	defer tick.Stop()

	var f rtpanic.RecoverCallback

	onlinedks := map[string]*DatakitClient{}

	f = func(_ []byte, _ error) {
		defer rtpanic.Recover(f, nil)

		for {
			select {
			case cli := <-clich:
				onlinedks[cli.UUID] = cli

			case id := <-hbch:
				if c, ok := onlinedks[id]; !ok {
					l.Warnf("datakit %s not online", id)
				} else {
					c.heartbeat = time.Now()
				}

			case m := <-msgch:

				for _, id := range m.Dest {
					dk, ok := onlinedks[id]
					if !ok {
						l.Warnf("datakit %s not found", id)
						continue
					}

					if err := ws.SendMsgToClient(m.raw, dk.Conn); err != nil {
						l.Errorf("ws.SendMsgToClient(): %s", err.Error())
					}
				}

			case <-tick.C:

				for _, dk := range onlinedks {
					if time.Since(dk.heartbeat) > time.Minute {
						l.Infof("remove datakit %s, last heartbeat %v", dk.UUID, dk.heartbeat)
						delete(onlinedks, dk.UUID)
						// TODO: fire keyevent: datakit lost
					}
				}
				l.Infof("total online datakit: %d", len(onlinedks))

			case <-kodo.Exit.Wait():
				for _, dk := range onlinedks {
					if err := dk.Conn.Close(); err != nil {
						l.Warn("dk.Conn.Close(): %s", err.Error())
					}
				}

				l.Info("datakit manager exit.")
				return
			}
		}
	}

	f(nil, nil)
}
