package linode

type AccountSwitchParams struct {
	AccountName string `json:"account_name" jsonschema:"required,description=Name of the account to switch to"`
}

type AccountInfo struct {
	Name     string `json:"name"`
	Label    string `json:"label"`
	IsCurrent bool   `json:"is_current"`
}

type AccountListResult struct {
	Accounts []AccountInfo `json:"accounts"`
	Current  string        `json:"current_account"`
}

type InstancesListResult struct {
	Instances []InstanceSummary `json:"instances"`
	Count     int               `json:"count"`
}

type InstanceSummary struct {
	ID       int      `json:"id"`
	Label    string   `json:"label"`
	Status   string   `json:"status"`
	Region   string   `json:"region"`
	Type     string   `json:"type"`
	IPv4     []string `json:"ipv4"`
	IPv6     string   `json:"ipv6"`
	Created  string   `json:"created"`
	Updated  string   `json:"updated"`
}