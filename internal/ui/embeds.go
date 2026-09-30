package ui

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"kenobi-bot/internal/dnd/model"
)

const (
	ColorDNDRed   = 0xE01A22
	ColorDNDGold  = 0xD4AF37
	ColorDNDBlue  = 0x3A86FF
	ColorDNDGreen = 0x2A9D8F
	ColorDNDDark  = 0x2B2D42
)

// BuildHubEmbed cria o Embed do Menu Principal do D&D 2024
func BuildHubEmbed() *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title:       "🐉 Compêndio Interativo D&D 5.5 (Edição 2024)",
		Description: "Explore todas as regras oficiais, classes, subclasses, magias, bestiário e itens revisados do **Player's Handbook 2024 / SRD 5.2**.\n\nUse os **botões e menus abaixo** para navegar instantaneamente!",
		Color:       ColorDNDRed,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "🛡️ Classes & Subclasses",
				Value:  "Consulte as 13 classes essenciais, atributos, dados de vida, proficiências e subclasses com suas sinopses completas.",
				Inline: false,
			},
			{
				Name:   "🧬 Espécies & Linhagens (Ravenloft & PHB 2024)",
				Value:  "Explore linhagens sombrias de Ravenloft (Dhampir, Hexblood, Lupin, Reborn) e as espécies essenciais do D&D 2024.",
				Inline: false,
			},
			{
				Name:   "✨ Grimório de Magias",
				Value:  "Pesquise e filtre mais de 340 magias oficiais por Nível (Truques ao 9º), Escola Mágica e Lista de Classes.",
				Inline: true,
			},
			{
				Name:   "🐉 Bestiário & Monstros",
				Value:  "Fichas completas de monstros com CA, PV, Atributos, Ações e Nível de Desafio (CR).",
				Inline: true,
			},
			{
				Name:   "⚔️ Itens & Maestria em Armas",
				Value:  "Armas com novas propriedades de Maestria (Cleave, Graze, Vex, Topple, Nick) e Itens Mágicos.",
				Inline: true,
			},
			{
				Name:   "📖 Glossário de Regras",
				Value:  "Consulte definições oficiais de Condições (Invisível, Agarrado, Cego), Ações e Regras de 2024.",
				Inline: true,
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: "D&D 5.5 / 2024 Core Rules • Selecione uma opção abaixo",
		},
	}
}

// BuildClassEmbed cria o Embed de Sinopse e Detalhes da Classe
func BuildClassEmbed(c *model.Class) *discordgo.MessageEmbed {
	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("%s Classe: %s (%s)", c.Emoji, c.Name, c.NamePT),
		Description: fmt.Sprintf("**Sinopse:**\n%s\n\n**Papel no Grupo:** `%s`", c.Overview, c.Role),
		Color:       c.Color,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "🎯 Atributo Primário",
				Value:  c.Traits.PrimaryAbility,
				Inline: true,
			},
			{
				Name:   "🎲 Dado de Vida",
				Value:  c.Traits.HitDie,
				Inline: true,
			},
			{
				Name:   "🛡️ Salvaguardas (Saves)",
				Value:  c.Traits.SavingThrows,
				Inline: true,
			},
			{
				Name:   "⚔️ Treinamento Marcial",
				Value:  fmt.Sprintf("**Armas:** %s\n**Armaduras:** %s", c.Traits.WeaponProf, c.Traits.ArmorTraining),
				Inline: false,
			},
			{
				Name:   "⚡ Maestria em Armas (2024)",
				Value:  c.Traits.WeaponMastery,
				Inline: false,
			},
			{
				Name:   "🎒 Perícias & Equipamento Inicial",
				Value:  fmt.Sprintf("**Perícias:** %s\n**Equipamento:** %s", c.Traits.SkillProf, c.Traits.StartingEquip),
				Inline: false,
			},
		},
	}

	// Adicionar resumo das principais Habilidades de Classe
	if len(c.Features) > 0 {
		var featList []string
		for _, f := range c.Features {
			featList = append(featList, fmt.Sprintf("• **Nv %d - %s:** %s", f.Level, f.Name, f.Description))
		}
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "✨ Habilidades Chave da Classe",
			Value:  strings.Join(featList, "\n"),
			Inline: false,
		})
	}

	// Subclasses disponíveis
	var subNames []string
	for _, sub := range c.Subclasses {
		subNames = append(subNames, fmt.Sprintf("• **%s**", sub.Name))
	}
	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
		Name:   fmt.Sprintf("🌟 Subclasses Disponíveis (%d) - Desbloqueadas no Nível 3", len(c.Subclasses)),
		Value:  strings.Join(subNames, "\n") + "\n\n*👇 Escolha uma subclasse no menu abaixo para ver sua sinopse e habilidades!*",
		Inline: false,
	})

	embed.Footer = &discordgo.MessageEmbedFooter{
		Text: fmt.Sprintf("D&D 2024 • %s • Use o menu abaixo para explorar subclasses", c.Name),
	}

	return embed
}

// BuildSubclassEmbed cria o Embed de Sinopse e Habilidades da Subclasse
func BuildSubclassEmbed(sub *model.Subclass, c *model.Class) *discordgo.MessageEmbed {
	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("%s Subclasse: %s (%s)", c.Emoji, sub.Name, c.Name),
		Description: fmt.Sprintf("**Sinopse Temática:**\n%s\n\n**Nível de Obtenção:** `Nível %d` (Padrão D&D 2024)", sub.Synopsis, sub.UnlockLevel),
		Color:       c.Color,
		Fields:      []*discordgo.MessageEmbedField{},
	}

	// Listar habilidades da subclasse
	if len(sub.Features) > 0 {
		var feats []string
		for _, f := range sub.Features {
			feats = append(feats, fmt.Sprintf("• **Nível %d - %s:**\n%s", f.Level, f.Name, f.Description))
		}
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "⚔️ Habilidades e Características Concedidas",
			Value:  strings.Join(feats, "\n\n"),
			Inline: false,
		})
	}

	embed.Footer = &discordgo.MessageEmbedFooter{
		Text: fmt.Sprintf("D&D 2024 • Subclasse de %s • Use os botões abaixo para voltar", c.Name),
	}

	return embed
}

// BuildSpellEmbed cria o Embed de exibição de Magia
func BuildSpellEmbed(s *model.Spell) *discordgo.MessageEmbed {
	levelStr := fmt.Sprintf("%dº Nível", s.Level)
	if s.Level == 0 {
		levelStr = "Truque (Cantrip)"
	}

	tags := []string{levelStr, s.School}
	if s.Concentration {
		tags = append(tags, "🎯 Concentração")
	}
	if s.Ritual {
		tags = append(tags, "📜 Ritual")
	}

	classesStr := "Geral"
	if len(s.Classes) > 0 {
		classesStr = strings.Join(s.Classes, ", ")
	}

	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("✨ Magia: %s", s.Name),
		Description: fmt.Sprintf("` %s `\n**Classes:** %s", strings.Join(tags, " • "), classesStr),
		Color:       ColorDNDBlue,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "⏱️ Tempo de Conjuração",
				Value:  s.CastingTime,
				Inline: true,
			},
			{
				Name:   "📏 Alcance / Área",
				Value:  s.Range,
				Inline: true,
			},
			{
				Name:   "⌛ Duração",
				Value:  s.Duration,
				Inline: true,
			},
			{
				Name:   "🔮 Componentes",
				Value:  s.Components,
				Inline: false,
			},
			{
				Name:   "📖 Descrição",
				Value:  s.Description,
				Inline: false,
			},
		},
	}

	if s.HigherLevels != "" {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "🔼 Em Níveis Mais Altos (Upcast)",
			Value:  s.HigherLevels,
			Inline: false,
		})
	}

	embed.Footer = &discordgo.MessageEmbedFooter{
		Text: "D&D 2024 Spellbook • Livro do Jogador",
	}

	return embed
}

// BuildMonsterEmbed cria o Embed da Ficha de Monstro
func BuildMonsterEmbed(m *model.Monster) *discordgo.MessageEmbed {
	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("🐉 Bestiário: %s", m.Name),
		Description: fmt.Sprintf("*%s %s, %s*\n**Nível de Desafio (CR):** `%s` (%s XP)", m.Size, m.Type, m.Alignment, m.ChallengeRating, m.XP),
		Color:       ColorDNDGold,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "🛡️ Classe de Armadura (CA)",
				Value:  m.ArmorClass,
				Inline: true,
			},
			{
				Name:   "❤️ Pontos de Vida (PV)",
				Value:  m.HitPoints,
				Inline: true,
			},
			{
				Name:   "💨 Deslocamento (Speed)",
				Value:  m.Speed,
				Inline: true,
			},
		},
	}

	if m.SavingThrows != "" || m.Skills != "" || m.Senses != "" || m.Languages != "" {
		var statsExtra []string
		if m.SavingThrows != "" {
			statsExtra = append(statsExtra, "**Salvaguardas:** "+m.SavingThrows)
		}
		if m.Skills != "" {
			statsExtra = append(statsExtra, "**Perícias:** "+m.Skills)
		}
		if m.Senses != "" {
			statsExtra = append(statsExtra, "**Sentidos:** "+m.Senses)
		}
		if m.Languages != "" {
			statsExtra = append(statsExtra, "**Idiomas:** "+m.Languages)
		}
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "📊 Perícias & Sentidos",
			Value:  strings.Join(statsExtra, "\n"),
			Inline: false,
		})
	}

	// Ações
	if len(m.Actions) > 0 {
		var acts []string
		for _, a := range m.Actions {
			acts = append(acts, fmt.Sprintf("• **%s:** %s", a.Name, a.Description))
		}
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "⚔️ Ações",
			Value:  strings.Join(acts, "\n"),
			Inline: false,
		})
	}

	// Ações Bônus / Reações
	if len(m.BonusActions) > 0 {
		var bActs []string
		for _, ba := range m.BonusActions {
			bActs = append(bActs, fmt.Sprintf("• **%s:** %s", ba.Name, ba.Description))
		}
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "⚡ Ações Bônus",
			Value:  strings.Join(bActs, "\n"),
			Inline: false,
		})
	}

	embed.Footer = &discordgo.MessageEmbedFooter{
		Text: "D&D 2024 Monster Manual / SRD 5.2",
	}

	return embed
}

// BuildItemEmbed cria o Embed de Item / Equipamento
func BuildItemEmbed(it *model.Item) *discordgo.MessageEmbed {
	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("⚔️ Item: %s", it.Name),
		Description: fmt.Sprintf("**Categoria:** `%s` • **Raridade:** `%s`", it.Category, it.Rarity),
		Color:       ColorDNDGreen,
		Fields:      []*discordgo.MessageEmbedField{},
	}

	if it.Damage != "" {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "💥 Dano & Propriedades",
			Value:  fmt.Sprintf("**Dano:** %s\n**Propriedades:** %s", it.Damage, it.Properties),
			Inline: false,
		})
	}

	if it.Mastery != "" {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "⚡ Maestria em Armas (Weapon Mastery 2024)",
			Value:  it.Mastery,
			Inline: false,
		})
	}

	if it.Cost != "" || it.Weight != "" {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "💰 Custo & Peso",
			Value:  fmt.Sprintf("**Custo:** %s • **Peso:** %s", it.Cost, it.Weight),
			Inline: true,
		})
	}

	if it.Description != "" {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "📜 Propriedades Mágicas / Descrição",
			Value:  it.Description,
			Inline: false,
		})
	}

	embed.Footer = &discordgo.MessageEmbedFooter{
		Text: "D&D 2024 Equipment & Magic Items",
	}

	return embed
}

// BuildRuleEmbed cria o Embed do Glossário de Regras
func BuildRuleEmbed(r *model.Rule) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("📖 Regra D&D 2024: %s", r.Name),
		Description: r.Description,
		Color:       ColorDNDDark,
		Footer: &discordgo.MessageEmbedFooter{
			Text: "D&D 2024 Rules Glossary • Regras Oficiais",
		},
	}
}

// BuildSpeciesEmbed cria o Embed com os detalhes e traços de uma Espécie / Raça
func BuildSpeciesEmbed(sp *model.Species) *discordgo.MessageEmbed {
	color := ColorDNDDark
	emoji := "🧬"
	if sp.Category == "Ravenloft" {
		color = 0x581845 // Roxo gótico de Ravenloft
		emoji = "🦇"
	}

	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("%s Espécie: %s (%s)", emoji, sp.Name, sp.NamePT),
		Description: fmt.Sprintf("**Categoria:** `%s` • **Origem:** `%s`\n\n%s", sp.Category, sp.Source, sp.Overview),
		Color:       color,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "👤 Tipo de Criatura",
				Value:  sp.CreatureType,
				Inline: true,
			},
			{
				Name:   "📏 Tamanho (Size)",
				Value:  sp.Size,
				Inline: true,
			},
			{
				Name:   "🏃 Deslocamento",
				Value:  sp.Speed,
				Inline: true,
			},
			{
				Name:   "👁️ Visão no Escuro",
				Value:  sp.Darkvision,
				Inline: true,
			},
		},
	}

	for _, trait := range sp.Traits {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   fmt.Sprintf("✨ %s", trait.Name),
			Value:  trait.Description,
			Inline: false,
		})
	}

	embed.Footer = &discordgo.MessageEmbedFooter{
		Text: fmt.Sprintf("%s • D&D 5.5 / 2024", sp.Source),
	}

	return embed
}

