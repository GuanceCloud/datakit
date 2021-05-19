package elb

type LoadbalancerV1 struct {
	VipAddress      string `json:"vip_address"`
	UpdateTime      string `json:"update_time"`
	CreateTime      string `json:"create_time"`
	ID              string `json:"id"`
	Status          string `json:"status"`
	Bandwidth       int    `json:"bandwidth"`
	SecurityGroupID string `json:"security_group_id"`
	VpcID           string `json:"vpc_id"`
	AdminStateUp    int    `json:"admin_state_up"`
	VipSubnetID     string `json:"vip_subnet_id"`
	Type            string `json:"type"`
	Name            string `json:"name"`
	Description     string `json:"description"`
}

type ListLoadbalancersV1 struct {
	Loadbalancers []LoadbalancerV1 `json:"loadbalancers"`
	InstanceNum   string           `json:"instance_num"`
}
