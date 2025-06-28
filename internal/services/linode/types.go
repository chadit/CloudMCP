package linode

type AccountSwitchParams struct {
	AccountName string `json:"account_name" jsonschema:"required,description=Name of the account to switch to"`
}

type AccountInfo struct {
	Name      string `json:"name"`
	Label     string `json:"label"`
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
	ID      int      `json:"id"`
	Label   string   `json:"label"`
	Status  string   `json:"status"`
	Region  string   `json:"region"`
	Type    string   `json:"type"`
	IPv4    []string `json:"ipv4"`
	IPv6    string   `json:"ipv6"`
	Created string   `json:"created"`
	Updated string   `json:"updated"`
}

// Instance Get.
type InstanceGetParams struct {
	InstanceID int `json:"instance_id" jsonschema:"required,description=ID of the Linode instance"`
}

type InstanceDetail struct {
	ID              int             `json:"id"`
	Label           string          `json:"label"`
	Status          string          `json:"status"`
	Region          string          `json:"region"`
	Type            string          `json:"type"`
	Image           string          `json:"image"`
	Group           string          `json:"group"`
	Tags            []string        `json:"tags"`
	IPv4            []string        `json:"ipv4"`
	IPv6            string          `json:"ipv6"`
	Created         string          `json:"created"`
	Updated         string          `json:"updated"`
	Hypervisor      string          `json:"hypervisor"`
	Specs           InstanceSpecs   `json:"specs"`
	Alerts          InstanceAlerts  `json:"alerts"`
	Backups         InstanceBackups `json:"backups"`
	WatchdogEnabled bool            `json:"watchdog_enabled"`
}

type InstanceSpecs struct {
	Disk     int `json:"disk"`
	Memory   int `json:"memory"`
	VCPUs    int `json:"vcpus"`
	GPUs     int `json:"gpus"`
	Transfer int `json:"transfer"`
}

type InstanceAlerts struct {
	CPU           int `json:"cpu"`
	NetworkIn     int `json:"network_in"`
	NetworkOut    int `json:"network_out"`
	TransferQuota int `json:"transfer_quota"`
	IO            int `json:"io"`
}

type InstanceBackups struct {
	Enabled        bool           `json:"enabled"`
	Schedule       BackupSchedule `json:"schedule"`
	LastSuccessful string         `json:"last_successful"`
	Available      bool           `json:"available"`
}

type BackupSchedule struct {
	Day    string `json:"day"`
	Window string `json:"window"`
}

// Volume types.
type VolumesListResult struct {
	Volumes []VolumeSummary `json:"volumes"`
	Count   int             `json:"count"`
}

type VolumeSummary struct {
	ID             int      `json:"id"`
	Label          string   `json:"label"`
	Status         string   `json:"status"`
	Size           int      `json:"size"`
	Region         string   `json:"region"`
	LinodeID       *int     `json:"linode_id"`
	LinodeLabel    string   `json:"linode_label"`
	Created        string   `json:"created"`
	Updated        string   `json:"updated"`
	FilesystemPath string   `json:"filesystem_path"`
	Tags           []string `json:"tags"`
}

type VolumeGetParams struct {
	VolumeID int `json:"volume_id" jsonschema:"required,description=ID of the volume"`
}

// IP types.
type IPsListResult struct {
	IPs   []IPInfo `json:"ips"`
	Count int      `json:"count"`
}

type IPInfo struct {
	Address    string `json:"address"`
	Gateway    string `json:"gateway"`
	SubnetMask string `json:"subnet_mask"`
	Prefix     int    `json:"prefix"`
	Type       string `json:"type"`
	Public     bool   `json:"public"`
	RDNS       string `json:"rdns"`
	LinodeID   int    `json:"linode_id"`
	Region     string `json:"region"`
}

type IPGetParams struct {
	Address string `json:"address" jsonschema:"required,description=IP address to get details for"`
}

// Instance operations.
type InstanceCreateParams struct {
	Region         string   `json:"region" jsonschema:"required,description=Region ID where the instance will be created"`
	Type           string   `json:"type" jsonschema:"required,description=Linode type ID (e.g. g6-nanode-1)"`
	Label          string   `json:"label" jsonschema:"required,description=Display label for the instance"`
	Image          string   `json:"image" jsonschema:"description=Image ID to deploy (e.g. linode/ubuntu22.04)"`
	RootPass       string   `json:"root_pass" jsonschema:"description=Root password for the instance"`
	AuthorizedKeys []string `json:"authorized_keys" jsonschema:"description=SSH public keys to add to root user"`
	StackscriptID  int      `json:"stackscript_id" jsonschema:"description=StackScript ID to run on first boot"`
	BackupsEnabled bool     `json:"backups_enabled" jsonschema:"description=Enable automatic backups"`
	PrivateIP      bool     `json:"private_ip" jsonschema:"description=Add a private IP address"`
	Tags           []string `json:"tags" jsonschema:"description=Tags to apply to the instance"`
}

type InstanceDeleteParams struct {
	InstanceID int `json:"instance_id" jsonschema:"required,description=ID of the Linode instance to delete"`
}

type InstanceBootParams struct {
	InstanceID int `json:"instance_id" jsonschema:"required,description=ID of the Linode instance to boot"`
	ConfigID   int `json:"config_id" jsonschema:"description=Configuration profile ID to boot"`
}

type InstanceShutdownParams struct {
	InstanceID int `json:"instance_id" jsonschema:"required,description=ID of the Linode instance to shutdown"`
}

type InstanceRebootParams struct {
	InstanceID int `json:"instance_id" jsonschema:"required,description=ID of the Linode instance to reboot"`
	ConfigID   int `json:"config_id" jsonschema:"description=Configuration profile ID to reboot into"`
}

// Volume operations.
type VolumeCreateParams struct {
	Label    string   `json:"label" jsonschema:"required,description=Display label for the volume"`
	Size     int      `json:"size" jsonschema:"required,description=Size of the volume in GB (10-8192)"`
	Region   string   `json:"region" jsonschema:"description=Region ID where the volume will be created"`
	LinodeID int      `json:"linode_id" jsonschema:"description=ID of Linode to attach the volume to"`
	Tags     []string `json:"tags" jsonschema:"description=Tags to apply to the volume"`
}

type VolumeDeleteParams struct {
	VolumeID int `json:"volume_id" jsonschema:"required,description=ID of the volume to delete"`
}

type VolumeAttachParams struct {
	VolumeID           int  `json:"volume_id" jsonschema:"required,description=ID of the volume to attach"`
	LinodeID           int  `json:"linode_id" jsonschema:"required,description=ID of the Linode to attach to"`
	PersistAcrossBoots bool `json:"persist_across_boots" jsonschema:"description=Keep volume attached when Linode reboots"`
}

type VolumeDetachParams struct {
	VolumeID int `json:"volume_id" jsonschema:"required,description=ID of the volume to detach"`
}

// Image types.
type ImagesListParams struct {
	IsPublic *bool `json:"is_public" jsonschema:"description=Filter to only public (true) or private (false) images"`
}

type ImagesListResult struct {
	Images []ImageSummary `json:"images"`
	Count  int            `json:"count"`
}

type ImageSummary struct {
	ID           string        `json:"id"`
	Label        string        `json:"label"`
	Description  string        `json:"description"`
	Created      string        `json:"created"`
	CreatedBy    string        `json:"created_by"`
	Deprecated   bool          `json:"deprecated"`
	IsPublic     bool          `json:"is_public"`
	Size         int           `json:"size"`
	Type         string        `json:"type"`
	Vendor       string        `json:"vendor"`
	Status       string        `json:"status"`
	Regions      []ImageRegion `json:"regions"`
	Tags         []string      `json:"tags"`
	TotalSize    int           `json:"total_size"`
	Capabilities []string      `json:"capabilities"`
}

type ImageRegion struct {
	Region string `json:"region"`
	Status string `json:"status"`
}

type ImageGetParams struct {
	ImageID string `json:"image_id" jsonschema:"required,description=ID of the image (e.g. linode/ubuntu22.04 or private/12345)"`
}

type ImageDetail struct {
	ID           string        `json:"id"`
	Label        string        `json:"label"`
	Description  string        `json:"description"`
	Created      string        `json:"created"`
	CreatedBy    string        `json:"created_by"`
	Deprecated   bool          `json:"deprecated"`
	IsPublic     bool          `json:"is_public"`
	Size         int           `json:"size"`
	Type         string        `json:"type"`
	Vendor       string        `json:"vendor"`
	Status       string        `json:"status"`
	Regions      []ImageRegion `json:"regions"`
	Tags         []string      `json:"tags"`
	TotalSize    int           `json:"total_size"`
	Capabilities []string      `json:"capabilities"`
	Updated      string        `json:"updated"`
	Expiry       *string       `json:"expiry"`
}

// Image operations.
type ImageCreateParams struct {
	DiskID      int      `json:"disk_id" jsonschema:"required,description=ID of the Linode disk to create image from"`
	Label       string   `json:"label" jsonschema:"required,description=Display label for the image"`
	Description string   `json:"description" jsonschema:"description=Detailed description of the image"`
	CloudInit   bool     `json:"cloud_init" jsonschema:"description=Whether this image supports cloud-init"`
	Tags        []string `json:"tags" jsonschema:"description=Tags to apply to the image"`
}

type ImageUpdateParams struct {
	ImageID     string   `json:"image_id" jsonschema:"required,description=ID of the image to update"`
	Label       string   `json:"label" jsonschema:"description=New display label for the image"`
	Description string   `json:"description" jsonschema:"description=New description for the image"`
	Tags        []string `json:"tags" jsonschema:"description=New tags for the image (replaces existing tags)"`
}

type ImageDeleteParams struct {
	ImageID string `json:"image_id" jsonschema:"required,description=ID of the image to delete"`
}

type ImageReplicateParams struct {
	ImageID string   `json:"image_id" jsonschema:"required,description=ID of the image to replicate"`
	Regions []string `json:"regions" jsonschema:"required,description=List of region IDs to replicate the image to"`
}

type ImageUploadParams struct {
	Label       string   `json:"label" jsonschema:"required,description=Display label for the uploaded image"`
	Region      string   `json:"region" jsonschema:"required,description=Initial region for the uploaded image"`
	Description string   `json:"description" jsonschema:"description=Description of the uploaded image"`
	CloudInit   bool     `json:"cloud_init" jsonschema:"description=Whether this image supports cloud-init"`
	Tags        []string `json:"tags" jsonschema:"description=Tags to apply to the image"`
}

type ImageUploadResult struct {
	ImageID  string `json:"image_id"`
	UploadTo string `json:"upload_to"`
}

// Firewall types.
type FirewallsListResult struct {
	Firewalls []FirewallSummary `json:"firewalls"`
	Count     int               `json:"count"`
}

type FirewallSummary struct {
	ID      int              `json:"id"`
	Label   string           `json:"label"`
	Status  string           `json:"status"`
	Tags    []string         `json:"tags"`
	Rules   FirewallRuleSet  `json:"rules"`
	Devices []FirewallDevice `json:"devices"`
	Created string           `json:"created"`
	Updated string           `json:"updated"`
}

type FirewallGetParams struct {
	FirewallID int `json:"firewall_id" jsonschema:"required,description=ID of the firewall"`
}

type FirewallDetail struct {
	ID      int              `json:"id"`
	Label   string           `json:"label"`
	Status  string           `json:"status"`
	Tags    []string         `json:"tags"`
	Rules   FirewallRuleSet  `json:"rules"`
	Devices []FirewallDevice `json:"devices"`
	Created string           `json:"created"`
	Updated string           `json:"updated"`
}

type FirewallRuleSet struct {
	Inbound        []FirewallRule `json:"inbound"`
	InboundPolicy  string         `json:"inbound_policy"`
	Outbound       []FirewallRule `json:"outbound"`
	OutboundPolicy string         `json:"outbound_policy"`
}

type FirewallRule struct {
	Ports       string          `json:"ports"`
	Protocol    string          `json:"protocol"`
	Addresses   FirewallAddress `json:"addresses"`
	Action      string          `json:"action"`
	Label       string          `json:"label"`
	Description string          `json:"description"`
}

type FirewallAddress struct {
	IPv4 []string `json:"ipv4"`
	IPv6 []string `json:"ipv6"`
}

type FirewallDevice struct {
	ID      int    `json:"id"`
	Type    string `json:"type"`
	Label   string `json:"label"`
	URL     string `json:"url"`
	Created string `json:"created"`
}

type FirewallCreateParams struct {
	Label string          `json:"label" jsonschema:"required,description=Display label for the firewall"`
	Rules FirewallRuleSet `json:"rules" jsonschema:"description=Firewall rules configuration"`
	Tags  []string        `json:"tags" jsonschema:"description=Tags to apply to the firewall"`
}

type FirewallUpdateParams struct {
	FirewallID int      `json:"firewall_id" jsonschema:"required,description=ID of the firewall to update"`
	Label      string   `json:"label" jsonschema:"description=New display label for the firewall"`
	Tags       []string `json:"tags" jsonschema:"description=New tags for the firewall"`
}

type FirewallDeleteParams struct {
	FirewallID int `json:"firewall_id" jsonschema:"required,description=ID of the firewall to delete"`
}

type FirewallRulesUpdateParams struct {
	FirewallID int             `json:"firewall_id" jsonschema:"required,description=ID of the firewall to update rules for"`
	Rules      FirewallRuleSet `json:"rules" jsonschema:"required,description=New firewall rules configuration"`
}

type FirewallDeviceCreateParams struct {
	FirewallID int    `json:"firewall_id" jsonschema:"required,description=ID of the firewall"`
	DeviceID   int    `json:"device_id" jsonschema:"required,description=ID of the device to assign to firewall"`
	DeviceType string `json:"device_type" jsonschema:"required,description=Type of device (linode or nodebalancer)"`
}

type FirewallDeviceDeleteParams struct {
	FirewallID int `json:"firewall_id" jsonschema:"required,description=ID of the firewall"`
	DeviceID   int `json:"device_id" jsonschema:"required,description=ID of the device to remove from firewall"`
}

// NodeBalancer types.
type NodeBalancersListResult struct {
	NodeBalancers []NodeBalancerSummary `json:"nodebalancers"`
	Count         int                   `json:"count"`
}

type NodeBalancerSummary struct {
	ID                 int                  `json:"id"`
	Label              string               `json:"label"`
	Region             string               `json:"region"`
	IPv4               string               `json:"ipv4"`
	IPv6               *string              `json:"ipv6"`
	ClientConnThrottle int                  `json:"client_conn_throttle"`
	Hostname           string               `json:"hostname"`
	Tags               []string             `json:"tags"`
	Created            string               `json:"created"`
	Updated            string               `json:"updated"`
	Transfer           NodeBalancerTransfer `json:"transfer"`
}

type NodeBalancerTransfer struct {
	In    float64 `json:"in"`
	Out   float64 `json:"out"`
	Total float64 `json:"total"`
}

type NodeBalancerGetParams struct {
	NodeBalancerID int `json:"nodebalancer_id" jsonschema:"required,description=ID of the NodeBalancer"`
}

type NodeBalancerDetail struct {
	ID                 int                  `json:"id"`
	Label              string               `json:"label"`
	Region             string               `json:"region"`
	IPv4               string               `json:"ipv4"`
	IPv6               *string              `json:"ipv6"`
	ClientConnThrottle int                  `json:"client_conn_throttle"`
	Hostname           string               `json:"hostname"`
	Tags               []string             `json:"tags"`
	Created            string               `json:"created"`
	Updated            string               `json:"updated"`
	Transfer           NodeBalancerTransfer `json:"transfer"`
	Configs            []NodeBalancerConfig `json:"configs"`
}

type NodeBalancerConfig struct {
	ID             int                    `json:"id"`
	Port           int                    `json:"port"`
	Protocol       string                 `json:"protocol"`
	Algorithm      string                 `json:"algorithm"`
	Stickiness     string                 `json:"stickiness"`
	Check          string                 `json:"check"`
	CheckInterval  int                    `json:"check_interval"`
	CheckTimeout   int                    `json:"check_timeout"`
	CheckAttempts  int                    `json:"check_attempts"`
	CheckPath      string                 `json:"check_path"`
	CheckBody      string                 `json:"check_body"`
	CheckPassive   bool                   `json:"check_passive"`
	ProxyProtocol  string                 `json:"proxy_protocol"`
	CipherSuite    string                 `json:"cipher_suite"`
	SSLCommonName  string                 `json:"ssl_commonname"`
	SSLFingerprint string                 `json:"ssl_fingerprint"`
	SSLCert        string                 `json:"ssl_cert"`
	SSLKey         string                 `json:"ssl_key"`
	NodesStatus    NodeBalancerNodeStatus `json:"nodes_status"`
}

type NodeBalancerNodeStatus struct {
	Up   int `json:"up"`
	Down int `json:"down"`
}

type NodeBalancerCreateParams struct {
	Label              string   `json:"label" jsonschema:"required,description=Display label for the NodeBalancer"`
	Region             string   `json:"region" jsonschema:"required,description=Region ID where the NodeBalancer will be created"`
	ClientConnThrottle int      `json:"client_conn_throttle" jsonschema:"description=Throttle connections per second (0-20)"`
	Tags               []string `json:"tags" jsonschema:"description=Tags to apply to the NodeBalancer"`
}

type NodeBalancerUpdateParams struct {
	NodeBalancerID     int      `json:"nodebalancer_id" jsonschema:"required,description=ID of the NodeBalancer to update"`
	Label              string   `json:"label" jsonschema:"description=New display label for the NodeBalancer"`
	ClientConnThrottle int      `json:"client_conn_throttle" jsonschema:"description=New throttle connections per second (0-20)"`
	Tags               []string `json:"tags" jsonschema:"description=New tags for the NodeBalancer"`
}

type NodeBalancerDeleteParams struct {
	NodeBalancerID int `json:"nodebalancer_id" jsonschema:"required,description=ID of the NodeBalancer to delete"`
}

type NodeBalancerConfigCreateParams struct {
	NodeBalancerID int    `json:"nodebalancer_id" jsonschema:"required,description=ID of the NodeBalancer"`
	Port           int    `json:"port" jsonschema:"required,description=Port to configure (1-65534)"`
	Protocol       string `json:"protocol" jsonschema:"required,description=Protocol (http, https, tcp)"`
	Algorithm      string `json:"algorithm" jsonschema:"description=Balancing algorithm (roundrobin, leastconn, source)"`
	Stickiness     string `json:"stickiness" jsonschema:"description=Session stickiness (none, table, http_cookie)"`
	Check          string `json:"check" jsonschema:"description=Health check type (none, connection, http, http_body)"`
	CheckInterval  int    `json:"check_interval" jsonschema:"description=Health check interval in seconds"`
	CheckTimeout   int    `json:"check_timeout" jsonschema:"description=Health check timeout in seconds"`
	CheckAttempts  int    `json:"check_attempts" jsonschema:"description=Health check attempts before marking down"`
	CheckPath      string `json:"check_path" jsonschema:"description=HTTP health check path"`
	CheckBody      string `json:"check_body" jsonschema:"description=Expected response body for http_body check"`
	CheckPassive   bool   `json:"check_passive" jsonschema:"description=Enable passive health checks"`
	ProxyProtocol  string `json:"proxy_protocol" jsonschema:"description=Proxy protocol version (none, v1, v2)"`
	SSLCert        string `json:"ssl_cert" jsonschema:"description=SSL certificate for HTTPS"`
	SSLKey         string `json:"ssl_key" jsonschema:"description=SSL private key for HTTPS"`
}

type NodeBalancerConfigUpdateParams struct {
	NodeBalancerID int    `json:"nodebalancer_id" jsonschema:"required,description=ID of the NodeBalancer"`
	ConfigID       int    `json:"config_id" jsonschema:"required,description=ID of the configuration to update"`
	Port           int    `json:"port" jsonschema:"description=Port to configure (1-65534)"`
	Protocol       string `json:"protocol" jsonschema:"description=Protocol (http, https, tcp)"`
	Algorithm      string `json:"algorithm" jsonschema:"description=Balancing algorithm (roundrobin, leastconn, source)"`
	Stickiness     string `json:"stickiness" jsonschema:"description=Session stickiness (none, table, http_cookie)"`
	Check          string `json:"check" jsonschema:"description=Health check type (none, connection, http, http_body)"`
	CheckInterval  int    `json:"check_interval" jsonschema:"description=Health check interval in seconds"`
	CheckTimeout   int    `json:"check_timeout" jsonschema:"description=Health check timeout in seconds"`
	CheckAttempts  int    `json:"check_attempts" jsonschema:"description=Health check attempts before marking down"`
	CheckPath      string `json:"check_path" jsonschema:"description=HTTP health check path"`
	CheckBody      string `json:"check_body" jsonschema:"description=Expected response body for http_body check"`
	CheckPassive   bool   `json:"check_passive" jsonschema:"description=Enable passive health checks"`
	ProxyProtocol  string `json:"proxy_protocol" jsonschema:"description=Proxy protocol version (none, v1, v2)"`
	SSLCert        string `json:"ssl_cert" jsonschema:"description=SSL certificate for HTTPS"`
	SSLKey         string `json:"ssl_key" jsonschema:"description=SSL private key for HTTPS"`
}

type NodeBalancerConfigDeleteParams struct {
	NodeBalancerID int `json:"nodebalancer_id" jsonschema:"required,description=ID of the NodeBalancer"`
	ConfigID       int `json:"config_id" jsonschema:"required,description=ID of the configuration to delete"`
}

// Domain types.
type DomainsListResult struct {
	Domains []DomainSummary `json:"domains"`
	Count   int             `json:"count"`
}

type DomainSummary struct {
	ID          int      `json:"id"`
	Domain      string   `json:"domain"`
	Type        string   `json:"type"`
	Status      string   `json:"status"`
	Description string   `json:"description"`
	SOAEmail    string   `json:"soa_email"`
	RetrySec    int      `json:"retry_sec"`
	MasterIPs   []string `json:"master_ips"`
	AXfrIPs     []string `json:"axfr_ips"`
	Tags        []string `json:"tags"`
	Created     string   `json:"created"`
	Updated     string   `json:"updated"`
}

type DomainGetParams struct {
	DomainID int `json:"domain_id" jsonschema:"required,description=ID of the domain"`
}

type DomainDetail struct {
	ID          int      `json:"id"`
	Domain      string   `json:"domain"`
	Type        string   `json:"type"`
	Status      string   `json:"status"`
	Description string   `json:"description"`
	SOAEmail    string   `json:"soa_email"`
	RetrySec    int      `json:"retry_sec"`
	MasterIPs   []string `json:"master_ips"`
	AXfrIPs     []string `json:"axfr_ips"`
	Tags        []string `json:"tags"`
	Created     string   `json:"created"`
	Updated     string   `json:"updated"`
	ExpireSec   int      `json:"expire_sec"`
	RefreshSec  int      `json:"refresh_sec"`
	TTLSec      int      `json:"ttl_sec"`
	Group       string   `json:"group"`
}

type DomainCreateParams struct {
	Domain      string   `json:"domain" jsonschema:"required,description=Domain name (e.g. example.com)"`
	Type        string   `json:"type" jsonschema:"required,description=Domain type (master or slave)"`
	SOAEmail    string   `json:"soa_email" jsonschema:"description=SOA email address"`
	Description string   `json:"description" jsonschema:"description=Description of the domain"`
	RetrySec    int      `json:"retry_sec" jsonschema:"description=Retry interval in seconds"`
	MasterIPs   []string `json:"master_ips" jsonschema:"description=Master IPs for slave domains"`
	AXfrIPs     []string `json:"axfr_ips" jsonschema:"description=IPs allowed to AXFR the entire zone"`
	ExpireSec   int      `json:"expire_sec" jsonschema:"description=Expire time in seconds"`
	RefreshSec  int      `json:"refresh_sec" jsonschema:"description=Refresh time in seconds"`
	TTLSec      int      `json:"ttl_sec" jsonschema:"description=Default TTL in seconds"`
	Tags        []string `json:"tags" jsonschema:"description=Tags to apply to the domain"`
	Group       string   `json:"group" jsonschema:"description=Group for the domain"`
}

type DomainUpdateParams struct {
	DomainID    int      `json:"domain_id" jsonschema:"required,description=ID of the domain to update"`
	Domain      string   `json:"domain" jsonschema:"description=New domain name"`
	Type        string   `json:"type" jsonschema:"description=New domain type (master or slave)"`
	SOAEmail    string   `json:"soa_email" jsonschema:"description=New SOA email address"`
	Description string   `json:"description" jsonschema:"description=New description"`
	RetrySec    int      `json:"retry_sec" jsonschema:"description=New retry interval in seconds"`
	MasterIPs   []string `json:"master_ips" jsonschema:"description=New master IPs for slave domains"`
	AXfrIPs     []string `json:"axfr_ips" jsonschema:"description=New IPs allowed to AXFR"`
	ExpireSec   int      `json:"expire_sec" jsonschema:"description=New expire time in seconds"`
	RefreshSec  int      `json:"refresh_sec" jsonschema:"description=New refresh time in seconds"`
	TTLSec      int      `json:"ttl_sec" jsonschema:"description=New default TTL in seconds"`
	Tags        []string `json:"tags" jsonschema:"description=New tags for the domain"`
	Group       string   `json:"group" jsonschema:"description=New group for the domain"`
}

type DomainDeleteParams struct {
	DomainID int `json:"domain_id" jsonschema:"required,description=ID of the domain to delete"`
}

// Domain Record types.
type DomainRecordsListParams struct {
	DomainID int `json:"domain_id" jsonschema:"required,description=ID of the domain"`
}

type DomainRecordsListResult struct {
	Records []DomainRecord `json:"records"`
	Count   int            `json:"count"`
}

type DomainRecord struct {
	ID       int    `json:"id"`
	Type     string `json:"type"`
	Name     string `json:"name"`
	Target   string `json:"target"`
	Priority int    `json:"priority"`
	Weight   int    `json:"weight"`
	Port     int    `json:"port"`
	Service  string `json:"service"`
	Protocol string `json:"protocol"`
	TTLSec   int    `json:"ttl_sec"`
	Tag      string `json:"tag"`
	Created  string `json:"created"`
	Updated  string `json:"updated"`
}

type DomainRecordGetParams struct {
	DomainID int `json:"domain_id" jsonschema:"required,description=ID of the domain"`
	RecordID int `json:"record_id" jsonschema:"required,description=ID of the domain record"`
}

type DomainRecordCreateParams struct {
	DomainID int    `json:"domain_id" jsonschema:"required,description=ID of the domain"`
	Type     string `json:"type" jsonschema:"required,description=Record type (A, AAAA, CNAME, MX, TXT, SRV, PTR, CAA, NS)"`
	Name     string `json:"name" jsonschema:"description=Record name (subdomain)"`
	Target   string `json:"target" jsonschema:"required,description=Record target (IP, hostname, etc.)"`
	Priority int    `json:"priority" jsonschema:"description=Record priority (for MX and SRV records)"`
	Weight   int    `json:"weight" jsonschema:"description=Record weight (for SRV records)"`
	Port     int    `json:"port" jsonschema:"description=Record port (for SRV records)"`
	Service  string `json:"service" jsonschema:"description=Service name (for SRV records)"`
	Protocol string `json:"protocol" jsonschema:"description=Protocol name (for SRV records)"`
	TTLSec   int    `json:"ttl_sec" jsonschema:"description=TTL in seconds"`
	Tag      string `json:"tag" jsonschema:"description=CAA record tag"`
}

type DomainRecordUpdateParams struct {
	DomainID int    `json:"domain_id" jsonschema:"required,description=ID of the domain"`
	RecordID int    `json:"record_id" jsonschema:"required,description=ID of the record to update"`
	Type     string `json:"type" jsonschema:"description=New record type"`
	Name     string `json:"name" jsonschema:"description=New record name"`
	Target   string `json:"target" jsonschema:"description=New record target"`
	Priority int    `json:"priority" jsonschema:"description=New record priority"`
	Weight   int    `json:"weight" jsonschema:"description=New record weight"`
	Port     int    `json:"port" jsonschema:"description=New record port"`
	Service  string `json:"service" jsonschema:"description=New service name"`
	Protocol string `json:"protocol" jsonschema:"description=New protocol name"`
	TTLSec   int    `json:"ttl_sec" jsonschema:"description=New TTL in seconds"`
	Tag      string `json:"tag" jsonschema:"description=New CAA record tag"`
}

type DomainRecordDeleteParams struct {
	DomainID int `json:"domain_id" jsonschema:"required,description=ID of the domain"`
	RecordID int `json:"record_id" jsonschema:"required,description=ID of the record to delete"`
}

// StackScript types.
type StackScriptsListResult struct {
	StackScripts []StackScriptSummary `json:"stackscripts"`
	Count        int                  `json:"count"`
}

type StackScriptSummary struct {
	ID                int      `json:"id"`
	Username          string   `json:"username"`
	Label             string   `json:"label"`
	Description       string   `json:"description"`
	IsPublic          bool     `json:"is_public"`
	Images            []string `json:"images"`
	DeploymentsTotal  int      `json:"deployments_total"`
	DeploymentsActive int      `json:"deployments_active"`
	UserGravatarID    string   `json:"user_gravatar_id"`
	Created           string   `json:"created"`
	Updated           string   `json:"updated"`
}

type StackScriptGetParams struct {
	StackScriptID int `json:"stackscript_id" jsonschema:"required,description=ID of the StackScript"`
}

type StackScriptDetail struct {
	ID                int                           `json:"id"`
	Username          string                        `json:"username"`
	Label             string                        `json:"label"`
	Description       string                        `json:"description"`
	Ordinal           int                           `json:"ordinal"`
	LogoURL           string                        `json:"logo_url"`
	Images            []string                      `json:"images"`
	DeploymentsTotal  int                           `json:"deployments_total"`
	DeploymentsActive int                           `json:"deployments_active"`
	IsPublic          bool                          `json:"is_public"`
	Mine              bool                          `json:"mine"`
	Created           string                        `json:"created"`
	Updated           string                        `json:"updated"`
	RevNote           string                        `json:"rev_note"`
	Script            string                        `json:"script"`
	UserDefinedFields []StackScriptUserDefinedField `json:"user_defined_fields"`
	UserGravatarID    string                        `json:"user_gravatar_id"`
}

type StackScriptUserDefinedField struct {
	Name    string `json:"name"`
	Label   string `json:"label"`
	Example string `json:"example"`
	OneOf   string `json:"oneof"`
	ManyOf  string `json:"manyof"`
	Default string `json:"default"`
}

type StackScriptCreateParams struct {
	Label       string   `json:"label" jsonschema:"required,description=Display label for the StackScript"`
	Description string   `json:"description" jsonschema:"description=Description of the StackScript"`
	Images      []string `json:"images" jsonschema:"required,description=Compatible image IDs"`
	Script      string   `json:"script" jsonschema:"required,description=StackScript code"`
	IsPublic    bool     `json:"is_public" jsonschema:"description=Whether the StackScript should be public"`
	RevNote     string   `json:"rev_note" jsonschema:"description=Revision note for this version"`
}

type StackScriptUpdateParams struct {
	StackScriptID int      `json:"stackscript_id" jsonschema:"required,description=ID of the StackScript to update"`
	Label         string   `json:"label" jsonschema:"description=New display label"`
	Description   string   `json:"description" jsonschema:"description=New description"`
	Images        []string `json:"images" jsonschema:"description=New compatible image IDs"`
	Script        string   `json:"script" jsonschema:"description=New StackScript code"`
	IsPublic      bool     `json:"is_public" jsonschema:"description=New public status"`
	RevNote       string   `json:"rev_note" jsonschema:"description=Revision note for this version"`
}

type StackScriptDeleteParams struct {
	StackScriptID int `json:"stackscript_id" jsonschema:"required,description=ID of the StackScript to delete"`
}

// LKE (Kubernetes) types.
type LKEClustersListResult struct {
	Clusters []LKEClusterSummary `json:"clusters"`
	Count    int                 `json:"count"`
}

type LKEClusterSummary struct {
	ID           int             `json:"id"`
	Label        string          `json:"label"`
	Region       string          `json:"region"`
	Status       string          `json:"status"`
	K8sVersion   string          `json:"k8s_version"`
	Tags         []string        `json:"tags"`
	Created      string          `json:"created"`
	Updated      string          `json:"updated"`
	ControlPlane LKEControlPlane `json:"control_plane"`
}

type LKEControlPlane struct {
	HighAvailability bool `json:"high_availability"`
}

type LKEClusterGetParams struct {
	ClusterID int `json:"cluster_id" jsonschema:"required,description=ID of the LKE cluster"`
}

type LKEClusterDetail struct {
	ID           int             `json:"id"`
	Label        string          `json:"label"`
	Region       string          `json:"region"`
	Status       string          `json:"status"`
	K8sVersion   string          `json:"k8s_version"`
	Tags         []string        `json:"tags"`
	Created      string          `json:"created"`
	Updated      string          `json:"updated"`
	ControlPlane LKEControlPlane `json:"control_plane"`
	NodePools    []LKENodePool   `json:"node_pools"`
}

type LKENodePool struct {
	ID         int                   `json:"id"`
	Count      int                   `json:"count"`
	Type       string                `json:"type"`
	Disks      []LKENodePoolDisk     `json:"disks"`
	Nodes      []LKENode             `json:"nodes"`
	Autoscaler LKENodePoolAutoscaler `json:"autoscaler"`
	Tags       []string              `json:"tags"`
}

type LKENodePoolDisk struct {
	Size int    `json:"size"`
	Type string `json:"type"`
}

type LKENode struct {
	ID         string `json:"id"`
	InstanceID int    `json:"instance_id"`
	Status     string `json:"status"`
}

type LKENodePoolAutoscaler struct {
	Enabled bool `json:"enabled"`
	Min     int  `json:"min"`
	Max     int  `json:"max"`
}

type LKEClusterCreateParams struct {
	Label        string              `json:"label" jsonschema:"required,description=Display label for the cluster"`
	Region       string              `json:"region" jsonschema:"required,description=Region ID where the cluster will be created"`
	K8sVersion   string              `json:"k8s_version" jsonschema:"required,description=Kubernetes version (e.g. 1.28)"`
	Tags         []string            `json:"tags" jsonschema:"description=Tags to apply to the cluster"`
	NodePools    []LKENodePoolSpec   `json:"node_pools" jsonschema:"required,description=Node pool specifications"`
	ControlPlane LKEControlPlaneSpec `json:"control_plane" jsonschema:"description=Control plane configuration"`
}

type LKENodePoolSpec struct {
	Type       string                 `json:"type" jsonschema:"required,description=Linode type for nodes (e.g. g6-standard-2)"`
	Count      int                    `json:"count" jsonschema:"required,description=Number of nodes in pool"`
	Disks      []LKENodePoolDisk      `json:"disks" jsonschema:"description=Disk configuration for nodes"`
	Autoscaler *LKENodePoolAutoscaler `json:"autoscaler" jsonschema:"description=Autoscaler configuration"`
	Tags       []string               `json:"tags" jsonschema:"description=Tags for the node pool"`
}

type LKEControlPlaneSpec struct {
	HighAvailability bool `json:"high_availability" jsonschema:"description=Enable high availability control plane"`
}

type LKEClusterUpdateParams struct {
	ClusterID    int                 `json:"cluster_id" jsonschema:"required,description=ID of the cluster to update"`
	Label        string              `json:"label" jsonschema:"description=New display label"`
	K8sVersion   string              `json:"k8s_version" jsonschema:"description=New Kubernetes version"`
	Tags         []string            `json:"tags" jsonschema:"description=New tags for the cluster"`
	ControlPlane LKEControlPlaneSpec `json:"control_plane" jsonschema:"description=New control plane configuration"`
}

type LKEClusterDeleteParams struct {
	ClusterID int `json:"cluster_id" jsonschema:"required,description=ID of the cluster to delete"`
}

type LKENodePoolCreateParams struct {
	ClusterID  int                    `json:"cluster_id" jsonschema:"required,description=ID of the cluster"`
	Type       string                 `json:"type" jsonschema:"required,description=Linode type for nodes"`
	Count      int                    `json:"count" jsonschema:"required,description=Number of nodes in pool"`
	Disks      []LKENodePoolDisk      `json:"disks" jsonschema:"description=Disk configuration for nodes"`
	Autoscaler *LKENodePoolAutoscaler `json:"autoscaler" jsonschema:"description=Autoscaler configuration"`
	Tags       []string               `json:"tags" jsonschema:"description=Tags for the node pool"`
}

type LKENodePoolUpdateParams struct {
	ClusterID  int                    `json:"cluster_id" jsonschema:"required,description=ID of the cluster"`
	PoolID     int                    `json:"pool_id" jsonschema:"required,description=ID of the node pool to update"`
	Count      int                    `json:"count" jsonschema:"description=New number of nodes"`
	Autoscaler *LKENodePoolAutoscaler `json:"autoscaler" jsonschema:"description=New autoscaler configuration"`
	Tags       []string               `json:"tags" jsonschema:"description=New tags for the node pool"`
}

type LKENodePoolDeleteParams struct {
	ClusterID int `json:"cluster_id" jsonschema:"required,description=ID of the cluster"`
	PoolID    int `json:"pool_id" jsonschema:"required,description=ID of the node pool to delete"`
}

type LKEKubeconfigParams struct {
	ClusterID int `json:"cluster_id" jsonschema:"required,description=ID of the cluster"`
}

// Database types.
type DatabasesListResult struct {
	Databases []DatabaseSummary `json:"databases"`
	Count     int               `json:"count"`
}

type DatabaseSummary struct {
	ID            int      `json:"id"`
	Label         string   `json:"label"`
	Engine        string   `json:"engine"`
	Version       string   `json:"version"`
	Region        string   `json:"region"`
	Status        string   `json:"status"`
	ClusterSize   int      `json:"cluster_size"`
	Type          string   `json:"type"`
	Encrypted     bool     `json:"encrypted"`
	Allow         []string `json:"allow_list"`
	Port          int      `json:"port"`
	SSLConnection bool     `json:"ssl_connection"`
	Created       string   `json:"created"`
	Updated       string   `json:"updated"`
	Tags          []string `json:"tags"`
}

type DatabaseGetParams struct {
	DatabaseID int `json:"database_id" jsonschema:"required,description=ID of the database"`
}

type DatabaseDetail struct {
	ID            int             `json:"id"`
	Label         string          `json:"label"`
	Engine        string          `json:"engine"`
	Version       string          `json:"version"`
	Region        string          `json:"region"`
	Status        string          `json:"status"`
	ClusterSize   int             `json:"cluster_size"`
	Type          string          `json:"type"`
	Encrypted     bool            `json:"encrypted"`
	Allow         []string        `json:"allow_list"`
	Port          int             `json:"port"`
	SSLConnection bool            `json:"ssl_connection"`
	Created       string          `json:"created"`
	Updated       string          `json:"updated"`
	Tags          []string        `json:"tags"`
	Hosts         DatabaseHosts   `json:"hosts"`
	Updates       DatabaseUpdates `json:"updates"`
	Backups       DatabaseBackups `json:"backups"`
}

type DatabaseHosts struct {
	Primary   string `json:"primary"`
	Secondary string `json:"secondary"`
}

type DatabaseUpdates struct {
	Frequency   string `json:"frequency"`
	Duration    int    `json:"duration"`
	HourOfDay   int    `json:"hour_of_day"`
	DayOfWeek   int    `json:"day_of_week"`
	WeekOfMonth int    `json:"week_of_month"`
}

type DatabaseBackups struct {
	Enabled  bool                   `json:"enabled"`
	Schedule DatabaseBackupSchedule `json:"schedule"`
}

type DatabaseBackupSchedule struct {
	Day    string `json:"day"`
	Window string `json:"window"`
}

type DatabaseCreateParams struct {
	Label         string   `json:"label" jsonschema:"required,description=Display label for the database"`
	Engine        string   `json:"engine" jsonschema:"required,description=Database engine (mysql or postgresql)"`
	Version       string   `json:"version" jsonschema:"description=Database version"`
	Region        string   `json:"region" jsonschema:"required,description=Region ID where the database will be created"`
	Type          string   `json:"type" jsonschema:"required,description=Database plan type"`
	ClusterSize   int      `json:"cluster_size" jsonschema:"description=Number of nodes in cluster (1 or 3)"`
	Encrypted     bool     `json:"encrypted" jsonschema:"description=Enable encryption at rest"`
	SSLConnection bool     `json:"ssl_connection" jsonschema:"description=Require SSL connections"`
	Allow         []string `json:"allow_list" jsonschema:"description=IP addresses allowed to connect"`
	Tags          []string `json:"tags" jsonschema:"description=Tags to apply to the database"`
}

type DatabaseUpdateParams struct {
	DatabaseID int             `json:"database_id" jsonschema:"required,description=ID of the database to update"`
	Label      string          `json:"label" jsonschema:"description=New display label"`
	Allow      []string        `json:"allow_list" jsonschema:"description=New IP addresses allowed to connect"`
	Updates    DatabaseUpdates `json:"updates" jsonschema:"description=New maintenance window settings"`
	Tags       []string        `json:"tags" jsonschema:"description=New tags for the database"`
}

type DatabaseDeleteParams struct {
	DatabaseID int `json:"database_id" jsonschema:"required,description=ID of the database to delete"`
}

type DatabaseBackupCreateParams struct {
	DatabaseID int    `json:"database_id" jsonschema:"required,description=ID of the database"`
	Label      string `json:"label" jsonschema:"required,description=Label for the backup"`
	Target     string `json:"target" jsonschema:"description=Backup target (primary or secondary)"`
}

type DatabaseBackupRestoreParams struct {
	DatabaseID int `json:"database_id" jsonschema:"required,description=ID of the database"`
	BackupID   int `json:"backup_id" jsonschema:"required,description=ID of the backup to restore"`
}

type DatabaseCredentialsResetParams struct {
	DatabaseID int    `json:"database_id" jsonschema:"required,description=ID of the database"`
	Username   string `json:"username" jsonschema:"required,description=Username to reset password for"`
}

// Object Storage types.
type ObjectStorageBucketsListResult struct {
	Buckets []ObjectStorageBucketSummary `json:"buckets"`
	Count   int                          `json:"count"`
}

type ObjectStorageBucketSummary struct {
	Label    string `json:"label"`
	Cluster  string `json:"cluster"`
	Hostname string `json:"hostname"`
	Created  string `json:"created"`
	Size     int64  `json:"size"`
	Objects  int    `json:"objects"`
}

type ObjectStorageBucketGetParams struct {
	Cluster string `json:"cluster" jsonschema:"required,description=Object Storage cluster ID"`
	Bucket  string `json:"bucket" jsonschema:"required,description=Bucket name"`
}

type ObjectStorageBucketDetail struct {
	Label    string `json:"label"`
	Cluster  string `json:"cluster"`
	Hostname string `json:"hostname"`
	Created  string `json:"created"`
	Size     int64  `json:"size"`
	Objects  int    `json:"objects"`
}

type ObjectStorageBucketCreateParams struct {
	Label   string `json:"label" jsonschema:"required,description=Bucket name/label"`
	Cluster string `json:"cluster" jsonschema:"description=Object Storage cluster ID (deprecated, use Region)"`
	Region  string `json:"region" jsonschema:"description=Object Storage region ID"`
	ACL     string `json:"acl" jsonschema:"description=Access control list (private, public-read, authenticated-read, public-read-write)"`
	CORS    bool   `json:"cors_enabled" jsonschema:"description=Enable CORS"`
}

type ObjectStorageBucketUpdateParams struct {
	Cluster string `json:"cluster" jsonschema:"required,description=Object Storage cluster ID"`
	Bucket  string `json:"bucket" jsonschema:"required,description=Bucket name"`
	ACL     string `json:"acl" jsonschema:"description=New access control list"`
	CORS    *bool  `json:"cors_enabled" jsonschema:"description=Enable or disable CORS"`
}

type ObjectStorageBucketDeleteParams struct {
	Cluster string `json:"cluster" jsonschema:"required,description=Object Storage cluster ID"`
	Bucket  string `json:"bucket" jsonschema:"required,description=Bucket name"`
}

type ObjectStorageKeysListResult struct {
	Keys  []ObjectStorageKeySummary `json:"keys"`
	Count int                       `json:"count"`
}

type ObjectStorageKeySummary struct {
	ID           int                         `json:"id"`
	Label        string                      `json:"label"`
	AccessKey    string                      `json:"access_key"`
	SecretKey    string                      `json:"secret_key"`
	Limited      bool                        `json:"limited"`
	BucketAccess []ObjectStorageBucketAccess `json:"bucket_access"`
}

type ObjectStorageBucketAccess struct {
	Cluster     string `json:"cluster"`
	BucketName  string `json:"bucket_name"`
	Permissions string `json:"permissions"`
}

type ObjectStorageKeyGetParams struct {
	KeyID int `json:"key_id" jsonschema:"required,description=ID of the Object Storage key"`
}

type ObjectStorageKeyDetail struct {
	ID           int                         `json:"id"`
	Label        string                      `json:"label"`
	AccessKey    string                      `json:"access_key"`
	SecretKey    string                      `json:"secret_key"`
	Limited      bool                        `json:"limited"`
	BucketAccess []ObjectStorageBucketAccess `json:"bucket_access"`
}

type ObjectStorageKeyCreateParams struct {
	Label        string                      `json:"label" jsonschema:"required,description=Display label for the key"`
	BucketAccess []ObjectStorageBucketAccess `json:"bucket_access" jsonschema:"description=Bucket access permissions"`
}

type ObjectStorageKeyUpdateParams struct {
	KeyID        int                         `json:"key_id" jsonschema:"required,description=ID of the key to update"`
	Label        string                      `json:"label" jsonschema:"description=New display label"`
	BucketAccess []ObjectStorageBucketAccess `json:"bucket_access" jsonschema:"description=New bucket access permissions"`
}

type ObjectStorageKeyDeleteParams struct {
	KeyID int `json:"key_id" jsonschema:"required,description=ID of the key to delete"`
}

type ObjectStorageClustersListResult struct {
	Clusters []ObjectStorageClusterSummary `json:"clusters"`
	Count    int                           `json:"count"`
}

type ObjectStorageClusterSummary struct {
	ID               string `json:"id"`
	Domain           string `json:"domain"`
	Status           string `json:"status"`
	Region           string `json:"region"`
	StaticSiteDomain string `json:"static_site_domain"`
}

// Advanced Networking types.
type ReservedIPsListResult struct {
	IPs   []ReservedIPSummary `json:"ips"`
	Count int                 `json:"count"`
}

type ReservedIPSummary struct {
	Address    string `json:"address"`
	Gateway    string `json:"gateway"`
	SubnetMask string `json:"subnet_mask"`
	Prefix     int    `json:"prefix"`
	Type       string `json:"type"`
	Public     bool   `json:"public"`
	RDNS       string `json:"rdns"`
	LinodeID   *int   `json:"linode_id"`
	Region     string `json:"region"`
}

type ReservedIPGetParams struct {
	Address string `json:"address" jsonschema:"required,description=IP address to get details for"`
}

type ReservedIPDetail struct {
	Address    string `json:"address"`
	Gateway    string `json:"gateway"`
	SubnetMask string `json:"subnet_mask"`
	Prefix     int    `json:"prefix"`
	Type       string `json:"type"`
	Public     bool   `json:"public"`
	RDNS       string `json:"rdns"`
	LinodeID   *int   `json:"linode_id"`
	Region     string `json:"region"`
}

type ReservedIPAllocateParams struct {
	Type     string `json:"type" jsonschema:"required,description=Type of IP (ipv4, ipv6)"`
	Public   bool   `json:"public" jsonschema:"description=Whether the IP should be public"`
	LinodeID int    `json:"linode_id" jsonschema:"description=Linode to assign the IP to"`
	Region   string `json:"region" jsonschema:"description=Region for the IP (required if no linode_id)"`
}

type ReservedIPAssignParams struct {
	Address  string `json:"address" jsonschema:"required,description=IP address to assign"`
	LinodeID int    `json:"linode_id" jsonschema:"description=Linode to assign to (0 to unassign)"`
	Region   string `json:"region" jsonschema:"required,description=Region where the IP assignment takes place"`
}

type ReservedIPUpdateParams struct {
	Address string `json:"address" jsonschema:"required,description=IP address to update"`
	RDNS    string `json:"rdns" jsonschema:"description=Reverse DNS for the IP"`
}

type VLANsListResult struct {
	VLANs []VLANSummary `json:"vlans"`
	Count int           `json:"count"`
}

type VLANSummary struct {
	Label   string `json:"label"`
	Linodes []int  `json:"linodes"`
	Region  string `json:"region"`
	Created string `json:"created"`
}

type VLANGetParams struct {
	Label string `json:"label" jsonschema:"required,description=VLAN label"`
}

type VLANDetail struct {
	Label   string `json:"label"`
	Linodes []int  `json:"linodes"`
	Region  string `json:"region"`
	Created string `json:"created"`
}

type IPv6PoolsListResult struct {
	Pools []IPv6PoolSummary `json:"pools"`
	Count int               `json:"count"`
}

type IPv6PoolSummary struct {
	Range  string `json:"range"`
	Region string `json:"region"`
}

type IPv6RangesListResult struct {
	Ranges []IPv6RangeSummary `json:"ranges"`
	Count  int                `json:"count"`
}

type IPv6RangeSummary struct {
	Range       string `json:"range"`
	Region      string `json:"region"`
	Prefix      int    `json:"prefix"`
	RouteTarget string `json:"route_target"`
}

// Monitoring (Longview) types.
type LongviewClientsListResult struct {
	Clients []LongviewClientSummary `json:"clients"`
	Count   int                     `json:"count"`
}

type LongviewClientSummary struct {
	ID      int    `json:"id"`
	Label   string `json:"label"`
	APIKey  string `json:"api_key"`
	Created string `json:"created"`
	Updated string `json:"updated"`
}

type LongviewClientGetParams struct {
	ClientID int `json:"client_id" jsonschema:"required,description=ID of the Longview client"`
}

type LongviewClientDetail struct {
	ID      int                    `json:"id"`
	Label   string                 `json:"label"`
	APIKey  string                 `json:"api_key"`
	Created string                 `json:"created"`
	Updated string                 `json:"updated"`
	Apps    map[string]interface{} `json:"apps"`
}

type LongviewClientCreateParams struct {
	Label string `json:"label" jsonschema:"required,description=Display label for the Longview client"`
}

type LongviewClientUpdateParams struct {
	ClientID int    `json:"client_id" jsonschema:"required,description=ID of the client to update"`
	Label    string `json:"label" jsonschema:"description=New display label"`
}

type LongviewClientDeleteParams struct {
	ClientID int `json:"client_id" jsonschema:"required,description=ID of the client to delete"`
}

// Support types.
type SupportTicketsListResult struct {
	Tickets []SupportTicketSummary `json:"tickets"`
	Count   int                    `json:"count"`
}

type SupportTicketSummary struct {
	ID          int                 `json:"id"`
	Summary     string              `json:"summary"`
	Description string              `json:"description"`
	Status      string              `json:"status"`
	Entity      SupportTicketEntity `json:"entity"`
	OpenedBy    string              `json:"opened_by"`
	UpdatedBy   string              `json:"updated_by"`
	Opened      string              `json:"opened"`
	Updated     string              `json:"updated"`
	Closeable   bool                `json:"closeable"`
}

type SupportTicketEntity struct {
	ID    int    `json:"id"`
	Label string `json:"label"`
	Type  string `json:"type"`
	URL   string `json:"url"`
}

type SupportTicketGetParams struct {
	TicketID int `json:"ticket_id" jsonschema:"required,description=ID of the support ticket"`
}

type SupportTicketDetail struct {
	ID          int                 `json:"id"`
	Summary     string              `json:"summary"`
	Description string              `json:"description"`
	Status      string              `json:"status"`
	Entity      SupportTicketEntity `json:"entity"`
	OpenedBy    string              `json:"opened_by"`
	UpdatedBy   string              `json:"updated_by"`
	Opened      string              `json:"opened"`
	Updated     string              `json:"updated"`
	ClosedBy    string              `json:"closed_by"`
	Closeable   bool                `json:"closeable"`
	GravatarID  string              `json:"gravatar_id"`
	Attachments []string            `json:"attachments"`
}

type SupportTicketReply struct {
	Description string `json:"description"`
	Created     string `json:"created"`
	CreatedBy   string `json:"created_by"`
	FromLinode  bool   `json:"from_linode"`
}

type SupportTicketCreateParams struct {
	Summary        string `json:"summary" jsonschema:"required,description=Brief summary of the issue"`
	Description    string `json:"description" jsonschema:"required,description=Detailed description of the issue"`
	LinodeID       *int   `json:"linode_id" jsonschema:"description=ID of related Linode"`
	DomainID       *int   `json:"domain_id" jsonschema:"description=ID of related domain"`
	NodebalancerID *int   `json:"nodebalancer_id" jsonschema:"description=ID of related NodeBalancer"`
	VolumeID       *int   `json:"volume_id" jsonschema:"description=ID of related volume"`
}

type SupportTicketReplyParams struct {
	TicketID    int    `json:"ticket_id" jsonschema:"required,description=ID of the ticket to reply to"`
	Description string `json:"description" jsonschema:"required,description=Reply content"`
}

// Additional Database types for MySQL/PostgreSQL specific operations

type MySQLDatabaseSummary struct {
	ID          int           `json:"id"`
	Label       string        `json:"label"`
	Engine      string        `json:"engine"`
	Version     string        `json:"version"`
	Region      string        `json:"region"`
	Type        string        `json:"type"`
	Status      string        `json:"status"`
	ClusterSize int           `json:"cluster_size"`
	Hosts       DatabaseHosts `json:"hosts"`
	Port        int           `json:"port"`
	Created     string        `json:"created"`
	Updated     string        `json:"updated"`
}

type PostgresDatabaseSummary struct {
	ID          int           `json:"id"`
	Label       string        `json:"label"`
	Engine      string        `json:"engine"`
	Version     string        `json:"version"`
	Region      string        `json:"region"`
	Type        string        `json:"type"`
	Status      string        `json:"status"`
	ClusterSize int           `json:"cluster_size"`
	Hosts       DatabaseHosts `json:"hosts"`
	Port        int           `json:"port"`
	Created     string        `json:"created"`
	Updated     string        `json:"updated"`
}

type MySQLDatabaseDetail struct {
	ID          int           `json:"id"`
	Label       string        `json:"label"`
	Engine      string        `json:"engine"`
	Version     string        `json:"version"`
	Region      string        `json:"region"`
	Type        string        `json:"type"`
	Status      string        `json:"status"`
	ClusterSize int           `json:"cluster_size"`
	Hosts       DatabaseHosts `json:"hosts"`
	Port        int           `json:"port"`
	AllowList   []string      `json:"allow_list"`
	Created     string        `json:"created"`
	Updated     string        `json:"updated"`
}

type PostgresDatabaseDetail struct {
	ID          int           `json:"id"`
	Label       string        `json:"label"`
	Engine      string        `json:"engine"`
	Version     string        `json:"version"`
	Region      string        `json:"region"`
	Type        string        `json:"type"`
	Status      string        `json:"status"`
	ClusterSize int           `json:"cluster_size"`
	Hosts       DatabaseHosts `json:"hosts"`
	Port        int           `json:"port"`
	AllowList   []string      `json:"allow_list"`
	Created     string        `json:"created"`
	Updated     string        `json:"updated"`
}

// Database parameter types

type MySQLDatabaseGetParams struct {
	DatabaseID int `json:"database_id" jsonschema:"required,description=ID of the MySQL database"`
}

type PostgresDatabaseGetParams struct {
	DatabaseID int `json:"database_id" jsonschema:"required,description=ID of the PostgreSQL database"`
}

type MySQLDatabaseCreateParams struct {
	Label       string   `json:"label" jsonschema:"required,description=Display label for the database"`
	Region      string   `json:"region" jsonschema:"required,description=Region where the database will be created"`
	Type        string   `json:"type" jsonschema:"required,description=Database type (e.g. g6-nanode-1)"`
	Engine      string   `json:"engine" jsonschema:"required,description=Database engine (e.g. mysql/8.0.30)"`
	ClusterSize int      `json:"cluster_size" jsonschema:"description=Number of nodes in the cluster (1 or 3)"`
	AllowList   []string `json:"allow_list" jsonschema:"description=List of IP addresses/ranges allowed to access the database"`
}

type PostgresDatabaseCreateParams struct {
	Label       string   `json:"label" jsonschema:"required,description=Display label for the database"`
	Region      string   `json:"region" jsonschema:"required,description=Region where the database will be created"`
	Type        string   `json:"type" jsonschema:"required,description=Database type (e.g. g6-nanode-1)"`
	Engine      string   `json:"engine" jsonschema:"required,description=Database engine (e.g. postgresql/14.9)"`
	ClusterSize int      `json:"cluster_size" jsonschema:"description=Number of nodes in the cluster (1 or 3)"`
	AllowList   []string `json:"allow_list" jsonschema:"description=List of IP addresses/ranges allowed to access the database"`
}

type MySQLDatabaseUpdateParams struct {
	DatabaseID int      `json:"database_id" jsonschema:"required,description=ID of the MySQL database to update"`
	Label      string   `json:"label" jsonschema:"description=New display label for the database"`
	AllowList  []string `json:"allow_list" jsonschema:"description=Updated list of IP addresses/ranges allowed to access the database"`
}

type PostgresDatabaseUpdateParams struct {
	DatabaseID int      `json:"database_id" jsonschema:"required,description=ID of the PostgreSQL database to update"`
	Label      string   `json:"label" jsonschema:"description=New display label for the database"`
	AllowList  []string `json:"allow_list" jsonschema:"description=Updated list of IP addresses/ranges allowed to access the database"`
}

type MySQLDatabaseDeleteParams struct {
	DatabaseID int `json:"database_id" jsonschema:"required,description=ID of the MySQL database to delete"`
}

type PostgresDatabaseDeleteParams struct {
	DatabaseID int `json:"database_id" jsonschema:"required,description=ID of the PostgreSQL database to delete"`
}

type MySQLDatabaseCredentialsParams struct {
	DatabaseID int `json:"database_id" jsonschema:"required,description=ID of the MySQL database"`
}

type PostgresDatabaseCredentialsParams struct {
	DatabaseID int `json:"database_id" jsonschema:"required,description=ID of the PostgreSQL database"`
}

type MySQLDatabaseCredentialsResetParams struct {
	DatabaseID int `json:"database_id" jsonschema:"required,description=ID of the MySQL database"`
}

type PostgresDatabaseCredentialsResetParams struct {
	DatabaseID int `json:"database_id" jsonschema:"required,description=ID of the PostgreSQL database"`
}
