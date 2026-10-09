package openstack

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/gophercloud/gophercloud/v2"
	"github.com/gophercloud/gophercloud/v2/openstack/compute/v2/servers"
	th "github.com/gophercloud/gophercloud/v2/testhelper"
	"github.com/gophercloud/gophercloud/v2/testhelper/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestClient builds a Client whose service clients point at a fake HTTP
// server, mirroring the resource bases set by openstack.NewComputeV2 etc.
func newTestClient() (*Client, th.FakeServer) {
	fs := th.SetupHTTP()

	svc := func(resourceBase string) *gophercloud.ServiceClient {
		s := client.ServiceClient(fs)
		if resourceBase != "" {
			s.ResourceBase = fs.Endpoint() + resourceBase
		}
		return s
	}

	return &Client{
		compute: svc(""),
		network: svc("v2.0/"),
		image:   svc("v2/"),
	}, fs
}

// writeJSON encodes v as a JSON response body.
func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func TestResolveImageID(t *testing.T) {
	c, fs := newTestClient()
	defer fs.Teardown()

	fs.Mux.HandleFunc("/v2/images", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		writeJSON(w, map[string]any{
			"images": []map[string]any{
				{"id": "img-old", "name": "gitlab-builder", "status": "active", "created_at": "2026-01-01T00:00:00Z"},
				{"id": "img-new", "name": "gitlab-builder", "status": "active", "created_at": "2026-01-02T00:00:00Z"},
			},
		})
	})

	id, err := c.ResolveImageID(context.Background(), "gitlab-builder")
	require.NoError(t, err)
	assert.Equal(t, "img-new", id)
}

func TestResolveImageIDNoMatch(t *testing.T) {
	c, fs := newTestClient()
	defer fs.Teardown()

	fs.Mux.HandleFunc("/v2/images", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]any{"images": []any{}})
	})

	_, err := c.ResolveImageID(context.Background(), "gitlab-builder")
	require.Error(t, err)
}

func TestDeleteServerIdempotent(t *testing.T) {
	c, fs := newTestClient()
	defer fs.Teardown()

	fs.Mux.HandleFunc("/servers/srv-1", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		w.WriteHeader(http.StatusNotFound)
	})

	require.NoError(t, c.DeleteServer(context.Background(), "srv-1"))
}

func TestGetServer(t *testing.T) {
	c, fs := newTestClient()
	defer fs.Teardown()

	fs.Mux.HandleFunc("/servers/srv-1", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		writeJSON(w, map[string]any{
			"server": map[string]any{"id": "srv-1", "name": "builder-abc", "status": "ACTIVE"},
		})
	})

	s, err := c.GetServer(context.Background(), "srv-1")
	require.NoError(t, err)
	assert.Equal(t, "srv-1", s.ID)
	assert.Equal(t, "ACTIVE", s.Status)
}

func TestFirstFixedAddress(t *testing.T) {
	s := &servers.Server{
		Addresses: map[string]any{
			"net-a": []any{
				map[string]any{"version": 4, "addr": "10.0.0.5"},
			},
		},
	}
	assert.Equal(t, "10.0.0.5", firstFixedAddress(s))
}

func TestFirstFixedAddressEmpty(t *testing.T) {
	assert.Equal(t, "", firstFixedAddress(&servers.Server{Addresses: nil}))
}

func TestIsTransient(t *testing.T) {
	assert.True(t, isTransient(gophercloud.ErrUnexpectedResponseCode{Actual: http.StatusServiceUnavailable}))
	assert.False(t, isTransient(gophercloud.ErrUnexpectedResponseCode{Actual: http.StatusNotFound}))
	assert.False(t, isTransient(nil))
}

func TestRetryStopsOnPermanentError(t *testing.T) {
	calls := 0
	err := retry(context.Background(), 5, func() error {
		calls++
		return gophercloud.ErrUnexpectedResponseCode{Actual: http.StatusBadRequest}
	})
	require.Error(t, err)
	assert.Equal(t, 1, calls, "permanent errors must not be retried")
}
