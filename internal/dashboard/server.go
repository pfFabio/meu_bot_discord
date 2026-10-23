package dashboard

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"log/slog"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"kenobi-bot/internal/audio"
	"kenobi-bot/internal/character"
	"kenobi-bot/internal/dice"
	"kenobi-bot/internal/dnd/repository"
	"kenobi-bot/internal/userpref"
)

//go:embed web/*
var embeddedWebFS embed.FS

// Config parâmetros para inicialização do servidor de dashboard local
type Config struct {
	Port         int
	GetSession   func() *discordgo.Session
	AudioClient  *audio.BridgeClient
	DNDRepo      repository.DNDRepository
	DiceRoller   dice.DiceRoller
	PrefService  userpref.Service
	CharService  character.Service
}

// LogEntry define uma linha de log estruturada
type LogEntry struct {
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	Message   string `json:"message"`
}

// Server gerencia o servidor HTTP local e a interface gráfica do bot
type Server struct {
	cfg          Config
	httpServer   *http.Server
	listener     net.Listener
	actualPort   int
	startTime    time.Time
	mu           sync.RWMutex
	logHistory   []LogEntry
	sseClients   map[chan string]struct{}
	clientsMu    sync.Mutex
}

var globalServer *Server

// NewServer cria uma nova instância do servidor de dashboard
func NewServer(cfg Config) *Server {
	if cfg.Port <= 0 {
		cfg.Port = 8080
	}

	s := &Server{
		cfg:        cfg,
		startTime:  time.Now(),
		logHistory: make([]LogEntry, 0, 200),
		sseClients: make(map[chan string]struct{}),
	}
	globalServer = s
	return s
}

// Start inicia o servidor HTTP em background
func (s *Server) Start() error {
	mux := http.NewServeMux()

	// 1. Registrar rotas da API REST
	s.registerAPIRoutes(mux)

	// 2. Servir os arquivos estáticos da interface web (HTML, CSS, JS)
	var fileSystem http.FileSystem
	// Se existir pasta local durante desenvolvimento, usa do disco para facilidade de hot-reload
	if info, err := os.Stat("internal/dashboard/web"); err == nil && info.IsDir() {
		fileSystem = http.Dir("internal/dashboard/web")
	} else {
		sub, err := fs.Sub(embeddedWebFS, "web")
		if err != nil {
			return fmt.Errorf("falha ao carregar filesystem embutido: %w", err)
		}
		fileSystem = http.FS(sub)
	}

	fileServer := http.FileServer(fileSystem)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Desabilita cache para a interface local responder imediatamente a qualquer atualização
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		fileServer.ServeHTTP(w, r)
	})

	// Tentar porta padrão e incrementar se ocupada
	port := s.cfg.Port
	var listener net.Listener
	var err error

	for attempts := 0; attempts < 10; attempts++ {
		addr := fmt.Sprintf("127.0.0.1:%d", port+attempts)
		listener, err = net.Listen("tcp", addr)
		if err == nil {
			s.actualPort = port + attempts
			break
		}
	}

	if listener == nil {
		return fmt.Errorf("não foi possível encontrar uma porta livre para o painel: %w", err)
	}

	s.listener = listener
	s.httpServer = &http.Server{
		Handler:      s.corsMiddleware(mux),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
	}

	slog.Info("Painel de Controle Local pronto!", "url", s.GetURL())

	go func() {
		if err := s.httpServer.Serve(listener); err != nil && err != http.ErrServerClosed {
			slog.Error("Erro no servidor HTTP do Dashboard", "error", err)
		}
	}()

	return nil
}

// Stop encerra o servidor de dashboard
func (s *Server) Stop() error {
	if s.httpServer == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return s.httpServer.Shutdown(ctx)
}

// GetURL retorna a URL local do painel
func (s *Server) GetURL() string {
	return fmt.Sprintf("http://127.0.0.1:%d", s.actualPort)
}

// AddLog registra uma entrada no histórico do console e notifica clientes SSE
func (s *Server) AddLog(level, message string) {
	entry := LogEntry{
		Timestamp: time.Now().Format("15:04:05"),
		Level:     level,
		Message:   message,
	}

	s.mu.Lock()
	if len(s.logHistory) >= 200 {
		s.logHistory = s.logHistory[1:]
	}
	s.logHistory = append(s.logHistory, entry)
	s.mu.Unlock()

	s.broadcastSSE(fmt.Sprintf(`{"type":"log","data":{"timestamp":"%s","level":"%s","message":%q}}`,
		entry.Timestamp, entry.Level, entry.Message))
}

// LogGlobal registra log no servidor global se instanciado
func LogGlobal(level, message string) {
	if globalServer != nil {
		globalServer.AddLog(level, message)
	}
}

func (s *Server) broadcastSSE(data string) {
	s.clientsMu.Lock()
	defer s.clientsMu.Unlock()

	for ch := range s.sseClients {
		select {
		case ch <- data:
		default:
		}
	}
}

func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
