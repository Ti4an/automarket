package storage

import (
	"encoding/json"
	"errors"
	"os"
	"sync"

	"automarket/models"
)

var ErrNotFound = errors.New("recurso no encontrado")

type data struct {
	Users    []models.User    `json:"users"`
	Vehicles []models.Vehicle `json:"vehicles"`
}

type Store struct {
	mu   sync.Mutex
	path string
	data data
}

func New(path string) (*Store, error) {
	s := &Store{path: path}
	content, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return s, nil
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(content, &s.data); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Store) save() error {
	content, err := json.MarshalIndent(s.data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, content, 0600)
}

func (s *Store) AddUser(u models.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data.Users = append(s.data.Users, u)
	return s.save()
}

func (s *Store) FindUserByEmail(email string) (models.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, u := range s.data.Users {
		if u.Email == email {
			return u, nil
		}
	}
	return models.User{}, ErrNotFound
}

func (s *Store) FindUserByID(id string) (models.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, u := range s.data.Users {
		if u.ID == id {
			return u, nil
		}
	}
	return models.User{}, ErrNotFound
}

func (s *Store) DeleteUser(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, u := range s.data.Users {
		if u.ID == id {
			s.data.Users = append(s.data.Users[:i], s.data.Users[i+1:]...)
			return s.save()
		}
	}
	return ErrNotFound
}

func (s *Store) AddVehicle(v models.Vehicle) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data.Vehicles = append(s.data.Vehicles, v)
	return s.save()
}

func (s *Store) FindVehicleByID(id string) (models.Vehicle, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, v := range s.data.Vehicles {
		if v.ID == id {
			return v, nil
		}
	}
	return models.Vehicle{}, ErrNotFound
}

func (s *Store) ListVehiclesByStatus(status models.VehicleStatus) []models.Vehicle {
	s.mu.Lock()
	defer s.mu.Unlock()
	result := []models.Vehicle{}
	for _, v := range s.data.Vehicles {
		if v.Status == status {
			result = append(result, v)
		}
	}
	return result
}

// ListVehicles devuelve los vehículos que cumplen el filtro dado.
func (s *Store) ListVehicles(keep func(models.Vehicle) bool) []models.Vehicle {
	s.mu.Lock()
	defer s.mu.Unlock()
	result := []models.Vehicle{}
	for _, v := range s.data.Vehicles {
		if keep(v) {
			result = append(result, v)
		}
	}
	return result
}

// ListUsersByRole devuelve los usuarios con el rol indicado.
func (s *Store) ListUsersByRole(role models.Role) []models.User {
	s.mu.Lock()
	defer s.mu.Unlock()
	result := []models.User{}
	for _, u := range s.data.Users {
		if u.Role == role {
			result = append(result, u)
		}
	}
	return result
}

func (s *Store) UpdateVehicle(updated models.Vehicle) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, v := range s.data.Vehicles {
		if v.ID == updated.ID {
			s.data.Vehicles[i] = updated
			return s.save()
		}
	}
	return ErrNotFound
}

func (s *Store) DeleteVehicle(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, v := range s.data.Vehicles {
		if v.ID == id {
			s.data.Vehicles = append(s.data.Vehicles[:i], s.data.Vehicles[i+1:]...)
			return s.save()
		}
	}
	return ErrNotFound
}

func (s *Store) DeleteVehiclesBySeller(sellerID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	kept := []models.Vehicle{}
	for _, v := range s.data.Vehicles {
		if v.SellerID != sellerID {
			kept = append(kept, v)
		}
	}
	s.data.Vehicles = kept
	return s.save()
}