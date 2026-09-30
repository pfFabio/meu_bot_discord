package audio

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// RecordStartResponse resposta do endpoint /record/start
type RecordStartResponse struct {
	Status     string `json:"status"`
	SessionDir string `json:"sessionDir"`
	SessionID  string `json:"sessionId"`
}

// RecordStopResponse resposta do endpoint /record/stop
type RecordStopResponse struct {
	Status            string   `json:"status"`
	SessionDir        string   `json:"sessionDir"`
	DurationMs        int64    `json:"durationMs"`
	Files             []string `json:"files"`
	MergedFile        string   `json:"mergedFile"`
	TranscriptionFile string   `json:"transcriptionFile"`
}

// RecordStatusResponse resposta do endpoint /record/status
type RecordStatusResponse struct {
	IsRecording  bool   `json:"isRecording"`
	SessionDir   string `json:"sessionDir"`
	DurationMs   int64  `json:"durationMs"`
	Participants int    `json:"participants"`
}

// StartRecording inicia a gravação da sessão de RPG
func (c *BridgeClient) StartRecording(guildID, channelID string) (*RecordStartResponse, error) {
	payload := map[string]string{
		"guildId":   guildID,
		"channelId": channelID,
	}
	data, _ := json.Marshal(payload)

	resp, err := c.httpClient.Post(c.baseURL+"/record/start", "application/json", bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("falha ao comunicar com serviço de áudio: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp map[string]string
		_ = json.NewDecoder(resp.Body).Decode(&errResp)
		if msg, ok := errResp["error"]; ok {
			return nil, fmt.Errorf("%s", msg)
		}
		return nil, fmt.Errorf("erro HTTP %d do serviço de gravação", resp.StatusCode)
	}

	var result RecordStartResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

// StopRecording para a gravação da sessão e retorna metadados dos arquivos
func (c *BridgeClient) StopRecording(guildID string) (*RecordStopResponse, error) {
	payload := map[string]string{"guildId": guildID}
	data, _ := json.Marshal(payload)

	resp, err := c.httpClient.Post(c.baseURL+"/record/stop", "application/json", bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("falha ao comunicar com serviço de áudio: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp map[string]string
		_ = json.NewDecoder(resp.Body).Decode(&errResp)
		if msg, ok := errResp["error"]; ok {
			return nil, fmt.Errorf("%s", msg)
		}
		return nil, fmt.Errorf("erro HTTP %d do serviço de gravação", resp.StatusCode)
	}

	var result RecordStopResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetRecordingStatus consulta o status da gravação ativa na guilda
func (c *BridgeClient) GetRecordingStatus(guildID string) (*RecordStatusResponse, error) {
	resp, err := c.httpClient.Get(fmt.Sprintf("%s/record/status/%s", c.baseURL, guildID))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result RecordStatusResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}
