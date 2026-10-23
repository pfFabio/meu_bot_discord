package handler

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/bwmarrin/discordgo"
	"kenobi-bot/internal/audio"
)

// RecordingHandler gerencia os slash commands de gravação de sessão de RPG
type RecordingHandler struct {
	client *audio.BridgeClient
}

// NewRecordingHandler cria uma nova instância do RecordingHandler
func NewRecordingHandler(client *audio.BridgeClient) *RecordingHandler {
	return &RecordingHandler{client: client}
}

// GetSlashCommands retorna os comandos slash de gravação
func (h *RecordingHandler) GetSlashCommands() []*discordgo.ApplicationCommand {
	return []*discordgo.ApplicationCommand{
		{
			Name:        "gravar",
			Description: "🎙️ Inicia ou para a gravação da sessão de RPG",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "acao",
					Description: "Ação de gravação",
					Required:    true,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{Name: "🔴 Iniciar Gravação", Value: "iniciar"},
						{Name: "⏹️ Parar Gravação", Value: "parar"},
						{Name: "📊 Status da Gravação", Value: "status"},
					},
				},
			},
		},
	}
}

// HandleSlashCommand processa o comando /gravar
func (h *RecordingHandler) HandleSlashCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	options := i.ApplicationCommandData().Options
	if len(options) == 0 {
		return
	}

	action := options[0].StringValue()
	guildID := i.GuildID

	switch action {
	case "iniciar":
		h.handleStart(s, i, guildID)
	case "parar":
		h.handleStop(s, i, guildID)
	case "status":
		h.handleStatus(s, i, guildID)
	}
}

func (h *RecordingHandler) handleStart(s *discordgo.Session, i *discordgo.InteractionCreate, guildID string) {
	// Encontrar o canal de voz do usuário que invocou o comando
	vs, err := s.State.VoiceState(guildID, i.Member.User.ID)
	if err != nil || vs.ChannelID == "" {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "❌ Você precisa estar em um canal de voz para iniciar a gravação.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	// Defer response (conexão pode demorar)
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})

	result, err := h.client.StartRecording(guildID, vs.ChannelID)
	if err != nil {
		errMsg := fmt.Sprintf("❌ Erro ao iniciar gravação: %s", err.Error())
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: &errMsg})
		return
	}

	slog.Info("Gravação de sessão iniciada",
		"guild_id", guildID,
		"channel_id", vs.ChannelID,
		"session_dir", result.SessionDir,
	)

	embed := &discordgo.MessageEmbed{
		Title:       "🔴 Gravação de Sessão Iniciada",
		Description: "A sessão de RPG está sendo gravada!\n\nUse `/gravar parar` quando a sessão terminar.",
		Color:       0xED4245, // Vermelho Discord
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "📂 Sessão",
				Value:  fmt.Sprintf("`%s`", result.SessionID),
				Inline: true,
			},
			{
				Name:   "🎤 Status",
				Value:  "Capturando áudio...",
				Inline: true,
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: "⚠️ Todos os participantes estão cientes de que a sessão está sendo gravada.",
		},
	}

	s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Embeds: &[]*discordgo.MessageEmbed{embed},
	})
}

func (h *RecordingHandler) handleStop(s *discordgo.Session, i *discordgo.InteractionCreate, guildID string) {
	// Defer response (conversão de áudio pode demorar)
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})

	result, err := h.client.StopRecording(guildID)
	if err != nil {
		errMsg := fmt.Sprintf("❌ Erro ao parar gravação: %s", err.Error())
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: &errMsg})
		return
	}

	// Formatar duração
	secs := result.DurationMs / 1000
	mins := secs / 60
	hours := mins / 60
	durationStr := fmt.Sprintf("%02d:%02d:%02d", hours, mins%60, secs%60)

	slog.Info("Gravação de sessão finalizada",
		"guild_id", guildID,
		"duration", durationStr,
		"files", len(result.Files),
	)

	filesInfo := "Nenhum áudio capturado"
	if len(result.Files) > 0 {
		filesInfo = fmt.Sprintf("%d faixa(s) de áudio", len(result.Files))
	}

	embed := &discordgo.MessageEmbed{
		Title:       "⏹️ Gravação Finalizada",
		Description: "A gravação da sessão de RPG foi encerrada com sucesso!",
		Color:       0x57F287, // Verde Discord
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "⏱️ Duração",
				Value:  fmt.Sprintf("`%s`", durationStr),
				Inline: true,
			},
			{
				Name:   "🎵 Arquivos",
				Value:  fmt.Sprintf("`%s`", filesInfo),
				Inline: true,
			},
		},
	}

	if result.MergedFile != "" {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:  "📦 Arquivo Mesclado",
			Value: "`sessao_completa.ogg`",
		})
	}

	var attachFiles []*discordgo.File
	txtPath := filepath.Join(result.SessionDir, "transcricao.txt")
	if b, err := os.ReadFile(txtPath); err == nil && len(b) > 0 {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:  "📝 Transcrição da Sessão",
			Value: "Identificação automática de cada jogador por canal de áudio!",
		})

		preview := string(b)
		lines := strings.Split(preview, "\n")
		var previewLines []string
		for _, l := range lines {
			if strings.HasPrefix(l, "[") {
				previewLines = append(previewLines, l)
				if len(previewLines) >= 5 {
					break
				}
			}
		}

		if len(previewLines) > 0 {
			previewText := strings.Join(previewLines, "\n")
			if len(previewText) > 600 {
				previewText = previewText[:600] + "..."
			}
			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
				Name:  "💬 Falas Registradas",
				Value: fmt.Sprintf("```\n%s\n```", previewText),
			})
		}

		// Anexar o arquivo transcricao.txt para download direto no Discord
		if f, err := os.Open(txtPath); err == nil {
			defer f.Close()
			attachFiles = append(attachFiles, &discordgo.File{
				Name:        "transcricao_sessao.txt",
				ContentType: "text/plain; charset=utf-8",
				Reader:      f,
			})
		}
	}

	embed.Footer = &discordgo.MessageEmbedFooter{
		Text: fmt.Sprintf("Pasta: %s", result.SessionDir),
	}

	webhookEdit := &discordgo.WebhookEdit{
		Embeds: &[]*discordgo.MessageEmbed{embed},
	}
	if len(attachFiles) > 0 {
		webhookEdit.Files = attachFiles
	}

	s.InteractionResponseEdit(i.Interaction, webhookEdit)
}

func (h *RecordingHandler) handleStatus(s *discordgo.Session, i *discordgo.InteractionCreate, guildID string) {
	status, err := h.client.GetRecordingStatus(guildID)
	if err != nil || !status.IsRecording {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "📭 Nenhuma gravação ativa neste servidor.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	secs := status.DurationMs / 1000
	mins := secs / 60
	hours := mins / 60
	durationStr := fmt.Sprintf("%02d:%02d:%02d", hours, mins%60, secs%60)

	embed := &discordgo.MessageEmbed{
		Title: "🔴 Gravação em Andamento",
		Color: 0xED4245,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "⏱️ Tempo",
				Value:  fmt.Sprintf("`%s`", durationStr),
				Inline: true,
			},
			{
				Name:   "👥 Participantes Gravados",
				Value:  fmt.Sprintf("`%d`", status.Participants),
				Inline: true,
			},
		},
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})
}
