package vlan

import (
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"os"
	"slices"
	"strings"
	"sync"

	"github.com/google/uuid"
)

var ErrNotFound = errors.New("not found")

// Store manages a JSON file to persist VLANs.
type Store struct {
	path      string
	vlansByID map[uuid.UUID]VLAN
	mu        sync.RWMutex
}

func NewStore(path string) (*Store, error) {
	store := &Store{
		path: path,
	}

	// store file does not exist
	if _, err := os.Stat(path); os.IsNotExist(err) {
		store.vlansByID = make(map[uuid.UUID]VLAN, 0)
		store.writeVLANs()
		return store, nil
	}

	// store file exists
	if err := store.readVLANs(); err != nil {
		return nil, err
	}
	return store, nil
}

func (s *Store) List() ([]VLAN, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if len(s.vlansByID) == 0 {
		return []VLAN{}, nil
	}
	return slices.Collect(maps.Values(s.vlansByID)), nil
}

func (s *Store) Get(id uuid.UUID) *VLAN {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if vlan, ok := s.vlansByID[id]; ok {
		return &vlan
	}
	return nil
}

func (s *Store) Save(vlan VLAN) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.vlansByID[vlan.ID] = vlan
	return s.writeVLANs()
}

func (s *Store) Update(vlan VLAN) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.vlansByID[vlan.ID]; !ok {
		return ErrNotFound
	}

	s.vlansByID[vlan.ID] = vlan
	return s.writeVLANs()
}

func (s *Store) Delete(id uuid.UUID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.vlansByID[id]; !ok {
		return ErrNotFound
	}

	delete(s.vlansByID, id)
	return s.writeVLANs()
}

func (s *Store) readVLANs() error {
	vlansFile, err := os.Open(s.path)
	if err != nil {
		return fmt.Errorf("failed to open %v: %w", s.path, err)
	}
	defer vlansFile.Close()

	vlans := []VLAN{}
	if err := json.NewDecoder(vlansFile).Decode(&vlans); err != nil {
		return fmt.Errorf("failed to decode %v: %w", s.path, err)
	}

	vlansByID := make(map[uuid.UUID]VLAN, len(vlans))
	for _, vlan := range vlans {
		if errors := vlan.Validate(); len(errors) > 0 {
			return fmt.Errorf("invalid VLAN in %s: %s", s.path, strings.Join(errors, ", "))
		}
		vlansByID[vlan.ID] = vlan
	}
	s.vlansByID = vlansByID
	return nil
}

func (s *Store) writeVLANs() error {
	vlansFile, err := os.Create(s.path)
	if err != nil {
		return fmt.Errorf("failed to open %v: %w", s.path, err)
	}
	defer vlansFile.Close()

	vlans := slices.Collect(maps.Values(s.vlansByID))
	if err := json.NewEncoder(vlansFile).Encode(vlans); err != nil {
		return fmt.Errorf("failed to encode %v: %w", s.path, err)
	}
	return nil
}
