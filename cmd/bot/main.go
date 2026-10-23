// Ponto de entrada da aplicação.
// Princípio SOLID — Single Responsibility & Dependency Inversion:
// Compõe todas as dependências, instancia o repositório, os handlers e o bot.
package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"kenobi-bot/internal/audio"
	"kenobi-bot/internal/bot"
	"kenobi-bot/internal/character"
	"kenobi-bot/internal/config"
	"kenobi-bot/internal/dashboard"
	"kenobi-bot/internal/dice"
	"kenobi-bot/internal/dnd/repository"
	"kenobi-bot/internal/handler"
	"kenobi-bot/internal/health"
	"kenobi-bot/internal/userpref"
)

func main() {
	// 0. Modo Diagnóstico (--doctor, --check, --diagnose)
	if len(os.Args) > 1 {
		arg := strings.ToLower(os.Args[1])
		if arg == "--check" || arg == "--doctor" || arg == "--diagnose" || arg == "doctor" || arg == "check" {
			report := health.RunDiagnostics(".")
			fmt.Print(report.FormatConsole())
			if report.HasCriticalFailures() {
				os.Exit(1)
			}
			os.Exit(0)
		}
	}

	// 0b. Pre-flight check automático de saúde antes de iniciar
	report := health.RunDiagnostics(".")
	if report.HasCriticalFailures() {
		fmt.Print(report.FormatConsole())
		slog.Error("Falha crítica nos testes de saúde pré-execução. Corrija os problemas acima antes de iniciar.")
		os.Exit(1)
	}
	if report.TotalWarn > 0 {
		slog.Warn("Preflight check concluído com avisos", "avisos", report.TotalWarn)
	} else {
		slog.Info("Preflight check: Todos os subsistemas operacionais e saudáveis!", "passou", report.TotalPass)
	}

	// 1. Carrega configuração
	cfg, err := config.Load()
	if err != nil {
		slog.Error("DISCORD_TOKEN não configurado. Adicione no .env ou variáveis de ambiente.")
		os.Exit(1)
	}

	// 2. Carregar repositório de dados do D&D 2024 / 5.5e
	dndRepo, err := repository.LoadDNDData(".")
	if err != nil {
		slog.Error("Falha ao carregar dados do D&D", "error", err)
		os.Exit(1)
	}

	// 3. Instanciar Handlers e Serviços
	dndHandler := handler.NewDNDHandler(dndRepo)
	helloHandler := handler.NewHelloThereHandler()
	prefService := userpref.NewPersistentService("user_preferences.json")
	diceRoller := dice.NewRoller()
	diceHandler := handler.NewDiceHandler(diceRoller, prefService)
	
	// 3b. Instanciar Serviço e Handler de Fichas de Personagem D&D
	charService := character.NewPersistentService("fichas.json")
	characterHandler := handler.NewCharacterHandler(charService)

	// 4. Instanciar Audio Bridge Client e Supervisor (DAVE E2EE)
	audioClient := audio.NewBridgeClient("http://127.0.0.1:2389")
	audioSupervisor := audio.NewSupervisor(audioClient)
	if err := audioSupervisor.Start(); err != nil {
		slog.Error("Falha ao iniciar Supervisor de Áudio", "error", err)
	}
	defer audioSupervisor.Stop()

	audioHandler := handler.NewAudioHandler(audioClient)

	// 4b. Instanciar Recording Handler (gravação de sessões de RPG)
	recordingHandler := handler.NewRecordingHandler(audioClient)

	// 5. Instanciar Bot com injeção de dependências
	b, err := bot.New(cfg.BotToken, dndHandler, diceHandler, audioHandler, recordingHandler, characterHandler)
	if err != nil {
		slog.Error("Falha ao criar instância do bot", "error", err)
		os.Exit(1)
	}

	// 6. Registrar Message Handlers
	b.AddMessageHandler(helloHandler)
	b.AddMessageHandler(diceHandler)
	b.AddMessageHandler(characterHandler)

	// 7. Iniciar conexão com o Discord
	if err := b.Start(); err != nil {
		slog.Error("Falha ao conectar com o Discord", "error", err)
		os.Exit(1)
	}

	slog.Info("Kenobi D&D Bot está online e pronto! Pressione Ctrl+C para encerrar.")

	// 7b. Iniciar Painel de Controle e Interface Local (GUI)
	dashServer := dashboard.NewServer(dashboard.Config{
		Port: 8080,
		GetSession: func() *discordgo.Session {
			return b.Session()
		},
		AudioClient: audioClient,
		DNDRepo:     dndRepo,
		DiceRoller:  diceRoller,
		PrefService: prefService,
		CharService: charService,
	})

	if err := dashServer.Start(); err != nil {
		slog.Error("Falha ao iniciar Painel de Controle Local", "error", err)
	} else {
		defer dashServer.Stop()
		slog.Info("Painel de Controle Local ativo!", "url", dashServer.GetURL())

		// Abre a janela dedicada independente na tela do usuário automaticamente
		openGUI := true
		for _, arg := range os.Args[1:] {
			if arg == "--no-gui" || arg == "--headless" {
				openGUI = false
				break
			}
		}

		if openGUI {
			go func() {
				// Aguarda o Gateway do Discord carregar dados das guilds antes de abrir a janela
				time.Sleep(1 * time.Second)
				dashboard.LaunchStandaloneWindow(dashServer.GetURL())
			}()
		}
	}

	// 8. Aguarda sinal do sistema operacional (Ctrl+C ou SIGTERM)
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	slog.Info("Recebido sinal de encerramento. Finalizando serviços...")
	dashServer.Stop()
	audioSupervisor.Stop()
	if err := b.Stop(); err != nil {
		slog.Error("Erro ao encerrar bot", "error", err)
	}
	slog.Info("Bot encerrado com sucesso.")
}
