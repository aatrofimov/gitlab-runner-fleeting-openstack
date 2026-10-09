package plugin

import (
	sdkplugin "gitlab.com/gitlab-org/fleeting/fleeting/plugin"
)

// Build-time values, overridable via -ldflags (see .goreleaser.yml).
var (
	NAME      = "gitlab-runner-fleeting-openstack"
	VERSION   = "dev"
	REVISION  = "HEAD"
	REFERENCE = "HEAD"
	BUILT     = "now"

	// Version is the plugin's version information.
	Version sdkplugin.VersionInfo
)

func init() {
	Version = sdkplugin.VersionInfo{
		Name:      NAME,
		Version:   VERSION,
		Revision:  REVISION,
		Reference: REFERENCE,
		BuiltAt:   BUILT,
	}
}
