package sqlserver

import (
	T "testing"

	dt "github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/assert"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/point"
)

type caseSpec struct {
	t *T.T

	name    string
	repo    string
	repoTag string
	envs    []string

	conf      []byte
	expectPts []*point.Point

	saPassword string

	pool     *dockertest.Pool
	resource *dockertest.Resource

	// TODO: test-result
}

func (cs *caseSpec) run() {
	p, err := dt.NewPool()
	assert.NoError(t, err)

	r, err := p.RunWithOptions(&dt.RunOptions{
		Repository: cs.repo,
		Tag:        cs.repoTag,
		PortBindings: map[docker.Port][]docker.PortBinding{
			"1433/tcp": []docker.PortBinding{{HostIP: "0.0.0.0", HostPort: "1433"}},
		},

		Name: "container-of-" + cs.name,
		Env:  cs.envs,
	}, func(c *docker.HostConfig) {
		c.AutoRemove = true
		c.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})

	assert.NoError(t, err)

	cs.pool = p
	cs.resource = r

	// TODO: run the test
}

func buildCases(t *T.T) []*caseSpec {
	t.Helper()

	basicSpecs := []struct {
		title         string
		tomlPath      string
		expectPtsPath string
	}{
		{
			name:      "remote-sqlserver",
			conf:      []byte(`some-conf`),
			expectPts: nil,
		},

		{
			name:      "remote-sqlserver-with-extra-tags",
			conf:      []byte(`some-conf`),
			expectPts: nil,
		},
	}

	images := [][2]string{
		[2]string{"mcr.microsoft.com/mssql/server", "2017-latest"},
		[2]string{"mcr.microsoft.com/mssql/server", "2019-latest"},
		[2]string{"mcr.microsoft.com/mssql/server", "2022-latest"},
	}

	// TODO: add per-image configs
	perImageCfgs := []interface{}{}
	_ = perImageCfgs

	for _, img := range images {
		for _, c := range basicSpecs {
			cases = append(cases, &caseSpec{
				t:    t,
				name: c.name,

				repo:    img[0],
				repoTag: img[1],

				conf:      c.conf,
				expectPts: c.expectPts,
			})
		}
	}
	return
}

func TestInput(t *T.T) {
	cases := buildCases(t)

	for _, tc := range cases {
		t.Run(tc.name, func(t *T.T) {
			tc.run()
		})
	}
}
