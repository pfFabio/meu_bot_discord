package handler

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/bwmarrin/discordgo"
	"kenobi-bot/internal/character"
	"kenobi-bot/internal/dice"
)

func TestCharacterHandlerSlashCommands(t *testing.T) {
	tempDir := t.TempDir()
	svc := character.NewPersistentService(filepath.Join(tempDir, "fichas.json"))
	h := NewCharacterHandler(svc)

	cmds := h.GetSlashCommands()
	if len(cmds) < 4 {
		t.Fatalf("Esperado no mínimo 4 comandos slash (/ficha, /hp, /magia, /descanso), obtido %d", len(cmds))
	}

	cmdMap := make(map[string]*discordgo.ApplicationCommand)
	for _, c := range cmds {
		cmdMap[c.Name] = c
	}

	if _, ok := cmdMap["ficha"]; !ok {
		t.Errorf("Comando /ficha não encontrado")
	}
	if _, ok := cmdMap["hp"]; !ok {
		t.Errorf("Comando /hp não encontrado")
	}
	if _, ok := cmdMap["magia"]; !ok {
		t.Errorf("Comando /magia não encontrado")
	}
	if _, ok := cmdMap["descanso"]; !ok {
		t.Errorf("Comando /descanso não encontrado")
	}
}

func TestCharacterHandlerWizardMessageHandle(t *testing.T) {
	tempDir := t.TempDir()
	svc := character.NewPersistentService(filepath.Join(tempDir, "fichas_test.json"))
	h := NewCharacterHandler(svc)

	userID := "user_456"
	channelID := "chan_789"

	// Quando não há sessão, não intercepta
	m1 := &discordgo.MessageCreate{
		Message: &discordgo.Message{
			Author:    &discordgo.User{ID: userID, Bot: false},
			ChannelID: channelID,
			Content:   "Olá bot",
		},
	}
	if h.Handle(nil, m1) {
		t.Errorf("Mensagem normal não deveria ser interceptada quando não há sessão")
	}

	// Inicia sessão
	_, _ = h.wizard.StartSession(userID, channelID)

	// Agora a mensagem deve ser interceptada pelo wizard
	m2 := &discordgo.MessageCreate{
		Message: &discordgo.Message{
			Author:    &discordgo.User{ID: userID, Bot: false},
			ChannelID: channelID,
			Content:   "Kaelen",
		},
	}
	if !h.Handle(nil, m2) {
		t.Errorf("Mensagem deveria ser interceptada pelo wizard durante criação ativa")
	}
}

func TestHasSavageAttacker(t *testing.T) {
	sheetWithout := &character.CharacterSheet{
		Name: "Guerreiro Comum",
		Features: []character.Feature{
			{Name: "Segundo Fôlego"},
		},
	}
	if hasSavageAttacker(sheetWithout) {
		t.Errorf("Não deveria ter atacante selvagem")
	}

	sheetWithPT := &character.CharacterSheet{
		Name: "Samanta Jaci",
		Features: []character.Feature{
			{Name: "Atacante Selvagem", Description: "Rola dano duas vezes"},
		},
	}
	if !hasSavageAttacker(sheetWithPT) {
		t.Errorf("Deveria detectar 'Atacante Selvagem'")
	}

	sheetWithEN := &character.CharacterSheet{
		Name: "Barbarian",
		Features: []character.Feature{
			{Name: "Savage Attacker feat"},
		},
	}
	if !hasSavageAttacker(sheetWithEN) {
		t.Errorf("Deveria detectar 'Savage Attacker'")
	}
}

func TestFormatDamageBreakdown(t *testing.T) {
	// 1 dado com modificador positivo
	roll1 := &dice.RollResult{
		Rolls:    []int{5},
		Modifier: 3,
		Total:    8,
	}
	str1 := formatDamageBreakdown(roll1, "Piercing")
	if !strings.Contains(str1, "[5]") || !strings.Contains(str1, "+ 3") || !strings.Contains(str1, "8") || !strings.Contains(str1, "Piercing") {
		t.Errorf("Formatação incorreta: %s", str1)
	}

	// 2 dados com modificador
	roll2 := &dice.RollResult{
		Rolls:    []int{3, 4},
		Modifier: 2,
		Total:    9,
	}
	str2 := formatDamageBreakdown(roll2, "Cortante")
	if !strings.Contains(str2, "[3, 4]") || !strings.Contains(str2, "+ 2") || !strings.Contains(str2, "9") {
		t.Errorf("Formatação incorreta: %s", str2)
	}

	// 1 dado sem modificador
	roll3 := &dice.RollResult{
		Rolls:    []int{6},
		Modifier: 0,
		Total:    6,
	}
	str3 := formatDamageBreakdown(roll3, "Perfurante")
	if !strings.Contains(str3, "[6]") || !strings.Contains(str3, "6") {
		t.Errorf("Formatação incorreta: %s", str3)
	}
}
