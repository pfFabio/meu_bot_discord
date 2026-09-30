package ui

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"kenobi-bot/internal/dnd/model"
)

// BuildHubComponents cria botões do menu inicial
func BuildHubComponents() []discordgo.MessageComponent {
	return []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "Classes & Subclasses",
					Style:    discordgo.PrimaryButton,
					CustomID: "dnd:nav:classes",
					Emoji:    &discordgo.ComponentEmoji{Name: "🛡️"},
				},
				discordgo.Button{
					Label:    "Espécies & Linhagens",
					Style:    discordgo.PrimaryButton,
					CustomID: "dnd:nav:species",
					Emoji:    &discordgo.ComponentEmoji{Name: "🧬"},
				},
				discordgo.Button{
					Label:    "Magias & Grimório",
					Style:    discordgo.PrimaryButton,
					CustomID: "dnd:nav:spells",
					Emoji:    &discordgo.ComponentEmoji{Name: "✨"},
				},
			},
		},
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "Bestiário & Monstros",
					Style:    discordgo.SecondaryButton,
					CustomID: "dnd:nav:monsters",
					Emoji:    &discordgo.ComponentEmoji{Name: "🐉"},
				},
				discordgo.Button{
					Label:    "Itens & Maestria",
					Style:    discordgo.SecondaryButton,
					CustomID: "dnd:nav:items",
					Emoji:    &discordgo.ComponentEmoji{Name: "⚔️"},
				},
				discordgo.Button{
					Label:    "Glossário de Regras",
					Style:    discordgo.SecondaryButton,
					CustomID: "dnd:nav:rules",
					Emoji:    &discordgo.ComponentEmoji{Name: "📖"},
				},
			},
		},
	}
}

// BuildClassSelectorComponents cria menus para selecionar Classe e Subclasse
func BuildClassSelectorComponents(classes []model.Class, selectedClass *model.Class) []discordgo.MessageComponent {
	var rows []discordgo.MessageComponent

	// 1. Menu de Seleção de Classes (máx 25)
	var classOptions []discordgo.SelectMenuOption
	for _, c := range classes {
		if len(classOptions) >= 25 {
			break
		}
		opt := discordgo.SelectMenuOption{
			Label:       fmt.Sprintf("%s (%s)", c.Name, c.NamePT),
			Value:       c.ID,
			Description: fmt.Sprintf("Dado: %s • Atributo: %s", c.Traits.HitDie, c.Traits.PrimaryAbility),
			Emoji:       &discordgo.ComponentEmoji{Name: c.Emoji},
		}
		if selectedClass != nil && selectedClass.ID == c.ID {
			opt.Default = true
		}
		classOptions = append(classOptions, opt)
	}

	rows = append(rows, discordgo.ActionsRow{
		Components: []discordgo.MessageComponent{
			discordgo.SelectMenu{
				CustomID:    "dnd:class:select",
				Placeholder: "🛡️ Escolha uma Classe de D&D...",
				Options:     classOptions,
			},
		},
	})

	// 2. Se uma classe foi selecionada, exibe o Menu de Subclasses! (máx 25)
	if selectedClass != nil && len(selectedClass.Subclasses) > 0 {
		var subOptions []discordgo.SelectMenuOption
		for _, sub := range selectedClass.Subclasses {
			if len(subOptions) >= 25 {
				break
			}
			desc := sub.Synopsis
			if len(desc) > 95 {
				desc = desc[:92] + "..."
			}
			subOptions = append(subOptions, discordgo.SelectMenuOption{
				Label:       sub.Name,
				Value:       fmt.Sprintf("%s:%s", selectedClass.ID, sub.ID),
				Description: desc,
				Emoji:       &discordgo.ComponentEmoji{Name: "🌟"},
			})
		}

		rows = append(rows, discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.SelectMenu{
					CustomID:    "dnd:subclass:select",
					Placeholder: fmt.Sprintf("🌟 Escolha uma Subclasse de %s...", selectedClass.Name),
					Options:     subOptions,
				},
			},
		})
	}

	// 3. Botão de Retorno
	rows = append(rows, discordgo.ActionsRow{
		Components: []discordgo.MessageComponent{
			discordgo.Button{
				Label:    "Menu Principal",
				Style:    discordgo.SecondaryButton,
				CustomID: "dnd:nav:home",
				Emoji:    &discordgo.ComponentEmoji{Name: "🏠"},
			},
		},
	})

	return rows
}

// BuildSubclassNavComponents cria botões de navegação ao visualizar uma subclasse
func BuildSubclassNavComponents(classID string) []discordgo.MessageComponent {
	return []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "Voltar para a Classe",
					Style:    discordgo.PrimaryButton,
					CustomID: "dnd:back:class:" + classID,
					Emoji:    &discordgo.ComponentEmoji{Name: "🔙"},
				},
				discordgo.Button{
					Label:    "Menu Principal",
					Style:    discordgo.SecondaryButton,
					CustomID: "dnd:nav:home",
					Emoji:    &discordgo.ComponentEmoji{Name: "🏠"},
				},
			},
		},
	}
}

// BuildSpellFilterComponents cria filtros e lista de seleção para Magias
func BuildSpellFilterComponents(spells []model.Spell, selectedLevel int, selectedSchool string) []discordgo.MessageComponent {
	var rows []discordgo.MessageComponent

	// 1. Filtro por Nível
	levelOptions := []discordgo.SelectMenuOption{
		{Label: "Todos os Níveis", Value: "-1", Description: "Exibir magias de todos os níveis"},
		{Label: "Truques (Cantrips - Nv 0)", Value: "0", Description: "Magias básicas de nível 0"},
		{Label: "1º Nível", Value: "1", Description: "Magias de 1º círculo"},
		{Label: "2º Nível", Value: "2", Description: "Magias de 2º círculo"},
		{Label: "3º Nível (ex: Bola de Fogo)", Value: "3", Description: "Magias de 3º círculo"},
		{Label: "4º Nível", Value: "4", Description: "Magias de 4º círculo"},
		{Label: "5º Nível", Value: "5", Description: "Magias de 5º círculo"},
		{Label: "6º ao 9º Nível (Alto Nível)", Value: "6", Description: "Magias épicas de 6º a 9º círculo"},
	}
	for i := range levelOptions {
		if fmt.Sprintf("%d", selectedLevel) == levelOptions[i].Value {
			levelOptions[i].Default = true
		}
	}

	rows = append(rows, discordgo.ActionsRow{
		Components: []discordgo.MessageComponent{
			discordgo.SelectMenu{
				CustomID:    "dnd:spell:filter:level",
				Placeholder: "🔮 Filtrar por Nível da Magia...",
				Options:     levelOptions,
			},
		},
	})

	// 2. Menu de Seleção de Magias encontradas (até 25)
	if len(spells) > 0 {
		var spellOptions []discordgo.SelectMenuOption
		limit := 25
		if len(spells) < limit {
			limit = len(spells)
		}

		for _, s := range spells[:limit] {
			lvlStr := fmt.Sprintf("Nv %d", s.Level)
			if s.Level == 0 {
				lvlStr = "Truque"
			}
			desc := fmt.Sprintf("%s • %s • %s", lvlStr, s.School, s.CastingTime)
			if len(desc) > 95 {
				desc = desc[:92] + "..."
			}
			spellOptions = append(spellOptions, discordgo.SelectMenuOption{
				Label:       s.Name,
				Value:       s.ID,
				Description: desc,
				Emoji:       &discordgo.ComponentEmoji{Name: "✨"},
			})
		}

		rows = append(rows, discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.SelectMenu{
					CustomID:    "dnd:spell:select",
					Placeholder: fmt.Sprintf("📖 Selecione uma Magia (%d encontradas)...", len(spells)),
					Options:     spellOptions,
				},
			},
		})
	}

	// 3. Botões
	rows = append(rows, discordgo.ActionsRow{
		Components: []discordgo.MessageComponent{
			discordgo.Button{
				Label:    "Menu Principal",
				Style:    discordgo.SecondaryButton,
				CustomID: "dnd:nav:home",
				Emoji:    &discordgo.ComponentEmoji{Name: "🏠"},
			},
		},
	})

	return rows
}

// BuildMonsterFilterComponents cria filtros e seleção para o Bestiário
func BuildMonsterFilterComponents(monsters []model.Monster, selectedCR string) []discordgo.MessageComponent {
	var rows []discordgo.MessageComponent

	// 1. Filtro por CR
	crOptions := []discordgo.SelectMenuOption{
		{Label: "Todos os Desafios (CR)", Value: "all", Description: "Todos os monstros"},
		{Label: "CR 0 a 1/2 (Iniciantes/Animais)", Value: "low", Description: "Cobras, Goblins, Bandidos"},
		{Label: "CR 1 a 4 (Desafio Baixo)", Value: "tier1", Description: "Carniçais, Ogros, Bugbears"},
		{Label: "CR 5 a 9 (Desafio Médio)", Value: "tier2", Description: "Elementais, Golens, Hidras"},
		{Label: "CR 10 a 15 (Desafio Alto)", Value: "tier3", Description: "Aboleth, Beholder, Dragão Jovem"},
		{Label: "CR 16+ (Lendários / Bosses)", Value: "tier4", Description: "Dragões Adultos, Balor, Lich"},
	}

	rows = append(rows, discordgo.ActionsRow{
		Components: []discordgo.MessageComponent{
			discordgo.SelectMenu{
				CustomID:    "dnd:monster:filter:cr",
				Placeholder: "🐉 Filtrar por Nível de Desafio (CR)...",
				Options:     crOptions,
			},
		},
	})

	// 2. Menu de Seleção de Monstros (até 25)
	if len(monsters) > 0 {
		var monOptions []discordgo.SelectMenuOption
		limit := 25
		if len(monsters) < limit {
			limit = len(monsters)
		}

		for _, m := range monsters[:limit] {
			desc := fmt.Sprintf("CR %s • %s • %s PV", m.ChallengeRating, m.Type, m.HitPoints)
			if len(desc) > 95 {
				desc = desc[:92] + "..."
			}
			monOptions = append(monOptions, discordgo.SelectMenuOption{
				Label:       m.Name,
				Value:       m.ID,
				Description: desc,
				Emoji:       &discordgo.ComponentEmoji{Name: "🐉"},
			})
		}

		rows = append(rows, discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.SelectMenu{
					CustomID:    "dnd:monster:select",
					Placeholder: fmt.Sprintf("📜 Selecione um Monstro (%d encontrados)...", len(monsters)),
					Options:     monOptions,
				},
			},
		})
	}

	// 3. Botão Home
	rows = append(rows, discordgo.ActionsRow{
		Components: []discordgo.MessageComponent{
			discordgo.Button{
				Label:    "Menu Principal",
				Style:    discordgo.SecondaryButton,
				CustomID: "dnd:nav:home",
				Emoji:    &discordgo.ComponentEmoji{Name: "🏠"},
			},
		},
	})

	return rows
}

// BuildItemFilterComponents cria filtros e seleção para Itens e Maestria
func BuildItemFilterComponents(items []model.Item, selectedCat string) []discordgo.MessageComponent {
	var rows []discordgo.MessageComponent

	// 1. Filtro por Categoria
	catOptions := []discordgo.SelectMenuOption{
		{Label: "Todos os Itens", Value: "all", Description: "Armas com Maestria e Itens Mágicos"},
		{Label: "Armas com Maestria (Weapon Mastery)", Value: "weapon", Description: "Cleave, Graze, Vex, Nick, Topple, Sap, Push, Slow"},
		{Label: "Itens Mágicos & Poções", Value: "magic", Description: "Poções com Ação Bônus, Itens Maravilhosos"},
	}

	rows = append(rows, discordgo.ActionsRow{
		Components: []discordgo.MessageComponent{
			discordgo.SelectMenu{
				CustomID:    "dnd:item:filter:cat",
				Placeholder: "⚔️ Filtrar por Categoria de Item...",
				Options:     catOptions,
			},
		},
	})

	// 2. Menu de Seleção de Itens (até 25)
	if len(items) > 0 {
		var itemOptions []discordgo.SelectMenuOption
		limit := 25
		if len(items) < limit {
			limit = len(items)
		}

		for _, it := range items[:limit] {
			desc := it.Category
			if it.Mastery != "" {
				desc = "Maestria: " + it.Mastery
			} else if it.Rarity != "" {
				desc = "Raridade: " + it.Rarity
			}
			if len(desc) > 95 {
				desc = desc[:92] + "..."
			}
			itemOptions = append(itemOptions, discordgo.SelectMenuOption{
				Label:       it.Name,
				Value:       it.ID,
				Description: desc,
				Emoji:       &discordgo.ComponentEmoji{Name: "⚔️"},
			})
		}

		rows = append(rows, discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.SelectMenu{
					CustomID:    "dnd:item:select",
					Placeholder: fmt.Sprintf("🛡️ Selecione um Item (%d encontrados)...", len(items)),
					Options:     itemOptions,
				},
			},
		})
	}

	// 3. Botão Home
	rows = append(rows, discordgo.ActionsRow{
		Components: []discordgo.MessageComponent{
			discordgo.Button{
				Label:    "Menu Principal",
				Style:    discordgo.SecondaryButton,
				CustomID: "dnd:nav:home",
				Emoji:    &discordgo.ComponentEmoji{Name: "🏠"},
			},
		},
	})

	return rows
}

// BuildRuleComponents cria menu de seleção para o Glossário de Regras
func BuildRuleComponents(rules []model.Rule) []discordgo.MessageComponent {
	var rows []discordgo.MessageComponent

	var ruleOptions []discordgo.SelectMenuOption
	limit := 25
	if len(rules) < limit {
		limit = len(rules)
	}

	for _, r := range rules[:limit] {
		desc := r.Summary
		if len(desc) > 95 {
			desc = desc[:92] + "..."
		}
		ruleOptions = append(ruleOptions, discordgo.SelectMenuOption{
			Label:       r.Name,
			Value:       r.ID,
			Description: desc,
			Emoji:       &discordgo.ComponentEmoji{Name: "📖"},
		})
	}

	rows = append(rows, discordgo.ActionsRow{
		Components: []discordgo.MessageComponent{
			discordgo.SelectMenu{
				CustomID:    "dnd:rule:select",
				Placeholder: "📖 Escolha uma Regra do D&D 2024 para consultar...",
				Options:     ruleOptions,
			},
		},
	})

	rows = append(rows, discordgo.ActionsRow{
		Components: []discordgo.MessageComponent{
			discordgo.Button{
				Label:    "Menu Principal",
				Style:    discordgo.SecondaryButton,
				CustomID: "dnd:nav:home",
				Emoji:    &discordgo.ComponentEmoji{Name: "🏠"},
			},
		},
	})

	return rows
}

// BuildSpeciesSelectorComponents cria componentes para selecionar Espécies e filtrar por categoria
func BuildSpeciesSelectorComponents(speciesList []model.Species, selected *model.Species, currentCategory string) []discordgo.MessageComponent {
	var rows []discordgo.MessageComponent

	// 1. Menu de Filtro por Categoria
	catOptions := []discordgo.SelectMenuOption{
		{Label: "🦇 Ravenloft (Linhagens Sombrias)", Value: "Ravenloft", Description: "Dhampir, Hexblood, Lupin, Reborn"},
		{Label: "👥 Espécies Comuns (PHB 2024)", Value: "Common", Description: "Humano, Elfo, Anão, Halfling, Aasimar, Draconato..."},
		{Label: "🌐 Todas as Espécies", Value: "all", Description: "Exibir todas as opções disponíveis"},
	}
	for idx := range catOptions {
		if strings.EqualFold(catOptions[idx].Value, currentCategory) {
			catOptions[idx].Default = true
		}
	}

	rows = append(rows, discordgo.ActionsRow{
		Components: []discordgo.MessageComponent{
			discordgo.SelectMenu{
				CustomID:    "dnd:species:filter:cat",
				Placeholder: "📁 Filtrar por Categoria (Ravenloft / 2024)...",
				Options:     catOptions,
			},
		},
	})

	// 2. Menu de Seleção de Espécies (máx 25)
	if len(speciesList) > 0 {
		var spOptions []discordgo.SelectMenuOption
		limit := 25
		if len(speciesList) < limit {
			limit = len(speciesList)
		}
		for _, sp := range speciesList[:limit] {
			emoji := "🧬"
			if sp.Category == "Ravenloft" {
				emoji = "🦇"
			}
			desc := fmt.Sprintf("%s • %s • %s", sp.CreatureType, sp.Size, sp.Speed)
			if len(desc) > 95 {
				desc = desc[:92] + "..."
			}
			opt := discordgo.SelectMenuOption{
				Label:       fmt.Sprintf("%s (%s)", sp.Name, sp.NamePT),
				Value:       sp.ID,
				Description: desc,
				Emoji:       &discordgo.ComponentEmoji{Name: emoji},
			}
			if selected != nil && selected.ID == sp.ID {
				opt.Default = true
			}
			spOptions = append(spOptions, opt)
		}

		rows = append(rows, discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.SelectMenu{
					CustomID:    "dnd:species:select",
					Placeholder: fmt.Sprintf("🧬 Escolha uma Espécie (%d disponíveis)...", len(speciesList)),
					Options:     spOptions,
				},
			},
		})
	}

	// 3. Botão Home
	rows = append(rows, discordgo.ActionsRow{
		Components: []discordgo.MessageComponent{
			discordgo.Button{
				Label:    "Menu Principal",
				Style:    discordgo.SecondaryButton,
				CustomID: "dnd:nav:home",
				Emoji:    &discordgo.ComponentEmoji{Name: "🏠"},
			},
		},
	})

	return rows
}

