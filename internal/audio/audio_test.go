package audio_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"kenobi-bot/internal/audio"
)

func TestAudioPresets(t *testing.T) {
	presets := audio.GetAmbientPresets()
	if len(presets) < 4 {
		t.Fatalf("Esperado pelo menos 4 presets de áudio de D&D, encontrou %d", len(presets))
	}

	t.Run("FindPreset encontra preset por chave (ex: taverna)", func(t *testing.T) {
		p, ok := audio.FindPreset("taverna")
		if !ok || p == nil {
			t.Fatalf("FindPreset('taverna') não encontrou o preset")
		}
		if p.URL == "" {
			t.Errorf("Preset 'taverna' possui URL vazia")
		}
	})

	t.Run("FindPreset inexistente retorna falso", func(t *testing.T) {
		_, ok := audio.FindPreset("inexistente")
		if ok {
			t.Errorf("FindPreset('inexistente') deveria retornar falso")
		}
	})
}

func TestBridgeClient(t *testing.T) {
	// Cria servidor HTTP simulado para testar o BridgeClient
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/health":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"status": "ok",
				"dave":   true,
			})
		case "/play":
			var req map[string]string
			_ = json.NewDecoder(r.Body).Decode(&req)
			_ = json.NewEncoder(w).Encode(audio.PlayResponse{
				Status:   "playing",
				Position: 0,
				Track: audio.PlayTrackInfo{
					Title:     req["query"],
					URL:       "https://stream.example.com",
					Duration:  "3:45",
					Requester: req["requester"],
					IsPreset:  false,
				},
			})
		case "/stop":
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "stopped"})
		case "/skip":
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "skipped"})
		case "/leave":
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "disconnected"})
		case "/queue/guild123":
			_ = json.NewEncoder(w).Encode(audio.QueueResponse{
				IsPlaying: true,
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := audio.NewBridgeClient(server.URL)

	t.Run("HealthCheck com servidor ativo", func(t *testing.T) {
		if !client.HealthCheck() {
			t.Errorf("HealthCheck() deveria retornar true")
		}
	})

	t.Run("Play envia requisição e retorna resposta formatada", func(t *testing.T) {
		resp, err := client.Play("guild123", "voice123", "taverna", "Mestre")
		if err != nil {
			t.Fatalf("Erro ao chamar Play: %v", err)
		}
		if resp.Status != "playing" {
			t.Errorf("resp.Status = %q, esperado 'playing'", resp.Status)
		}
		if resp.Track.Title != "taverna" {
			t.Errorf("resp.Track.Title = %q, esperado 'taverna'", resp.Track.Title)
		}
	})

	t.Run("Stop limpa reprodução", func(t *testing.T) {
		if err := client.Stop("guild123"); err != nil {
			t.Errorf("Stop() retornou erro: %v", err)
		}
	})

	t.Run("Skip pula faixa", func(t *testing.T) {
		skipped, err := client.Skip("guild123")
		if err != nil || !skipped {
			t.Errorf("Skip() retornou (%v, %v), esperado (true, nil)", skipped, err)
		}
	})

	t.Run("Leave desconecta", func(t *testing.T) {
		if err := client.Leave("guild123"); err != nil {
			t.Errorf("Leave() retornou erro: %v", err)
		}
	})

	t.Run("GetQueue retorna status da fila", func(t *testing.T) {
		qResp, err := client.GetQueue("guild123")
		if err != nil {
			t.Fatalf("GetQueue() retornou erro: %v", err)
		}
		if !qResp.IsPlaying {
			t.Errorf("qResp.IsPlaying deveria ser true")
		}
	})
}
