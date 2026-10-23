package userpref

import (
	"encoding/json"
	"os"
	"sync"
)

// DefaultSuccessThreshold é o valor padrão a ser batido no dado para ser considerado sucesso
const DefaultSuccessThreshold = 5

// UserPreferences armazena configurações individuais de cada jogador
type UserPreferences struct {
	UserID           string `json:"user_id"`
	HalflingLuck     bool   `json:"halfling_luck"`      // Sorte Halfling: Rerrola automaticamente 1s em d20s
	RerollOn1All     bool   `json:"reroll_on_1_all"`    // Rerrola automaticamente 1s em qualquer dado
	SuccessThreshold int    `json:"success_threshold"`  // Valor mínimo para contar sucesso (ex: >= 5)
}

// Service define o contrato de gerenciamento de preferências de jogadores
type Service interface {
	GetPreferences(userID string) UserPreferences
	ToggleHalflingLuck(userID string) bool
	ToggleRerollOn1All(userID string) bool
	SetSuccessThreshold(userID string, threshold int) int
}

// MemoryService implementa Service em memória thread-safe com persistência em disco opcional
type MemoryService struct {
	mu       sync.RWMutex
	prefs    map[string]UserPreferences
	filePath string
}

// NewMemoryService instancia o serviço de preferências exclusivamente em memória
func NewMemoryService() *MemoryService {
	return &MemoryService{
		prefs: make(map[string]UserPreferences),
	}
}

// NewPersistentService instancia o serviço de preferências com arquivo JSON para persistência
func NewPersistentService(filePath string) *MemoryService {
	s := &MemoryService{
		prefs:    make(map[string]UserPreferences),
		filePath: filePath,
	}
	s.loadFromFile()
	return s
}

func (s *MemoryService) loadFromFile() {
	if s.filePath == "" {
		return
	}
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		return
	}
	var loaded map[string]UserPreferences
	if err := json.Unmarshal(data, &loaded); err == nil && loaded != nil {
		s.prefs = loaded
	}
}

func (s *MemoryService) saveToFileLocked() {
	if s.filePath == "" {
		return
	}
	data, err := json.MarshalIndent(s.prefs, "", "  ")
	if err == nil {
		_ = os.WriteFile(s.filePath, data, 0644)
	}
}

// GetPreferences retorna as preferências do usuário (ou valores padrão se não existir)
func (s *MemoryService) GetPreferences(userID string) UserPreferences {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if p, ok := s.prefs[userID]; ok {
		if p.SuccessThreshold <= 0 {
			p.SuccessThreshold = DefaultSuccessThreshold
		}
		return p
	}
	return UserPreferences{
		UserID:           userID,
		HalflingLuck:     false,
		RerollOn1All:     false,
		SuccessThreshold: DefaultSuccessThreshold,
	}
}

// ToggleHalflingLuck alterna a configuração de Sorte Halfling (d20)
func (s *MemoryService) ToggleHalflingLuck(userID string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	p, ok := s.prefs[userID]
	if !ok {
		p = UserPreferences{
			UserID:           userID,
			SuccessThreshold: DefaultSuccessThreshold,
		}
	}
	p.HalflingLuck = !p.HalflingLuck
	s.prefs[userID] = p
	s.saveToFileLocked()
	return p.HalflingLuck
}

// ToggleRerollOn1All alterna a configuração de rerrolar 1 em todos os dados
func (s *MemoryService) ToggleRerollOn1All(userID string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	p, ok := s.prefs[userID]
	if !ok {
		p = UserPreferences{
			UserID:           userID,
			SuccessThreshold: DefaultSuccessThreshold,
		}
	}
	p.RerollOn1All = !p.RerollOn1All
	s.prefs[userID] = p
	s.saveToFileLocked()
	return p.RerollOn1All
}

// SetSuccessThreshold define o limite de sucesso configurado para o jogador
func (s *MemoryService) SetSuccessThreshold(userID string, threshold int) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	p, ok := s.prefs[userID]
	if !ok {
		p = UserPreferences{
			UserID:           userID,
			SuccessThreshold: DefaultSuccessThreshold,
		}
	}
	if threshold < 1 {
		threshold = 1
	}
	if threshold > 100 {
		threshold = 100
	}
	p.SuccessThreshold = threshold
	s.prefs[userID] = p
	s.saveToFileLocked()
	return p.SuccessThreshold
}
