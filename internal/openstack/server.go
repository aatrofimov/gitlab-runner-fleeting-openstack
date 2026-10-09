package openstack

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"sort"
	"time"

	"github.com/gophercloud/gophercloud/v2"
	"github.com/gophercloud/gophercloud/v2/openstack/compute/v2/keypairs"
	"github.com/gophercloud/gophercloud/v2/openstack/compute/v2/servers"
	"github.com/gophercloud/gophercloud/v2/openstack/image/v2/images"
	"github.com/gophercloud/gophercloud/v2/openstack/networking/v2/extensions/layer3/floatingips"
	"github.com/gophercloud/gophercloud/v2/openstack/networking/v2/networks"
	"github.com/gophercloud/gophercloud/v2/openstack/networking/v2/ports"
)

// Server is a minimal view of a Nova server for the plugin.
type Server struct {
	ID      string
	Name    string
	Status  string
	Address string
}

// CreateOpts describes a server to create.
type CreateOpts struct {
	Name              string
	ImageID           string
	FlavorID          string
	KeypairName       string
	AvailabilityZone  string
	Networks          []string
	SecurityGroups    []string
	Tags              []string
	AssignFloatingIP  bool
	FloatingIPNetwork string
}

const (
	defaultCreateTimeout = 10 * time.Minute
	maxRetries           = 5
)

// CreateServer creates a server and waits for it to become ACTIVE. When
// AssignFloatingIP is set, a floating IP is allocated and attached; the
// returned Server.Address is then the floating IP, otherwise the first fixed
// IPv4 address.
func (c *Client) CreateServer(ctx context.Context, opts CreateOpts) (*Server, error) {
	createOpts := servers.CreateOpts{
		Name:             opts.Name,
		ImageRef:         opts.ImageID,
		FlavorRef:        opts.FlavorID,
		SecurityGroups:   opts.SecurityGroups,
		AvailabilityZone: opts.AvailabilityZone,
		Tags:             opts.Tags,
	}

	if len(opts.Networks) > 0 {
		nets := make([]servers.Network, 0, len(opts.Networks))
		for _, netID := range opts.Networks {
			nets = append(nets, servers.Network{UUID: netID})
		}
		createOpts.Networks = nets
	}

	var builder servers.CreateOptsBuilder = createOpts
	if opts.KeypairName != "" {
		builder = &keypairs.CreateOptsExt{
			CreateOptsBuilder: createOpts,
			KeyName:           opts.KeypairName,
		}
	}

	var server *servers.Server
	err := retry(ctx, maxRetries, func() error {
		s, err := servers.Create(ctx, c.compute, builder, nil).Extract()
		if err != nil {
			return err
		}
		server = s
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("creating server: %w", err)
	}

	s, err := c.waitForStatus(ctx, server.ID, "ACTIVE", defaultCreateTimeout)
	if err != nil {
		// Best-effort cleanup of a server that never became usable.
		_ = c.DeleteServer(context.Background(), server.ID)
		return nil, err
	}

	result := &Server{ID: s.ID, Name: s.Name, Status: s.Status}

	if opts.AssignFloatingIP {
		addr, err := c.assignFloatingIP(ctx, s.ID, opts.FloatingIPNetwork)
		if err != nil {
			_ = c.DeleteServer(context.Background(), s.ID)
			return nil, fmt.Errorf("assigning floating ip: %w", err)
		}
		result.Address = addr
	} else {
		result.Address = firstFixedAddress(s)
	}

	return result, nil
}

// GetServer returns the current state of a server.
func (c *Client) GetServer(ctx context.Context, id string) (*Server, error) {
	var server *servers.Server
	err := retry(ctx, maxRetries, func() error {
		s, err := servers.Get(ctx, c.compute, id).Extract()
		if err != nil {
			return err
		}
		server = s
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("getting server %s: %w", id, err)
	}

	return &Server{ID: server.ID, Name: server.Name, Status: server.Status, Address: firstFixedAddress(server)}, nil
}

// DeleteServer deletes a server. Deleting an already-deleted server is treated
// as a success (idempotent).
func (c *Client) DeleteServer(ctx context.Context, id string) error {
	err := retry(ctx, maxRetries, func() error {
		return servers.Delete(ctx, c.compute, id).ExtractErr()
	})
	if err != nil && !gophercloud.ResponseCodeIs(err, http.StatusNotFound) {
		return fmt.Errorf("deleting server %s: %w", id, err)
	}
	return nil
}

// ListServers lists the servers whose name matches the given prefix.
func (c *Client) ListServers(ctx context.Context, namePrefix string) ([]Server, error) {
	allPages, err := servers.List(c.compute, servers.ListOpts{Name: namePrefix}).AllPages(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing servers: %w", err)
	}

	all, err := servers.ExtractServers(allPages)
	if err != nil {
		return nil, fmt.Errorf("extracting servers: %w", err)
	}

	result := make([]Server, 0, len(all))
	for _, s := range all {
		result = append(result, Server{ID: s.ID, Name: s.Name, Status: s.Status, Address: firstFixedAddress(&s)})
	}
	return result, nil
}

// ResolveImageID resolves an image name to the most recently created image ID.
func (c *Client) ResolveImageID(ctx context.Context, imageName string) (string, error) {
	allPages, err := images.List(c.image, images.ListOpts{
		Name:   imageName,
		Status: images.ImageStatusActive,
	}).AllPages(ctx)
	if err != nil {
		return "", fmt.Errorf("listing images: %w", err)
	}

	all, err := images.ExtractImages(allPages)
	if err != nil {
		return "", fmt.Errorf("extracting images: %w", err)
	}

	if len(all) == 0 {
		return "", fmt.Errorf("no active image found with name %q", imageName)
	}

	sort.Slice(all, func(i, j int) bool {
		return all[i].CreatedAt.After(all[j].CreatedAt)
	})

	return all[0].ID, nil
}

// EnsureKeypair imports the provided public key as an OpenStack keypair. If a
// keypair with the same name already exists it is replaced, so the public key
// always matches the plugin's freshly generated private key.
func (c *Client) EnsureKeypair(ctx context.Context, name, publicKey string) error {
	create := func() error {
		_, err := keypairs.Create(ctx, c.compute, keypairs.CreateOpts{
			Name:      name,
			PublicKey: publicKey,
			Type:      "ssh",
		}).Extract()
		return err
	}

	if err := create(); err == nil {
		return nil
	} else if !gophercloud.ResponseCodeIs(err, http.StatusConflict) {
		return fmt.Errorf("creating keypair: %w", err)
	}

	// Conflict: replace the existing keypair to refresh the public key.
	if err := keypairs.Delete(ctx, c.compute, name, nil).ExtractErr(); err != nil &&
		!gophercloud.ResponseCodeIs(err, http.StatusNotFound) {
		return fmt.Errorf("deleting existing keypair: %w", err)
	}
	if err := create(); err != nil {
		return fmt.Errorf("recreating keypair: %w", err)
	}
	return nil
}

// DeleteKeypair removes a keypair. Deleting a missing keypair is idempotent.
func (c *Client) DeleteKeypair(ctx context.Context, name string) error {
	if err := keypairs.Delete(ctx, c.compute, name, nil).ExtractErr(); err != nil &&
		!gophercloud.ResponseCodeIs(err, http.StatusNotFound) {
		return fmt.Errorf("deleting keypair %s: %w", name, err)
	}
	return nil
}

// waitForStatus polls the server until it reaches the desired status or enters
// an error state.
func (c *Client) waitForStatus(ctx context.Context, id, want string, timeout time.Duration) (*servers.Server, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	backoff := 2 * time.Second
	for {
		s, err := servers.Get(ctx, c.compute, id).Extract()
		if err != nil {
			return nil, fmt.Errorf("getting server %s: %w", id, err)
		}

		switch s.Status {
		case want:
			return s, nil
		case "ERROR", "FAILED":
			return nil, fmt.Errorf("server %s entered %q state", id, s.Status)
		}

		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("waiting for server %s to reach %q: %w", id, want, ctx.Err())
		case <-time.After(backoff):
		}
		if backoff < 30*time.Second {
			backoff *= 2
		}
	}
}

// assignFloatingIP allocates a floating IP on the given external network and
// attaches it to the server's first port.
func (c *Client) assignFloatingIP(ctx context.Context, serverID, externalNetwork string) (string, error) {
	networkID, err := c.resolveNetworkID(ctx, externalNetwork)
	if err != nil {
		return "", err
	}

	portID, err := c.firstPortID(ctx, serverID)
	if err != nil {
		return "", err
	}

	fip, err := floatingips.Create(ctx, c.network, floatingips.CreateOpts{
		FloatingNetworkID: networkID,
		PortID:            portID,
	}).Extract()
	if err != nil {
		return "", fmt.Errorf("creating floating ip: %w", err)
	}

	return fip.FloatingIP, nil
}

// resolveNetworkID resolves a network name or ID to an ID.
func (c *Client) resolveNetworkID(ctx context.Context, nameOrID string) (string, error) {
	if net, err := networks.Get(ctx, c.network, nameOrID).Extract(); err == nil {
		return net.ID, nil
	}

	allPages, err := networks.List(c.network, networks.ListOpts{Name: nameOrID}).AllPages(ctx)
	if err != nil {
		return "", fmt.Errorf("listing networks: %w", err)
	}
	all, err := networks.ExtractNetworks(allPages)
	if err != nil {
		return "", fmt.Errorf("extracting networks: %w", err)
	}
	if len(all) == 0 {
		return "", fmt.Errorf("network %q not found", nameOrID)
	}
	return all[0].ID, nil
}

// firstPortID returns the first port attached to the server.
func (c *Client) firstPortID(ctx context.Context, serverID string) (string, error) {
	allPages, err := ports.List(c.network, ports.ListOpts{DeviceID: serverID}).AllPages(ctx)
	if err != nil {
		return "", fmt.Errorf("listing ports: %w", err)
	}
	all, err := ports.ExtractPorts(allPages)
	if err != nil {
		return "", fmt.Errorf("extracting ports: %w", err)
	}
	if len(all) == 0 {
		return "", fmt.Errorf("server %s has no ports", serverID)
	}
	return all[0].ID, nil
}

// firstFixedAddress returns the first IPv4 address from a server's addresses.
func firstFixedAddress(s *servers.Server) string {
	b, err := json.Marshal(s.Addresses)
	if err != nil {
		return ""
	}
	var addrs map[string][]servers.Address
	if err := json.Unmarshal(b, &addrs); err != nil {
		return ""
	}
	for _, list := range addrs {
		for _, a := range list {
			if a.Version == 4 && a.Address != "" {
				return a.Address
			}
		}
	}
	return ""
}

// retry runs fn up to attempts times, retrying only transient errors (network
// failures and HTTP 5xx).
func retry(ctx context.Context, attempts int, fn func() error) error {
	var err error
	for i := 0; i < attempts; i++ {
		err = fn()
		if err == nil {
			return nil
		}
		if !isTransient(err) {
			return err
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Duration(i+1) * time.Second):
		}
	}
	return err
}

// isTransient reports whether err represents a retryable failure.
func isTransient(err error) bool {
	for _, code := range []int{http.StatusRequestTimeout, http.StatusTooManyRequests, http.StatusInternalServerError, http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout} {
		if gophercloud.ResponseCodeIs(err, code) {
			return true
		}
	}
	var netErr net.Error
	return errors.As(err, &netErr)
}
