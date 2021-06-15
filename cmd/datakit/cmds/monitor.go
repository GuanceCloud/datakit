package cmds

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"golang.org/x/term"

	markdown "github.com/MichaelMure/go-term-markdown"
	dkhttp "gitlab.jiagouyun.com/cloudcare-tools/datakit/http"
)

func CMDMonitor(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		return nil, fmt.Errorf("%s", string(body))
	}

	ds := dkhttp.DatakitStats{
		DisableMonofont: true,
	}
	if err := json.Unmarshal(body, &ds); err != nil {
		return nil, err
	}

	mdtxt, err := ds.Markdown("")
	if err != nil {
		return nil, err
	}

	width := 100
	if term.IsTerminal(0) {
		w, _, err := term.GetSize(0)
		if err == nil {
			width = w
		}
	}

	if err != nil {
		return nil, err
	} else {
		if len(mdtxt) == 0 {
			return nil, fmt.Errorf("no monitor info available")
		} else {
			result := markdown.Render(string(mdtxt), width, 2)
			return result, nil
		}
	}
}
