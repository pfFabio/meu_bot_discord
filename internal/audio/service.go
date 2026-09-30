package audio

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/jonas747/dca"
)

// GuildVoiceState gerencia o estado da fila e conexão de voz por servidor
type GuildVoiceState struct {
	GuildID         string
	VoiceConnection *discordgo.VoiceConnection
	Queue           []Track
	CurrentTrack    *Track
	IsPlaying       bool
	StopChan        chan struct{}
	EncodeSession   *dca.EncodeSession
	mu              sync.Mutex
}

// Service gerencia as conexões e reproduções de áudio do bot
type Service struct {
	states   map[string]*GuildVoiceState
	resolver *Resolver
	mu       sync.RWMutex
}

// NewService cria uma nova instância do serviço de áudio
func NewService(resolver *Resolver) *Service {
	return &Service{
		states:   make(map[string]*GuildVoiceState),
		resolver: resolver,
	}
}

func (s *Service) getOrCreateState(guildID string) *GuildVoiceState {
	s.mu.Lock()
	defer s.mu.Unlock()

	state, ok := s.states[guildID]
	if !ok {
		state = &GuildVoiceState{
			GuildID:  guildID,
			Queue:    make([]Track, 0),
			StopChan: make(chan struct{}),
		}
		s.states[guildID] = state
	}
	return state
}

// JoinChannel conecta o bot ao canal de voz especificado
func (s *Service) JoinChannel(discord *discordgo.Session, guildID, channelID string) (*discordgo.VoiceConnection, error) {
	state := s.getOrCreateState(guildID)
	state.mu.Lock()
	defer state.mu.Unlock()

	// Se já estiver conectado ao mesmo canal e pronto, reutiliza
	if state.VoiceConnection != nil && state.VoiceConnection.Ready {
		return state.VoiceConnection, nil
	}

	vc, err := discord.ChannelVoiceJoin(guildID, channelID, false, true)
	if err != nil {
		return nil, fmt.Errorf("falha ao entrar no canal de voz: %w", err)
	}

	// Aguardar conexão estar pronta (máximo 5 segundos)
	timeout := time.After(5 * time.Second)
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			return vc, nil
		case <-ticker.C:
			if vc.Ready {
				state.VoiceConnection = vc
				return vc, nil
			}
		}
	}
}

// Enqueue adiciona uma faixa à fila e inicia a reprodução se o bot estiver ocioso
func (s *Service) Enqueue(discord *discordgo.Session, guildID, channelID string, track Track) (int, error) {
	state := s.getOrCreateState(guildID)

	_, err := s.JoinChannel(discord, guildID, channelID)
	if err != nil {
		return 0, err
	}

	state.mu.Lock()
	state.Queue = append(state.Queue, track)
	queueLen := len(state.Queue)
	isPlaying := state.IsPlaying
	state.mu.Unlock()

	if !isPlaying {
		go s.playLoop(state)
	}

	return queueLen, nil
}

// playLoop executa as faixas da fila continuamente até esgotar
func (s *Service) playLoop(state *GuildVoiceState) {
	for {
		state.mu.Lock()
		if len(state.Queue) == 0 {
			state.IsPlaying = false
			state.CurrentTrack = nil
			if state.VoiceConnection != nil && state.VoiceConnection.Ready {
				_ = state.VoiceConnection.Speaking(false)
			}
			state.mu.Unlock()
			return
		}

		// Pega a próxima música da fila
		track := state.Queue[0]
		state.Queue = state.Queue[1:]
		state.CurrentTrack = &track
		state.IsPlaying = true
		state.StopChan = make(chan struct{})
		vc := state.VoiceConnection
		state.mu.Unlock()

		if vc == nil || !vc.Ready {
			slog.Warn("Conexão de voz não está pronta para reprodução", "guild_id", state.GuildID)
			time.Sleep(1 * time.Second)
			continue
		}

		slog.Info("Iniciando reprodução de áudio",
			"guild_id", state.GuildID,
			"title", track.Title,
			"url", track.URL,
		)

		_ = vc.Speaking(true)

		encodeOpts := *dca.StdEncodeOptions
		encodeOpts.Bitrate = 96
		encodeOpts.Application = dca.AudioApplicationAudio

		encodingSession, err := dca.EncodeFile(track.URL, &encodeOpts)
		if err != nil {
			slog.Error("Falha ao codificar áudio com ffmpeg/dca", "error", err, "title", track.Title)
			_ = vc.Speaking(false)
			continue
		}

		state.mu.Lock()
		state.EncodeSession = encodingSession
		state.mu.Unlock()

		s.streamFrames(vc, encodingSession, state.StopChan)

		encodingSession.Cleanup()
		_ = vc.Speaking(false)

		state.mu.Lock()
		state.EncodeSession = nil
		state.mu.Unlock()
	}
}

// streamFrames envia os frames Opus para o canal de voz do Discord
func (s *Service) streamFrames(vc *discordgo.VoiceConnection, session *dca.EncodeSession, stopChan chan struct{}) {
	for {
		select {
		case <-stopChan:
			return
		default:
			frame, err := session.OpusFrame()
			if err != nil {
				if err != io.EOF && !errors.Is(err, io.ErrUnexpectedEOF) {
					slog.Debug("Fim da reprodução de áudio", "info", err)
				}
				return
			}

			select {
			case vc.OpusSend <- frame:
			case <-stopChan:
				return
			case <-time.After(1 * time.Second):
				return
			}
		}
	}
}

// Stop interrompe a reprodução atual e limpa toda a fila
func (s *Service) Stop(guildID string) bool {
	state := s.getOrCreateState(guildID)
	state.mu.Lock()
	defer state.mu.Unlock()

	state.Queue = make([]Track, 0)
	state.CurrentTrack = nil
	state.IsPlaying = false

	if state.StopChan != nil {
		select {
		case <-state.StopChan:
		default:
			close(state.StopChan)
		}
	}

	if state.VoiceConnection != nil && state.VoiceConnection.Ready {
		_ = state.VoiceConnection.Speaking(false)
	}

	return true
}

// Skip pula a faixa atual para a próxima da fila
func (s *Service) Skip(guildID string) bool {
	state := s.getOrCreateState(guildID)
	state.mu.Lock()
	defer state.mu.Unlock()

	if !state.IsPlaying {
		return false
	}

	if state.StopChan != nil {
		select {
		case <-state.StopChan:
		default:
			close(state.StopChan)
		}
	}
	return true
}

// Leave desconecta o bot do canal de voz e limpa os recursos
func (s *Service) Leave(guildID string) error {
	s.Stop(guildID)

	state := s.getOrCreateState(guildID)
	state.mu.Lock()
	defer state.mu.Unlock()

	if state.VoiceConnection != nil {
		err := state.VoiceConnection.Disconnect()
		state.VoiceConnection = nil
		return err
	}
	return nil
}

// GetStatus retorna a faixa atual e a lista de faixas na fila
func (s *Service) GetStatus(guildID string) (*Track, []Track) {
	state := s.getOrCreateState(guildID)
	state.mu.Lock()
	defer state.mu.Unlock()

	queueCopy := make([]Track, len(state.Queue))
	copy(queueCopy, state.Queue)
	return state.CurrentTrack, queueCopy
}

// GetResolver retorna o resolvedor de áudio
func (s *Service) GetResolver() *Resolver {
	return s.resolver
}
