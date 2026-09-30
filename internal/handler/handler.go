// Package handler define a interface e as implementações dos handlers
// de eventos do Discord.
// Princípio SOLID:
//   - Interface Segregation: MessageHandler é uma interface enxuta com um único método.
//   - Dependency Inversion: o pacote bot depende da abstração (interface),
//     não da implementação concreta.
//   - Open/Closed: novos handlers podem ser adicionados implementando a interface,
//     sem alterar o código existente.
package handler

import "github.com/bwmarrin/discordgo"

// MessageHandler define o contrato para qualquer handler que reage
// a mensagens recebidas no Discord.
type MessageHandler interface {
	// Handle processa um evento de criação de mensagem.
	// Retorna true se a mensagem foi tratada por este handler, false caso contrário.
	Handle(s *discordgo.Session, m *discordgo.MessageCreate) bool
}
