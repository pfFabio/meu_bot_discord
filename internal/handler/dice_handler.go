package handler

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/bwmarrin/discordgo"
	"kenobi-bot/internal/dice"
	"kenobi-bot/internal/userpref"
)

// DiceHandler cuida do processamento de rolagens de dados por chat e Slash Command
type DiceHandler struct {
	roller      dice.DiceRoller
	prefService userpref.Service
}

// NewDiceHandler cria uma nova instância do DiceHandler
func NewDiceHandler(roller dice.DiceRoller, prefService userpref.Service) *DiceHandler {
	return &DiceHandler{
		roller:      roller,
		prefService: prefService,
	}
}

// ShouldHandle verifica se a mensagem é uma notação de dados válida
func (h *DiceHandler) ShouldHandle(authorID, botID, content string) bool {
	if authorID == botID {
		return false
	}
	content = strings.TrimSpace(content)
	if !h.roller.IsDiceNotation(content) {
		return false
	}
	// Valida se os limites (ex: <= d100) são respeitados
	_, err := h.roller.Roll(content)
	return err == nil
}

// Handle processa mensagens de texto no chat com comandos de dados (ex: d20, 2d10, 1d8+2)
func (h *DiceHandler) Handle(s *discordgo.Session, m *discordgo.MessageCreate) bool {
	botID := ""
	if s != nil && s.State != nil && s.State.User != nil {
		botID = s.State.User.ID
	}

	if !h.ShouldHandle(m.Author.ID, botID, m.Content) {
		return false
	}

	// Consultar preferências do jogador
	opts := dice.RollOptions{}
	if h.prefService != nil {
		pref := h.prefService.GetPreferences(m.Author.ID)
		if pref.RerollOn1All {
			opts.RerollOn1 = true
		} else if pref.HalflingLuck && strings.Contains(strings.ToLower(m.Content), "d20") {
			opts.RerollOn1 = true
		}
	}

	res, err := h.roller.RollWithOptions(m.Content, opts)
	if err != nil {
		return false
	}

	if s != nil {
		reply := fmt.Sprintf("<@%s> %s", m.Author.ID, res.Format())
		_, sendErr := s.ChannelMessageSend(m.ChannelID, reply)
		if sendErr != nil {
			slog.Error("Falha ao enviar rolagem de dados",
				"channel_id", m.ChannelID,
				"error", sendErr,
			)
		}
	}

	return true
}

// HandleSlashCommand processa o comando de barra /rolar ou /sorte
func (h *DiceHandler) HandleSlashCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	cmdName := i.ApplicationCommandData().Name
	userID := ""
	if i.Member != nil && i.Member.User != nil {
		userID = i.Member.User.ID
	} else if i.User != nil {
		userID = i.User.ID
	}

	// Comando de configuração de sorte / rerrolagem
	if cmdName == "sorte" || cmdName == "dados_config" {
		h.sendConfigPanel(s, i, userID)
		return
	}

	// Comando /rolar
	data := i.ApplicationCommandData()
	notation := "1d20"

	for _, opt := range data.Options {
		if opt.Name == "expressao" || opt.Name == "dado" {
			notation = opt.StringValue()
		}
	}

	opts := dice.RollOptions{}
	if h.prefService != nil && userID != "" {
		pref := h.prefService.GetPreferences(userID)
		if pref.RerollOn1All {
			opts.RerollOn1 = true
		} else if pref.HalflingLuck && strings.Contains(strings.ToLower(notation), "d20") {
			opts.RerollOn1 = true
		}
	}

	res, err := h.roller.RollWithOptions(notation, opts)
	if err != nil {
		_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("❌ Erro ao rolar dados: %s. Exemplo válido: `/rolar expressao: 2d10+3` ou `1d20`", err.Error()),
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("<@%s> %s", userID, res.Format()),
		},
	})
}

// HandleComponentInteraction responde a cliques nos botões de configuração
func (h *DiceHandler) HandleComponentInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	customID := i.MessageComponentData().CustomID
	userID := ""
	if i.Member != nil && i.Member.User != nil {
		userID = i.Member.User.ID
	} else if i.User != nil {
		userID = i.User.ID
	}

	switch customID {
	case "dice:toggle:halfling":
		if h.prefService != nil {
			h.prefService.ToggleHalflingLuck(userID)
		}
		h.updateConfigPanel(s, i, userID)
	case "dice:toggle:reroll1":
		if h.prefService != nil {
			h.prefService.ToggleRerollOn1All(userID)
		}
		h.updateConfigPanel(s, i, userID)
	}
}

func (h *DiceHandler) sendConfigPanel(s *discordgo.Session, i *discordgo.InteractionCreate, userID string) {
	embed, components := h.buildConfigUI(userID)
	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds:     []*discordgo.MessageEmbed{embed},
			Components: components,
			Flags:      discordgo.MessageFlagsEphemeral, // Privado para o próprio jogador
		},
	})
}

func (h *DiceHandler) updateConfigPanel(s *discordgo.Session, i *discordgo.InteractionCreate, userID string) {
	embed, components := h.buildConfigUI(userID)
	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Embeds:     []*discordgo.MessageEmbed{embed},
			Components: components,
		},
	})
}

func (h *DiceHandler) buildConfigUI(userID string) (*discordgo.MessageEmbed, []discordgo.MessageComponent) {
	pref := userpref.UserPreferences{}
	if h.prefService != nil {
		pref = h.prefService.GetPreferences(userID)
	}

	halflingStatus := "🔴 Desativado"
	halflingStyle := discordgo.SecondaryButton
	if pref.HalflingLuck {
		halflingStatus = "🟢 Ativado (Rerrola 1 no d20)"
		halflingStyle = discordgo.SuccessButton
	}

	reroll1Status := "🔴 Desativado"
	reroll1Style := discordgo.SecondaryButton
	if pref.RerollOn1All {
		reroll1Status = "🟢 Ativado (Rerrola 1 em Todos os Dados)"
		reroll1Style = discordgo.SuccessButton
	}

	embed := &discordgo.MessageEmbed{
		Title:       "⚙️ Configurações de Rolagem & Habilidades de Rerrolagem",
		Description: "Configure aqui suas habilidades automáticas de rerrolagem de dados de D&D.\nQuando ativadas, o bot rerrolará automaticamente se sair o número 1 e exibirá o dado anterior riscado (ex: `[~~1~~ ➔ 15]`).",
		Color:       0x52B788,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "🍀 Sorte Halfling (Halfling Luck)",
				Value:  fmt.Sprintf("Status: **%s**\n*Rerrola automaticamente o número 1 em jogadas de d20.*", halflingStatus),
				Inline: false,
			},
			{
				Name:   "⚔️ Luta com Armas Grandes / Rerrolar 1s (GWF & Geral)",
				Value:  fmt.Sprintf("Status: **%s**\n*Rerrola automaticamente o número 1 em qualquer dado de dano ou geral.*", reroll1Status),
				Inline: false,
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Clique nos botões abaixo para Ativar ou Desativar para o seu usuário.",
		},
	}

	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "🍀 Sorte Halfling (d20)",
					Style:    halflingStyle,
					CustomID: "dice:toggle:halfling",
				},
				discordgo.Button{
					Label:    "⚔️ Rerrolar 1s em Tudo",
					Style:    reroll1Style,
					CustomID: "dice:toggle:reroll1",
				},
			},
		},
	}

	return embed, components
}
