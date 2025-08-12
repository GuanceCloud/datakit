// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package gitlab

import (

	// nolint:gosec
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/point"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/metrics"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/ntp"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type customTime struct {
	time.Time
}

func (t *customTime) UnmarshalJSON(b []byte) (err error) {
	layout := []string{
		"2006-01-02 15:04:05 MST",
		"2006-01-02 15:04:05 Z07:00",
		"2006-01-02 15:04:05 Z0700",
		time.RFC3339,
	}
	s := strings.Trim(string(b), "\"")
	if s == "null" {
		t.Time = time.Time{}
		return
	}
	for _, l := range layout {
		t.Time, err = time.Parse(l, s)
		if err == nil {
			break
		}
	}
	return
}

// plEventPayload contains the information for GitLab's pipeline status change event.
type plEventPayload struct {
	ObjectKind       *string       `json:"object_kind"`
	User             *user         `json:"user"`
	Project          *project      `json:"project"`
	Commit           *commit       `json:"commit"`
	ObjectAttributes *plObjAttrs   `json:"object_attributes"`
	MergeRequest     *mergeRequest `json:"merge_request"`
	Builds           []build       `json:"builds"`
}

// jobEventPayload contains the information for GitLab's Job status change.
type jobEventPayload struct {
	ObjectKind  *string `json:"object_kind"`
	Ref         *string `json:"ref"`
	Tag         *bool   `json:"tag"`
	BeforeSHA   *string `json:"before_sha"`
	SHA         *string `json:"sha"`
	BuildID     *int64  `json:"build_id"`
	BuildName   *string `json:"build_name"`
	BuildStage  *string `json:"build_stage"`
	BuildStatus *string `json:"build_status"`

	BuildCreatedAt  *customTime `json:"build_created_at"`
	BuildStartedAt  *customTime `json:"build_started_at"`
	BuildFinishedAt *customTime `json:"build_finished_at"`

	BuildQueuedDuration *float64 `json:"build_queued_duration"`
	BuildDuration       *float64 `json:"build_duration"`

	BuildAllowFailure  *bool        `json:"build_allow_failure"`
	BuildFailureReason *string      `json:"build_failure_reason"`
	PipelineID         *int64       `json:"pipeline_id"`
	ProjectID          *int64       `json:"project_id"`
	ProjectName        *string      `json:"project_name"`
	User               *user        `json:"user"`
	Commit             *buildCommit `json:"commit"`
	Repository         *repository  `json:"repository"`
	Runner             *runner      `json:"runner"`
}

// build contains all the GitLab build information.
type build struct {
	ID            *int64         `json:"id"`
	Stage         *string        `json:"stage"`
	Name          *string        `json:"name"`
	Status        *string        `json:"status"`
	CreatedAt     *customTime    `json:"created_at"`
	StartedAt     *customTime    `json:"started_at"`
	FinishedAt    *customTime    `json:"finished_at"`
	When          *string        `json:"when"`
	Manual        *bool          `json:"manual"`
	User          *user          `json:"user"`
	Runner        *runner        `json:"runner"`
	ArtifactsFile *artifactsFile `json:"artifactsfile"`
}

// runner represents a runner agent.
type runner struct {
	ID          *int64  `json:"id"`
	Description *string `json:"description"`
	Active      *bool   `json:"active"`
	IsShared    *bool   `json:"is_shared"`
}

// artifactsFile contains all the GitLab artifact information.
type artifactsFile struct {
	Filename *string `json:"filename"`
	Size     *string `json:"size"`
}

// commit contains all the GitLab commit information.
type commit struct {
	ID        *string     `json:"id"`
	Message   *string     `json:"message"`
	Title     *string     `json:"title"`
	Timestamp *customTime `json:"timestamp"`
	URL       *string     `json:"url"`
	Author    *author     `json:"author"`
	Added     *[]string   `json:"added"`
	Modified  *[]string   `json:"modified"`
	Removed   *[]string   `json:"removed"`
}

// buildCommit contains all the GitLab build commit information.
type buildCommit struct {
	ID          *int64      `json:"id"`
	SHA         *string     `json:"sha"`
	Message     *string     `json:"message"`
	AuthorName  *string     `json:"author_name"`
	AuthorEmail *string     `json:"author_email"`
	Status      *string     `json:"status"`
	Duration    *float64    `json:"duration"`
	StartedAt   *customTime `json:"started_at"`
	FinishedAt  *customTime `json:"finished_at"`
}

// user contains all the GitLab user information.
type user struct {
	ID        *int64  `json:"id"`
	Name      *string `json:"name"`
	UserName  *string `json:"username"`
	AvatarURL *string `json:"avatar_url"`
	Email     *string `json:"email"`
}

// project contains all the GitLab project information.
type project struct {
	ID                *int64  `json:"id"`
	Name              *string `json:"name"`
	Description       *string `json:"description"`
	WebURL            *string `json:"web_url"`
	AvatarURL         *string `json:"avatar_url"`
	GitSSHURL         *string `json:"git_ssh_url"`
	GitHTTPURL        *string `json:"git_http_url"`
	Namespace         *string `json:"namespace"`
	VisibilityLevel   *int64  `json:"visibility_level"`
	PathWithNamespace *string `json:"path_with_namespace"`
	DefaultBranch     *string `json:"default_branch"`
	Homepage          *string `json:"homepage"`
	URL               *string `json:"url"`
	SSHURL            *string `json:"ssh_url"`
	HTTPURL           *string `json:"http_url"`
}

// repository contains all the GitLab repository information.
type repository struct {
	Name            *string `json:"name"`
	URL             *string `json:"url"`
	Description     *string `json:"description"`
	Homepage        *string `json:"homepage"`
	GitSSHURL       *string `json:"git_ssh_url"`
	GitHTTPURL      *string `json:"git_http_url"`
	VisibilityLevel *int64  `json:"visibility_level"`
}

// plObjAttrs contains pipeline specific GitLab object attributes information.
type plObjAttrs struct {
	ID         *int64         `json:"id"`
	Ref        *string        `json:"ref"`
	Tag        *bool          `json:"tag"`
	SHA        *string        `json:"sha"`
	BeforeSHA  *string        `json:"before_sha"`
	Source     *string        `json:"source"`
	Status     *string        `json:"status"`
	Stages     *[]string      `json:"stages"`
	CreatedAt  *customTime    `json:"created_at"`
	FinishedAt *customTime    `json:"finished_at"`
	Duration   *int64         `json:"duration"`
	Variables  *[]plVariables `json:"variables"`
}

// plVariables contains pipeline variables.
type plVariables struct {
	Key   *string `json:"key"`
	Value *string `json:"value"`
}

// mergeRequest contains all the GitLab merge request information.
type mergeRequest struct {
	ID              *int64      `json:"id"`
	TargetBranch    *string     `json:"target_branch"`
	SourceBranch    *string     `json:"source_branch"`
	SourceProjectID *int64      `json:"source_project_id"`
	AssigneeID      *int64      `json:"assignee_id"`
	AuthorID        *int64      `json:"author_id"`
	Title           *string     `json:"title"`
	CreatedAt       *customTime `json:"created_at"`
	UpdatedAt       *customTime `json:"updated_at"`
	MilestoneID     *int64      `json:"milestone_id"`
	State           *string     `json:"state"`
	MergeStatus     *string     `json:"merge_status"`
	TargetProjectID *int64      `json:"target_project_id"`
	IID             *int64      `json:"iid"`
	Description     *string     `json:"description"`
	Position        *int64      `json:"position"`
	LockedAt        *customTime `json:"locked_at"`
	Source          *source     `json:"source"`
	Target          *target     `json:"target"`
	LastCommit      *lastCommit `json:"last_commit"`
	WorkInProgress  *bool       `json:"work_in_progress"`
	Assignee        *assignee   `json:"assignee"`
	URL             *string     `json:"url"`
}

// assignee contains all the GitLab assignee information.
type assignee struct {
	ID        *int64  `json:"id"`
	Name      *string `json:"name"`
	Username  *string `json:"username"`
	AvatarURL *string `json:"avatar_url"`
	Email     *string `json:"email"`
}

// source contains all the GitLab source information.
type source struct {
	Name              *string `json:"name"`
	Description       *string `json:"description"`
	WebURL            *string `json:"web_url"`
	AvatarURL         *string `json:"avatar_url"`
	GitSSHURL         *string `json:"git_ssh_url"`
	GitHTTPURL        *string `json:"git_http_url"`
	Namespace         *string `json:"namespace"`
	VisibilityLevel   *int64  `json:"visibility_level"`
	PathWithNamespace *string `json:"path_with_namespace"`
	DefaultBranch     *string `json:"default_branch"`
	Homepage          *string `json:"homepage"`
	URL               *string `json:"url"`
	SSHURL            *string `json:"ssh_url"`
	HTTPURL           *string `json:"http_url"`
}

// target contains all the GitLab target information.
type target struct {
	Name              *string `json:"name"`
	Description       *string `json:"description"`
	WebURL            *string `json:"web_url"`
	AvatarURL         *string `json:"avatar_url"`
	GitSSHURL         *string `json:"git_ssh_url"`
	GitHTTPURL        *string `json:"git_http_url"`
	Namespace         *string `json:"namespace"`
	VisibilityLevel   *int64  `json:"visibility_level"`
	PathWithNamespace *string `json:"path_with_namespace"`
	DefaultBranch     *string `json:"default_branch"`
	Homepage          *string `json:"homepage"`
	URL               *string `json:"url"`
	SSHURL            *string `json:"ssh_url"`
	HTTPURL           *string `json:"http_url"`
}

// lastCommit contains all the GitLab last commit information.
type lastCommit struct {
	ID        *string     `json:"id"`
	Message   *string     `json:"message"`
	Timestamp *customTime `json:"timestamp"`
	URL       *string     `json:"url"`
	Author    *author     `json:"author"`
}

// author contains all the GitLab author information.
type author struct {
	Name  *string `json:"name"`
	Email *string `json:"email"`
}

func getJobEventFields(j jobEventPayload) map[string]interface{} {
	var (
		fields = map[string]interface{}{}
		createdAt,
		startedAt,
		finishedAt,
		duration,
		queueDuration int64
	)

	if j.BuildID != nil {
		fields["build_id"] = strconv.FormatInt(*j.BuildID, 10)
	}

	if j.BuildCreatedAt != nil {
		createdAt = j.BuildCreatedAt.UnixMilli()
		fields["build_created_at"] = createdAt
	}

	if j.BuildStartedAt != nil {
		startedAt = j.BuildStartedAt.UnixMilli()
		fields["build_started_at"] = startedAt
	}

	if j.BuildQueuedDuration != nil {
		queueDuration = int64(*j.BuildQueuedDuration * float64(time.Millisecond))
		fields["queued_duration"] = queueDuration
	}

	if j.BuildFinishedAt != nil {
		finishedAt = j.BuildFinishedAt.UnixMilli()
		fields["build_finished_at"] = finishedAt
	}
	if j.BuildDuration != nil {
		duration = int64(*j.BuildDuration * float64(time.Second/time.Microsecond))
		fields["build_duration"] = duration
	}

	if j.PipelineID != nil {
		fields["pipeline_id"] = strconv.FormatInt(*j.PipelineID, 10)
	}
	if j.ProjectID != nil {
		fields["project_id"] = strconv.FormatInt(*j.ProjectID, 10)
	}
	if j.Runner != nil && j.Runner.ID != nil {
		fields["runner_id"] = strconv.FormatInt(*j.Runner.ID, 10)
	}
	if j.Commit != nil && j.Commit.Message != nil {
		fields["build_commit_message"] = *j.Commit.Message
		fields["message"] = *j.Commit.Message
	}
	return fields
}

func getJobEventTags(j jobEventPayload) map[string]string {
	tags := map[string]string{}
	if j.ObjectKind != nil {
		tags["object_kind"] = *j.ObjectKind
	}
	if j.SHA != nil {
		tags["sha"] = *j.SHA
	}
	if j.BuildName != nil {
		tags["build_name"] = *j.BuildName
	}
	if j.BuildStage != nil {
		tags["build_stage"] = *j.BuildStage
	}
	if j.BuildStatus != nil {
		tags["build_status"] = *j.BuildStatus
	}
	if j.ProjectName != nil {
		tags["project_name"] = *j.ProjectName
	}
	if j.BuildFailureReason != nil {
		tags["build_failure_reason"] = *j.BuildFailureReason
	}
	if j.User != nil && j.User.Email != nil {
		tags["user_email"] = *j.User.Email
	}
	if j.Commit != nil && j.Commit.SHA != nil {
		tags["build_commit_sha"] = *j.Commit.SHA
	}
	if j.Repository != nil && j.Repository.Name != nil {
		tags["build_repo_name"] = *j.Repository.Name
	}
	return tags
}

func getPipelineEventFields(pl plEventPayload) map[string]interface{} {
	var (
		fields = map[string]interface{}{}
		duration,
		queueDuration,
		createdAt,
		finishedAt int64
	)

	if pl.ObjectAttributes != nil && pl.ObjectAttributes.ID != nil {
		fields["pipeline_id"] = strconv.FormatInt(*pl.ObjectAttributes.ID, 10)
	}
	if pl.ObjectAttributes != nil && pl.ObjectAttributes.Duration != nil {
		duration = *pl.ObjectAttributes.Duration * int64(time.Second/time.Microsecond)
	} else {
		fields["duration"] = int64(0)
	}

	if pl.Commit != nil && pl.Commit.Message != nil {
		fields["commit_message"] = *pl.Commit.Message
		fields["message"] = *pl.Commit.Message
	}

	if pl.ObjectAttributes != nil {
		if pl.ObjectAttributes.CreatedAt != nil {
			createdAt = pl.ObjectAttributes.CreatedAt.UnixMilli()
		}
		if pl.ObjectAttributes.FinishedAt != nil {
			finishedAt = pl.ObjectAttributes.FinishedAt.UnixMilli()
		}

		queueDuration = finishedAt - createdAt - duration/int64(time.Microsecond)
	}

	fields["finished_at"] = finishedAt
	fields["created_at"] = createdAt
	fields["queued_duration"] = queueDuration
	fields["duration"] = duration

	return fields
}

func getPipelineEventTags(pl plEventPayload) map[string]string {
	tags := map[string]string{}
	if pl.ObjectKind != nil {
		tags["object_kind"] = *pl.ObjectKind
	}
	if pl.ObjectAttributes != nil && pl.ObjectAttributes.Status != nil {
		tags["ci_status"] = *pl.ObjectAttributes.Status
	}
	if pl.Project != nil && pl.Project.PathWithNamespace != nil {
		tags["pipeline_name"] = *pl.Project.PathWithNamespace
	}
	if pl.Project != nil && pl.Project.WebURL != nil && pl.ObjectAttributes != nil && pl.ObjectAttributes.ID != nil {
		tags["pipeline_url"] = *pl.Project.WebURL + "/pipelines/" + strconv.FormatInt(*pl.ObjectAttributes.ID, 10)
	}
	if pl.Commit != nil && pl.Commit.ID != nil {
		tags["commit_sha"] = *pl.Commit.ID
	}
	if pl.Commit != nil && pl.Commit.Author != nil && pl.Commit.Author.Email != nil {
		tags["author_email"] = *pl.Commit.Author.Email
	}
	if pl.Project != nil && pl.Project.GitHTTPURL != nil {
		tags["repository_url"] = *pl.Project.GitHTTPURL
	}
	if pl.ObjectAttributes != nil && pl.ObjectAttributes.Source != nil {
		tags["pipeline_source"] = *pl.ObjectAttributes.Source
	}
	if pl.ObjectKind != nil {
		tags["operation_name"] = *pl.ObjectKind
	}
	if pl.Project != nil && pl.Project.Name != nil {
		tags["resource"] = *pl.Project.Name
	}
	if pl.ObjectAttributes != nil && pl.ObjectAttributes.Ref != nil {
		tags["ref"] = *pl.ObjectAttributes.Ref
	}
	return tags
}

func (ipt *Input) getPoint(data []byte, eventType string) ([]*point.Point, error) {
	var tags map[string]string
	var fields map[string]interface{}
	var measurementName string
	switch eventType {
	case pipelineHook:
		measurementName = "gitlab_pipeline"
		var pl plEventPayload
		if err := json.Unmarshal(data, &pl); err != nil {
			return nil, err
		}
		// We need pipeline event with ci status success or failed only.
		if pl.ObjectAttributes == nil || pl.ObjectAttributes.Status == nil {
			l.Debugf("ignore pipeline event with empty ci_status")
			return nil, nil
		}
		if *pl.ObjectAttributes.Status != "success" && *pl.ObjectAttributes.Status != "failed" {
			l.Debugf("ignore pipeline event with ci_status = %s", *pl.ObjectAttributes.Status)
			return nil, nil
		}
		tags = getPipelineEventTags(pl)
		fields = getPipelineEventFields(pl)

	case jobHook:
		measurementName = "gitlab_job"
		var j jobEventPayload
		if err := json.Unmarshal(data, &j); err != nil {
			return nil, fmt.Errorf("invalid job event: %w", err)
		}
		// We need job event with build status success or failed only.
		if j.BuildStatus == nil {
			l.Debugf("ignore job event with empty build_status")
			return nil, nil
		}
		if *j.BuildStatus != "success" && *j.BuildStatus != "failed" {
			l.Infof("ignore job event on status %q", *j.BuildStatus)
			return nil, nil
		}
		tags = getJobEventTags(j)
		fields = getJobEventFields(j)

	default:
		return nil, fmt.Errorf("unrecognized event payload: %v", eventType)
	}

	// add event raw json
	fields["event_raw"] = string(data)

	ipt.addExtraTags(tags)

	opts := point.DefaultLoggingOptions()
	opts = append(opts, point.WithTime(ntp.Now()))

	if ipt.Election {
		tags = inputs.MergeTagsWrapper(tags, ipt.Tagger.ElectionTags(), ipt.Tags, ipt.URL)
	} else {
		tags = inputs.MergeTagsWrapper(tags, ipt.Tagger.HostTags(), ipt.Tags, ipt.URL)
	}

	pt := point.NewPoint(measurementName,
		append(point.NewTags(tags), point.NewKVs(fields)...), opts...)

	return []*point.Point{pt}, nil
}

func (ipt *Input) addExtraTags(tags map[string]string) {
	for k, v := range ipt.CIExtraTags {
		// Existing tags will not be overwritten.
		if _, has := tags[k]; has {
			continue
		}
		tags[k] = v
	}
}

func (ipt *Input) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	event := req.Header.Get(gitlabEventHeader)

	if event != pipelineHook && event != jobHook {
		// Webhooks that return failure codes in the 4xx range
		// are considered to be misconfigured, and these are
		// disabled until you manually re-enable them.
		// Here we still return 200 to prevent webhook from disabling,
		// and log that we receive unrecognized event payload.
		l.Warnf("receive unrecognized event payload: %s, webhook may be misconfigured", event)
		return
	}

	data, err := io.ReadAll(req.Body)
	if err != nil {
		l.Errorf("fail to read from webhook request body: %v", err)
		resp.WriteHeader(http.StatusInternalServerError)
		return
	}

	digest := md5.Sum(data) //nolint:gosec
	if ipt.reqMemo.has(digest) {
		// Skip duplicated requests.
		l.Debugf("skip duplicated event with md5 = %v", digest)
		return
	}
	ipt.reqMemo.add(digest)

	pts, err := ipt.getPoint(data, event)
	if err != nil {
		l.Errorf("get point: %v", err)
		resp.WriteHeader(http.StatusInternalServerError)
		ipt.reqMemo.remove(digest)
		return
	}
	if pts == nil {
		// Skip unwanted events.
		return
	}

	if err := ipt.feeder.Feed(point.Logging, pts,
		dkio.WithElection(ipt.Election),
		dkio.WithSource("gitlab_ci")); err != nil {
		ipt.feeder.FeedLastError(err.Error(),
			metrics.WithLastErrorInput(inputName),
			metrics.WithLastErrorCategory(point.Logging),
		)
		l.Errorf("feed measurement: %s", err)

		resp.WriteHeader(http.StatusInternalServerError)
		ipt.reqMemo.remove(digest)
		return
	}
}

type requestMemo struct {
	memoMap     map[[16]byte]time.Time
	hasReqCh    chan hasRequest
	addReqCh    chan [16]byte
	removeReqCh chan [16]byte
	semStop     *cliutils.Sem
}

type hasRequest struct {
	digest [16]byte
	respCh chan bool
}

func (m *requestMemo) has(digest [16]byte) bool {
	r := hasRequest{
		digest: digest,
		respCh: make(chan bool, 1),
	}
	m.hasReqCh <- r
	return <-r.respCh
}

func (m *requestMemo) add(digest [16]byte) {
	m.addReqCh <- digest
}

func (m *requestMemo) remove(digest [16]byte) {
	m.removeReqCh <- digest
}

// memoMaintainer is in charge of maintaining a memo of requests' md5 sums.
// Saving requests' md5 sums is to prevent sending duplicated points to datakit io.
// Each md5 sum will be expired in du, then they will be removed from memory.
// memoMaintainer will traverse memoMap and remove expired md5 sums every du seconds.
func (m *requestMemo) memoMaintainer(du time.Duration) {
	// Clear expired md5 sums every du seconds.
	ticker := time.NewTicker(du)

	for {
		select {
		case <-ticker.C:
			for k, v := range m.memoMap {
				if time.Since(v) > du {
					delete(m.memoMap, k)
				}
			}

		case r := <-m.hasReqCh:
			_, has := m.memoMap[r.digest]
			r.respCh <- has

		case r := <-m.addReqCh:
			l.Debugf("md5 = %v is added to memo", r)
			m.memoMap[r] = time.Now()

		case r := <-m.removeReqCh:
			l.Debugf("md5 = %v is removed from memo", r)
			delete(m.memoMap, r)

		case <-datakit.Exit.Wait():
			l.Debugf("memoMaintainer exited")
			return

		case <-m.semStop.Wait():
			l.Debugf("memoMaintainer exited")
			return
		}
	}
}
