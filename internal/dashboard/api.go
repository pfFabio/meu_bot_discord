package dashboard

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"kenobi-bot/internal/audio"
	"kenobi-bot/internal/dice"
	"kenobi-bot/internal/dnd/model"
	"kenobi-bot/internal/health"
	"kenobi-bot/internal/ui"
)

func (s *Server) registerAPIRoutes(mux *http.ServeMux) {
	// Status & Diagnósticos
	mux.HandleFunc("/api/status", s.handleStatus)
	mux.HandleFunc("/api/diagnostics", s.handleDiagnostics)
	mux.HandleFunc("/api/events", s.handleSSE)
	mux.HandleFunc("/api/logs", s.handleLogs)

	// Servidores & Canais Discord
	mux.HandleFunc("/api/guilds", s.handleGuilds)
	mux.HandleFunc("/api/chat/send", s.handleChatSend)

	// Voz
	mux.HandleFunc("/api/voice/connect", s.handleVoiceConnect)
	mux.HandleFunc("/api/voice/disconnect", s.handleVoiceDisconnect)

	// Áudio & Reprodução
	mux.HandleFunc("/api/audio/status", s.handleAudioStatus)
	mux.HandleFunc("/api/audio/play", s.handleAudioPlay)
	mux.HandleFunc("/api/audio/pause", s.handleAudioPause)
	mux.HandleFunc("/api/audio/resume", s.handleAudioResume)
	mux.HandleFunc("/api/audio/skip", s.handleAudioSkip)
	mux.HandleFunc("/api/audio/previous", s.handleAudioPrevious)
	mux.HandleFunc("/api/audio/stop", s.handleAudioStop)
	mux.HandleFunc("/api/audio/volume", s.handleAudioVolume)
	mux.HandleFunc("/api/audio/loop", s.handleAudioLoop)
	mux.HandleFunc("/api/audio/queue/remove", s.handleAudioQueueRemove)
	mux.HandleFunc("/api/audio/queue/clear", s.handleAudioQueueClear)
	mux.HandleFunc("/api/audio/queue/promote", s.handleAudioQueuePromote)
	mux.HandleFunc("/api/audio/queue/jump", s.handleAudioQueueJump)
	mux.HandleFunc("/api/audio/library", s.handleAudioLibrary)

	// Gravação de Sessões
	mux.HandleFunc("/api/record/start", s.handleRecordStart)
	mux.HandleFunc("/api/record/stop", s.handleRecordStop)
	mux.HandleFunc("/api/record/status", s.handleRecordStatus)
	mux.HandleFunc("/api/record/files", s.handleRecordFiles)
	mux.HandleFunc("/api/record/download/", s.handleRecordDownload)

	// Rolagens de Dados
	mux.HandleFunc("/api/dice/roll", s.handleDiceRoll)

	// Compêndio D&D 2024
	mux.HandleFunc("/api/dnd/classes", s.handleDNDClasses)
	mux.HandleFunc("/api/dnd/spells", s.handleDNDSpells)
	mux.HandleFunc("/api/dnd/monsters", s.handleDNDMonsters)
	mux.HandleFunc("/api/dnd/items", s.handleDNDItems)
	mux.HandleFunc("/api/dnd/rules", s.handleDNDRules)
	mux.HandleFunc("/api/dnd/species", s.handleDNDSpecies)
	mux.HandleFunc("/api/dnd/send", s.handleDNDSend)

	// Fichas de Personagem D&D
	mux.HandleFunc("/api/characters", s.handleCharacters)
	mux.HandleFunc("/api/character/send", s.handleCharacterSend)
	mux.HandleFunc("/api/character/hp", s.handleCharacterHP)
	mux.HandleFunc("/api/character/delete", s.handleCharacterDelete)
	mux.HandleFunc("/api/character/update", s.handleCharacterUpdate)
}

func (s *Server) jsonResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(data)
}

func (s *Server) errorResponse(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

// ═══════════════════════════════════════════════════════════════
// STATUS & EVENTOS EM TEMPO REAL (SSE)
// ═══════════════════════════════════════════════════════════════

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	sess := s.cfg.GetSession()
	isOnline := sess != nil && sess.State != nil && sess.State.User != nil

	var username, avatarURL string
	var guildCount int
	var pingMs int64

	if isOnline {
		username = sess.State.User.Username + "#" + sess.State.User.Discriminator
		avatarURL = sess.State.User.AvatarURL("128")
		guildCount = len(sess.State.Guilds)
		pingMs = sess.HeartbeatLatency().Milliseconds()
	}

	bridgeHealthy := false
	if s.cfg.AudioClient != nil {
		bridgeHealthy = s.cfg.AudioClient.HealthCheck()
	}

	s.jsonResponse(w, map[string]interface{}{
		"online":        isOnline,
		"username":      username,
		"avatarUrl":     avatarURL,
		"guildCount":    guildCount,
		"pingMs":        pingMs,
		"uptimeSeconds": int(time.Since(s.startTime).Seconds()),
		"bridgeHealthy": bridgeHealthy,
		"appUrl":        s.GetURL(),
	})
}

func (s *Server) handleDiagnostics(w http.ResponseWriter, r *http.Request) {
	report := health.RunDiagnostics(".")
	s.jsonResponse(w, report)
}

func (s *Server) handleLogs(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	s.jsonResponse(w, s.logHistory)
}

func (s *Server) handleSSE(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	clientChan := make(chan string, 10)
	s.clientsMu.Lock()
	s.sseClients[clientChan] = struct{}{}
	s.clientsMu.Unlock()

	defer func() {
		s.clientsMu.Lock()
		delete(s.sseClients, clientChan)
		close(clientChan)
		s.clientsMu.Unlock()
	}()

	// Enviar ping inicial
	fmt.Fprintf(w, "data: {\"type\":\"connected\"}\n\n")
	flusher.Flush()

	notify := r.Context().Done()
	for {
		select {
		case <-notify:
			return
		case msg := <-clientChan:
			fmt.Fprintf(w, "data: %s\n\n", msg)
			flusher.Flush()
		}
	}
}

// ═══════════════════════════════════════════════════════════════
// SERVIDORES & CANAIS DISCORD
// ═══════════════════════════════════════════════════════════════

type ChannelInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type int    `json:"type"`
}

type GuildDetails struct {
	ID                  string        `json:"id"`
	Name                string        `json:"name"`
	IconURL             string        `json:"iconUrl"`
	TextChannels        []ChannelInfo `json:"textChannels"`
	VoiceChannels       []ChannelInfo `json:"voiceChannels"`
	CurrentVoiceChannel string        `json:"currentVoiceChannel"`
}

func (s *Server) handleGuilds(w http.ResponseWriter, r *http.Request) {
	sess := s.cfg.GetSession()
	if sess == nil || sess.State == nil {
		s.jsonResponse(w, []GuildDetails{})
		return
	}

	var results []GuildDetails
	botID := ""
	if sess.State.User != nil {
		botID = sess.State.User.ID
	}

	for _, g := range sess.State.Guilds {
		// Checa canais
		channels, err := sess.GuildChannels(g.ID)
		if err != nil {
			channels = g.Channels
		}

		var textChans []ChannelInfo
		var voiceChans []ChannelInfo

		for _, ch := range channels {
			switch ch.Type {
			case discordgo.ChannelTypeGuildText:
				textChans = append(textChans, ChannelInfo{
					ID:   ch.ID,
					Name: ch.Name,
					Type: int(ch.Type),
				})
			case discordgo.ChannelTypeGuildVoice, discordgo.ChannelTypeGuildStageVoice:
				voiceChans = append(voiceChans, ChannelInfo{
					ID:   ch.ID,
					Name: ch.Name,
					Type: int(ch.Type),
				})
			}
		}

		currentVoice := ""
		for _, vs := range g.VoiceStates {
			if vs.UserID == botID {
				currentVoice = vs.ChannelID
				break
			}
		}

		results = append(results, GuildDetails{
			ID:                  g.ID,
			Name:                g.Name,
			IconURL:             g.IconURL("128"),
			TextChannels:        textChans,
			VoiceChannels:       voiceChans,
			CurrentVoiceChannel: currentVoice,
		})
	}

	s.jsonResponse(w, results)
}

func (s *Server) handleChatSend(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.errorResponse(w, http.StatusMethodNotAllowed, "Método não permitido")
		return
	}
	var req struct {
		ChannelID string `json:"channelId"`
		Message   string `json:"message"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ChannelID == "" || req.Message == "" {
		s.errorResponse(w, http.StatusBadRequest, "channelId e message são obrigatórios")
		return
	}

	sess := s.cfg.GetSession()
	if sess == nil {
		s.errorResponse(w, http.StatusServiceUnavailable, "Discord não conectado")
		return
	}

	msg, err := sess.ChannelMessageSend(req.ChannelID, req.Message)
	if err != nil {
		s.errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.AddLog("INFO", fmt.Sprintf("Mensagem enviada no canal %s pelo Mestre", req.ChannelID))
	s.jsonResponse(w, map[string]interface{}{"status": "sent", "id": msg.ID})
}

// ═══════════════════════════════════════════════════════════════
// CONTROLES DE VOZ & CONEXÃO
// ═══════════════════════════════════════════════════════════════

func (s *Server) handleVoiceConnect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.errorResponse(w, http.StatusMethodNotAllowed, "Método não permitido")
		return
	}
	var req struct {
		GuildID   string `json:"guildId"`
		ChannelID string `json:"channelId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.GuildID == "" || req.ChannelID == "" {
		s.errorResponse(w, http.StatusBadRequest, "guildId e channelId são obrigatórios")
		return
	}

	if s.cfg.AudioClient == nil {
		s.errorResponse(w, http.StatusServiceUnavailable, "Audio Bridge indisponível")
		return
	}

	if err := s.cfg.AudioClient.Join(req.GuildID, req.ChannelID); err != nil {
		s.errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.AddLog("INFO", fmt.Sprintf("Bot conectado ao canal de voz %s (guild %s)", req.ChannelID, req.GuildID))
	s.jsonResponse(w, map[string]string{"status": "connected"})
}

func (s *Server) handleVoiceDisconnect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.errorResponse(w, http.StatusMethodNotAllowed, "Método não permitido")
		return
	}
	var req struct {
		GuildID string `json:"guildId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.GuildID == "" {
		s.errorResponse(w, http.StatusBadRequest, "guildId é obrigatório")
		return
	}

	if s.cfg.AudioClient == nil {
		s.errorResponse(w, http.StatusServiceUnavailable, "Audio Bridge indisponível")
		return
	}

	if err := s.cfg.AudioClient.Leave(req.GuildID); err != nil {
		s.errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.AddLog("INFO", fmt.Sprintf("Bot desconectado do canal de voz na guild %s", req.GuildID))
	s.jsonResponse(w, map[string]string{"status": "disconnected"})
}

// ═══════════════════════════════════════════════════════════════
// CONTROLES DE ÁUDIO & REPRODUÇÃO
// ═══════════════════════════════════════════════════════════════

func (s *Server) handleAudioStatus(w http.ResponseWriter, r *http.Request) {
	guildID := r.URL.Query().Get("guildId")
	if guildID == "" {
		s.errorResponse(w, http.StatusBadRequest, "guildId é obrigatório")
		return
	}

	if s.cfg.AudioClient == nil {
		s.jsonResponse(w, audio.QueueResponse{})
		return
	}

	qResp, err := s.cfg.AudioClient.GetQueue(guildID)
	if err != nil {
		s.jsonResponse(w, audio.QueueResponse{})
		return
	}

	s.jsonResponse(w, qResp)
}

func (s *Server) handleAudioPlay(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.errorResponse(w, http.StatusMethodNotAllowed, "Método não permitido")
		return
	}
	var req struct {
		GuildID      string `json:"guildId"`
		ChannelID    string `json:"channelId"`
		Query        string `json:"query"`
		Requester    string `json:"requester"`
		PlayNext     bool   `json:"playNext"`
		Transition   bool   `json:"transition"`
		FadeDuration int    `json:"fadeDuration"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.GuildID == "" || req.Query == "" {
		s.errorResponse(w, http.StatusBadRequest, "guildId e query são obrigatórios")
		return
	}

	if req.Requester == "" {
		req.Requester = "Mestre (Painel Local)"
	}

	// Se channelId não foi passado, tenta descobrir o canal onde o bot já está conectado
	if req.ChannelID == "" {
		sess := s.cfg.GetSession()
		if sess != nil && sess.State != nil {
			botID := sess.State.User.ID
			for _, g := range sess.State.Guilds {
				if g.ID == req.GuildID {
					for _, vs := range g.VoiceStates {
						if vs.UserID == botID {
							req.ChannelID = vs.ChannelID
							break
						}
					}
				}
			}
		}
	}

	if req.ChannelID == "" {
		s.errorResponse(w, http.StatusBadRequest, "Selecione um canal de voz para conectar o bot antes de tocar!")
		return
	}

	if s.cfg.AudioClient == nil {
		s.errorResponse(w, http.StatusServiceUnavailable, "Audio Bridge indisponível")
		return
	}

	resp, err := s.cfg.AudioClient.PlayWithOptions(req.GuildID, req.ChannelID, req.Query, req.Requester, audio.PlayOptions{
		PlayNext:     req.PlayNext,
		Transition:   req.Transition,
		FadeDuration: req.FadeDuration,
	})
	if err != nil {
		s.errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	modeStr := ""
	if req.Transition {
		modeStr = " [transição suave]"
	} else if req.PlayNext {
		modeStr = " [a tocar a seguir]"
	}
	s.AddLog("INFO", fmt.Sprintf("Música enviada para reprodução: %s (%s%s)", req.Query, resp.Status, modeStr))
	s.jsonResponse(w, resp)
}

func (s *Server) handleAudioPause(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.errorResponse(w, http.StatusMethodNotAllowed, "Método não permitido")
		return
	}
	var req struct {
		GuildID string `json:"guildId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.GuildID == "" {
		s.errorResponse(w, http.StatusBadRequest, "guildId é obrigatório")
		return
	}

	if err := s.cfg.AudioClient.Pause(req.GuildID); err != nil {
		s.errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.AddLog("INFO", "Reprodução de música pausada pelo painel")
	s.jsonResponse(w, map[string]string{"status": "paused"})
}

func (s *Server) handleAudioResume(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.errorResponse(w, http.StatusMethodNotAllowed, "Método não permitido")
		return
	}
	var req struct {
		GuildID string `json:"guildId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.GuildID == "" {
		s.errorResponse(w, http.StatusBadRequest, "guildId é obrigatório")
		return
	}

	if err := s.cfg.AudioClient.Resume(req.GuildID); err != nil {
		s.errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.AddLog("INFO", "Reprodução de música retomada pelo painel")
	s.jsonResponse(w, map[string]string{"status": "resumed"})
}

func (s *Server) handleAudioSkip(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.errorResponse(w, http.StatusMethodNotAllowed, "Método não permitido")
		return
	}
	var req struct {
		GuildID      string `json:"guildId"`
		Transition   bool   `json:"transition"`
		FadeDuration int    `json:"fadeDuration"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.GuildID == "" {
		s.errorResponse(w, http.StatusBadRequest, "guildId é obrigatório")
		return
	}

	skipped, err := s.cfg.AudioClient.SkipWithOptions(req.GuildID, audio.SkipOptions{
		Transition:   req.Transition,
		FadeDuration: req.FadeDuration,
	})
	if err != nil {
		s.errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	transStr := ""
	if req.Transition {
		transStr = " com transição suave"
	}
	s.AddLog("INFO", fmt.Sprintf("Música avançada (skip)%s pelo painel: %v", transStr, skipped))
	s.jsonResponse(w, map[string]interface{}{"status": "skipped", "success": skipped})
}

func (s *Server) handleAudioPrevious(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.errorResponse(w, http.StatusMethodNotAllowed, "Método não permitido")
		return
	}
	var req struct {
		GuildID string `json:"guildId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.GuildID == "" {
		s.errorResponse(w, http.StatusBadRequest, "guildId é obrigatório")
		return
	}

	prevTrack, err := s.cfg.AudioClient.Previous(req.GuildID)
	if err != nil {
		s.errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.AddLog("INFO", "Música voltada (previous) pelo painel")
	s.jsonResponse(w, map[string]interface{}{"status": "previous", "track": prevTrack})
}

func (s *Server) handleAudioStop(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.errorResponse(w, http.StatusMethodNotAllowed, "Método não permitido")
		return
	}
	var req struct {
		GuildID string `json:"guildId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.GuildID == "" {
		s.errorResponse(w, http.StatusBadRequest, "guildId é obrigatório")
		return
	}

	if err := s.cfg.AudioClient.Stop(req.GuildID); err != nil {
		s.errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.AddLog("INFO", "Reprodução interrompida e fila limpa pelo painel")
	s.jsonResponse(w, map[string]string{"status": "stopped"})
}

func (s *Server) handleAudioVolume(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.errorResponse(w, http.StatusMethodNotAllowed, "Método não permitido")
		return
	}
	var req struct {
		GuildID string  `json:"guildId"`
		Volume  float64 `json:"volume"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.GuildID == "" {
		s.errorResponse(w, http.StatusBadRequest, "guildId e volume são obrigatórios")
		return
	}

	if err := s.cfg.AudioClient.SetVolume(req.GuildID, req.Volume); err != nil {
		s.errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.jsonResponse(w, map[string]interface{}{"status": "ok", "volume": req.Volume})
}

func (s *Server) handleAudioLoop(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.errorResponse(w, http.StatusMethodNotAllowed, "Método não permitido")
		return
	}
	var req struct {
		GuildID string `json:"guildId"`
		Mode    string `json:"mode"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.GuildID == "" {
		s.errorResponse(w, http.StatusBadRequest, "guildId é obrigatório")
		return
	}

	newMode, err := s.cfg.AudioClient.Loop(req.GuildID, req.Mode)
	if err != nil {
		s.errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.jsonResponse(w, map[string]string{"status": "ok", "loopMode": newMode})
}

func (s *Server) handleAudioQueueRemove(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.errorResponse(w, http.StatusMethodNotAllowed, "Método não permitido")
		return
	}
	var req struct {
		GuildID string `json:"guildId"`
		Index   int    `json:"index"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.GuildID == "" {
		s.errorResponse(w, http.StatusBadRequest, "guildId e index são obrigatórios")
		return
	}

	if err := s.cfg.AudioClient.RemoveQueueIndex(req.GuildID, req.Index); err != nil {
		s.errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.jsonResponse(w, map[string]string{"status": "removed"})
}

func (s *Server) handleAudioQueueClear(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.errorResponse(w, http.StatusMethodNotAllowed, "Método não permitido")
		return
	}
	var req struct {
		GuildID string `json:"guildId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.GuildID == "" {
		s.errorResponse(w, http.StatusBadRequest, "guildId é obrigatório")
		return
	}

	if err := s.cfg.AudioClient.ClearQueue(req.GuildID); err != nil {
		s.errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.jsonResponse(w, map[string]string{"status": "cleared"})
}

func (s *Server) handleAudioQueuePromote(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.errorResponse(w, http.StatusMethodNotAllowed, "Método não permitido")
		return
	}
	var req struct {
		GuildID string `json:"guildId"`
		Index   int    `json:"index"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.GuildID == "" {
		s.errorResponse(w, http.StatusBadRequest, "guildId e index são obrigatórios")
		return
	}

	if err := s.cfg.AudioClient.PromoteQueue(req.GuildID, req.Index); err != nil {
		s.errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.AddLog("INFO", fmt.Sprintf("Música na posição %d promovida para próxima a tocar", req.Index+1))
	s.jsonResponse(w, map[string]string{"status": "promoted"})
}

func (s *Server) handleAudioQueueJump(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.errorResponse(w, http.StatusMethodNotAllowed, "Método não permitido")
		return
	}
	var req struct {
		GuildID      string `json:"guildId"`
		Index        int    `json:"index"`
		Transition   bool   `json:"transition"`
		FadeDuration int    `json:"fadeDuration"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.GuildID == "" {
		s.errorResponse(w, http.StatusBadRequest, "guildId e index são obrigatórios")
		return
	}

	if err := s.cfg.AudioClient.JumpQueue(req.GuildID, req.Index, req.Transition, req.FadeDuration); err != nil {
		s.errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	transStr := ""
	if req.Transition {
		transStr = " (transição suave ativada)"
	}
	s.AddLog("INFO", fmt.Sprintf("Saltando imediatamente para música na fila (índice %d)%s", req.Index+1, transStr))
	s.jsonResponse(w, map[string]string{"status": "jumped"})
}

func (s *Server) handleAudioLibrary(w http.ResponseWriter, r *http.Request) {
	localMusic := audio.ScanLocalMusic()
	ambience := audio.GetCurseOfStrahdAmbienceList()
	themes := audio.GetThemeTracks()

	presets := []map[string]string{
		{"key": "taverna", "title": "🍻 Taverna Medieval", "description": "Conversas, copos brindando, lareira e alaúde ao fundo"},
		{"key": "combate", "title": "⚔️ Combate Épico", "description": "Tambores de guerra, metais orquestrais e adrenalina pura"},
		{"key": "masmorra", "title": "🕯️ Masmorra & Caverna", "description": "Gotas d'água ecoando, correntes e clima sombrio claustrofóbico"},
		{"key": "floresta", "title": "🌿 Floresta Feérica", "description": "Pássaros cantando, brisa suave e magia da natureza"},
	}

	s.jsonResponse(w, map[string]interface{}{
		"local":    localMusic,
		"ambience": ambience,
		"themes":   themes,
		"presets":  presets,
	})
}

// ═══════════════════════════════════════════════════════════════
// GRAVAÇÃO DE SESSÕES
// ═══════════════════════════════════════════════════════════════

func (s *Server) handleRecordStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.errorResponse(w, http.StatusMethodNotAllowed, "Método não permitido")
		return
	}
	var req struct {
		GuildID   string `json:"guildId"`
		ChannelID string `json:"channelId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.GuildID == "" || req.ChannelID == "" {
		s.errorResponse(w, http.StatusBadRequest, "guildId e channelId são obrigatórios")
		return
	}

	res, err := s.cfg.AudioClient.StartRecording(req.GuildID, req.ChannelID)
	if err != nil {
		s.errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.AddLog("INFO", fmt.Sprintf("🔴 Gravação de sessão iniciada no canal %s", req.ChannelID))
	s.jsonResponse(w, res)
}

func (s *Server) handleRecordStop(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.errorResponse(w, http.StatusMethodNotAllowed, "Método não permitido")
		return
	}
	var req struct {
		GuildID string `json:"guildId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.GuildID == "" {
		s.errorResponse(w, http.StatusBadRequest, "guildId é obrigatório")
		return
	}

	res, err := s.cfg.AudioClient.StopRecording(req.GuildID)
	if err != nil {
		s.errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.AddLog("INFO", fmt.Sprintf("⏹️ Gravação de sessão finalizada! Arquivo: %s", res.MergedFile))
	s.jsonResponse(w, res)
}

func (s *Server) handleRecordStatus(w http.ResponseWriter, r *http.Request) {
	guildID := r.URL.Query().Get("guildId")
	if guildID == "" {
		s.errorResponse(w, http.StatusBadRequest, "guildId é obrigatório")
		return
	}

	res, err := s.cfg.AudioClient.GetRecordingStatus(guildID)
	if err != nil {
		s.jsonResponse(w, audio.RecordStatusResponse{IsRecording: false})
		return
	}
	s.jsonResponse(w, res)
}

type RecordingFileInfo struct {
	Name         string `json:"name"`
	Path         string `json:"path"`
	SizeBytes    int64  `json:"sizeBytes"`
	FormattedSize string `json:"formattedSize"`
	ModifiedAt   string `json:"modifiedAt"`
}

func (s *Server) handleRecordFiles(w http.ResponseWriter, r *http.Request) {
	recordingsDir := "gravacoes"
	entries, err := os.ReadDir(recordingsDir)
	if err != nil {
		s.jsonResponse(w, []RecordingFileInfo{})
		return
	}

	var files []RecordingFileInfo
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}
		// Pode ser arquivo direto ou pasta de sessão
		size := info.Size()
		if entry.IsDir() {
			// Calcula tamanho da pasta
			_ = filepath.Walk(filepath.Join(recordingsDir, entry.Name()), func(_ string, f os.FileInfo, _ error) error {
				if f != nil && !f.IsDir() {
					size += f.Size()
				}
				return nil
			})
		}

		formatted := formatBytes(size)
		files = append(files, RecordingFileInfo{
			Name:          entry.Name(),
			Path:          filepath.Join(recordingsDir, entry.Name()),
			SizeBytes:     size,
			FormattedSize: formatted,
			ModifiedAt:    info.ModTime().Format("02/01/2006 15:04"),
		})
	}

	// Ordena mais recentes primeiro
	sort.Slice(files, func(i, j int) bool {
		return files[i].ModifiedAt > files[j].ModifiedAt
	})

	s.jsonResponse(w, files)
}

func (s *Server) handleRecordDownload(w http.ResponseWriter, r *http.Request) {
	subPath := strings.TrimPrefix(r.URL.Path, "/api/record/download/")
	cleanPath := filepath.Clean(filepath.Join("gravacoes", subPath))

	if !strings.HasPrefix(cleanPath, "gravacoes") {
		http.Error(w, "Acesso negado", http.StatusForbidden)
		return
	}

	http.ServeFile(w, r, cleanPath)
}

func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

// ═══════════════════════════════════════════════════════════════
// ROLADOR DE DADOS & RPG
// ═══════════════════════════════════════════════════════════════

func (s *Server) handleDiceRoll(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.errorResponse(w, http.StatusMethodNotAllowed, "Método não permitido")
		return
	}

	var req struct {
		Notation      string `json:"notation"`
		RollName      string `json:"rollName"`
		Threshold     int    `json:"threshold"`
		IsSuccessMode bool   `json:"isSuccessMode"`
		RerollOn1     bool   `json:"rerollOn1"`
		SendToDiscord bool   `json:"sendToDiscord"`
		GuildID       string `json:"guildId"`
		ChannelID     string `json:"channelId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Notation == "" {
		s.errorResponse(w, http.StatusBadRequest, "notation é obrigatória")
		return
	}

	roller := s.cfg.DiceRoller
	if roller == nil {
		s.errorResponse(w, http.StatusServiceUnavailable, "Rolador de dados não configurado")
		return
	}

	opts := dice.RollOptions{RerollOn1: req.RerollOn1}

	if req.IsSuccessMode {
		threshold := req.Threshold
		if threshold <= 0 {
			threshold = 5 // padrão D&D/WoD
		}

		res, err := roller.RollSuccess(req.Notation, threshold, opts)
		if err != nil {
			s.errorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		s.AddLog("INFO", fmt.Sprintf("🎲 Rolagem de Sucessos: %s (Alvo ≥ %d) ➔ %d Sucessos, %d Falhas",
			res.Notation, res.Threshold, res.Successes, res.Failures))

		// Enviar para o Discord se solicitado
		if req.SendToDiscord && req.ChannelID != "" {
			sess := s.cfg.GetSession()
			if sess != nil {
				title := "🎯 Rolagem de Sucessos & Falhas"
				if req.RollName != "" {
					title = fmt.Sprintf("🎯 %s: Sucessos & Falhas", req.RollName)
				}
				color := 0x2A9D8F
				if res.Successes == 0 {
					color = 0xE01A22
				}

				embed := &discordgo.MessageEmbed{
					Title:       title,
					Description: fmt.Sprintf("**Notação:** `%s` • **Meta:** ≥ %d\n**Resultado:** **%d Sucesso(s)** e **%d Falha(s)**", res.Notation, res.Threshold, res.Successes, res.Failures),
					Color:       color,
					Fields: []*discordgo.MessageEmbedField{
						{
							Name:   "🎲 Dados Rolados",
							Value:  fmt.Sprintf("%v", res.Rolls),
							Inline: false,
						},
					},
					Footer: &discordgo.MessageEmbedFooter{
						Text: "Rolado pelo Mestre via Painel de Controle Kenobi",
					},
					Timestamp: time.Now().Format(time.RFC3339),
				}
				_, _ = sess.ChannelMessageSendEmbed(req.ChannelID, embed)
			}
		}

		s.jsonResponse(w, map[string]interface{}{
			"type":      "success",
			"result":    res,
			"notation":  res.Notation,
			"successes": res.Successes,
			"failures":  res.Failures,
			"rolls":     res.Rolls,
		})
		return
	}

	// Rolagem comum
	res, err := roller.RollWithOptions(req.Notation, opts)
	if err != nil {
		s.errorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	s.AddLog("INFO", fmt.Sprintf("🎲 Rolagem: %s ➔ Total: %d (Dados: %v)", res.Notation, res.Total, res.Rolls))

	// Enviar para o Discord se solicitado
	if req.SendToDiscord && req.ChannelID != "" {
		sess := s.cfg.GetSession()
		if sess != nil {
			title := "🎲 Rolagem de Dados"
			if req.RollName != "" {
				title = fmt.Sprintf("🎲 %s", req.RollName)
			}

			color := 0x3A86FF
			if res.Sides == 20 && len(res.Rolls) == 1 {
				switch res.Rolls[0] {
				case 20:
					color = 0xD4AF37 // Crítico Dourado
				case 1:
					color = 0xE01A22 // Falha Crítica Vermelha
				}
			}

			embed := &discordgo.MessageEmbed{
				Title:       title,
				Description: fmt.Sprintf("**Notação:** `%s`\n# 🎲 Total: **%d**", res.Notation, res.Total),
				Color:       color,
				Fields: []*discordgo.MessageEmbedField{
					{
						Name:   "Dados Individuais",
						Value:  fmt.Sprintf("%v", res.Rolls),
						Inline: true,
					},
				},
				Footer: &discordgo.MessageEmbedFooter{
					Text: "Rolado pelo Mestre via Painel de Controle Kenobi",
				},
				Timestamp: time.Now().Format(time.RFC3339),
			}
			if res.Modifier != 0 {
				embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
					Name:   "Modificador",
					Value:  fmt.Sprintf("%+d", res.Modifier),
					Inline: true,
				})
			}
			_, _ = sess.ChannelMessageSendEmbed(req.ChannelID, embed)
		}
	}

	s.jsonResponse(w, map[string]interface{}{
		"type":     "standard",
		"result":   res,
		"notation": res.Notation,
		"total":    res.Total,
		"rolls":    res.Rolls,
		"details":  res.Details,
		"modifier": res.Modifier,
	})
}

// ═══════════════════════════════════════════════════════════════
// COMPÊNDIO D&D 2024
// ═══════════════════════════════════════════════════════════════

func (s *Server) handleDNDClasses(w http.ResponseWriter, r *http.Request) {
	if s.cfg.DNDRepo == nil {
		s.jsonResponse(w, []model.Class{})
		return
	}
	classes := s.cfg.DNDRepo.GetClasses()
	s.jsonResponse(w, classes)
}

func (s *Server) handleDNDSpells(w http.ResponseWriter, r *http.Request) {
	if s.cfg.DNDRepo == nil {
		s.jsonResponse(w, []model.Spell{})
		return
	}
	name := r.URL.Query().Get("q")
	levelStr := r.URL.Query().Get("level")
	school := r.URL.Query().Get("school")
	className := r.URL.Query().Get("class")

	level := -1
	if levelStr != "" {
		if lvl, err := strconv.Atoi(levelStr); err == nil {
			level = lvl
		}
	}

	spells := s.cfg.DNDRepo.SearchSpells(name, level, school, className, 50)
	s.jsonResponse(w, spells)
}

func (s *Server) handleDNDMonsters(w http.ResponseWriter, r *http.Request) {
	if s.cfg.DNDRepo == nil {
		s.jsonResponse(w, []model.Monster{})
		return
	}
	name := r.URL.Query().Get("q")
	cr := r.URL.Query().Get("cr")
	monType := r.URL.Query().Get("type")

	monsters := s.cfg.DNDRepo.SearchMonsters(name, cr, monType, 50)
	s.jsonResponse(w, monsters)
}

func (s *Server) handleDNDItems(w http.ResponseWriter, r *http.Request) {
	if s.cfg.DNDRepo == nil {
		s.jsonResponse(w, []model.Item{})
		return
	}
	name := r.URL.Query().Get("q")
	cat := r.URL.Query().Get("category")
	rarity := r.URL.Query().Get("rarity")
	mastery := r.URL.Query().Get("mastery")

	items := s.cfg.DNDRepo.SearchItems(name, cat, rarity, mastery, 50)
	s.jsonResponse(w, items)
}

func (s *Server) handleDNDRules(w http.ResponseWriter, r *http.Request) {
	if s.cfg.DNDRepo == nil {
		s.jsonResponse(w, []model.Rule{})
		return
	}
	query := r.URL.Query().Get("q")
	rules := s.cfg.DNDRepo.SearchRules(query, 50)
	s.jsonResponse(w, rules)
}

func (s *Server) handleDNDSpecies(w http.ResponseWriter, r *http.Request) {
	if s.cfg.DNDRepo == nil {
		s.jsonResponse(w, []model.Species{})
		return
	}
	cat := r.URL.Query().Get("category")
	species := s.cfg.DNDRepo.GetSpeciesList(cat)
	s.jsonResponse(w, species)
}

func (s *Server) handleDNDSend(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.errorResponse(w, http.StatusMethodNotAllowed, "Método não permitido")
		return
	}
	var req struct {
		Type      string `json:"type"` // "spell", "monster", "class", "item", "rule", "species"
		Name      string `json:"name"`
		ChannelID string `json:"channelId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Type == "" || req.Name == "" || req.ChannelID == "" {
		s.errorResponse(w, http.StatusBadRequest, "type, name e channelId são obrigatórios")
		return
	}

	sess := s.cfg.GetSession()
	if sess == nil {
		s.errorResponse(w, http.StatusServiceUnavailable, "Discord não conectado")
		return
	}

	repo := s.cfg.DNDRepo
	if repo == nil {
		s.errorResponse(w, http.StatusServiceUnavailable, "Compêndio D&D indisponível")
		return
	}

	var embed *discordgo.MessageEmbed

	switch req.Type {
	case "spell":
		spell, ok := repo.GetSpell(req.Name)
		if !ok {
			s.errorResponse(w, http.StatusNotFound, "Magia não encontrada")
			return
		}
		embed = ui.BuildSpellEmbed(spell)
	case "monster":
		monster, ok := repo.GetMonster(req.Name)
		if !ok {
			s.errorResponse(w, http.StatusNotFound, "Monstro não encontrado")
			return
		}
		embed = ui.BuildMonsterEmbed(monster)
	case "class":
		class, ok := repo.GetClass(req.Name)
		if !ok {
			s.errorResponse(w, http.StatusNotFound, "Classe não encontrada")
			return
		}
		embed = ui.BuildClassEmbed(class)
	case "item":
		item, ok := repo.GetItem(req.Name)
		if !ok {
			s.errorResponse(w, http.StatusNotFound, "Item não encontrado")
			return
		}
		embed = ui.BuildItemEmbed(item)
	case "rule":
		rule, ok := repo.GetRule(req.Name)
		if !ok {
			s.errorResponse(w, http.StatusNotFound, "Regra não encontrada")
			return
		}
		embed = ui.BuildRuleEmbed(rule)
	case "species":
		sp, ok := repo.GetSpecies(req.Name)
		if !ok {
			s.errorResponse(w, http.StatusNotFound, "Espécie não encontrada")
			return
		}
		embed = ui.BuildSpeciesEmbed(sp)
	default:
		s.errorResponse(w, http.StatusBadRequest, "Tipo de compêndio desconhecido")
		return
	}

	msg, err := sess.ChannelMessageSendEmbed(req.ChannelID, embed)
	if err != nil {
		s.errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.AddLog("INFO", fmt.Sprintf("Ficha de D&D (%s: %s) postada no canal %s", req.Type, req.Name, req.ChannelID))
	s.jsonResponse(w, map[string]interface{}{"status": "sent", "id": msg.ID})
}

// handleCharacters retorna fichas de RPG gerenciadas pelo bot
func (s *Server) handleCharacters(w http.ResponseWriter, r *http.Request) {
	if s.cfg.CharService == nil {
		s.jsonResponse(w, map[string]interface{}{"characters": []interface{}{}})
		return
	}

	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		userID = r.URL.Query().Get("userId")
	}
	if userID != "" {
		sheets := s.cfg.CharService.ListByUser(userID)
		active, _ := s.cfg.CharService.GetActive(userID)
		s.jsonResponse(w, map[string]interface{}{
			"characters": sheets,
			"active":     active,
		})
		return
	}

	// Sem user_id específico, lista todas
	sheets := s.cfg.CharService.GetAll()
	s.jsonResponse(w, map[string]interface{}{
		"characters": sheets,
		"count":      len(sheets),
	})
}

// handleCharacterSend envia embed formatado da ficha de personagem para o Discord
func (s *Server) handleCharacterSend(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.errorResponse(w, http.StatusMethodNotAllowed, "Método não permitido")
		return
	}
	var req struct {
		CharacterID string `json:"characterId"`
		ChannelID   string `json:"channelId"`
		View        string `json:"view"` // "full", "combat", "skills"
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.CharacterID == "" || req.ChannelID == "" {
		s.errorResponse(w, http.StatusBadRequest, "characterId e channelId são obrigatórios")
		return
	}

	if s.cfg.CharService == nil {
		s.errorResponse(w, http.StatusServiceUnavailable, "Serviço de fichas indisponível")
		return
	}

	sheet, ok := s.cfg.CharService.GetByID(req.CharacterID)
	if !ok {
		s.errorResponse(w, http.StatusNotFound, "Ficha de personagem não encontrada")
		return
	}

	sess := s.cfg.GetSession()
	if sess == nil {
		s.errorResponse(w, http.StatusServiceUnavailable, "Sessão do Discord indisponível")
		return
	}

	var embed *discordgo.MessageEmbed
	tab := "overview"
	switch req.View {
	case "attributes":
		embed = ui.BuildCharacterAttributesEmbed(sheet)
		tab = "attributes"
	case "spells":
		embed = ui.BuildCharacterSpellsEmbed(sheet)
		tab = "spells"
	case "features":
		embed = ui.BuildCharacterFeaturesEmbed(sheet)
		tab = "features"
	case "inventory":
		embed = ui.BuildCharacterInventoryEmbed(sheet)
		tab = "inventory"
	default:
		embed = ui.BuildCharacterOverviewEmbed(sheet)
	}

	components := ui.BuildCharacterActionComponents(sheet, tab)
	msg, err := sess.ChannelMessageSendComplex(req.ChannelID, &discordgo.MessageSend{
		Embeds:     []*discordgo.MessageEmbed{embed},
		Components: components,
	})
	if err != nil {
		s.errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.AddLog("INFO", fmt.Sprintf("Ficha de %s postada no canal %s", sheet.Name, req.ChannelID))
	s.jsonResponse(w, map[string]interface{}{"status": "sent", "id": msg.ID})
}

// handleCharacterHP gerencia alterações de PV da ficha pelo painel
func (s *Server) handleCharacterHP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.errorResponse(w, http.StatusMethodNotAllowed, "Método não permitido")
		return
	}
	var req struct {
		CharacterID string `json:"characterId"`
		CurrentHP   int    `json:"currentHp"`
		TempHP      int    `json:"tempHp"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.CharacterID == "" {
		s.errorResponse(w, http.StatusBadRequest, "characterId é obrigatório")
		return
	}

	if s.cfg.CharService == nil {
		s.errorResponse(w, http.StatusServiceUnavailable, "Serviço de fichas indisponível")
		return
	}

	sheet, ok := s.cfg.CharService.GetByID(req.CharacterID)
	if !ok {
		s.errorResponse(w, http.StatusNotFound, "Ficha não encontrada")
		return
	}

	sheet.CurrentHP = req.CurrentHP
	if sheet.CurrentHP < 0 {
		sheet.CurrentHP = 0
	}
	if sheet.CurrentHP > sheet.MaxHP {
		sheet.CurrentHP = sheet.MaxHP
	}
	if req.TempHP >= 0 {
		sheet.TempHP = req.TempHP
	}

	if err := s.cfg.CharService.Save(sheet); err != nil {
		s.errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.AddLog("INFO", fmt.Sprintf("PVs de %s atualizados para %d/%d (Temp: %d)", sheet.Name, sheet.CurrentHP, sheet.MaxHP, sheet.TempHP))
	s.jsonResponse(w, map[string]interface{}{
		"status":    "updated",
		"character": sheet,
	})
}

// handleCharacterDelete exclui permanentemente uma ficha pelo ID
func (s *Server) handleCharacterDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodDelete {
		s.errorResponse(w, http.StatusMethodNotAllowed, "Método não permitido")
		return
	}
	var req struct {
		CharacterID string `json:"characterId"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	if req.CharacterID == "" {
		req.CharacterID = r.URL.Query().Get("characterId")
	}
	if req.CharacterID == "" {
		s.errorResponse(w, http.StatusBadRequest, "characterId é obrigatório")
		return
	}

	if s.cfg.CharService == nil {
		s.errorResponse(w, http.StatusServiceUnavailable, "Serviço de fichas indisponível")
		return
	}

	charName := "Ficha"
	if sheet, ok := s.cfg.CharService.GetByID(req.CharacterID); ok && sheet != nil {
		charName = sheet.Name
	}

	if err := s.cfg.CharService.Delete(req.CharacterID); err != nil {
		s.errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.AddLog("INFO", fmt.Sprintf("Ficha de %s excluída permanentemente via painel", charName))
	s.jsonResponse(w, map[string]interface{}{
		"status": "deleted",
		"id":     req.CharacterID,
	})
}

// handleCharacterUpdate atualiza os campos de uma ficha de RPG pelo painel
func (s *Server) handleCharacterUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodPut {
		s.errorResponse(w, http.StatusMethodNotAllowed, "Método não permitido")
		return
	}

	var req struct {
		CharacterID string `json:"characterId"`
		Name        string `json:"name"`
		Class       string `json:"class"`
		Subclass    string `json:"subclass"`
		UserID      string `json:"userId"`
		Level       int    `json:"level"`
		Race        string `json:"race"`
		Background  string `json:"background"`
		Alignment   string `json:"alignment"`
		ArmorClass  int    `json:"armorClass"`
		Speed       string `json:"speed"`
		MaxHP       int    `json:"maxHp"`
		CurrentHP   int    `json:"currentHp"`
		TempHP      int    `json:"tempHp"`
		STR         int    `json:"str"`
		DEX         int    `json:"dex"`
		CON         int    `json:"con"`
		INT         int    `json:"int"`
		WIS         int    `json:"wis"`
		CHA         int    `json:"cha"`
		GP          int    `json:"gp"`
		SP          int    `json:"sp"`
		CP          int    `json:"cp"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.CharacterID == "" {
		s.errorResponse(w, http.StatusBadRequest, "characterId e dados válidos são obrigatórios")
		return
	}

	if s.cfg.CharService == nil {
		s.errorResponse(w, http.StatusServiceUnavailable, "Serviço de fichas indisponível")
		return
	}

	sheet, ok := s.cfg.CharService.GetByID(req.CharacterID)
	if !ok || sheet == nil {
		s.errorResponse(w, http.StatusNotFound, "Ficha de personagem não encontrada")
		return
	}

	if req.UserID != "" && req.UserID != sheet.UserID {
		_, _ = s.cfg.CharService.AssignSheet(sheet.ID, req.UserID)
	}

	if strings.TrimSpace(req.Name) != "" {
		sheet.Name = strings.TrimSpace(req.Name)
	}
	if req.Class != "" {
		sheet.Class = req.Class
	}
	sheet.Subclass = req.Subclass
	if req.Level > 0 {
		sheet.Level = req.Level
	}
	sheet.Race = req.Race
	sheet.Background = req.Background
	sheet.Alignment = req.Alignment
	if req.ArmorClass > 0 {
		sheet.ArmorClass = req.ArmorClass
	}
	if req.Speed != "" {
		sheet.Speed = req.Speed
	}
	if req.MaxHP > 0 {
		sheet.MaxHP = req.MaxHP
	}
	if req.CurrentHP >= 0 {
		sheet.CurrentHP = req.CurrentHP
		if sheet.CurrentHP > sheet.MaxHP {
			sheet.CurrentHP = sheet.MaxHP
		}
	}
	if req.TempHP >= 0 {
		sheet.TempHP = req.TempHP
	}

	// Atributos base
	if req.STR > 0 {
		sheet.Attributes.STR.Score = req.STR
	}
	if req.DEX > 0 {
		sheet.Attributes.DEX.Score = req.DEX
	}
	if req.CON > 0 {
		sheet.Attributes.CON.Score = req.CON
	}
	if req.INT > 0 {
		sheet.Attributes.INT.Score = req.INT
	}
	if req.WIS > 0 {
		sheet.Attributes.WIS.Score = req.WIS
	}
	if req.CHA > 0 {
		sheet.Attributes.CHA.Score = req.CHA
	}

	// Moedas
	if req.GP >= 0 {
		sheet.Coins.GP = req.GP
	}
	if req.SP >= 0 {
		sheet.Coins.SP = req.SP
	}
	if req.CP >= 0 {
		sheet.Coins.CP = req.CP
	}

	sheet.CalculateDerived()

	if err := s.cfg.CharService.Save(sheet); err != nil {
		s.errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.AddLog("INFO", fmt.Sprintf("Ficha de %s editada com sucesso no painel", sheet.Name))
	s.jsonResponse(w, map[string]interface{}{
		"status":    "updated",
		"character": sheet,
	})
}

