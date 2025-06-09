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
