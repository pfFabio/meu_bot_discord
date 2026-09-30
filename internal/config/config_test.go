package config_test

import (
	"os"
	"testing"

	"kenobi-bot/internal/config"
)

func TestConfigLoad(t *testing.T) {
	t.Run("Carrega token existente das variáveis de ambiente", func(t *testing.T) {
		expectedToken := "test_token_12345"
		os.Setenv("DISCORD_TOKEN", expectedToken)
		defer os.Unsetenv("DISCORD_TOKEN")

		cfg, err := config.Load()
		if err != nil {
			t.Fatalf("Erro inesperado ao carregar config: %v", err)
		}
		if cfg.BotToken != expectedToken {
			t.Errorf("BotToken = %q, esperado %q", cfg.BotToken, expectedToken)
		}
	})

	t.Run("Retorna erro quando DISCORD_TOKEN não está definido", func(t *testing.T) {
		os.Unsetenv("DISCORD_TOKEN")

		_, err := config.Load()
		if err == nil {
			t.Errorf("Esperado erro ao carregar config sem DISCORD_TOKEN, mas retornou nil")
		}
	})
}
