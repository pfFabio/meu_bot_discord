package character

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Service define os contratos para gerenciamento e persistência de fichas de personagens
type Service interface {
	Save(sheet *CharacterSheet) error
	GetActive(userID string) (*CharacterSheet, bool)
	SetActive(userID, sheetID string) error
	GetByID(sheetID string) (*CharacterSheet, bool)
	ListByUser(userID string) []*CharacterSheet
	GetAll() []*CharacterSheet
	Delete(sheetID string) error
	ApplyDamage(userID string, amount int) (*CharacterSheet, int, bool, error)
	ApplyHealing(userID string, amount int) (*CharacterSheet, int, error)
	SetTempHP(userID string, amount int) (*CharacterSheet, error)
	UseSpellSlot(userID string, level int) (*CharacterSheet, error)
	RestoreSpellSlot(userID string, level int) (*CharacterSheet, error)
	PerformLongRest(userID string) (*CharacterSheet, error)
	PerformShortRest(userID string) (*CharacterSheet, error)
	UseResource(userID string, resourceName string) (*CharacterSheet, error)
	RestoreResource(userID string, resourceName string) (*CharacterSheet, error)
	AssignSheet(sheetID string, targetUserID string) (*CharacterSheet, error)
	ImportFromPDF(userID, pdfPath string) (*CharacterSheet, error)
}

type storageData struct {
	Sheets       map[string]*CharacterSheet `json:"sheets"`        // sheetID -> sheet
	ActiveSheets map[string]string          `json:"active_sheets"` // userID -> sheetID
}

// MemoryService implementa o serviço de fichas com persistência thread-safe em JSON
type MemoryService struct {
	mu           sync.RWMutex
	sheets       map[string]*CharacterSheet
	activeSheets map[string]string // userID -> sheetID
	filePath     string
	pythonPath   string
}

// NewPersistentService cria o serviço de fichas com persistência no arquivo especificado
func NewPersistentService(filePath string) *MemoryService {
	s := &MemoryService{
		sheets:       make(map[string]*CharacterSheet),
		activeSheets: make(map[string]string),
		filePath:     filePath,
		pythonPath:   "python",
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

	var stored storageData
	if err := json.Unmarshal(data, &stored); err == nil {
		if stored.Sheets != nil {
			s.sheets = stored.Sheets
		}
		if stored.ActiveSheets != nil {
			s.activeSheets = stored.ActiveSheets
		}
		slog.Info("Fichas de personagens carregadas com sucesso", "total", len(s.sheets))
	}
}

func (s *MemoryService) saveToFileLocked() error {
	if s.filePath == "" {
		return nil
	}
	stored := storageData{
		Sheets:       s.sheets,
		ActiveSheets: s.activeSheets,
	}

	data, err := json.MarshalIndent(stored, "", "  ")
	if err != nil {
		return err
	}

	dir := filepath.Dir(s.filePath)
	if dir != "" && dir != "." {
		_ = os.MkdirAll(dir, 0755)
	}

	return os.WriteFile(s.filePath, data, 0644)
}

// Save salva ou atualiza uma ficha no repositório
func (s *MemoryService) Save(sheet *CharacterSheet) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if sheet.ID == "" {
		sheet.ID = fmt.Sprintf("sheet_%d", time.Now().UnixNano())
	}
	if sheet.CreatedAt.IsZero() {
		sheet.CreatedAt = time.Now()
	}
	sheet.CalculateDerived()

	s.sheets[sheet.ID] = sheet

	// Se o usuário não tiver uma ficha ativa, define esta como ativa
	if sheet.UserID != "" {
		if _, ok := s.activeSheets[sheet.UserID]; !ok {
			s.activeSheets[sheet.UserID] = sheet.ID
		}
	}

	return s.saveToFileLocked()
}

// GetActive retorna a ficha atualmente ativa de um usuário
func (s *MemoryService) GetActive(userID string) (*CharacterSheet, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	activeID, ok := s.activeSheets[userID]
	if !ok {
		// Se não tiver ativa explícita, procura a primeira ficha do usuário
		for _, sheet := range s.sheets {
			if sheet.UserID == userID {
				return sheet, true
			}
		}
		return nil, false
	}

	sheet, found := s.sheets[activeID]
	return sheet, found
}

// SetActive define a ficha ativa para o usuário
func (s *MemoryService) SetActive(userID, sheetID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	sheet, ok := s.sheets[sheetID]
	if !ok {
		return fmt.Errorf("ficha '%s' não encontrada", sheetID)
	}
	if sheet.UserID != userID {
		return fmt.Errorf("você não tem permissão para ativar uma ficha que pertence a outro usuário")
	}

	s.activeSheets[userID] = sheetID
	return s.saveToFileLocked()
}

// AssignSheet vincula uma ficha existente a um novo usuário e a define como ativa para ele
func (s *MemoryService) AssignSheet(sheetID string, targetUserID string) (*CharacterSheet, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	sheet, ok := s.sheets[sheetID]
	if !ok {
		// Procura por nome case-insensitive
		sLow := strings.ToLower(sheetID)
		for _, sh := range s.sheets {
			if strings.EqualFold(sh.Name, sheetID) || strings.Contains(strings.ToLower(sh.Name), sLow) {
				sheet = sh
				ok = true
				break
			}
		}
		if !ok {
			return nil, fmt.Errorf("ficha '%s' não encontrada", sheetID)
		}
	}

	sheet.UserID = targetUserID
	s.activeSheets[targetUserID] = sheet.ID
	sheet.UpdatedAt = time.Now()
	_ = s.saveToFileLocked()
	return sheet, nil
}

// GetByID busca uma ficha pelo ID
func (s *MemoryService) GetByID(sheetID string) (*CharacterSheet, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sheet, ok := s.sheets[sheetID]
	return sheet, ok
}

// ListByUser lista todas as fichas pertencentes a um usuário. Se userID for vazio, retorna todas as fichas cadastradas.
func (s *MemoryService) ListByUser(userID string) []*CharacterSheet {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*CharacterSheet
	for _, sheet := range s.sheets {
		if userID == "" || sheet.UserID == userID {
			result = append(result, sheet)
		}
	}
	return result
}

// GetAll retorna todas as fichas cadastradas no repositório
func (s *MemoryService) GetAll() []*CharacterSheet {
	return s.ListByUser("")
}

// Delete exclui uma ficha pelo ID
func (s *MemoryService) Delete(sheetID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	sheet, ok := s.sheets[sheetID]
	if !ok {
		return fmt.Errorf("ficha não encontrada")
	}

	delete(s.sheets, sheetID)
	if s.activeSheets[sheet.UserID] == sheetID {
		delete(s.activeSheets, sheet.UserID)
		// Se houver outra ficha do mesmo usuário, promove a ativa
		for _, other := range s.sheets {
			if other.UserID == sheet.UserID {
				s.activeSheets[sheet.UserID] = other.ID
				break
			}
		}
	}

	return s.saveToFileLocked()
}

// ApplyDamage aplica dano à ficha ativa do usuário
func (s *MemoryService) ApplyDamage(userID string, amount int) (*CharacterSheet, int, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	sheet, ok := s.getActiveLocked(userID)
	if !ok {
		return nil, 0, false, fmt.Errorf("nenhuma ficha ativa encontrada para este usuário. Use `/ficha criar` ou `/ficha importar`")
	}

	lost, isDown := sheet.TakeDamage(amount)
	_ = s.saveToFileLocked()
	return sheet, lost, isDown, nil
}

// ApplyHealing cura a ficha ativa do usuário
func (s *MemoryService) ApplyHealing(userID string, amount int) (*CharacterSheet, int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	sheet, ok := s.getActiveLocked(userID)
	if !ok {
		return nil, 0, fmt.Errorf("nenhuma ficha ativa encontrada")
	}

	healed := sheet.Heal(amount)
	_ = s.saveToFileLocked()
	return sheet, healed, nil
}

// SetTempHP define pontos de vida temporários na ficha ativa do usuário
func (s *MemoryService) SetTempHP(userID string, amount int) (*CharacterSheet, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	sheet, ok := s.getActiveLocked(userID)
	if !ok {
		return nil, fmt.Errorf("nenhuma ficha ativa encontrada")
	}

	sheet.SetTemporaryHP(amount)
	_ = s.saveToFileLocked()
	return sheet, nil
}

// UseSpellSlot gasta um espaço de magia na ficha ativa
func (s *MemoryService) UseSpellSlot(userID string, level int) (*CharacterSheet, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	sheet, ok := s.getActiveLocked(userID)
	if !ok {
		return nil, fmt.Errorf("nenhuma ficha ativa encontrada")
	}

	if err := sheet.UseSpellSlot(level); err != nil {
		return nil, err
	}

	_ = s.saveToFileLocked()
	return sheet, nil
}

// RestoreSpellSlot recupera um espaço de magia na ficha ativa
func (s *MemoryService) RestoreSpellSlot(userID string, level int) (*CharacterSheet, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	sheet, ok := s.getActiveLocked(userID)
	if !ok {
		return nil, fmt.Errorf("nenhuma ficha ativa encontrada")
	}

	if err := sheet.RestoreSpellSlot(level); err != nil {
		return nil, err
	}

	_ = s.saveToFileLocked()
	return sheet, nil
}

// PerformLongRest realiza descanso longo na ficha ativa
func (s *MemoryService) PerformLongRest(userID string) (*CharacterSheet, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	sheet, ok := s.getActiveLocked(userID)
	if !ok {
		return nil, fmt.Errorf("nenhuma ficha ativa encontrada")
	}

	sheet.LongRest()
	_ = s.saveToFileLocked()
	return sheet, nil
}

// PerformShortRest realiza descanso curto na ficha ativa
func (s *MemoryService) PerformShortRest(userID string) (*CharacterSheet, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	sheet, ok := s.getActiveLocked(userID)
	if !ok {
		return nil, fmt.Errorf("nenhuma ficha ativa encontrada")
	}

	sheet.ShortRest()
	_ = s.saveToFileLocked()
	return sheet, nil
}

// UseResource gasta um recurso limitado na ficha ativa
func (s *MemoryService) UseResource(userID string, resourceName string) (*CharacterSheet, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	sheet, ok := s.getActiveLocked(userID)
	if !ok {
		return nil, fmt.Errorf("nenhuma ficha ativa encontrada")
	}

	if err := sheet.UseResource(resourceName); err != nil {
		return nil, err
	}

	_ = s.saveToFileLocked()
	return sheet, nil
}

// RestoreResource recupera um recurso limitado na ficha ativa
func (s *MemoryService) RestoreResource(userID string, resourceName string) (*CharacterSheet, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	sheet, ok := s.getActiveLocked(userID)
	if !ok {
		return nil, fmt.Errorf("nenhuma ficha ativa encontrada")
	}

	if err := sheet.RestoreResource(resourceName); err != nil {
		return nil, err
	}

	_ = s.saveToFileLocked()
	return sheet, nil
}

// findScriptPath localiza o parse_character_pdf.py no diretório atual ou ancestrais
func findScriptPath() string {
	candidatePaths := []string{
		"parse_character_pdf.py",
		filepath.Join("..", "parse_character_pdf.py"),
		filepath.Join("..", "..", "parse_character_pdf.py"),
	}
	for _, p := range candidatePaths {
		if abs, err := filepath.Abs(p); err == nil {
			if _, err := os.Stat(abs); err == nil {
				return abs
			}
		}
	}
	return "parse_character_pdf.py"
}

// ImportFromPDF executa o script parse_character_pdf.py e importa a ficha para o usuário
func (s *MemoryService) ImportFromPDF(userID, pdfPath string) (*CharacterSheet, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	scriptPath := findScriptPath()
	cmd := exec.CommandContext(ctx, s.pythonPath, scriptPath, pdfPath)
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("erro ao analisar PDF: %s (detalhes: %s)", err, string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("falha ao invocar extrator de PDF: %w", err)
	}

	var sheet CharacterSheet
	if err := json.Unmarshal(output, &sheet); err != nil {
		return nil, fmt.Errorf("falha ao interpretar dados extraídos do PDF: %w", err)
	}

	sheet.ID = fmt.Sprintf("sheet_%d", time.Now().UnixNano())
	sheet.UserID = userID
	sheet.CreatedAt = time.Now()
	sheet.CalculateDerived()

	if err := s.Save(&sheet); err != nil {
		return nil, fmt.Errorf("falha ao salvar ficha importada: %w", err)
	}

	// Ativa a nova ficha automaticamente
	_ = s.SetActive(userID, sheet.ID)

	slog.Info("Ficha importada com sucesso via PDF",
		"user_id", userID,
		"name", sheet.Name,
		"class", sheet.Class,
		"level", sheet.Level,
	)

	return &sheet, nil
}

func (s *MemoryService) getActiveLocked(userID string) (*CharacterSheet, bool) {
	activeID, ok := s.activeSheets[userID]
	if !ok {
		for _, sheet := range s.sheets {
			if sheet.UserID == userID {
				return sheet, true
			}
		}
		return nil, false
	}
	sheet, found := s.sheets[activeID]
	return sheet, found
}
