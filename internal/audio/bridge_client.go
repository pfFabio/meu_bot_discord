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
	Current *struct {
		Title     string `json:"title"`
		URL       string `json:"url"`
		Duration  string `json:"duration"`
		Requester string `json:"requester"`
	} `json:"current"`
	Queue []struct {
		Title     string `json:"title"`
		URL       string `json:"url"`
		Duration  string `json:"duration"`
		Requester string `json:"requester"`
	} `json:"queue"`
	IsPlaying bool   `json:"isPlaying"`
	LoopMode  string `json:"loopMode"` // "off", "track", "queue"
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
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(c.baseURL + "/health")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

// Play envia o comando para tocar áudio ou adicionar à fila
func (c *BridgeClient) Play(guildID, channelID, query, requester string) (*PlayResponse, error) {
	payload := map[string]string{
		"guildId":   guildID,
		"channelId": channelID,
		"query":     query,
		"requester": requester,
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
	payload := map[string]string{"guildId": guildID}
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


