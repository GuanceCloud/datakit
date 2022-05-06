package gitlab

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	tu "gitlab.jiagouyun.com/cloudcare-tools/cliutils/testutil"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
)

const (
	pipelineJson1 = `
{
  "object_kind": "pipeline",
  "object_attributes":{
    "id": 31,
    "ref": "master",
    "tag": false,
    "sha": "bcbb5ec396a2c0f828686f14fac9b80b780504f2",
    "before_sha": "bcbb5ec396a2c0f828686f14fac9b80b780504f2",
    "source": "merge_request_event",
    "status": "success",
    "stages":[
      "build",
      "test",
      "deploy"
    ],
    "created_at": "2016-08-12 15:23:28 UTC",
    "finished_at": "2016-08-12 15:26:29 UTC",
    "duration": 63,
    "variables": [
      {
        "key": "NESTOR_PROD_ENVIRONMENT",
        "value": "us-west-1"
      }
    ]
  },
  "merge_request": {
    "id": 1,
    "iid": 1,
    "title": "Test",
    "source_branch": "test",
    "source_project_id": 1,
    "target_branch": "master",
    "target_project_id": 1,
    "state": "opened",
    "merge_status": "can_be_merged",
    "url": "http://192.168.64.1:3005/gitlab-org/gitlab-test/merge_requests/1"
  },
  "user":{
    "id": 1,
    "name": "Administrator",
    "username": "root",
    "avatar_url": "http://www.gravatar.com/avatar/e32bd13e2add097461cb96824b7a829c?s=80\u0026d=identicon",
    "email": "user_email@gitlab.com"
  },
  "project":{
    "id": 1,
    "name": "Gitlab Test",
    "description": "Atque in sunt eos similique dolores voluptatem.",
    "web_url": "http://192.168.64.1:3005/gitlab-org/gitlab-test",
    "avatar_url": null,
    "git_ssh_url": "git@192.168.64.1:gitlab-org/gitlab-test.git",
    "git_http_url": "http://192.168.64.1:3005/gitlab-org/gitlab-test.git",
    "namespace": "Gitlab Org",
    "visibility_level": 20,
    "path_with_namespace": "gitlab-org/gitlab-test",
    "default_branch": "master"
  },
  "commit":{
    "id": "bcbb5ec396a2c0f828686f14fac9b80b780504f2",
    "message": "test\n",
    "timestamp": "2016-08-12T17:23:21+02:00",
    "url": "http://example.com/gitlab-org/gitlab-test/commit/bcbb5ec396a2c0f828686f14fac9b80b780504f2",
    "author":{
      "name": "User",
      "email": "user@gitlab.com"
    }
  },
  "builds":[
    {
      "id": 380,
      "stage": "deploy",
      "name": "production",
      "status": "skipped",
      "created_at": "2016-08-12 15:23:28 UTC",
      "started_at": null,
      "finished_at": null,
      "when": "manual",
      "manual": true,
      "allow_failure": false,
      "user":{
        "id": 1,
        "name": "Administrator",
        "username": "root",
        "avatar_url": "http://www.gravatar.com/avatar/e32bd13e2add097461cb96824b7a829c?s=80\u0026d=identicon",
        "email": "admin@example.com"
      },
      "runner": null,
      "artifacts_file":{
        "filename": null,
        "size": null
      },
      "environment": {
        "name": "production",
        "action": "start",
        "deployment_tier": "production"
      }
    },
    {
      "id": 377,
      "stage": "test",
      "name": "test-image",
      "status": "success",
      "created_at": "2016-08-12 15:23:28 UTC",
      "started_at": "2016-08-12 15:26:12 UTC",
      "finished_at": null,
      "when": "on_success",
      "manual": false,
      "allow_failure": false,
      "user":{
        "id": 1,
        "name": "Administrator",
        "username": "root",
        "avatar_url": "http://www.gravatar.com/avatar/e32bd13e2add097461cb96824b7a829c?s=80\u0026d=identicon",
        "email": "admin@example.com"
      },
      "runner": {
        "id": 380987,
        "description": "shared-runners-manager-6.gitlab.com",
        "active": true,
        "runner_type": "instance_type",
        "is_shared": true,
        "tags": [
          "linux",
          "docker",
          "shared-runner"
        ]
      },
      "artifacts_file":{
        "filename": null,
        "size": null
      },
      "environment": null
    },
    {
      "id": 378,
      "stage": "test",
      "name": "test-build",
      "status": "success",
      "created_at": "2016-08-12 15:23:28 UTC",
      "started_at": "2016-08-12 15:26:12 UTC",
      "finished_at": "2016-08-12 15:26:29 UTC",
      "when": "on_success",
      "manual": false,
      "allow_failure": false,
      "user":{
        "id": 1,
        "name": "Administrator",
        "username": "root",
        "avatar_url": "http://www.gravatar.com/avatar/e32bd13e2add097461cb96824b7a829c?s=80\u0026d=identicon",
        "email": "admin@example.com"
      },
      "runner": {
        "id":380987,
        "description":"shared-runners-manager-6.gitlab.com",
        "active":true,
        "runner_type": "instance_type",
        "is_shared": true,
        "tags": [
          "linux",
          "docker"
        ]
      },
      "artifacts_file":{
        "filename": null,
        "size": null
      },
      "environment": null
    },
    {
      "id": 376,
      "stage": "build",
      "name": "build-image",
      "status": "success",
      "created_at": "2016-08-12 15:23:28 UTC",
      "started_at": "2016-08-12 15:24:56 UTC",
      "finished_at": "2016-08-12 15:25:26 UTC",
      "when": "on_success",
      "manual": false,
      "allow_failure": false,
      "user":{
        "id": 1,
        "name": "Administrator",
        "username": "root",
        "avatar_url": "http://www.gravatar.com/avatar/e32bd13e2add097461cb96824b7a829c?s=80\u0026d=identicon",
        "email": "admin@example.com"
      },
      "runner": {
        "id": 380987,
        "description": "shared-runners-manager-6.gitlab.com",
        "active": true,
        "runner_type": "instance_type",
        "is_shared": true,
        "tags": [
          "linux",
          "docker"
        ]
      },
      "artifacts_file":{
        "filename": null,
        "size": null
      },
      "environment": null
    },
    {
      "id": 379,
      "stage": "deploy",
      "name": "staging",
      "status": "created",
      "created_at": "2016-08-12 15:23:28 UTC",
      "started_at": null,
      "finished_at": null,
      "when": "on_success",
      "manual": false,
      "allow_failure": false,
      "user":{
        "id": 1,
        "name": "Administrator",
        "username": "root",
        "avatar_url": "http://www.gravatar.com/avatar/e32bd13e2add097461cb96824b7a829c?s=80\u0026d=identicon",
        "email": "admin@example.com"
      },
      "runner": null,
      "artifacts_file":{
        "filename": null,
        "size": null
      },
      "environment": {
        "name": "staging",
        "action": "start",
        "deployment_tier": "staging"
      }
    }
  ]
}`
	pipelineJson2 = `
{
  "object_kind": "pipeline",
  "object_attributes": {
    "id": 110437,
    "ref": "iss-449",
    "tag": false,
    "sha": "1cabf86e6a687e684926538f9e605454e0b64fab",
    "before_sha": "803160bd046069fbcdd11a30c5c64f3ac8ab031b",
    "source": "push",
    "status": "failed",
    "detailed_status": "failed",
    "stages": [
      "deploy"
    ],
    "created_at": "2022-04-01 13:49:55 +0800",
    "finished_at": "2022-04-01 15:00:27 +0800",
    "duration": null,
    "queued_duration": null,
    "variables": [

    ]
  },
  "merge_request": null,
  "user": {
    "id": 397,
    "name": "pengyonghui",
    "username": "pengyonghui",
    "avatar_url": "http://gitlab.jiagouyun.com/uploads/-/system/user/avatar/397/avatar.png",
    "email": "421640644@qq.com"
  },
  "project": {
    "id": 800,
    "name": "datakit",
    "description": "",
    "web_url": "http://gitlab.jiagouyun.com/pengyonghui/datakit",
    "avatar_url": null,
    "git_ssh_url": "ssh://git@gitlab.jiagouyun.com:40022/pengyonghui/datakit.git",
    "git_http_url": "http://gitlab.jiagouyun.com/pengyonghui/datakit.git",
    "namespace": "pengyonghui",
    "visibility_level": 0,
    "path_with_namespace": "pengyonghui/datakit",
    "default_branch": "dev",
    "ci_config_path": null
  },
  "commit": {
    "id": "1cabf86e6a687e684926538f9e605454e0b64fab",
    "message": "Update test-ci.txt",
    "title": "Update test-ci.txt",
    "timestamp": "2022-04-01T13:49:54+08:00",
    "url": "http://gitlab.jiagouyun.com/pengyonghui/datakit/-/commit/1cabf86e6a687e684926538f9e605454e0b64fab",
    "author": {
      "name": "pengyonghui",
      "email": "397-pengyonghui@users.noreply.gitlab.jiagouyun.com"
    }
  },
  "builds": [
    {
      "id": 120257,
      "stage": "deploy",
      "name": "release-nothing-only-build",
      "status": "failed",
      "created_at": "2022-04-01 13:49:55 +0800",
      "started_at": null,
      "finished_at": "2022-04-01 15:00:27 +0800",
      "duration": null,
      "queued_duration": 96317.11524972,
      "when": "on_success",
      "manual": false,
      "allow_failure": false,
      "user": {
        "id": 397,
        "name": "pengyonghui",
        "username": "pengyonghui",
        "avatar_url": "http://gitlab.jiagouyun.com/uploads/-/system/user/avatar/397/avatar.png",
        "email": "421640644@qq.com"
      },
      "runner": null,
      "artifacts_file": {
        "filename": null,
        "size": null
      },
      "environment": null
    }
  ]
}`
	jobJson = `
{
  "object_kind": "build",
  "ref": "gitlab-script-trigger",
  "tag": false,
  "before_sha": "2293ada6b400935a1378653304eaf6221e0fdb8f",
  "sha": "2293ada6b400935a1378653304eaf6221e0fdb8f",
  "build_id": 1977,
  "build_name": "test",
  "build_stage": "test",
  "build_status": "success",
  "build_created_at": "2021-02-23T02:41:37.886Z",
  "build_started_at": "2021-02-23T02:41:37.886Z",
  "build_finished_at": null,
  "build_duration": null,
  "build_allow_failure": false,
  "build_failure_reason": "script_failure",
  "pipeline_id": 2366,
  "project_id": 380,
  "project_name": "gitlab-org/gitlab-test",
  "user": {
    "id": 3,
    "name": "User",
    "email": "user@gitlab.com",
    "avatar_url": "http://www.gravatar.com/avatar/e32bd13e2add097461cb96824b7a829c?s=80\u0026d=identicon"
  },
  "commit": {
    "id": 2366,
    "sha": "2293ada6b400935a1378653304eaf6221e0fdb8f",
    "message": "test\n",
    "author_name": "User",
    "author_email": "user@gitlab.com",
    "status": "created",
    "duration": null,
    "started_at": null,
    "finished_at": null
  },
  "repository": {
    "name": "gitlab_test",
    "description": "Atque in sunt eos similique dolores voluptatem.",
    "homepage": "http://192.168.64.1:3005/gitlab-org/gitlab-test",
    "git_ssh_url": "git@192.168.64.1:gitlab-org/gitlab-test.git",
    "git_http_url": "http://192.168.64.1:3005/gitlab-org/gitlab-test.git",
    "visibility_level": 20
  },
  "runner": {
    "active": true,
    "runner_type": "project_type",
    "is_shared": false,
    "id": 380987,
    "description": "shared-runners-manager-6.gitlab.com",
    "tags": [
      "linux",
      "docker"
    ]
  },
  "environment": null
}`
	failedJobJson = `{
  "object_kind": "build",
  "ref": "testing-ci4",
  "tag": false,
  "before_sha": "ef286a974090ddb2c6b399d13c806562e8fbc334",
  "sha": "d062ab827c6c0dfe1164a6c6f02128778e93e592",
  "build_id": 121672,
  "build_name": "release-testing",
  "build_stage": "deploy",
  "build_status": "failed",
  "build_created_at": "2022-04-17 10:12:05 +0800",
  "build_started_at": "2022-04-17 10:12:15 +0800",
  "build_finished_at": "2022-04-17 10:12:32 +0800",
  "build_duration": 16.791868,
  "build_queued_duration": 10.100971,
  "build_allow_failure": false,
  "build_failure_reason": "script_failure",
  "pipeline_id": 111536,
  "runner": {
    "id": 33,
    "description": "ubuntu-3",
    "runner_type": "project_type",
    "active": true,
    "is_shared": false,
    "tags": [
      "cloudcare-ft"
    ]
  },
  "project_id": 806,
  "project_name": "pengyonghui / datakit",
  "user": {
    "id": 397,
    "name": "pengyonghui",
    "username": "pengyonghui",
    "avatar_url": "http://gitlab.jiagouyun.com/uploads/-/system/user/avatar/397/avatar.png",
    "email": "421640644@qq.com"
  },
  "commit": {
    "id": 111536,
    "sha": "d062ab827c6c0dfe1164a6c6f02128778e93e592",
    "message": "Update .gitlab-ci.yml",
    "author_name": "pengyonghui",
    "author_email": "397-pengyonghui@users.noreply.gitlab.jiagouyun.com",
    "author_url": "http://gitlab.jiagouyun.com/pengyonghui",
    "status": "failed",
    "duration": 16,
    "started_at": "2022-04-16 22:20:43 +0800",
    "finished_at": "2022-04-17 10:12:32 +0800"
  },
  "repository": {
    "name": "datakit",
    "url": "ssh://git@gitlab.jiagouyun.com:40022/pengyonghui/datakit.git",
    "description": "",
    "homepage": "http://gitlab.jiagouyun.com/pengyonghui/datakit",
    "git_http_url": "http://gitlab.jiagouyun.com/pengyonghui/datakit.git",
    "git_ssh_url": "ssh://git@gitlab.jiagouyun.com:40022/pengyonghui/datakit.git",
    "visibility_level": 0
  },
  "environment": null
}`
	unwantedJobJson = `
{
  "object_kind": "build",
  "ref": "gitlab-script-trigger",
  "tag": false,
  "before_sha": "2293ada6b400935a1378653304eaf6221e0fdb8f",
  "sha": "2293ada6b400935a1378653304eaf6221e0fdb8f",
  "build_id": 1977,
  "build_name": "test",
  "build_stage": "test",
  "build_status": "unwanted",
  "build_created_at": "2021-02-23T02:41:37.886Z",
  "build_started_at": "2021-02-23T02:41:37.886Z",
  "build_finished_at": null,
  "build_duration": null,
  "build_allow_failure": false,
  "build_failure_reason": "script_failure",
  "pipeline_id": 2366,
  "project_id": 380,
  "project_name": "gitlab-org/gitlab-test",
  "user": {
    "id": 3,
    "name": "User",
    "email": "user@gitlab.com",
    "avatar_url": "http://www.gravatar.com/avatar/e32bd13e2add097461cb96824b7a829c?s=80\u0026d=identicon"
  },
  "commit": {
    "id": 2366,
    "sha": "2293ada6b400935a1378653304eaf6221e0fdb8f",
    "message": "test\n",
    "author_name": "User",
    "author_email": "user@gitlab.com",
    "status": "created",
    "duration": null,
    "started_at": null,
    "finished_at": null
  },
  "repository": {
    "name": "gitlab_test",
    "description": "Atque in sunt eos similique dolores voluptatem.",
    "homepage": "http://192.168.64.1:3005/gitlab-org/gitlab-test",
    "git_ssh_url": "git@192.168.64.1:gitlab-org/gitlab-test.git",
    "git_http_url": "http://192.168.64.1:3005/gitlab-org/gitlab-test.git",
    "visibility_level": 20
  },
  "runner": {
    "active": true,
    "runner_type": "project_type",
    "is_shared": false,
    "id": 380987,
    "description": "shared-runners-manager-6.gitlab.com",
    "tags": [
      "linux",
      "docker"
    ]
  },
  "environment": null
}`
)

type mockWriter struct {
	statusCode int
}

func (m *mockWriter) Header() http.Header {
	return nil
}

func (m *mockWriter) Write(i []byte) (int, error) {
	return 0, nil
}

func (m *mockWriter) WriteHeader(statusCode int) {
	m.statusCode = statusCode
}

func newMockWriter() *mockWriter {
	return &mockWriter{statusCode: 200}
}

func TestPipelineJson(t *testing.T) {
	var ppl PipelineEventPayload
	if err := json.Unmarshal([]byte(pipelineJson1), &ppl); err != nil {
		t.Error(err)
	}
	if err := json.Unmarshal([]byte(pipelineJson2), &ppl); err != nil {
		t.Error(err)
	}
}

func TestJobJson(t *testing.T) {
	var h JobEventPayload
	if err := json.Unmarshal([]byte(jobJson), &h); err != nil {
		t.Error(err)
	}
}

func TestGetPipelineTagsAndFields(t *testing.T) {
	testCases := []struct {
		name           string
		eventJson      string
		expectedTags   map[string]string
		expectedFields map[string]interface{}
	}{
		{
			"success pipeline",
			pipelineJson1,
			map[string]string{
				"ci_status":       "success",
				"pipeline_name":   "gitlab-org/gitlab-test",
				"author_email":    "user@gitlab.com",
				"pipeline_source": "merge_request_event",
				"operation_name":  "pipeline",
				"resource":        "Gitlab Test",
				"object_kind":     "pipeline",
				"pipeline_url":    "http://192.168.64.1:3005/gitlab-org/gitlab-test/pipelines/31",
				"commit_sha":      "bcbb5ec396a2c0f828686f14fac9b80b780504f2",
				"repository_url":  "http://192.168.64.1:3005/gitlab-org/gitlab-test.git",
				"ref":             "master",
			},
			map[string]interface{}{
				"duration":       int64(63),
				"pipeline_id":    "31",
				"commit_message": "test\n",
				"message":        "test\n",
				"created_at":     int64(1471015408),
				"finished_at":    int64(1471015589),
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var pipeline PipelineEventPayload
			if err := json.Unmarshal([]byte(tc.eventJson), &pipeline); err != nil {
				t.Error(err)
			}
			tags := getPipelineEventTags(pipeline)
			fields := getPipelineEventFields(pipeline)
			tu.Equals(t, tc.expectedTags, tags)
			tu.Equals(t, tc.expectedFields, fields)
		})
	}
}

func TestGetJobTagsAndFields(t *testing.T) {
	testCases := []struct {
		name           string
		eventJson      string
		expectedTags   map[string]string
		expectedFields map[string]interface{}
	}{
		{
			"success job",
			jobJson,
			map[string]string{
				"object_kind":          "build",
				"build_status":         "success",
				"project_name":         "gitlab-org/gitlab-test",
				"user_email":           "user@gitlab.com",
				"build_repo_name":      "gitlab_test",
				"sha":                  "2293ada6b400935a1378653304eaf6221e0fdb8f",
				"build_name":           "test",
				"build_stage":          "test",
				"build_failure_reason": "script_failure",
				"build_commit_sha":     "2293ada6b400935a1378653304eaf6221e0fdb8f",
			},
			map[string]interface{}{
				"runner_id":            "380987",
				"build_id":             "1977",
				"pipeline_id":          "2366",
				"project_id":           "380",
				"build_started_at":     int64(1614048097),
				"build_commit_message": "test\n",
				"message":              "test\n",
				"build_duration":       0.0,
			},
		},
		{
			"failed job",
			failedJobJson,
			map[string]string{
				"object_kind":          "build",
				"build_status":         "failed",
				"project_name":         "pengyonghui / datakit",
				"user_email":           "421640644@qq.com",
				"build_repo_name":      "datakit",
				"sha":                  "d062ab827c6c0dfe1164a6c6f02128778e93e592",
				"build_name":           "release-testing",
				"build_stage":          "deploy",
				"build_failure_reason": "script_failure",
				"build_commit_sha":     "d062ab827c6c0dfe1164a6c6f02128778e93e592",
			},
			map[string]interface{}{
				"runner_id":            "33",
				"build_id":             "121672",
				"pipeline_id":          "111536",
				"project_id":           "806",
				"build_started_at":     int64(1650161535),
				"build_commit_message": "Update .gitlab-ci.yml",
				"message":              "Update .gitlab-ci.yml",
				"build_duration":       16.791868,
				"build_finished_at":    int64(1650161552),
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var job JobEventPayload
			if err := json.Unmarshal([]byte(tc.eventJson), &job); err != nil {
				t.Error(err)
			}
			tags := getJobEventTags(job)
			fields := getJobEventFields(job)
			tu.Equals(t, tc.expectedTags, tags)
			tu.Equals(t, tc.expectedFields, fields)
		})
	}
}

func TestServeHTTP(t *testing.T) {
	ipt := getInput(30 * time.Second)
	table := []struct {
		name               string
		request            *http.Request
		expectedStatusCode int
	}{
		{
			request:            getPipelineRequest(pipelineJson1),
			expectedStatusCode: http.StatusOK,
		},
		{
			request:            getJobRequest(jobJson),
			expectedStatusCode: http.StatusOK,
		},
	}
	for _, tc := range table {
		t.Run(tc.name, func(t *testing.T) {
			w := newMockWriter()
			ipt.ServeHTTP(w, tc.request)
			assert.Equal(t, tc.expectedStatusCode, w.statusCode)
		})
	}
}

func TestRemove(t *testing.T) {
	ipt := getInput(1 * time.Second)
	r := getPipelineRequest(pipelineJson1)
	ipt.ServeHTTP(newMockWriter(), r)
	digest := md5.Sum([]byte(pipelineJson1))
	assert.True(t, ipt.reqMemo.has(digest))
	time.Sleep(2 * time.Second)
	assert.False(t, ipt.reqMemo.has(md5.Sum([]byte(pipelineJson1))))
}

func TestFailToFeed(t *testing.T) {
	ipt := getInput(30 * time.Second)
	ipt.feed = func(name, category string, pts []*io.Point, opt *io.Option) error {
		return fmt.Errorf("mock error")
	}
	r := getPipelineRequest(pipelineJson1)
	digest := md5.Sum([]byte(pipelineJson1))
	w := newMockWriter()
	ipt.ServeHTTP(w, r)
	assert.Equal(t, http.StatusInternalServerError, w.statusCode)
	assert.False(t, ipt.reqMemo.has(digest))
}

func TestUnwantedEvent(t *testing.T) {
	ipt := getInput(30 * time.Second)
	r := getJobRequest(unwantedJobJson)
	digest := md5.Sum([]byte(unwantedJobJson))
	w := newMockWriter()
	ipt.ServeHTTP(w, r)
	assert.Equal(t, http.StatusOK, w.statusCode)
	assert.True(t, ipt.reqMemo.has(digest))
}

func TestAddExtraTags(t *testing.T) {
	testCases := []struct {
		name     string
		existing map[string]string
		extra    map[string]string
		expected map[string]string
	}{
		{
			"add extra tags",
			map[string]string{"a": "b"},
			map[string]string{"c": "d"},
			map[string]string{"a": "b", "c": "d"},
		},
		{
			"do not overwrite existing tags",
			map[string]string{"a": "b"},
			map[string]string{"a": "c"},
			map[string]string{"a": "b"},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ipt := newInput()
			ipt.CIExtraTags = tc.extra
			tags := tc.existing
			ipt.addExtraTags(tags)
			assert.Equal(t, tc.expected, tags)
		})
	}
}

func getInput(expired time.Duration) *Input {
	ipt := newInput()
	ipt.feed = func(name, category string, pts []*io.Point, opt *io.Option) error {
		return nil
	}
	ipt.feedLastError = func(inputName string, err string) {}
	go ipt.reqMemo.memoMaintainer(expired)
	return ipt
}

func getPipelineRequest(reqBody string) *http.Request {
	r := httptest.NewRequest("POST", "/", bytes.NewReader([]byte(reqBody)))
	r.Header.Set(gitlabEventHeader, pipelineHook)
	return r
}

func getJobRequest(reqBody string) *http.Request {
	r := httptest.NewRequest("POST", "/", bytes.NewReader([]byte(reqBody)))
	r.Header.Set(gitlabEventHeader, jobHook)
	return r
}
