// Package bot é responsável por inicializar a sessão do Discord,
// registrar os handlers de eventos, comandos slash e interações de componentes.
package bot

import (
	"log/slog"
	"strings"

	"github.com/bwmarrin/discordgo"
	"kenobi-bot/internal/handler"
)

// Bot encapsula a sessão do Discord, handlers de mensagem e handlers de comandos
type Bot struct {
	session          *discordgo.Session
	msgHandlers      []handler.MessageHandler
	dndHandler       *handler.DNDHandler
	diceHandler      *handler.DiceHandler
	audioHandler     *handler.AudioHandler
	recordingHandler *handler.RecordingHandler
	characterHandler *handler.CharacterHandler
	createdCommands  []*discordgo.ApplicationCommand
}

// New cria uma nova instância do Bot a partir de um token e dos handlers
func New(
	token string,
	dndHandler *handler.DNDHandler,
	diceHandler *handler.DiceHandler,
	audioHandler *handler.AudioHandler,
	recordingHandler *handler.RecordingHandler,
	characterHandler *handler.CharacterHandler,
) (*Bot, error) {
	session, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, err
	}

	session.Identify.Intents = discordgo.IntentsGuilds |
		discordgo.IntentsGuildMessages |
		discordgo.IntentMessageContent |
		discordgo.IntentsGuildVoiceStates

	return &Bot{
		session:          session,
		msgHandlers:      make([]handler.MessageHandler, 0),
		dndHandler:       dndHandler,
		diceHandler:      diceHandler,
		audioHandler:     audioHandler,
		recordingHandler: recordingHandler,
		characterHandler: characterHandler,
	}, nil
}

// Session retorna a sessão do DiscordGo subjacente
func (b *Bot) Session() *discordgo.Session {
	return b.session
}

// AddMessageHandler registra um MessageHandler na cadeia de processamento
func (b *Bot) AddMessageHandler(h handler.MessageHandler) {
	b.msgHandlers = append(b.msgHandlers, h)
}

// Start abre a conexão WebSocket com o Discord Gateway e registra comandos e listeners
func (b *Bot) Start() error {
	// Registrar listener para mensagens normais
	b.session.AddHandler(b.onMessageCreate)

	// Registrar listener para interações (Slash Commands, Select Menus, Botões)
	b.session.AddHandler(b.onInteractionCreate)

	if err := b.session.Open(); err != nil {
		return err
	}

	slog.Info("Bot conectado ao Discord com sucesso!",
		"user", b.session.State.User.Username+"#"+b.session.State.User.Discriminator,
	)

	// Registrar Slash Commands globalmente
	b.registerSlashCommands()

	return nil
}

// registerSlashCommands registra os comandos slash na API do Discord
func (b *Bot) registerSlashCommands() {
	appID := b.session.State.User.ID
	var commands []*discordgo.ApplicationCommand

	if b.dndHandler != nil {
		commands = append(commands, b.dndHandler.GetSlashCommands()...)
	}

	// Adicionar comando /rolar
	commands = append(commands, &discordgo.ApplicationCommand{
		Name:        "rolar",
		Description: "🎲 Rola dados de RPG (ex: d20, 2d10, 3d20k2, 4d6k3, 1d8+2 até d100)",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "expressao",
				Description: "Notação do dado (ex: d20, 2d10, 3d20k2, 4d6k3, 1d8+2, d100)",
				Required:    false,
			},
		},
	})

	// Adicionar comando /sorte (Configuração de Rerrolagem de 1s)
	commands = append(commands, &discordgo.ApplicationCommand{
		Name:        "sorte",
		Description: "🍀 Ativa/Desativa Rerrolagem Automática de 1s (Sorte Halfling / GWF)",
	})

	// Adicionar comando /sus (Rolagem com contagem de sucessos e falhas)
	commands = append(commands, &discordgo.ApplicationCommand{
		Name:        "sus",
		Description: "🎯 Rola dados e contabiliza sucessos e falhas (ex: 4d6+5d20+2d10 ou 6d10)",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "dados",
				Description: "Dados separados por + (ex: 4d6+5d20+2d10, 6d10 ou apenas 6)",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        "alvo",
				Description: "Valor mínimo para sucesso nesta rolagem (opcional; padrão usa o salvo no /limsus)",
				Required:    false,
			},
		},
	})

	// Adicionar comando /sucs (Alias de /sus)
	commands = append(commands, &discordgo.ApplicationCommand{
		Name:        "sucs",
		Description: "🎯 (Alias de /sus) Rola dados e contabiliza sucessos e falhas",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "dados",
				Description: "Dados separados por + (ex: 4d6+5d20+2d10, 6d10 ou 6)",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        "alvo",
				Description: "Valor mínimo para sucesso nesta rolagem (opcional)",
				Required:    false,
			},
		},
	})

	// Adicionar comando /limsus (Configuração de limite de sucesso)
	commands = append(commands, &discordgo.ApplicationCommand{
		Name:        "limsus",
		Description: "⚙️ Define ou consulta o valor mínimo no dado para contar sucesso (ex: 5 para ≥ 5)",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        "valor",
				Description: "Valor mínimo a ser batido no dado para ser um sucesso (ex: 5 para ≥ 5)",
				Required:    false,
			},
		},
	})

	// Adicionar comando /limsucs (Alias de /limsus)
	commands = append(commands, &discordgo.ApplicationCommand{
		Name:        "limsucs",
		Description: "⚙️ (Alias de /limsus) Define ou consulta o valor mínimo para sucesso",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        "valor",
				Description: "Valor mínimo a ser batido no dado para ser um sucesso (ex: 5 para ≥ 5)",
				Required:    false,
			},
		},
	})

	// Adicionar comando /limitesus (Alias de /limsus)
	commands = append(commands, &discordgo.ApplicationCommand{
		Name:        "limitesus",
		Description: "⚙️ (Alias) Define ou consulta o valor mínimo no dado para contar sucesso",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        "valor",
				Description: "Valor mínimo a ser batido no dado para ser um sucesso (ex: 5 para ≥ 5)",
				Required:    false,
			},
		},
	})

	// Adicionar comandos de áudio
	if b.audioHandler != nil {
		commands = append(commands, b.audioHandler.GetSlashCommands()...)
	}

	// Adicionar comandos de gravação de sessão
	if b.recordingHandler != nil {
		commands = append(commands, b.recordingHandler.GetSlashCommands()...)
	}

	// Adicionar comandos do Sistema de Fichas de Personagens D&D
	if b.characterHandler != nil {
		commands = append(commands, b.characterHandler.GetSlashCommands()...)
	}

	slog.Info("Registrando Slash Commands...", "total", len(commands))

	// 1. Registro em massa global (atômico e limpa comandos antigos)
	createdGlobal, err := b.session.ApplicationCommandBulkOverwrite(appID, "", commands)
	if err != nil {
		slog.Error("Falha ao registrar comandos slash globais", "error", err)
	} else {
		b.createdCommands = append(b.createdCommands, createdGlobal...)
		slog.Info("Slash Commands globais registrados com sucesso", "total", len(createdGlobal))
	}

	// 2. Registro imediato nas Guilds (servidores) onde o bot está presente (aparece na hora no Discord)
	guilds, err := b.session.UserGuilds(100, "", "", false)
	if err == nil {
		for _, g := range guilds {
			createdGuild, err := b.session.ApplicationCommandBulkOverwrite(appID, g.ID, commands)
			if err != nil {
				slog.Warn("Falha ao registrar comandos na guild", "guild_id", g.ID, "guild_name", g.Name, "error", err)
			} else {
				slog.Info("Slash Commands registrados instantaneamente na guild", "guild_name", g.Name, "total", len(createdGuild))
			}
		}
	} else {
		slog.Warn("Não foi possível listar guilds para registro direto", "error", err)
	}
}

// Stop encerra a conexão com o Discord de forma segura e faz cleanup
func (b *Bot) Stop() error {
	slog.Info("Encerrando conexão com o Discord...")
	return b.session.Close()
}

// onMessageCreate despacha mensagens para os handlers registrados
func (b *Bot) onMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	for _, h := range b.msgHandlers {
		if h.Handle(s, m) {
			return
		}
	}
}

// onInteractionCreate despacha Slash Commands, Component Interactions (Menus/Botões) e Modais
func (b *Bot) onInteractionCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		cmdName := i.ApplicationCommandData().Name
		if cmdName == "rolar" || cmdName == "roll" || cmdName == "sorte" || cmdName == "dados_config" ||
			cmdName == "sus" || cmdName == "sucs" || cmdName == "limsus" || cmdName == "limsucs" ||
			cmdName == "limitesus" || cmdName == "limitesucs" {
			if b.diceHandler != nil {
				b.diceHandler.HandleSlashCommand(s, i)
			}
			return
		}
		if cmdName == "play" || cmdName == "tocar" || cmdName == "parar" || cmdName == "stop" ||
			cmdName == "pular" || cmdName == "skip" || cmdName == "fila" || cmdName == "queue" ||
			cmdName == "sair" || cmdName == "leave" || cmdName == "ambiente" || cmdName == "ambience" ||
			cmdName == "musicas" || cmdName == "temas" || cmdName == "tema" || cmdName == "loop" || cmdName == "repetir" {
			if b.audioHandler != nil {
				b.audioHandler.HandleSlashCommand(s, i)
			}
			return
		}
		if cmdName == "gravar" {
			if b.recordingHandler != nil {
				b.recordingHandler.HandleSlashCommand(s, i)
			}
			return
		}
		if cmdName == "ficha" || cmdName == "hp" || cmdName == "magia" || cmdName == "descanso" {
			if b.characterHandler != nil {
				b.characterHandler.HandleSlashCommand(s, i)
			}
			return
		}
		if b.dndHandler != nil {
			b.dndHandler.HandleSlashCommand(s, i)
		}
	case discordgo.InteractionApplicationCommandAutocomplete:
		cmdName := i.ApplicationCommandData().Name
		if cmdName == "play" || cmdName == "tocar" || cmdName == "temas" || cmdName == "tema" {
			if b.audioHandler != nil {
				b.audioHandler.HandleAutocomplete(s, i)
			}
			return
		}
	case discordgo.InteractionMessageComponent:
		customID := i.MessageComponentData().CustomID
		if strings.HasPrefix(customID, "dice:") {
			if b.diceHandler != nil {
				b.diceHandler.HandleComponentInteraction(s, i)
			}
			return
		}
		if strings.HasPrefix(customID, "audio:") {
			if b.audioHandler != nil {
				b.audioHandler.HandleComponentInteraction(s, i)
			}
			return
		}
		if strings.HasPrefix(customID, "sheet:") {
			if b.characterHandler != nil {
				b.characterHandler.HandleComponentInteraction(s, i)
			}
			return
		}
		if b.dndHandler != nil {
			b.dndHandler.HandleComponentInteraction(s, i)
		}
	case discordgo.InteractionModalSubmit:
		customID := i.ModalSubmitData().CustomID
		if strings.HasPrefix(customID, "sheet:modal:") {
			if b.characterHandler != nil {
				b.characterHandler.HandleModalSubmit(s, i)
			}
			return
		}
	}
}
