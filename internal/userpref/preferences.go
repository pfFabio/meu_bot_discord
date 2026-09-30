package userpref

import "sync"

// UserPreferences armazena configurações individuais de cada jogador
type UserPreferences struct {
	UserID       string
	HalflingLuck bool // Sorte Halfling: Rerrola automaticamente 1s em d20s
	RerollOn1All bool // Rerrola automaticamente 1s em qualquer dado (ex: dano / Great Weapon Fighting)
}

// Service define o contrato de gerenciamento de preferências de jogadores
type Service interface {
	GetPreferences(userID string) UserPreferences
	ToggleHalflingLuck(userID string) bool
	ToggleRerollOn1All(userID string) bool
}

// MemoryService implementa Service em memória thread-safe
type MemoryService struct {
	mu    sync.RWMutex
	prefs map[string]UserPreferences
}

// NewMemoryService instancia o serviço de preferências
func NewMemoryService() *MemoryService {
	return &MemoryService{
		prefs: make(map[string]UserPreferences),
	}
}

// GetPreferences retorna as preferências do usuário (ou valores padrão se não existir)
func (s *MemoryService) GetPreferences(userID string) UserPreferences {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if p, ok := s.prefs[userID]; ok {
		return p
	}
	return UserPreferences{
		UserID:       userID,
		HalflingLuck: false,
		RerollOn1All: false,
	}
}

// ToggleHalflingLuck alterna a configuração de Sorte Halfling (d20)
func (s *MemoryService) ToggleHalflingLuck(userID string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	p, ok := s.prefs[userID]
	if !ok {
		p = UserPreferences{UserID: userID}
	}
	p.HalflingLuck = !p.HalflingLuck
	s.prefs[userID] = p
	return p.HalflingLuck
}

// ToggleRerollOn1All alterna a configuração de rerrolar 1 em todos os dados
func (s *MemoryService) ToggleRerollOn1All(userID string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	p, ok := s.prefs[userID]
	if !ok {
		p = UserPreferences{UserID: userID}
	}
	p.RerollOn1All = !p.RerollOn1All
	s.prefs[userID] = p
	return p.RerollOn1All
}
