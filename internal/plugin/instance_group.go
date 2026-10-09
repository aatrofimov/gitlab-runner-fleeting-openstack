package plugin

import (
	"context"

	"gitlab.com/gitlab-org/fleeting/fleeting/provider"
)

// Update reconciles the plugin's tracked instances with the cloud, calling
// update for every tracked instance with its current provider.State.
func (g *InstanceGroup) Update(ctx context.Context, update func(id string, state provider.State)) error {
	servers, err := g.client.ListServers(ctx, g.InstanceGroup.Name)
	if err != nil {
		return err
	}

	cloudState := make(map[string]provider.State, len(servers))
	for _, s := range servers {
		cloudState[s.ID] = novaToProviderState(s.Status)
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	for id, inst := range g.instances {
		if state, ok := cloudState[id]; ok {
			inst.status = state
			update(id, state)
			continue
		}

		// The server vanished from the cloud (preempted or deleted out of
		// band): report it as deleted and stop tracking it.
		update(id, provider.StateDeleted)
		delete(g.instances, id)
	}

	return nil
}

// novaToProviderState maps a Nova server status to a provider.State.
func novaToProviderState(status string) provider.State {
	switch status {
	case "ACTIVE":
		return provider.StateRunning
	case "BUILD", "BUILDING", "REBUILD", "REBOOT", "HARD_REBOOT", "MIGRATING", "RESIZE", "VERIFY_RESIZE", "REVERT_RESIZE":
		return provider.StateCreating
	case "ERROR", "FAILED", "SHUTOFF", "DELETED", "SOFT_DELETED":
		return provider.StateDeleted
	default:
		return provider.StateCreating
	}
}
