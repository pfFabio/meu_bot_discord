package handler_test

import (
	"testing"

	"kenobi-bot/internal/handler"
)

func TestHelloThereHandler(t *testing.T) {
	h := handler.NewHelloThereHandler()

	tests := []struct {
		name         string
		content      string
		authorID     string
		botID        string
		expectHandle bool
	}{
		{
			name:         "Mensagem exata em minúsculas",
			content:      "hello there",
			authorID:     "user123",
			botID:        "bot999",
			expectHandle: true,
		},
		{
			name:         "Mensagem em maiúsculas (Case Insensitive)",
			content:      "HELLO THERE",
			authorID:     "user123",
			botID:        "bot999",
			expectHandle: true,
		},
		{
			name:         "Mensagem com espaços nas bordas",
			content:      "   hello there   ",
			authorID:     "user123",
			botID:        "bot999",
			expectHandle: true,
		},
		{
			name:         "Mensagem com pontuação / texto diferente",
			content:      "hello there!",
			authorID:     "user123",
			botID:        "bot999",
			expectHandle: false,
		},
		{
			name:         "Outra mensagem qualquer",
			content:      "bom dia pessoal",
			authorID:     "user123",
			botID:        "bot999",
			expectHandle: false,
		},
		{
			name:         "Mensagem enviada pelo próprio bot (Anti-Loop)",
			content:      "hello there",
			authorID:     "bot999",
			botID:        "bot999",
			expectHandle: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handled := h.ShouldHandle(tt.authorID, tt.botID, tt.content)
			if handled != tt.expectHandle {
				t.Errorf("ShouldHandle() = %v, esperado %v para mensagem %q (author: %s, bot: %s)",
					handled, tt.expectHandle, tt.content, tt.authorID, tt.botID)
			}
		})
	}
}
