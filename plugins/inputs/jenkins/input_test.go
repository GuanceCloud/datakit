// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package jenkins

import (
	"bytes"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	pipelineEventTrace = `[
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
	jobEventTraceSuccess = `[
    [
        {
            "trace_id": 6894633842882167706,
            "span_id": 622433946346465891,
            "parent_id": 8583937955644618326,
            "name": "jenkins.step",
            "resource": "Shell Script",
            "service": "jenkins-instance",
            "type": "ci",
            "meta": {
                "ci.job.number": "33",
                "ci.status": "success",
                "ci.job.name": "Shell Script",
                "git.commit_sha": "6eecc513e6435d2c654c94f9c86f76250f53402e",
                "ci.workspace_path": "/var/lib/jenkins/jobs/test-ci/branches/master/workspace",
                "user.name": "root",
                "ci.job.script": "go build",
                "language": "",
                "_dd.ci.level": "job",
                "git.branch": "master",
                "error": "false",
                "jenkins.step.args.script": "go build",
                "git.repository_url": "https://gitee.com/ethan97/leetcode.git",
                "jenkins.result": "success",
                "ci.job.url": "http://192.168.47.4:8080/job/test-ci/job/master/5/execution/node/33/",
                "git.commit.sha": "6eecc513e6435d2c654c94f9c86f76250f53402e",
                "ci.job.result": "success",
                "_dd.origin": "ciapp-pipeline",
                "ci.stage.name": "build",
                "ci.node.name": "master",
                "ci.node.labels": "[\"built-in\"]",
                "ci.pipeline.id": "jenkins-test-ci-master-5",
                "_dd.ci.internal": "false",
                "ci.provider.name": "jenkins",
                "_dd.ci.build_level": "job",
                "ci.pipeline.name": "test-ci"
            },
            "metrics": {
                "_sampling_priority_v1": 1,
                "ci.queue_time": 0
            },
            "start": 1652020914276000000,
            "duration": 821000000
        }
    ]
]`
	jobEventTraceFailed = `[
	[
		{
            "trace_id": 6894633842882167706,
            "span_id": 7957055028247124595,
            "parent_id": 8583937955644618326,
            "error": 1,
            "name": "jenkins.step",
            "resource": "Shell Script",
            "service": "jenkins-instance",
            "type": "ci",
            "meta": {
                "ci.job.number": "34",
                "ci.status": "error",
                "ci.job.name": "Shell Script",
                "git.commit_sha": "6eecc513e6435d2c654c94f9c86f76250f53402e",
                "ci.workspace_path": "/var/lib/jenkins/jobs/test-ci/branches/master/workspace",
                "user.name": "root",
                "ci.job.script": "chmod 777 what",
                "language": "",
                "_dd.ci.level": "job",
                "git.branch": "master",
                "error": "true",
                "jenkins.step.args.script": "chmod 777 what",
                "git.repository_url": "https://gitee.com/ethan97/leetcode.git",
                "jenkins.result": "error",
                "ci.job.url": "http://192.168.47.4:8080/job/test-ci/job/master/5/execution/node/34/",
                "git.commit.sha": "6eecc513e6435d2c654c94f9c86f76250f53402e",
                "ci.job.result": "error",
                "_dd.origin": "ciapp-pipeline",
                "ci.stage.name": "build",
                "ci.node.name": "master",
                "ci.node.labels": "[\"built-in\"]",
                "ci.pipeline.id": "jenkins-test-ci-master-5",
                "_dd.ci.internal": "false",
                "error.type": "hudson.AbortException",
                "ci.provider.name": "jenkins",
                "_dd.ci.build_level": "job",
                "error.message": "script returned exit code 1",
                "error.stack": "hudson.AbortException: script returned exit code 1\n\tat org.jenkinsci.plugins.workflow.steps.durable_task.DurableTaskStep$Execution.handleExit(DurableTaskStep.java:664)\n\tat org.jenkinsci.plugins.workflow.steps.durable_task.DurableTaskStep$Execution.check(DurableTaskStep.java:610)\n\tat org.jenkinsci.plugins.workflow.steps.durable_task.DurableTaskStep$Execution.run(DurableTaskStep.java:554)\n\tat java.util.concurrent.Executors$RunnableAdapter.call(Executors.java:511)\n\tat java.util.concurrent.FutureTask.run(FutureTask.java:266)\n\tat java.util.concurrent.ScheduledThreadPoolExecutor$ScheduledFutureTask.access$201(ScheduledThreadPoolExecutor.java:180)\n\tat java.util.concurrent.ScheduledThreadPoolExecutor$ScheduledFutureTask.run(ScheduledThreadPoolExecutor.java:293)\n\tat java.util.concurrent.ThreadPoolExecutor.runWorker(ThreadPoolExecutor.java:1149)\n\tat java.util.concurrent.ThreadPoolExecutor$Worker.run(ThreadPoolExecutor.java:624)\n\tat java.lang.Thread.run(Thread.java:748)\n",
                "ci.pipeline.name": "test-ci"
            },
            "metrics": {
                "_sampling_priority_v1": 1,
                "ci.queue_time": 0
            },
            "start": 1652020915097000000,
            "duration": 279000000
        }
	]
]`
	stageEventTrace = `[
	[
		{
            "trace_id": 6894633842882167706,
            "span_id": 8583937955644618326,
            "parent_id": 1713178964574047858,
            "error": 1,
            "name": "jenkins.stage",
            "resource": "build",
            "service": "jenkins-instance",
            "type": "ci",
            "meta": {
                "ci.status": "error",
                "git.commit_sha": "6eecc513e6435d2c654c94f9c86f76250f53402e",
                "ci.workspace_path": "/var/lib/jenkins/jobs/test-ci/branches/master/workspace",
                "user.name": "root",
                "language": "",
                "_dd.ci.level": "stage",
                "git.branch": "master",
                "error": "true",
                "git.repository_url": "https://gitee.com/ethan97/leetcode.git",
                "jenkins.result": "success",
                "git.commit.sha": "6eecc513e6435d2c654c94f9c86f76250f53402e",
                "_dd.origin": "ciapp-pipeline",
                "ci.stage.url": "http://192.168.47.4:8080/job/test-ci/job/master/5/execution/node/32/",
                "ci.stage.name": "build",
                "ci.node.name": "master",
                "ci.node.labels": "[\"built-in\"]",
                "ci.pipeline.id": "jenkins-test-ci-master-5",
                "_dd.ci.internal": "false",
                "error.type": "hudson.AbortException",
                "ci.provider.name": "jenkins",
                "ci.stage.number": "32",
                "_dd.ci.build_level": "stage",
                "error.message": "script returned exit code 1",
                "error.stack": "hudson.AbortException: script returned exit code 1\n\tat org.jenkinsci.plugins.workflow.steps.durable_task.DurableTaskStep$Execution.handleExit(DurableTaskStep.java:664)\n\tat org.jenkinsci.plugins.workflow.steps.durable_task.DurableTaskStep$Execution.check(DurableTaskStep.java:610)\n\tat org.jenkinsci.plugins.workflow.steps.durable_task.DurableTaskStep$Execution.run(DurableTaskStep.java:554)\n\tat java.util.concurrent.Executors$RunnableAdapter.call(Executors.java:511)\n\tat java.util.concurrent.FutureTask.run(FutureTask.java:266)\n\tat java.util.concurrent.ScheduledThreadPoolExecutor$ScheduledFutureTask.access$201(ScheduledThreadPoolExecutor.java:180)\n\tat java.util.concurrent.ScheduledThreadPoolExecutor$ScheduledFutureTask.run(ScheduledThreadPoolExecutor.java:293)\n\tat java.util.concurrent.ThreadPoolExecutor.runWorker(ThreadPoolExecutor.java:1149)\n\tat java.util.concurrent.ThreadPoolExecutor$Worker.run(ThreadPoolExecutor.java:624)\n\tat java.lang.Thread.run(Thread.java:748)\n",
                "ci.pipeline.name": "test-ci",
                "ci.stage.result": "error"
            },
            "metrics": {
                "_sampling_priority_v1": 1,
                "ci.queue_time": 0
            },
            "start": 1652020914233000000,
            "duration": 1143000000
        }
	]
]`
)

func TestDecodeTraces(t *testing.T) {
	reader := bytes.NewReader([]byte(pipelineEventTrace))
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
	testCases := []struct {
		name           string
		in             string
		failed         bool
		expectedTags   map[string]string
		expectedFields map[string]interface{}
	}{
		{
			"pipeline event",
			pipelineEventTrace,
			false,
			map[string]string{
				"pipeline_url":   "http://192.168.47.4:8080/job/test-ci/job/master/5/",
				"ref":            "master",
				"repository_url": "https://gitee.com/ethan97/leetcode.git",
				"commit_sha":     "6eecc513e6435d2c654c94f9c86f76250f53402e",
				"object_kind":    "pipeline",
				"operation_name": "pipeline",
				"pipeline_name":  "test-ci",
				"resource":       "leetcode",
				"author_email":   "421640644@qq.com",
				"ci_status":      "failed",
			},
			map[string]interface{}{
				"pipeline_id":    "jenkins-test-ci-master-5",
				"commit_message": "modify\n",
				"created_at":     int64(1652020909),
				"duration":       int64(7),
				"finished_at":    int64(1652020916),
				"message":        "jenkins-test-ci-master-5",
			},
		},
		{
			"job event success",
			jobEventTraceSuccess,
			false,
			map[string]string{
				"build_commit_sha": "6eecc513e6435d2c654c94f9c86f76250f53402e",
				"build_name":       "Shell Script",
				"build_repo_name":  "https://gitee.com/ethan97/leetcode.git",
				"build_stage":      "build",
				"build_status":     "success",
				"object_kind":      "job",
				"project_name":     "leetcode",
				"sha":              "6eecc513e6435d2c654c94f9c86f76250f53402e",
			},
			map[string]interface{}{
				"pipeline_id":       "jenkins-test-ci-master-5",
				"runner_id":         "master",
				"build_duration":    int64(0),
				"build_finished_at": int64(1652020915),
				"build_id":          "33",
				"build_started_at":  int64(1652020914),
				"message":           "Shell Script",
			},
		},
		{
			"job event failed",
			jobEventTraceFailed,
			false,
			map[string]string{
				"build_status":         "failed",
				"object_kind":          "job",
				"project_name":         "leetcode",
				"build_commit_sha":     "6eecc513e6435d2c654c94f9c86f76250f53402e",
				"build_name":           "Shell Script",
				"build_stage":          "build",
				"build_failure_reason": "script returned exit code 1",
				"build_repo_name":      "https://gitee.com/ethan97/leetcode.git",
				"sha":                  "6eecc513e6435d2c654c94f9c86f76250f53402e",
			},
			map[string]interface{}{
				"message":           "Shell Script",
				"pipeline_id":       "jenkins-test-ci-master-5",
				"runner_id":         "master",
				"build_duration":    int64(0),
				"build_finished_at": int64(1652020915),
				"build_id":          "34",
				"build_started_at":  int64(1652020915),
			},
		},
		{
			"stage event should ignore",
			stageEventTrace,
			true,
			map[string]string{},
			map[string]interface{}{},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			i := Input{}
			traces, err := getTraces(tc.in)
			assert.Nil(t, err)
			for _, trace := range traces {
				for _, span := range trace {
					p, err := i.getPoint(span)
					if tc.failed {
						assert.Nil(t, p)
						return
					}
					assert.Nil(t, err)
					assert.Equal(t, tc.expectedTags, p.Tags())
					fields, err := p.Fields()
					assert.Nil(t, err)
					assert.Equal(t, tc.expectedFields, fields)
				}
			}
		})
	}
}

func TestCIExtraTags(t *testing.T) {
	testCases := []struct {
		name              string
		in                string
		failed            bool
		expectedExtraTags map[string]string
	}{
		{
			"extra tags in pipeline event",
			pipelineEventTrace,
			false,
			map[string]string{
				"custom_key":  "custom_value",
				"another_key": "another_value",
			},
		},
		{
			"extra tags in job event",
			jobEventTraceSuccess,
			false,
			map[string]string{
				"hello": "hi",
				"what":  "sup",
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			i := Input{CIExtraTags: tc.expectedExtraTags}
			traces, err := getTraces(tc.in)
			assert.Nil(t, err)
			for _, trace := range traces {
				for _, span := range trace {
					p, err := i.getPoint(span)
					assert.Nil(t, err)
					for k, v := range tc.expectedExtraTags {
						assert.Contains(t, p.Tags(), k)
						assert.True(t, p.Tags()[k] == v)
					}
				}
			}
		})
	}
}

func getTraces(eventJson string) (ddTraces, error) {
	reader := bytes.NewReader([]byte(eventJson))
	req := httptest.NewRequest("POST", "/v0.3/traces", reader)
	return decodeTraces(req)
}
