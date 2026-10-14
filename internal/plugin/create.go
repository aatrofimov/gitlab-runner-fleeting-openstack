package plugin

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/aatrofimov/gitlab-runner-fleeting-openstack/internal/openstack"

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

// maxCreateBackoff caps the exponential backoff between server-create attempts.
const maxCreateBackoff = 60 * time.Second

// createOne provisions a single server from the baked image and records it,
// retrying up to OpenStack.MaxCreateAttempts times with exponential backoff
// when a server fails to become usable.
func (g *InstanceGroup) createOne(ctx context.Context) error {
	attempts := g.OpenStack.MaxCreateAttempts
	backoff := g.OpenStack.CreateRetryBackoff.Duration()

	var err error
	for attempt := 1; attempt <= attempts; attempt++ {
		err = g.createOnce(ctx)
		if err == nil {
			return nil
		}

		if attempt == attempts {
			break
		}

		g.log.Warn("creating instance failed, retrying",
			"attempt", attempt,
			"max_attempts", attempts,
			"backoff", backoff.String(),
			"err", err,
		)

		if !waitBackoff(ctx, backoff) {
			return ctx.Err()
		}

		if backoff < maxCreateBackoff {
			backoff *= 2
			if backoff > maxCreateBackoff {
				backoff = maxCreateBackoff
			}
		}
	}

	return err
}

// createOnce performs a single server-create attempt.
func (g *InstanceGroup) createOnce(ctx context.Context) error {
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
		UserData:          buildCloudInitUserData(g.publicKey),
		ConfigDrive:       true,
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

// waitBackoff blocks for d, returning false early if ctx is canceled.
func waitBackoff(ctx context.Context, d time.Duration) bool {
	if d <= 0 {
		select {
		case <-ctx.Done():
			return false
		default:
			return true
		}
	}

	timer := time.NewTimer(d)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
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
