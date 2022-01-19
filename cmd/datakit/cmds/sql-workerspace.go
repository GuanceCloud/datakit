package cmds

import (
	"strings"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
)

var wss []*workerSpace

type workerSpace struct {
	Name     string
	ID       string
	Token    string
	createAt string
	expireAt string
}

func switchToken(s string) {
	// 切换空间 前提: 之前没有切换过 或者切回到默认的空间
	// 查询缓存中有没有改token或者workspace
	// 将token赋值之后 提示‘已经切换到了xxx空间’
	name := strings.TrimSpace(strings.TrimPrefix(s, "use"))
	if name == "" {
		infof("use workspace command is:use tkn_xxxx\n")
		return
	}
	if name == config.GetToken() {
		temporaryToken = ""
		infof("change workerSpace to default\n")
		return
	}
	if temporaryToken != "" {
		infof("can't change to : %s because workspace is not default\n", name)
		return
	}
	if wss == nil {
		infof("no workerSpace to switch, use 'show_workspaces()'\n")
		return
	}
	for _, space := range wss {
		if space.Name == name || space.Token == name {
			temporaryToken = space.Token
			infof("change workerSpace to %s\n", space.Name)
			return
		}
	}
	infof("[error] invalid workerSpace:'%s' , use 'show_workspaces()'\n", name)
}

func cacheWorkerSpace(c []*queryResult) {
	for _, result := range c {
		cache(result)
	}
}

func cache(c *queryResult) {
	if wss == nil {
		wss = make([]*workerSpace, 0)
	}
	wsIndex, tokenIndex, expIndex, creatIndex, nameIndex := 0, 0, 0, 0, 0
	for _, row := range c.Series {
		for i, column := range row.Columns {
			switch column {
			case "wsuuid":
				wsIndex = i
			case "token":
				tokenIndex = i
			case "expireAt":
				expIndex = i
			case "createAt":
				creatIndex = i
			case "name":
				nameIndex = i
			default:

			}
		}
		for _, value := range row.Values {
			ws := &workerSpace{Name: (value[nameIndex]).(string),
				ID:       (value[wsIndex]).(string),
				Token:    (value[tokenIndex]).(string),
				createAt: (value[creatIndex]).(string),
				expireAt: (value[expIndex]).(string),
			}
			wss = append(wss, ws)
		}
	}
}
