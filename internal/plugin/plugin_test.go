package plugin

import (
	"context"
	"strconv"
	"strings"
	"testing"

	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aatrofimov/gitlab-runner-fleeting-openstack/config"
	"github.com/aatrofimov/gitlab-runner-fleeting-openstack/internal/openstack"

	"gitlab.com/gitlab-org/fleeting/fleeting/provider"
)

type fakeClient struct {
	createFn    func(context.Context, openstack.CreateOpts) (*openstack.Server, error)
	getFn       func(context.Context, string) (*openstack.Server, error)
	deleteFn    func(context.Context, string) error
	listFn      func(context.Context, string) ([]openstack.Server, error)
	imageFn     func(context.Context, string) (string, error)
	ensureFn    func(context.Context, string, string) error
	deleteKeyFn func(context.Context, string) error

	createCalls []openstack.CreateOpts
	deleteCalls []string
	ensureCalls []string
}

func newFakeClient() *fakeClient {
	f := &fakeClient{}
	f.createFn = func(_ context.Context, opts openstack.CreateOpts) (*openstack.Server, error) {
		return &openstack.Server{ID: "srv-" + strconv.Itoa(len(f.createCalls)), Name: opts.Name, Status: "ACTIVE", Address: "10.0.0.9"}, nil
	}
	f.getFn = func(_ context.Context, id string) (*openstack.Server, error) {
		return &openstack.Server{ID: id, Status: "ACTIVE"}, nil
	}
	f.deleteFn = func(_ context.Context, id string) error { return nil }
	f.listFn = func(_ context.Context, _ string) ([]openstack.Server, error) { return nil, nil }
	f.imageFn = func(_ context.Context, _ string) (string, error) { return "img-1", nil }
	f.ensureFn = func(_ context.Context, _ string, _ string) error { return nil }
	f.deleteKeyFn = func(_ context.Context, _ string) error { return nil }
	return f
}

func (f *fakeClient) CreateServer(ctx context.Context, opts openstack.CreateOpts) (*openstack.Server, error) {
	f.createCalls = append(f.createCalls, opts)
	return f.createFn(ctx, opts)
}

func (f *fakeClient) GetServer(ctx context.Context, id string) (*openstack.Server, error) {
	return f.getFn(ctx, id)
}

func (f *fakeClient) DeleteServer(ctx context.Context, id string) error {
	f.deleteCalls = append(f.deleteCalls, id)
	return f.deleteFn(ctx, id)
}

func (f *fakeClient) ListServers(ctx context.Context, namePrefix string) ([]openstack.Server, error) {
	return f.listFn(ctx, namePrefix)
}

func (f *fakeClient) ResolveImageID(ctx context.Context, imageName string) (string, error) {
	return f.imageFn(ctx, imageName)
}

func (f *fakeClient) EnsureKeypair(ctx context.Context, name, publicKey string) error {
	f.ensureCalls = append(f.ensureCalls, name)
	return f.ensureFn(ctx, name, publicKey)
}

func (f *fakeClient) DeleteKeypair(ctx context.Context, name string) error {
	return f.deleteKeyFn(ctx, name)
}

func newTestGroup(t *testing.T, fake *fakeClient, cfg config.Config) *InstanceGroup {
	t.Helper()

	old := newClient
	t.Cleanup(func() { newClient = old })
	newClient = func(context.Context, openstack.Options) (openstackClient, error) {
		return fake, nil
	}

	g := &InstanceGroup{OpenStack: cfg.OpenStack, InstanceGroup: cfg.InstanceGroup}
	_, err := g.Init(context.Background(), hclog.NewNullLogger(), provider.Settings{})
	require.NoError(t, err)
	return g
}

func validConfig() config.Config {
	return (config.Config{
		OpenStack: config.OpenStack{
			ImageID:  "img-1",
			FlavorID: "flavor-1",
		},
		InstanceGroup: config.Group{Name: "builder", MaxInstances: 5},
	}).Default()
}

func TestInit(t *testing.T) {
	fake := newFakeClient()
	g := newTestGroup(t, fake, validConfig())

	assert.Equal(t, "img-1", g.imageID)
	assert.Equal(t, "builder-key", g.keypairName)
	assert.NotEmpty(t, g.privateKey)
	assert.Len(t, fake.ensureCalls, 1)
	assert.Equal(t, "builder-key", fake.ensureCalls[0])
}

func TestInitResolvesImageName(t *testing.T) {
	fake := newFakeClient()
	cfg := validConfig()
	cfg.OpenStack.ImageID = ""
	cfg.OpenStack.ImageName = "gitlab-builder"

	g := newTestGroup(t, fake, cfg)
	assert.Equal(t, "img-1", g.imageID)
}

func TestInitValidationError(t *testing.T) {
	fake := newFakeClient()
	cfg := validConfig()
	cfg.OpenStack.FlavorID = ""

	old := newClient
	t.Cleanup(func() { newClient = old })
	newClient = func(context.Context, openstack.Options) (openstackClient, error) { return fake, nil }

	g := &InstanceGroup{OpenStack: cfg.OpenStack, InstanceGroup: cfg.InstanceGroup}
	_, err := g.Init(context.Background(), hclog.NewNullLogger(), provider.Settings{})
	require.Error(t, err)
}

func TestIncrease(t *testing.T) {
	fake := newFakeClient()
	g := newTestGroup(t, fake, validConfig())

	n, err := g.Increase(context.Background(), 3)
	require.NoError(t, err)
	assert.Equal(t, 3, n)
	assert.Len(t, fake.createCalls, 3)

	assert.Equal(t, []string{g.OpenStack.PreemptibleTag}, fake.createCalls[0].Tags)
	assert.Equal(t, g.keypairName, fake.createCalls[0].KeypairName)
	assert.Equal(t, "img-1", fake.createCalls[0].ImageID)
	assert.True(t, fake.createCalls[0].ConfigDrive)
	assert.Contains(t, string(fake.createCalls[0].UserData), g.publicKey)

	g.mu.Lock()
	assert.Len(t, g.instances, 3)
	g.mu.Unlock()
}

func TestDecrease(t *testing.T) {
	fake := newFakeClient()
	g := newTestGroup(t, fake, validConfig())

	_, err := g.Increase(context.Background(), 2)
	require.NoError(t, err)

	succeeded, err := g.Decrease(context.Background(), []string{"srv-1", "srv-2"})
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"srv-1", "srv-2"}, succeeded)
	assert.ElementsMatch(t, []string{"srv-1", "srv-2"}, fake.deleteCalls)

	g.mu.Lock()
	assert.Empty(t, g.instances)
	g.mu.Unlock()
}

func TestUpdateReconciliation(t *testing.T) {
	fake := newFakeClient()
	g := newTestGroup(t, fake, validConfig())

	_, err := g.Increase(context.Background(), 2)
	require.NoError(t, err)

	fake.listFn = func(_ context.Context, _ string) ([]openstack.Server, error) {
		return []openstack.Server{
			{ID: "srv-1", Status: "ACTIVE"},
		}, nil
	}

	states := map[string]provider.State{}
	err = g.Update(context.Background(), func(id string, state provider.State) {
		states[id] = state
	})
	require.NoError(t, err)

	assert.Equal(t, provider.StateRunning, states["srv-1"])
	assert.Equal(t, provider.StateDeleted, states["srv-2"], "vanished instance must be reported deleted")
}

func TestConnectInfo(t *testing.T) {
	fake := newFakeClient()
	cfg := validConfig()
	g := newTestGroup(t, fake, cfg)

	_, err := g.Increase(context.Background(), 1)
	require.NoError(t, err)

	info, err := g.ConnectInfo(context.Background(), "srv-1")
	require.NoError(t, err)

	assert.Equal(t, "srv-1", info.ID)
	assert.Equal(t, "10.0.0.9", info.ExternalAddr)
	assert.Equal(t, provider.ProtocolSSH, info.Protocol)
	assert.Equal(t, 22, info.ProtocolPort)
	assert.Equal(t, g.OpenStack.SSHUser, info.Username)
	assert.Equal(t, g.privateKey, info.Key)
}

func TestConnectInfoUnknown(t *testing.T) {
	fake := newFakeClient()
	g := newTestGroup(t, fake, validConfig())

	_, err := g.ConnectInfo(context.Background(), "missing")
	require.Error(t, err)
}

func TestHeartbeat(t *testing.T) {
	fake := newFakeClient()
	g := newTestGroup(t, fake, validConfig())

	require.NoError(t, g.Heartbeat(context.Background(), "srv-1"))

	fake.getFn = func(_ context.Context, id string) (*openstack.Server, error) {
		return &openstack.Server{ID: id, Status: "ERROR"}, nil
	}
	require.ErrorIs(t, g.Heartbeat(context.Background(), "srv-1"), provider.ErrInstanceUnhealthy)
}

func TestSuspendResumeUnsupported(t *testing.T) {
	fake := newFakeClient()
	g := newTestGroup(t, fake, validConfig())

	_, err := g.Suspend(context.Background(), []string{"srv-1"})
	require.ErrorIs(t, err, provider.ErrSuspendResumeNotSupported)

	_, err = g.Resume(context.Background(), []string{"srv-1"})
	require.ErrorIs(t, err, provider.ErrSuspendResumeNotSupported)
}

func TestShutdown(t *testing.T) {
	fake := newFakeClient()
	g := newTestGroup(t, fake, validConfig())

	_, err := g.Increase(context.Background(), 1)
	require.NoError(t, err)

	require.NoError(t, g.Shutdown(context.Background()))
	assert.ElementsMatch(t, []string{"srv-1"}, fake.deleteCalls)
}

func TestBuildCloudInitUserData(t *testing.T) {
	ud := string(buildCloudInitUserData("ssh-ed25519 AAA test"))

	assert.Contains(t, ud, "ssh-ed25519 AAA test")
	assert.Contains(t, ud, "name: root")
	assert.Equal(t, 2, strings.Count(ud, "ssh-ed25519 AAA test"),
		"key must be injected into both the default user and root")
}

func TestNovaToProviderState(t *testing.T) {
	tests := []struct {
		status string
		want   provider.State
	}{
		{"ACTIVE", provider.StateRunning},
		{"BUILD", provider.StateCreating},
		{"BUILDING", provider.StateCreating},
		{"ERROR", provider.StateDeleted},
		{"SHUTOFF", provider.StateDeleted},
		{"DELETED", provider.StateDeleted},
		{"UNKNOWN", provider.StateCreating},
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			assert.Equal(t, tt.want, novaToProviderState(tt.status))
		})
	}
}
