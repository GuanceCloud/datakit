package cmds

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	markdown "github.com/MichaelMure/go-term-markdown"
	"github.com/c-bata/go-prompt"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/man"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

func cmdMan() {

	switch runtime.GOOS {
	case "windows":
		fmt.Println("\n[E] --man do not support Windows")
		return
	}

	// load input-names
	for k, _ := range inputs.Inputs {
		suggestions = append(suggestions, prompt.Suggest{Text: k, Description: ""})
	}

	for k, _ := range man.OtherDocs {
		suggestions = append(suggestions, prompt.Suggest{Text: k, Description: ""})
	}

	// TODO: add suggestions for pipeline

	c, _ := newCompleter()

	p := prompt.New(
		runMan,
		c.Complete,
		prompt.OptionTitle("man: DataKit manual query"),
		prompt.OptionPrefix("man > "),
	)

	p.Run()
}

func exportMan(to, skipList, ver string) error {
	if err := os.MkdirAll(to, os.ModePerm); err != nil {
		return err
	}

	arr := strings.Split(skipList, ",")
	skip := map[string]bool{}
	for _, x := range arr {
		skip[x] = true
	}

	for k, _ := range inputs.Inputs {
		if skip[k] {
			continue
		}

		data, err := man.BuildMarkdownManual(k, &man.Option{ManVersion: ver, WithCSS: false, IgnoreMissing: true})
		if err != nil {
			return err
		}

		if len(data) == 0 {
			continue
		}

		if err := ioutil.WriteFile(filepath.Join(to, k+".md"), data, os.ModePerm); err != nil {
			return err
		}
	}

	for k, _ := range man.OtherDocs {

		if skip[k] {
			continue
		}

		data, err := man.BuildMarkdownManual(k, &man.Option{ManVersion: ver, WithCSS: false})
		if err != nil {
			return err
		}

		if len(data) == 0 {
			continue
		}

		if err := ioutil.WriteFile(filepath.Join(to, k+".md"), data, os.ModePerm); err != nil {
			return err
		}
	}

	return nil
}

func runMan(txt string) {
	s := strings.Join(strings.Fields(strings.TrimSpace(txt)), " ")
	if s == "" {
		return
	}

	switch s {
	case "Q", "q", "exit":
		fmt.Println("Bye!")
		os.Exit(0)
	default:
		x, err := man.BuildMarkdownManual(s,
			&man.Option{
				WithCSS:                       false,
				DisableMonofontOnTagFieldName: true,
			})

		width := 80
		leftPad := 6
		if err != nil {
			fmt.Printf("[E] %s\n", err.Error())
		} else {
			if len(x) == 0 {
				fmt.Printf("[E] intput %s got no manual", s)
			} else {
				result := markdown.Render(string(x), width, leftPad)
				fmt.Println(string(result))
			}
		}
	}
}
