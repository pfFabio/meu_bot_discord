package audio

import (
	"context"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"
)

// Supervisor gerencia o ciclo de vida do processo Node.js de áudio com DAVE E2EE
type Supervisor struct {
	cmdPath    string
	client     *BridgeClient
	cmd        *exec.Cmd
	ctx        context.Context
	cancel     context.CancelFunc
	mu         sync.Mutex
	isStopping bool
}

// NewSupervisor cria uma nova instância do supervisor de processos
func NewSupervisor(client *BridgeClient) *Supervisor {
	ctx, cancel := context.WithCancel(context.Background())
	return &Supervisor{
		cmdPath: filepath.Join("audio-bridge", "index.js"),
		client:  client,
		ctx:     ctx,
		cancel:  cancel,
	}
}

// Start inicia o processo do microserviço Node.js e dispara o monitoramento de saúde
func (s *Supervisor) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, err := os.Stat(s.cmdPath); os.IsNotExist(err) {
		slog.Warn("Arquivo do audio-bridge não encontrado, pulando inicialização automática", "path", s.cmdPath)
		return nil
	}

	s.startProcess()

	// Iniciar rotina de monitoramento e recuperação automática (Health Watchdog)
	go s.watchdogLoop()

	return nil
}

func (s *Supervisor) startProcess() {
	slog.Info("Iniciando microserviço de áudio DAVE (Node.js Sidecar)...")
	cmd := exec.CommandContext(s.ctx, "node", s.cmdPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()

	if err := cmd.Start(); err != nil {
		slog.Error("Falha ao iniciar processo Node.js do audio-bridge", "error", err)
		return
	}

	s.cmd = cmd
	slog.Info("Audio Bridge iniciado com sucesso!", "pid", cmd.Process.Pid)
}

// watchdogLoop verifica periodicamente a saúde do microserviço e reinicia se necessário
func (s *Supervisor) watchdogLoop() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	// Aguardar inicialização inicial do Node.js e conexão ao Discord
	time.Sleep(8 * time.Second)

	consecutiveFailures := 0

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.mu.Lock()
			stopping := s.isStopping
			s.mu.Unlock()

			if stopping {
				return
			}

			isHealthy := s.client.HealthCheck()
			if !isHealthy {
				consecutiveFailures++
				slog.Warn("Health check do Audio Bridge falhou",
					"consecutive_failures", consecutiveFailures,
				)

				if consecutiveFailures >= 5 {
					slog.Error("Audio Bridge inativo após múltiplas tentativas. Reiniciando processo...")
					s.restartProcess()
					consecutiveFailures = 0
				}
			} else {
				consecutiveFailures = 0
			}
		}
	}
}

func (s *Supervisor) restartProcess() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.cmd != nil && s.cmd.Process != nil {
		_ = s.cmd.Process.Kill()
		_ = s.cmd.Wait()
	}

	s.startProcess()
}

// Stop encerra o processo Node.js de forma graciosa
func (s *Supervisor) Stop() {
	s.mu.Lock()
	s.isStopping = true
	s.cancel()
	s.mu.Unlock()

	if s.cmd != nil && s.cmd.Process != nil {
		slog.Info("Encerrando processo do Audio Bridge...")
		_ = s.cmd.Process.Kill()
		_ = s.cmd.Wait()
	}
}
