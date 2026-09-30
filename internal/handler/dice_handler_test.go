package handler_test

import (
	"testing"

	"kenobi-bot/internal/dice"
	"kenobi-bot/internal/handler"
	"kenobi-bot/internal/userpref"
)

func TestDiceHandler(t *testing.T) {
	roller := dice.NewRoller()
	prefService := userpref.NewMemoryService()
	h := handler.NewDiceHandler(roller, prefService)

	tests := []struct {
		name         string
		content      string
		authorID     string
		botID        string
		expectHandle bool
	}{
		{
			name:         "Mensagem 'd20'",
			content:      "d20",
			authorID:     "user1",
			botID:        "bot1",
			expectHandle: true,
		},
		{
			name:         "Mensagem '2d10'",
			content:      "2d10",
			authorID:     "user1",
			botID:        "bot1",
			expectHandle: true,
		},
		{
			name:         "Mensagem '1d8+2'",
			content:      "1d8+2",
			authorID:     "user1",
			botID:        "bot1",
			expectHandle: true,
		},
		{
			name:         "Mensagem '4d6 - 1' com espaços",
			content:      "4d6 - 1",
			authorID:     "user1",
			botID:        "bot1",
			expectHandle: true,
		},
		{
			name:         "Mensagem 'd100'",
			content:      "d100",
			authorID:     "user1",
			botID:        "bot1",
			expectHandle: true,
		},
		{
			name:         "Mensagem com tag explícita de reroll 'd20 r1'",
			content:      "d20 r1",
			authorID:     "user1",
			botID:        "bot1",
			expectHandle: true,
		},
		{
			name:         "Mensagem com keep '3d20k2'",
			content:      "3d20k2",
			authorID:     "user1",
			botID:        "bot1",
			expectHandle: true,
		},
		{
			name:         "Mensagem com keep e modificador '4d6k3+2'",
			content:      "4d6k3+2",
			authorID:     "user1",
			botID:        "bot1",
			expectHandle: true,
		},
		{
			name:         "Mensagem inválida 'd101' (acima do limite d100)",
			content:      "d101",
			authorID:     "user1",
			botID:        "bot1",
			expectHandle: false,
		},
		{
			name:         "Mensagem de texto comum",
			content:      "oi galera como estão?",
			authorID:     "user1",
			botID:        "bot1",
			expectHandle: false,
		},
		{
			name:         "Mensagem do próprio bot",
			content:      "d20",
			authorID:     "bot1",
			botID:        "bot1",
			expectHandle: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handled := h.ShouldHandle(tt.authorID, tt.botID, tt.content)
			if handled != tt.expectHandle {
				t.Errorf("ShouldHandle() = %v, esperado %v para %q", handled, tt.expectHandle, tt.content)
			}
		})
	}

	t.Run("Integração com preferências de usuário (Halfling Luck e Reroll)", func(t *testing.T) {
		prefService.ToggleHalflingLuck("user_luck")
		pref := prefService.GetPreferences("user_luck")
		if !pref.HalflingLuck {
			t.Errorf("HalflingLuck deveria estar ativado para user_luck")
		}
	})
}
