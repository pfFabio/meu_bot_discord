// Ponto de entrada da aplicação.
// Princípio SOLID — Single Responsibility & Dependency Inversion:
// Compõe todas as dependências, instancia o repositório, os handlers e o bot.
package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"kenobi-bot/internal/audio"
	"kenobi-bot/internal/bot"
	"kenobi-bot/internal/config"
	"kenobi-bot/internal/dice"
	"kenobi-bot/internal/dnd/repository"
	"kenobi-bot/internal/handler"
	"kenobi-bot/internal/userpref"
)

func main() {
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
	prefService := userpref.NewMemoryService()
	diceRoller := dice.NewRoller()
	diceHandler := handler.NewDiceHandler(diceRoller, prefService)
	
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
	b, err := bot.New(cfg.BotToken, dndHandler, diceHandler, audioHandler, recordingHandler)
	if err != nil {
		slog.Error("Falha ao criar instância do bot", "error", err)
		os.Exit(1)
	}

	// 6. Registrar Message Handlers
	b.AddMessageHandler(helloHandler)
	b.AddMessageHandler(diceHandler)

	// 7. Iniciar conexão com o Discord
	if err := b.Start(); err != nil {
		slog.Error("Falha ao conectar com o Discord", "error", err)
		os.Exit(1)
	}

	slog.Info("Kenobi D&D Bot está online e pronto! Pressione Ctrl+C para encerrar.")

	// 8. Aguarda sinal do sistema operacional (Ctrl+C ou SIGTERM)
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	slog.Info("Recebido sinal de encerramento. Finalizando serviços...")
	audioSupervisor.Stop()
	if err := b.Stop(); err != nil {
		slog.Error("Erro ao encerrar bot", "error", err)
	}
	slog.Info("Bot encerrado com sucesso.")
}
