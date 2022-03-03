package worker

import (
	"os"
	"testing"
)

func TestCall(t *testing.T) {
	LoadDefaultDotPScript2Store()
	ReloadAllDefaultDotPScript2Store()
	LoadGitReposDotPScript2Store(nil)
	ReloadAllGitReposDotPScript2Store(nil)
	LoadRemoteDotPScript2Store(nil)
	ReloadAllRemoteDotPScript2Store(nil)
	_ = os.WriteFile("/tmp/nginx-time123.p", []byte(`
	json(_, time)
	set_tag(bb, "aa0")
	default_time(time)
	`), os.FileMode(0o755))
	loadDotPScript2StoreWithNS("xxxx", []string{"/tmp/nginx-time.p123"}, "")
	_ = os.Remove("/tmp/nginx-time123.p")
	loadDotPScript2StoreWithNS("xxx", nil, "")
}
