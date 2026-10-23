package handler

import (
	"fmt"
	"log/slog"
	"strconv"
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

// ShouldHandle verifica se a mensagem é uma notação de dados válida ou comando de sucessos
func (h *DiceHandler) ShouldHandle(authorID, botID, content string) bool {
	if authorID == botID {
		return false
	}
	content = strings.TrimSpace(content)
	clean := strings.ToLower(content)
	if strings.HasPrefix(clean, "/") || strings.HasPrefix(clean, "!") {
		clean = clean[1:]
	}
	clean = strings.TrimSpace(clean)

	if strings.HasPrefix(clean, "sus") || strings.HasPrefix(clean, "sucs") ||
		strings.HasPrefix(clean, "limsus") || strings.HasPrefix(clean, "limsucs") ||
		strings.HasPrefix(clean, "limitesus") || strings.HasPrefix(clean, "limitesucs") {
		return true
	}

	if !h.roller.IsDiceNotation(content) {
		return false
	}
	// Valida se os limites (ex: <= d100) são respeitados
	_, err := h.roller.Roll(content)
	return err == nil
}

// Handle processa mensagens de texto no chat com comandos de dados (ex: d20, 2d10, sus 3d61d20, limsus 5)
func (h *DiceHandler) Handle(s *discordgo.Session, m *discordgo.MessageCreate) bool {
	botID := ""
	if s != nil && s.State != nil && s.State.User != nil {
		botID = s.State.User.ID
	}

	if !h.ShouldHandle(m.Author.ID, botID, m.Content) {
		return false
	}

	clean := strings.TrimSpace(strings.ToLower(m.Content))
	if strings.HasPrefix(clean, "/") || strings.HasPrefix(clean, "!") {
		clean = clean[1:]
	}
	clean = strings.TrimSpace(clean)

	// Caso A: Comando de configuração de limite de sucesso (limsus / limsucs / limitesus / limitesucs)
	if strings.HasPrefix(clean, "limsus") || strings.HasPrefix(clean, "limsucs") ||
		strings.HasPrefix(clean, "limitesus") || strings.HasPrefix(clean, "limitesucs") {
		return h.handleLimSusTextMessage(s, m, clean)
	}

	// Caso B: Comando de rolagem de sucessos (sus / sucs)
	if strings.HasPrefix(clean, "sus") || strings.HasPrefix(clean, "sucs") {
		return h.handleSusTextMessage(s, m, clean)
	}

	// Caso C: Rolagem padrão (d20, 2d10, 1d8+2)
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

func (h *DiceHandler) handleLimSusTextMessage(s *discordgo.Session, m *discordgo.MessageCreate, clean string) bool {
	parts := strings.Fields(clean)
	if len(parts) > 1 {
		val, err := strconv.Atoi(parts[1])
		if err != nil || val < 1 || val > 100 {
			if s != nil {
				_, _ = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("<@%s> ❌ O limite de sucesso deve ser um número entre 1 e 100.", m.Author.ID))
			}
			return true
		}
		if h.prefService != nil {
			h.prefService.SetSuccessThreshold(m.Author.ID, val)
		}
		if s != nil {
			reply := fmt.Sprintf("🎯 <@%s> **Limite de Sucesso configurado para ≥ %d!**\nAgora quando você rolar com `sus` ou `/sus`, dados com resultado **%d ou mais** serão contados como sucesso.", m.Author.ID, val, val)
			_, _ = s.ChannelMessageSend(m.ChannelID, reply)
		}
		return true
	}

	current := userpref.DefaultSuccessThreshold
	if h.prefService != nil {
		current = h.prefService.GetPreferences(m.Author.ID).SuccessThreshold
	}
	if s != nil {
		reply := fmt.Sprintf("🎯 <@%s> Seu limite de sucesso atual é **≥ %d**.\nPara alterar, digite `limsus <número>` (ex: `limsus 6`) ou `/limsus valor: 6`.", m.Author.ID, current)
		_, _ = s.ChannelMessageSend(m.ChannelID, reply)
	}
	return true
}

func (h *DiceHandler) handleSusTextMessage(s *discordgo.Session, m *discordgo.MessageCreate, clean string) bool {
	args := ""
	if strings.HasPrefix(clean, "sus") {
		args = strings.TrimSpace(clean[len("sus"):])
	} else if strings.HasPrefix(clean, "sucs") {
		args = strings.TrimSpace(clean[len("sucs"):])
	}

	if args == "" {
		args = "6d10"
	}

	pref := userpref.UserPreferences{SuccessThreshold: userpref.DefaultSuccessThreshold}
	if h.prefService != nil {
		pref = h.prefService.GetPreferences(m.Author.ID)
	}

	opts := dice.RollOptions{}
	if pref.RerollOn1All {
		opts.RerollOn1 = true
	} else if pref.HalflingLuck && strings.Contains(strings.ToLower(args), "d20") {
		opts.RerollOn1 = true
	}

	res, err := h.roller.RollSuccess(args, pref.SuccessThreshold, opts)
	if err != nil {
		if s != nil {
			reply := fmt.Sprintf("<@%s> ❌ Erro ao rolar sucessos: %s.\nExemplos: `sus 4d6+5d20+2d10`, `sus 6d10`, `sus 8d6`, `sus 6` ou `sus 4d6+5d20 5`.", m.Author.ID, err.Error())
			_, _ = s.ChannelMessageSend(m.ChannelID, reply)
		}
		return true
	}

	if s != nil {
		reply := fmt.Sprintf("<@%s>\n%s", m.Author.ID, res.Format())
		_, _ = s.ChannelMessageSend(m.ChannelID, reply)
	}
	return true
}

// HandleSlashCommand processa comandos de barra como /rolar, /sorte, /sus e /limsus
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

	// Comando de configuração de limite de sucesso
	if cmdName == "limsus" || cmdName == "limsucs" || cmdName == "limitesus" || cmdName == "limitesucs" {
		h.handleLimSusSlashCommand(s, i, userID)
		return
	}

	// Comando de rolagem de sucessos
	if cmdName == "sus" || cmdName == "sucs" {
		h.handleSusSlashCommand(s, i, userID)
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

func (h *DiceHandler) handleLimSusSlashCommand(s *discordgo.Session, i *discordgo.InteractionCreate, userID string) {
	data := i.ApplicationCommandData()
	var val *int64
	for _, opt := range data.Options {
		if opt.Name == "valor" || opt.Name == "limite" || opt.Name == "alvo" {
			v := opt.IntValue()
			val = &v
		}
	}

	if val != nil {
		newLimit := int(*val)
		if newLimit < 1 || newLimit > 100 {
			_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "❌ O valor do limite de sucesso deve ser entre 1 e 100.",
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
			return
		}
		if h.prefService != nil {
			h.prefService.SetSuccessThreshold(userID, newLimit)
		}
		_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("🎯 <@%s> **Limite de Sucesso configurado para ≥ %d!**\n"+
					"Agora em suas rolagens com `/sus`, qualquer dado tirando **%d ou mais** será contado como um sucesso.",
					userID, newLimit, newLimit),
			},
		})
		return
	}

	current := userpref.DefaultSuccessThreshold
	if h.prefService != nil {
		current = h.prefService.GetPreferences(userID).SuccessThreshold
	}
	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("🎯 <@%s> Seu limite de sucesso atual é **≥ %d**.\n"+
				"Para alterar, use `/limsus valor: <número>` (exemplo: `/limsus valor: 6`).",
				userID, current),
			Flags: discordgo.MessageFlagsEphemeral,
		},
	})
}

func (h *DiceHandler) handleSusSlashCommand(s *discordgo.Session, i *discordgo.InteractionCreate, userID string) {
	data := i.ApplicationCommandData()
	notation := ""
	var overrideTarget int

	for _, opt := range data.Options {
		if opt.Name == "dados" || opt.Name == "expressao" || opt.Name == "quantidade" {
			notation = strings.TrimSpace(opt.StringValue())
		}
		if opt.Name == "alvo" || opt.Name == "limite" || opt.Name == "valor" {
			overrideTarget = int(opt.IntValue())
		}
	}

	if notation == "" {
		notation = "6d10"
	}

	pref := userpref.UserPreferences{SuccessThreshold: userpref.DefaultSuccessThreshold}
	if h.prefService != nil && userID != "" {
		pref = h.prefService.GetPreferences(userID)
	}

	threshold := pref.SuccessThreshold
	if overrideTarget > 0 {
		threshold = overrideTarget
	}

	opts := dice.RollOptions{}
	if pref.RerollOn1All {
		opts.RerollOn1 = true
	} else if pref.HalflingLuck && strings.Contains(strings.ToLower(notation), "d20") {
		opts.RerollOn1 = true
	}

	res, err := h.roller.RollSuccess(notation, threshold, opts)
	if err != nil {
		_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("❌ Erro ao rolar sucessos: %s.\nExemplos válidos: `/sus dados: 4d6+5d20+2d10`, `/sus dados: 6d10` ou `/sus dados: 4d6+5d20 alvo: 6`.", err.Error()),
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("<@%s>\n%s", userID, res.Format()),
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
			{
				Name:   "🎯 Limite de Sucesso (/limsus)",
				Value:  fmt.Sprintf("Alvo atual para rolagens `/sus`: **≥ %d**\n*Use `/limsus valor: <número>` para alterar seu alvo padrão.*", pref.SuccessThreshold),
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
