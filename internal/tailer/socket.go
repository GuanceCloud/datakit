package tailer

import (
	"errors"
	"net"
	"strings"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/worker"
)

/*
支持socket接受日志.
*/
const (
	name          = "socketLog"
	MaxPending    = 200
	ReadBufferLen = 1024 * 4
	defSendTime   = time.Second * 5
)

var l = logger.DefaultSLogger("socketLog")

type sample struct {
	data []string
}

type socketLogger struct {
	// 存放连接，释放连接使用
	tcpListeners    map[string]net.Listener
	udpConns        map[string]net.Conn
	socketBufferLen int // read buffer lens
	ignorePatterns  []string

	// 配置
	opt  *Option
	stop chan struct{}

	// pending 发送到pipline模块
	pending []worker.TaskData
	msg     chan *sample
}

func NewWithOpt(opt *Option, ignorePatterns ...[]string) (sl *socketLogger, err error) {
	sl = &socketLogger{
		tcpListeners:    make(map[string]net.Listener),
		udpConns:        make(map[string]net.Conn),
		socketBufferLen: ReadBufferLen,
		ignorePatterns: func() []string {
			if len(ignorePatterns) > 0 {
				return ignorePatterns[0]
			}
			return nil
		}(),
		opt:     opt,
		stop:    make(chan struct{}, 1),
		pending: make([]worker.TaskData, 0),
		msg:     make(chan *sample, 10),
	}
	if err := sl.opt.init(); err != nil {
		return nil, err
	}
	l = logger.SLogger("socketLog")
	return sl, nil
}

func (sl *socketLogger) Start() {
	if len(sl.opt.Sockets) == 0 {
		sl.opt.log.Warnf("no socket config")
		return
	}
	go sl.startReceive()
	for _, socket := range sl.opt.Sockets {
		strs := strings.Split(socket, "://")
		if len(strs) != 2 {
			sl.opt.log.Warnf("no socket config")
			continue
		}
		network := strs[0]
		listenAddr := strs[1]
		sl.opt.log.Infof("network=%s addr=%s", strs[0], strs[1])
		switch network {
		case "", "tcp", "tcp4", "tcp6": // 建议使用tcp
			listener, err := net.Listen(network, listenAddr)
			if err != nil {
				sl.opt.log.Warnf("socket listen port error:%v", err)
				continue
			}
			l.Infof("logging is listener TCP addr=%s", listenAddr)
			sl.tcpListeners[socket] = listener
			go sl.accept(listener)
		case "udp", "udp4", "udp6":
			udpAddr, err := net.ResolveUDPAddr(network, listenAddr)
			if err != nil {
				sl.opt.log.Warnf("resolve UDP addr error:%v", err)
				continue
			}
			l.Infof("logging is listener udp addr=%s", listenAddr)
			conn, err := net.ListenUDP(network, udpAddr)
			if err == nil {
				go sl.doSocket(conn)
			}
		default:
			l.Warnf("socket config like this: socket=[tcp://127.0.0.1:9540] , and please check your logging.conf")
		}
	}
}

func (sl *socketLogger) accept(listener net.Listener) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return
			}
			sl.opt.log.Warnf("Error accepting:%s", err.Error())
			continue
		}
		go sl.doSocket(conn)
	}
}

func (sl *socketLogger) doSocket(conn net.Conn) {
	var cacheLine string
	for {
		data := make([]byte, sl.socketBufferLen)
		n, err := conn.Read(data)
		// see:$GOROOT/src/io/io.go:83
		if err != nil && n == 0 {
			return
		}
		var pipDate []string
		var cacheM string
		pipDate, cacheM = sl.spiltBuffer(cacheLine, string(data[:n]), n == sl.socketBufferLen)
		cacheLine = cacheM
		if len(pipDate) != 0 {
			sl.msg <- &sample{pipDate}
		}
	}
}

func (sl *socketLogger) spiltBuffer(fromCache string, date string, full bool) (pipdata []string, cacheDate string) {
	lines := strings.Split(date, "\n")
	logLen := len(lines)
	if full && logLen == 1 {
		fromCache += lines[0]
		return pipdata, fromCache
	}
	lines[0] = fromCache + lines[0]
	if strings.HasSuffix(date, "\n") {
		pipdata = append(pipdata, lines[0:len(lines)-1]...)
	} else {
		cacheDate = lines[logLen-1]
		pipdata = append(pipdata, lines[0:logLen-1]...)
	}
	return pipdata, cacheDate
}

func (sl *socketLogger) startReceive() {
	for {
		select {
		case d := <-sl.msg:
			for _, datum := range d.data {
				if datum != "" {
					sl.pending = append(sl.pending, &SocketTaskData{Tag: sl.opt.GlobalTags, Log: datum, Source: sl.opt.Source})
				}
				l.Debugf("datam = %s", datum)
			}
			if len(sl.pending) > MaxPending {
				err := sl.sendToPip(sl.pending)
				if err != nil {
					// 保留最新data
					sl.pending = sl.pending[len(d.data)-1:]
				} else {
					sl.pending = make([]worker.TaskData, 0)
				}
			}
		case <-time.After(defSendTime):
			if len(sl.pending) > 0 {
				err := sl.sendToPip(sl.pending)
				if err == nil {
					sl.pending = make([]worker.TaskData, 0)
				}
			}
		case <-sl.stop:
			return
		}
	}
}

type SocketTaskData struct {
	Log    string
	Source string
	Tag    map[string]string
}

func (std *SocketTaskData) GetContent() string {
	return std.Log
}

func (std *SocketTaskData) Handler(result *worker.Result) error {
	// result.SetSource(std.source)
	if std.Tag != nil && len(std.Tag) != 0 {
		for k, v := range std.Tag {
			result.SetTag(k, v)
		}
	}
	return nil
}

func (sl *socketLogger) sendToPip(pending []worker.TaskData) error {
	task := &worker.Task{
		TaskName:   name,
		ScriptName: sl.opt.Pipeline,
		Source:     sl.opt.Source,
		Data:       pending,
		Opt: &worker.TaskOpt{
			IgnoreStatus:          sl.opt.IgnoreStatus,
			DisableAddStatusField: sl.opt.DisableAddStatusField,
		},
		TS: time.Now(),
	}
	return worker.FeedPipelineTask(task)
}

func (sl *socketLogger) Close() {
	sl.stop <- struct{}{}
	for _, listener := range sl.tcpListeners {
		err := listener.Close()
		sl.opt.log.Infof("close tcp port err=%v", err)
	}
	for _, listener := range sl.udpConns {
		err := listener.Close()
		sl.opt.log.Infof("close udp port err=%v", err)
	}
}
