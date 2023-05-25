// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package build

import (
	T "testing"

	"github.com/stretchr/testify/assert"
)

func Test_doMatch(t *T.T) {
	t.Run("punctuation", func(t *T.T) {
		_, d := doMatch(`中,通.非!无:就;`, true)
		assert.Equal(t, `中，通。非！无：就；`, d)

		_, d = doMatch(`中; `, true) // with extra space
		assert.Equal(t, `中；`, d)

		_, d = doMatch(`:就`, true)
		assert.Equal(t, `：就`, d)

		_, d = doMatch(`abc中`, true)
		assert.Equal(t, `abc 中`, d)

		_, d = doMatch(`中abc`, true)
		assert.Equal(t, `中 abc`, d)

		// multiline
		_, d = doMatch(`中abc
		abc中`, true)
		assert.Equal(t, `中 abc
		abc 中`, d)
	})

	t.Run("external-link", func(t *T.T) {
		txt := `[abc](http://xxx)`
		arr, d := doMatch(txt, true)
		assert.Equal(t, d, `[abc](http://xxx){:target="_blank"}`)
		assert.NotEmpty(t, arr)
		t.Logf("arr: %+#v, d: %q", arr, d)

		txt = `[abc](http://yyy.com){:target="_blank"}`
		arr, d = doMatch(txt, true)
		t.Logf("arr: %+#v, d: %q", arr, d)
		assert.Empty(t, arr)
		assert.Equal(t, txt, d)

		txt = `[*datakit.yaml*](https://static.guance.com/datakit/datakit.yaml){:target="_blank"}，其中开启了很多[默认采集器](datakit-input-conf.md#default-enabled-inputs)`
		arr, d = doMatch(txt, true)
		t.Logf("arr: %+#v, d: %q", arr, d)
		assert.Empty(t, arr, 0)
		assert.Equal(t, txt, d)

		// external links that exclude
		txt = `[some-image](https://static.guance.com/datakit/datakit.png)，其中开启了很多[默认采集器](datakit-input-conf.md#default-enabled-inputs)`
		arr, d = doMatch(txt, true)
		t.Logf("arr: %+#v, d: %q", arr, d)
		assert.Empty(t, arr, 0)

		txt = `

			*mkdocs.sh* 依赖的基础环境：

			<!-- 1. 先将[文档库](https://gitlab.jiagouyun.com/zy-docs/dataflux-doc){:target="_blank"} clone 到本地目录 *~/git/dataflux-doc*，此处默认即使用这个本地目录。*mkdocs.sh* 会将 Datakit 文档生成并拷贝到该 repo 对应的目录下。 -->
			1. 先将[文档库](https://gitlab.jiagouyun.com/zy-docs/dataflux-doc) clone 到本地目录 *~/git/dataflux-doc*，此处默认即使用这个本地目录。*mkdocs.sh* 会将 Datakit 文档生成并拷贝到该 repo 对应的目录下。
			1. 在 *dataflux-doc* 项目下，有一个 *requirements.txt*，执行  pip install -r requirements.txt  安装对应依赖
			1. 回到 Datakit 代码目录，执行根目录下的  ./mkdocs.sh  即可
			`
		arr, d = doMatch(txt, true)
		t.Logf("arr: %+#v", arr)
		t.Logf("fixed: %q", d)
		assert.NotEmpty(t, arr)
		assert.Equal(t, d, `

			*mkdocs.sh* 依赖的基础环境：

			<!-- 1. 先将[文档库](https://gitlab.jiagouyun.com/zy-docs/dataflux-doc){:target="_blank"} clone 到本地目录 *~/git/dataflux-doc*，此处默认即使用这个本地目录。*mkdocs.sh* 会将 Datakit 文档生成并拷贝到该 repo 对应的目录下。 -->
			1. 先将[文档库](https://gitlab.jiagouyun.com/zy-docs/dataflux-doc){:target="_blank"} clone 到本地目录 *~/git/dataflux-doc*，此处默认即使用这个本地目录。*mkdocs.sh* 会将 Datakit 文档生成并拷贝到该 repo 对应的目录下。
			1. 在 *dataflux-doc* 项目下，有一个 *requirements.txt*，执行  pip install -r requirements.txt  安装对应依赖
			1. 回到 Datakit 代码目录，执行根目录下的  ./mkdocs.sh  即可
			`)

		txt = "### `user_agent()` {#fn-user-agent}"
		arr, d = doMatch(txt, true)
		assert.Empty(t, arr)
		assert.Equal(t, txt, d)
	})

	t.Run("section", func(t *T.T) {
		var (
			arr    []string
			txt, d string
		)

		txt = "### `hahah` {#h}"
		arr, d = doMatch(txt, true)
		assert.Empty(t, arr)
		assert.Equal(t, txt, d)

		txt = "# haha" // head section NOT checked
		arr, d = doMatch(txt, true)
		assert.Empty(t, arr)
		assert.Equal(t, txt, d)

		txt = "## haha" // head section NOT checked
		arr, d = doMatch(txt, true)
		assert.NotEmpty(t, arr)
		assert.Equal(t, txt, d) // can't fix invalid section
	})
}
