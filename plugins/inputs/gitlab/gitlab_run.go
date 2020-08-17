package gitlab

import (
	"io/ioutil"
	"time"

	"github.com/xanzy/go-gitlab"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
)

func (g *GitlabParam) gather() {
	var start, stop time.Time
	var d time.Duration
	var err error

	switch g.input.Interval.(type) {
	case int64:
		d = time.Duration(g.input.Interval.(int64)) * time.Second
	case string:
		d, err = time.ParseDuration(g.input.Interval.(string))
		if err != nil {
			g.log.Errorf("parse interval err: %s", err.Error())
			return
		}
	default:
		g.log.Errorf("interval type unsupported")
		return
	}
	client, err := gitlab.NewClient(g.input.Token, gitlab.WithBaseURL(g.input.Host))
	if err != nil {
		g.log.Errorf("NewClient err: %s", err.Error())
		return
	}

	foundPBM := make(map[interface{}]map[string]bool)
	ticker := time.NewTicker(d)
	defer ticker.Stop()

	ticker1 := time.NewTicker(time.Duration(10) * time.Minute)
	defer ticker1.Stop()

	err = g.input.getProjectAndBranch(client, foundPBM)
	if err != nil {
		g.log.Errorf("getProjectAndBranch err: %s", err.Error())
	}

	start = g.getStartDate()

	for {
		select {
		case <-ticker1.C:
			err := g.input.getProjectAndBranch(client, foundPBM)
			if err != nil {
				g.log.Errorf("getProjectAndBranch err: %s", err.Error())
			}

		case <-ticker.C:
			stop = g.getStopDate(start)
			g.log.Debugf("| %v -> %v | %v", start.Format(time.RFC3339), stop.Format(time.RFC3339), foundPBM)
			for p, pj := range foundPBM {
				for b, _ := range pj {
					param := *g
					param.input.Project = p
					param.input.Branch = b
					err = param.getCommitMetrics(client, start, stop)
					if err != nil {
						g.log.Errorf("getCommitMetrics err: %s", err.Error())
					}
				}
			}
			start = stop
			g.updateTimeFile(stop)

		case <-datakit.Exit.Wait():
			g.log.Info("input gitlab exit")
			return
		}
	}
}

func (t *GitlabInput) getProjectAndBranch(client *gitlab.Client, pBM map[interface{}]map[string]bool) error {
	if t.Project == nil {
		listOps := gitlab.ListProjectsOptions{}
		nextPage := 1
		listOps.PerPage = 100

		for {
			listOps.Page = nextPage
			ps, resp, _ := client.Projects.ListProjects(&listOps)
			for _, p := range ps {
				pBM[p.ID] = make(map[string]bool)
			}
			nextPage = resp.NextPage
			if nextPage == 0 {
				break
			}
		}
	} else {
		pBM[t.Project] = make(map[string]bool)
	}

	for p, _ := range pBM {
		if t.Branch == "" {
			bs, err := t.getBranchsByProject(client, p)
			if err != nil {
				continue
			}
			for _, b := range bs {
				pBM[p][b] = true
			}
		} else {
			pBM[p][t.Branch] = true
		}
	}

	return nil
}

func (t *GitlabInput) getBranchsByProject(client *gitlab.Client, project interface{}) ([]string, error) {
	bs := make([]string, 0)
	nextPage := 1

	listOps := gitlab.ListBranchesOptions{}
	listOps.PerPage = 100

	for {
		listOps.Page = nextPage
		branch, resp, err := client.Branches.ListBranches(project, &listOps)
		if err != nil {
			return bs, err
		}
		for _, b := range branch {
			bs = append(bs, b.Name)
		}

		nextPage = resp.NextPage
		if nextPage == 0 {
			break
		}
	}

	return bs, nil
}

func (p *GitlabParam) getCommitMetrics(client *gitlab.Client, start time.Time, stop time.Time) error {
	var tags map[string]string
	var fields map[string]interface{}

	pj, _, err := client.Projects.GetProject(p.input.Project, &gitlab.GetProjectOptions{})
	if err != nil {
		return err
	}

	nextPage := 1
	listOps := gitlab.ListCommitsOptions{}
	listOps.PerPage = 100
	listOps.RefName = gitlab.String(p.input.Branch)
	listOps.Since = gitlab.Time(start)
	listOps.Until = gitlab.Time(stop)

	for {
		listOps.Page = nextPage
		commits, resp, err := client.Commits.ListCommits(p.input.Project, &listOps)
		if err != nil {
			return nil
		}
		for _, commit := range commits {
			tags = make(map[string]string)
			for tag, tagV := range p.input.Tags {
				tags[tag] = tagV
			}
			fields = make(map[string]interface{})

			tags["host"] = p.input.Host
			tags["branch"] = p.input.Branch
			tags["project_name"] = pj.Name
			tags["author_name"] = commit.AuthorName
			tags["author_email"] = commit.AuthorEmail
			tags["comitter_name"] = commit.CommitterName
			tags["comitter_email"] = commit.CommitterEmail

			fields["commit_id"] = commit.ID
			fields["title"] = commit.Title
			fields["message"] = commit.Message

			pts, err := io.MakeMetric(p.input.MetricsName, tags, fields, *commit.CreatedAt)
			if err != nil {
				return err
			}

			err = p.output.IoFeed(pts, io.Metric, inputName)
			p.log.Debugf(string(pts))
			if err != nil {
				return err
			}
		}
		nextPage = resp.NextPage
		if nextPage == 0 {
			break
		}
	}
	return nil
}

func (p *GitlabParam) getStartDate() time.Time {
	var err error
	var t time.Time

	content, err := ioutil.ReadFile(p.input.jsFile)
	if err != nil {
		t, err = parseTimeStr(string(content))
		if err == nil {
			return t
		}
	}

	t, err = parseTimeStr(p.input.StartDate)
	if err == nil {
		return t
	}

	t, _ = parseTimeStr(defaultStartDate)
	return t
}

func (p *GitlabParam) getStopDate(s time.Time) time.Time {
	var stopTime time.Time
	now := time.Now()

	stopTime = s.Add(time.Duration(p.input.HoursBatch) * time.Hour)
	if stopTime.After(now) {
		return now
	}
	return stopTime
}

func parseTimeStr(timeStr string) (time.Time, error) {
	startTime, err := time.Parse("2006-01-02T15:04:05", timeStr)
	if err != nil {
		startTime, err = time.Parse(time.RFC3339, timeStr)
		if err != nil {
			return startTime, err
		}
	}
	return startTime, nil
}

func (p *GitlabParam) updateTimeFile(stop time.Time) {
	tStr := stop.Format("2006-01-02T15:04:05")
	ioutil.WriteFile(p.input.jsFile, []byte(tStr), 0x666)
}
