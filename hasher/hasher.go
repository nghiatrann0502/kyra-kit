package hasher

import (
	"errors"
	"fmt"
	"strings"
)

// ===============================
// Interfaces & Manager
// ===============================

type Provider interface {
	ID() string
	Hash(password []byte) (encoded string, err error)
	Verify(password []byte, encoded string) (ok bool, err error)
	// NeedsRehash(encoded string) bool
}

type Manager struct {
	providers map[string]Provider
	defaultID string
}

type WeightedID struct {
	ID     string
	Weight uint64 // Weight > 0
}

func NewManager(defaultID string, providers ...Provider) (*Manager, error) {
	if len(providers) == 0 {
		return nil, errors.New("hasher: at least one provider is required")
	}

	m := &Manager{providers: make(map[string]Provider), defaultID: strings.ToLower(defaultID)}
	for _, p := range providers {
		if p == nil || p.ID() == "" {
			return nil, errors.New("hasher: invalid provider")
		}
		id := strings.ToLower(p.ID())
		if _, exists := m.providers[id]; exists {
			return nil, fmt.Errorf("hasher: duplicate provider id %q", id)
		}
		m.providers[id] = p
	}
	if _, ok := m.providers[m.defaultID]; !ok {
		return nil, fmt.Errorf("hasher: default provider %q not registered", m.defaultID)
	}
	return m, nil
}

func MustNewManager(defaultID string, providers ...Provider) *Manager {
	m, err := NewManager(defaultID, providers...)
	if err != nil {
		panic(err)
	}
	return m
}

func (m *Manager) Hash(password []byte) (string, error) {
	return m.providers[m.defaultID].Hash(password)
}

func (m *Manager) HashWith(id string, password []byte) (string, error) {
	p := m.providers[strings.ToLower(id)]
	if p == nil {
		return "", fmt.Errorf("hasher: provider %q not registered", id)
	}
	return p.Hash(password)
}

func (m *Manager) HashRandom(password []byte, ids ...string) (string, string, error) {
	pool := ids
	if len(pool) == 0 {
		pool = make([]string, 0, len(m.providers))
		for id := range m.providers {
			pool = append(pool, id)
		}
	}

	filtered := make([]string, 0, len(pool))
	for _, id := range pool {
		id = strings.ToLower(id)
		if _, ok := m.providers[id]; ok {
			filtered = append(filtered, id)
		}
	}
	if len(filtered) == 0 {
		return "", "", errors.New("hasher: no valid providers to choose from")
	}
	idx, err := cryptoRandIndex(len(filtered))
	if err != nil {
		return "", "", err
	}
	picked := filtered[idx]
	enc, err := m.providers[picked].Hash(password)
	return picked, enc, err
}

func (m *Manager) HashWeighted(password []byte, items ...WeightedID) (string, string, error) {
	type entry struct {
		id string
		w  uint64
	}
	var list []entry
	var total uint64
	for _, it := range items {
		id := strings.ToLower(it.ID)
		if it.Weight == 0 {
			continue
		}
		if _, ok := m.providers[id]; !ok {
			continue
		}
		list = append(list, entry{id: id, w: it.Weight})
		total += it.Weight
	}
	if len(list) == 0 {
		return "", "", errors.New("hasher: no weighted providers")
	}
	r, err := cryptoRandUint64N(total)
	if err != nil {
		return "", "", err
	}
	var acc uint64
	for _, e := range list {
		acc += e.w
		if r < acc {
			enc, err := m.providers[e.id].Hash(password)
			return e.id, enc, err
		}
	}
	return "", "", errors.New("hasher: selection failed")
}

func (m *Manager) Verify(password []byte, encoded string) (bool, error) {
	p, id := m.detect(encoded)
	if p == nil {
		return false, fmt.Errorf("hasher: unknown hash format (id=%q)", id)
	}
	return p.Verify(password, encoded)
}

//
// func (m *Manager) NeedsRehash(encoded string) bool {
// 	p, _ := m.detect(encoded)
// 	if p == nil {
// 		return true
// 	}
// 	if strings.ToLower(p.ID()) != m.defaultID {
// 		return true
// 	}
// 	return p.NeedsRehash(encoded)
// }

func (m *Manager) VerifyAndUpgrade(password []byte, encoded string) (upgraded bool, newEncoded string, err error) {
	ok, err := m.Verify(password, encoded)
	if err != nil || !ok {
		return false, "", err
	}
	// if m.NeedsRehash(encoded) {
	// 	ne, err := m.Hash(password)
	// 	return err == nil, ne, err
	// }
	return false, "", nil
}

func (m *Manager) detect(encoded string) (Provider, string) {
	id := strings.ToLower(extractPHCID(encoded))
	if id == "" {
		if isBcryptNative(encoded) {
			id = "bcrypt"
		}
	}
	if id == "" {
		return nil, ""
	}
	p := m.providers[id]
	return p, id
}
