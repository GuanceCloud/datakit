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

// PipelineEventPayload contains the information for GitLab's pipeline status change event.
type PipelineEventPayload struct {
	ObjectKind       *string                   `json:"object_kind"`
	User             *User                     `json:"user"`
	Project          *Project                  `json:"project"`
	Commit           *Commit                   `json:"commit"`
	ObjectAttributes *PipelineObjectAttributes `json:"object_attributes"`
	MergeRequest     *MergeRequest             `json:"merge_request"`
	Builds           *[]Build                  `json:"builds"`
}

// JobEventPayload contains the information for GitLab's Job status change.
type JobEventPayload struct {
	ObjectKind         *string      `json:"object_kind"`
	Ref                *string      `json:"ref"`
	Tag                *bool        `json:"tag"`
	BeforeSHA          *string      `json:"before_sha"`
	SHA                *string      `json:"sha"`
	BuildID            *int64       `json:"build_id"`
	BuildName          *string      `json:"build_name"`
	BuildStage         *string      `json:"build_stage"`
	BuildStatus        *string      `json:"build_status"`
	BuildStartedAt     *customTime  `json:"build_started_at"`
	BuildFinishedAt    *customTime  `json:"build_finished_at"`
	BuildDuration      *float64     `json:"build_duration"`
	BuildAllowFailure  *bool        `json:"build_allow_failure"`
	BuildFailureReason *string      `json:"build_failure_reason"`
	PipelineID         *int64       `json:"pipeline_id"`
	ProjectID          *int64       `json:"project_id"`
	ProjectName        *string      `json:"project_name"`
	User               *User        `json:"user"`
	Commit             *BuildCommit `json:"commit"`
	Repository         *Repository  `json:"repository"`
	Runner             *Runner      `json:"runner"`
}

// Build contains all the GitLab Build information.
type Build struct {
	ID            *int64         `json:"id"`
	Stage         *string        `json:"stage"`
	Name          *string        `json:"name"`
	Status        *string        `json:"status"`
	CreatedAt     *customTime    `json:"created_at"`
	StartedAt     *customTime    `json:"started_at"`
	FinishedAt    *customTime    `json:"finished_at"`
	When          *string        `json:"when"`
	Manual        *bool          `json:"manual"`
	User          *User          `json:"user"`
	Runner        *Runner        `json:"runner"`
	ArtifactsFile *ArtifactsFile `json:"artifactsfile"`
}

// Runner represents a runner agent.
type Runner struct {
	ID          *int64  `json:"id"`
	Description *string `json:"description"`
	Active      *bool   `json:"active"`
	IsShared    *bool   `json:"is_shared"`
}

// ArtifactsFile contains all the GitLab artifact information.
type ArtifactsFile struct {
	Filename *string `json:"filename"`
	Size     *string `json:"size"`
}

// Commit contains all the GitLab commit information.
type Commit struct {
	ID        *string     `json:"id"`
	Message   *string     `json:"message"`
	Title     *string     `json:"title"`
	Timestamp *customTime `json:"timestamp"`
	URL       *string     `json:"url"`
	Author    *Author     `json:"author"`
	Added     *[]string   `json:"added"`
	Modified  *[]string   `json:"modified"`
	Removed   *[]string   `json:"removed"`
}

// BuildCommit contains all the GitLab build commit information.
type BuildCommit struct {
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

// User contains all the GitLab user information.
type User struct {
	ID        *int64  `json:"id"`
	Name      *string `json:"name"`
	UserName  *string `json:"username"`
	AvatarURL *string `json:"avatar_url"`
	Email     *string `json:"email"`
}

// Project contains all the GitLab project information.
type Project struct {
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

// Repository contains all the GitLab repository information.
type Repository struct {
	Name            *string `json:"name"`
	URL             *string `json:"url"`
	Description     *string `json:"description"`
	Homepage        *string `json:"homepage"`
	GitSSHURL       *string `json:"git_ssh_url"`
	GitHTTPURL      *string `json:"git_http_url"`
	VisibilityLevel *int64  `json:"visibility_level"`
}

// ObjectAttributes contains all the GitLab object attributes information.
type ObjectAttributes struct {
	ID               *int64      `json:"id"`
	Title            *string     `json:"title"`
	AssigneeIDS      *[]int64    `json:"assignee_ids"`
	AssigneeID       *int64      `json:"assignee_id"`
	AuthorID         *int64      `json:"author_id"`
	ProjectID        *int64      `json:"project_id"`
	CreatedAt        *customTime `json:"created_at"`
	UpdatedAt        *customTime `json:"updated_at"`
	UpdatedByID      *int64      `json:"updated_by_id"`
	LastEditedAt     *customTime `json:"last_edited_at"`
	LastEditedByID   *int64      `json:"last_edited_by_id"`
	RelativePosition *int64      `json:"relative_position"`
	Position         *Position   `json:"position"`
	BranchName       *string     `json:"branch_name"`
	Description      *string     `json:"description"`
	MilestoneID      *int64      `json:"milestone_id"`
	State            *string     `json:"state"`
	StateID          *int64      `json:"state_id"`
	Confidential     *bool       `json:"confidential"`
	DiscussionLocked *bool       `json:"discussion_locked"`
	DueDate          *customTime `json:"due_date"`
	TimeEstimate     *int64      `json:"time_estimate"`
	TotalTimeSpent   *int64      `json:"total_time_spent"`
	IID              *int64      `json:"iid"`
	URL              *string     `json:"url"`
	Action           *string     `json:"action"`
	TargetBranch     *string     `json:"target_branch"`
	SourceBranch     *string     `json:"source_branch"`
	SourceProjectID  *int64      `json:"source_project_id"`
	TargetProjectID  *int64      `json:"target_project_id"`
	StCommits        *string     `json:"st_commits"`
	MergeStatus      *string     `json:"merge_status"`
	Content          *string     `json:"content"`
	Format           *string     `json:"format"`
	Message          *string     `json:"message"`
	Slug             *string     `json:"slug"`
	Ref              *string     `json:"ref"`
	Tag              *bool       `json:"tag"`
	SHA              *string     `json:"sha"`
	BeforeSHA        *string     `json:"before_sha"`
	Status           *string     `json:"status"`
	Stages           *[]string   `json:"stages"`
	Duration         *int64      `json:"duration"`
	Note             *string     `json:"note"`
	NotebookType     *string     `json:"noteable_type"` // nolint:misspell
	At               *customTime `json:"attachment"`
	LineCode         *string     `json:"line_code"`
	CommitID         *string     `json:"commit_id"`
	NoteableID       *int64      `json:"noteable_id"` // nolint: misspell
	System           *bool       `json:"system"`
	WorkInProgress   *bool       `json:"work_in_progress"`
	StDiffs          *[]StDiff   `json:"st_diffs"`
	Source           *Source     `json:"source"`
	Target           *Target     `json:"target"`
	LastCommit       *LastCommit `json:"last_commit"`
	Assignee         *Assignee   `json:"assignee"`
}

// PipelineObjectAttributes contains pipeline specific GitLab object attributes information.
type PipelineObjectAttributes struct {
	ID         *int64      `json:"id"`
	Ref        *string     `json:"ref"`
	Tag        *bool       `json:"tag"`
	SHA        *string     `json:"sha"`
	BeforeSHA  *string     `json:"before_sha"`
	Source     *string     `json:"source"`
	Status     *string     `json:"status"`
	Stages     *[]string   `json:"stages"`
	CreatedAt  *customTime `json:"created_at"`
	FinishedAt *customTime `json:"finished_at"`
	Duration   *int64      `json:"duration"`
	Variables  *[]Variable `json:"variables"`
}

// Variable contains pipeline variables.
type Variable struct {
	Key   *string `json:"key"`
	Value *string `json:"value"`
}

// Position defines a specific location, identified by paths line numbers and
// image coordinates, within a specific diff, identified by start, head and
// base commit ids.
//
// Text position will have: new_line and old_line
// Image position will have: width, height, x, y.
type Position struct {
	BaseSHA      *string `json:"base_sha"`
	StartSHA     *string `json:"start_sha"`
	HeadSHA      *string `json:"head_sha"`
	OldPath      *string `json:"old_path"`
	NewPath      *string `json:"new_path"`
	PositionType *string `json:"position_type"`
	OldLine      *int64  `json:"old_line"`
	NewLine      *int64  `json:"new_line"`
	Width        *int64  `json:"width"`
	Height       *int64  `json:"height"`
	X            *int64  `json:"x"`
	Y            *int64  `json:"y"`
}

// MergeRequest contains all the GitLab merge request information.
type MergeRequest struct {
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
	Source          *Source     `json:"source"`
	Target          *Target     `json:"target"`
	LastCommit      *LastCommit `json:"last_commit"`
	WorkInProgress  *bool       `json:"work_in_progress"`
	Assignee        *Assignee   `json:"assignee"`
	URL             *string     `json:"url"`
}

// Assignee contains all the GitLab assignee information.
type Assignee struct {
	ID        *int64  `json:"id"`
	Name      *string `json:"name"`
	Username  *string `json:"username"`
	AvatarURL *string `json:"avatar_url"`
	Email     *string `json:"email"`
}

// StDiff contains all the GitLab diff information.
type StDiff struct {
	Diff        *string `json:"diff"`
	NewPath     *string `json:"new_path"`
	OldPath     *string `json:"old_path"`
	AMode       *string `json:"a_mode"`
	BMode       *string `json:"b_mode"`
	NewFile     *bool   `json:"new_file"`
	RenamedFile *bool   `json:"renamed_file"`
	DeletedFile *bool   `json:"deleted_file"`
}

// Source contains all the GitLab source information.
type Source struct {
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

// Target contains all the GitLab target information.
type Target struct {
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

// LastCommit contains all the GitLab last commit information.
type LastCommit struct {
	ID        *string     `json:"id"`
	Message   *string     `json:"message"`
	Timestamp *customTime `json:"timestamp"`
	URL       *string     `json:"url"`
	Author    *Author     `json:"author"`
}

// Author contains all the GitLab author information.
type Author struct {
	Name  *string `json:"name"`
	Email *string `json:"email"`
}

func getJobEventFields(j JobEventPayload) map[string]interface{} {
	fields := map[string]interface{}{}
	if j.BuildID != nil {
		fields["build_id"] = strconv.FormatInt(*j.BuildID, 10)
	}
	if j.BuildStartedAt != nil {
		fields["build_started_at"] = j.BuildStartedAt.UnixMilli()
	}
	if j.BuildFinishedAt != nil {
		fields["build_finished_at"] = j.BuildFinishedAt.UnixMilli()
	}
	if j.BuildDuration != nil {
		fields["build_duration"] = int64(*j.BuildDuration * float64(time.Second/time.Microsecond))
	} else {
		fields["build_duration"] = int64(0)
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

func getJobEventTags(j JobEventPayload) map[string]string {
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

func getPipelineEventFields(pl PipelineEventPayload) map[string]interface{} {
	fields := map[string]interface{}{}
	if pl.ObjectAttributes != nil && pl.ObjectAttributes.ID != nil {
		fields["pipeline_id"] = strconv.FormatInt(*pl.ObjectAttributes.ID, 10)
	}
	if pl.ObjectAttributes != nil && pl.ObjectAttributes.Duration != nil {
		fields["duration"] = *pl.ObjectAttributes.Duration * int64(time.Second/time.Microsecond)
	} else {
		fields["duration"] = int64(0)
	}
	if pl.Commit != nil && pl.Commit.Message != nil {
		fields["commit_message"] = *pl.Commit.Message
		fields["message"] = *pl.Commit.Message
	}
	if pl.ObjectAttributes != nil {
		if pl.ObjectAttributes.CreatedAt != nil {
			fields["created_at"] = pl.ObjectAttributes.CreatedAt.UnixMilli()
		}
		if pl.ObjectAttributes.FinishedAt != nil {
			fields["finished_at"] = pl.ObjectAttributes.FinishedAt.UnixMilli()
		}
	}
	return fields
}

func getPipelineEventTags(pl PipelineEventPayload) map[string]string {
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
		var pl PipelineEventPayload
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
		var j JobEventPayload
		if err := json.Unmarshal(data, &j); err != nil {
			return nil, err
		}
		// We need job event with build status success or failed only.
		if j.BuildStatus == nil {
			l.Debugf("ignore job event with empty build_status")
			return nil, nil
		}
		if *j.BuildStatus != "success" && *j.BuildStatus != "failed" {
			l.Debugf("ignore job event with build_status = '%s'", *j.BuildStatus)
			return nil, nil
		}
		tags = getJobEventTags(j)
		fields = getJobEventFields(j)

	default:
		return nil, fmt.Errorf("unrecognized event payload: %v", eventType)
	}
	ipt.addExtraTags(tags)

	opts := point.DefaultLoggingOptions()
	opts = append(opts, point.WithTime(time.Now()))

	if ipt.Election {
		tags = inputs.MergeTagsWrapper(tags, ipt.Tagger.ElectionTags(), ipt.Tags, ipt.URL)
	} else {
		tags = inputs.MergeTagsWrapper(tags, ipt.Tagger.HostTags(), ipt.Tags, ipt.URL)
	}

	pt := point.NewPointV2(measurementName,
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

	if err := ipt.feeder.FeedV2(point.Logging, pts,
		dkio.WithElection(ipt.Election),
		dkio.WithInputName("gitlab_ci")); err != nil {
		ipt.feeder.FeedLastError(err.Error(),
			dkio.WithLastErrorInput(inputName),
			dkio.WithLastErrorCategory(point.Logging),
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
