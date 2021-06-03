package docker

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/pkg/stdcopy"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	iod "gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
)

const (
	// Maximum bytes of a log line before it will be split, size is mirroring
	// docker code:
	// https://github.com/moby/moby/blob/master/daemon/logger/copier.go#L21
	maxLineBytes = 16 * 1024

	// ES value can be at most 32766 bytes long
	maxFieldsLength = 32766

	pipelineTimeField = "time"

	useIOHighFreq = true
)

func (this *Input) addToContainerList(containerID string, cancel context.CancelFunc) error {
	this.mu.Lock()
	defer this.mu.Unlock()
	this.containerLogList[containerID] = cancel
	return nil
}

func (this *Input) removeFromContainerList(containerID string) error {
	this.mu.Lock()
	defer this.mu.Unlock()
	delete(this.containerLogList, containerID)
	return nil
}

func (this *Input) containerInContainerList(containerID string) bool {
	this.mu.Lock()
	defer this.mu.Unlock()
	_, ok := this.containerLogList[containerID]
	return ok
}

func (this *Input) cancelTails() error {
	this.mu.Lock()
	defer this.mu.Unlock()
	for _, cancel := range this.containerLogList {
		cancel()
	}
	return nil
}

func (this *Input) gatherLog() {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, this.apiTimeoutDuration)
	defer cancel()

	cList, err := this.client.ContainerList(ctx, this.opts)
	if err != nil {
		l.Error(err)
		iod.FeedLastError(inputName, fmt.Sprintf("gather logging: %s", err.Error()))
		return
	}

	for _, container := range cList {
		if this.containerInContainerList(container.ID) {
			continue
		}

		ctx, cancel := context.WithCancel(context.Background())
		this.addToContainerList(container.ID, cancel)

		// Start a new goroutine for every new container that has logs to collect
		this.wg.Add(1)
		go func(container types.Container) {
			defer this.wg.Done()
			defer this.removeFromContainerList(container.ID)

			err = this.tailContainerLogs(ctx, container)
			if err != nil && err != context.Canceled {
				l.Error(err)
				iod.FeedLastError(inputName, fmt.Sprintf("gather logging: %s", err.Error()))
			}
		}(container)
	}
}

func (this *Input) tailContainerLogs(ctx context.Context, container types.Container) error {
	// ignore imageVersion
	imageName, _ := ParseImage(container.Image)
	containerName := getContainerName(container.Names)

	tags := map[string]string{
		"container_name": containerName,
		"container_id":   container.ID,
		"image_name":     imageName,
	}
	for k, v := range this.Tags {
		tags[k] = v
	}

	hasTTY, err := this.hasTTY(ctx, container)
	if err != nil {
		return err
	}

	logReader, err := this.client.ContainerLogs(ctx, container.ID, this.containerLogsOptions)
	if err != nil {
		return err
	}

	var source string
	if contianerIsFromKubernetes(getContainerName(container.Names)) {
		uid, err := this.kubernetes.GatherPodUID(container.ID)
		if err != nil {
			l.Debugf("gather k8s podUID error: %s", err)
		} else {
			name, err := this.kubernetes.GatherWorkName(uid)
			if err != nil {
				l.Debugf("gather k8s workname error: %s", err)
			} else {
				source = name
			}
		}
	}

	// If the container is using a TTY, there is only a single stream
	// (stdout), and data is copied directly from the container output stream,
	// no extra multiplexing or headers.
	//
	// If the container is *not* using a TTY, streams for stdout and stderr are
	// multiplexed.

	if hasTTY {
		return tailStream(logReader, "tty", container, this.LogFilters, tags, source)
	} else {
		return tailMultiplexed(logReader, container, this.LogFilters, tags, source)
	}

}

func (this *Input) hasTTY(ctx context.Context, container types.Container) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, this.apiTimeoutDuration)
	defer cancel()
	c, err := this.client.ContainerInspect(ctx, container.ID)
	if err != nil {
		return false, err
	}
	return c.Config.Tty, nil
}

func tailStream(reader io.ReadCloser, stream string, container types.Container, logFilters LogFilters, baseTags map[string]string, source string) error {
	defer reader.Close()

	tags := make(map[string]string, len(baseTags)+1)
	for k, v := range baseTags {
		tags[k] = v
	}
	tags["stream"] = stream

	r := bufio.NewReaderSize(reader, maxLineBytes)

	for {
		line, err := r.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		if len(line) == 0 {
			continue
		}

		message := strings.TrimSpace(string(line))

		containerName := getContainerName(container.Names)
		// measurement 默认使用容器名，如果该容器是 k8s 创建，则尝试获取它的 work name（work-load）
		// 如果该字段值（即 source 参数）不为空，则使用
		var measurement = containerName
		if source != "" {
			measurement = source
		}

		var fields = make(map[string]interface{})

		for _, lf := range logFilters {
			if lf.MatchMessage(message) {
				if lf.Source != "" {
					measurement = lf.Source
				}

				var err error
				fields, err = lf.RunPipeline(message)
				if err != nil {
					l.Debug(err)
				}

				if lf.Service != "" {
					fields["service"] = lf.Service
				}
				break
			}
		}

		// 没有对应的 logFilters
		if len(fields) == 0 {
			fields["service"] = containerName
			fields["message"] = message
		}
		// 额外添加
		fields["from_kubernetes"] = contianerIsFromKubernetes(containerName)

		// l.Debugf("get %d bytes from source: %s", len(message), measurement)

		if err := checkFieldsLength(fields, maxFieldsLength); err != nil {
			// 只有在碰到非 message 字段，且长度超过最大限制时才会返回 error
			// 防止通过 pipeline 添加巨长字段的恶意行为
			l.Error(err)
			continue
		}

		addStatus(fields)

		// pipeline切割的日志时间
		ts, err := takeTime(fields)
		if err != nil {
			ts = time.Now()
			l.Error(err)
		}

		pt, err := iod.MakePoint(measurement, tags, fields, ts)
		if err != nil {
			l.Error(err)
		} else {
			if err := iod.Feed(inputName, datakit.Logging, []*iod.Point{pt}, &iod.Option{HighFreq: useIOHighFreq}); err != nil {
				l.Error(err)
			}
		}
	}
}

func tailMultiplexed(src io.ReadCloser, container types.Container, lf LogFilters, baseTags map[string]string, source string) error {
	outReader, outWriter := io.Pipe()
	errReader, errWriter := io.Pipe()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := tailStream(outReader, "stdout", container, lf, baseTags, source)
		if err != nil {
			l.Error(err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := tailStream(errReader, "stderr", container, lf, baseTags, source)
		if err != nil {
			l.Error(err)
		}
	}()

	_, err := stdcopy.StdCopy(outWriter, errWriter, src)
	outWriter.Close()
	errWriter.Close()
	src.Close()
	wg.Wait()
	return err
}

// Adapts some of the logic from the actual Docker library's image parsing
// routines:
// https://github.com/docker/distribution/blob/release/2.7/reference/normalize.go
func ParseImage(image string) (string, string) {
	domain := ""
	remainder := ""

	i := strings.IndexRune(image, '/')

	if i == -1 || (!strings.ContainsAny(image[:i], ".:") && image[:i] != "localhost") {
		remainder = image
	} else {
		domain, remainder = image[:i], image[i+1:]
	}

	imageName := ""
	imageVersion := "unknown"

	i = strings.LastIndex(remainder, ":")
	if i > -1 {
		imageVersion = remainder[i+1:]
		imageName = remainder[:i]
	} else {
		imageName = remainder
	}

	if domain != "" {
		imageName = domain + "/" + imageName
	}

	return imageName, imageVersion
}

func takeTime(fields map[string]interface{}) (ts time.Time, err error) {
	// time should be nano-second
	if v, ok := fields[pipelineTimeField]; ok {
		nanots, ok := v.(int64)
		if !ok {
			err = fmt.Errorf("invalid filed `%s: %v', should be nano-second, but got `%s'",
				pipelineTimeField, v, reflect.TypeOf(v).String())
			return
		}

		ts = time.Unix(nanots/int64(time.Second), nanots%int64(time.Second))
		delete(fields, pipelineTimeField)
	} else {
		ts = time.Now()
	}

	return
}

// checkFieldsLength 指定字段长度 "小于等于" maxlength
func checkFieldsLength(fields map[string]interface{}, maxlength int) error {
	for k, v := range fields {
		switch vv := v.(type) {
		// FIXME:
		// need  "case []byte" ?
		case string:
			if len(vv) <= maxlength {
				continue
			}
			if k == "message" {
				fields[k] = vv[:maxlength]
			} else {
				return fmt.Errorf("fields: %s, length=%d, out of maximum length", k, len(vv))
			}
		default:
			// nil
		}
	}
	return nil
}

var statusMap = map[string]string{
	"f":        "emerg",
	"emerg":    "emerg",
	"a":        "alert",
	"alert":    "alert",
	"c":        "critical",
	"critical": "critical",
	"e":        "error",
	"error":    "error",
	"w":        "warning",
	"warning":  "warning",
	"i":        "info",
	"info":     "info",
	"d":        "debug",
	"trace":    "debug",
	"verbose":  "debug",
	"debug":    "debug",
	"o":        "OK",
	"s":        "OK",
	"ok":       "OK",
}

func addStatus(fields map[string]interface{}) {
	// map 有 "status" 字段
	statusField, ok := fields["status"]
	if !ok {
		fields["status"] = "info"
		return
	}
	// "status" 类型必须是 string
	statusStr, ok := statusField.(string)
	if !ok {
		fields["status"] = "info"
		return
	}

	// 查询 statusMap 枚举表并替换
	if v, ok := statusMap[strings.ToLower(statusStr)]; !ok {
		fields["status"] = "info"
	} else {
		fields["status"] = v
	}
}
