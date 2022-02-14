package tailer

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/worker"
)

/*
支持 socket 接收日志.
*/

const (
	ReadBufferLen = 1024 * 4
)

var l = logger.DefaultSLogger("socketLog")

type server struct {
	addr string       // udpConns's or tcpListeners's  key
	lis  net.Listener // tcp listener
	conn net.Conn     // udp.Conn
}

type socketLogger struct {
	// 存放连接，释放连接使用
	tcpListeners    map[string]net.Listener
	udpConns        map[string]net.Conn
	socketBufferLen int // read buffer lens
	ignorePatterns  []string
	tags            map[string]string
	// 配置
	opt  *Option
	stop chan struct{}

	servers []*server
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
		tags: buildTags(opt.GlobalTags),
		opt:  opt,
		stop: make(chan struct{}, 1),
	}
	if err := sl.opt.init(); err != nil {
		return nil, err
	}
	l = logger.SLogger("socketLog")
	return sl, nil
}

func buildTags(globalTags map[string]string) map[string]string {
	tags := make(map[string]string)
	for k, v := range globalTags {
		tags[k] = v
	}
	if _, ok := tags["log_source"]; !ok {
		tags["log_source"] = "socket"
	}
	return tags
}

func (sl *socketLogger) Start() {
	if len(sl.opt.Sockets) == 0 {
		sl.opt.log.Warnf("logging sockets is empty")
		return
	}
	for _, socket := range sl.opt.Sockets {
		s, err := mkServer(socket)
		if err != nil {
			l.Error(err)
			return
		}
		sl.servers = append(sl.servers, s)
	}

	// 配置无误之后 开始accept
	sl.toReceive()
}

func mkServer(socket string) (s *server, err error) {
	s = &server{addr: socket}
	socketURL, err := url.Parse(socket)
	if err != nil {
		return s, fmt.Errorf("error socket config err=%w", err)
	}
	network := socketURL.Scheme
	listenAddr := socketURL.Host
	l.Infof("check logging socket Scheme=%s listenerAddr=%s", network, listenAddr)
	switch network {
	case "", "tcp", "tcp4", "tcp6": // 建议使用tcp
		listener, err := net.Listen(network, listenAddr)
		if err != nil {
			return s, fmt.Errorf("socket listen port error:%w", err)
		}
		s.lis = listener
	case "udp", "udp4", "udp6":
		udpAddr, err := net.ResolveUDPAddr(network, listenAddr)
		if err != nil {
			return s, fmt.Errorf("resolve UDP addr error:%w", err)
		}
		conn, err := net.ListenUDP(network, udpAddr)
		if err != nil {
			return s, fmt.Errorf(" net.ListenUDP error:%w", err)
		}
		s.conn = conn
	default:
		err = fmt.Errorf("socket config like this: socket=[tcp://127.0.0.1:9540] , and please check your logging.conf")
	}
	return s, err
}

// toReceive: 根据listen或udp.conn 开始接收数据.
func (sl *socketLogger) toReceive() {
	if sl.servers == nil || len(sl.servers) == 0 {
		return
	}
	for _, s := range sl.servers {
		if s.lis != nil {
			sl.tcpListeners[s.addr] = s.lis
			l.Infof("TCP port:%s start to accept", s.addr)
			dkio.FeedEventLog(&dkio.Reporter{Message: fmt.Sprintf("[DataKit-logging] First Message. sockets log port: %s is open, source: %s",
				s.addr, sl.opt.Source), Logtype: "event"})
			go sl.accept(s.lis)
		}
		if s.conn != nil {
			sl.udpConns[s.addr] = s.conn
			l.Infof("UDP port:%s start to accept", s.addr)
			dkio.FeedEventLog(&dkio.Reporter{Message: fmt.Sprintf("[DataKit-logging] First Message. sockets log port: %s is open, source: %s",
				s.addr, sl.opt.Source), Logtype: "event"})
			go sl.doSocket(s.conn)
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
			sl.sendToPipeline(pipDate)
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

func (sl *socketLogger) sendToPipeline(pending []string) {
	taskDates := make([]worker.TaskData, 0)
	for _, data := range pending {
		if data != "" {
			taskDates = append(taskDates, &SocketTaskData{Tag: sl.tags, Log: data, Source: sl.opt.Source})
		}
	}
	if len(taskDates) != 0 {
		task := &worker.Task{
			TaskName:   "socklogging/" + sl.opt.InputName,
			ScriptName: sl.opt.Pipeline,
			Source:     sl.opt.Source,
			Data:       taskDates,
			Opt: &worker.TaskOpt{
				IgnoreStatus:          sl.opt.IgnoreStatus,
				DisableAddStatusField: sl.opt.DisableAddStatusField,
			},
			TS: time.Now(),
		}
		// 阻塞型channel
		_ = worker.FeedPipelineTaskBlock(task)
	}
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
