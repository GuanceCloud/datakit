package gitlab

import (
	"log"
	"time"

	"github.com/xanzy/go-gitlab"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal"
)

func (t *GitlabTarget) active() error {
	pbm, err := t.getProjectAndBranch()
	if err != nil {
		return err
	}
	for k, v := range pbm {
		for _, b := range v {
			input := GitlabInput{
				GitlabTarget: *t,
				MetricName:   metricName,
			}
			input.Project = k
			input.Branch  = b

			output := GitlabOutput{acc}
			p := GitlabParam{input, output}
			go p.gather()
		}
	}
	return nil
}

func (t *GitlabTarget) getProjectAndBranch() (map[interface{}][]string, error) {
	pBM := make(map[interface{}][]string)

	client, err := gitlab.NewClient(t.Token, gitlab.WithBaseURL(t.Host))
	if err != nil {
		return nil, err
	}

	if t.Project == nil {
		listOps := gitlab.ListProjectsOptions{}
		nextPage := 1
		listOps.PerPage = 100

		for {
			listOps.Page = nextPage
			ps, resp, _ := client.Projects.ListProjects(&listOps)
			for _, p := range ps {
				pBM[p.ID] = make([]string, 0)
			}
			nextPage = resp.NextPage
			if nextPage == 0 {
				break
			}
		}
	} else {
		pBM[t.Project] = make([]string, 0)
	}

	for p, _ := range pBM {
		if t.Branch == "" {
			bs, err := t.getBranchsByProject(client, p)
			if err != nil {
				continue
			}
			pBM[p] = append(pBM[p], bs...)
		} else {
			pBM[p] = append(pBM[p], t.Branch)
		}
	}
	return pBM, nil

}
func (t *GitlabTarget) getBranchsByProject(client *gitlab.Client, project interface{}) ([]string, error) {
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

func (p *GitlabParam) gather() {
	var stopTime  time.Time
	var startTime time.Time

	startTime, err := getCommitStartDate(p.input.StartDate)
	if err != nil {
		log.Printf("W! [gitlab] %s", err.Error())
		return
	}

	client, err := gitlab.NewClient(p.input.Token, gitlab.WithBaseURL(p.input.Host))
	if err != nil {
		log.Printf("W! [gitlab] %s", err.Error())
		return
	}

	pj, _, err := client.Projects.GetProject(p.input.Project, &gitlab.GetProjectOptions{})
	if err != nil {
		log.Printf("W! [gitlab] %s", err.Error())
		return
	}

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		stopTime = getCommitStopDate(startTime, p.input.HoursBatch)
		err := p.getCommitMetrics(client, pj.Name, startTime, stopTime)
		if err != nil {
			if err != nil {
				log.Printf("W! [gitlab] %s", err.Error())
			}
		}

		startTime = stopTime

		err = internal.SleepContext(ctx, time.Duration(p.input.Interval)*time.Second)
		if err != nil {
			log.Printf("W! [gitlab] %s", err.Error())
		}

	}
}

func (p *GitlabParam) getCommitMetrics(client *gitlab.Client, pjName string, start time.Time, stop time.Time) error {
	var tags map[string]string
	var fields map[string]interface{}

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
			fields = make(map[string]interface{})

			tags["host"] = p.input.Host
			tags["branch"] = p.input.Branch
			tags["project_name"] = pjName
			tags["author_name"] = commit.AuthorName
			tags["author_email"] = commit.AuthorEmail
			tags["comitter_name"] = commit.CommitterName
			tags["comitter_email"] = commit.CommitterEmail

			fields["commit_id"] = commit.ID
			fields["title"] = commit.Title
			fields["message"] = commit.Message

			p.output.acc.AddFields(p.input.MetricName, fields, tags, *commit.CreatedAt)

		}
		nextPage = resp.NextPage
		if nextPage == 0 {
			break
		}
	}
	return nil
}

func getCommitStartDate(timeStr string) (time.Time, error) {
	var startTime time.Time
	var err error

	startTime, err = time.Parse("2006-01-02T15:04:05", timeStr)
	if err != nil {
		startTime, err = time.Parse(time.RFC3339, timeStr)
		if err != nil {
			return startTime, err
		}
	}
	return startTime, nil
}

func getCommitStopDate(s time.Time, hoursBatch int) time.Time {
	var stopTime time.Time
	now := time.Now()

	stopTime = s.Add(time.Duration(hoursBatch)*time.Hour)
	if stopTime.After(now) {
		return now
	}
	return stopTime
}