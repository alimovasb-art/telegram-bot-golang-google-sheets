package fsm

import (
	"sync"

	"github.com/alimovasb-art/telegram-bot-golang-google-sheets/internal/models"
)

type Store struct {
	mu            sync.RWMutex
	states        map[int64]models.State
	registrations map[int64]*models.RegistrationData
}

func NewStore() *Store {
	return &Store{
		states:        make(map[int64]models.State),
		registrations: make(map[int64]*models.RegistrationData),
	}
}

func (s *Store) SetState(userID int64, state models.State) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.states[userID] = state
}

func (s *Store) GetState(userID int64) models.State {
	s.mu.RLock()
	defer s.mu.RUnlock()
	state, ok := s.states[userID]
	if !ok {
		return models.StateIdle
	}
	return state
}

func (s *Store) SetRegistrationData(userID int64, data *models.RegistrationData) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.registrations[userID] = data
}

func (s *Store) GetRegistrationData(userID int64) *models.RegistrationData {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.registrations[userID]
}

func (s *Store) Clear(userID int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.states, userID)
	delete(s.registrations, userID)
}
