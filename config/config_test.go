package config

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigDefault(t *testing.T) {
	cfg := Config{
		OpenStack: OpenStack{
			ImageID:  "img",
			FlavorID: "flavor",
		},
	}.Default()

	assert.Equal(t, "2.72", cfg.OpenStack.ComputeMicroversion)
	assert.Equal(t, "preemptible", cfg.OpenStack.PreemptibleTag)
	assert.Equal(t, "debian", cfg.OpenStack.SSHUser)
	assert.Equal(t, "builder", cfg.InstanceGroup.Name)
	assert.Equal(t, defaultMaxInstances, cfg.InstanceGroup.MaxInstances)
}

func TestConfigDefaultDoesNotOverrideExplicitValues(t *testing.T) {
	cfg := Config{
		OpenStack: OpenStack{
			ImageID:             "img",
			FlavorID:            "flavor",
			ComputeMicroversion: "2.90",
			PreemptibleTag:      "spot",
			SSHUser:             "ubuntu",
			ImageName:           "custom",
		},
		InstanceGroup: Group{
			Name:         "workers",
			MaxInstances: 12,
		},
	}.Default()

	assert.Equal(t, "2.90", cfg.OpenStack.ComputeMicroversion)
	assert.Equal(t, "spot", cfg.OpenStack.PreemptibleTag)
	assert.Equal(t, "ubuntu", cfg.OpenStack.SSHUser)
	assert.Equal(t, "custom", cfg.OpenStack.ImageName)
	assert.Equal(t, "workers", cfg.InstanceGroup.Name)
	assert.Equal(t, 12, cfg.InstanceGroup.MaxInstances)
}

func TestConfigValidate(t *testing.T) {
	valid := func() Config {
		return (Config{
			OpenStack: OpenStack{
				ImageID:  "img",
				FlavorID: "flavor",
			},
			InstanceGroup: Group{Name: "builder"},
		}).Default()
	}

	t.Run("valid", func(t *testing.T) {
		require.NoError(t, valid().Validate())
	})

	t.Run("image name instead of id", func(t *testing.T) {
		cfg := valid()
		cfg.OpenStack.ImageID = ""
		cfg.OpenStack.ImageName = "gitlab-builder"
		require.NoError(t, cfg.Validate())
	})

	t.Run("missing flavor", func(t *testing.T) {
		cfg := valid()
		cfg.OpenStack.FlavorID = ""
		require.ErrorContains(t, cfg.Validate(), "flavor_id")
	})

	t.Run("missing image", func(t *testing.T) {
		cfg := valid()
		cfg.OpenStack.ImageID = ""
		cfg.OpenStack.ImageName = ""
		require.ErrorContains(t, cfg.Validate(), "image_id or openstack.image_name")
	})

	t.Run("floating ip without network", func(t *testing.T) {
		cfg := valid()
		cfg.OpenStack.AssignFloatingIP = true
		require.ErrorContains(t, cfg.Validate(), "floating_ip_network")
	})
}

func TestDurationUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		want    Duration
		wantErr bool
	}{
		{name: "minutes", raw: `"1m"`, want: Duration(1_000_000_000 * 60)},
		{name: "seconds", raw: `"30s"`, want: Duration(30_000_000_000)},
		{name: "empty", raw: `""`, want: 0},
		{name: "nanoseconds", raw: `1500000000`, want: Duration(1_500_000_000)},
		{name: "invalid", raw: `"nope"`, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var d Duration
			err := json.Unmarshal([]byte(tt.raw), &d)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, d)
		})
	}
}
