// Package version implements datakit's version parsing/compare
package version

import (
	"fmt"
	"strconv"
	"strings"
)

type VerInfo struct {
	VersionString string `json:"version"`
	Commit        string `json:"commit"`
	ReleaseDate   string `json:"date_utc"`

	DownloadURL        string `json:"-"`
	DownloadURLTesting string `json:"-"`

	tag   string
	major uint64
	minor uint64
	min   uint64
	rc    string
	build uint64
}

func (vi *VerInfo) Compare(i *VerInfo) int {
	if vi == nil {
		return 1
	}

	a := vi.major*1024*1024*1024 + vi.minor*1024*1024 + vi.min
	b := i.major*1024*1024*1024 + i.minor*1024*1024 + i.min

	if a > b {
		return 1
	}

	// same version number: 1.1.7
	n := strings.Compare(vi.rc, i.rc)
	if n > 0 {
		return 1
	} else if n < 0 {
		return -1
	}

	if vi.build > i.build {
		return 1
	} else if vi.build < i.build {
		return -1
	}

	return 0 // same version
}

func (vi *VerInfo) String() string {
	return fmt.Sprintf("datakit %s/%s", vi.VersionString, vi.Commit)
}

func (vi *VerInfo) parseNumbers(s string) error {
	arr := strings.Split(s, ".")
	if len(arr) != 3 { //nolint:gomnd
		return fmt.Errorf("invalid number: %s", vi.VersionString)
	}

	var err error

	vi.major, err = strconv.ParseUint(arr[0], 10, 64) //nolint:gomnd
	if err != nil {
		return fmt.Errorf("invalid number: %s", arr[0])
	}

	vi.minor, err = strconv.ParseUint(arr[1], 10, 64) //nolint:gomnd
	if err != nil {
		return fmt.Errorf("invalid number: %s", arr[1])
	}

	if vi.minor > 1024 { //nolint:gomnd
		return fmt.Errorf("too large minor number: %d (should <=1024)", vi.minor)
	}

	vi.min, err = strconv.ParseUint(arr[2], 10, 64) //nolint:gomnd
	if err != nil {
		return fmt.Errorf("invalid number: %s", arr[2])
	}

	if vi.min > 1024 { //nolint:gomnd
		return fmt.Errorf("too large min number: %d (should <=1024)", vi.min)
	}
	return nil
}

//nolint:gomnd
func (vi *VerInfo) Parse() error {
	arr := strings.Split(vi.VersionString, "_") // there should be no `_' within version string
	if len(arr) >= 2 {
		vi.tag = arr[1]
	}

	// older version has prefix `v', this crash semver.Parse()
	verstr := strings.TrimPrefix(arr[0], "v")

	parts := strings.Split(verstr, "-")

	if err := vi.parseNumbers(parts[0]); err != nil {
		return err
	}

	switch len(parts) {
	case 1: // like 1.1.7
		return nil

	case 2: // like 1.1.7-rc2
		if err := vi.parseNumbers(parts[0]); err != nil {
			return err
		}
		if strings.HasPrefix(parts[1], "rc") {
			vi.rc = parts[1]
		}
		return nil

	case 3: // like 1.2.0-123-g40c4860c
		if err := vi.parseNumbers(parts[0]); err != nil {
			return err
		}

		var err error
		vi.build, err = strconv.ParseUint(parts[1], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid build number %s: %w", parts[2], err)
		}

		vi.Commit = parts[2]

		return nil

	case 4: // like 1.1.7-rc1-125-g40c4860c
		if err := vi.parseNumbers(parts[0]); err != nil {
			return err
		}

		if strings.HasPrefix(parts[1], "rc") {
			vi.rc = parts[1]
		}

		var err error

		vi.build, err = strconv.ParseUint(parts[2], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid build number %s: %w", parts[2], err)
		}
		vi.Commit = parts[3]

		return nil
	}

	return fmt.Errorf("invalid version string %s", vi.VersionString)
}

func IsNewVersion(newVer, curver *VerInfo, acceptRC bool) bool {
	if newVer.Compare(curver) > 0 { // new version
		if newVer.rc == "" { // no rc version
			return true
		}

		if acceptRC {
			return true
		}
	}

	return false
}

func IsValidReleaseVersion(releaseVer string) bool {
	ver := &VerInfo{VersionString: releaseVer}
	err := ver.Parse()
	if err == nil && ver.build == 0 { // new version
		return true
	}

	return false
}

func (vi *VerInfo) GetMinor() uint64 {
	return vi.minor
}

func (vi *VerInfo) GetMajor() uint64 {
	return vi.major
}

func (vi *VerInfo) GetMin() uint64 {
	return vi.min
}

func (vi *VerInfo) IsStable() bool {
	return vi.minor%2 == 0
}
