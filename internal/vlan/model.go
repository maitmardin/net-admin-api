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
	return errors
}
