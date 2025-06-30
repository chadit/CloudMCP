package linode

// AccountSwitchParams represents the parameters required to switch between configured Linode accounts.
type AccountSwitchParams struct {
	AccountName string `json:"accountName" jsonschema:"required,description=Name of the account to switch to"`
}

// AccountInfo represents basic information about a configured Linode account.
type AccountInfo struct {
	Name      string `json:"name"`
	Label     string `json:"label"`
	IsCurrent bool   `json:"isCurrent"`
}

// AccountListResult contains the list of configured Linode accounts and the currently active account.
type AccountListResult struct {
	Accounts []AccountInfo `json:"accounts"`
	Current  string        `json:"currentAccount"`
}

// InstancesListResult contains a list of Linode instances and the total count.
type InstancesListResult struct {
	Instances []InstanceSummary `json:"instances"`
	Count     int               `json:"count"`
}

// InstanceSummary provides basic information about a Linode instance.
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

// InstanceGetParams represents the parameters required to retrieve a specific Linode instance.
type InstanceGetParams struct {
	InstanceID int `json:"instanceId" jsonschema:"required,description=ID of the Linode instance"`
}

// InstanceDetail provides comprehensive information about a Linode instance including specs, alerts, and backup status.
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
	WatchdogEnabled bool            `json:"watchdogEnabled"`
}

// InstanceSpecs represents the specifications of a Linode instance.
type InstanceSpecs struct {
	Disk     int `json:"disk"`
	Memory   int `json:"memory"`
	VCPUs    int `json:"vcpus"`
	GPUs     int `json:"gpus"`
	Transfer int `json:"transfer"`
}

// InstanceAlerts represents alert thresholds for a Linode instance.
type InstanceAlerts struct {
	CPU           int `json:"cpu"`
	NetworkIn     int `json:"networkIn"`
	NetworkOut    int `json:"networkOut"`
	TransferQuota int `json:"transferQuota"`
	IO            int `json:"io"`
}

// InstanceBackups represents backup configuration for a Linode instance.
type InstanceBackups struct {
	Enabled        bool           `json:"enabled"`
	Schedule       BackupSchedule `json:"schedule"`
	LastSuccessful string         `json:"lastSuccessful"`
	Available      bool           `json:"available"`
}

// BackupSchedule represents the schedule for automatic backups.
type BackupSchedule struct {
	Day    string `json:"day"`
	Window string `json:"window"`
}

// Volume types.

// VolumesListResult contains a list of Linode volumes and the total count.
type VolumesListResult struct {
	Volumes []VolumeSummary `json:"volumes"`
	Count   int             `json:"count"`
}

// VolumeSummary provides basic information about a Linode volume.
type VolumeSummary struct {
	ID             int      `json:"id"`
	Label          string   `json:"label"`
	Status         string   `json:"status"`
	Size           int      `json:"size"`
	Region         string   `json:"region"`
	LinodeID       *int     `json:"linodeId"`
	LinodeLabel    string   `json:"linodeLabel"`
	Created        string   `json:"created"`
	Updated        string   `json:"updated"`
	FilesystemPath string   `json:"filesystemPath"`
	Tags           []string `json:"tags"`
}

// VolumeGetParams represents the parameters required to retrieve a specific volume.
type VolumeGetParams struct {
	VolumeID int `json:"volumeId" jsonschema:"required,description=ID of the volume"`
}

// IP types.

// IPsListResult contains a list of IP addresses and the total count.
type IPsListResult struct {
	IPs   []IPInfo `json:"ips"`
	Count int      `json:"count"`
}

// IPInfo provides information about an IP address.
type IPInfo struct {
	Address    string `json:"address"`
	Gateway    string `json:"gateway"`
	SubnetMask string `json:"subnetMask"`
	Prefix     int    `json:"prefix"`
	Type       string `json:"type"`
	Public     bool   `json:"public"`
	RDNS       string `json:"rdns"`
	LinodeID   int    `json:"linodeId"`
	Region     string `json:"region"`
}

// IPGetParams represents the parameters required to retrieve IP address details.
type IPGetParams struct {
	Address string `json:"address" jsonschema:"required,description=IP address to get details for"`
}

// Instance operations.

// InstanceCreateParams represents the parameters required to create a new Linode instance.
type InstanceCreateParams struct {
	Region         string   `json:"region"         jsonschema:"required,description=Region ID where the instance will be created"`
	Type           string   `json:"type"           jsonschema:"required,description=Linode type ID (e.g. g6-nanode-1)"`
	Label          string   `json:"label"          jsonschema:"required,description=Display label for the instance"`
	Image          string   `json:"image"          jsonschema:"description=Image ID to deploy (e.g. linode/ubuntu22.04)"`
	RootPass       string   `json:"rootPass"       jsonschema:"description=Root password for the instance"`
	AuthorizedKeys []string `json:"authorizedKeys" jsonschema:"description=SSH public keys to add to root user"`
	StackscriptID  int      `json:"stackscriptId"  jsonschema:"description=StackScript ID to run on first boot"`
	BackupsEnabled bool     `json:"backupsEnabled" jsonschema:"description=Enable automatic backups"`
	PrivateIP      bool     `json:"privateIp"      jsonschema:"description=Add a private IP address"`
	Tags           []string `json:"tags"           jsonschema:"description=Tags to apply to the instance"`
}

// InstanceDeleteParams represents the parameters required to delete a Linode instance.
type InstanceDeleteParams struct {
	InstanceID int `json:"instanceId" jsonschema:"required,description=ID of the Linode instance to delete"`
}

// InstanceBootParams represents the parameters required to boot a Linode instance.
type InstanceBootParams struct {
	InstanceID int `json:"instanceId" jsonschema:"required,description=ID of the Linode instance to boot"`
	ConfigID   int `json:"configId"   jsonschema:"description=Configuration profile ID to boot"`
}

// InstanceShutdownParams represents the parameters required to shutdown a Linode instance.
type InstanceShutdownParams struct {
	InstanceID int `json:"instanceId" jsonschema:"required,description=ID of the Linode instance to shutdown"`
}

// InstanceRebootParams represents the parameters required to reboot a Linode instance.
type InstanceRebootParams struct {
	InstanceID int `json:"instanceId" jsonschema:"required,description=ID of the Linode instance to reboot"`
	ConfigID   int `json:"configId"   jsonschema:"description=Configuration profile ID to reboot into"`
}

// Volume operations.

// VolumeCreateParams represents the parameters required to create a new volume.
type VolumeCreateParams struct {
	Label    string   `json:"label"    jsonschema:"required,description=Display label for the volume"`
	Size     int      `json:"size"     jsonschema:"required,description=Size of the volume in GB (10-8192)"`
	Region   string   `json:"region"   jsonschema:"description=Region ID where the volume will be created"`
	LinodeID int      `json:"linodeId" jsonschema:"description=ID of Linode to attach the volume to"`
	Tags     []string `json:"tags"     jsonschema:"description=Tags to apply to the volume"`
}

// VolumeDeleteParams represents the parameters required to delete a volume.
type VolumeDeleteParams struct {
	VolumeID int `json:"volumeId" jsonschema:"required,description=ID of the volume to delete"`
}

// VolumeAttachParams represents the parameters required to attach a volume to a Linode.
type VolumeAttachParams struct {
	VolumeID           int  `json:"volumeId"           jsonschema:"required,description=ID of the volume to attach"`
	LinodeID           int  `json:"linodeId"           jsonschema:"required,description=ID of the Linode to attach to"`
	PersistAcrossBoots bool `json:"persistAcrossBoots" jsonschema:"description=Keep volume attached when Linode reboots"`
}

// VolumeDetachParams represents the parameters required to detach a volume from a Linode.
type VolumeDetachParams struct {
	VolumeID int `json:"volumeId" jsonschema:"required,description=ID of the volume to detach"`
}

// Image types.

// ImagesListParams represents the parameters for listing images.
type ImagesListParams struct {
	IsPublic *bool `json:"isPublic" jsonschema:"description=Filter to only public (true) or private (false) images"`
}

// ImagesListResult contains a list of images and the total count.
type ImagesListResult struct {
	Images []ImageSummary `json:"images"`
	Count  int            `json:"count"`
}

// ImageSummary provides basic information about a Linode image.
type ImageSummary struct {
	ID           string        `json:"id"`
	Label        string        `json:"label"`
	Description  string        `json:"description"`
	Created      string        `json:"created"`
	CreatedBy    string        `json:"createdBy"`
	Deprecated   bool          `json:"deprecated"`
	IsPublic     bool          `json:"isPublic"`
	Size         int           `json:"size"`
	Type         string        `json:"type"`
	Vendor       string        `json:"vendor"`
	Status       string        `json:"status"`
	Regions      []ImageRegion `json:"regions"`
	Tags         []string      `json:"tags"`
	TotalSize    int           `json:"totalSize"`
	Capabilities []string      `json:"capabilities"`
}

// ImageRegion represents the availability status of an image in a specific region.
type ImageRegion struct {
	Region string `json:"region"`
	Status string `json:"status"`
}

// ImageGetParams represents the parameters required to retrieve a specific image.
type ImageGetParams struct {
	ImageID string `json:"imageId" jsonschema:"required,description=ID of the image (e.g. linode/ubuntu22.04 or private/12345)"`
}

// ImageDetail provides comprehensive information about a Linode image.
type ImageDetail struct {
	ID           string        `json:"id"`
	Label        string        `json:"label"`
	Description  string        `json:"description"`
	Created      string        `json:"created"`
	CreatedBy    string        `json:"createdBy"`
	Deprecated   bool          `json:"deprecated"`
	IsPublic     bool          `json:"isPublic"`
	Size         int           `json:"size"`
	Type         string        `json:"type"`
	Vendor       string        `json:"vendor"`
	Status       string        `json:"status"`
	Regions      []ImageRegion `json:"regions"`
	Tags         []string      `json:"tags"`
	TotalSize    int           `json:"totalSize"`
	Capabilities []string      `json:"capabilities"`
	Updated      string        `json:"updated"`
	Expiry       *string       `json:"expiry"`
}

// Image operations.

// ImageCreateParams represents the parameters required to create a new image from a disk.
type ImageCreateParams struct {
	DiskID      int      `json:"diskId"      jsonschema:"required,description=ID of the Linode disk to create image from"`
	Label       string   `json:"label"       jsonschema:"required,description=Display label for the image"`
	Description string   `json:"description" jsonschema:"description=Detailed description of the image"`
	CloudInit   bool     `json:"cloudInit"   jsonschema:"description=Whether this image supports cloud-init"`
	Tags        []string `json:"tags"        jsonschema:"description=Tags to apply to the image"`
}

// ImageUpdateParams represents the parameters required to update an existing image.
type ImageUpdateParams struct {
	ImageID     string   `json:"imageId"     jsonschema:"required,description=ID of the image to update"`
	Label       string   `json:"label"       jsonschema:"description=New display label for the image"`
	Description string   `json:"description" jsonschema:"description=New description for the image"`
	Tags        []string `json:"tags"        jsonschema:"description=New tags for the image (replaces existing tags)"`
}

// ImageDeleteParams represents the parameters required to delete an image.
type ImageDeleteParams struct {
	ImageID string `json:"imageId" jsonschema:"required,description=ID of the image to delete"`
}

// ImageReplicateParams represents the parameters required to replicate an image to other regions.
type ImageReplicateParams struct {
	ImageID string   `json:"imageId" jsonschema:"required,description=ID of the image to replicate"`
	Regions []string `json:"regions" jsonschema:"required,description=List of region IDs to replicate the image to"`
}

// ImageUploadParams represents the parameters required to upload a new image.
type ImageUploadParams struct {
	Label       string   `json:"label"       jsonschema:"required,description=Display label for the uploaded image"`
	Region      string   `json:"region"      jsonschema:"required,description=Initial region for the uploaded image"`
	Description string   `json:"description" jsonschema:"description=Description of the uploaded image"`
	CloudInit   bool     `json:"cloudInit"   jsonschema:"description=Whether this image supports cloud-init"`
	Tags        []string `json:"tags"        jsonschema:"description=Tags to apply to the image"`
}

// ImageUploadResult contains the result of an image upload operation.
type ImageUploadResult struct {
	ImageID  string `json:"imageId"`
	UploadTo string `json:"uploadTo"`
}

// Firewall types.

// FirewallsListResult contains a list of firewalls and the total count.
type FirewallsListResult struct {
	Firewalls []FirewallSummary `json:"firewalls"`
	Count     int               `json:"count"`
}

// FirewallSummary provides basic information about a Linode firewall.
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

// FirewallGetParams represents the parameters required to retrieve a specific firewall.
type FirewallGetParams struct {
	FirewallID int `json:"firewallId" jsonschema:"required,description=ID of the firewall"`
}

// FirewallDetail provides comprehensive information about a Linode firewall.
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

// FirewallRuleSet represents the complete set of firewall rules and policies.
type FirewallRuleSet struct {
	Inbound        []FirewallRule `json:"inbound"`
	InboundPolicy  string         `json:"inboundPolicy"`
	Outbound       []FirewallRule `json:"outbound"`
	OutboundPolicy string         `json:"outboundPolicy"`
}

// FirewallRule represents a single firewall rule.
type FirewallRule struct {
	Ports       string          `json:"ports"`
	Protocol    string          `json:"protocol"`
	Addresses   FirewallAddress `json:"addresses"`
	Action      string          `json:"action"`
	Label       string          `json:"label"`
	Description string          `json:"description"`
}

// FirewallAddress represents the address constraints for a firewall rule.
type FirewallAddress struct {
	IPv4 []string `json:"ipv4"`
	IPv6 []string `json:"ipv6"`
}

// FirewallDevice represents a device attached to a firewall.
type FirewallDevice struct {
	ID      int    `json:"id"`
	Type    string `json:"type"`
	Label   string `json:"label"`
	URL     string `json:"url"`
	Created string `json:"created"`
}

// FirewallCreateParams represents the parameters required to create a new firewall.
type FirewallCreateParams struct {
	Label string          `json:"label" jsonschema:"required,description=Display label for the firewall"`
	Rules FirewallRuleSet `json:"rules" jsonschema:"description=Firewall rules configuration"`
	Tags  []string        `json:"tags"  jsonschema:"description=Tags to apply to the firewall"`
}

// FirewallUpdateParams represents the parameters required to update an existing firewall.
type FirewallUpdateParams struct {
	FirewallID int      `json:"firewallId" jsonschema:"required,description=ID of the firewall to update"`
	Label      string   `json:"label"      jsonschema:"description=New display label for the firewall"`
	Tags       []string `json:"tags"       jsonschema:"description=New tags for the firewall"`
}

// FirewallDeleteParams represents the parameters required to delete a firewall.
type FirewallDeleteParams struct {
	FirewallID int `json:"firewallId" jsonschema:"required,description=ID of the firewall to delete"`
}

// FirewallRulesUpdateParams represents the parameters required to update firewall rules.
type FirewallRulesUpdateParams struct {
	FirewallID int             `json:"firewallId" jsonschema:"required,description=ID of the firewall to update rules for"`
	Rules      FirewallRuleSet `json:"rules"      jsonschema:"required,description=New firewall rules configuration"`
}

// FirewallDeviceCreateParams represents the parameters required to attach a device to a firewall.
type FirewallDeviceCreateParams struct {
	FirewallID int    `json:"firewallId" jsonschema:"required,description=ID of the firewall"`
	DeviceID   int    `json:"deviceId"   jsonschema:"required,description=ID of the device to assign to firewall"`
	DeviceType string `json:"deviceType" jsonschema:"required,description=Type of device (linode or nodebalancer)"`
}

// FirewallDeviceDeleteParams represents the parameters required to remove a device from a firewall.
type FirewallDeviceDeleteParams struct {
	FirewallID int `json:"firewallId" jsonschema:"required,description=ID of the firewall"`
	DeviceID   int `json:"deviceId"   jsonschema:"required,description=ID of the device to remove from firewall"`
}

// NodeBalancer types.

// NodeBalancersListResult contains a list of NodeBalancers and the total count.
type NodeBalancersListResult struct {
	NodeBalancers []NodeBalancerSummary `json:"nodebalancers"`
	Count         int                   `json:"count"`
}

// NodeBalancerSummary provides basic information about a NodeBalancer.
type NodeBalancerSummary struct {
	ID                 int                  `json:"id"`
	Label              string               `json:"label"`
	Region             string               `json:"region"`
	IPv4               string               `json:"ipv4"`
	IPv6               *string              `json:"ipv6"`
	ClientConnThrottle int                  `json:"clientConnThrottle"`
	Hostname           string               `json:"hostname"`
	Tags               []string             `json:"tags"`
	Created            string               `json:"created"`
	Updated            string               `json:"updated"`
	Transfer           NodeBalancerTransfer `json:"transfer"`
}

// NodeBalancerTransfer represents transfer statistics for a NodeBalancer.
type NodeBalancerTransfer struct {
	In    float64 `json:"in"`
	Out   float64 `json:"out"`
	Total float64 `json:"total"`
}

// NodeBalancerGetParams represents the parameters required to retrieve a specific NodeBalancer.
type NodeBalancerGetParams struct {
	NodeBalancerID int `json:"nodebalancerId" jsonschema:"required,description=ID of the NodeBalancer"`
}

// NodeBalancerDetail provides comprehensive information about a NodeBalancer.
type NodeBalancerDetail struct {
	ID                 int                  `json:"id"`
	Label              string               `json:"label"`
	Region             string               `json:"region"`
	IPv4               string               `json:"ipv4"`
	IPv6               *string              `json:"ipv6"`
	ClientConnThrottle int                  `json:"clientConnThrottle"`
	Hostname           string               `json:"hostname"`
	Tags               []string             `json:"tags"`
	Created            string               `json:"created"`
	Updated            string               `json:"updated"`
	Transfer           NodeBalancerTransfer `json:"transfer"`
	Configs            []NodeBalancerConfig `json:"configs"`
}

// NodeBalancerConfig represents a configuration for a NodeBalancer port.
type NodeBalancerConfig struct {
	ID             int                    `json:"id"`
	Port           int                    `json:"port"`
	Protocol       string                 `json:"protocol"`
	Algorithm      string                 `json:"algorithm"`
	Stickiness     string                 `json:"stickiness"`
	Check          string                 `json:"check"`
	CheckInterval  int                    `json:"checkInterval"`
	CheckTimeout   int                    `json:"checkTimeout"`
	CheckAttempts  int                    `json:"checkAttempts"`
	CheckPath      string                 `json:"checkPath"`
	CheckBody      string                 `json:"checkBody"`
	CheckPassive   bool                   `json:"checkPassive"`
	ProxyProtocol  string                 `json:"proxyProtocol"`
	CipherSuite    string                 `json:"cipherSuite"`
	SSLCommonName  string                 `json:"sslCommonname"`
	SSLFingerprint string                 `json:"sslFingerprint"`
	SSLCert        string                 `json:"sslCert"`
	SSLKey         string                 `json:"sslKey"`
	NodesStatus    NodeBalancerNodeStatus `json:"nodesStatus"`
}

// NodeBalancerNodeStatus represents the status of nodes in a NodeBalancer configuration.
type NodeBalancerNodeStatus struct {
	Up   int `json:"up"`
	Down int `json:"down"`
}

// NodeBalancerCreateParams represents the parameters required to create a new NodeBalancer.
type NodeBalancerCreateParams struct {
	Label              string   `json:"label"              jsonschema:"required,description=Display label for the NodeBalancer"`
	Region             string   `json:"region"             jsonschema:"required,description=Region ID where the NodeBalancer will be created"`
	ClientConnThrottle int      `json:"clientConnThrottle" jsonschema:"description=Throttle connections per second (0-20)"`
	Tags               []string `json:"tags"               jsonschema:"description=Tags to apply to the NodeBalancer"`
}

// NodeBalancerUpdateParams represents the parameters required to update an existing NodeBalancer.
type NodeBalancerUpdateParams struct {
	NodeBalancerID     int      `json:"nodebalancerId"     jsonschema:"required,description=ID of the NodeBalancer to update"`
	Label              string   `json:"label"              jsonschema:"description=New display label for the NodeBalancer"`
	ClientConnThrottle int      `json:"clientConnThrottle" jsonschema:"description=New throttle connections per second (0-20)"`
	Tags               []string `json:"tags"               jsonschema:"description=New tags for the NodeBalancer"`
}

// NodeBalancerDeleteParams represents the parameters required to delete a NodeBalancer.
type NodeBalancerDeleteParams struct {
	NodeBalancerID int `json:"nodebalancerId" jsonschema:"required,description=ID of the NodeBalancer to delete"`
}

// NodeBalancerConfigCreateParams represents the parameters required to create a NodeBalancer configuration.
type NodeBalancerConfigCreateParams struct {
	NodeBalancerID int    `json:"nodebalancerId" jsonschema:"required,description=ID of the NodeBalancer"`
	Port           int    `json:"port"           jsonschema:"required,description=Port to configure (1-65534)"`
	Protocol       string `json:"protocol"       jsonschema:"required,description=Protocol (http, https, tcp)"`
	Algorithm      string `json:"algorithm"      jsonschema:"description=Balancing algorithm (roundrobin, leastconn, source)"`
	Stickiness     string `json:"stickiness"     jsonschema:"description=Session stickiness (none, table, http_cookie)"`
	Check          string `json:"check"          jsonschema:"description=Health check type (none, connection, http, http_body)"`
	CheckInterval  int    `json:"checkInterval"  jsonschema:"description=Health check interval in seconds"`
	CheckTimeout   int    `json:"checkTimeout"   jsonschema:"description=Health check timeout in seconds"`
	CheckAttempts  int    `json:"checkAttempts"  jsonschema:"description=Health check attempts before marking down"`
	CheckPath      string `json:"checkPath"      jsonschema:"description=HTTP health check path"`
	CheckBody      string `json:"checkBody"      jsonschema:"description=Expected response body for http_body check"`
	CheckPassive   bool   `json:"checkPassive"   jsonschema:"description=Enable passive health checks"`
	ProxyProtocol  string `json:"proxyProtocol"  jsonschema:"description=Proxy protocol version (none, v1, v2)"`
	SSLCert        string `json:"sslCert"        jsonschema:"description=SSL certificate for HTTPS"`
	SSLKey         string `json:"sslKey"         jsonschema:"description=SSL private key for HTTPS"`
}

// NodeBalancerConfigUpdateParams represents the parameters required to update a NodeBalancer configuration.
type NodeBalancerConfigUpdateParams struct {
	NodeBalancerID int    `json:"nodebalancerId" jsonschema:"required,description=ID of the NodeBalancer"`
	ConfigID       int    `json:"configId"       jsonschema:"required,description=ID of the configuration to update"`
	Port           int    `json:"port"           jsonschema:"description=Port to configure (1-65534)"`
	Protocol       string `json:"protocol"       jsonschema:"description=Protocol (http, https, tcp)"`
	Algorithm      string `json:"algorithm"      jsonschema:"description=Balancing algorithm (roundrobin, leastconn, source)"`
	Stickiness     string `json:"stickiness"     jsonschema:"description=Session stickiness (none, table, http_cookie)"`
	Check          string `json:"check"          jsonschema:"description=Health check type (none, connection, http, http_body)"`
	CheckInterval  int    `json:"checkInterval"  jsonschema:"description=Health check interval in seconds"`
	CheckTimeout   int    `json:"checkTimeout"   jsonschema:"description=Health check timeout in seconds"`
	CheckAttempts  int    `json:"checkAttempts"  jsonschema:"description=Health check attempts before marking down"`
	CheckPath      string `json:"checkPath"      jsonschema:"description=HTTP health check path"`
	CheckBody      string `json:"checkBody"      jsonschema:"description=Expected response body for http_body check"`
	CheckPassive   bool   `json:"checkPassive"   jsonschema:"description=Enable passive health checks"`
	ProxyProtocol  string `json:"proxyProtocol"  jsonschema:"description=Proxy protocol version (none, v1, v2)"`
	SSLCert        string `json:"sslCert"        jsonschema:"description=SSL certificate for HTTPS"`
	SSLKey         string `json:"sslKey"         jsonschema:"description=SSL private key for HTTPS"`
}

// NodeBalancerConfigDeleteParams represents the parameters required to delete a NodeBalancer configuration.
type NodeBalancerConfigDeleteParams struct {
	NodeBalancerID int `json:"nodebalancerId" jsonschema:"required,description=ID of the NodeBalancer"`
	ConfigID       int `json:"configId"       jsonschema:"required,description=ID of the configuration to delete"`
}

// Domain types.

// DomainsListResult contains a list of domains and the total count.
type DomainsListResult struct {
	Domains []DomainSummary `json:"domains"`
	Count   int             `json:"count"`
}

// DomainSummary provides basic information about a domain.
type DomainSummary struct {
	ID          int      `json:"id"`
	Domain      string   `json:"domain"`
	Type        string   `json:"type"`
	Status      string   `json:"status"`
	Description string   `json:"description"`
	SOAEmail    string   `json:"soaEmail"`
	RetrySec    int      `json:"retrySec"`
	MasterIPs   []string `json:"masterIps"`
	AXfrIPs     []string `json:"axfrIps"`
	Tags        []string `json:"tags"`
	Created     string   `json:"created"`
	Updated     string   `json:"updated"`
}

// DomainGetParams represents the parameters required to retrieve a specific domain.
type DomainGetParams struct {
	DomainID int `json:"domainId" jsonschema:"required,description=ID of the domain"`
}

// DomainDetail provides comprehensive information about a domain.
type DomainDetail struct {
	ID          int      `json:"id"`
	Domain      string   `json:"domain"`
	Type        string   `json:"type"`
	Status      string   `json:"status"`
	Description string   `json:"description"`
	SOAEmail    string   `json:"soaEmail"`
	RetrySec    int      `json:"retrySec"`
	MasterIPs   []string `json:"masterIps"`
	AXfrIPs     []string `json:"axfrIps"`
	Tags        []string `json:"tags"`
	Created     string   `json:"created"`
	Updated     string   `json:"updated"`
	ExpireSec   int      `json:"expireSec"`
	RefreshSec  int      `json:"refreshSec"`
	TTLSec      int      `json:"ttlSec"`
}

// DomainCreateParams represents the parameters required to create a new domain.
type DomainCreateParams struct {
	Domain      string   `json:"domain"      jsonschema:"required,description=Domain name (e.g. example.com)"`
	Type        string   `json:"type"        jsonschema:"required,description=Domain type (master or slave)"`
	SOAEmail    string   `json:"soaEmail"    jsonschema:"description=SOA email address"`
	Description string   `json:"description" jsonschema:"description=Description of the domain"`
	RetrySec    int      `json:"retrySec"    jsonschema:"description=Retry interval in seconds"`
	MasterIPs   []string `json:"masterIps"   jsonschema:"description=Master IPs for slave domains"`
	AXfrIPs     []string `json:"axfrIps"     jsonschema:"description=IPs allowed to AXFR the entire zone"`
	ExpireSec   int      `json:"expireSec"   jsonschema:"description=Expire time in seconds"`
	RefreshSec  int      `json:"refreshSec"  jsonschema:"description=Refresh time in seconds"`
	TTLSec      int      `json:"ttlSec"      jsonschema:"description=Default TTL in seconds"`
	Tags        []string `json:"tags"        jsonschema:"description=Tags to apply to the domain"`
}

// DomainUpdateParams represents the parameters required to update an existing domain.
type DomainUpdateParams struct {
	DomainID    int      `json:"domainId"    jsonschema:"required,description=ID of the domain to update"`
	Domain      string   `json:"domain"      jsonschema:"description=New domain name"`
	Type        string   `json:"type"        jsonschema:"description=New domain type (master or slave)"`
	SOAEmail    string   `json:"soaEmail"    jsonschema:"description=New SOA email address"`
	Description string   `json:"description" jsonschema:"description=New description"`
	RetrySec    int      `json:"retrySec"    jsonschema:"description=New retry interval in seconds"`
	MasterIPs   []string `json:"masterIps"   jsonschema:"description=New master IPs for slave domains"`
	AXfrIPs     []string `json:"axfrIps"     jsonschema:"description=New IPs allowed to AXFR"`
	ExpireSec   int      `json:"expireSec"   jsonschema:"description=New expire time in seconds"`
	RefreshSec  int      `json:"refreshSec"  jsonschema:"description=New refresh time in seconds"`
	TTLSec      int      `json:"ttlSec"      jsonschema:"description=New default TTL in seconds"`
	Tags        []string `json:"tags"        jsonschema:"description=New tags for the domain"`
}

// DomainDeleteParams represents the parameters required to delete a domain.
type DomainDeleteParams struct {
	DomainID int `json:"domainId" jsonschema:"required,description=ID of the domain to delete"`
}

// Domain Record types.

// DomainRecordsListParams represents the parameters required to list domain records.
type DomainRecordsListParams struct {
	DomainID int `json:"domainId" jsonschema:"required,description=ID of the domain"`
}

// DomainRecordsListResult contains a list of domain records and the total count.
type DomainRecordsListResult struct {
	Records []DomainRecord `json:"records"`
	Count   int            `json:"count"`
}

// DomainRecord represents a DNS record for a domain.
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
	TTLSec   int    `json:"ttlSec"`
	Tag      string `json:"tag"`
	Created  string `json:"created"`
	Updated  string `json:"updated"`
}

// DomainRecordGetParams represents the parameters required to retrieve a specific domain record.
type DomainRecordGetParams struct {
	DomainID int `json:"domainId" jsonschema:"required,description=ID of the domain"`
	RecordID int `json:"recordId" jsonschema:"required,description=ID of the domain record"`
}

// DomainRecordCreateParams represents the parameters required to create a new domain record.
type DomainRecordCreateParams struct {
	DomainID int    `json:"domainId" jsonschema:"required,description=ID of the domain"`
	Type     string `json:"type"     jsonschema:"required,description=Record type (A, AAAA, CNAME, MX, TXT, SRV, PTR, CAA, NS)"`
	Name     string `json:"name"     jsonschema:"description=Record name (subdomain)"`
	Target   string `json:"target"   jsonschema:"required,description=Record target (IP, hostname, etc.)"`
	Priority int    `json:"priority" jsonschema:"description=Record priority (for MX and SRV records)"`
	Weight   int    `json:"weight"   jsonschema:"description=Record weight (for SRV records)"`
	Port     int    `json:"port"     jsonschema:"description=Record port (for SRV records)"`
	Service  string `json:"service"  jsonschema:"description=Service name (for SRV records)"`
	Protocol string `json:"protocol" jsonschema:"description=Protocol name (for SRV records)"`
	TTLSec   int    `json:"ttlSec"   jsonschema:"description=TTL in seconds"`
	Tag      string `json:"tag"      jsonschema:"description=CAA record tag"`
}

// DomainRecordUpdateParams represents the parameters required to update an existing domain record.
type DomainRecordUpdateParams struct {
	DomainID int    `json:"domainId" jsonschema:"required,description=ID of the domain"`
	RecordID int    `json:"recordId" jsonschema:"required,description=ID of the record to update"`
	Type     string `json:"type"     jsonschema:"description=New record type"`
	Name     string `json:"name"     jsonschema:"description=New record name"`
	Target   string `json:"target"   jsonschema:"description=New record target"`
	Priority int    `json:"priority" jsonschema:"description=New record priority"`
	Weight   int    `json:"weight"   jsonschema:"description=New record weight"`
	Port     int    `json:"port"     jsonschema:"description=New record port"`
	Service  string `json:"service"  jsonschema:"description=New service name"`
	Protocol string `json:"protocol" jsonschema:"description=New protocol name"`
	TTLSec   int    `json:"ttlSec"   jsonschema:"description=New TTL in seconds"`
	Tag      string `json:"tag"      jsonschema:"description=New CAA record tag"`
}

// DomainRecordDeleteParams represents the parameters required to delete a domain record.
type DomainRecordDeleteParams struct {
	DomainID int `json:"domainId" jsonschema:"required,description=ID of the domain"`
	RecordID int `json:"recordId" jsonschema:"required,description=ID of the record to delete"`
}

// StackScript types.

// StackScriptsListResult contains a list of StackScripts and the total count.
type StackScriptsListResult struct {
	StackScripts []StackScriptSummary `json:"stackscripts"`
	Count        int                  `json:"count"`
}

// StackScriptSummary provides basic information about a StackScript.
type StackScriptSummary struct {
	ID                int      `json:"id"`
	Username          string   `json:"username"`
	Label             string   `json:"label"`
	Description       string   `json:"description"`
	IsPublic          bool     `json:"isPublic"`
	Images            []string `json:"images"`
	DeploymentsTotal  int      `json:"deploymentsTotal"`
	DeploymentsActive int      `json:"deploymentsActive"`
	UserGravatarID    string   `json:"userGravatarId"`
	Created           string   `json:"created"`
	Updated           string   `json:"updated"`
}

// StackScriptGetParams represents the parameters required to retrieve a specific StackScript.
type StackScriptGetParams struct {
	StackScriptID int `json:"stackscriptId" jsonschema:"required,description=ID of the StackScript"`
}

// StackScriptDetail provides comprehensive information about a StackScript.
type StackScriptDetail struct {
	ID                int                           `json:"id"`
	Username          string                        `json:"username"`
	Label             string                        `json:"label"`
	Description       string                        `json:"description"`
	Ordinal           int                           `json:"ordinal"`
	LogoURL           string                        `json:"logoUrl"`
	Images            []string                      `json:"images"`
	DeploymentsTotal  int                           `json:"deploymentsTotal"`
	DeploymentsActive int                           `json:"deploymentsActive"`
	IsPublic          bool                          `json:"isPublic"`
	Mine              bool                          `json:"mine"`
	Created           string                        `json:"created"`
	Updated           string                        `json:"updated"`
	RevNote           string                        `json:"revNote"`
	Script            string                        `json:"script"`
	UserDefinedFields []StackScriptUserDefinedField `json:"userDefinedFields"`
	UserGravatarID    string                        `json:"userGravatarId"`
}

// StackScriptUserDefinedField represents a user-defined field in a StackScript.
type StackScriptUserDefinedField struct {
	Name    string `json:"name"`
	Label   string `json:"label"`
	Example string `json:"example"`
	OneOf   string `json:"oneof"`
	ManyOf  string `json:"manyof"`
	Default string `json:"default"`
}

// StackScriptCreateParams represents the parameters required to create a new StackScript.
type StackScriptCreateParams struct {
	Label       string   `json:"label"       jsonschema:"required,description=Display label for the StackScript"`
	Description string   `json:"description" jsonschema:"description=Description of the StackScript"`
	Images      []string `json:"images"      jsonschema:"required,description=Compatible image IDs"`
	Script      string   `json:"script"      jsonschema:"required,description=StackScript code"`
	IsPublic    bool     `json:"isPublic"    jsonschema:"description=Whether the StackScript should be public"`
	RevNote     string   `json:"revNote"     jsonschema:"description=Revision note for this version"`
}

// StackScriptUpdateParams represents the parameters required to update an existing StackScript.
type StackScriptUpdateParams struct {
	StackScriptID int      `json:"stackscriptId" jsonschema:"required,description=ID of the StackScript to update"`
	Label         string   `json:"label"         jsonschema:"description=New display label"`
	Description   string   `json:"description"   jsonschema:"description=New description"`
	Images        []string `json:"images"        jsonschema:"description=New compatible image IDs"`
	Script        string   `json:"script"        jsonschema:"description=New StackScript code"`
	IsPublic      bool     `json:"isPublic"      jsonschema:"description=New public status"`
	RevNote       string   `json:"revNote"       jsonschema:"description=Revision note for this version"`
}

// StackScriptDeleteParams represents the parameters required to delete a StackScript.
type StackScriptDeleteParams struct {
	StackScriptID int `json:"stackscriptId" jsonschema:"required,description=ID of the StackScript to delete"`
}

// LKE (Kubernetes) types.

// LKEClustersListResult contains a list of LKE clusters and the total count.
type LKEClustersListResult struct {
	Clusters []LKEClusterSummary `json:"clusters"`
	Count    int                 `json:"count"`
}

// LKEClusterSummary provides basic information about an LKE cluster.
type LKEClusterSummary struct {
	ID           int             `json:"id"`
	Label        string          `json:"label"`
	Region       string          `json:"region"`
	Status       string          `json:"status"`
	K8sVersion   string          `json:"k8sVersion"`
	Tags         []string        `json:"tags"`
	Created      string          `json:"created"`
	Updated      string          `json:"updated"`
	ControlPlane LKEControlPlane `json:"controlPlane"`
}

// LKEControlPlane represents the control plane configuration for an LKE cluster.
type LKEControlPlane struct {
	HighAvailability bool `json:"highAvailability"`
}

// LKEClusterGetParams represents the parameters required to retrieve a specific LKE cluster.
type LKEClusterGetParams struct {
	ClusterID int `json:"clusterId" jsonschema:"required,description=ID of the LKE cluster"`
}

// LKEClusterDetail provides comprehensive information about an LKE cluster.
type LKEClusterDetail struct {
	ID           int             `json:"id"`
	Label        string          `json:"label"`
	Region       string          `json:"region"`
	Status       string          `json:"status"`
	K8sVersion   string          `json:"k8sVersion"`
	Tags         []string        `json:"tags"`
	Created      string          `json:"created"`
	Updated      string          `json:"updated"`
	ControlPlane LKEControlPlane `json:"controlPlane"`
	NodePools    []LKENodePool   `json:"nodePools"`
}

// LKENodePool represents a node pool in an LKE cluster.
type LKENodePool struct {
	ID         int                   `json:"id"`
	Count      int                   `json:"count"`
	Type       string                `json:"type"`
	Disks      []LKENodePoolDisk     `json:"disks"`
	Nodes      []LKENode             `json:"nodes"`
	Autoscaler LKENodePoolAutoscaler `json:"autoscaler"`
	Tags       []string              `json:"tags"`
}

// LKENodePoolDisk represents disk configuration for nodes in an LKE node pool.
type LKENodePoolDisk struct {
	Size int    `json:"size"`
	Type string `json:"type"`
}

// LKENode represents a node in an LKE node pool.
type LKENode struct {
	ID         string `json:"id"`
	InstanceID int    `json:"instanceId"`
	Status     string `json:"status"`
}

// LKENodePoolAutoscaler represents autoscaler configuration for an LKE node pool.
type LKENodePoolAutoscaler struct {
	Enabled bool `json:"enabled"`
	Min     int  `json:"min"`
	Max     int  `json:"max"`
}

// LKEClusterCreateParams represents the parameters required to create a new LKE cluster.
type LKEClusterCreateParams struct {
	Label        string              `json:"label"        jsonschema:"required,description=Display label for the cluster"`
	Region       string              `json:"region"       jsonschema:"required,description=Region ID where the cluster will be created"`
	K8sVersion   string              `json:"k8sVersion"   jsonschema:"required,description=Kubernetes version (e.g. 1.28)"`
	Tags         []string            `json:"tags"         jsonschema:"description=Tags to apply to the cluster"`
	NodePools    []LKENodePoolSpec   `json:"nodePools"    jsonschema:"required,description=Node pool specifications"`
	ControlPlane LKEControlPlaneSpec `json:"controlPlane" jsonschema:"description=Control plane configuration"`
}

// LKENodePoolSpec represents the specification for creating an LKE node pool.
type LKENodePoolSpec struct {
	Type       string                 `json:"type"       jsonschema:"required,description=Linode type for nodes (e.g. g6-standard-2)"`
	Count      int                    `json:"count"      jsonschema:"required,description=Number of nodes in pool"`
	Disks      []LKENodePoolDisk      `json:"disks"      jsonschema:"description=Disk configuration for nodes"`
	Autoscaler *LKENodePoolAutoscaler `json:"autoscaler" jsonschema:"description=Autoscaler configuration"`
	Tags       []string               `json:"tags"       jsonschema:"description=Tags for the node pool"`
}

// LKEControlPlaneSpec represents the specification for an LKE control plane.
type LKEControlPlaneSpec struct {
	HighAvailability bool `json:"highAvailability" jsonschema:"description=Enable high availability control plane"`
}

// LKEClusterUpdateParams represents the parameters required to update an existing LKE cluster.
type LKEClusterUpdateParams struct {
	ClusterID    int                 `json:"clusterId"    jsonschema:"required,description=ID of the cluster to update"`
	Label        string              `json:"label"        jsonschema:"description=New display label"`
	K8sVersion   string              `json:"k8sVersion"   jsonschema:"description=New Kubernetes version"`
	Tags         []string            `json:"tags"         jsonschema:"description=New tags for the cluster"`
	ControlPlane LKEControlPlaneSpec `json:"controlPlane" jsonschema:"description=New control plane configuration"`
}

// LKEClusterDeleteParams represents the parameters required to delete an LKE cluster.
type LKEClusterDeleteParams struct {
	ClusterID int `json:"clusterId" jsonschema:"required,description=ID of the cluster to delete"`
}

// LKENodePoolCreateParams represents the parameters required to create a new node pool.
type LKENodePoolCreateParams struct {
	ClusterID  int                    `json:"clusterId"  jsonschema:"required,description=ID of the cluster"`
	Type       string                 `json:"type"       jsonschema:"required,description=Linode type for nodes"`
	Count      int                    `json:"count"      jsonschema:"required,description=Number of nodes in pool"`
	Disks      []LKENodePoolDisk      `json:"disks"      jsonschema:"description=Disk configuration for nodes"`
	Autoscaler *LKENodePoolAutoscaler `json:"autoscaler" jsonschema:"description=Autoscaler configuration"`
	Tags       []string               `json:"tags"       jsonschema:"description=Tags for the node pool"`
}

// LKENodePoolUpdateParams represents the parameters required to update an existing node pool.
type LKENodePoolUpdateParams struct {
	ClusterID  int                    `json:"clusterId"  jsonschema:"required,description=ID of the cluster"`
	PoolID     int                    `json:"poolId"     jsonschema:"required,description=ID of the node pool to update"`
	Count      int                    `json:"count"      jsonschema:"description=New number of nodes"`
	Autoscaler *LKENodePoolAutoscaler `json:"autoscaler" jsonschema:"description=New autoscaler configuration"`
	Tags       []string               `json:"tags"       jsonschema:"description=New tags for the node pool"`
}

// LKENodePoolDeleteParams represents the parameters required to delete a node pool.
type LKENodePoolDeleteParams struct {
	ClusterID int `json:"clusterId" jsonschema:"required,description=ID of the cluster"`
	PoolID    int `json:"poolId"    jsonschema:"required,description=ID of the node pool to delete"`
}

// LKEKubeconfigParams represents the parameters required to retrieve cluster kubeconfig.
type LKEKubeconfigParams struct {
	ClusterID int `json:"clusterId" jsonschema:"required,description=ID of the cluster"`
}

// Database types.

// DatabasesListResult contains a list of databases and the total count.
type DatabasesListResult struct {
	Databases []DatabaseSummary `json:"databases"`
	Count     int               `json:"count"`
}

// DatabaseSummary provides basic information about a database.
type DatabaseSummary struct {
	ID            int      `json:"id"`
	Label         string   `json:"label"`
	Engine        string   `json:"engine"`
	Version       string   `json:"version"`
	Region        string   `json:"region"`
	Status        string   `json:"status"`
	ClusterSize   int      `json:"clusterSize"`
	Type          string   `json:"type"`
	Encrypted     bool     `json:"encrypted"`
	Allow         []string `json:"allowList"`
	Port          int      `json:"port"`
	SSLConnection bool     `json:"sslConnection"`
	Created       string   `json:"created"`
	Updated       string   `json:"updated"`
	Tags          []string `json:"tags"`
}

// DatabaseGetParams represents the parameters required to retrieve a specific database.
type DatabaseGetParams struct {
	DatabaseID int `json:"databaseId" jsonschema:"required,description=ID of the database"`
}

// DatabaseDetail provides comprehensive information about a database.
type DatabaseDetail struct {
	ID            int             `json:"id"`
	Label         string          `json:"label"`
	Engine        string          `json:"engine"`
	Version       string          `json:"version"`
	Region        string          `json:"region"`
	Status        string          `json:"status"`
	ClusterSize   int             `json:"clusterSize"`
	Type          string          `json:"type"`
	Encrypted     bool            `json:"encrypted"`
	Allow         []string        `json:"allowList"`
	Port          int             `json:"port"`
	SSLConnection bool            `json:"sslConnection"`
	Created       string          `json:"created"`
	Updated       string          `json:"updated"`
	Tags          []string        `json:"tags"`
	Hosts         DatabaseHosts   `json:"hosts"`
	Updates       DatabaseUpdates `json:"updates"`
	Backups       DatabaseBackups `json:"backups"`
}

// DatabaseHosts represents the host information for a database.
type DatabaseHosts struct {
	Primary   string `json:"primary"`
	Secondary string `json:"secondary"`
}

// DatabaseUpdates represents the maintenance window configuration for a database.
type DatabaseUpdates struct {
	Frequency   string `json:"frequency"`
	Duration    int    `json:"duration"`
	HourOfDay   int    `json:"hourOfDay"`
	DayOfWeek   int    `json:"dayOfWeek"`
	WeekOfMonth int    `json:"weekOfMonth"`
}

// DatabaseBackups represents the backup configuration for a database.
type DatabaseBackups struct {
	Enabled  bool                   `json:"enabled"`
	Schedule DatabaseBackupSchedule `json:"schedule"`
}

// DatabaseBackupSchedule represents the schedule for database backups.
type DatabaseBackupSchedule struct {
	Day    string `json:"day"`
	Window string `json:"window"`
}

// DatabaseCreateParams represents the parameters required to create a new database.
type DatabaseCreateParams struct {
	Label         string   `json:"label"         jsonschema:"required,description=Display label for the database"`
	Engine        string   `json:"engine"        jsonschema:"required,description=Database engine (mysql or postgresql)"`
	Version       string   `json:"version"       jsonschema:"description=Database version"`
	Region        string   `json:"region"        jsonschema:"required,description=Region ID where the database will be created"`
	Type          string   `json:"type"          jsonschema:"required,description=Database plan type"`
	ClusterSize   int      `json:"clusterSize"   jsonschema:"description=Number of nodes in cluster (1 or 3)"`
	Encrypted     bool     `json:"encrypted"     jsonschema:"description=Enable encryption at rest"`
	SSLConnection bool     `json:"sslConnection" jsonschema:"description=Require SSL connections"`
	Allow         []string `json:"allowList"     jsonschema:"description=IP addresses allowed to connect"`
	Tags          []string `json:"tags"          jsonschema:"description=Tags to apply to the database"`
}

// DatabaseUpdateParams represents the parameters required to update an existing database.
type DatabaseUpdateParams struct {
	DatabaseID int             `json:"databaseId" jsonschema:"required,description=ID of the database to update"`
	Label      string          `json:"label"      jsonschema:"description=New display label"`
	Allow      []string        `json:"allowList"  jsonschema:"description=New IP addresses allowed to connect"`
	Updates    DatabaseUpdates `json:"updates"    jsonschema:"description=New maintenance window settings"`
	Tags       []string        `json:"tags"       jsonschema:"description=New tags for the database"`
}

// DatabaseDeleteParams represents the parameters required to delete a database.
type DatabaseDeleteParams struct {
	DatabaseID int `json:"databaseId" jsonschema:"required,description=ID of the database to delete"`
}

// DatabaseBackupCreateParams represents the parameters required to create a database backup.
type DatabaseBackupCreateParams struct {
	DatabaseID int    `json:"databaseId" jsonschema:"required,description=ID of the database"`
	Label      string `json:"label"      jsonschema:"required,description=Label for the backup"`
	Target     string `json:"target"     jsonschema:"description=Backup target (primary or secondary)"`
}

// DatabaseBackupRestoreParams represents the parameters required to restore a database backup.
type DatabaseBackupRestoreParams struct {
	DatabaseID int `json:"databaseId" jsonschema:"required,description=ID of the database"`
	BackupID   int `json:"backupId"   jsonschema:"required,description=ID of the backup to restore"`
}

// DatabaseCredentialsResetParams represents the parameters required to reset database credentials.
type DatabaseCredentialsResetParams struct {
	DatabaseID int    `json:"databaseId" jsonschema:"required,description=ID of the database"`
	Username   string `json:"username"   jsonschema:"required,description=Username to reset password for"`
}

// Object Storage types.

// ObjectStorageBucketsListResult contains a list of Object Storage buckets and the total count.
type ObjectStorageBucketsListResult struct {
	Buckets []ObjectStorageBucketSummary `json:"buckets"`
	Count   int                          `json:"count"`
}

// ObjectStorageBucketSummary provides basic information about an Object Storage bucket.
type ObjectStorageBucketSummary struct {
	Label    string `json:"label"`
	Region   string `json:"region"`
	Hostname string `json:"hostname"`
	Created  string `json:"created"`
	Size     int64  `json:"size"`
	Objects  int    `json:"objects"`
}

// ObjectStorageBucketGetParams represents the parameters required to retrieve a specific bucket.
type ObjectStorageBucketGetParams struct {
	Region string `json:"region" jsonschema:"required,description=Object Storage region ID"`
	Bucket string `json:"bucket" jsonschema:"required,description=Bucket name"`
}

// ObjectStorageBucketDetail provides comprehensive information about an Object Storage bucket.
type ObjectStorageBucketDetail struct {
	Label    string `json:"label"`
	Region   string `json:"region"`
	Hostname string `json:"hostname"`
	Created  string `json:"created"`
	Size     int64  `json:"size"`
	Objects  int    `json:"objects"`
}

// ObjectStorageBucketCreateParams represents the parameters required to create a new bucket.
type ObjectStorageBucketCreateParams struct {
	Label  string `json:"label"       jsonschema:"required,description=Bucket name/label"`
	Region string `json:"region"      jsonschema:"required,description=Object Storage region ID"`
	ACL    string `json:"acl"         jsonschema:"description=Access control list (private, public-read, authenticated-read, public-read-write)"`
	CORS   bool   `json:"corsEnabled" jsonschema:"description=Enable CORS"`
}

// ObjectStorageBucketUpdateParams represents the parameters required to update an existing bucket.
type ObjectStorageBucketUpdateParams struct {
	Region string `json:"region"      jsonschema:"required,description=Object Storage region ID"`
	Bucket string `json:"bucket"      jsonschema:"required,description=Bucket name"`
	ACL    string `json:"acl"         jsonschema:"description=New access control list"`
	CORS   *bool  `json:"corsEnabled" jsonschema:"description=Enable or disable CORS"`
}

// ObjectStorageBucketDeleteParams represents the parameters required to delete a bucket.
type ObjectStorageBucketDeleteParams struct {
	Region string `json:"region" jsonschema:"required,description=Object Storage region ID"`
	Bucket string `json:"bucket" jsonschema:"required,description=Bucket name"`
}

// ObjectStorageKeysListResult contains a list of Object Storage keys and the total count.
type ObjectStorageKeysListResult struct {
	Keys  []ObjectStorageKeySummary `json:"keys"`
	Count int                       `json:"count"`
}

// ObjectStorageKeySummary provides basic information about an Object Storage key.
type ObjectStorageKeySummary struct {
	ID           int                         `json:"id"`
	Label        string                      `json:"label"`
	AccessKey    string                      `json:"accessKey"`
	SecretKey    string                      `json:"secretKey"`
	Limited      bool                        `json:"limited"`
	BucketAccess []ObjectStorageBucketAccess `json:"bucketAccess"`
}

// ObjectStorageBucketAccess represents access permissions for a specific bucket.
type ObjectStorageBucketAccess struct {
	Region      string `json:"region"`
	BucketName  string `json:"bucketName"`
	Permissions string `json:"permissions"`
}

// ObjectStorageKeyGetParams represents the parameters required to retrieve a specific Object Storage key.
type ObjectStorageKeyGetParams struct {
	KeyID int `json:"keyId" jsonschema:"required,description=ID of the Object Storage key"`
}

// ObjectStorageKeyDetail provides comprehensive information about an Object Storage key.
type ObjectStorageKeyDetail struct {
	ID           int                         `json:"id"`
	Label        string                      `json:"label"`
	AccessKey    string                      `json:"accessKey"`
	SecretKey    string                      `json:"secretKey"`
	Limited      bool                        `json:"limited"`
	BucketAccess []ObjectStorageBucketAccess `json:"bucketAccess"`
}

// ObjectStorageKeyCreateParams represents the parameters required to create a new Object Storage key.
type ObjectStorageKeyCreateParams struct {
	Label        string                      `json:"label"        jsonschema:"required,description=Display label for the key"`
	BucketAccess []ObjectStorageBucketAccess `json:"bucketAccess" jsonschema:"description=Bucket access permissions"`
}

// ObjectStorageKeyUpdateParams represents the parameters required to update an existing Object Storage key.
type ObjectStorageKeyUpdateParams struct {
	KeyID        int                         `json:"keyId"        jsonschema:"required,description=ID of the key to update"`
	Label        string                      `json:"label"        jsonschema:"description=New display label"`
	BucketAccess []ObjectStorageBucketAccess `json:"bucketAccess" jsonschema:"description=New bucket access permissions"`
}

// ObjectStorageKeyDeleteParams represents the parameters required to delete an Object Storage key.
type ObjectStorageKeyDeleteParams struct {
	KeyID int `json:"keyId" jsonschema:"required,description=ID of the key to delete"`
}

// ObjectStorageClustersListResult contains a list of Object Storage clusters and the total count.
type ObjectStorageClustersListResult struct {
	Clusters []ObjectStorageClusterSummary `json:"clusters"`
	Count    int                           `json:"count"`
}

// ObjectStorageClusterSummary provides basic information about an Object Storage cluster.
type ObjectStorageClusterSummary struct {
	ID               string `json:"id"`
	Domain           string `json:"domain"`
	Status           string `json:"status"`
	Region           string `json:"region"`
	StaticSiteDomain string `json:"staticSiteDomain"`
}

// Advanced Networking types.

// ReservedIPsListResult contains a list of reserved IPs and the total count.
type ReservedIPsListResult struct {
	IPs   []ReservedIPSummary `json:"ips"`
	Count int                 `json:"count"`
}

// ReservedIPSummary provides basic information about a reserved IP address.
type ReservedIPSummary struct {
	Address    string `json:"address"`
	Gateway    string `json:"gateway"`
	SubnetMask string `json:"subnetMask"`
	Prefix     int    `json:"prefix"`
	Type       string `json:"type"`
	Public     bool   `json:"public"`
	RDNS       string `json:"rdns"`
	LinodeID   *int   `json:"linodeId"`
	Region     string `json:"region"`
}

// ReservedIPGetParams represents the parameters required to retrieve a specific reserved IP.
type ReservedIPGetParams struct {
	Address string `json:"address" jsonschema:"required,description=IP address to get details for"`
}

// ReservedIPDetail provides comprehensive information about a reserved IP address.
type ReservedIPDetail struct {
	Address    string `json:"address"`
	Gateway    string `json:"gateway"`
	SubnetMask string `json:"subnetMask"`
	Prefix     int    `json:"prefix"`
	Type       string `json:"type"`
	Public     bool   `json:"public"`
	RDNS       string `json:"rdns"`
	LinodeID   *int   `json:"linodeId"`
	Region     string `json:"region"`
}

// ReservedIPAllocateParams represents the parameters required to allocate a new reserved IP.
type ReservedIPAllocateParams struct {
	Type     string `json:"type"     jsonschema:"required,description=Type of IP (ipv4, ipv6)"`
	Public   bool   `json:"public"   jsonschema:"description=Whether the IP should be public"`
	LinodeID int    `json:"linodeId" jsonschema:"description=Linode to assign the IP to"`
	Region   string `json:"region"   jsonschema:"description=Region for the IP (required if no linode_id)"`
}

// ReservedIPAssignParams represents the parameters required to assign a reserved IP.
type ReservedIPAssignParams struct {
	Address  string `json:"address"  jsonschema:"required,description=IP address to assign"`
	LinodeID int    `json:"linodeId" jsonschema:"description=Linode to assign to (0 to unassign)"`
	Region   string `json:"region"   jsonschema:"required,description=Region where the IP assignment takes place"`
}

// ReservedIPUpdateParams represents the parameters required to update a reserved IP.
type ReservedIPUpdateParams struct {
	Address string `json:"address" jsonschema:"required,description=IP address to update"`
	RDNS    string `json:"rdns"    jsonschema:"description=Reverse DNS for the IP"`
}

// VLANsListResult contains a list of VLANs and the total count.
type VLANsListResult struct {
	VLANs []VLANSummary `json:"vlans"`
	Count int           `json:"count"`
}

// VLANSummary provides basic information about a VLAN.
type VLANSummary struct {
	Label   string `json:"label"`
	Linodes []int  `json:"linodes"`
	Region  string `json:"region"`
	Created string `json:"created"`
}

// VLANGetParams represents the parameters required to retrieve a specific VLAN.
type VLANGetParams struct {
	Label string `json:"label" jsonschema:"required,description=VLAN label"`
}

// VLANDetail provides comprehensive information about a VLAN.
type VLANDetail struct {
	Label   string `json:"label"`
	Linodes []int  `json:"linodes"`
	Region  string `json:"region"`
	Created string `json:"created"`
}

// IPv6PoolsListResult contains a list of IPv6 pools and the total count.
type IPv6PoolsListResult struct {
	Pools []IPv6PoolSummary `json:"pools"`
	Count int               `json:"count"`
}

// IPv6PoolSummary provides basic information about an IPv6 pool.
type IPv6PoolSummary struct {
	Range  string `json:"range"`
	Region string `json:"region"`
}

// IPv6RangesListResult contains a list of IPv6 ranges and the total count.
type IPv6RangesListResult struct {
	Ranges []IPv6RangeSummary `json:"ranges"`
	Count  int                `json:"count"`
}

// IPv6RangeSummary provides basic information about an IPv6 range.
type IPv6RangeSummary struct {
	Range       string `json:"range"`
	Region      string `json:"region"`
	Prefix      int    `json:"prefix"`
	RouteTarget string `json:"routeTarget"`
}

// Monitoring (Longview) types.

// LongviewClientsListResult contains a list of Longview clients and the total count.
type LongviewClientsListResult struct {
	Clients []LongviewClientSummary `json:"clients"`
	Count   int                     `json:"count"`
}

// LongviewClientSummary provides basic information about a Longview client.
type LongviewClientSummary struct {
	ID      int    `json:"id"`
	Label   string `json:"label"`
	APIKey  string `json:"apiKey"`
	Created string `json:"created"`
	Updated string `json:"updated"`
}

// LongviewClientGetParams represents the parameters required to retrieve a specific Longview client.
type LongviewClientGetParams struct {
	ClientID int `json:"clientId" jsonschema:"required,description=ID of the Longview client"`
}

// LongviewClientDetail provides comprehensive information about a Longview client.
type LongviewClientDetail struct {
	ID      int                    `json:"id"`
	Label   string                 `json:"label"`
	APIKey  string                 `json:"apiKey"`
	Created string                 `json:"created"`
	Updated string                 `json:"updated"`
	Apps    map[string]interface{} `json:"apps"`
}

// LongviewClientCreateParams represents the parameters required to create a new Longview client.
type LongviewClientCreateParams struct {
	Label string `json:"label" jsonschema:"required,description=Display label for the Longview client"`
}

// LongviewClientUpdateParams represents the parameters required to update an existing Longview client.
type LongviewClientUpdateParams struct {
	ClientID int    `json:"clientId" jsonschema:"required,description=ID of the client to update"`
	Label    string `json:"label"    jsonschema:"description=New display label"`
}

// LongviewClientDeleteParams represents the parameters required to delete a Longview client.
type LongviewClientDeleteParams struct {
	ClientID int `json:"clientId" jsonschema:"required,description=ID of the client to delete"`
}

// Support types.

// SupportTicketsListResult contains a list of support tickets and the total count.
type SupportTicketsListResult struct {
	Tickets []SupportTicketSummary `json:"tickets"`
	Count   int                    `json:"count"`
}

// SupportTicketSummary provides basic information about a support ticket.
type SupportTicketSummary struct {
	ID          int                 `json:"id"`
	Summary     string              `json:"summary"`
	Description string              `json:"description"`
	Status      string              `json:"status"`
	Entity      SupportTicketEntity `json:"entity"`
	OpenedBy    string              `json:"openedBy"`
	UpdatedBy   string              `json:"updatedBy"`
	Opened      string              `json:"opened"`
	Updated     string              `json:"updated"`
	Closeable   bool                `json:"closeable"`
}

// SupportTicketEntity represents the entity associated with a support ticket.
type SupportTicketEntity struct {
	ID    int    `json:"id"`
	Label string `json:"label"`
	Type  string `json:"type"`
	URL   string `json:"url"`
}

// SupportTicketGetParams represents the parameters required to retrieve a specific support ticket.
type SupportTicketGetParams struct {
	TicketID int `json:"ticketId" jsonschema:"required,description=ID of the support ticket"`
}

// SupportTicketDetail provides comprehensive information about a support ticket.
type SupportTicketDetail struct {
	ID          int                 `json:"id"`
	Summary     string              `json:"summary"`
	Description string              `json:"description"`
	Status      string              `json:"status"`
	Entity      SupportTicketEntity `json:"entity"`
	OpenedBy    string              `json:"openedBy"`
	UpdatedBy   string              `json:"updatedBy"`
	Opened      string              `json:"opened"`
	Updated     string              `json:"updated"`
	ClosedBy    string              `json:"closedBy"`
	Closeable   bool                `json:"closeable"`
	GravatarID  string              `json:"gravatarId"`
	Attachments []string            `json:"attachments"`
}

// SupportTicketReply represents a reply to a support ticket.
type SupportTicketReply struct {
	Description string `json:"description"`
	Created     string `json:"created"`
	CreatedBy   string `json:"createdBy"`
	FromLinode  bool   `json:"fromLinode"`
}

// SupportTicketCreateParams represents the parameters required to create a new support ticket.
type SupportTicketCreateParams struct {
	Summary        string `json:"summary"        jsonschema:"required,description=Brief summary of the issue"`
	Description    string `json:"description"    jsonschema:"required,description=Detailed description of the issue"`
	LinodeID       *int   `json:"linodeId"       jsonschema:"description=ID of related Linode"`
	DomainID       *int   `json:"domainId"       jsonschema:"description=ID of related domain"`
	NodebalancerID *int   `json:"nodebalancerId" jsonschema:"description=ID of related NodeBalancer"`
	VolumeID       *int   `json:"volumeId"       jsonschema:"description=ID of related volume"`
}

// SupportTicketReplyParams represents the parameters required to reply to a support ticket.
type SupportTicketReplyParams struct {
	TicketID    int    `json:"ticketId"    jsonschema:"required,description=ID of the ticket to reply to"`
	Description string `json:"description" jsonschema:"required,description=Reply content"`
}

// Additional Database types for MySQL/PostgreSQL specific operations.

// MySQLDatabaseSummary provides basic information about a MySQL database.
type MySQLDatabaseSummary struct {
	ID          int           `json:"id"`
	Label       string        `json:"label"`
	Engine      string        `json:"engine"`
	Version     string        `json:"version"`
	Region      string        `json:"region"`
	Type        string        `json:"type"`
	Status      string        `json:"status"`
	ClusterSize int           `json:"clusterSize"`
	Hosts       DatabaseHosts `json:"hosts"`
	Port        int           `json:"port"`
	Created     string        `json:"created"`
	Updated     string        `json:"updated"`
}

// PostgresDatabaseSummary provides basic information about a PostgreSQL database.
type PostgresDatabaseSummary struct {
	ID          int           `json:"id"`
	Label       string        `json:"label"`
	Engine      string        `json:"engine"`
	Version     string        `json:"version"`
	Region      string        `json:"region"`
	Type        string        `json:"type"`
	Status      string        `json:"status"`
	ClusterSize int           `json:"clusterSize"`
	Hosts       DatabaseHosts `json:"hosts"`
	Port        int           `json:"port"`
	Created     string        `json:"created"`
	Updated     string        `json:"updated"`
}

// MySQLDatabaseDetail provides comprehensive information about a MySQL database.
type MySQLDatabaseDetail struct {
	ID          int           `json:"id"`
	Label       string        `json:"label"`
	Engine      string        `json:"engine"`
	Version     string        `json:"version"`
	Region      string        `json:"region"`
	Type        string        `json:"type"`
	Status      string        `json:"status"`
	ClusterSize int           `json:"clusterSize"`
	Hosts       DatabaseHosts `json:"hosts"`
	Port        int           `json:"port"`
	AllowList   []string      `json:"allowList"`
	Created     string        `json:"created"`
	Updated     string        `json:"updated"`
}

// PostgresDatabaseDetail provides comprehensive information about a PostgreSQL database.
type PostgresDatabaseDetail struct {
	ID          int           `json:"id"`
	Label       string        `json:"label"`
	Engine      string        `json:"engine"`
	Version     string        `json:"version"`
	Region      string        `json:"region"`
	Type        string        `json:"type"`
	Status      string        `json:"status"`
	ClusterSize int           `json:"clusterSize"`
	Hosts       DatabaseHosts `json:"hosts"`
	Port        int           `json:"port"`
	AllowList   []string      `json:"allowList"`
	Created     string        `json:"created"`
	Updated     string        `json:"updated"`
}

// Database parameter types.

// MySQLDatabaseGetParams represents the parameters required to retrieve a specific MySQL database.
type MySQLDatabaseGetParams struct {
	DatabaseID int `json:"databaseId" jsonschema:"required,description=ID of the MySQL database"`
}

// PostgresDatabaseGetParams represents the parameters required to retrieve a specific PostgreSQL database.
type PostgresDatabaseGetParams struct {
	DatabaseID int `json:"databaseId" jsonschema:"required,description=ID of the PostgreSQL database"`
}

// MySQLDatabaseCreateParams represents the parameters required to create a new MySQL database.
type MySQLDatabaseCreateParams struct {
	Label       string   `json:"label"       jsonschema:"required,description=Display label for the database"`
	Region      string   `json:"region"      jsonschema:"required,description=Region where the database will be created"`
	Type        string   `json:"type"        jsonschema:"required,description=Database type (e.g. g6-nanode-1)"`
	Engine      string   `json:"engine"      jsonschema:"required,description=Database engine (e.g. mysql/8.0.30)"`
	ClusterSize int      `json:"clusterSize" jsonschema:"description=Number of nodes in the cluster (1 or 3)"`
	AllowList   []string `json:"allowList"   jsonschema:"description=List of IP addresses/ranges allowed to access the database"`
}

// PostgresDatabaseCreateParams represents the parameters required to create a new PostgreSQL database.
type PostgresDatabaseCreateParams struct {
	Label       string   `json:"label"       jsonschema:"required,description=Display label for the database"`
	Region      string   `json:"region"      jsonschema:"required,description=Region where the database will be created"`
	Type        string   `json:"type"        jsonschema:"required,description=Database type (e.g. g6-nanode-1)"`
	Engine      string   `json:"engine"      jsonschema:"required,description=Database engine (e.g. postgresql/14.9)"`
	ClusterSize int      `json:"clusterSize" jsonschema:"description=Number of nodes in the cluster (1 or 3)"`
	AllowList   []string `json:"allowList"   jsonschema:"description=List of IP addresses/ranges allowed to access the database"`
}

// MySQLDatabaseUpdateParams represents the parameters required to update an existing MySQL database.
type MySQLDatabaseUpdateParams struct {
	DatabaseID int      `json:"databaseId" jsonschema:"required,description=ID of the MySQL database to update"`
	Label      string   `json:"label"      jsonschema:"description=New display label for the database"`
	AllowList  []string `json:"allowList"  jsonschema:"description=Updated list of IP addresses/ranges allowed to access the database"`
}

// PostgresDatabaseUpdateParams represents the parameters required to update an existing PostgreSQL database.
type PostgresDatabaseUpdateParams struct {
	DatabaseID int      `json:"databaseId" jsonschema:"required,description=ID of the PostgreSQL database to update"`
	Label      string   `json:"label"      jsonschema:"description=New display label for the database"`
	AllowList  []string `json:"allowList"  jsonschema:"description=Updated list of IP addresses/ranges allowed to access the database"`
}

// MySQLDatabaseDeleteParams represents the parameters required to delete a MySQL database.
type MySQLDatabaseDeleteParams struct {
	DatabaseID int `json:"databaseId" jsonschema:"required,description=ID of the MySQL database to delete"`
}

// PostgresDatabaseDeleteParams represents the parameters required to delete a PostgreSQL database.
type PostgresDatabaseDeleteParams struct {
	DatabaseID int `json:"databaseId" jsonschema:"required,description=ID of the PostgreSQL database to delete"`
}

// MySQLDatabaseCredentialsParams represents the parameters required to retrieve MySQL database credentials.
type MySQLDatabaseCredentialsParams struct {
	DatabaseID int `json:"databaseId" jsonschema:"required,description=ID of the MySQL database"`
}

// PostgresDatabaseCredentialsParams represents the parameters required to retrieve PostgreSQL database credentials.
type PostgresDatabaseCredentialsParams struct {
	DatabaseID int `json:"databaseId" jsonschema:"required,description=ID of the PostgreSQL database"`
}

// MySQLDatabaseCredentialsResetParams represents the parameters required to reset MySQL database credentials.
type MySQLDatabaseCredentialsResetParams struct {
	DatabaseID int `json:"databaseId" jsonschema:"required,description=ID of the MySQL database"`
}

// PostgresDatabaseCredentialsResetParams represents the parameters required to reset PostgreSQL database credentials.
type PostgresDatabaseCredentialsResetParams struct {
	DatabaseID int `json:"databaseId" jsonschema:"required,description=ID of the PostgreSQL database"`
}
