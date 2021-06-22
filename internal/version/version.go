package version

import (
	"fmt"
	"strconv"
	"strings"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
)

var (
	l = logger.DefaultSLogger("version")
)

type VerInfo struct {
	VersionString string `json:"version"`
	Commit        string `json:"commit"`
	ReleaseDate   string `json:"date_utc"`

	DownloadURL        string `json:"-"`
	DownloadURLTesting string `json:"-"`

	major uint64
	minor uint64
	min   uint64
	rc    string
	build uint64
}

func (x *VerInfo) Compare(y *VerInfo) int {
	if x == nil {
		return 1
	}

	a := x.major*1024*1024*1024 + x.minor*1024*1024 + x.min
	b := y.major*1024*1024*1024 + y.minor*1024*1024 + y.min

	l.Debugf("v1: %d, v2: %d", a, b)

	if a > b {
		return 1
	} else if a < 0 {
		return -1
	}

	// same version number: 1.1.7
	n := strings.Compare(x.rc, y.rc)
	if n > 0 {
		return 1
	} else if n < 0 {
		return -1
	}

	if x.build > y.build {
		return 1
	} else if x.build < y.build {
		return -1
	}

	return 0 // same version
}

func (vi *VerInfo) String() string {
	return fmt.Sprintf("datakit %s/%s", vi.VersionString, vi.Commit)
}

func (vi *VerInfo) parseNumbers(s string) error {
	arr := strings.Split(s, ".")
	if len(arr) != 3 {
		return fmt.Errorf("invalid number: %s", vi.VersionString)
	}

	var err error

	vi.major, err = strconv.ParseUint(arr[0], 10, 64)
	if err != nil {
		return fmt.Errorf("invalid number: %s", arr[0])
	}

	vi.minor, err = strconv.ParseUint(arr[1], 10, 64)
	if err != nil {
		return fmt.Errorf("invalid number: %s", arr[1])
	}

	if vi.minor > 1024 {
		return fmt.Errorf("too large minor number: %d (should <=1024)", vi.minor)
	}

	vi.min, err = strconv.ParseUint(arr[2], 10, 64)
	if err != nil {
		return fmt.Errorf("invalid number: %s", arr[2])
	}

	if vi.min > 1024 {
		return fmt.Errorf("too large min number: %d (should <=1024)", vi.min)
	}
	return nil
}

func (vi *VerInfo) Parse() error {
	verstr := strings.TrimPrefix(vi.VersionString, "v") // older version has prefix `v', this crash semver.Parse()

	parts := strings.Split(verstr, "-")
	switch len(parts) {
	case 2: // like 1.1.7-rc2
		if err := vi.parseNumbers(parts[0]); err != nil {
			return err
		}
		if strings.HasPrefix(parts[1], "rc") {
			vi.rc = parts[1]
		}
		return nil
	case 4: //like 1.1.7-rc1-125-g40c4860c
		if err := vi.parseNumbers(parts[0]); err != nil {
			return err
		}

		if strings.HasPrefix(parts[1], "rc") {
			vi.rc = parts[1]
		}

		var err error

		vi.build, err = strconv.ParseUint(parts[2], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid build number %s: %s", parts[2], err.Error())
		}

		return nil
	}

	return fmt.Errorf("unknown version string %s, expect format is `1.1.7-rc2' or `1.1.7-rc1-125-g40c4860c'", verstr)
}

func IsNewVersion(newVer, curver *VerInfo, acceptRC bool) bool {

	if newVer.Compare(curver) > 0 { // new version
		if len(newVer.rc) == 0 { // no rc version
			return true
		}

		if acceptRC {
			return true
		}
	}

	return false
}
