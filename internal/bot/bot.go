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
	createdCommands  []*discordgo.ApplicationCommand
}

// New cria uma nova instância do Bot a partir de um token e dos handlers
func New(
	token string,
	dndHandler *handler.DNDHandler,
	diceHandler *handler.DiceHandler,
	audioHandler *handler.AudioHandler,
	recordingHandler *handler.RecordingHandler,
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
	}, nil
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

	// Adicionar comandos de áudio
	if b.audioHandler != nil {
		commands = append(commands, b.audioHandler.GetSlashCommands()...)
	}

	// Adicionar comandos de gravação de sessão
	if b.recordingHandler != nil {
		commands = append(commands, b.recordingHandler.GetSlashCommands()...)
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

// onInteractionCreate despacha Slash Commands e Component Interactions (Menus/Botões)
func (b *Bot) onInteractionCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		cmdName := i.ApplicationCommandData().Name
		if cmdName == "rolar" || cmdName == "roll" || cmdName == "sorte" || cmdName == "dados_config" {
			if b.diceHandler != nil {
				b.diceHandler.HandleSlashCommand(s, i)
			}
			return
		}
		if cmdName == "play" || cmdName == "tocar" || cmdName == "parar" || cmdName == "stop" ||
			cmdName == "pular" || cmdName == "skip" || cmdName == "fila" || cmdName == "queue" ||
			cmdName == "sair" || cmdName == "leave" || cmdName == "ambiente" || cmdName == "musicas" {
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
		if b.dndHandler != nil {
			b.dndHandler.HandleSlashCommand(s, i)
		}
	case discordgo.InteractionApplicationCommandAutocomplete:
		cmdName := i.ApplicationCommandData().Name
		if cmdName == "play" || cmdName == "tocar" {
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
		if b.dndHandler != nil {
			b.dndHandler.HandleComponentInteraction(s, i)
		}
	}
}
