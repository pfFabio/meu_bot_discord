package handler

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"kenobi-bot/internal/audio"
)

// AudioHandler processa comandos slash e botões de áudio e música usando o Bridge DAVE E2EE
type AudioHandler struct {
	client *audio.BridgeClient
}

// NewAudioHandler cria uma nova instância do AudioHandler
func NewAudioHandler(client *audio.BridgeClient) *AudioHandler {
	return &AudioHandler{
		client: client,
	}
}

// GetSlashCommands retorna as definições dos comandos slash de música e som
func (h *AudioHandler) GetSlashCommands() []*discordgo.ApplicationCommand {
	return []*discordgo.ApplicationCommand{
		{
			Name:        "play",
			Description: "🎵 Toca uma música local, preset de D&D ou link do YouTube/SoundCloud",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:         discordgo.ApplicationCommandOptionString,
					Name:         "busca",
					Description:  "Nome da música local, preset (taverna, combate) ou link",
					Required:     false,
					Autocomplete: true,
				},
			},
		},
		{
			Name:        "tocar",
			Description: "🎵 Lista as músicas da pasta para escolher ou toca uma música diretamente",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:         discordgo.ApplicationCommandOptionString,
					Name:         "busca",
					Description:  "Nome da música ou link (deixe vazio para listar todas as músicas da pasta)",
					Required:     false,
					Autocomplete: true,
				},
			},
		},
		{
			Name:        "musicas",
			Description: "📂 Lista e toca as músicas locais da pasta musicas (Baróvia / D&D)",
		},
		{
			Name:        "ambiente",
			Description: "🏰 Painel de Trilhas Sonoras e Sons de Ambiente para Mesas de D&D",
		},
		{
			Name:        "parar",
			Description: "⏹️ Interrompe a música e limpa a fila de reprodução",
		},
		{
			Name:        "pular",
			Description: "⏭️ Pula para a próxima música da fila",
		},
		{
			Name:        "fila",
			Description: "📜 Exibe a lista de músicas na fila de reprodução",
		},
		{
			Name:        "loop",
			Description: "🔁 Ativa ou altera a repetição da música atual ou de toda a fila",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "modo",
					Description: "Modo de repetição (música atual, toda a fila ou desativar)",
					Required:    false,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{Name: "🔂 Repetir Música Atual", Value: "track"},
						{Name: "🔁 Repetir Toda a Fila", Value: "queue"},
						{Name: "❌ Desativar Repetição", Value: "off"},
					},
				},
			},
		},
		{
			Name:        "repetir",
			Description: "🔁 Repete a música atual ou toda a fila (atalho do /loop)",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "modo",
					Description: "Modo de repetição (música atual, toda a fila ou desativar)",
					Required:    false,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{Name: "🔂 Repetir Música Atual", Value: "track"},
						{Name: "🔁 Repetir Toda a Fila", Value: "queue"},
						{Name: "❌ Desativar Repetição", Value: "off"},
					},
				},
			},
		},
		{
			Name:        "sair",
			Description: "👋 Desconecta o bot do canal de voz",
		},
	}
}

// HandleSlashCommand despacha os comandos de áudio
func (h *AudioHandler) HandleSlashCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	cmdName := i.ApplicationCommandData().Name
	guildID := i.GuildID
	if guildID == "" {
		_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "❌ Comandos de áudio só podem ser usados dentro de um servidor.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	switch cmdName {
	case "play", "tocar":
		h.handlePlay(s, i)
	case "musicas":
		h.handleLocalMusicPanel(s, i)
	case "ambiente":
		h.handleAmbientePanel(s, i)
	case "loop", "repetir":
		h.handleLoop(s, i)
	case "parar", "stop":
		_ = h.client.Stop(guildID)
		_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "⏹️ Reprodução de áudio interrompida e fila limpa!",
			},
		})
	case "pular", "skip":
		skipped, err := h.client.Skip(guildID)
		if err != nil || !skipped {
			_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "⚠️ Nenhuma música na fila para pular.",
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
		} else {
			_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "⏭️ Faixa pulada com sucesso!",
				},
			})
		}
	case "fila", "queue":
		h.handleQueue(s, i)
	case "sair", "leave":
		_ = h.client.Leave(guildID)
		_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "👋 Desconectado do canal de voz.",
			},
		})
	}
}

func (h *AudioHandler) handlePlay(s *discordgo.Session, i *discordgo.InteractionCreate) {
	query := ""
	for _, opt := range i.ApplicationCommandData().Options {
		if opt.Name == "busca" {
			query = strings.TrimSpace(opt.StringValue())
		}
	}

	// Se não informou busca (ou digitou "musicas"/"lista"), lista todas as músicas da pasta para o usuário escolher
	if query == "" || strings.EqualFold(query, "musicas") || strings.EqualFold(query, "lista") || strings.EqualFold(query, "listar") {
		h.handleLocalMusicPanel(s, i)
		return
	}

	voiceChannelID := h.findUserVoiceChannel(s, i.GuildID, i.Member.User.ID)
	if voiceChannelID == "" {
		_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "❌ Você precisa estar conectado a um canal de voz para tocar áudio!",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	// Deferir resposta para permitir tempo de busca
	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})

	username := i.Member.User.Username
	playResp, err := h.client.Play(i.GuildID, voiceChannelID, query, username)
	if err != nil {
		_, _ = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: fmt.Sprintf("❌ Erro ao tocar áudio: %v", err),
		})
		return
	}

	switch playResp.Status {
	case "playing":
		_, _ = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: fmt.Sprintf("🎶 **Tocando agora:** `%s` (Duração: %s) • *Pedido por: %s*",
				playResp.Track.Title, playResp.Track.Duration, username),
		})
	case "playlist":
		_, _ = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: fmt.Sprintf("📑 **Playlist Adicionada:** **%d faixas** enfileiradas!\n▶️ **Tocando agora:** `%s` • *Pedido por: %s*",
				playResp.Count, playResp.Track.Title, username),
		})
	case "playlist_queued":
		_, _ = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: fmt.Sprintf("📑 **Playlist Adicionada à Fila:** **%d faixas** adicionadas à fila! • *Pedido por: %s*",
				playResp.Count, username),
		})
	default:
		_, _ = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: fmt.Sprintf("➕ **Adicionado à fila (#%d):** `%s` • *Pedido por: %s*",
				playResp.Position, playResp.Track.Title, username),
		})
	}
}

// handleLoop ativa, altera ou desativa a repetição da música atual ou de toda a fila
func (h *AudioHandler) handleLoop(s *discordgo.Session, i *discordgo.InteractionCreate) {
	mode := ""
	for _, opt := range i.ApplicationCommandData().Options {
		if opt.Name == "modo" {
			mode = opt.StringValue()
		}
	}

	newMode, err := h.client.Loop(i.GuildID, mode)
	if err != nil {
		_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("⚠️ %v", err),
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	var msg string
	switch newMode {
	case "track":
		msg = "🔂 **Modo Loop Ativado:** A **música atual** será repetida continuamente!\n*Use `/loop modo: Desativar` ou o botão 🔁 para desligar.*"
	case "queue":
		msg = "🔁 **Modo Loop Ativado:** A **fila inteira** será repetida continuamente!\n*Use `/loop modo: Desativar` ou o botão 🔁 para desligar.*"
	default:
		msg = "▶️ **Modo Loop Desativado.** As músicas serão tocadas normalmente uma única vez."
	}

	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: msg,
		},
	})
}

// handleLocalMusicPanel exibe o painel interativo com as músicas locais da pasta musicas
func (h *AudioHandler) handleLocalMusicPanel(s *discordgo.Session, i *discordgo.InteractionCreate) {
	localTracks, _ := h.client.GetLocalTracks()
	if len(localTracks) == 0 {
		_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "📂 Nenhuma música encontrada na pasta `musicas/`. Adicione arquivos `.mp3`, `.wav`, `.ogg` na pasta do bot!",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	var fields []*discordgo.MessageEmbedField
	var menuOptions []discordgo.SelectMenuOption

	for idx, track := range localTracks {
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   fmt.Sprintf("%s %s", track.Emoji, track.Title),
			Value:  fmt.Sprintf("📄 `%s` • ⏱️ Duração: %s", track.Filename, track.Duration),
			Inline: false,
		})

		if idx < 25 {
			menuOptions = append(menuOptions, discordgo.SelectMenuOption{
				Label:       track.Title,
				Value:       "local:" + track.Filename,
				Description: fmt.Sprintf("Arquivo: %s (%s)", track.Filename, track.Duration),
				Emoji:       &discordgo.ComponentEmoji{Name: track.Emoji},
				Default:     false,
			})
		}
	}

	embed := &discordgo.MessageEmbed{
		Title:       "🎵 Escolha uma Música para Tocar",
		Description: "Faixas disponíveis na pasta `musicas/`.\nSelecione no menu suspenso ou clique nos botões rápidos abaixo:",
		Color:       0x800020, // Tom Bordô / Baróvia
		Fields:      fields,
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("Kenobi Bot • %d música(s) disponível(is) na pasta", len(localTracks)),
		},
	}

	var components []discordgo.MessageComponent

	// 1. Menu Dropdown
	components = append(components, discordgo.ActionsRow{
		Components: []discordgo.MessageComponent{
			discordgo.SelectMenu{
				CustomID:    "audio:select_local",
				Placeholder: "🎵 Escolha uma música para tocar agora...",
				Options:     menuOptions,
			},
		},
	})

	// 2. Botões Rápidos (até 4 faixas em destaque)
	var quickButtons []discordgo.MessageComponent
	for idx, track := range localTracks {
		if idx >= 4 {
			break
		}
		btnLabel := track.Title
		if len(btnLabel) > 20 {
			btnLabel = btnLabel[:17] + "..."
		}
		quickButtons = append(quickButtons, discordgo.Button{
			Label:    btnLabel,
			Style:    discordgo.PrimaryButton,
			CustomID: "audio:local:" + track.Filename,
			Emoji:    &discordgo.ComponentEmoji{Name: track.Emoji},
		})
	}

	if len(quickButtons) > 0 {
		components = append(components, discordgo.ActionsRow{
			Components: quickButtons,
		})
	}

	// 3. Controles de Reprodução (Parar, Pular, Fila, Loop, Sair)
	components = append(components, discordgo.ActionsRow{
		Components: []discordgo.MessageComponent{
			discordgo.Button{
				Label:    "Parar",
				Style:    discordgo.SecondaryButton,
				CustomID: "audio:action:stop",
				Emoji:    &discordgo.ComponentEmoji{Name: "⏹️"},
			},
			discordgo.Button{
				Label:    "Pular",
				Style:    discordgo.SecondaryButton,
				CustomID: "audio:action:skip",
				Emoji:    &discordgo.ComponentEmoji{Name: "⏭️"},
			},
			discordgo.Button{
				Label:    "Fila",
				Style:    discordgo.SecondaryButton,
				CustomID: "audio:action:queue",
				Emoji:    &discordgo.ComponentEmoji{Name: "📜"},
			},
			discordgo.Button{
				Label:    "Loop",
				Style:    discordgo.SecondaryButton,
				CustomID: "audio:action:loop",
				Emoji:    &discordgo.ComponentEmoji{Name: "🔁"},
			},
			discordgo.Button{
				Label:    "Sair",
				Style:    discordgo.SecondaryButton,
				CustomID: "audio:action:leave",
				Emoji:    &discordgo.ComponentEmoji{Name: "👋"},
			},
		},
	})

	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds:     []*discordgo.MessageEmbed{embed},
			Components: components,
		},
	})
}

func (h *AudioHandler) handleAmbientePanel(s *discordgo.Session, i *discordgo.InteractionCreate) {
	embed := &discordgo.MessageEmbed{
		Title:       "🏰 Painel de Áudio Ambiente para Mesas de D&D",
		Description: "Escolha uma trilha sonora atmosférica contínua para sua sessão de RPG.\n*O bot entrará no seu canal de voz usando criptografia DAVE E2EE e tocará a trilha escolhida.*",
		Color:       0x4A6FA5,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "🍻 Taverna Medieval", Value: "Alaúdes, flautas e o calor de uma taberna acolhedora.", Inline: true},
			{Name: "⚔️ Combate Épico", Value: "Tambores de guerra e orquestra para batalhas tensas.", Inline: true},
			{Name: "🕯️ Masmorra & Caverna", Value: "Gotejamento, ecos sombrios e mistério subterrâneo.", Inline: true},
			{Name: "🌿 Floresta Feérica", Value: "Sons da natureza, harpas místicas e brisa suave.", Inline: true},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: "D&D 2024 Soundboard • Clique abaixo para tocar",
		},
	}

	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "Taverna",
					Style:    discordgo.PrimaryButton,
					CustomID: "audio:preset:taverna",
					Emoji:    &discordgo.ComponentEmoji{Name: "🍻"},
				},
				discordgo.Button{
					Label:    "Combate",
					Style:    discordgo.DangerButton,
					CustomID: "audio:preset:combate",
					Emoji:    &discordgo.ComponentEmoji{Name: "⚔️"},
				},
				discordgo.Button{
					Label:    "Masmorra",
					Style:    discordgo.SecondaryButton,
					CustomID: "audio:preset:masmorra",
					Emoji:    &discordgo.ComponentEmoji{Name: "🕯️"},
				},
				discordgo.Button{
					Label:    "Floresta",
					Style:    discordgo.SuccessButton,
					CustomID: "audio:preset:floresta",
					Emoji:    &discordgo.ComponentEmoji{Name: "🌿"},
				},
			},
		},
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "Repetir (Loop)",
					Style:    discordgo.SecondaryButton,
					CustomID: "audio:action:loop",
					Emoji:    &discordgo.ComponentEmoji{Name: "🔁"},
				},
				discordgo.Button{
					Label:    "Parar Áudio",
					Style:    discordgo.SecondaryButton,
					CustomID: "audio:action:stop",
					Emoji:    &discordgo.ComponentEmoji{Name: "⏹️"},
				},
				discordgo.Button{
					Label:    "Desconectar",
					Style:    discordgo.SecondaryButton,
					CustomID: "audio:action:leave",
					Emoji:    &discordgo.ComponentEmoji{Name: "👋"},
				},
			},
		},
	}

	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds:     []*discordgo.MessageEmbed{embed},
			Components: components,
		},
	})
}

// HandleComponentInteraction gerencia cliques no painel de ambiente
func (h *AudioHandler) HandleComponentInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	customID := i.MessageComponentData().CustomID
	guildID := i.GuildID

	if strings.HasPrefix(customID, "audio:preset:") {
		presetKey := strings.TrimPrefix(customID, "audio:preset:")
		voiceChannelID := h.findUserVoiceChannel(s, guildID, i.Member.User.ID)
		if voiceChannelID == "" {
			_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "❌ Conecte-se a um canal de voz primeiro para tocar a trilha!",
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
			return
		}

		_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		})

		_ = h.client.Stop(guildID)
		playResp, err := h.client.Play(guildID, voiceChannelID, presetKey, i.Member.User.Username)
		if err != nil {
			_, _ = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
				Content: fmt.Sprintf("❌ Erro ao iniciar trilha: %v", err),
			})
			return
		}

		_, _ = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: fmt.Sprintf("🏰 **Iniciando Trilha de D&D:** `%s`", playResp.Track.Title),
		})
		return
	}

	if customID == "audio:select_local" {
		values := i.MessageComponentData().Values
		if len(values) == 0 {
			return
		}
		selectedValue := values[0]
		voiceChannelID := h.findUserVoiceChannel(s, guildID, i.Member.User.ID)
		if voiceChannelID == "" {
			_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "❌ Você precisa estar conectado a um canal de voz para tocar áudio!",
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
			return
		}

		_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		})

		playResp, err := h.client.Play(guildID, voiceChannelID, selectedValue, i.Member.User.Username)
		if err != nil {
			_, _ = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
				Content: fmt.Sprintf("❌ Erro ao tocar música local: %v", err),
			})
			return
		}

		if playResp.Status == "playing" {
			_, _ = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
				Content: fmt.Sprintf("🎶 **Tocando agora:** `%s` (Duração: %s) • *Pedido por: %s*",
					playResp.Track.Title, playResp.Track.Duration, i.Member.User.Username),
			})
		} else {
			_, _ = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
				Content: fmt.Sprintf("➕ **Adicionado à fila (#%d):** `%s` • *Pedido por: %s*",
					playResp.Position, playResp.Track.Title, i.Member.User.Username),
			})
		}
		return
	}

	if strings.HasPrefix(customID, "audio:local:") {
		filename := strings.TrimPrefix(customID, "audio:local:")
		voiceChannelID := h.findUserVoiceChannel(s, guildID, i.Member.User.ID)
		if voiceChannelID == "" {
			_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "❌ Você precisa estar conectado a um canal de voz para tocar áudio!",
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
			return
		}

		_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		})

		playResp, err := h.client.Play(guildID, voiceChannelID, "local:"+filename, i.Member.User.Username)
		if err != nil {
			_, _ = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
				Content: fmt.Sprintf("❌ Erro ao iniciar música: %v", err),
			})
			return
		}

		if playResp.Status == "playing" {
			_, _ = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
				Content: fmt.Sprintf("🎶 **Tocando agora:** `%s` (Duração: %s) • *Pedido por: %s*",
					playResp.Track.Title, playResp.Track.Duration, i.Member.User.Username),
			})
		} else {
			_, _ = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
				Content: fmt.Sprintf("➕ **Adicionado à fila (#%d):** `%s` • *Pedido por: %s*",
					playResp.Position, playResp.Track.Title, i.Member.User.Username),
			})
		}
		return
	}

	if customID == "audio:action:stop" {
		_ = h.client.Stop(guildID)
		_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "⏹️ Reprodução de áudio interrompida e fila limpa!",
			},
		})
		return
	}

	if customID == "audio:action:skip" {
		skipped, err := h.client.Skip(guildID)
		if err != nil || !skipped {
			_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "⚠️ Nenhuma música na fila para pular.",
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
		} else {
			_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "⏭️ Faixa pulada com sucesso!",
				},
			})
		}
		return
	}

	if customID == "audio:action:queue" {
		h.handleQueue(s, i)
		return
	}

	if customID == "audio:action:loop" {
		newMode, err := h.client.Loop(guildID, "")
		if err != nil {
			_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("⚠️ %v", err),
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
			return
		}

		var msg string
		switch newMode {
		case "track":
			msg = "🔂 **Repetição da Música Atual ATIVADA!** (Loop)"
		case "queue":
			msg = "🔁 **Repetição da Fila ATIVADA!** (Loop)"
		default:
			msg = "▶️ **Repetição DESATIVADA.**"
		}

		_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: msg,
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	if customID == "audio:action:leave" {
		_ = h.client.Leave(guildID)
		_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "👋 Desconectado do canal de voz.",
			},
		})
		return
	}
}

// HandleAutocomplete processa sugestões em tempo real para os comandos /play e /tocar
func (h *AudioHandler) HandleAutocomplete(s *discordgo.Session, i *discordgo.InteractionCreate) {
	data := i.ApplicationCommandData()
	var currentQuery string
	for _, opt := range data.Options {
		if opt.Focused {
			currentQuery = strings.TrimSpace(opt.StringValue())
			break
		}
	}

	lowerQuery := strings.ToLower(currentQuery)
	var choices []*discordgo.ApplicationCommandOptionChoice

	// 1. Sugerir faixas locais da pasta musicas
	localTracks := audio.ScanLocalMusic()
	for _, track := range localTracks {
		nameWithEmoji := fmt.Sprintf("%s %s (Local)", track.Emoji, track.Title)
		if lowerQuery == "" ||
			strings.Contains(strings.ToLower(track.Title), lowerQuery) ||
			strings.Contains(strings.ToLower(track.Filename), lowerQuery) {
			choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
				Name:  nameWithEmoji,
				Value: "local:" + track.Filename,
			})
		}
		if len(choices) >= 15 {
			break
		}
	}

	// 2. Sugerir Presets de D&D
	presets := audio.GetAmbientPresets()
	for _, p := range presets {
		nameWithEmoji := fmt.Sprintf("%s %s (Preset)", p.Emoji, p.Name)
		if lowerQuery == "" ||
			strings.Contains(strings.ToLower(p.Name), lowerQuery) ||
			strings.Contains(strings.ToLower(p.Key), lowerQuery) {
			choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
				Name:  nameWithEmoji,
				Value: p.Key,
			})
		}
		if len(choices) >= 25 {
			break
		}
	}

	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionApplicationCommandAutocompleteResult,
		Data: &discordgo.InteractionResponseData{
			Choices: choices,
		},
	})
}

func (h *AudioHandler) handleQueue(s *discordgo.Session, i *discordgo.InteractionCreate) {
	qResp, err := h.client.GetQueue(i.GuildID)
	if err != nil || (qResp.Current == nil && len(qResp.Queue) == 0) {
		_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "📜 A fila de reprodução está vazia no momento.",
			},
		})
		return
	}

	var fields []*discordgo.MessageEmbedField
	if qResp.Current != nil {
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   "▶️ Tocando Agora",
			Value:  fmt.Sprintf("**%s** (Duração: %s)\n*Pedido por: %s*", qResp.Current.Title, qResp.Current.Duration, qResp.Current.Requester),
			Inline: false,
		})
	}

	if len(qResp.Queue) > 0 {
		var queueList []string
		for idx, t := range qResp.Queue {
			if idx >= 10 {
				queueList = append(queueList, fmt.Sprintf("... e mais %d faixas na fila", len(qResp.Queue)-10))
				break
			}
			queueList = append(queueList, fmt.Sprintf("`%d.` **%s** • *%s*", idx+1, t.Title, t.Requester))
		}
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   "📜 A Seguir na Fila",
			Value:  strings.Join(queueList, "\n"),
			Inline: false,
		})
	}

	var loopStatus string
	switch qResp.LoopMode {
	case "track":
		loopStatus = "🔂 Música Atual (Loop)"
	case "queue":
		loopStatus = "🔁 Toda a Fila (Loop)"
	default:
		loopStatus = "Desativado"
	}

	embed := &discordgo.MessageEmbed{
		Title:  "🎵 Fila de Reprodução",
		Color:  0x3A86FF,
		Fields: fields,
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("Modo de Repetição: %s", loopStatus),
		},
	}

	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})
}

// findUserVoiceChannel encontra o canal de voz onde o usuário está conectado no servidor
func (h *AudioHandler) findUserVoiceChannel(s *discordgo.Session, guildID, userID string) string {
	guild, err := s.State.Guild(guildID)
	if err != nil || guild == nil {
		g, fetchErr := s.Guild(guildID)
		if fetchErr != nil {
			return ""
		}
		guild = g
	}

	for _, vs := range guild.VoiceStates {
		if vs.UserID == userID {
			return vs.ChannelID
		}
	}
	return ""
}
