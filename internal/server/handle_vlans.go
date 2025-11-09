package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"net-admin-api/internal/vlan"
)

func (s *Server) HandleListVLANs(respWriter http.ResponseWriter, req *http.Request) {
	vlans, err := s.vlanStore.List()
	if err != nil {
		log.Printf("failed to read vlans: %v", err)
		internalError(respWriter, "failed to read vlans")
		return
	}
	writeJSONResponse(respWriter, vlans)
}

func (s *Server) HandleCreateVLAN(respWriter http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	vlan := &vlan.VLAN{}
	if err := json.NewDecoder(req.Body).Decode(&vlan); err != nil {
		invalidInput(respWriter, fmt.Sprintf("failed to parse vlan: %v", err))
		return
	}

	if errors := vlan.Validate(); len(errors) > 0 {
		invalidInput(respWriter, strings.Join(errors, ", "))
		return
	}

	// Generate new ID and save
	vlan.ID = uuid.New()
	if err := s.vlanStore.Save(*vlan); err != nil {
		log.Printf("failed to save vlan: %v", err)
		internalError(respWriter, "failed to save vlan")
		return
	}

	respWriter.Header().Set("Location", fmt.Sprintf("%s/%s", req.URL.RequestURI(), vlan.ID.String()))
	respWriter.WriteHeader(http.StatusCreated)
}

func (s *Server) HandleReadVLAN(respWriter http.ResponseWriter, req *http.Request) {
	vlanID, err := uuid.Parse(req.PathValue("id"))
	if err != nil {
		invalidInput(respWriter, "invalid vlan id")
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
		invalidInput(respWriter, "invalid vlan id")
		return
	}

	defer req.Body.Close()
	v := &vlan.VLAN{}
	if err := json.NewDecoder(req.Body).Decode(&v); err != nil {
		invalidInput(respWriter, fmt.Sprintf("failed to parse vlan: %v", err))
		return
	}
	if vlanID != v.ID {
		invalidInput(respWriter, "mismatching vlan id in request body")
		return
	}
	if errors := v.Validate(); len(errors) > 0 {
		invalidInput(respWriter, strings.Join(errors, ", "))
		return
	}

	if err := s.vlanStore.Update(*v); err != nil {
		if errors.Is(err, vlan.ErrNotFound) {
			http.NotFound(respWriter, req)
			return
		}

		log.Printf("failed to update vlan: %v", err)
		internalError(respWriter, "failed to update vlan")
		return
	}
}

func (s *Server) HandleDeleteVLAN(respWriter http.ResponseWriter, req *http.Request) {
	vlanID, err := uuid.Parse(req.PathValue("id"))
	if err != nil {
		invalidInput(respWriter, "invalid vlan id")
		return
	}

	if err := s.vlanStore.Delete(vlanID); err != nil {
		if err == vlan.ErrNotFound {
			http.NotFound(respWriter, req)
			return
		}
		log.Printf("failed to delete vlan: %v", err)
		internalError(respWriter, "failed to delete vlan")
		return
	}
}
