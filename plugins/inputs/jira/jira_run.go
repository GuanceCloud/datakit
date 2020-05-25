package jira

import (
	"fmt"
	"log"
	"runtime"
	"time"

	"github.com/andygrunwald/go-jira"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal"
)

func (t *JiraTarget) active() {
	var queue chan string
	cnt := 0
	lastQueueId := -1
	issueHasFound := make(map[string]bool)

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		issues, err := t.findIssues()
		logError(err)
		for _, issue := range issues {
			if _, ok := issueHasFound[issue]; ok {
				continue
			}
			issueHasFound[issue] = true
			cnt += 1
			if cnt/maxIssuesPerQueue != lastQueueId {
				lastQueueId = cnt / maxIssuesPerQueue
				queue = make(chan string, maxIssuesPerQueue)
				go t.gather(queue)
			}
			queue <- issue
		}

		//20分钟扫描一次
		internal.SleepContext(ctx, time.Duration(20)*time.Minute)
	}
}

func (t *JiraTarget) gather(queue chan string) {
	issueL := make([]string, 0)
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		isBreak := false
		for {
			if isBreak {
				break
			}
			select {
			case issue := <-queue:
				issueL = append(issueL, issue)
			case <-time.NewTimer(time.Duration(1) * time.Second).C:
				isBreak = true
			}
		}

		for _, issue := range issueL {
			input := JiraInput{
				JiraTarget: *t,
				MetricName: metricName,
			}
			input.Issue = issue
			output := JiraOutput{acc}

			params := JiraParam{input, output}
			params.getMetrics()
		}

		err := internal.SleepContext(ctx, time.Duration(t.Interval)*time.Second)
		logError(err)
	}
}

func (p *JiraParam) getMetrics() {
	c, err := p.input.JiraTarget.makeJiraClient()
	if err != nil {
		logError(err)
		return
	}

	i, resp, err := c.Issue.Get(p.input.Issue, &jira.GetQueryOptions{})
	if err != nil {
		logError(err)
		return
	}
	if resp.StatusCode != 200 {
		return
	}
	resp.Body.Close()

	tags := make(map[string]string)
	fields := make(map[string]interface{})

	tags["host"] = p.input.Host
	tags["project_key"] = i.Fields.Project.Key
	tags["project_id"] = i.Fields.Project.ID
	tags["project_name"] = i.Fields.Project.Name

	fields["id"] = i.ID
	fields["key"] = i.Key
	fields["url"] = i.Self
	if i.Fields != nil {
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
	p.output.acc.AddFields(p.input.MetricName, fields, tags, time.Time(i.Fields.Updated))
}

func (t *JiraTarget) makeJiraClient() (*jira.Client, error) {
	tp := jira.BasicAuthTransport{
		Username: t.Username,
		Password: t.Password,
	}
	return jira.NewClient(tp.Client(), t.Host)
}

func (t *JiraTarget) findIssues() ([]string, error) {
	//指定issue ID
	if t.Issue != "" {
		return t.findIssueById()
	}
	//遍历所有项目下所有问题，耗时较长
	if t.Project == "" && t.Issue == "" {
		return t.findAllIssue()
	}
	//获取指定项目下所有issue
	return t.findIssuesByProject()
}

func (t *JiraTarget) findAllIssue() ([]string, error) {
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
		case <-ctx.Done():
			return issueL, ctx.Err()
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
		if cnt != ops.MaxResults {
			break
		}
		ops.StartAt += ops.MaxResults
	}
	return issueL, nil
}

func (t *JiraTarget) findIssuesByProject() ([]string, error) {
	issueL := make([]string, 0)

	c, err := t.makeJiraClient()
	if err != nil {
		return issueL, err
	}

	p, resp, err := c.Project.Get(t.Project)
	resp.Body.Close()
	if resp.StatusCode != 200 {
		return issueL, nil
	}
	if err != nil {
		return issueL, err
	}
	ops := jira.SearchOptions{}
	ops.StartAt = 0
	ops.MaxResults = 300
	sql := fmt.Sprintf("project=%s", p.Key)

	for {
		select {
		case <-ctx.Done():
			return issueL, ctx.Err()
		default:
		}

		issues, _, _ := c.Issue.Search(sql, &ops)
		resp.Body.Close()
		if resp.StatusCode != 200 {
			return issueL, nil
		}
		if err != nil {
			return issueL, err
		}

		cnt := 0
		for _, i := range issues {
			cnt += 1
			issueL = append(issueL, i.ID)
		}

		if cnt != ops.MaxResults {
			break
		}
		ops.StartAt += ops.MaxResults
	}
	return issueL, nil
}

func (t *JiraTarget) findIssueById() ([]string, error) {
	issueL := make([]string, 0)
	issueL = append(issueL, t.Issue)
	return issueL, nil
}

func logError(err error) {
	if err == nil {
		return
	}
	_, file, line, _ := runtime.Caller(1)
	log.Printf("E! [jira] file: %s line: %d Err: %s", file, line, err.Error())
}

//func (t *JiraTarget) findAllIssue() ([]string, error) {
//	maxIssueGoroutiune := 500
//	issueL := make([]string, 0)
//	c, err := t.makeJiraClient()
//	if err != nil {
//		return issueL, err
//	}
//
//	ops := jira.SearchOptions{}
//	ops.StartAt = 0
//	ops.MaxResults = 1
//	_, resp, err := c.Issue.Search("", &ops)
//	resp.Body.Close()
//	if resp.StatusCode != 200 || err != nil {
//		return issueL, nil
//	}
//
//
//	gn := 0
//	wg := sync.WaitGroup{}
//	mutex := sync.Mutex{}
//
//	for {
//		wg.Add(1)
//		go func(batchNum int) {
//			defer wg.Done()
//
//			c, err := t.makeJiraClient()
//			if err != nil {
//				return
//			}
//
//			total := 0
//			op := jira.SearchOptions{}
//			op.StartAt = batchNum * maxIssueGoroutiune
//			op.MaxResults = 300
//
//			for {
//				select {
//				case <-ctx.Done():
//					return
//				default:
//				}
//
//				cnt := 0
//				isues, rp, err := c.Issue.Search("", &op)
//				rp.Body.Close()
//				if rp.StatusCode != 200 || err != nil {
//					return
//				}
//
//				for _, i := range isues {
//					cnt += 1
//					mutex.Lock()
//					issueL = append(issueL, i.ID)
//					mutex.Unlock()
//				}
//				if cnt != op.MaxResults {
//					return
//				}
//				total += cnt
//				if total >= maxIssueGoroutiune {
//					return
//				}
//				op.StartAt += op.MaxResults
//			}
//		}(gn)
//
//		gn += 1
//		if gn >= resp.Total/maxIssueGoroutiune {
//			break
//		}
//	}
//
//	wg.Wait()
//	return issueL, nil
//}