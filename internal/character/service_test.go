package character

import (
	"os"
	"path/filepath"
	"testing"
)

func TestServiceCRUDAndActiveSheet(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "test_fichas.json")

	svc := NewPersistentService(filePath)

	sheet1 := &CharacterSheet{
		Name:      "Valerian",
		Class:     "Bardo",
		Level:     3,
		UserID:    "user_123",
		MaxHP:     24,
		CurrentHP: 24,
	}

	if err := svc.Save(sheet1); err != nil {
		t.Fatalf("Erro ao salvar sheet1: %v", err)
	}

	// Como é a primeira ficha do usuário, ela deve ser a ativa
	active, ok := svc.GetActive("user_123")
	if !ok || active.Name != "Valerian" {
		t.Fatalf("Esperado Valerian como ativo, obtido %v (ok=%v)", active, ok)
	}

	// Adiciona segunda ficha
	sheet2 := &CharacterSheet{
		Name:      "Lyana",
		Class:     "Ladino",
		Level:     3,
		UserID:    "user_123",
		MaxHP:     24,
		CurrentHP: 24,
	}

	if err := svc.Save(sheet2); err != nil {
		t.Fatalf("Erro ao salvar sheet2: %v", err)
	}

	// Valerian ainda deve ser o ativo
	active, _ = svc.GetActive("user_123")
	if active.Name != "Valerian" {
		t.Errorf("Valerian ainda deveria ser a ficha ativa")
	}

	// Troca a ativa para Lyana
	if err := svc.SetActive("user_123", sheet2.ID); err != nil {
		t.Fatalf("Erro ao ativar Lyana: %v", err)
	}

	active, _ = svc.GetActive("user_123")
	if active.Name != "Lyana" {
		t.Errorf("Lyana deveria ser a nova ficha ativa")
	}

	// Lista todas do usuário
	list := svc.ListByUser("user_123")
	if len(list) != 2 {
		t.Errorf("Esperado 2 fichas para user_123, obtido %d", len(list))
	}

	// Testa aplicação de dano na ativa (Lyana)
	sheet, lost, isDown, err := svc.ApplyDamage("user_123", 10)
	if err != nil {
		t.Fatalf("Erro ao aplicar dano: %v", err)
	}
	if lost != 10 || isDown {
		t.Errorf("Esperado 10 perdido, isDown false. Obtido: lost=%d, isDown=%v", lost, isDown)
	}
	if sheet.CurrentHP != 14 {
		t.Errorf("Esperado 14 HP na Lyana, obtido %d", sheet.CurrentHP)
	}

	// Testa persistência recarregando de arquivo
	svcReloaded := NewPersistentService(filePath)
	activeReloaded, ok := svcReloaded.GetActive("user_123")
	if !ok || activeReloaded.Name != "Lyana" || activeReloaded.CurrentHP != 14 {
		t.Errorf("Falha na persistência ao recarregar dados do disco: %v", activeReloaded)
	}
}

func TestServicePDFImport(t *testing.T) {
	pdfPath := filepath.Join("..", "..", "jogadores", "lyana hidefeather.pdf")
	if _, err := os.Stat(pdfPath); os.IsNotExist(err) {
		t.Skip("PDF de teste não encontrado, pulando teste de integração de PDF")
	}

	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "pdf_test_fichas.json")
	svc := NewPersistentService(filePath)

	sheet, err := svc.ImportFromPDF("user_test", pdfPath)
	if err != nil {
		t.Fatalf("Erro ao importar PDF: %v", err)
	}

	if sheet.Name == "" {
		t.Errorf("Nome não deveria ser vazio na ficha importada")
	}
	if sheet.Level != 3 {
		t.Errorf("Esperado nível 3 para Lyana, obtido %d", sheet.Level)
	}
	if sheet.MaxHP != 24 {
		t.Errorf("Esperado MaxHP 24 para Lyana, obtido %d", sheet.MaxHP)
	}

	// Verifica se a ficha ativa foi definida
	active, ok := svc.GetActive("user_test")
	if !ok || active.ID != sheet.ID {
		t.Errorf("Ficha importada deveria ser definida como ativa automaticamente")
	}
}
