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

// Load carrega as variáveis de ambiente do arquivo .env padrão
func Load() (*Config, error) {
	return LoadFrom(".env")
}

// LoadFrom carrega as variáveis de ambiente a partir de caminhos específicos
func LoadFrom(filenames ...string) (*Config, error) {
	_ = godotenv.Load(filenames...)

	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		return nil, os.ErrNotExist
	}

	return &Config{
		BotToken: token,
	}, nil
}
