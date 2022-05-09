package jenkins

import (
	"bytes"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

const tracesJson = `[
    [
        {
            "trace_id": 6894633842882167706,
            "span_id": 1713178964574047858,
            "error": 1,
            "name": "jenkins.build",
            "resource": "test-ci",
            "service": "jenkins-instance",
            "type": "ci",
            "meta": {
                "ci.status": "error",
                "git.commit_sha": "6eecc513e6435d2c654c94f9c86f76250f53402e",
                "user.name": "root",
                "ci.workspace_path": "/var/lib/jenkins/jobs/test-ci/branches/master/workspace",
                "_dd.ci.level": "pipeline",
                "git.branch": "master",
                "ci.pipeline.result": "error",
                "git.repository_url": "https://gitee.com/ethan97/leetcode.git",
                "git.commit.message": "modify\n",
                "user.email": "421640644@qq.com",
                "git.commit.author.name": "pengyonghui",
                "git.commit.committer.date": "2022-05-08T14:41:32.000Z",
                "git.commit.author.email": "421640644@qq.com",
                "jenkins.result": "failure",
                "git.commit.committer.email": "421640644@qq.com",
                "jenkins.tag": "jenkins-test-ci-master-5",
                "git.commit.sha": "6eecc513e6435d2c654c94f9c86f76250f53402e",
                "_dd.ci.stages": "[{\"name\":\"Declarative: Checkout SCM\",\"duration\":599000000},{\"name\":\"prepare\",\"duration\":317000000},{\"name\":\"test\",\"duration\":1133000000},{\"name\":\"build\",\"duration\":1143000000},{\"name\":\"deploy\",\"duration\":15000000}]",
                "_dd.origin": "ciapp-pipeline",
                "ci.pipeline.url": "http://192.168.47.4:8080/job/test-ci/job/master/5/",
                "ci.node.name": "master",
                "jenkins.executor.number": "",
                "ci.node.labels": "[\"built-in\"]",
                "ci.pipeline.number": "5",
                "ci.pipeline.id": "jenkins-test-ci-master-5",
                "_dd.ci.internal": "false",
                "ci.provider.name": "jenkins",
                "_dd.ci.build_level": "pipeline",
                "git.commit.committer.name": "pengyonghui",
                "git.commit.author.date": "2022-05-08T14:41:32.000Z",
                "ci.pipeline.name": "test-ci"
            },
            "metrics": {
                "_sampling_priority_v1": 1,
                "ci.queue_time": 0
            },
            "start": 1652020909552000000,
            "duration": 7319000000
        }
    ]
]`

func TestDecodeTraces(t *testing.T) {
	reader := bytes.NewReader([]byte(tracesJson))
	req := httptest.NewRequest("POST", "/v0.3/traces", reader)
	_, err := decodeTraces(req)
	assert.Nil(t, err)
}

func TestExtractProjectName(t *testing.T) {
	testCases := []struct {
		name   string
		in     string
		expect string
	}{
		{
			"normal",
			"https://github.com/kevindurant/leetcode.git",
			"leetcode",
		},
		{
			".git missing",
			"https://github.com/kevindurant/leetcode",
			"",
		},
		{
			"random url",
			"https://baidu.com",
			"",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expect, extractProjectName(tc.in))
		})
	}
}

func TestGetPoint(t *testing.T) {
	traces, err := getTraces()
	assert.Nil(t, err)
	i := Input{CIExtraTags: map[string]string{"some_key": "some_value"}}
	for _, trace := range traces {
		for _, span := range trace {
			p, err := i.getPoint(span)
			assert.Nil(t, err)
			assert.Contains(t, p.Tags(), "some_key")
		}
	}
}

func getTraces() (ddTraces, error) {
	reader := bytes.NewReader([]byte(tracesJson))
	req := httptest.NewRequest("POST", "/v0.3/traces", reader)
	return decodeTraces(req)
}
