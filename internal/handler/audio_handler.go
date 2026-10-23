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
				{
					Type:        discordgo.ApplicationCommandOptionBoolean,
					Name:        "proxima",
					Description: "⏭️ Tocar a seguir (coloca como a próxima da fila sem parar a atual)",
					Required:    false,
				},
				{
					Type:        discordgo.ApplicationCommandOptionBoolean,
					Name:        "transicao",
					Description: "✨ Transição suave com fade-out da música atual (sem corte seco)",
					Required:    false,
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
				{
					Type:        discordgo.ApplicationCommandOptionBoolean,
					Name:        "proxima",
					Description: "⏭️ Tocar a seguir (coloca como a próxima da fila sem parar a atual)",
					Required:    false,
				},
				{
					Type:        discordgo.ApplicationCommandOptionBoolean,
					Name:        "transicao",
					Description: "✨ Transição suave com fade-out da música atual (sem corte seco)",
					Required:    false,
				},
			},
		},
		{
			Name:        "musicas",
			Description: "📂 Lista e toca as músicas locais da pasta musicas (Baróvia / D&D)",
		},
		{
			Name:        "temas",
			Description: "🎭 Temas e Trilhas Sonoras de Personagens, Batalhas e Relíquias (Curse of Strahd)",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:         discordgo.ApplicationCommandOptionString,
					Name:         "busca",
					Description:  "Nome do personagem, vilão, combate ou relíquia (deixe vazio para abrir o painel)",
					Required:     false,
					Autocomplete: true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionBoolean,
					Name:        "proxima",
					Description: "⏭️ Tocar a seguir (coloca como a próxima da fila sem parar a atual)",
					Required:    false,
				},
				{
					Type:        discordgo.ApplicationCommandOptionBoolean,
					Name:        "transicao",
					Description: "✨ Transição suave com fade-out da música atual (sem corte seco)",
					Required:    false,
				},
			},
		},
		{
			Name:        "ambience",
			Description: "🏰 Painel de Sons de Ambiente e Cenários de Baróvia (Curse of Strahd)",
		},
		{
			Name:        "ambiente",
			Description: "🏰 Painel de Sons de Ambiente e Cenários de Baróvia (Curse of Strahd)",
		},
		{
			Name:        "parar",
			Description: "⏹️ Interrompe a música e limpa a fila de reprodução",
		},
		{
			Name:        "pular",
			Description: "⏭️ Pula para a próxima música da fila",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionBoolean,
					Name:        "suave",
					Description: "✨ Fazer transição suave (fade-out) em vez de corte imediato",
					Required:    false,
				},
			},
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
	case "temas", "tema":
		h.handleTemasPanel(s, i)
	case "ambiente", "ambience":
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
		suave := false
		for _, opt := range i.ApplicationCommandData().Options {
			if opt.Name == "suave" {
				suave = opt.BoolValue()
			}
		}
		skipped, err := h.client.SkipWithOptions(guildID, audio.SkipOptions{
			Transition:   suave,
			FadeDuration: 1500,
		})
		if err != nil || !skipped {
			_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "⚠️ Nenhuma música na fila para pular.",
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
		} else {
			transText := ""
			if suave {
				transText = " *(com transição suave ✨)*"
			}
			_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("⏭️ Faixa pulada com sucesso%s!", transText),
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
	playNext := false
	transition := false
	for _, opt := range i.ApplicationCommandData().Options {
		switch opt.Name {
		case "busca":
			query = strings.TrimSpace(opt.StringValue())
		case "proxima":
			playNext = opt.BoolValue()
		case "transicao":
			transition = opt.BoolValue()
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
	playResp, err := h.client.PlayWithOptions(i.GuildID, voiceChannelID, query, username, audio.PlayOptions{
		PlayNext:     playNext,
		Transition:   transition,
		FadeDuration: 1500,
	})
	if err != nil {
		_, _ = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: fmt.Sprintf("❌ Erro ao tocar áudio: %v", err),
		})
		return
	}

	switch playResp.Status {
	case "next":
		_, _ = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: fmt.Sprintf("⏭️ **Definida como a PRÓXIMA música a tocar:** `%s` • *Pedido por: %s*",
				playResp.Track.Title, username),
		})
	case "playing":
		transText := ""
		if transition {
			transText = " *(com transição suave ✨)*"
		}
		_, _ = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: fmt.Sprintf("🎶 **Tocando agora%s:** `%s` (Duração: %s) • *Pedido por: %s*",
				transText, playResp.Track.Title, playResp.Track.Duration, username),
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
	case "queue":
		msg = "🔁 **Modo Loop Ativado:** A **fila inteira** será repetida continuamente!\n*Músicas adicionadas entram automaticamente no ciclo. Use `/loop modo: Desativar` ou o botão 🔁 para desligar.*"
	case "track":
		msg = "🔂 **Modo Loop Ativado:** A **música atual** será repetida continuamente!\n*Músicas adicionadas à fila expandem e integram o ciclo de repetição.*"
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
	ambienceList := audio.GetCurseOfStrahdAmbienceList()
	availableTracks := audio.GetAvailableAmbienceTracks()

	var options []discordgo.SelectMenuOption
	for idx, track := range ambienceList {
		if idx >= 25 {
			break
		}
		desc := track.Description
		if len(desc) > 100 {
			desc = desc[:97] + "..."
		}
		label := track.Title
		if len(label) > 100 {
			label = label[:97] + "..."
		}
		options = append(options, discordgo.SelectMenuOption{
			Label:       label,
			Value:       "local:" + track.Filename,
			Description: desc,
			Emoji:       &discordgo.ComponentEmoji{Name: track.Emoji},
		})
	}

	embed := &discordgo.MessageEmbed{
		Title:       "🏰 Painel de Áudio Ambiente de Baróvia • Curse of Strahd",
		Description: "Trilhas sonoras atmosféricas contínuas para cada locação e momento de Baróvia.\n" +
			"Selecione o cenário no menu suspenso ou clique nos botões rápidos de locais mais frequentes!\n" +
			"*Ao escolher um cenário, o áudio anterior é interrompido e a nova atmosfera se inicia.*",
		Color:       0x800020, // Bordô gótico de Baróvia
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "🌲 Exploração & Regiões",
				Value:  "• **Svalich Woods** • **Névoas de Ravenloft**\n• **Passo de Tsolenka** • **Ruínas de Berez**",
				Inline: true,
			},
			{
				Name:   "🏘️ Povoados & Civilização",
				Value:  "• **Vila de Baróvia** • **Cidade de Vallaki**\n• **Vila de Krezk** • **Mago dos Vinhos**",
				Inline: true,
			},
			{
				Name:   "🏰 Masmorras & Covis Sombrios",
				Value:  "• **Castelo Ravenloft** • **Catacumbas**\n• **Casa da Morte & Masmorra** • **Templo de Âmbar**",
				Inline: true,
			},
			{
				Name:   "⛪ Marcos & Encontros Notáveis",
				Value:  "• **Velho Moedor de Ossos** • **Argynvostholt**\n• **Abadia de Santa Markovia** • **Torre de Van Richten**",
				Inline: false,
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("Curse of Strahd Ambience • %d/%d faixas locais prontas • Kenobi Bot", len(availableTracks), len(ambienceList)),
		},
	}

	var components []discordgo.MessageComponent

	// 1. Menu Dropdown com todos os 18 cenários de Baróvia
	if len(options) > 0 {
		components = append(components, discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.SelectMenu{
					CustomID:    "audio:select_ambience",
					Placeholder: "🏰 Selecione o cenário de Baróvia para tocar...",
					Options:     options,
				},
			},
		})
	}

	// 2. Botões Rápidos dos locais mais icônicos
	components = append(components, discordgo.ActionsRow{
		Components: []discordgo.MessageComponent{
			discordgo.Button{
				Label:    "Svalich Woods",
				Style:    discordgo.SuccessButton,
				CustomID: "audio:ambience:CoS - Svalich Woods.mp3",
				Emoji:    &discordgo.ComponentEmoji{Name: "🌲"},
			},
			discordgo.Button{
				Label:    "Vila de Baróvia",
				Style:    discordgo.PrimaryButton,
				CustomID: "audio:ambience:CoS - Village of Barovia.mp3",
				Emoji:    &discordgo.ComponentEmoji{Name: "🏚️"},
			},
			discordgo.Button{
				Label:    "Vallaki",
				Style:    discordgo.PrimaryButton,
				CustomID: "audio:ambience:CoS - Town Of Vallaki.mp3",
				Emoji:    &discordgo.ComponentEmoji{Name: "🏘️"},
			},
			discordgo.Button{
				Label:    "Castelo Ravenloft",
				Style:    discordgo.DangerButton,
				CustomID: "audio:ambience:CoS - Castle Ravenloft.mp3",
				Emoji:    &discordgo.ComponentEmoji{Name: "🏰"},
			},
			discordgo.Button{
				Label:    "Névoas",
				Style:    discordgo.SecondaryButton,
				CustomID: "audio:ambience:CoS - Mists of Ravenloft.mp3",
				Emoji:    &discordgo.ComponentEmoji{Name: "🌫️"},
			},
		},
	})

	// 3. Controles de Reprodução (Loop, Parar, Desconectar)
	components = append(components, discordgo.ActionsRow{
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
	})

	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds:     []*discordgo.MessageEmbed{embed},
			Components: components,
		},
	})
}

// handleTemasPanel processa o comando /temas (busca direta ou painel de seleção)
func (h *AudioHandler) handleTemasPanel(s *discordgo.Session, i *discordgo.InteractionCreate) {
	query := ""
	playNext := false
	transition := false
	for _, opt := range i.ApplicationCommandData().Options {
		switch opt.Name {
		case "busca":
			query = strings.TrimSpace(opt.StringValue())
		case "proxima":
			playNext = opt.BoolValue()
		case "transicao":
			transition = opt.BoolValue()
		}
	}

	if query != "" {
		track, found := audio.FindThemeTrack(query)
		if !found {
			_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("⚠️ Não encontrei nenhum tema correspondente a `%s`.\n💡 Use `/temas` sem opções para abrir o painel completo com todos os 94 temas de Baróvia!", query),
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
			return
		}

		voiceChannelID := h.findUserVoiceChannel(s, i.GuildID, i.Member.User.ID)
		if voiceChannelID == "" {
			_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "❌ Você precisa estar conectado a um canal de voz para tocar um tema!",
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
			return
		}

		_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		})

		target := "local:" + track.Filename
		if !track.IsAvailable && track.YouTubeURL != "" {
			target = track.YouTubeURL
		}

		username := i.Member.User.Username
		playResp, err := h.client.PlayWithOptions(i.GuildID, voiceChannelID, target, username, audio.PlayOptions{
			PlayNext:     playNext,
			Transition:   transition,
			FadeDuration: 1500,
		})
		if err != nil {
			_, _ = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
				Content: fmt.Sprintf("❌ Erro ao tocar tema de Baróvia: %v", err),
			})
			return
		}

		switch playResp.Status {
		case "next":
			_, _ = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
				Content: fmt.Sprintf("⏭️ **Definido como PRÓXIMO tema a tocar:** `%s` • *Pedido por: %s*",
					playResp.Track.Title, username),
			})
		case "playing":
			transText := ""
			if transition {
				transText = " *(com transição suave ✨)*"
			}
			_, _ = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
				Content: fmt.Sprintf("🎭 **Tema de Baróvia Iniciado%s:** `%s` (Duração: %s)\n🔊 *Trilha sonora iniciada por %s!*",
					transText, playResp.Track.Title, playResp.Track.Duration, username),
			})
		default:
			_, _ = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
				Content: fmt.Sprintf("➕ **Tema Adicionado à Fila (#%d):** `%s` (Duração: %s)\n🔁 *Integrado à lista de reprodução e ao loop por %s!*",
					playResp.Position, playResp.Track.Title, playResp.Track.Duration, username),
			})
		}
		return
	}

	embed, components := h.renderTemasPanelData("todos")
	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds:     []*discordgo.MessageEmbed{embed},
			Components: components,
		},
	})
}

// renderTemasPanelData monta o embed e componentes interativos do painel de temas
func (h *AudioHandler) renderTemasPanelData(selectedCategory string) (*discordgo.MessageEmbed, []discordgo.MessageComponent) {
	allTracks := audio.GetThemeTracks()
	categories := audio.GetThemeCategories()

	var displayTracks []audio.ThemeTrack
	if selectedCategory != "" && selectedCategory != "todos" {
		displayTracks = audio.GetThemeTracksByCategory(selectedCategory)
	} else {
		displayTracks = allTracks
	}
	if len(displayTracks) > 25 {
		displayTracks = displayTracks[:25]
	}

	var menuOptions []discordgo.SelectMenuOption
	for _, track := range displayTracks {
		desc := fmt.Sprintf("%s • %s", track.Category, track.Duration)
		if len(desc) > 100 {
			desc = desc[:97] + "..."
		}
		lbl := track.Title
		if len(lbl) > 100 {
			lbl = lbl[:97] + "..."
		}
		val := "local:" + track.Filename
		if !track.IsAvailable && track.YouTubeURL != "" {
			val = track.YouTubeURL
		}
		menuOptions = append(menuOptions, discordgo.SelectMenuOption{
			Label:       lbl,
			Value:       val,
			Description: desc,
			Emoji:       &discordgo.ComponentEmoji{Name: track.Emoji},
		})
	}

	var catOptions []discordgo.SelectMenuOption
	catOptions = append(catOptions, discordgo.SelectMenuOption{
		Label:       "🎭 Todos os Temas (Destaques)",
		Value:       "todos",
		Description: "Exibir temas e trilhas principais de Baróvia",
		Emoji:       &discordgo.ComponentEmoji{Name: "🎭"},
		Default:     selectedCategory == "" || selectedCategory == "todos",
	})
	for _, cat := range categories {
		isDef := strings.EqualFold(cat, selectedCategory)
		catOptions = append(catOptions, discordgo.SelectMenuOption{
			Label:       cat,
			Value:       cat,
			Description: fmt.Sprintf("Filtrar faixas de %s", cat),
			Default:     isDef,
		})
	}

	embed := &discordgo.MessageEmbed{
		Title:       "🎭 Painel de Temas e Trilhas Sonoras • Curse of Strahd",
		Description: "Trilhas sonoras completas de personagens, encontros dramáticos, combates e relíquias de Baróvia.\n" +
			"Selecione uma faixa no menu suspenso ou nos botões rápidos para tocar ou adicionar à fila!\n" +
			"*Músicas adicionadas à fila integram automaticamente o ciclo de Loop ativo.*",
		Color:       0x800020, // Bordô gótico
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "👑 Personagens & NPCs",
				Value:  "• Strahd von Zarovich • Ireena Kolyana\n• Ismark • Rahadin • Noivas • Van Richten",
				Inline: true,
			},
			{
				Name:   "⚔️ Combates & Batalhas",
				Value:  "• Batalha Strahd • Encontro no Pântano\n• Casa da Morte • Bruxas • Deva",
				Inline: true,
			},
			{
				Name:   "✨ Relíquias & Destino",
				Value:  "• Espada Solar • Símbolo Sagrado\n• Leitura de Tarokka • Tomo de Strahd",
				Inline: true,
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("Kenobi Bot • %d temas catalogados de Curse of Strahd", len(allTracks)),
		},
	}

	var components []discordgo.MessageComponent

	// 1. Menu de faixas da categoria selecionada
	components = append(components, discordgo.ActionsRow{
		Components: []discordgo.MessageComponent{
			discordgo.SelectMenu{
				CustomID:    "audio:theme:select",
				Placeholder: fmt.Sprintf("🎵 Escolha um tema (%d faixas listadas)...", len(displayTracks)),
				Options:     menuOptions,
			},
		},
	})

	// 2. Menu seletor de categoria
	components = append(components, discordgo.ActionsRow{
		Components: []discordgo.MessageComponent{
			discordgo.SelectMenu{
				CustomID:    "audio:theme:category",
				Placeholder: "📂 Filtrar temas por Categoria...",
				Options:     catOptions,
			},
		},
	})

	// 3. Botões Rápidos
	components = append(components, discordgo.ActionsRow{
		Components: []discordgo.MessageComponent{
			discordgo.Button{
				Label:    "Strahd von Zarovich",
				Style:    discordgo.DangerButton,
				CustomID: "audio:theme:quick:CoS - Strahd von Zarovich.mp3",
				Emoji:    &discordgo.ComponentEmoji{Name: "🩸"},
			},
			discordgo.Button{
				Label:    "Batalha: Strahd",
				Style:    discordgo.DangerButton,
				CustomID: "audio:theme:quick:CoS - Strahd Battle Theme.mp3",
				Emoji:    &discordgo.ComponentEmoji{Name: "⚔️"},
			},
			discordgo.Button{
				Label:    "Ireena Kolyana",
				Style:    discordgo.PrimaryButton,
				CustomID: "audio:theme:quick:CoS - Ireena Kolyana Theme.mp3",
				Emoji:    &discordgo.ComponentEmoji{Name: "👑"},
			},
			discordgo.Button{
				Label:    "Espada Solar",
				Style:    discordgo.SuccessButton,
				CustomID: "audio:theme:quick:CoS - Sunsword.mp3",
				Emoji:    &discordgo.ComponentEmoji{Name: "✨"},
			},
		},
	})

	// 4. Controles (Loop, Pular, Parar, Fila, Sair)
	components = append(components, discordgo.ActionsRow{
		Components: []discordgo.MessageComponent{
			discordgo.Button{
				Label:    "Loop",
				Style:    discordgo.SecondaryButton,
				CustomID: "audio:action:loop",
				Emoji:    &discordgo.ComponentEmoji{Name: "🔁"},
			},
			discordgo.Button{
				Label:    "Pular",
				Style:    discordgo.SecondaryButton,
				CustomID: "audio:action:skip",
				Emoji:    &discordgo.ComponentEmoji{Name: "⏭️"},
			},
			discordgo.Button{
				Label:    "Parar",
				Style:    discordgo.SecondaryButton,
				CustomID: "audio:action:stop",
				Emoji:    &discordgo.ComponentEmoji{Name: "⏹️"},
			},
			discordgo.Button{
				Label:    "Fila",
				Style:    discordgo.SecondaryButton,
				CustomID: "audio:action:queue",
				Emoji:    &discordgo.ComponentEmoji{Name: "📜"},
			},
			discordgo.Button{
				Label:    "Sair",
				Style:    discordgo.SecondaryButton,
				CustomID: "audio:action:leave",
				Emoji:    &discordgo.ComponentEmoji{Name: "👋"},
			},
		},
	})

	return embed, components
}

// HandleComponentInteraction gerencia cliques no painel de ambiente e temas
func (h *AudioHandler) HandleComponentInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	customID := i.MessageComponentData().CustomID
	guildID := i.GuildID

	// Interações do Painel de Temas
	if customID == "audio:theme:select" || strings.HasPrefix(customID, "audio:theme:quick:") {
		var targetTrack string
		if customID == "audio:theme:select" {
			values := i.MessageComponentData().Values
			if len(values) == 0 {
				return
			}
			targetTrack = values[0]
		} else {
			filename := strings.TrimPrefix(customID, "audio:theme:quick:")
			targetTrack = "local:" + filename
		}

		voiceChannelID := h.findUserVoiceChannel(s, guildID, i.Member.User.ID)
		if voiceChannelID == "" {
			_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "❌ Conecte-se a um canal de voz primeiro para tocar o tema!",
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
			return
		}

		_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		})

		playResp, err := h.client.Play(guildID, voiceChannelID, targetTrack, i.Member.User.Username)
		if err != nil {
			_, _ = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
				Content: fmt.Sprintf("❌ Erro ao iniciar tema: %v", err),
			})
			return
		}

		if playResp.Status == "playing" {
			_, _ = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
				Content: fmt.Sprintf("🎭 **Tema de Baróvia Iniciado:** `%s`\n🔊 *Trilha temática iniciada por %s!*",
					playResp.Track.Title, i.Member.User.Username),
			})
		} else {
			_, _ = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
				Content: fmt.Sprintf("➕ **Tema Adicionado à Fila (#%d):** `%s`\n🔁 *Integrado ao loop de reprodução por %s!*",
					playResp.Position, playResp.Track.Title, i.Member.User.Username),
			})
		}
		return
	}

	if customID == "audio:theme:category" {
		values := i.MessageComponentData().Values
		selectedCat := "todos"
		if len(values) > 0 {
			selectedCat = values[0]
		}
		embed, components := h.renderTemasPanelData(selectedCat)
		_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseUpdateMessage,
			Data: &discordgo.InteractionResponseData{
				Embeds:     []*discordgo.MessageEmbed{embed},
				Components: components,
			},
		})
		return
	}

	if customID == "audio:select_ambience" || strings.HasPrefix(customID, "audio:ambience:") {
		var targetTrack string
		if customID == "audio:select_ambience" {
			values := i.MessageComponentData().Values
			if len(values) == 0 {
				return
			}
			targetTrack = values[0]
		} else {
			filename := strings.TrimPrefix(customID, "audio:ambience:")
			targetTrack = "local:" + filename
		}

		voiceChannelID := h.findUserVoiceChannel(s, guildID, i.Member.User.ID)
		if voiceChannelID == "" {
			_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "❌ Conecte-se a um canal de voz primeiro para tocar o áudio ambiente!",
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
			return
		}

		_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		})

		playResp, err := h.client.Play(guildID, voiceChannelID, targetTrack, i.Member.User.Username)
		if err != nil {
			_, _ = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
				Content: fmt.Sprintf("❌ Erro ao iniciar ambiente de Baróvia: %v", err),
			})
			return
		}

		if playResp.Status == "playing" {
			_, _ = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
				Content: fmt.Sprintf("🏰 **Entrando no Cenário:** `%s`\n🔊 *Música ambiente de Baróvia iniciada por %s!*",
					playResp.Track.Title, i.Member.User.Username),
			})
		} else {
			_, _ = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
				Content: fmt.Sprintf("➕ **Ambiente Adicionado à Fila (#%d):** `%s`\n🔁 *Integrado ao loop de reprodução por %s!*",
					playResp.Position, playResp.Track.Title, i.Member.User.Username),
			})
		}
		return
	}

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

		playResp, err := h.client.Play(guildID, voiceChannelID, presetKey, i.Member.User.Username)
		if err != nil {
			_, _ = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
				Content: fmt.Sprintf("❌ Erro ao iniciar trilha: %v", err),
			})
			return
		}

		if playResp.Status == "playing" {
			_, _ = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
				Content: fmt.Sprintf("🏰 **Iniciando Trilha de D&D:** `%s`", playResp.Track.Title),
			})
		} else {
			_, _ = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
				Content: fmt.Sprintf("➕ **Trilha Adicionada à Fila (#%d):** `%s`\n🔁 *Integrada ao loop de reprodução!*",
					playResp.Position, playResp.Track.Title),
			})
		}
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
		case "queue":
			msg = "🔁 **Repetição da Fila ATIVADA!** (Loop de todas as faixas adicionadas)"
		case "track":
			msg = "🔂 **Repetição da Música Atual ATIVADA!** (Loop da faixa atual)"
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

// HandleAutocomplete processa sugestões em tempo real para os comandos /play, /tocar e /temas
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

	// Se for comando /temas, sugere exclusivamente faixas temáticas de Curse of Strahd
	if data.Name == "temas" || data.Name == "tema" {
		themeTracks := audio.GetThemeTracks()
		for _, track := range themeTracks {
			if lowerQuery == "" ||
				strings.Contains(strings.ToLower(track.Title), lowerQuery) ||
				strings.Contains(strings.ToLower(track.Category), lowerQuery) ||
				strings.Contains(strings.ToLower(track.Key), lowerQuery) ||
				strings.Contains(strings.ToLower(track.Filename), lowerQuery) {
				choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
					Name:  fmt.Sprintf("%s %s (%s)", track.Emoji, track.Title, track.Category),
					Value: track.Filename,
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
		return
	}

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
		if len(choices) >= 20 {
			break
		}
	}

	// 3. Sugerir Temas de Curse of Strahd se ainda couber
	if len(choices) < 25 {
		themeTracks := audio.GetThemeTracks()
		for _, track := range themeTracks {
			if lowerQuery != "" &&
				(strings.Contains(strings.ToLower(track.Title), lowerQuery) ||
					strings.Contains(strings.ToLower(track.Filename), lowerQuery)) {
				choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
					Name:  fmt.Sprintf("%s %s (Tema CoS)", track.Emoji, track.Title),
					Value: "local:" + track.Filename,
				})
			}
			if len(choices) >= 25 {
				break
			}
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
