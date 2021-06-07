package cmds

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"bytes"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"
	"github.com/c-bata/go-prompt"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/git"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/kubernetes"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

func BuildK8sConfig(name string, interactive bool) {
	switch runtime.GOOS {
	case "windows":
		fmt.Println("\n[E] --man do not support Windows")
		return
	}

	// load input-names
	for k, _ := range inputs.Inputs {
		suggestions = append(suggestions, prompt.Suggest{Text: k, Description: ""})
	}

	k := &KubeDeploy{
		DeployName: name,
		Inputs: make(map[string]string),
	}

	if interactive {
		c, _ := newCompleter()
		p := prompt.New(
			k.run,
			c.Complete,
			prompt.OptionTitle("k8s deploy: generate k8s deploy config"),
			prompt.OptionPrefix("k8s > "),
		)

		p.Run()
	} else {
        // generate all
        for item, _ := range inputs.Inputs {
        	err := k.buildConfig(item, &Option{})
			if err != nil {
				fmt.Printf("[E] %s\n", err.Error())
			}
		}
	}
}

type KubeDeploy struct {
	DeployName   string
	Inputs       map[string]string
	Version      string
	ReleaseDate  string
}

func (k *KubeDeploy) run(txt string) {
	k.runCmd(os.Stdout, txt)
}

func (k *KubeDeploy) runCmd(writer io.Writer, txt string) {
	s := strings.Join(strings.Fields(strings.TrimSpace(txt)), " ")
	if s == "" {
		return
	}

	switch s {
	case "Q", "q", "exit":
		fmt.Fprint(writer, "Bye!")
		os.Exit(0)
		return
	case "render":
		fmt.Fprint(writer, "start render kubernetes deploy config...")
		if err := k.render(); err != nil {
			fmt.Printf("[E] generate k8s config error %s\n", err.Error())
		}
		os.Exit(0)
		return
	default:
		err := k.buildConfig(s, &Option{})
		if err != nil {
			fmt.Fprintf(writer, "[E] %s\n", err.Error())
		}
	}
}

type Option struct {
	IgnoreMissing    bool
}

func (k *KubeDeploy) buildConfig(name string, opt *Option) error {
	ver := git.Version
	c, ok := inputs.Inputs[name]
	if !ok {
		return fmt.Errorf("intput %s not found", name)
	}

	input := c()
	switch i := input.(type) {
	case inputs.InputV2:
		k.Version = ver
		k.ReleaseDate = git.BuildAt
		k.Version = ver
		k.Inputs[name] = i.SampleConfig()
	default:
		l.Warnf("incomplete input: %s", name)
		return nil
	}
    return nil
}

func (k *KubeDeploy) render() error {
	K8sDeployDir := filepath.Join(datakit.InstallDir, "deploy")
	filename     := filepath.Join(K8sDeployDir, k.DeployName + ".yaml")

	if err := os.MkdirAll(K8sDeployDir, os.ModePerm); err != nil {
		return err
	}

	md, err := kubernetes.Get(k.DeployName)
	if err != nil {
		return err
	}

	temp, err := template.New(k.DeployName).Parse(md)
	if err != nil {
		fmt.Printf("[E] template new error %v \n", err)
		return err
	}

	var buf bytes.Buffer
	if err := temp.Execute(&buf, k); err != nil {
		fmt.Printf("[E] template render error %v \n", err)
		return err
	}

	if err := ioutil.WriteFile(filename, buf.Bytes(), os.ModePerm); err != nil {
		return err
	}

	return nil
}

