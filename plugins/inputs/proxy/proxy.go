package proxy

import (
	"bytes"
	"compress/gzip"
	"context"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/koding/websocketproxy"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/luascript"
	uhttp "gitlab.jiagouyun.com/cloudcare-tools/cliutils/network/http"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	httpd "gitlab.jiagouyun.com/cloudcare-tools/datakit/http"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"

	"github.com/gin-gonic/gin"
)

const (
	inputName = "proxy"

	defaultMeasurement = "proxy"

	sampleCfg = `
[[inputs.proxy]]
    ## http server route path
		## required: don't change
    path = "/proxy"
	  ws_bind = "0.0.0.0:5588"
`
)

var l = logger.DefaultSLogger(inputName)

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &Proxy{}
	})
}

type Proxy struct {
	Path   string `toml:"path"`
	WSBind string `toml:"ws_bind,bind"`

	PointsLuaFiles []string            `toml:"-"`
	ObjectLuaFiles []string            `toml:"-"`
	Global         []map[string]string `toml:"-"`

	ls   *luascript.LuaScript
	cron *luascript.LuaCron

	enable bool
	mut    sync.Mutex
}

func (*Proxy) SampleConfig() string {
	return sampleCfg
}

func (*Proxy) Catalog() string {
	return inputName
}

func (*Proxy) Test() (*inputs.TestResult, error) {
	test := &inputs.TestResult{}
	return test, nil
}

func (d *Proxy) Run() {
	l = logger.SLogger(inputName)

	if d.initCfg() {
		return
	}

	if len(d.PointsLuaFiles) != 0 || len(d.ObjectLuaFiles) != 0 {
		d.ls.Run()
	}

	if len(d.Global) != 0 {
		d.cron.Run()
	}

	d.enable = true

	wsurl := datakit.Cfg.MainCfg.DataWay.BuildWSURL(datakit.Cfg.MainCfg)
	wsurl.RawQuery = ""
	server := &http.Server{Addr: d.WSBind, Handler: websocketproxy.NewProxy(wsurl)}
	l.Infof("[info] starting WebSocket proxy on %s, remote: %s", d.WSBind, wsurl.String())
	go func() {
		if err := server.ListenAndServe(); err != nil {
			l.Error(err.Error())
			return
		}
	}()
	l.Infof("proxy input started...")

	select {
	case <-datakit.Exit.Wait():
		d.stop()
		if err := server.Shutdown(context.Background()); err != nil {
			l.Errorf("[error] shutdown websocket server: %s", err.Error())
		}
		l.Infof("[info] websocketproxy closed ")
		return
	}
}

func (d *Proxy) stop() {
	d.mut.Lock()
	d.enable = false
	d.mut.Unlock()

	d.ls.Stop()
	d.cron.Stoping()
}

func (d *Proxy) initCfg() bool {
	var err error
	d.mut = sync.Mutex{}
	d.cron = luascript.NewLuaCron()
	d.ls = luascript.NewLuaScript(2)

	for {
		select {
		case <-datakit.Exit.Wait():
			return true

		default:
			for _, global := range d.Global {
				err = d.cron.AddLuaFromFile(global["lua_file"], global["schedule"])
				if err != nil {
					l.Error(err)
					goto lable
				}
			}

			if d.PointsLuaFiles != nil {
				err = d.ls.AddLuaCodesFromFile("points", d.PointsLuaFiles)
				if err != nil {
					l.Error(err)
					goto lable
				}
			}

			if d.ObjectLuaFiles != nil {
				err = d.ls.AddLuaCodesFromFile("object", d.ObjectLuaFiles)
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
	l.Debugf("crontab lua list: %v", d.Global)
	l.Debugf("points lua: %v", d.PointsLuaFiles)
	l.Debugf("object lua: %v", d.ObjectLuaFiles)

	return false
}

func (d *Proxy) RegHttpHandler() {
	httpd.RegGinHandler("POST", d.Path, d.handle)
}

func (d *Proxy) handle(c *gin.Context) {
	if !d.enable {
		l.Warnf("worker does not exist")
		return
	}

	category := c.Query("category")

	gz := (c.Request.Header.Get("Content-Encoding") == "gzip")

	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		l.Errorf("read body, %s", err.Error())
		goto end
	}
	defer c.Request.Body.Close()

	if gz {
		r, err := gzip.NewReader(bytes.NewReader(body))
		if err != nil {
			l.Errorf("NewReader(): %s", err.Error())
			uhttp.HttpErr(c, err)
			return
		}

		body, err = ioutil.ReadAll(r)
		if err != nil {
			l.Errorf("ReadAll(): %s", err.Error())
			uhttp.HttpErr(c, err)
			return
		}
	}

	l.Debugf("receive data, category: %s, len(%d bytes)", category, len(body))

	switch category {
	case io.Metric, io.Logging, io.KeyEvent, io.Tracing, io.Rum:
		if len(d.PointsLuaFiles) == 0 {
			if err := io.NamedFeed(body, category, inputName); err != nil {
				l.Error(err)
			}
			goto end
		}

		pts, err := ParsePoints(body, "ns")
		if err != nil {
			l.Errorf("parse points, %s", err.Error())
			goto end
		}

		p, err := NewPointsData("points", category, pts)
		if err != nil {
			l.Errorf("new points data, %s", err.Error())
			goto end
		}

		err = d.ls.SendData(p)
		if err != nil {
			l.Error(err)
			goto end
		}

	case io.Object:
		if len(d.ObjectLuaFiles) == 0 {
			if err := io.NamedFeed(body, category, inputName); err != nil {
				l.Error(err)
			}
			goto end
		}

		j, err := NewObjectData("object", category, body)
		if err != nil {
			l.Error(err)
			goto end
		}

		err = d.ls.SendData(j)
		if err != nil {
			l.Error(err)
			goto end
		}

	default:
		l.Errorf("invalid category: `%s'", category)
	}

end:
	c.Writer.WriteHeader(http.StatusOK)
}
