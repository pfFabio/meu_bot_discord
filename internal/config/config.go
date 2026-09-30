// Package config é responsável exclusivamente pelo carregamento
// e validação da configuração da aplicação a partir de variáveis de ambiente.
// Princípio SOLID: Single Responsibility — este pacote cuida apenas de configuração.
package config

import (
	"os"

	"github.com/joho/godotenv"
)

// Config armazena todas as configurações necessárias para o bot.
// Cada campo corresponde a uma variável de ambiente obrigatória.
type Config struct {
	BotToken string
}

// Load carrega as variáveis de ambiente do arquivo .env (se existir)
// e retorna uma instância validada de Config ou erro se não configurada.
func Load() (*Config, error) {
	// Tenta carregar o .env; ignora erro se não existir (ex: em produção usa vars de ambiente direto)
	_ = godotenv.Load()

	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		return nil, os.ErrNotExist
	}

	return &Config{
		BotToken: token,
	}, nil
}
