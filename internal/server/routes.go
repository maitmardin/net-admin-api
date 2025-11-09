package server

import (
	"net/http"
)

func (s *Server) RegisterRoutes() http.Handler {
	mux := http.NewServeMux()

	// vlans
	mux.HandleFunc("GET    /api/v1/vlans",      s.HandleListVLANs)
	mux.HandleFunc("POST   /api/v1/vlans",      s.HandleCreateVLAN)
	mux.HandleFunc("GET    /api/v1/vlans/{id}", s.HandleReadVLAN)
	mux.HandleFunc("PUT    /api/v1/vlans/{id}", s.HandleUpdateVLAN)
	mux.HandleFunc("DELETE /api/v1/vlans/{id}", s.HandleDeleteVLAN)

	// monitoring
	mux.HandleFunc("GET /health", s.HandleHealth)

	return s.corsMiddleware(mux)
}

func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type")
		w.Header().Set("Access-Control-Allow-Credentials", "false")

		// Handle preflight OPTIONS requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// Proceed with the next handler
		next.ServeHTTP(w, r)
	})
}
