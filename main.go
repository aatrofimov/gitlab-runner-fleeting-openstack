// Command gitlab-runner-fleeting-openstack is the fleeting plugin binary served
// to GitLab Runner's docker-autoscaler executor.
package main

import (
	sdkplugin "gitlab.com/gitlab-org/fleeting/fleeting/plugin"

	osplugin "github.com/aatrofimov/gitlab-runner-fleeting-openstack/internal/plugin"
)

func main() {
	sdkplugin.Main(&osplugin.InstanceGroup{}, osplugin.Version)
}
