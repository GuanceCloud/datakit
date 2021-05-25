package jira

import (
	"fmt"
	"sync"
	"time"

	"github.com/andygrunwald/go-jira"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
)

func (p *JiraParam) active() {
	var queue chan string
	wg := sync.WaitGroup{}

	cnt := 0
	lastQueueId := -1
	issueHasFound := make(map[string]bool)
	ticker := time.NewTicker(time.Duration(20) * time.Minute)
	defer ticker.Stop()

	task := func() {
		issues, err := p.findIssues()
		if err != nil {
			p.log.Errorf("findIssues err: %s", err.Error())
		}
		for _, issue := range issues {
			if _, ok := issueHasFound[issue]; ok {
				continue
			}
			p.log.Debugf("New issue %s", issue)
			issueHasFound[issue] = true
			cnt += 1
			if cnt/maxIssuesPerQueue != lastQueueId {
				lastQueueId = cnt / maxIssuesPerQueue
				queue = make(chan string, maxIssuesPerQueue)
				wg.Add(1)
				go p.gather(queue, &wg)
			}
			queue <- issue
		}
	}

	task()

	for {
		select {
		case <-ticker.C:
			task()
		case <-datakit.Exit.Wait():
			wg.Wait()
			p.log.Info("input jira exit")
			return
		}
	}
}

func (p *JiraParam) gather(queue <-chan string, wg *sync.WaitGroup) {
	var d time.Duration
	var err error

	switch p.input.Interval.(type) {
	case int64:
		d = time.Duration(p.input.Interval.(int64)) * time.Second
	case string:
		d, err = time.ParseDuration(p.input.Interval.(string))
		if err != nil {
			p.log.Errorf("parse interval err: %s", err.Error())
			return
		}
	default:
		p.log.Errorf("interval type unsupported")
		return
	}

	issueL := make([]string, 0)
	ticker := time.NewTicker(d)
	defer ticker.Stop()

	c, err := p.makeJiraClient()
	if err != nil {
		p.log.Errorf("makeJiraClient err: %s", err.Error())
		return
	}

	for {
		select {
		case issue := <-queue:
			p.log.Debugf("Recv issue %v", issue)
			issueL = append(issueL, issue)

		case <-ticker.C:
			for _, issue := range issueL {
				select {
				case <-datakit.Exit.Wait():
					wg.Done()
					return
				default:
				}
				params := *p
				params.input.Issue = issue
				err := params.getMetrics(c)
				if err != nil {
					p.log.Errorf("getMetrics err: %s", err.Error())
				}
			}

		case <-datakit.Exit.Wait():
			wg.Done()
			return
		}
	}
}

func (p *JiraParam) getMetrics(c *jira.Client) error {
	i, resp, err := c.Issue.Get(p.input.Issue, &jira.GetQueryOptions{})
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("http issue get code: %d", resp.StatusCode)
	}
	resp.Body.Close()

	p.log.Debugf("get issue %v", p.input.Issue)

	tags := make(map[string]string)
	fields := make(map[string]interface{})

	fields["id"] = i.ID
	fields["key"] = i.Key
	fields["url"] = i.Self

	if i.Fields != nil {
		tags["project_key"] = i.Fields.Project.Key
		tags["project_id"] = i.Fields.Project.ID
		tags["project_name"] = i.Fields.Project.Name

		fields["type"] = i.Fields.Type.Name
		fields["summary"] = i.Fields.Summary
		if i.Fields.Creator != nil {
			fields["creator"] = i.Fields.Creator.Name
		}
		if i.Fields.Assignee != nil {
			fields["assignee"] = i.Fields.Assignee.Name
		}
		if i.Fields.Reporter != nil {
			fields["reporter"] = i.Fields.Reporter.Name
		}
		if i.Fields.Priority != nil {
			fields["priority"] = i.Fields.Priority.Name
		}
		if i.Fields.Status != nil {
			fields["status"] = i.Fields.Status.Name
		}
	}

	tags["host"] = p.input.Host
	for tag, tagV := range p.input.Tags {
		tags[tag] = tagV
	}

	pts, err := io.MakeMetric(p.input.MetricsName, tags, fields, time.Time(i.Fields.Updated))
	if err != nil {
		return err
	}
	p.log.Debug(string(pts))

	err = p.output.IoFeed(pts, datakit.Metric, inputName)
	return err
}

func (p *JiraParam) makeJiraClient() (*jira.Client, error) {
	tp := jira.BasicAuthTransport{
		Username: p.input.Username,
		Password: p.input.Password,
	}
	return jira.NewClient(tp.Client(), p.input.Host)
}

func (t *JiraParam) findIssues() ([]string, error) {
	//指定issue ID
	if t.input.Issue != "" {
		return t.findIssueById()
	}
	//遍历所有项目下所有问题，耗时较长
	if t.input.Project == "" && t.input.Issue == "" {
		return t.findAllIssue()
	}
	//获取指定项目下所有issue
	return t.findIssuesByProject()
}

func (t *JiraParam) findAllIssue() ([]string, error) {
	issueL := make([]string, 0)

	c, err := t.makeJiraClient()
	if err != nil {
		return issueL, err
	}

	ops := jira.SearchOptions{}
	ops.StartAt = 0
	ops.MaxResults = 300
	for {
		select {
		case <-datakit.Exit.Wait():
			return nil, nil
		default:
		}

		cnt := 0
		issues, resp, err := c.Issue.Search("", &ops)
		resp.Body.Close()
		if resp.StatusCode != 200 {
			return issueL, nil
		}
		if err != nil {
			return issueL, err
		}

		for _, i := range issues {
			cnt += 1
			issueL = append(issueL, i.ID)
		}
		if cnt < ops.MaxResults || cnt == 0 {
			break
		}
		ops.StartAt += ops.MaxResults
	}
	return issueL, nil
}

func (t *JiraParam) findIssuesByProject() ([]string, error) {
	issueL := make([]string, 0)
	t.log.Debugf("findIssuesByProject")
	c, err := t.makeJiraClient()
	if err != nil {
		return issueL, err
	}

	p, resp, err := c.Project.Get(t.input.Project)
	if err != nil {
		return issueL, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return issueL, fmt.Errorf("get project %v with response code %v",
			t.input.Project, resp.StatusCode)
	}
	t.log.Debugf("project key:%v name:%v id:%v", p.Key, p.Name, p.ID)
	ops := jira.SearchOptions{}
	ops.StartAt = 0
	ops.MaxResults = 300
	sql := fmt.Sprintf("project=%s", p.ID)
	t.log.Debugf("sql=%v", sql)
	for {
		select {
		case <-datakit.Exit.Wait():
			return nil, nil
		default:
		}

		issues, _, _ := c.Issue.Search(sql, &ops)
		if err != nil {
			return issueL, err
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			return issueL, fmt.Errorf("search issue with response code %v", resp.StatusCode)
		}

		cnt := 0
		for _, i := range issues {
			cnt += 1
			issueL = append(issueL, i.ID)
		}
		t.log.Debugf("cnt = %d StartAt = %d, MaxResults = %d Total = %d", cnt, resp.StartAt, resp.MaxResults, resp.Total)
		if cnt == resp.Total {
			break
		}
		ops.StartAt += ops.MaxResults
	}
	return issueL, nil
}

func (t *JiraParam) findIssueById() ([]string, error) {
	issueL := make([]string, 0)
	issueL = append(issueL, t.input.Issue)
	return issueL, nil
}
