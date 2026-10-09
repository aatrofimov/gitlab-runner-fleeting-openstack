package plugin

import (
	"context"
)

// Decrease deletes the given instances and returns the IDs that were
// successfully requested for removal.
func (g *InstanceGroup) Decrease(ctx context.Context, instances []string) ([]string, error) {
	succeeded := make([]string, 0, len(instances))

	for _, id := range instances {
		if err := g.client.DeleteServer(ctx, id); err != nil {
			g.log.Error("deleting instance", "instance", id, "err", err)
			continue
		}

		g.mu.Lock()
		delete(g.instances, id)
		g.mu.Unlock()

		succeeded = append(succeeded, id)
	}

	return succeeded, nil
}
