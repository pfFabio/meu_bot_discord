package handler_test

import (
	"testing"

	"kenobi-bot/internal/audio"
	"kenobi-bot/internal/handler"
)

func TestAudioHandler_GetSlashCommands(t *testing.T) {
	client := audio.NewBridgeClient("http://127.0.0.1:2389")
	h := handler.NewAudioHandler(client)

	commands := h.GetSlashCommands()
	if len(commands) == 0 {
		t.Fatalf("Esperava comandos slash registrados")
	}

	var foundTocar, foundPlay bool
	for _, cmd := range commands {
		if cmd.Name == "tocar" {
			foundTocar = true
			if len(cmd.Options) == 0 {
				t.Errorf("Comando /tocar deveria ter a opção busca")
			} else {
				if cmd.Options[0].Required {
					t.Errorf("A opção 'busca' do /tocar deve ter Required=false para permitir listar as músicas")
				}
				if !cmd.Options[0].Autocomplete {
					t.Errorf("A opção 'busca' do /tocar deve ter Autocomplete=true")
				}
			}
		}
		if cmd.Name == "play" {
			foundPlay = true
			if len(cmd.Options) > 0 && cmd.Options[0].Required {
				t.Errorf("A opção 'busca' do /play deve ter Required=false")
			}
		}
	}

	if !foundTocar {
		t.Errorf("Comando /tocar não encontrado nos slash commands")
	}
	if !foundPlay {
		t.Errorf("Comando /play não encontrado nos slash commands")
	}
}
