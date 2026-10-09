package plugin

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/trofimovaa/gitlab-runner-fleeting-openstack/internal/openstack"

	"gitlab.com/gitlab-org/fleeting/fleeting/provider"
)

// Increase requests n more instances and returns how many were created.
func (g *InstanceGroup) Increase(ctx context.Context, n int) (int, error) {
	g.log.Info("increasing instance group", "count", n)

	succeeded := 0
	for i := 0; i < n; i++ {
		if err := g.createOne(ctx); err != nil {
			g.log.Error("creating instance", "err", err)
			return succeeded, err
		}
		succeeded++
	}

	return succeeded, nil
}

// createOne provisions a single server from the baked image and records it.
func (g *InstanceGroup) createOne(ctx context.Context) error {
	name := fmt.Sprintf("%s-%s", g.InstanceGroup.Name, randomSuffix())

	srv, err := g.client.CreateServer(ctx, openstack.CreateOpts{
		Name:              name,
		ImageID:           g.imageID,
		FlavorID:          g.OpenStack.FlavorID,
		KeypairName:       g.keypairName,
		AvailabilityZone:  g.OpenStack.AvailabilityZone,
		Networks:          g.OpenStack.Networks,
		SecurityGroups:    g.OpenStack.SecurityGroups,
		Tags:              []string{g.OpenStack.PreemptibleTag},
		AssignFloatingIP:  g.OpenStack.AssignFloatingIP,
		FloatingIPNetwork: g.OpenStack.FloatingIPNetwork,
	})
	if err != nil {
		return err
	}

	g.mu.Lock()
	g.instances[srv.ID] = &instance{
		id:      srv.ID,
		name:    srv.Name,
		address: srv.Address,
		status:  provider.StateRunning,
	}
	g.mu.Unlock()

	return nil
}

// randomSuffix returns a short random hex string used to make server names
// unique.
func randomSuffix() string {
	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		return "unknown"
	}
	return hex.EncodeToString(b)
}
