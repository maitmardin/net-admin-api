package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"net-admin/internal/vlan"

	"github.com/google/uuid"
)

func (s *Server) HandleListVLANs(respWriter http.ResponseWriter, req *http.Request) {
	vlans, err := s.vlanStore.List()
	if err != nil {
		log.Printf("failed to read vlans: %v", err)
		http.Error(respWriter, "failed to read vlans", http.StatusInternalServerError)
		return
	}
	writeJSONResponse(respWriter, vlans)
}

func (s *Server) HandleCreateVLAN(respWriter http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	vlan := &vlan.VLAN{}
	if err := json.NewDecoder(req.Body).Decode(&vlan); err != nil {
		http.Error(respWriter, fmt.Sprintf("failed to parse vlan: %v", err), http.StatusBadRequest)
		return
	}

	if errors := vlan.Validate(); len(errors) > 0 {
		http.Error(respWriter, strings.Join(errors, ", "), http.StatusBadRequest)
		return
	}

	// Generate new ID and save
	vlan.ID = uuid.New()
	if err := s.vlanStore.Save(*vlan); err != nil {
		log.Printf("failed to save vlan: %v", err)
		http.Error(respWriter, "failed to save vlan", http.StatusInternalServerError)
		return
	}

	respWriter.Header().Set("Location", fmt.Sprintf("%s/%s", req.URL.RequestURI(), vlan.ID.String()))
	respWriter.WriteHeader(http.StatusCreated)
}

func (s *Server) HandleReadVLAN(respWriter http.ResponseWriter, req *http.Request) {
	vlanID, err := uuid.Parse(req.PathValue("id"))
	if err != nil {
		http.Error(respWriter, "invalid vlan id", http.StatusBadRequest)
		return
	}
	vlan := s.vlanStore.Get(vlanID)
	if vlan == nil {
		http.NotFound(respWriter, req)
		return
	}
	writeJSONResponse(respWriter, vlan)
}

func (s *Server) HandleUpdateVLAN(respWriter http.ResponseWriter, req *http.Request) {
	vlanID, err := uuid.Parse(req.PathValue("id"))
	if err != nil {
		http.Error(respWriter, "invalid vlan id", http.StatusBadRequest)
		return
	}

	defer req.Body.Close()
	v := &vlan.VLAN{}
	if err := json.NewDecoder(req.Body).Decode(&v); err != nil {
		http.Error(respWriter, fmt.Sprintf("failed to parse vlan: %v", err), http.StatusBadRequest)
		return
	}
	if vlanID != v.ID {
		http.Error(respWriter, "mismatching vlan id in request body", http.StatusBadRequest)
		return
	}
	if errors := v.Validate(); len(errors) > 0 {
		http.Error(respWriter, strings.Join(errors, ", "), http.StatusBadRequest)
		return
	}

	if err := s.vlanStore.Update(*v); err != nil {
		if errors.Is(err, vlan.ErrNotFound) {
			http.NotFound(respWriter, req)
			return
		}

		log.Printf("failed to update vlan: %v", err)
		http.Error(respWriter, "failed to update vlan", http.StatusInternalServerError)
		return
	}
}

func (s *Server) HandleDeleteVLAN(respWriter http.ResponseWriter, req *http.Request) {
	vlanID, err := uuid.Parse(req.PathValue("id"))
	if err != nil {
		http.Error(respWriter, "invalid vlan id", http.StatusBadRequest)
		return
	}

	if err := s.vlanStore.Delete(vlanID); err != nil {
		if err == vlan.ErrNotFound {
			http.NotFound(respWriter, req)
			return
		}
		log.Printf("failed to delete vlan: %v", err)
		http.Error(respWriter, "failed to delete vlan", http.StatusInternalServerError)
		return
	}
}

func writeJSONResponse(respWriter http.ResponseWriter, v any) {
	respWriter.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(respWriter).Encode(v); err != nil {
		log.Printf("failed to write response: %v", err)
	}
}
