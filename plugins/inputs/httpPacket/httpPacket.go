package httpPacket

import (
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/luascript"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	httpd "gitlab.jiagouyun.com/cloudcare-tools/datakit/http"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"

	"github.com/gin-gonic/gin"
)

const (
	inputName = "httpPacket"

	defaultMeasurement = "httpPacket"

	sampleCfg = `
[[inputs.httpPacket]]
    ## http server route path
    ## required
    path = "/httpPacket"
    lua_files = []
`
)

var l = logger.DefaultSLogger(inputName)

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &HttpPacket{}
	})
}

type HttpPacket struct {
	Path     string   `toml:"path"`
	LuaFiles []string `toml:"lua_files"`
	ls       *luascript.LuaScript

	enable bool
	mut    sync.Mutex
}

func (*HttpPacket) SampleConfig() string {
	return sampleCfg
}

func (*HttpPacket) Catalog() string {
	return inputName
}

func (h *HttpPacket) Run() {
	l = logger.SLogger(inputName)

	if h.initCfg() {
		return
	}

	h.ls.Run()
	h.enable = true

	l.Infof("httpPacket input started...")

	for {
		select {
		case <-datakit.Exit.Wait():
			h.stop()
			return
		default:
		}
	}
}

func (h *HttpPacket) stop() {
	h.mut.Lock()
	h.enable = false
	h.mut.Unlock()

	h.ls.Stop()
}

func (h *HttpPacket) initCfg() bool {
	var err error
	h.mut = sync.Mutex{}
	h.ls = luascript.NewLuaScript(2)

	for {
		select {
		case <-datakit.Exit.Wait():
			return true
		default:
			if len(h.LuaFiles) != 0 {
				err = h.ls.AddLuaCodesFromFile("points", h.LuaFiles)
				if err != nil {
					l.Error(err)
					goto lable
				}
			}
		}
		break
	lable:
		time.Sleep(time.Second)
	}

	l.Debugf("init lua success")
	l.Debugf("points lua: %v", h.LuaFiles)

	return false
}

func (h *HttpPacket) RegHttpHandler() {
	httpd.RegGinHandler("POST", h.Path, h.handle)
}

func (h *HttpPacket) handle(c *gin.Context) {
	if !h.enable {
		l.Warnf("worker does not exist")
		return
	}

	body, err := ioutil.ReadAll(c.Request.Body)
	defer c.Request.Body.Close()

	l.Debug("read body", string(body))

	if err != nil {
		l.Errorf("read body, %s", err.Error())
		c.Writer.WriteHeader(http.StatusBadRequest)
		return
	}

	if len(h.LuaFiles) == 0 {
		if err := io.NamedFeed(body, datakit.Metric, inputName); err != nil {
			l.Error(err)
			c.Writer.WriteHeader(http.StatusBadRequest)
			return
		}

		l.Debug("point", string(body))

		c.Writer.WriteHeader(http.StatusOK)
		return
	}

	pts, err := ParsePoints(body, "ns")
	if err != nil {
		l.Errorf("parse points, %s", err.Error())

		c.Writer.WriteHeader(http.StatusBadRequest)
		return
	}

	// new Points struct
	p, err := NewPointsData("points", datakit.Metric, pts)
	if err != nil {
		l.Errorf("new points data, %s", err.Error())

		c.Writer.WriteHeader(http.StatusBadRequest)
		return
	}

	// send point
	err = h.ls.SendData(p)

	if err != nil {
		l.Error(err)

		c.Writer.WriteHeader(http.StatusBadRequest)
		return
	}

	c.Writer.WriteHeader(http.StatusOK)
}
