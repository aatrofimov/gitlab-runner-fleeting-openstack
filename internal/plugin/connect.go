package plugin

import (
	"context"
	"fmt"

	"gitlab.com/gitlab-org/fleeting/fleeting/provider"
)

// ConnectInfo returns the connection details needed to reach the Docker daemon
// on an instance over SSH.
func (g *InstanceGroup) ConnectInfo(ctx context.Context, id string) (provider.ConnectInfo, error) {
	g.mu.Lock()
	inst, ok := g.instances[id]
	g.mu.Unlock()
	if !ok {
		return provider.ConnectInfo{}, fmt.Errorf("instance %s not found", id)
	}

	conn := g.settings.ConnectorConfig
	conn.Key = g.privateKey
	if conn.Username == "" {
		conn.Username = g.OpenStack.SSHUser
	}
	if conn.Protocol == "" {
		conn.Protocol = provider.ProtocolSSH
	}
	if conn.ProtocolPort == 0 {
		conn.ProtocolPort = 22
	}

	return provider.ConnectInfo{
		ConnectorConfig: conn,
		ID:              id,
		ExternalAddr:    inst.address,
		InternalAddr:    inst.address,
	}, nil
}

// Heartbeat reports whether an instance is still healthy. It returns an error
// when the server is gone or has entered an unusable state (preempted).
func (g *InstanceGroup) Heartbeat(ctx context.Context, id string) error {
	srv, err := g.client.GetServer(ctx, id)
	if err != nil {
		return err
	}

	switch srv.Status {
	case "ACTIVE":
		return nil
	case "ERROR", "FAILED", "SHUTOFF", "DELETED", "SOFT_DELETED":
		return provider.ErrInstanceUnhealthy
	default:
		return nil
	}
}
