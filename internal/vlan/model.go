package vlan

import (
	"fmt"
	"net/netip"

	"github.com/google/uuid"
)

type VLAN struct {
	ID      uuid.UUID    `json:"id"`
	VID     uint16       `json:"vid"`
	Name    string       `json:"name"`
	Subnet  netip.Prefix `json:"subnet"`
	Gateway netip.Addr   `json:"gateway"`
	Status  string       `json:"status"`
}

func (v *VLAN) Validate() []string {
	errors := make([]string, 0)
	if v.VID < 1 || v.VID > 4094 {
		errors = append(errors, fmt.Sprintf("invalid VLAN ID %v (expected range 1..4094)", v.VID))
	}
	if v.Name == "" {
		errors = append(errors, "name must not be empty")
	}
	if !v.Subnet.Contains(v.Gateway) {
		errors = append(errors, fmt.Sprintf("gateway %s must belong to subnet %s", v.Gateway, v.Subnet))
	}
	return errors
}
