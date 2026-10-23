package audio

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

// PlayTrackInfo representa os metadados de uma faixa retornada na resposta de reprodução
type PlayTrackInfo struct {
	Title     string `json:"title"`
	URL       string `json:"url"`
	Duration  string `json:"duration"`
	Requester string `json:"requester"`
	IsPreset  bool   `json:"isPreset"`
	IsLocal   bool   `json:"isLocal"`
	Filename  string `json:"filename"`
}

// PlayResponse resposta da requisição de play
type PlayResponse struct {
	Status   string        `json:"status"` // "playing", "queued", "playlist", "playlist_queued"
	Position int           `json:"position"`
	Count    int           `json:"count"`
	Track    PlayTrackInfo `json:"track"`
}

// QueueResponse resposta da requisição de fila
type QueueResponse struct {
	Current            *PlayTrackInfo  `json:"current"`
	Queue              []PlayTrackInfo `json:"queue"`
	History            []PlayTrackInfo `json:"history"`
	IsPlaying          bool            `json:"isPlaying"`
	IsPaused           bool            `json:"isPaused"`
	LoopMode           string          `json:"loopMode"` // "off", "track", "queue"
	Volume             float64         `json:"volume"`
	PlaybackDurationMs int64           `json:"playbackDurationMs"`
	TrackStartTime     int64           `json:"trackStartTime"`
}

// BridgeClient faz chamadas HTTP para o microserviço Node.js local com DAVE E2EE
type BridgeClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewBridgeClient cria uma nova instância do cliente do sidecar
func NewBridgeClient(baseURL string) *BridgeClient {
	if baseURL == "" {
		baseURL = "http://127.0.0.1:2389"
	}
	return &BridgeClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// HealthCheck verifica se o sidecar está rodando e pronto com DAVE E2EE
func (c *BridgeClient) HealthCheck() bool {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(c.baseURL + "/health")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

// PlayOptions opções para reprodução
type PlayOptions struct {
	PlayNext     bool `json:"playNext,omitempty"`
	Transition   bool `json:"transition,omitempty"`
	FadeDuration int  `json:"fadeDuration,omitempty"` // em ms (padrão 1500)
}

// SkipOptions opções para avançar de música
type SkipOptions struct {
	Transition   bool `json:"transition,omitempty"`
	FadeDuration int  `json:"fadeDuration,omitempty"` // em ms (padrão 1500)
}

// Play envia o comando para tocar áudio ou adicionar à fila
func (c *BridgeClient) Play(guildID, channelID, query, requester string) (*PlayResponse, error) {
	return c.PlayWithOptions(guildID, channelID, query, requester, PlayOptions{})
}

// PlayWithOptions envia o comando para tocar áudio com suporte a PlayNext e transição suave
func (c *BridgeClient) PlayWithOptions(guildID, channelID, query, requester string, opts PlayOptions) (*PlayResponse, error) {
	payload := map[string]interface{}{
		"guildId":      guildID,
		"channelId":    channelID,
		"query":        query,
		"requester":    requester,
		"playNext":     opts.PlayNext,
		"transition":   opts.Transition,
		"fadeDuration": opts.FadeDuration,
	}
	data, _ := json.Marshal(payload)

	resp, err := c.httpClient.Post(c.baseURL+"/play", "application/json", bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("falha ao comunicar com serviço de áudio: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp map[string]string
		_ = json.NewDecoder(resp.Body).Decode(&errResp)
		if msg, ok := errResp["error"]; ok {
			return nil, errors.New(msg)
		}
		return nil, fmt.Errorf("erro HTTP %d do serviço de áudio", resp.StatusCode)
	}

	var playResp PlayResponse
	if err := json.NewDecoder(resp.Body).Decode(&playResp); err != nil {
		return nil, err
	}
	return &playResp, nil
}

// Stop para a reprodução e limpa a fila
func (c *BridgeClient) Stop(guildID string) error {
	payload := map[string]string{"guildId": guildID}
	data, _ := json.Marshal(payload)

	resp, err := c.httpClient.Post(c.baseURL+"/stop", "application/json", bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

// Skip pula a música atual
func (c *BridgeClient) Skip(guildID string) (bool, error) {
	return c.SkipWithOptions(guildID, SkipOptions{})
}

// SkipWithOptions pula a música atual com opções de transição suave
func (c *BridgeClient) SkipWithOptions(guildID string, opts SkipOptions) (bool, error) {
	payload := map[string]interface{}{
		"guildId":      guildID,
		"transition":   opts.Transition,
		"fadeDuration": opts.FadeDuration,
	}
	data, _ := json.Marshal(payload)

	resp, err := c.httpClient.Post(c.baseURL+"/skip", "application/json", bytes.NewBuffer(data))
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	_ = json.NewDecoder(resp.Body).Decode(&result)
	return result["status"] == "skipped", nil
}

// Join conecta o bot a um canal de voz
func (c *BridgeClient) Join(guildID, channelID string) error {
	payload := map[string]string{
		"guildId":   guildID,
		"channelId": channelID,
	}
	data, _ := json.Marshal(payload)

	resp, err := c.httpClient.Post(c.baseURL+"/join", "application/json", bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

// Leave desconecta do canal de voz
func (c *BridgeClient) Leave(guildID string) error {
	payload := map[string]string{"guildId": guildID}
	data, _ := json.Marshal(payload)

	resp, err := c.httpClient.Post(c.baseURL+"/leave", "application/json", bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

// GetQueue consulta o estado e fila da guilda
func (c *BridgeClient) GetQueue(guildID string) (*QueueResponse, error) {
	resp, err := c.httpClient.Get(fmt.Sprintf("%s/queue/%s", c.baseURL, guildID))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var qResp QueueResponse
	if err := json.NewDecoder(resp.Body).Decode(&qResp); err != nil {
		return nil, err
	}
	return &qResp, nil
}

// GetLocalTracks busca a lista de faixas locais do microserviço Node.js
func (c *BridgeClient) GetLocalTracks() ([]LocalTrack, error) {
	resp, err := c.httpClient.Get(c.baseURL + "/local-tracks")
	if err != nil {
		// Fallback para varredura direta no disco se o sidecar estiver iniciando
		return ScanLocalMusic(), nil
	}
	defer resp.Body.Close()

	var result struct {
		Tracks []LocalTrack `json:"tracks"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return ScanLocalMusic(), nil
	}
	return result.Tracks, nil
}

// Loop altera ou alterna o modo de repetição ("off", "track", "queue")
func (c *BridgeClient) Loop(guildID, mode string) (string, error) {
	payload := map[string]string{
		"guildId": guildID,
		"mode":    mode,
	}
	data, _ := json.Marshal(payload)

	resp, err := c.httpClient.Post(c.baseURL+"/loop", "application/json", bytes.NewBuffer(data))
	if err != nil {
		return "", fmt.Errorf("falha ao comunicar com serviço de áudio: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp map[string]string
		_ = json.NewDecoder(resp.Body).Decode(&errResp)
		if msg, ok := errResp["error"]; ok {
			return "", errors.New(msg)
		}
		return "", fmt.Errorf("erro HTTP %d do serviço de áudio", resp.StatusCode)
	}

	var result struct {
		LoopMode string `json:"loopMode"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	return result.LoopMode, nil
}

// Pause pausa a reprodução da música atual
func (c *BridgeClient) Pause(guildID string) error {
	payload := map[string]string{"guildId": guildID}
	data, _ := json.Marshal(payload)

	resp, err := c.httpClient.Post(c.baseURL+"/pause", "application/json", bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

// Resume retoma a reprodução pausada
func (c *BridgeClient) Resume(guildID string) error {
	payload := map[string]string{"guildId": guildID}
	data, _ := json.Marshal(payload)

	resp, err := c.httpClient.Post(c.baseURL+"/resume", "application/json", bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

// Previous volta para a música anterior ou reinicia a atual
func (c *BridgeClient) Previous(guildID string) (*PlayTrackInfo, error) {
	payload := map[string]string{"guildId": guildID}
	data, _ := json.Marshal(payload)

	resp, err := c.httpClient.Post(c.baseURL+"/previous", "application/json", bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Status string         `json:"status"`
		Track  *PlayTrackInfo `json:"track"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result.Track, nil
}

// SetVolume ajusta o volume de reprodução (0.0 a 3.0)
func (c *BridgeClient) SetVolume(guildID string, volume float64) error {
	payload := map[string]interface{}{
		"guildId": guildID,
		"volume":  volume,
	}
	data, _ := json.Marshal(payload)

	resp, err := c.httpClient.Post(c.baseURL+"/volume", "application/json", bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

// RemoveQueueIndex remove uma faixa da fila pelo seu índice
func (c *BridgeClient) RemoveQueueIndex(guildID string, index int) error {
	payload := map[string]interface{}{
		"guildId": guildID,
		"index":   index,
	}
	data, _ := json.Marshal(payload)

	resp, err := c.httpClient.Post(c.baseURL+"/queue/remove", "application/json", bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

// ClearQueue limpa a fila de reprodução
func (c *BridgeClient) ClearQueue(guildID string) error {
	payload := map[string]string{"guildId": guildID}
	data, _ := json.Marshal(payload)

	resp, err := c.httpClient.Post(c.baseURL+"/queue/clear", "application/json", bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

// PromoteQueue move uma faixa da fila para a posição 0 (tocar a seguir)
func (c *BridgeClient) PromoteQueue(guildID string, index int) error {
	payload := map[string]interface{}{
		"guildId": guildID,
		"index":   index,
	}
	data, _ := json.Marshal(payload)

	resp, err := c.httpClient.Post(c.baseURL+"/queue/promote", "application/json", bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

// JumpQueue pula imediatamente para uma música específica da fila, com opção de transição suave
func (c *BridgeClient) JumpQueue(guildID string, index int, transition bool, fadeDuration int) error {
	payload := map[string]interface{}{
		"guildId":      guildID,
		"index":        index,
		"transition":   transition,
		"fadeDuration": fadeDuration,
	}
	data, _ := json.Marshal(payload)

	resp, err := c.httpClient.Post(c.baseURL+"/queue/jump", "application/json", bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}



