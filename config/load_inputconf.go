package config

import (
	"bytes"
	"io/ioutil"
	"path/filepath"
	"reflect"

	bstoml "github.com/BurntSushi/toml"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

func doLoadConf(confData string, creators map[string]inputs.Creator) (map[string][]inputs.Input, error) {
	ret := map[string][]inputs.Input{}

	var res map[string]interface{}

	if _, err := bstoml.Decode(confData, &res); err != nil {
		l.Warnf("bstoml.Decode: %s, ignored", err)
		return nil, err
	}

	//	对现有的 conf 而言，无非如下格式：
	//	  [inputs.xxx]
	//	或者
	//	  [[inputs.xxx]]
	//	不管怎么解析，第一层是一个 map[string]interface{}，这里的 string 就是 inputs
	//	第二层，就是具体的 inputs 名称，但它本质上也是一个 map[string]interface{}，这里的 string 就是上面的 xxx
	//	第三层，就是具体的采集器了，它可以是数组形式的，也可以是非数组形式的

	for k, v := range res { // 第一层
		if k != "inputs" {
			l.Warnf("ingore none input conf: %s, ignored", k)
			continue
		}

		switch x := v.(type) {
		case map[string]interface{}: // 第二层
			for inputName, b := range x {

				c, ok := creators[inputName]
				if !ok {
					l.Warnf("unknown input: %s, ignored", inputName)
					continue
				}

				switch y := b.(type) { // 第三层
				case []map[string]interface{}: // it's a inputs array: [[inputs.xxx]]
					for _, input := range y {

						l.Debugf("input: %+#v", input)

						if i, err := constructInput(input, c); err != nil {
							l.Errorf("constructInput: %s, ignored", err)
						} else {
							ret[inputName] = append(ret[inputName], i)
						}
					}

				case map[string]interface{}: // it's a single input: [inputs.xxx]
					if i, err := constructInput(y, c); err != nil {
						l.Errorf("constructInput: %s, ignored", err)
					} else {
						ret[inputName] = append(ret[inputName], i)
					}

				default:
					l.Warnf("unexpect input conf, got type %s, ignored", reflect.TypeOf(b).String())
				}
			}

		default:
			l.Warnf("ignore %s, got type %s", k, reflect.TypeOf(v).String())
		}
	}

	return ret, nil
}

// LoadInputConf read all inputs configures(toml) from @root,
// then create various inputs object
func LoadInputConf(root string) (ret map[string][]inputs.Input) {
	confs := SearchDir(root, ".conf")

	ret = map[string][]inputs.Input{}

	for _, fp := range confs {
		data, err := ioutil.ReadFile(filepath.Clean(fp))
		if err != nil {
			l.Errorf("ioutil.ReadFile: %s", err.Error())
			return nil
		}

		x, err := doLoadConf(string(data), inputs.Inputs)
		if err != nil {
			l.Warnf("load conf(%s) failed: %s, ignored", fp, err)
			continue
		}

		for k, v := range x {
			ret[k] = append(ret[k], v...)
		}
	}

	return ret
}

func constructInput(x interface{}, c inputs.Creator) (inputs.Input, error) {
	i := c()
	var buf bytes.Buffer

	l.Debugf("input type %s", reflect.TypeOf(x).String())

	if err := bstoml.NewEncoder(&buf).Encode((x)); err != nil {
		l.Errorf("Encode: %s", err)
		return nil, err
	}

	l.Debugf("buf: %s", buf.String())

	if _, err := bstoml.Decode(buf.String(), i); err != nil {
		l.Errorf("Decode: %s", err)
		return nil, err
	}

	return i, nil
}
