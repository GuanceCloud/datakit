package gitlab

import (
	"encoding/json"
	"testing"

	tu "gitlab.jiagouyun.com/cloudcare-tools/cliutils/testutil"
)

var pipelineJson1 = `
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

var pipelineJson2 = `
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

var jobJson = `
{
  "object_kind": "build",
  "ref": "gitlab-script-trigger",
  "tag": false,
  "before_sha": "2293ada6b400935a1378653304eaf6221e0fdb8f",
  "sha": "2293ada6b400935a1378653304eaf6221e0fdb8f",
  "build_id": 1977,
  "build_name": "test",
  "build_stage": "test",
  "build_status": "created",
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

func TestPipelineTagsAndFields(t *testing.T) {
	var ppl PipelineEventPayload
	if err := json.Unmarshal([]byte(pipelineJson1), &ppl); err != nil {
		t.Error(err)
	}
	tags := getPipelineEventTags(ppl)
	fields := getPipelineEventFields(ppl)
	expectedFields := map[string]interface{}{
		"duration":       int64(63),
		"pipeline_id":    int64(31),
		"commit_message": "test\n",
	}
	expectedTags := map[string]string{
		"ci_status":      "success",
		"pipeline_name":  "gitlab-org/gitlab-test",
		"author_email":   "user@gitlab.com",
		"source":         "merge_request_event",
		"operation_name": "pipeline",
		"resource":       "Gitlab Test",
		"object_kind":    "pipeline",
		"pipeline_url":   "http://192.168.64.1:3005/gitlab-org/gitlab-test/pipelines/31",
		"commit_sha":     "bcbb5ec396a2c0f828686f14fac9b80b780504f2",
		"repository_url": "http://192.168.64.1:3005/gitlab-org/gitlab-test.git",
		"ref":            "master",
	}
	tu.Equals(t, expectedTags, tags)
	tu.Equals(t, expectedFields, fields)
}

func TestJobTagsAndFields(t *testing.T) {
	var job JobEventPayload
	if err := json.Unmarshal([]byte(jobJson), &job); err != nil {
		t.Error(err)
	}
	tags := getJobEventTags(job)
	fields := getJobEventFields(job)
	expectedTags := map[string]string{
		"object_kind":          "build",
		"build_status":         "created",
		"project_name":         "gitlab-org/gitlab-test",
		"user_email":           "user@gitlab.com",
		"build_repo_name":      "gitlab_test",
		"sha":                  "2293ada6b400935a1378653304eaf6221e0fdb8f",
		"build_name":           "test",
		"build_stage":          "test",
		"build_failure_reason": "script_failure",
		"build_commit_sha":     "2293ada6b400935a1378653304eaf6221e0fdb8f",
	}
	expectedFields := map[string]interface{}{
		"runner_id":            int64(380987),
		"build_id":             int64(1977),
		"pipeline_id":          int64(2366),
		"project_id":           int64(380),
		"build_started_at":     int64(1614048097),
		"build_commit_message": "test\n",
	}
	tu.Equals(t, expectedTags, tags)
	tu.Equals(t, expectedFields, fields)
}
