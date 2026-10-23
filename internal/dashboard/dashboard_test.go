package dashboard

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bwmarrin/discordgo"
	"kenobi-bot/internal/audio"
	"kenobi-bot/internal/dice"
	"kenobi-bot/internal/dnd/repository"
	"kenobi-bot/internal/userpref"
)

func createTestServer() *Server {
	repo := repository.NewMemoryRepository(nil, nil, nil, nil, nil, nil)
	pref := userpref.NewPersistentService("test_pref.json")
	roller := dice.NewRoller()
	audioClient := audio.NewBridgeClient("http://127.0.0.1:2389")

	cfg := Config{
		Port: 8089,
		GetSession: func() *discordgo.Session {
			return nil
		},
		AudioClient: audioClient,
		DNDRepo:     repo,
		DiceRoller:  roller,
		PrefService: pref,
	}

	return NewServer(cfg)
}

func TestDashboardEndpoints(t *testing.T) {
	s := createTestServer()
	mux := http.NewServeMux()
	s.registerAPIRoutes(mux)

	t.Run("GET /api/status retorna status válido", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/status", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("Esperava código 200, obteve %d", w.Code)
		}

		var resp map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("Erro ao decodificar JSON: %v", err)
		}

		if _, ok := resp["online"]; !ok {
			t.Errorf("Campo 'online' ausente na resposta de status")
		}
	})

	t.Run("GET /api/audio/library retorna biblioteca", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/audio/library", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("Esperava código 200, obteve %d", w.Code)
		}

		var resp map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("Erro ao decodificar JSON: %v", err)
		}

		if _, ok := resp["presets"]; !ok {
			t.Errorf("Campo 'presets' ausente na biblioteca")
		}
	})

	t.Run("POST /api/dice/roll processa rolagem comum", func(t *testing.T) {
		payload := map[string]interface{}{
			"notation": "1d20+3",
		}
		data, _ := json.Marshal(payload)
		req := httptest.NewRequest(http.MethodPost, "/api/dice/roll", bytes.NewBuffer(data))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("Esperava código 200, obteve %d: %s", w.Code, w.Body.String())
		}

		var resp struct {
			Type  string `json:"type"`
			Total int    `json:"total"`
			Rolls []int  `json:"rolls"`
		}
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("Erro ao decodificar JSON: %v", err)
		}

		if resp.Type != "standard" {
			t.Errorf("Esperava type 'standard', obteve %s", resp.Type)
		}
		if len(resp.Rolls) != 1 {
			t.Errorf("Esperava 1 dado rolado, obteve %d", len(resp.Rolls))
		}
		if resp.Total < 4 || resp.Total > 23 {
			t.Errorf("Total de 1d20+3 fora do intervalo: %d", resp.Total)
		}
	})

	t.Run("POST /api/dice/roll processa rolagem de sucessos", func(t *testing.T) {
		payload := map[string]interface{}{
			"notation":      "4d6",
			"threshold":     4,
			"isSuccessMode": true,
		}
		data, _ := json.Marshal(payload)
		req := httptest.NewRequest(http.MethodPost, "/api/dice/roll", bytes.NewBuffer(data))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("Esperava código 200, obteve %d: %s", w.Code, w.Body.String())
		}

		var resp struct {
			Type      string `json:"type"`
			Successes int    `json:"successes"`
			Failures  int    `json:"failures"`
			Rolls     []int  `json:"rolls"`
		}
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("Erro ao decodificar JSON: %v", err)
		}

		if resp.Type != "success" {
			t.Errorf("Esperava type 'success', obteve %s", resp.Type)
		}
		if resp.Successes+resp.Failures != 4 {
			t.Errorf("Soma de sucessos e falhas diferente de 4: %d", resp.Successes+resp.Failures)
		}
	})
}
