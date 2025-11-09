package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHandleHealth(t *testing.T) {
	apiServer := &Server{}
	testServer := httptest.NewServer(http.HandlerFunc(apiServer.HandleHealth))
	t.Cleanup(testServer.Close)

	resp, err := http.Get(testServer.URL + "/health")
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)
}
