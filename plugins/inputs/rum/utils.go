// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package rum

import (
	"archive/zip"
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	lp "github.com/GuanceCloud/cliutils/lineproto"
	uhttp "github.com/GuanceCloud/cliutils/network/http"
	"github.com/go-sourcemap/sourcemap"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	dkhttp "gitlab.jiagouyun.com/cloudcare-tools/datakit/http"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/cmds"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/path"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/ip2isp"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/ptinput/funcs"
)

const iosDSYMFilePath = "Contents/Resources/DWARF"

const CacheInitCap = 16

const (
	SdkWeb     = "df_web_rum_sdk"
	SdkAndroid = "df_android_rum_sdk"
	SdkIOS     = "df_ios_rum_sdk"
)

const (
	srcMapDirWeb     = "web"
	srcMapDirAndroid = "android"
	srcMapDirIOS     = "ios"
)

const (
	JavaCrash   = "java_crash"
	NativeCrash = "native_crash"
	IOSCrash    = "ios_crash"
)

var (
	NDKAvailableABI = map[string]struct{}{
		"armeabi-v7a": {},
		"arm64-v8a":   {},
		"x86":         {},
		"x86_64":      {},
	}
	srcMapDirs = map[string]string{
		SdkWeb:     srcMapDirWeb,
		SdkAndroid: srcMapDirAndroid,
		SdkIOS:     srcMapDirIOS,
	}
	rumMetricNames = map[string]bool{
		`view`:      true,
		`resource`:  true,
		`error`:     true,
		`long_task`: true,
		`action`:    true,
	}
	rumMetricAppID        = "app_id"
	sourceMapTokenBuckets = newExecCmdTokenBuckets(64)
	sourcemapCache        = make(map[string]map[string]*sourcemap.Consumer)
	sourcemapLock         sync.RWMutex
	latestCheckFileTime   = time.Now().Unix()
	uncompressLock        sync.Mutex
)

func loadSourcemapFile() error {
	rumDir := getRumSourcemapDir(srcMapDirWeb)
	files, err := ioutil.ReadDir(rumDir)
	if err != nil {
		log.Warnf("load rum sourcemap dir failed: %s", err.Error())
		return fmt.Errorf("load web source map file fial: %w", err)
	}

	sourcemapLock.Lock()
	defer sourcemapLock.Unlock()

	for _, file := range files {
		if !file.IsDir() {
			fileName := file.Name()
			if strings.HasSuffix(fileName, ".zip") {
				sourcemapItem, err := loadZipFile(filepath.Join(rumDir, fileName))
				if err != nil {
					log.Warnf("load zip file %s failed, %s", fileName, err.Error())
					continue
				}

				sourcemapCache[fileName] = sourcemapItem
			}
		}
	}

	return nil
}

func getRumSourcemapDir(sdkName string) string {
	dir, ok := srcMapDirs[sdkName]
	if !ok {
		dir = srcMapDirWeb
	}
	rumDir := filepath.Join(datakit.DataDir, "rum", dir)

	return rumDir
}

type execCmdTokenBuckets struct {
	buckets chan struct{}
}

func newExecCmdTokenBuckets(size int) *execCmdTokenBuckets {
	if size < 16 {
		size = 16
	}
	tb := &execCmdTokenBuckets{
		buckets: make(chan struct{}, size),
	}
	for i := 0; i < size; i++ {
		tb.buckets <- struct{}{}
	}

	return tb
}

func (e *execCmdTokenBuckets) getToken() struct{} {
	return <-e.buckets
}

func (e *execCmdTokenBuckets) sendBackToken(token struct{}) {
	e.buckets <- token
}

func geoTags(srcip string) (tags map[string]string) {
	// default set to be unknown
	tags = map[string]string{
		"city":     "unknown",
		"province": "unknown",
		"country":  "unknown",
		"isp":      "unknown",
		"ip":       srcip,
	}

	if srcip == "" {
		return
	}

	ipInfo, err := funcs.Geo(srcip)

	log.Debugf("ipinfo(%s): %+#v", srcip, ipInfo)

	if err != nil {
		log.Warnf("geo failed: %s, ignored", err)
		return
	}

	// avoid nil pointer error
	if ipInfo == nil {
		return tags
	}

	if len(ipInfo.City) > 0 {
		tags["city"] = ipInfo.City
	}
	if len(ipInfo.Region) > 0 {
		tags["province"] = ipInfo.Region
	}
	if len(ipInfo.Country) > 0 {
		tags["country"] = ipInfo.Country
	}
	if len(srcip) > 0 {
		tags["ip"] = srcip
	}

	if isp := ip2isp.SearchISP(srcip); len(isp) > 0 {
		tags["isp"] = isp
	}

	return tags
}

type jsonPoint struct {
	Measurement string                 `json:"measurement"`
	Tags        map[string]string      `json:"tags,omitempty"`
	Fields      map[string]interface{} `json:"fields"`
	Time        int64                  `json:"time,omitempty"`
}

// convert json point to lineproto point.
func (jp *jsonPoint) point(opt *lp.Option) (*point.Point, error) {
	p, err := lp.MakeLineProtoPoint(jp.Measurement, jp.Tags, jp.Fields, opt)
	if err != nil {
		return nil, err
	}

	return &point.Point{Point: p}, nil
}

func jsonPoints(body []byte, opt *lp.Option) ([]*point.Point, error) {
	var jps []jsonPoint
	if err := json.Unmarshal(body, &jps); err != nil {
		log.Error(err.Error())

		return nil, dkhttp.ErrInvalidJSONPoint
	}

	if opt == nil {
		opt = lp.DefaultOption
	}

	var pts []*point.Point
	for _, jp := range jps {
		if jp.Time != 0 { // use time from json point
			opt.Time = time.Unix(0, jp.Time)
		}

		if p, err := jp.point(opt); err != nil {
			log.Error(err.Error())

			return nil, uhttp.Error(dkhttp.ErrInvalidJSONPoint, err.Error())
		} else {
			pts = append(pts, p)
		}
	}

	return pts, nil
}

func contains(str string, list []string) bool {
	if len(list) == 0 {
		return true
	}
	for _, a := range list {
		if a == str {
			return true
		}
	}
	return false
}

func getSrcIP(ac *dkhttp.APIConfig, req *http.Request) (ip string) {
	if ac != nil {
		ip = req.Header.Get(ac.RUMOriginIPHeader)
		log.Debugf("get ip from %s: %s", ac.RUMOriginIPHeader, ip)
		if ip == "" {
			for k, v := range req.Header {
				log.Debugf("%s: %s", k, strings.Join(v, ","))
			}
		}
	} else {
		log.Info("apiConfig not set")
	}

	if ip != "" {
		log.Debugf("header remote addr: %s", ip)
		parts := strings.Split(ip, ",")
		if len(parts) > 0 {
			ip = parts[0] // 注意：此处只取第一个 IP 作为源 IP
			return
		}
	} else { // 默认取 http 框架带进来的 IP
		log.Debugf("gin remote addr: %s", req.RemoteAddr)
		host, _, err := net.SplitHostPort(req.RemoteAddr)
		if err == nil {
			ip = host
			return
		} else {
			log.Warnf("net.SplitHostPort(%s): %s, ignored", req.RemoteAddr, err)
		}
	}

	return ip
}

type iosCrashAddress struct {
	start     string
	end       string
	originStr string
}

func scanIOSCrashAddress(originErrStack string) (map[string]map[string][]iosCrashAddress, error) {
	// for match
	// 4   App                                         0x0000000104fd0728 0x104f30000 + 657192
	//
	// $1 "App"
	// $2 "0x0000000104fd0728"
	// $3 "0x104f30000 + 657192"  // 用于解析后对原始堆栈进行替换
	// $4 0x104f30000
	// $5 "657192"
	expStr := `(\S+)\s+(0x[0-9a-fA-F]+)\s+((0x[0-9a-fA-F]+)\s*\+\s*(\d+|0x[0-9a-fA-F]+))`
	re, err := regexp.Compile(expStr)
	if err != nil {
		return nil, fmt.Errorf("compile regexp [%s] fail: %w", expStr, err)
	}

	matches := re.FindAllStringSubmatch(originErrStack, -1)

	crashAddress := make(map[string]map[string][]iosCrashAddress)
	for _, match := range matches {
		if len(match) != 6 {
			continue
		}

		moduleName := match[1]
		startAddr := match[4]

		if _, ok := crashAddress[moduleName]; !ok {
			crashAddress[moduleName] = make(map[string][]iosCrashAddress)
		}

		crashAddress[moduleName][startAddr] = append(crashAddress[moduleName][startAddr], iosCrashAddress{
			start:     startAddr,
			end:       match[2],
			originStr: match[3],
		})
	}

	return crashAddress, nil
}

func scanABI(crashReport string) string {
	for _, fieldName := range []string{"ABI list:", "ABI:"} {
		idx := strings.Index(crashReport, fieldName)
		log.Infof("index of [%s] is %d", fieldName, idx)
		if idx > -1 {
			if eolIdx := strings.IndexAny(crashReport[idx:], "\r\n"); eolIdx > -1 {
				log.Infof("index of EOL is %d", eolIdx)
				abi := strings.Trim(strings.TrimSpace(crashReport[idx+len(fieldName):idx+eolIdx]), "'")
				abiList := []string{abi}
				if strings.ContainsRune(abi, ',') {
					abiList = strings.Split(abi, ",")
				}
				for _, abi := range abiList {
					if _, ok := NDKAvailableABI[abi]; ok {
						return abi
					}
				}
				log.Errorf("find illegal abi [%s], ignore", abi)
			}
		}
	}

	return ""
}

func checkJavaShrinkTool(mappingFile string) (string, error) {
	fp, err := os.Open(mappingFile) // nolint:gosec
	if err != nil {
		return "", fmt.Errorf("open mapping file [%s] fail: %w", mappingFile, err)
	}
	defer func() {
		_ = fp.Close()
	}()

	scanner := bufio.NewScanner(fp)

	rows := 0
	for scanner.Scan() && rows < 10 {
		rows++
		line := scanner.Text()
		if strings.HasPrefix(line, "#") &&
			(strings.Contains(line, "compiler: R8") || strings.Contains(line, "com.android.tools.r8.mapping")) {
			return cmds.AndroidCommandLineTools, nil
		}
	}
	return cmds.Proguard, nil
}

type simpleQueue struct {
	buckets []string
}

func newSimpleQueue(initCap ...uint) *simpleQueue {
	if len(initCap) > 0 {
		return &simpleQueue{
			buckets: make([]string, 0, initCap[0]),
		}
	}
	return &simpleQueue{}
}

func (sq *simpleQueue) IsEmpty() bool {
	return len(sq.buckets) == 0
}

func (sq *simpleQueue) enqueue(item string) {
	sq.buckets = append(sq.buckets, item)
}

func (sq *simpleQueue) dequeue() string {
	if sq.IsEmpty() {
		return ""
	}
	item := sq.buckets[0]
	sq.buckets = sq.buckets[1:]
	return item
}

func scanModuleSymbolFile(dir string, moduleName string) (string, error) {
	queue := newSimpleQueue(4)
	queue.enqueue(dir)

	for !queue.IsEmpty() {
		dir = queue.dequeue()
		files, _ := os.ReadDir(dir)

		for _, entry := range files {
			if strings.HasPrefix(entry.Name(), moduleName) &&
				strings.HasSuffix(strings.ToLower(entry.Name()), ".dsym") {
				if entry.IsDir() {
					// for path SomeName.app.dSYM/Contents/Resources/DWARF/SomeName
					dSYMFile := filepath.Join(dir, entry.Name(), iosDSYMFilePath, moduleName)
					if path.IsFileExists(dSYMFile) {
						return dSYMFile, nil
					} else {
						log.Infof("file [%s] not exists, ignore...", dSYMFile)
					}
				} else if entry.Type().IsRegular() {
					return filepath.Join(dir, entry.Name()), nil
				}
			}

			if entry.IsDir() {
				// 放入队列继续查找
				queue.enqueue(filepath.Join(dir, entry.Name()))
			}
		}
	}

	return "", fmt.Errorf("IOS dSYM symbol file for module[%s] not found", moduleName)
}

func getSourcemap(errStack string, sourcemapItem map[string]*sourcemap.Consumer) string {
	reg := regexp.MustCompile(`@ .*:\d+:\d+`)

	replaceStr := reg.ReplaceAllStringFunc(errStack, func(str string) string {
		return str[0:2] + getSourceMapString(str[2:], sourcemapItem)
	})
	return replaceStr
}

func getSourceMapString(str string, sourcemapItem map[string]*sourcemap.Consumer) string {
	parts := strings.Split(str, ":")
	partsLen := len(parts)
	if partsLen < 3 {
		return str
	}
	rowNumber, err := strconv.Atoi(parts[partsLen-2])
	if err != nil {
		return str
	}
	colNumber, err := strconv.Atoi(parts[partsLen-1])
	if err != nil {
		return str
	}

	srcPath := strings.Join(parts[:partsLen-2], ":") // http://localhost:5500/dist/bundle.js

	urlObj, err := url.Parse(srcPath)
	if err != nil {
		log.Debugf("parse url failed, %s, %s", srcPath, err.Error())
		return str
	}

	urlPath := strings.TrimPrefix(urlObj.Path, "/")
	sourceMapFileName := urlPath + ".map"

	smap, ok := sourcemapItem[sourceMapFileName]
	if !ok {
		log.Debugf("no sourcemap: %s", sourceMapFileName)
		return str
	}

	file, _, line, col, ok := smap.Source(rowNumber, colNumber)

	if ok {
		return fmt.Sprintf("%s:%v:%v", file, line, col)
	}

	return str
}

// GetSourcemapZipFileName  zip file name.
func GetSourcemapZipFileName(applicatinID, env, version string) string {
	fileName := fmt.Sprintf("%s-%s-%s.zip", applicatinID, env, version)

	return strings.ReplaceAll(fileName, string(filepath.Separator), "__")
}

func copyZipItem(item *zip.File, dst string) error {
	reader, err := item.Open()
	if err != nil {
		return fmt.Errorf("open zip item fail: %w", err)
	}
	defer func() {
		_ = reader.Close()
	}()

	dstDir := filepath.Dir(dst)
	if !path.IsDir(dstDir) {
		if err := os.MkdirAll(dstDir, 0o750); err != nil {
			return fmt.Errorf("mkdir [%s] fail: %w", dstDir, err)
		}
	}

	writer, err := os.Create(dst) // nolint:gosec
	if err != nil {
		return fmt.Errorf("create dst file fail: %w", err)
	}
	defer func() {
		_ = writer.Close()
	}()

	_, err = io.Copy(writer, reader) // nolint:gosec
	if err != nil {
		return fmt.Errorf("copy zip item fail: %w", err)
	}

	return nil
}

func uncompressZipFile(zipFileAbsPath string) error {
	zipFileAbsDir := strings.TrimSuffix(zipFileAbsPath, filepath.Ext(zipFileAbsPath))

	needUncompress := false

	dirStatInfo, dirErr := os.Stat(zipFileAbsDir)
	zipFileStat, fileErr := os.Stat(zipFileAbsPath)

	if os.IsNotExist(dirErr) || !dirStatInfo.IsDir() {
		if os.IsNotExist(fileErr) || !zipFileStat.Mode().IsRegular() {
			return fmt.Errorf("sourcemap zip file [%s] not exists", zipFileAbsPath)
		}
		needUncompress = true
	} else if fileErr == nil && zipFileStat.Mode().IsRegular() {
		if dirErr == nil && dirStatInfo.ModTime().Before(zipFileStat.ModTime()) {
			log.Infof("文件夹修改时间早于zip文件，重新解压缩zip包")
			needUncompress = true
		}
	}

	if needUncompress {
		// 加锁，避免消耗太多磁盘资源
		uncompressLock.Lock()
		defer uncompressLock.Unlock()

		zipFileDir := filepath.Dir(zipFileAbsPath)
		baseZipName := strings.TrimSuffix(filepath.Base(zipFileAbsPath), filepath.Ext(zipFileAbsPath))

		reader, err := zip.OpenReader(zipFileAbsPath)
		if err != nil {
			return fmt.Errorf("open zip file [%s] fail: %w", zipFileAbsPath, err)
		}
		defer func() {
			_ = reader.Close()
		}()

		for _, item := range reader.File {
			var absItemPath string
			if !strings.HasPrefix(filepath.Clean(item.Name), baseZipName) {
				absItemPath = filepath.Join(zipFileDir, baseZipName, item.Name) // nolint:gosec
			} else {
				absItemPath = filepath.Join(zipFileDir, item.Name) // nolint:gosec
			}

			if item.FileInfo().IsDir() {
				if err := os.MkdirAll(absItemPath, 0o750); err != nil {
					return fmt.Errorf("can not mkdir: %w", err)
				}
			} else if err := copyZipItem(item, absItemPath); err != nil {
				return err
			}
		}
	}

	return nil
}

func loadZipFile(zipFile string) (map[string]*sourcemap.Consumer, error) {
	sourcemapItem := make(map[string]*sourcemap.Consumer)

	zipReader, err := zip.OpenReader(zipFile)
	if err != nil {
		return nil, err
	}
	defer zipReader.Close() //nolint:errcheck

	for _, f := range zipReader.File {
		if !f.FileInfo().IsDir() && strings.HasSuffix(f.Name, ".map") {
			file, err := f.Open()
			if err != nil {
				log.Debugf("ignore sourcemap %s, %s", f.Name, err.Error())
				continue
			}
			defer file.Close() // nolint:errcheck

			content, err := ioutil.ReadAll(file)
			if err != nil {
				log.Debugf("ignore sourcemap %s, %s", f.Name, err.Error())
				continue
			}

			smap, err := sourcemap.Parse(f.Name, content)
			if err != nil {
				log.Debugf("sourcemap parse failed, %s", err.Error())
				continue
			}

			sourcemapItem[f.Name] = smap
		}
	}

	return sourcemapItem, nil
}

func updateSourcemapCache(zipFile string) error {
	fileName := filepath.Base(zipFile)
	if !strings.HasSuffix(fileName, ".zip") {
		return fmt.Errorf(`suffix name is not ".zip" [%s]`, zipFile)
	}
	sourcemapItem, err := loadZipFile(zipFile)
	if err != nil {
		log.Errorf("load zip file error: %s", err.Error())
		return fmt.Errorf("load zip file [%s] err: %w", zipFile, err)
	}
	sourcemapLock.Lock()
	defer sourcemapLock.Unlock()
	sourcemapCache[fileName] = sourcemapItem
	log.Debugf("load sourcemap success: %s", fileName)

	return nil
}

func deleteSourcemapCache(zipFile string) {
	fileName := filepath.Base(zipFile)
	if strings.HasSuffix(fileName, ".zip") {
		sourcemapLock.Lock()
		defer sourcemapLock.Unlock()
		delete(sourcemapCache, fileName)
	}
}

// isDomainName checks if a string is a presentation-format domain name,
// this func is copied from net package.
func isDomainName(s string) bool {
	// The root domain name is valid. See golang.org/issue/45715.
	if s == "." {
		return true
	}

	// See RFC 1035, RFC 3696.
	// Presentation format has dots before every label except the first, and the
	// terminal empty label is optional here because we assume fully-qualified
	// (absolute) input. We must therefore reserve space for the first and last
	// labels' length octets in wire format, where they are necessary and the
	// maximum total length is 255.
	// So our _effective_ maximum is 253, but 254 is not rejected if the last
	// character is a dot.
	l := len(s)
	if l == 0 || l > 254 || l == 254 && s[l-1] != '.' {
		return false
	}

	last := byte('.')
	nonNumeric := false // true once we've seen a letter or hyphen
	partlen := 0
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		default:
			return false
		case 'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z' || c == '_':
			nonNumeric = true
			partlen++
		case '0' <= c && c <= '9':
			// fine
			partlen++
		case c == '-':
			// Byte before dash cannot be dot.
			if last == '.' {
				return false
			}
			partlen++
			nonNumeric = true
		case c == '.':
			// Byte before dot cannot be dot, dash.
			if last == '.' || last == '-' {
				return false
			}
			if partlen > 63 || partlen == 0 {
				return false
			}
			partlen = 0
		}
		last = c
	}
	if last == '-' || partlen > 63 {
		return false
	}

	return nonNumeric
}

type QueueNode[T any] struct {
	Data T
	Prev *QueueNode[T]
	Next *QueueNode[T]
}

func NewQueueNode[T any](t T) *QueueNode[T] {
	return &QueueNode[T]{
		Data: t,
		Prev: nil,
		Next: nil,
	}
}

type Queue[T any] struct {
	front *QueueNode[T]
	rear  *QueueNode[T]
	size  int
}

func NewQueue[T any]() *Queue[T] {
	var _nil T
	head, tail := NewQueueNode[T](_nil), NewQueueNode[T](_nil)
	head.Next = tail
	tail.Prev = head
	return &Queue[T]{
		front: head,
		rear:  tail,
		size:  0,
	}
}

func (q *Queue[T]) dump() {
	node := q.front.Next
	for node != q.rear {
		fmt.Printf("%v\n", node.Data)
		node = node.Next
	}
	fmt.Println("--------------------------------")
}

func (q *Queue[T]) Size() int {
	return q.size
}

func (q *Queue[T]) Empty() bool {
	return q.size == 0
}

func (q *Queue[T]) FrontNode() *QueueNode[T] {
	if q.Empty() {
		return nil
	}
	return q.front.Next
}

func (q *Queue[T]) RearNode() *QueueNode[T] {
	if q.Empty() {
		return nil
	}
	return q.rear.Prev
}

func (q *Queue[T]) Enqueue(node *QueueNode[T]) *QueueNode[T] {
	node.Next = q.front.Next
	node.Prev = q.front
	q.front.Next.Prev = node
	q.front.Next = node
	q.size++
	return node
}

func (q *Queue[T]) Dequeue() *QueueNode[T] {
	if q.size == 0 {
		return nil
	}
	node := q.rear.Prev
	q.Remove(node)
	return node
}

func (q *Queue[T]) Remove(node *QueueNode[T]) {
	if node == nil {
		return
	}
	node.Prev.Next = node.Next
	node.Next.Prev = node.Prev
	q.size--
}

func (q *Queue[T]) MoveToFront(node *QueueNode[T]) {
	if q.Size() > 1 {
		q.Remove(node)
		q.Enqueue(node)
	}
}

type cdnResolved struct {
	domain  string
	cname   string
	cdnName string
	created time.Time
}

func newCDNResolved(domain, cname, cdnName string, created time.Time) *cdnResolved {
	return &cdnResolved{
		domain:  domain,
		cname:   cname,
		cdnName: cdnName,
		created: created,
	}
}

type lruCDNCache struct {
	cdnMap  map[string]*QueueNode[*cdnResolved]
	queue   *Queue[*cdnResolved]
	maxSize int
}

func newLruCDNCache(maxCapacity int) *lruCDNCache {
	return &lruCDNCache{
		cdnMap:  make(map[string]*QueueNode[*cdnResolved], CacheInitCap),
		queue:   NewQueue[*cdnResolved](),
		maxSize: maxCapacity,
	}
}

func (lru *lruCDNCache) push(cdn *cdnResolved) {
	if lru.queue.Size() >= lru.maxSize {
		oldest := lru.queue.Dequeue()
		if oldest != nil {
			delete(lru.cdnMap, oldest.Data.domain)

			log.Warnf("Reach to max limit of cache，the oldest data is dropped，domain = [%s], created = [%s]",
				oldest.Data.domain, oldest.Data.created.Format(time.RFC3339))
		}
	}

	node := NewQueueNode(cdn)
	lru.queue.Enqueue(node)
	lru.cdnMap[cdn.domain] = node
}

func (lru *lruCDNCache) get(domain string) *QueueNode[*cdnResolved] {
	if node, ok := lru.cdnMap[domain]; ok {
		return node
	}
	return nil
}

func (lru *lruCDNCache) moveToFront(node *QueueNode[*cdnResolved]) {
	lru.queue.MoveToFront(node)
}

func (lru *lruCDNCache) drop(domain string) *cdnResolved { //nolint: unused
	if cdn, ok := lru.cdnMap[domain]; ok {
		delete(lru.cdnMap, domain)
		lru.queue.Remove(cdn)

		if len(lru.cdnMap) != lru.queue.Size() {
			log.Warnf("cache map size do not equals queue size, map size = [%d], queue size = [%d]",
				len(lru.cdnMap), lru.queue.Size())
		}
		return cdn.Data
	}
	return nil
}
