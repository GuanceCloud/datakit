package container

import (
	"testing"
)

func TestParseImage(t *testing.T) {
	cases := []struct {
		image          string
		imageName      string
		imageShortName string
		imageTag       string
	}{
		{
			"influxdb:1.7.9",
			"influxdb",
			"influxdb",
			"1.7.9",
		},
		{
			"gitlab/gitlab-ee:latest",
			"gitlab/gitlab-ee",
			"gitlab-ee",
			"latest",
		},
		{
			"registry.aliyuncs.com/google_containers/pause:3.1",
			"registry.aliyuncs.com/google_containers/pause",
			"pause",
			"3.1",
		},
		{
			"testing-none",
			"testing-none",
			"testing-none",
			"unknown",
		},
	}

	for idx, tc := range cases {
		name, short, version := ParseImage(tc.image)
		if name != tc.imageName {
			t.Errorf("[%d][ERROR] should be %s, got %s\n", idx, tc.imageName, name)
		}
		if short != tc.imageShortName {
			t.Errorf("[%d][ERROR] should be %s, got %s\n", idx, tc.imageShortName, short)
		}
		if version != tc.imageTag {
			t.Errorf("[%d][ERROR] should be %s, got %s\n", idx, tc.imageTag, version)
		}

		t.Logf("[%d][OK] %v\n", idx, tc)
	}
}

func TestRegexpMatchString(t *testing.T) {
	cases := []struct {
		regexps []string
		target  string
		fail    bool
	}{
		{
			[]string{`hello-*`, `hello1-*`, `hello2-*`},
			"hello2-image",
			false,
		},
		{
			[]string{`registry*`, `registry.aliyuncs*`, `registry.aliyuncs.com*`},
			"registry.aliyuncs.com/google_containers",
			false,
		},
		{
			[]string{`hello-*`, `hello1-*`, `hello2-*`},
			"registry.aliyuncs.com/google_containers/kube-apiserver",
			true,
		},
	}

	for idx, tc := range cases {
		matchRes := regexpMatchString(tc.regexps, tc.target)
		switch {
		case matchRes && !tc.fail:
			t.Logf("[ OK ] index: %d \n", idx)
		case matchRes:
			t.Logf("[ ERROR ] index: %d\n", idx)
		default:
			t.Logf("[ OK ] index: %d\n", idx)
		}
	}
}
