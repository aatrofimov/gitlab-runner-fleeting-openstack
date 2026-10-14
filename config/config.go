// Package config defines the schema, defaults and validation for the plugin's
// runtime configuration.
//
// The GitLab Runner passes the plugin its configuration as JSON: the
// [runners.autoscaler.plugin_config] table of the runner's config.toml is
// converted to JSON before being handed to the plugin. As a result every field
// is tagged with a `json:` name. Connection-level settings (username, port,
// private key, ...) are delivered separately through provider.Settings and are
// not part of this schema.
package config

import (
	"encoding/json"
	"fmt"
	"time"
)

const (
	defaultComputeMicroversion = "2.72"
	defaultPreemptibleTag      = "preemptible"
	defaultImageName           = "gitlab-builder"
	defaultSSHUser             = "debian"
	defaultInstanceGroupName   = "builder"
	defaultMaxInstances        = 5
	defaultMaxCreateAttempts   = 5
	defaultCreateRetryBackoff  = 10 * time.Second
)

// Config is the top-level plugin configuration.
type Config struct {
	OpenStack     OpenStack `json:"openstack"`
	InstanceGroup Group     `json:"instance_group"`
}

// OpenStack holds the cloud-specific settings used to provision servers.
type OpenStack struct {
	// Cloud is the named cloud in clouds.yaml to use. When empty, the plugin
	// authenticates from the standard OS_* environment variables.
	Cloud string `json:"cloud"`

	// AuthURL overrides the identity endpoint (OS_AUTH_URL / clouds.yaml).
	AuthURL string `json:"auth_url"`

	// Region overrides the region (OS_REGION_NAME / clouds.yaml).
	Region string `json:"region"`

	// ComputeMicroversion is the Nova microversion to request. Selectel needs
	// 2.72 to honor the preemptible server tag.
	ComputeMicroversion string `json:"compute_microversion"`

	// ImageID is the Glance image UUID to boot from. Takes precedence over
	// ImageName.
	ImageID string `json:"image_id"`

	// ImageName is resolved to the latest image UUID via Glance when ImageID
	// is empty.
	ImageName string `json:"image_name"`

	// FlavorID is the Nova flavor UUID for the build servers.
	FlavorID string `json:"flavor_id"`

	// KeypairName is the OpenStack keypair whose public key is injected into
	// each server. The plugin generates the matching private key at startup.
	KeypairName string `json:"keypair_name"`

	// AvailabilityZone optionally pins servers to an availability zone.
	AvailabilityZone string `json:"availability_zone"`

	// Networks lists the Neutron network UUIDs to attach to each server.
	Networks []string `json:"networks"`

	// SecurityGroups lists security group names applied to each server.
	SecurityGroups []string `json:"security_groups"`

	// AssignFloatingIP allocates and attaches a floating IP to each server.
	AssignFloatingIP bool `json:"assign_floating_ip"`

	// FloatingIPNetwork is the external network (name or ID) the floating IP
	// is allocated from. Required when AssignFloatingIP is true.
	FloatingIPNetwork string `json:"floating_ip_network"`

	// PreemptibleTag is the Nova server tag that marks a server preemptible.
	PreemptibleTag string `json:"preemptible_tag"`

	// SSHUser is the default SSH user, used only when the runner's
	// connector_config does not specify a username.
	SSHUser string `json:"ssh_user"`

	// MaxCreateAttempts caps the number of server-create attempts per
	// instance within a single Increase call when a server enters the
	// OpenStack ERROR state.
	MaxCreateAttempts int `json:"max_create_attempts"`

	// CreateRetryBackoff is the initial delay between create attempts. It
	// grows exponentially (doubling) up to 60s on each retry.
	CreateRetryBackoff Duration `json:"create_retry_backoff"`
}

// Group holds the instance-group-scoped settings.
type Group struct {
	// Name is the instance group name, used as the server name prefix.
	Name string `json:"name"`

	// IdleTime is the baseline interval for the plugin's polling/retry loops.
	IdleTime Duration `json:"idle_time"`

	// MaxInstances is the maximum number of instances the plugin reports as
	// its capacity (ProviderInfo.MaxSize).
	MaxInstances int `json:"max_instances"`
}

// Duration is a time.Duration that unmarshals from a Go duration string
// (for example "1m" or "30s") as well as from a bare number of nanoseconds.
type Duration time.Duration

// Duration returns the value as a time.Duration.
func (d Duration) Duration() time.Duration { return time.Duration(d) }

// UnmarshalJSON implements json.Unmarshaler.
func (d *Duration) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err == nil {
		if s == "" {
			*d = 0
			return nil
		}
		v, err := time.ParseDuration(s)
		if err != nil {
			return fmt.Errorf("parse duration %q: %w", s, err)
		}
		*d = Duration(v)
		return nil
	}

	var n int64
	if err := json.Unmarshal(b, &n); err != nil {
		return fmt.Errorf("duration must be a string or an integer number of nanoseconds")
	}
	*d = Duration(n)
	return nil
}

// Default returns a copy of the configuration with defaults applied.
func (c Config) Default() Config {
	if c.OpenStack.ComputeMicroversion == "" {
		c.OpenStack.ComputeMicroversion = defaultComputeMicroversion
	}
	if c.OpenStack.PreemptibleTag == "" {
		c.OpenStack.PreemptibleTag = defaultPreemptibleTag
	}
	if c.OpenStack.ImageID == "" && c.OpenStack.ImageName == "" {
		c.OpenStack.ImageName = defaultImageName
	}
	if c.OpenStack.SSHUser == "" {
		c.OpenStack.SSHUser = defaultSSHUser
	}
	if c.InstanceGroup.Name == "" {
		c.InstanceGroup.Name = defaultInstanceGroupName
	}
	if c.InstanceGroup.MaxInstances <= 0 {
		c.InstanceGroup.MaxInstances = defaultMaxInstances
	}
	if c.OpenStack.MaxCreateAttempts == 0 {
		c.OpenStack.MaxCreateAttempts = defaultMaxCreateAttempts
	}
	if c.OpenStack.CreateRetryBackoff == 0 {
		c.OpenStack.CreateRetryBackoff = Duration(defaultCreateRetryBackoff)
	}
	return c
}

// Validate reports configuration errors.
func (c Config) Validate() error {
	if c.OpenStack.FlavorID == "" {
		return fmt.Errorf("openstack.flavor_id is required")
	}
	if c.OpenStack.ImageID == "" && c.OpenStack.ImageName == "" {
		return fmt.Errorf("openstack.image_id or openstack.image_name is required")
	}
	if c.OpenStack.AssignFloatingIP && c.OpenStack.FloatingIPNetwork == "" {
		return fmt.Errorf("openstack.floating_ip_network is required when openstack.assign_floating_ip is true")
	}
	if c.InstanceGroup.Name == "" {
		return fmt.Errorf("instance_group.name is required")
	}
	if c.OpenStack.MaxCreateAttempts < 1 || c.OpenStack.MaxCreateAttempts > 10 {
		return fmt.Errorf("openstack.max_create_attempts must be between 1 and 10")
	}
	if c.OpenStack.CreateRetryBackoff.Duration() < 0 {
		return fmt.Errorf("openstack.create_retry_backoff must not be negative")
	}
	return nil
}
