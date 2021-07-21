package elb

type Listener struct {
	ID string `json:"id"`
}

type Pool struct {
	ID string `json:"id"`
}
type LoadbalancerV2 struct {
	Description         string     `json:"description"`
	AdminStateUp        bool       `json:"admin_state_up"`
	TenantID            string     `json:"tenant_id"`
	ProjectID           string     `json:"project_id"`
	ProvisioningStatus  string     `json:"provisioning_status"`
	VipSubnetID         string     `json:"vip_subnet_id"`
	Listeners           []Listener `json:"listeners"`
	VipAddress          string     `json:"vip_address"`
	VipPortID           string     `json:"vip_port_id"`
	Provider            string     `json:"provider"`
	Pools               []Pool     `json:"pools"`
	ID                  string     `json:"id"`
	OperatingStatus     string     `json:"operating_status"`
	Tags                []string   `json:"tags"`
	Name                string     `json:"name"`
	CreatedAt           string     `json:"created_at"`
	UpdatedAt           string     `json:"updated_at"`
	EnterpriseProjectID string     `json:"enterprise_project_id "`
}

type ListLoadbalancersV2 struct {
	Loadbalancers []LoadbalancerV2 `json:"loadbalancers"`
}

type LoadbalancerLink struct {
	Href string `json:"href"`
	Rel  string `json:"rel"`
}

type ListLoadbalancersV20 struct {
	Loadbalancers     []LoadbalancerV2   `json:"loadbalancers"`
	LoadbalancerLinks []LoadbalancerLink `json:"loadbalancers_links"`
}
