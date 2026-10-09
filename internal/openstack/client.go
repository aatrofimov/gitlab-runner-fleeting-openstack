// Package openstack is a thin wrapper over gophercloud that exposes exactly the
// operations the fleeting plugin needs: authenticate, create/get/delete/list
// servers, resolve an image name to an ID, manage the ephemeral keypair and
// attach floating IPs.
package openstack

import (
	"context"
	"fmt"

	"github.com/gophercloud/gophercloud/v2"
	"github.com/gophercloud/gophercloud/v2/openstack"
	"github.com/gophercloud/utils/v2/openstack/clientconfig"
)

// Options configures the OpenStack client.
type Options struct {
	// Cloud is the named cloud in clouds.yaml. When empty, authentication
	// falls back to the OS_* environment variables.
	Cloud string

	// AuthURL optionally overrides the identity endpoint.
	AuthURL string

	// Region optionally overrides the region.
	Region string

	// Microversion is the Nova microversion requested for compute calls
	// (e.g. "2.72" for Selectel preemptible servers).
	Microversion string
}

// Client wraps the gophercloud service clients needed by the plugin.
type Client struct {
	compute *gophercloud.ServiceClient
	network *gophercloud.ServiceClient
	image   *gophercloud.ServiceClient
}

// New authenticates against OpenStack and returns a ready-to-use Client.
func New(ctx context.Context, opts Options) (*Client, error) {
	provider, err := authenticate(ctx, opts)
	if err != nil {
		return nil, err
	}

	compute, err := openstack.NewComputeV2(provider, gophercloud.EndpointOpts{Region: opts.Region})
	if err != nil {
		return nil, fmt.Errorf("creating compute client: %w", err)
	}
	if opts.Microversion != "" {
		compute.Microversion = opts.Microversion
	}

	network, err := openstack.NewNetworkV2(provider, gophercloud.EndpointOpts{Region: opts.Region})
	if err != nil {
		return nil, fmt.Errorf("creating network client: %w", err)
	}

	image, err := openstack.NewImageV2(provider, gophercloud.EndpointOpts{Region: opts.Region})
	if err != nil {
		return nil, fmt.Errorf("creating image client: %w", err)
	}

	return &Client{compute: compute, network: network, image: image}, nil
}

// authenticate resolves authentication options from clouds.yaml / OS_* env and
// returns an authenticated provider client. An explicit AuthURL overrides the
// resolved identity endpoint.
func authenticate(ctx context.Context, opts Options) (*gophercloud.ProviderClient, error) {
	clientOpts := &clientconfig.ClientOpts{
		Cloud:      opts.Cloud,
		RegionName: opts.Region,
	}

	if opts.AuthURL != "" {
		authOpts, err := clientconfig.AuthOptions(clientOpts)
		if err != nil {
			return nil, fmt.Errorf("resolving auth options: %w", err)
		}
		authOpts.IdentityEndpoint = opts.AuthURL
		return openstack.AuthenticatedClient(ctx, *authOpts)
	}

	provider, err := clientconfig.AuthenticatedClient(ctx, clientOpts)
	if err != nil {
		return nil, fmt.Errorf("authenticating: %w", err)
	}
	return provider, nil
}
