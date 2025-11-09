package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"path"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"net-admin/internal/vlan"
)

func TestHandlers_OK(t *testing.T) {
	t.Parallel()
	vlanStorePath := filepath.Join(t.TempDir(), "vlans.json")
	server := newHTTPServer(t, vlanStorePath)

	// No VLANs initially
	vlansFromAPI := readVLANs(t, server)
	require.Len(t, vlansFromAPI, 0)

	// Create one VLAN
	vlan1 := newVLAN(t, 1, "test1", "192.168.0.0/24", "192.168.0.1")
	createVLAN(t, server, vlan1)

	// Created VLAN is now in the list
	vlansFromAPI = readVLANs(t, server)
	require.Len(t, vlansFromAPI, 1)
	require.Equal(t, *vlan1, vlansFromAPI[vlan1.ID])

	// Created VLAN can be retrieved by ID
	vlan1FromAPI := readVLAN(t, server, vlan1.ID)
	require.Equal(t, *vlan1, *vlan1FromAPI)

	// Update previously created VLAN
	vlan1.Name = "a better name"
	vlan1.Status = "disabled"
	updateVLAN(t, server, vlan1)

	// Updated VLAN is now in the list
	vlansFromAPI = readVLANs(t, server)
	require.Len(t, vlansFromAPI, 1)
	require.Equal(t, *vlan1, vlansFromAPI[vlan1.ID])

	// Delete the only VLAN
	deleteVLAN(t, server, vlan1.ID)

	// No VLANs after delete
	vlansFromAPI = readVLANs(t, server)
	require.Len(t, vlansFromAPI, 0)
}

func TestServerRestart(t *testing.T) {
	t.Parallel()
	vlanStorePath := filepath.Join(t.TempDir(), "vlans.json")
	server := newHTTPServer(t, vlanStorePath)

	count := uint16(100)
	vlans := make([]*vlan.VLAN, count)
	for i := uint16(1); i <= count; i++ {
		vlan := newVLAN(t, i,
			fmt.Sprintf("test %d", i),
			fmt.Sprintf("192.168.%d.0/24", i),
			fmt.Sprintf("192.168.%d.1", i))
		createVLAN(t, server, vlan)
		vlans = append(vlans, vlan)
	}

	// 100 VLANs present
	vlansFromAPI := readVLANs(t, server)
	require.Len(t, vlansFromAPI, int(count))

	// Shut down the existing server, and create a new one
	server.Close()
	server = newHTTPServer(t, vlanStorePath)

	// Still 100 VLANs present
	vlansFromAPI = readVLANs(t, server)
	require.Len(t, vlansFromAPI, int(count))
}

func TestHandleReadVLAN_NOK(t *testing.T) {
	t.Parallel()
	vlanStorePath := filepath.Join(t.TempDir(), "vlans.json")
	server := newHTTPServer(t, vlanStorePath)

	// Attempt a read with invalid ID
	resp, err := server.Client().Get(server.URL + "/api/v1/vlans/invalid")
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, resp.StatusCode, http.StatusBadRequest)

	// Attempt a read with non-existent ID
	resp, err = server.Client().Get(server.URL + "/api/v1/vlans/" + uuid.New().String())
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, resp.StatusCode, http.StatusNotFound)
}

func TestHandleCreateVLAN_NOK(t *testing.T) {
	t.Parallel()
	vlanStorePath := filepath.Join(t.TempDir(), "vlans.json")
	server := newHTTPServer(t, vlanStorePath)

	// Create VLAN with invalid JSON
	vlan1 := newVLAN(t, 5000, "test1", "192.168.0.0/24", "192.168.0.1")
	resp, err := server.Client().Post(server.URL + "/api/v1/vlans", "application/json", bytes.NewBuffer([]byte("invalid")))
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, resp.StatusCode, http.StatusBadRequest)

	// Create VLAN with invalid VID
	vlan1 = newVLAN(t, 5000, "test1", "192.168.0.0/24", "192.168.0.1")
	resp, err = server.Client().Post(server.URL + "/api/v1/vlans", "application/json", encodeVLAN(t, vlan1))
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, resp.StatusCode, http.StatusBadRequest)
}

func TestHandleUpdateVLAN_NOK(t *testing.T) {
	t.Parallel()
	vlanStorePath := filepath.Join(t.TempDir(), "vlans.json")
	server := newHTTPServer(t, vlanStorePath)

	// Create one VLAN
	vlan1 := newVLAN(t, 1, "test1", "192.168.0.0/24", "192.168.0.1")
	vlan2 := newVLAN(t, 2, "test2", "192.168.1.0/24", "192.168.1.1")
	createVLAN(t, server, vlan1)

	// Update with invalid ID
	req, err := http.NewRequest("PUT", server.URL + "/api/v1/vlans/invalid", encodeVLAN(t, vlan1))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	resp, err := server.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, resp.StatusCode, http.StatusBadRequest)

	// Update non-existent VLAN
	req, err = http.NewRequest("PUT", server.URL + "/api/v1/vlans/" + vlan2.ID.String(), encodeVLAN(t, vlan2))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	resp, err = server.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, resp.StatusCode, http.StatusNotFound)

	// Update with a mismatching ID in request body
	req, err = http.NewRequest("PUT", server.URL + "/api/v1/vlans/" + vlan1.ID.String(), encodeVLAN(t, vlan2))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	resp, err = server.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, resp.StatusCode, http.StatusBadRequest)

	// Update with invalid JSON
	req, err = http.NewRequest("PUT", server.URL + "/api/v1/vlans/" + vlan1.ID.String(), bytes.NewBuffer([]byte("invalid")))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	resp, err = server.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, resp.StatusCode, http.StatusBadRequest)

	// Update with an invalid VID
	vlan1.VID = 9999
	req, err = http.NewRequest("PUT", server.URL + "/api/v1/vlans/" + vlan1.ID.String(), encodeVLAN(t, vlan1))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	resp, err = server.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, resp.StatusCode, http.StatusBadRequest)
}

func TestHandleDeleteVLAN_NOK(t *testing.T) {
	t.Parallel()
	vlanStorePath := filepath.Join(t.TempDir(), "vlans.json")
	server := newHTTPServer(t, vlanStorePath)

	// Create one VLAN
	vlan1 := newVLAN(t, 1, "test1", "192.168.0.0/24", "192.168.0.1")

	// Delete with invalid ID
	req, err := http.NewRequest("DELETE", server.URL + "/api/v1/vlans/invalid", nil)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	resp, err := server.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, resp.StatusCode, http.StatusBadRequest)

	// Delete non-existent VLAN
	req, err = http.NewRequest("DELETE", server.URL + "/api/v1/vlans/" + vlan1.ID.String(), nil)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	resp, err = server.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, resp.StatusCode, http.StatusNotFound)
}

func newHTTPServer(t *testing.T, vlanStorePath string) *httptest.Server {
	t.Helper()
	apiServer, err := NewServer(0, vlanStorePath)
	if err != nil {
		t.Fatalf("error creating API server: %v", err)
	}

	testServer := httptest.NewServer(apiServer.Handler)

	// Only cleanup testServer, apiServer was never started started
	t.Cleanup(testServer.Close)

	return testServer
}

func readVLANs(t *testing.T, server *httptest.Server) map[uuid.UUID]vlan.VLAN {
	t.Helper()

	resp, err := server.Client().Get(server.URL + "/api/v1/vlans")
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, resp.StatusCode, http.StatusOK)

	vlans := []vlan.VLAN{}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&vlans))

	vlansByID := make(map[uuid.UUID]vlan.VLAN, len(vlans))
	for _, vlan := range vlans {
		vlansByID[vlan.ID] = vlan
	}
	return vlansByID
}

func readVLAN(t *testing.T, server *httptest.Server, id uuid.UUID) *vlan.VLAN {
	t.Helper()

	resp, err := server.Client().Get(server.URL + "/api/v1/vlans/" + id.String())
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, resp.StatusCode, http.StatusOK)

	vlan := vlan.VLAN{}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&vlan))
	return &vlan
}

// Creates a new VLAN using the API server, updates the VLAN with ID generated assigned by API
func createVLAN(t *testing.T, server *httptest.Server, vlan *vlan.VLAN) {
	t.Helper()

	resp, err := server.Client().Post(server.URL + "/api/v1/vlans", "application/json", encodeVLAN(t, vlan))
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, resp.StatusCode, http.StatusCreated)
	vlanLocation, err := resp.Location()
	require.NoError(t, err)

	vlanIDStr := path.Base(vlanLocation.String())
	require.Len(t, vlanIDStr, 36, "unexpected UUID length")

	vlanID, err := uuid.Parse(vlanIDStr)
	require.NoError(t, err)
	vlan.ID = vlanID
}

func updateVLAN(t *testing.T, server *httptest.Server, vlan *vlan.VLAN) {
	t.Helper()

	req, err := http.NewRequest("PUT", server.URL + "/api/v1/vlans/" + vlan.ID.String(), encodeVLAN(t, vlan))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := server.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, resp.StatusCode, http.StatusOK)
	require.Equal(t, int64(0), resp.ContentLength, "unexpected content in VLAN update response")
}

func deleteVLAN(t *testing.T, server *httptest.Server, id uuid.UUID) {
	t.Helper()

	req, err := http.NewRequest("DELETE", server.URL + "/api/v1/vlans/" + id.String(), nil)
	require.NoError(t, err)

	resp, err := server.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, resp.StatusCode, http.StatusOK)
	require.Equal(t, int64(0), resp.ContentLength, "unexpected content in VLAN delete response")
}

func encodeVLAN(t *testing.T, vlan *vlan.VLAN) *bytes.Buffer {
	t.Helper()
	buf := &bytes.Buffer{}
	require.NoError(t, json.NewEncoder(buf).Encode(vlan))
	return buf
}

func newVLAN(t *testing.T, vid uint16, name, subnetStr, gatewayStr string) *vlan.VLAN {
	t.Helper()
	subnet, err := netip.ParsePrefix(subnetStr)
	require.NoError(t, err)
	gateway, err := netip.ParseAddr(gatewayStr)
	require.NoError(t, err)
	return &vlan.VLAN{
		VID:     vid,
		Name:    name,
		Subnet:  subnet,
		Gateway: gateway,
		Status:  "enabled",
	}
}
