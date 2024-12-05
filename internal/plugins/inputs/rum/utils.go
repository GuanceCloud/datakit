// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package rum

import (
	"archive/zip"
	"bufio"
	"errors"
	"fmt"
	"io"
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

	"github.com/go-sourcemap/sourcemap"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/path"
)

const iosDSYMFilePath = "Contents/Resources/DWARF"

const CacheInitCap = 16

const TmpExpiredDirExt = ".expired-tmp"

const (
	SdkWeb        = "df_web_rum_sdk"
	SdkWebMiniApp = "df_miniapp_rum_sdk"
	SdkWebUniApp  = "df_uniapp_rum_sdk"
	SdkAndroid    = "df_android_rum_sdk"
	SdkIOS        = "df_ios_rum_sdk"
)

const (
	JavaCrash   = "java_crash"
	NativeCrash = "native_crash"
	IOSCrash    = "ios_crash"
)

type webSourceMapDict map[string]map[string]*sourcemap.Consumer

var (
	ClientRealIPHeaders = []string{"X-Forwarded-For", "X-Real-IP"}

	NDKAvailableABI = map[string]struct{}{
		"armeabi-v7a": {},
		"arm64-v8a":   {},
		"x86":         {},
		"x86_64":      {},
	}
	srcMapDirs = map[string]string{
		SdkWeb:        SourceMapDirWeb,
		SdkWebMiniApp: SourceMapDirMini,
		SdkWebUniApp:  SourceMapDirMini,
		SdkAndroid:    SourceMapDirAndroid,
		SdkIOS:        SourceMapDirIOS,
	}

	rumMetricAppID         = "app_id"
	sourceMapTokenBuckets  = newExecCmdTokenBuckets(200)
	webSourcemapCache      = make(webSourceMapDict)
	webSourceCacheLoadTime = make(map[string]time.Time)
	webSourcemapLock       sync.RWMutex
	ExtractZipLock         sync.Mutex

	// IOSAddressRegexp for match
	// 4   App                                         0x0000000104fd0728 0x104f30000 + 657192
	//
	// $1 "App"
	// $2 "0x0000000104fd0728"
	// $3 "0x104f30000 + 657192"  // 用于解析后对原始堆栈进行替换
	// $4 0x104f30000
	// $5 "657192".
	IOSAddressRegexp = regexp.MustCompile(`(\S+)\s+(0x[0-9a-fA-F]+)\s+((0x[0-9a-fA-F]+)\s*\+\s*(\d+|0x[0-9a-fA-F]+))`)
	replaceRegexp    = regexp.MustCompile(`@ .*:\d+:\d+`)
)

var errLimitReader = errors.New("limit reader err")

type limitReader struct {
	r io.ReadCloser
}

func newLimitReader(r io.ReadCloser, max int64) io.ReadCloser {
	return &limitReader{
		r: http.MaxBytesReader(nil, r, max),
	}
}

func (l *limitReader) Read(p []byte) (int, error) {
	n, err := l.r.Read(p)
	if err != nil {
		if err == io.EOF { //nolint:errorlint
			return n, err
		}
		// wrap the errLimitReader
		return n, fmt.Errorf("%w: %s", errLimitReader, err)
	}
	return n, nil
}

func (l *limitReader) Close() error {
	return l.r.Close()
}

func (ipt *Input) getWebSourceMapDirs() map[string]struct{} {
	sourceMapDirs := make(map[string]struct{}, 2)

	for _, sdkName := range []string{SdkWeb, SdkWebMiniApp, SdkWebUniApp} {
		sourceMapDirs[ipt.getRumSourcemapDir(sdkName)] = struct{}{}
	}
	return sourceMapDirs
}

func (ipt *Input) loadSourcemapFile() error {
	sourceMapDirs := ipt.getWebSourceMapDirs()

	webSourcemapLock.Lock()
	defer webSourcemapLock.Unlock()

	for dir := range sourceMapDirs {
		archives, err := scanArchives(dir)
		for _, archive := range archives {
			sourcemapItem, err := loadZipFile(archive.Filepath, MaxSourceMapFileSize, ".map")
			if err != nil {
				log.Warnf("load zip file %s failed, %s", archive.Filepath, err.Error())
				continue
			}

			filename := filepath.Base(archive.Filepath)

			webSourcemapCache[filename] = sourcemapItem
			webSourceCacheLoadTime[filename] = time.Now()
		}
		if err != nil {
			log.Warnf("scan source map dir [%s] encounter error: %s", dir, err)
		}
	}

	loadedZipGauge.WithLabelValues(SourceMapDirWeb).Set(float64(len(webSourcemapCache)))
	return nil
}

func (ipt *Input) getRumSourcemapDir(sdkName string) string {
	dir, ok := srcMapDirs[sdkName]
	if !ok {
		dir = SourceMapDirWeb
	}

	return filepath.Join(ipt.rumDataDir, dir)
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

func isPrivateIP(ip net.IP) bool {
	if ip.IsLoopback() {
		return true
	}

	if dotIP := ip.To4(); dotIP != nil {
		switch {
		case dotIP[0] == 10:
			return true
		case dotIP[0] == 172 && dotIP[1] >= 16 && dotIP[1] <= 31:
			return true
		case dotIP[0] == 192 && dotIP[1] == 168:
			return true
		}
	}

	return false
}

type iosCrashAddress struct {
	start     string
	end       string
	originStr string
}

func scanIOSCrashAddress(originErrStack string) (map[string]map[string][]iosCrashAddress, error) {
	matches := IOSAddressRegexp.FindAllStringSubmatch(originErrStack, -1)

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
			return AndroidCommandLineTools, nil
		}
	}
	return Proguard, nil
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

func getSourcemap(errStack string, sourcemapItem map[string]*sourcemap.Consumer, status *sourceMapStatus) string {
	replaceStr := replaceRegexp.ReplaceAllStringFunc(errStack, func(str string) string {
		return str[0:2] + getSourceMapString(str[2:], sourcemapItem, status)
	})
	return replaceStr
}

func getSourceMapString(str string, sourcemapItem map[string]*sourcemap.Consumer, status *sourceMapStatus) string {
	var err error
	defer func() {
		if err != nil && status != nil {
			status.status = StatusError
			status.reason = err.Error()
		}
	}()

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
		err = fmt.Errorf("parse url failed, %s, %w", srcPath, err)
		log.Warn(err.Error())
		return str
	}

	urlPath := strings.TrimPrefix(urlObj.Path, "/")
	sourceMapFileName := urlPath + ".map"

	smap, ok := sourcemapItem[sourceMapFileName]
	if !ok {
		err = fmt.Errorf("parse sourcemap for [%s] failed: sourcemap file [%s] is required", str, sourceMapFileName)
		log.Warn(err.Error())
		return str
	}

	file, _, line, col, ok := smap.Source(rowNumber, colNumber)

	if ok {
		return fmt.Sprintf("%s:%v:%v", file, line, col)
	} else {
		err = fmt.Errorf("fetch original source information failed, make sure sourcemap file [%s] is valid", sourceMapFileName)
		log.Warn(err.Error())
	}

	return str
}

func copyZipItem(item *zip.File, dst string) error {
	reader, err := item.Open()
	if err != nil {
		return fmt.Errorf("open zip item fail: %w", err)
	}

	defer reader.Close() //nolint:errcheck

	dstDir := filepath.Dir(dst)
	if !path.IsDir(dstDir) { // path not exist
		if err := os.MkdirAll(dstDir, 0o750); err != nil {
			return fmt.Errorf("mkdir %q fail: %w", dstDir, err)
		}
	}

	writer, err := os.Create(dst) // nolint:gosec
	if err != nil {
		return fmt.Errorf("create dst file fail: %w", err)
	}

	defer writer.Close() //nolint:errcheck,gosec

	log.Infof("copy zip file to %q", dst)

	_, err = io.Copy(writer, reader) // nolint:gosec
	if err != nil {
		return fmt.Errorf("copy zip item fail: %w", err)
	}

	return nil
}

func extractZipFile(zipFileAbsPath string) error {
	zipFileDir := filepath.Dir(zipFileAbsPath)
	extractTo := strings.TrimSuffix(filepath.Base(zipFileAbsPath), ZipExt)
	tmpExtractTo := "." + extractTo

	reader, err := zip.OpenReader(zipFileAbsPath)
	if err != nil {
		return fmt.Errorf("open zip file [%s] fail: %w", zipFileAbsPath, err)
	}

	defer reader.Close() // nolint:errcheck

	for _, item := range reader.File {
		fname := checkZipEntry(item)
		if fname == "" {
			continue
		}

		var absItemPath string
		if !strings.HasPrefix(filepath.Clean(fname), extractTo) {
			absItemPath = filepath.Join(zipFileDir, tmpExtractTo, fname) // nolint:gosec
		} else {
			itemName := strings.Replace(fname, extractTo, tmpExtractTo, 1)
			absItemPath = filepath.Join(zipFileDir, itemName) // nolint:gosec
		}

		if item.FileInfo().IsDir() {
			if err := os.MkdirAll(absItemPath, 0o750); err != nil {
				return fmt.Errorf("can not mkdir: %w", err)
			} else {
				continue
			}
		} else if err := copyZipItem(item, absItemPath); err != nil {
			return err
		}
	}

	absExtractTo := filepath.Join(zipFileDir, extractTo)
	absTmpExtractTo := filepath.Join(zipFileDir, tmpExtractTo)
	expiredDirName := ""
	if isDir(absExtractTo) {
		expiredDirName = filepath.Join(zipFileDir, extractTo+"."+strconv.FormatInt(time.Now().UnixNano(), 32)+TmpExpiredDirExt)
		if err := os.Rename(absExtractTo, expiredDirName); err != nil {
			return fmt.Errorf("unable to rename %s to %s: %w", absExtractTo, expiredDirName, err)
		}
	}
	if err := os.Rename(absTmpExtractTo, absExtractTo); err != nil {
		return fmt.Errorf("unable to rename %s to %s: %w", tmpExtractTo, extractTo, err)
	}

	// todo: remove in another goroutine
	if expiredDirName != "" {
		if err := os.RemoveAll(expiredDirName); err != nil {
			log.Warnf("unable to remove dir [%s]: %s", expiredDirName, err)
		}
	}

	return nil
}

// checkSourcemapZipEntry check if filename in zip are valid.
//
// Only *.map file are checked, subdir in zip ignored.
func checkSourcemapZipEntry(f *zip.File, maxSize uint64, suffix string) string {
	if f.FileInfo().IsDir() {
		log.Infof("ignore zip dir %q", f.Name)
		return ""
	}

	if !strings.HasSuffix(f.Name, suffix) {
		log.Infof("ignore zip file %q, not suffix with %q", f.Name, suffix)
		return ""
	}

	// TODO: is dir in zip got UncompressedSize64?
	if f.UncompressedSize64 > maxSize {
		log.Warnf("ignore zip file %q with size %d(>%d)", f.Name, f.UncompressedSize64, MaxSourceMapFileSize)
		return ""
	}

	fname := checkZipEntry(f)
	if fname == "" {
		return ""
	}

	return fname
}

// checkZipEntry check if filename(include subdir) in zip are valid.
//
// We check for file name start with ../ or abstract path, these files
// may cause security issue and is not valid sourcemap upload.
func checkZipEntry(f *zip.File) string {
	fname := filepath.Clean(f.Name)

	if strings.HasPrefix(fname, "..") {
		log.Warnf("ignore zip file with dir inject %q(%q)", fname, f.Name)
		return ""
	}

	if filepath.IsAbs(fname) {
		log.Warnf("ignore zip file with abs path %q(%q)", fname, f.Name)
		return ""
	}

	return fname
}

func readZipSourcemap(f *zip.File, maxSize uint64, suffix string) (string, *sourcemap.Consumer, error) {
	fname := checkSourcemapZipEntry(f, maxSize, suffix)
	if fname == "" {
		return "", nil, nil
	}

	file, err := f.Open()
	if err != nil {
		return "", nil, err
	}
	defer file.Close() // nolint:errcheck

	content, err := io.ReadAll(file)
	if err != nil {
		return "", nil, err
	}

	smap, err := sourcemap.Parse(fname, content)
	if err != nil {
		return "", nil, err
	}

	return fname, smap, nil
}

func loadZipFile(zipFile string, maxZipFileSize uint64, suffix string) (map[string]*sourcemap.Consumer, error) {
	sourcemapItem := make(map[string]*sourcemap.Consumer)

	zipReader, err := zip.OpenReader(zipFile)
	if err != nil {
		return nil, fmt.Errorf("zip.OpenReader(%q): %w", zipFile, err)
	}
	defer zipReader.Close() //nolint:errcheck

	for _, f := range zipReader.File {
		fname, smap, err := readZipSourcemap(f, maxZipFileSize, suffix)
		if err != nil {
			log.Errorf("parse sourcemap %q failed: %s", f.Name, err.Error())
			continue
		}

		if smap != nil && fname != "" {
			log.Infof("load sourcemap file %q", fname)
			sourcemapItem[fname] = smap
		}
	}

	return sourcemapItem, nil
}

const (
	MaxSourceMapFileSize = uint64(4 * (1 << 30))
)

func updateSourcemapCache(zipFile string) error {
	fileName := filepath.Base(zipFile)
	if !strings.HasSuffix(fileName, ZipExt) {
		return fmt.Errorf(`suffix name is not ".zip" [%s]`, zipFile)
	}

	log.Infof("reload source map archive: %s", zipFile)
	sourcemapItem, err := loadZipFile(zipFile, MaxSourceMapFileSize, ".map")
	if err != nil {
		log.Errorf("load zip file error: %s", err.Error())
		return fmt.Errorf("load zip file [%s] err: %w", zipFile, err)
	}

	webSourcemapLock.Lock()
	defer webSourcemapLock.Unlock()

	if _, ok := webSourcemapCache[fileName]; !ok {
		loadedZipGauge.WithLabelValues(SourceMapDirWeb).Inc()
	}

	webSourcemapCache[fileName] = sourcemapItem
	webSourceCacheLoadTime[fileName] = time.Now()

	log.Infof("load sourcemap success: %s", fileName)

	return nil
}

func deleteSourcemapCache(zipFiles ...string) {
	if len(zipFiles) == 0 {
		return
	}
	webSourcemapLock.Lock()
	defer webSourcemapLock.Unlock()

	cnt := 0
	for _, zipFile := range zipFiles {
		fileName := filepath.Base(zipFile)
		if strings.HasSuffix(fileName, ".zip") {
			delete(webSourcemapCache, fileName)
			delete(webSourceCacheLoadTime, fileName)
			log.Infof("web zip archive [%s] removed from cache", fileName)
			cnt++
		}
	}
	if cnt > 0 {
		loadedZipGauge.WithLabelValues(SourceMapDirWeb).Sub(float64(cnt))
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

type Queue[T any] struct {
	front *QueueNode[T]
	rear  *QueueNode[T]
	size  int
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
