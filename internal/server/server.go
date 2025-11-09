package server

import (
	"fmt"
	"net/http"
	"time"

	"net-admin-api/internal/vlan"
)

// Network Administration API server.
type Server struct {
	port      int
	vlanStore *vlan.Store
}

func NewServer(port int, vlanStorePath string) (*http.Server, error) {
	vlanStore, err := vlan.NewStore(vlanStorePath)
	if err != nil {
		return nil, err
	}
	server := &Server{
		port:      port,
		vlanStore: vlanStore,
	}

	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", server.port),
		Handler:      server.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return httpServer, nil
}
