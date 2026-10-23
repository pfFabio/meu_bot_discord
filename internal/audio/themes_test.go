package audio

import (
	"testing"
)

func TestThemesCatalog(t *testing.T) {
	tracks := GetThemeTracks()
	if len(tracks) == 0 {
		t.Fatalf("Esperava faixas de temas no catálogo, obteve 0")
	}

	// Testa busca por tema específico
	track, ok := FindThemeTrack("ireena")
	if !ok || track == nil {
		t.Fatalf("Esperava encontrar tema de Ireena")
	}
	if track.Key != "ireena_kolyana" && !track.IsAvailable {
		// ok
	}

	// Testa categorias
	categories := GetThemeCategories()
	if len(categories) == 0 {
		t.Fatalf("Esperava categorias de temas")
	}

	npcTracks := GetThemeTracksByCategory("Personagens")
	if len(npcTracks) == 0 {
		t.Fatalf("Esperava faixas na categoria Personagens")
	}
}
