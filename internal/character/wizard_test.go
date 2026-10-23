package character

import (
	"path/filepath"
	"testing"
)

func TestWizardCreationFlow(t *testing.T) {
	tempDir := t.TempDir()
	svc := NewPersistentService(filepath.Join(tempDir, "wizard_fichas.json"))
	wizard := NewWizardManager(svc)

	userID := "user_wizard_1"
	channelID := "chan_1"

	// 1. Inicia sessão
	prompt, err := wizard.StartSession(userID, channelID)
	if err != nil || prompt == "" {
		t.Fatalf("Erro ao iniciar sessão: %v", err)
	}
	if !wizard.HasActiveSession(userID, channelID) {
		t.Errorf("Deveria haver sessão ativa")
	}

	// Step 0: Nome
	resp, sheet, _ := wizard.ProcessInput(userID, "Eldrin")
	if sheet != nil || resp == "" {
		t.Errorf("Não deveria terminar no passo 0")
	}

	// Step 1: Raça
	wizard.ProcessInput(userID, "Elfo da Floresta")

	// Step 2: Classe
	wizard.ProcessInput(userID, "Ladino Assassino")

	// Step 3: Nível
	wizard.ProcessInput(userID, "3")

	// Step 4: Atributos com rolagem automática "rolar"
	respAttr, _, _ := wizard.ProcessInput(userID, "rolar")
	if respAttr == "" {
		t.Errorf("Resposta de rolagem não deve ser vazia")
	}

	// Step 5: HP e CA
	wizard.ProcessInput(userID, "24 HP, 15 CA")

	// Step 6: Salvaguardas
	wizard.ProcessInput(userID, "Destreza e Inteligência")

	// Step 7: Perícias
	wizard.ProcessInput(userID, "Furtividade (Expertise), Acrobacia, Percepção, Intuição")

	// Step 8: Habilidades
	wizard.ProcessInput(userID, "Visão no Escuro, Ataque Furtivo, Sortuda (2 usos)")

	// Step 9: Magias
	wizard.ProcessInput(userID, "nenhuma")

	// Step 10: Equipamento
	finalResp, finishedSheet, err := wizard.ProcessInput(userID, "Espada Curta (+5, 1d6+3 Piercing), Adaga (+5, 1d4+3 Piercing), Armadura de Couro, 50 PO")

	if err != nil {
		t.Fatalf("Erro no passo final: %v", err)
	}
	if finishedSheet == nil {
		t.Fatalf("finishedSheet não deveria ser nil ao concluir")
	}
	if finishedSheet.Name != "Eldrin" {
		t.Errorf("Esperado nome Eldrin, obtido %s", finishedSheet.Name)
	}
	if finishedSheet.Level != 3 {
		t.Errorf("Esperado nível 3, obtido %d", finishedSheet.Level)
	}
	if finishedSheet.MaxHP != 24 {
		t.Errorf("Esperado MaxHP 24, obtido %d", finishedSheet.MaxHP)
	}
	if finishedSheet.ArmorClass != 15 {
		t.Errorf("Esperado CA 15, obtido %d", finishedSheet.ArmorClass)
	}
	if finishedSheet.Coins.GP != 50 {
		t.Errorf("Esperado 50 GP, obtido %d", finishedSheet.Coins.GP)
	}
	if len(finishedSheet.LimitedUses) == 0 {
		t.Errorf("Esperado recurso limitado Sortuda com 2 usos")
	}

	// Verifica se foi salvo no serviço e ativado
	active, ok := svc.GetActive(userID)
	if !ok || active.Name != "Eldrin" {
		t.Errorf("A nova ficha deveria estar salva e ativa no serviço")
	}
	if finalResp == "" {
		t.Errorf("finalResp não deveria ser vazio")
	}
}

func TestWizardCancel(t *testing.T) {
	tempDir := t.TempDir()
	svc := NewPersistentService(filepath.Join(tempDir, "wizard_cancel_fichas.json"))
	wizard := NewWizardManager(svc)

	userID := "user_cancel"
	channelID := "chan_1"

	wizard.StartSession(userID, channelID)
	resp, sheet, err := wizard.ProcessInput(userID, "cancelar")
	if err != nil || sheet != nil {
		t.Errorf("Erro ou sheet não-nil ao cancelar: %v", err)
	}
	if wizard.HasActiveSession(userID, channelID) {
		t.Errorf("Sessão não deveria estar mais ativa após cancelar")
	}
	if resp == "" {
		t.Errorf("Deveria haver mensagem de cancelamento")
	}
}
