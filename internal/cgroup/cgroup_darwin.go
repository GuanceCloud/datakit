package cgroup

type Cgroup struct {
	opt *CgroupOptions
	err error
}

func (c *Cgroup) start() {
	l.Infof("not support darwin system, exit")
}
