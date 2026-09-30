package handler

import (
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"kenobi-bot/internal/dnd/model"
	"kenobi-bot/internal/dnd/repository"
	"kenobi-bot/internal/ui"
)

// DNDHandler processa todas as interações e comandos relacionados a D&D 2024
type DNDHandler struct {
	repo repository.DNDRepository
}

// NewDNDHandler cria uma nova instância do handler
func NewDNDHandler(repo repository.DNDRepository) *DNDHandler {
	return &DNDHandler{repo: repo}
}

// GetSlashCommands retorna as definições dos Comandos Slash do D&D 2024
func (h *DNDHandler) GetSlashCommands() []*discordgo.ApplicationCommand {
	return []*discordgo.ApplicationCommand{
		{
			Name:        "dnd",
			Description: "Abre o Compêndio Interativo de D&D 5.5 (Edição 2024).",
		},
		{
			Name:        "classes",
			Description: "Navegue pelas 12 classes e 48 subclasses do D&D 2024.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "nome",
					Description: "Nome da classe ou subclasse para consultar diretamente.",
					Required:    false,
				},
			},
		},
		{
			Name:        "magias",
			Description: "Consulte o grimório de magias oficiais do D&D 2024.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "nome",
					Description: "Nome da magia (ex: Fireball, Shield, Cure Wounds).",
					Required:    false,
				},
			},
		},
		{
			Name:        "monstros",
			Description: "Consulte o Bestiário oficial do D&D 2024.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "nome",
					Description: "Nome do monstro (ex: Dragon, Aboleth, Goblin).",
					Required:    false,
				},
			},
		},
		{
			Name:        "itens",
			Description: "Consulte armas com Maestria e itens mágicos do D&D 2024.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "nome",
					Description: "Nome do item ou arma.",
					Required:    false,
				},
			},
		},
		{
			Name:        "regras",
			Description: "Consulte o Glossário de Regras oficiais do D&D 2024.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "termo",
					Description: "Termo ou condição a consultar (ex: Invisível, Agarrado, Descanso).",
					Required:    false,
				},
			},
		},
		{
			Name:        "racas",
			Description: "Consulte espécies e linhagens de Ravenloft e D&D 2024 (Dhampir, Hexblood, Lupin, Reborn...).",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "nome",
					Description: "Nome da espécie ou linhagem (ex: Dhampir, Hexblood, Lupin, Reborn, Elfo, Humano).",
					Required:    false,
				},
			},
		},
	}
}

// HandleSlashCommand processa comandos de barra
func (h *DNDHandler) HandleSlashCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	cmdName := i.ApplicationCommandData().Name

	switch cmdName {
	case "dnd":
		h.respondHome(s, i)

	case "classes":
		opts := i.ApplicationCommandData().Options
		if len(opts) > 0 && opts[0].StringValue() != "" {
			query := opts[0].StringValue()
			if cls, found := h.repo.GetClass(query); found {
				h.respondClass(s, i, cls)
				return
			}
			if sub, found := h.repo.GetSubclass("", query); found {
				cls, _ := h.repo.GetClass(sub.ClassName)
				h.respondSubclass(s, i, sub, cls)
				return
			}
		}
		// Se não passou parâmetro, abre o menu de classes
		classes := h.repo.GetClasses()
		embed := ui.BuildHubEmbed()
		embed.Title = "🛡️ Classes & Subclasses de D&D 2024"
		embed.Description = "Selecione uma das 12 classes no menu abaixo para ver sua sinopse, dados de vida, proficiências e subclasses!"
		comps := ui.BuildClassSelectorComponents(classes, nil)

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds:     []*discordgo.MessageEmbed{embed},
				Components: comps,
			},
		})

	case "magias":
		opts := i.ApplicationCommandData().Options
		if len(opts) > 0 && opts[0].StringValue() != "" {
			query := opts[0].StringValue()
			if spell, found := h.repo.GetSpell(query); found {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Embeds:     []*discordgo.MessageEmbed{ui.BuildSpellEmbed(spell)},
						Components: ui.BuildSubclassNavComponents(""),
					},
				})
				return
			}
		}
		h.showSpellFilter(s, i, -1, "")

	case "monstros":
		opts := i.ApplicationCommandData().Options
		if len(opts) > 0 && opts[0].StringValue() != "" {
			query := opts[0].StringValue()
			if mon, found := h.repo.GetMonster(query); found {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Embeds:     []*discordgo.MessageEmbed{ui.BuildMonsterEmbed(mon)},
						Components: ui.BuildSubclassNavComponents(""),
					},
				})
				return
			}
		}
		h.showMonsterFilter(s, i, "all")

	case "itens":
		opts := i.ApplicationCommandData().Options
		if len(opts) > 0 && opts[0].StringValue() != "" {
			query := opts[0].StringValue()
			if item, found := h.repo.GetItem(query); found {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Embeds:     []*discordgo.MessageEmbed{ui.BuildItemEmbed(item)},
						Components: ui.BuildSubclassNavComponents(""),
					},
				})
				return
			}
		}
		h.showItemFilter(s, i, "all")

	case "regras":
		opts := i.ApplicationCommandData().Options
		if len(opts) > 0 && opts[0].StringValue() != "" {
			query := opts[0].StringValue()
			if rule, found := h.repo.GetRule(query); found {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Embeds:     []*discordgo.MessageEmbed{ui.BuildRuleEmbed(rule)},
						Components: ui.BuildSubclassNavComponents(""),
					},
				})
				return
			}
		}
		h.showRulesMenu(s, i)

	case "racas", "especies":
		opts := i.ApplicationCommandData().Options
		if len(opts) > 0 && opts[0].StringValue() != "" {
			query := opts[0].StringValue()
			if sp, found := h.repo.GetSpecies(query); found {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Embeds:     []*discordgo.MessageEmbed{ui.BuildSpeciesEmbed(sp)},
						Components: ui.BuildSubclassNavComponents(""),
					},
				})
				return
			}
		}
		h.showSpeciesMenu(s, i, "Ravenloft")
	}
}

// HandleComponentInteraction processa cliques em botões e menus suspensos (SelectMenu)
func (h *DNDHandler) HandleComponentInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	customID := i.MessageComponentData().CustomID

	switch {
	case customID == "dnd:nav:home":
		h.updateHome(s, i)

	case customID == "dnd:nav:classes":
		classes := h.repo.GetClasses()
		embed := ui.BuildHubEmbed()
		embed.Title = "🛡️ Classes & Subclasses de D&D 2024"
		embed.Description = "Selecione uma das 12 classes no menu abaixo para ver sua sinopse, dados de vida, proficiências e subclasses!"
		comps := ui.BuildClassSelectorComponents(classes, nil)
		h.updateResponse(s, i, embed, comps)

	case customID == "dnd:nav:spells":
		h.updateSpellFilter(s, i, -1, "")

	case customID == "dnd:nav:monsters":
		h.updateMonsterFilter(s, i, "all")

	case customID == "dnd:nav:items":
		h.updateItemFilter(s, i, "all")

	case customID == "dnd:nav:rules":
		h.updateRulesMenu(s, i)

	case customID == "dnd:nav:species":
		h.updateSpeciesMenu(s, i, "Ravenloft", nil)

	case customID == "dnd:species:filter:cat":
		values := i.MessageComponentData().Values
		if len(values) > 0 {
			h.updateSpeciesMenu(s, i, values[0], nil)
		}

	case customID == "dnd:species:select":
		values := i.MessageComponentData().Values
		if len(values) > 0 {
			if sp, found := h.repo.GetSpecies(values[0]); found {
				embed := ui.BuildSpeciesEmbed(sp)
				comps := ui.BuildSubclassNavComponents("")
				h.updateResponse(s, i, embed, comps)
			}
		}

	case customID == "dnd:class:select":
		values := i.MessageComponentData().Values
		if len(values) > 0 {
			classID := values[0]
			if cls, found := h.repo.GetClass(classID); found {
				embed := ui.BuildClassEmbed(cls)
				comps := ui.BuildClassSelectorComponents(h.repo.GetClasses(), cls)
				h.updateResponse(s, i, embed, comps)
			}
		}

	case customID == "dnd:subclass:select":
		values := i.MessageComponentData().Values
		if len(values) > 0 {
			parts := strings.Split(values[0], ":")
			if len(parts) == 2 {
				classID, subID := parts[0], parts[1]
				if sub, found := h.repo.GetSubclass(classID, subID); found {
					cls, _ := h.repo.GetClass(classID)
					embed := ui.BuildSubclassEmbed(sub, cls)
					comps := ui.BuildSubclassNavComponents(classID)
					h.updateResponse(s, i, embed, comps)
				}
			}
		}

	case strings.HasPrefix(customID, "dnd:back:class:"):
		classID := strings.TrimPrefix(customID, "dnd:back:class:")
		if cls, found := h.repo.GetClass(classID); found {
			embed := ui.BuildClassEmbed(cls)
			comps := ui.BuildClassSelectorComponents(h.repo.GetClasses(), cls)
			h.updateResponse(s, i, embed, comps)
		} else {
			h.updateHome(s, i)
		}

	case customID == "dnd:spell:filter:level":
		values := i.MessageComponentData().Values
		if len(values) > 0 {
			lvl, _ := strconv.Atoi(values[0])
			h.updateSpellFilter(s, i, lvl, "")
		}

	case customID == "dnd:spell:select":
		values := i.MessageComponentData().Values
		if len(values) > 0 {
			if spell, found := h.repo.GetSpell(values[0]); found {
				embed := ui.BuildSpellEmbed(spell)
				comps := ui.BuildSubclassNavComponents("")
				h.updateResponse(s, i, embed, comps)
			}
		}

	case customID == "dnd:monster:filter:cr":
		values := i.MessageComponentData().Values
		if len(values) > 0 {
			h.updateMonsterFilter(s, i, values[0])
		}

	case customID == "dnd:monster:select":
		values := i.MessageComponentData().Values
		if len(values) > 0 {
			if mon, found := h.repo.GetMonster(values[0]); found {
				embed := ui.BuildMonsterEmbed(mon)
				comps := ui.BuildSubclassNavComponents("")
				h.updateResponse(s, i, embed, comps)
			}
		}

	case customID == "dnd:item:filter:cat":
		values := i.MessageComponentData().Values
		if len(values) > 0 {
			h.updateItemFilter(s, i, values[0])
		}

	case customID == "dnd:item:select":
		values := i.MessageComponentData().Values
		if len(values) > 0 {
			if it, found := h.repo.GetItem(values[0]); found {
				embed := ui.BuildItemEmbed(it)
				comps := ui.BuildSubclassNavComponents("")
				h.updateResponse(s, i, embed, comps)
			}
		}

	case customID == "dnd:rule:select":
		values := i.MessageComponentData().Values
		if len(values) > 0 {
			if rule, found := h.repo.GetRule(values[0]); found {
				embed := ui.BuildRuleEmbed(rule)
				comps := ui.BuildSubclassNavComponents("")
				h.updateResponse(s, i, embed, comps)
			}
		}
	}
}

func (h *DNDHandler) respondHome(s *discordgo.Session, i *discordgo.InteractionCreate) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds:     []*discordgo.MessageEmbed{ui.BuildHubEmbed()},
			Components: ui.BuildHubComponents(),
		},
	})
}

func (h *DNDHandler) updateHome(s *discordgo.Session, i *discordgo.InteractionCreate) {
	h.updateResponse(s, i, ui.BuildHubEmbed(), ui.BuildHubComponents())
}

func (h *DNDHandler) respondClass(s *discordgo.Session, i *discordgo.InteractionCreate, cls *model.Class) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds:     []*discordgo.MessageEmbed{ui.BuildClassEmbed(cls)},
			Components: ui.BuildClassSelectorComponents(h.repo.GetClasses(), cls),
		},
	})
}

func (h *DNDHandler) respondSubclass(s *discordgo.Session, i *discordgo.InteractionCreate, sub *model.Subclass, cls *model.Class) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds:     []*discordgo.MessageEmbed{ui.BuildSubclassEmbed(sub, cls)},
			Components: ui.BuildSubclassNavComponents(cls.ID),
		},
	})
}

func (h *DNDHandler) showSpellFilter(s *discordgo.Session, i *discordgo.InteractionCreate, level int, school string) {
	spells := h.repo.SearchSpells("", level, school, "", 25)
	embed := &discordgo.MessageEmbed{
		Title:       "✨ Grimório de Magias D&D 2024",
		Description: fmt.Sprintf("Mostrando magias encontradas (%d no total). Use o menu para escolher o nível ou abrir os detalhes da magia!", len(spells)),
		Color:       ui.ColorDNDBlue,
	}
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds:     []*discordgo.MessageEmbed{embed},
			Components: ui.BuildSpellFilterComponents(spells, level, school),
		},
	})
}

func (h *DNDHandler) updateSpellFilter(s *discordgo.Session, i *discordgo.InteractionCreate, level int, school string) {
	spells := h.repo.SearchSpells("", level, school, "", 25)
	lvlName := "Todos os Níveis"
	if level == 0 {
		lvlName = "Truques"
	} else if level > 0 {
		lvlName = fmt.Sprintf("%dº Nível", level)
	}

	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("✨ Grimório de Magias - %s", lvlName),
		Description: fmt.Sprintf("Foram encontradas **%d magias**. Escolha uma magia no menu para ver tempo de conjuração, alcance, componentes e descrição completa!", len(spells)),
		Color:       ui.ColorDNDBlue,
	}
	h.updateResponse(s, i, embed, ui.BuildSpellFilterComponents(spells, level, school))
}

func (h *DNDHandler) showMonsterFilter(s *discordgo.Session, i *discordgo.InteractionCreate, cr string) {
	monsters := h.filterMonstersByCR(cr)
	embed := &discordgo.MessageEmbed{
		Title:       "🐉 Bestiário de Monstros D&D 2024",
		Description: "Filtre monstros pelo Nível de Desafio (CR) e selecione para ver ficha de combate completa!",
		Color:       ui.ColorDNDGold,
	}
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds:     []*discordgo.MessageEmbed{embed},
			Components: ui.BuildMonsterFilterComponents(monsters, cr),
		},
	})
}

func (h *DNDHandler) updateMonsterFilter(s *discordgo.Session, i *discordgo.InteractionCreate, cr string) {
	monsters := h.filterMonstersByCR(cr)
	embed := &discordgo.MessageEmbed{
		Title:       "🐉 Bestiário de Monstros D&D 2024",
		Description: fmt.Sprintf("Mostrando **%d monstros**. Selecione um monstro no menu abaixo para ver CA, PV, Atributos e Ações!", len(monsters)),
		Color:       ui.ColorDNDGold,
	}
	h.updateResponse(s, i, embed, ui.BuildMonsterFilterComponents(monsters, cr))
}

func (h *DNDHandler) filterMonstersByCR(cr string) []model.Monster {
	all := h.repo.SearchMonsters("", "", "", 100)
	if cr == "all" || cr == "" {
		if len(all) > 25 {
			return all[:25]
		}
		return all
	}

	var filtered []model.Monster
	for _, m := range all {
		crLow := strings.ToLower(m.ChallengeRating)
		switch cr {
		case "low":
			if strings.Contains(crLow, "0") || strings.Contains(crLow, "1/8") || strings.Contains(crLow, "1/4") || strings.Contains(crLow, "1/2") {
				filtered = append(filtered, m)
			}
		case "tier1":
			if strings.HasPrefix(crLow, "1 ") || strings.HasPrefix(crLow, "2") || strings.HasPrefix(crLow, "3") || strings.HasPrefix(crLow, "4") {
				filtered = append(filtered, m)
			}
		case "tier2":
			if strings.HasPrefix(crLow, "5") || strings.HasPrefix(crLow, "6") || strings.HasPrefix(crLow, "7") || strings.HasPrefix(crLow, "8") || strings.HasPrefix(crLow, "9") {
				filtered = append(filtered, m)
			}
		case "tier3":
			if strings.HasPrefix(crLow, "10") || strings.HasPrefix(crLow, "11") || strings.HasPrefix(crLow, "12") || strings.HasPrefix(crLow, "13") || strings.HasPrefix(crLow, "14") || strings.HasPrefix(crLow, "15") {
				filtered = append(filtered, m)
			}
		case "tier4":
			if strings.HasPrefix(crLow, "16") || strings.HasPrefix(crLow, "17") || strings.HasPrefix(crLow, "18") || strings.HasPrefix(crLow, "19") || strings.HasPrefix(crLow, "2") {
				filtered = append(filtered, m)
			}
		}
		if len(filtered) >= 25 {
			break
		}
	}
	if len(filtered) == 0 {
		if len(all) > 25 {
			return all[:25]
		}
		return all
	}
	return filtered
}

func (h *DNDHandler) showItemFilter(s *discordgo.Session, i *discordgo.InteractionCreate, cat string) {
	items := h.filterItemsByCat(cat)
	embed := &discordgo.MessageEmbed{
		Title:       "⚔️ Equipamentos & Itens Mágicos D&D 2024",
		Description: "Explore armas com Maestria (Weapon Mastery) e itens mágicos.",
		Color:       ui.ColorDNDGreen,
	}
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds:     []*discordgo.MessageEmbed{embed},
			Components: ui.BuildItemFilterComponents(items, cat),
		},
	})
}

func (h *DNDHandler) updateItemFilter(s *discordgo.Session, i *discordgo.InteractionCreate, cat string) {
	items := h.filterItemsByCat(cat)
	embed := &discordgo.MessageEmbed{
		Title:       "⚔️ Equipamentos & Itens Mágicos D&D 2024",
		Description: fmt.Sprintf("Mostrando **%d itens**. Selecione um item no menu para ver dano, propriedades de maestria e efeitos mágicos!", len(items)),
		Color:       ui.ColorDNDGreen,
	}
	h.updateResponse(s, i, embed, ui.BuildItemFilterComponents(items, cat))
}

func (h *DNDHandler) filterItemsByCat(cat string) []model.Item {
	all := h.repo.SearchItems("", "", "", "", 50)
	if cat == "all" || cat == "" {
		return all
	}
	var filtered []model.Item
	for _, it := range all {
		if cat == "weapon" && it.Mastery != "" {
			filtered = append(filtered, it)
		} else if cat == "magic" && it.Rarity != "" && !strings.EqualFold(it.Rarity, "Mundane") {
			filtered = append(filtered, it)
		}
	}
	if len(filtered) == 0 {
		return all
	}
	return filtered
}

func (h *DNDHandler) showRulesMenu(s *discordgo.Session, i *discordgo.InteractionCreate) {
	rules := h.repo.SearchRules("", 25)
	embed := &discordgo.MessageEmbed{
		Title:       "📖 Glossário de Regras Oficiais D&D 2024",
		Description: "Escolha uma regra ou condição no menu abaixo para ler sua definição completa.",
		Color:       ui.ColorDNDDark,
	}
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds:     []*discordgo.MessageEmbed{embed},
			Components: ui.BuildRuleComponents(rules),
		},
	})
}

func (h *DNDHandler) updateRulesMenu(s *discordgo.Session, i *discordgo.InteractionCreate) {
	rules := h.repo.SearchRules("", 25)
	embed := &discordgo.MessageEmbed{
		Title:       "📖 Glossário de Regras Oficiais D&D 2024",
		Description: "Escolha uma regra ou condição no menu abaixo para ler sua definição completa.",
		Color:       ui.ColorDNDDark,
	}
	h.updateResponse(s, i, embed, ui.BuildRuleComponents(rules))
}

func (h *DNDHandler) showSpeciesMenu(s *discordgo.Session, i *discordgo.InteractionCreate, category string) {
	speciesList := h.repo.GetSpeciesList(category)
	embed := ui.BuildHubEmbed()
	embed.Title = "🧬 Espécies & Linhagens de D&D (Ravenloft & 2024)"
	embed.Description = "Selecione uma espécie no menu abaixo para ver traços raciais, velocidade, visão no escuro e habilidades especiais!\n\n🦇 **Destaque Ravenloft:** Dhampir, Hexblood, Lupin e Reborn."
	comps := ui.BuildSpeciesSelectorComponents(speciesList, nil, category)

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds:     []*discordgo.MessageEmbed{embed},
			Components: comps,
		},
	})
}

func (h *DNDHandler) updateSpeciesMenu(s *discordgo.Session, i *discordgo.InteractionCreate, category string, selected *model.Species) {
	speciesList := h.repo.GetSpeciesList(category)
	embed := ui.BuildHubEmbed()
	embed.Title = "🧬 Espécies & Linhagens de D&D (Ravenloft & 2024)"
	embed.Description = "Selecione uma espécie no menu abaixo para ver traços raciais, velocidade, visão no escuro e habilidades especiais!\n\n🦇 **Destaque Ravenloft:** Dhampir, Hexblood, Lupin e Reborn."
	comps := ui.BuildSpeciesSelectorComponents(speciesList, selected, category)
	h.updateResponse(s, i, embed, comps)
}

func (h *DNDHandler) updateResponse(s *discordgo.Session, i *discordgo.InteractionCreate, embed *discordgo.MessageEmbed, comps []discordgo.MessageComponent) {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Embeds:     []*discordgo.MessageEmbed{embed},
			Components: comps,
		},
	})
	if err != nil {
		slog.Error("Erro ao atualizar componente de interação", "error", err)
	}
}
