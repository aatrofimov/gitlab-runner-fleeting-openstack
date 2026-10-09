package plugin

import (
	"context"

	"gitlab.com/gitlab-org/fleeting/fleeting/provider"
)

// Suspend is not supported for preemptible build servers.
func (g *InstanceGroup) Suspend(context.Context, []string) ([]string, error) {
	return nil, provider.ErrSuspendResumeNotSupported
}

// Resume is not supported for preemptible build servers.
func (g *InstanceGroup) Resume(context.Context, []string) ([]string, error) {
	return nil, provider.ErrSuspendResumeNotSupported
}
