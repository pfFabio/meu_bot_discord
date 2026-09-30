package audio_test

import (
	"testing"

	"kenobi-bot/internal/audio"
)

func TestScanLocalMusic(t *testing.T) {
	tracks := audio.ScanLocalMusic()
	if len(tracks) == 0 {
		t.Skip("Nenhuma música encontrada na pasta musicas, pulando teste de varredura real")
	}

	foundMago := false
	foundVistani := false
	for _, tr := range tracks {
		if tr.Filename == "mago e strahd.mp3" {
			foundMago = true
			if tr.Emoji != "🧙‍♂️" {
				t.Errorf("Emoji esperado 🧙‍♂️, recebido: %s", tr.Emoji)
			}
		}
		if tr.Filename == "vistani introdução.mp3" {
			foundVistani = true
			if tr.Emoji != "🎪" {
				t.Errorf("Emoji esperado 🎪, recebido: %s", tr.Emoji)
			}
		}
	}

	if !foundMago {
		t.Errorf("mago e strahd.mp3 não foi encontrado pelo scanner")
	}
	if !foundVistani {
		t.Errorf("vistani introdução.mp3 não foi encontrado pelo scanner")
	}
}

func TestFindLocalTrack(t *testing.T) {
	// Teste por nome parcial
	track, ok := audio.FindLocalTrack("mago")
	if !ok {
		t.Fatalf("Esperava encontrar faixa com 'mago'")
	}
	if track.Filename != "mago e strahd.mp3" {
		t.Errorf("Faixa incorreta: %s", track.Filename)
	}

	// Teste por prefixo local:
	track2, ok2 := audio.FindLocalTrack("local:vistani introdução.mp3")
	if !ok2 {
		t.Fatalf("Esperava encontrar faixa com 'local:vistani introdução.mp3'")
	}
	if track2.Filename != "vistani introdução.mp3" {
		t.Errorf("Faixa incorreta: %s", track2.Filename)
	}

	// Teste sem correspondência
	_, ok3 := audio.FindLocalTrack("musica_inexistente_xyz_123")
	if ok3 {
		t.Errorf("Não deveria encontrar música inexistente")
	}
}

func TestResolverLocalTrack(t *testing.T) {
	resolver := audio.NewResolver()

	track, err := resolver.Resolve("vistani", "Jogador")
	if err != nil {
		t.Fatalf("Erro ao resolver faixa local: %v", err)
	}
	if !track.IsLocal {
		t.Errorf("Esperava IsLocal = true")
	}
	if track.Requester != "Jogador" {
		t.Errorf("Requester incorreto: %s", track.Requester)
	}
}
