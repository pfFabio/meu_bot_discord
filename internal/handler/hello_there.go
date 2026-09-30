// Package handler - HelloThereHandler implementa MessageHandler.
// Princípio SOLID: Single Responsibility — este handler cuida exclusivamente
// da interação "hello there" → "General Kenobi!".
package handler

import (
	"log/slog"
	"strings"

	"github.com/bwmarrin/discordgo"
)

// HelloThereHandler responde "General Kenobi!" quando alguém
// envia "hello there" em qualquer canal visível pelo bot.
type HelloThereHandler struct{}

// NewHelloThereHandler cria uma nova instância do handler.
// Usar construtor facilita futuras injeções de dependência (ex: logger, DB).
func NewHelloThereHandler() *HelloThereHandler {
	return &HelloThereHandler{}
}

// ShouldHandle verifica a regra de negócio se a mensagem deve ser processada
func (h *HelloThereHandler) ShouldHandle(authorID, botID, content string) bool {
	if authorID == botID {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(content), "hello there")
}

// Handle verifica se a mensagem contém "hello there" (case-insensitive)
// e responde com "General Kenobi!" no mesmo canal.
func (h *HelloThereHandler) Handle(s *discordgo.Session, m *discordgo.MessageCreate) bool {
	botID := ""
	if s != nil && s.State != nil && s.State.User != nil {
		botID = s.State.User.ID
	}

	if !h.ShouldHandle(m.Author.ID, botID, m.Content) {
		return false
	}

	if s != nil {
		_, err := s.ChannelMessageSend(m.ChannelID, "**General Kenobi!** ⚔️")
		if err != nil {
			slog.Error("Falha ao enviar resposta",
				"channel_id", m.ChannelID,
				"error", err,
			)
		}
	}
	return true
}
