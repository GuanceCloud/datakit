// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package gitlab

import (
	T "testing"

	"github.com/stretchr/testify/assert"
)

func TestPipelineEvent(t *T.T) {
	t.Run(`pl-queue-duration`, func(t *T.T) {
		ipt := defaultInput()
		for _, v := range plevents {
			pts, err := ipt.getPoint([]byte(v), pipelineHook)
			assert.NoError(t, err)
			assert.Len(t, pts, 1)

			t.Logf("point: %s", pts[0].Pretty())

			assert.Equal(t, int64(118000), pts[0].Get("queued_duration").(int64))
			assert.Equal(t, v, pts[0].Get("event_raw").(string))
		}
	})
}

func TestJobEvent(t *T.T) {
	t.Run(`basic`, func(t *T.T) {
		ipt := defaultInput()
		for _, v := range jobevents {
			pts, err := ipt.getPoint([]byte(v), jobHook)
			assert.NoError(t, err)
			assert.Len(t, pts, 1)

			t.Logf("point: %s", pts[0].Pretty())

			assert.Equal(t, int64(1095588715), pts[0].Get("queued_duration").(int64))
			assert.Equal(t, v, pts[0].Get("event_raw").(string))
		}
	})
}

var (
	jobevents = map[string]string{
		"16.0": `{
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
  "build_started_at": "2021-02-23T02:41:38.886Z",
  "build_finished_at": "2021-02-23T02:41:39.886Z",
  "build_duration": 123.4,
  "build_queued_duration": 1095.588715,
  "build_allow_failure": false,
  "build_failure_reason": "script_failure",
  "retries_count": 2,
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
    "name": "Build pipeline",
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
}`,
	}

	plevents = map[string]string{
		"16.0": `{
   "object_kind": "pipeline",
   "object_attributes":{
      "id": 31,
      "iid": 3,
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
      "detailed_merge_status": "mergeable",
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
      "message": "test",
      "timestamp": "2016-08-12T17:23:21+02:00",
      "url": "http://example.com/gitlab-org/gitlab-test/commit/bcbb5ec396a2c0f828686f14fac9b80b780504f2",
      "author":{
         "name": "User",
         "email": "user@gitlab.com"
      }
   },
   "source_pipeline":{
      "project":{
        "id": 41,
        "web_url": "https://gitlab.example.com/gitlab-org/upstream-project",
        "path_with_namespace": "gitlab-org/upstream-project"
      },
      "pipeline_id": 30,
      "job_id": 3401
   },
   "builds":[]
}`,
	}
)
