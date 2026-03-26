# Global Settings {#global}

these global settings are applied for all sections in this markdown.

- the {{.Version}} comes from *gitlab-ci.yml* on variable `CI_VERSION`, you can grep on `cat gitlab-ci.yml | grep -w "CI_VERSION"`
- read gitlab milestone `https://gitlab.jiagouyun.com/cloudcare-tools/datakit/-/milestones`, select the milestone same as {{.Version}}

## For milestone changelog {#ms-cl}

export changelog with the following requirements:

- read history change log in *internal/export/doc/zh/changelog-{{.Year}}.md*
- read the gitlab access token from ENV `GITLAB_ACCESS_TOKEN`
- read gitlab milestone
- for all closed/merged issue, add new changelog to the top of changelog-{{.Year}}.md above. do not use emoji
- also add english changelog to *internal/export/doc/en/changelog-{{.Year}}.md*

## For milestone code review {#ms-mr}

review milestone code with the following requirements:

- read the gitlab access token from ENV `GITLAB_REVIEW_ACCESS_TOKEN`
- read gitlab milestone `https://gitlab.jiagouyun.com/cloudcare-tools/datakit/-/milestones`, select the milestone same as {{.Version}}
- for all issue:
    - read the issue and it's related merge request(closed or open), one issue may have one or more merge request(ignore the closed merge request)
    - review the code for the issue, give advice to code lines if any.
    - add or update a comment(conclusion) for the issue's merge request

## For merge request review {#mr}

- review specific merge request, add your summary to merge request's comment list.
- for the same merge request, if review again, update your summary and create a new one.
