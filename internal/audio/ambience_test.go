package audio_test

import (
	"testing"

	"kenobi-bot/internal/audio"
)

func TestCurseOfStrahdAmbienceList(t *testing.T) {
	list := audio.GetCurseOfStrahdAmbienceList()
	if len(list) != 18 {
		t.Fatalf("Esperado 18 faixas de ambiente para Curse of Strahd, recebido %d", len(list))
	}

	keys := make(map[string]bool)
	for _, item := range list {
		if item.Key == "" {
			t.Errorf("Faixa %s possui Key vazia", item.Title)
		}
		if keys[item.Key] {
			t.Errorf("Key duplicada: %s", item.Key)
		}
		keys[item.Key] = true

		if item.Title == "" {
			t.Errorf("Faixa com Key %s possui Title vazio", item.Key)
		}
		if item.Filename == "" {
			t.Errorf("Faixa %s possui Filename vazio", item.Title)
		}
		if item.Emoji == "" {
			t.Errorf("Faixa %s possui Emoji vazio", item.Title)
		}
		if item.Description == "" {
			t.Errorf("Faixa %s possui Description vazia", item.Title)
		}
	}
}

func TestGetAmbienceDir(t *testing.T) {
	dir := audio.GetAmbienceDir()
	if dir == "" {
		t.Fatalf("GetAmbienceDir retornou string vazia")
	}
}

func TestGetAvailableAmbienceTracks(t *testing.T) {
	tracks := audio.GetAvailableAmbienceTracks()
	if len(tracks) != 18 {
		t.Errorf("Esperado 18 faixas de ambiente prontas na pasta ambiente/, encontrado %d", len(tracks))
	}
}

