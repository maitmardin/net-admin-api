package server

import (
	"fmt"
	"net/http"
)

func (s *Server) HandleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Cache-Control", "no-cache")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "OK")
}
