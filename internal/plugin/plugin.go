// Package plugin implements the fleeting provider.InstanceGroup interface for
// OpenStack. It is the gRPC plugin served to GitLab Runner's docker-autoscaler
// executor.
package plugin

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"path"
	"sync"

	"github.com/hashicorp/go-hclog"
	"golang.org/x/crypto/ssh"

	"github.com/aatrofimov/gitlab-runner-fleeting-openstack/config"
	"github.com/aatrofimov/gitlab-runner-fleeting-openstack/internal/openstack"

	"gitlab.com/gitlab-org/fleeting/fleeting/provider"
)

// openstackClient is the subset of the OpenStack wrapper the plugin depends
// on. It is implemented by *openstack.Client and faked in tests.
type openstackClient interface {
	CreateServer(ctx context.Context, opts openstack.CreateOpts) (*openstack.Server, error)
	GetServer(ctx context.Context, id string) (*openstack.Server, error)
	DeleteServer(ctx context.Context, id string) error
	ListServers(ctx context.Context, namePrefix string) ([]openstack.Server, error)
	ResolveImageID(ctx context.Context, imageName string) (string, error)
	EnsureKeypair(ctx context.Context, name, publicKey string) error
	DeleteKeypair(ctx context.Context, name string) error
}

// newClient is the OpenStack client factory, overridable in tests.
var newClient = func(ctx context.Context, opts openstack.Options) (openstackClient, error) {
	return openstack.New(ctx, opts)
}

// instance tracks a single server created by this plugin process.
type instance struct {
	id      string
	name    string
	address string
	status  provider.State
}

// InstanceGroup implements provider.InstanceGroup.
type InstanceGroup struct {
	// OpenStack and InstanceGroup are populated from the JSON configuration
	// the runner passes via [runners.autoscaler.plugin_config].
	OpenStack     config.OpenStack `json:"openstack"`
	InstanceGroup config.Group     `json:"instance_group"`

	log         hclog.Logger
	settings    provider.Settings
	client      openstackClient
	imageID     string
	keypairName string
	privateKey  []byte

	mu        sync.Mutex
	instances map[string]*instance
}

var _ provider.InstanceGroup = (*InstanceGroup)(nil)

// Init prepares the plugin: it authenticates, resolves the image, generates
// the ephemeral SSH keypair and imports its public half into OpenStack.
func (g *InstanceGroup) Init(ctx context.Context, log hclog.Logger, settings provider.Settings) (provider.ProviderInfo, error) {
	cfg := (config.Config{OpenStack: g.OpenStack, InstanceGroup: g.InstanceGroup}).Default()
	if err := cfg.Validate(); err != nil {
		return provider.ProviderInfo{}, err
	}
	g.OpenStack = cfg.OpenStack
	g.InstanceGroup = cfg.InstanceGroup

	g.log = log.With("region", cfg.OpenStack.Region, "name", cfg.InstanceGroup.Name)
	g.settings = settings
	g.instances = make(map[string]*instance)

	client, err := newClient(ctx, openstack.Options{
		Cloud:        cfg.OpenStack.Cloud,
		AuthURL:      cfg.OpenStack.AuthURL,
		Region:       cfg.OpenStack.Region,
		Microversion: cfg.OpenStack.ComputeMicroversion,
	})
	if err != nil {
		return provider.ProviderInfo{}, err
	}
	g.client = client

	if cfg.OpenStack.ImageID != "" {
		g.imageID = cfg.OpenStack.ImageID
	} else {
		id, err := client.ResolveImageID(ctx, cfg.OpenStack.ImageName)
		if err != nil {
			return provider.ProviderInfo{}, fmt.Errorf("resolving image %q: %w", cfg.OpenStack.ImageName, err)
		}
		g.imageID = id
	}

	g.keypairName = cfg.OpenStack.KeypairName
	if g.keypairName == "" {
		g.keypairName = cfg.InstanceGroup.Name + "-key"
	}

	privatePEM, publicSSH, err := generateKeypair()
	if err != nil {
		return provider.ProviderInfo{}, err
	}
	if err := client.EnsureKeypair(ctx, g.keypairName, publicSSH); err != nil {
		return provider.ProviderInfo{}, fmt.Errorf("ensuring keypair: %w", err)
	}
	g.privateKey = privatePEM

	id := path.Join("openstack", cfg.OpenStack.Region, cfg.InstanceGroup.Name)
	return provider.ProviderInfo{
		ID:      id,
		MaxSize: cfg.InstanceGroup.MaxInstances,
	}, nil
}

// Shutdown deletes any servers this plugin still tracks and removes the
// ephemeral keypair.
func (g *InstanceGroup) Shutdown(ctx context.Context) error {
	g.mu.Lock()
	ids := make([]string, 0, len(g.instances))
	for id := range g.instances {
		ids = append(ids, id)
	}
	g.mu.Unlock()

	for _, id := range ids {
		if err := g.client.DeleteServer(ctx, id); err != nil {
			g.log.Error("deleting instance during shutdown", "instance", id, "err", err)
		}
	}

	if g.keypairName != "" {
		if err := g.client.DeleteKeypair(ctx, g.keypairName); err != nil {
			g.log.Error("deleting keypair during shutdown", "keypair", g.keypairName, "err", err)
		}
	}

	return nil
}

// generateKeypair creates an ephemeral ed25519 keypair, returning the private
// key in PKCS#8 PEM form and the public key in OpenSSH authorized_keys form.
func generateKeypair() (privatePEM []byte, publicSSH string, err error) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, "", fmt.Errorf("generating ed25519 key: %w", err)
	}

	der, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return nil, "", fmt.Errorf("marshaling private key: %w", err)
	}
	privatePEM = pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der})

	sshPub, err := ssh.NewPublicKey(pub)
	if err != nil {
		return nil, "", fmt.Errorf("encoding public key: %w", err)
	}
	publicSSH = string(ssh.MarshalAuthorizedKey(sshPub))

	return privatePEM, publicSSH, nil
}
