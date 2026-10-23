package ui

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"kenobi-bot/internal/character"
)

// Helper para barra de vida visual
func renderHPBar(cur, max, temp int) string {
	if max <= 0 {
		return "N/A"
	}
	ratio := float64(cur) / float64(max)
	if ratio < 0 {
		ratio = 0
	} else if ratio > 1 {
		ratio = 1
	}

	totalBlocks := 10
	filled := int(ratio * float64(totalBlocks))
	if filled == 0 && cur > 0 {
		filled = 1
	}

	var blockChar string
	if ratio > 0.5 {
		blockChar = "🟩"
	} else if ratio > 0.25 {
		blockChar = "🟨"
	} else if ratio > 0 {
		blockChar = "🟥"
	} else {
		blockChar = "⬛"
	}

	bar := strings.Repeat(blockChar, filled) + strings.Repeat("⬜", totalBlocks-filled)
	res := fmt.Sprintf("%s **%d/%d PV**", bar, cur, max)
	if temp > 0 {
		res += fmt.Sprintf(" *(+%d Temp)*", temp)
	}
	return res
}

// formatAdvantageLabel formata o indicador textual do modo de vantagem/desvantagem
func formatAdvantageLabel(mode string) string {
	switch mode {
	case "advantage":
		return "✨ `Vantagem`"
	case "disadvantage":
		return "💀 `Desvantagem`"
	default:
		return "⚖️ `Normal`"
	}
}

// buildAdvantageButton cria o botão dinâmico para alternar Vantagem / Desvantagem / Normal
func buildAdvantageButton(sheet *character.CharacterSheet) discordgo.Button {
	mode := sheet.GetAdvantageMode()
	switch mode {
	case "advantage":
		return discordgo.Button{
			Label:    "Vantagem (Ativa)",
			Emoji:    &discordgo.ComponentEmoji{Name: "✨"},
			Style:    discordgo.SuccessButton,
			CustomID: "sheet:adv:toggle",
		}
	case "disadvantage":
		return discordgo.Button{
			Label:    "Desvantagem (Ativa)",
			Emoji:    &discordgo.ComponentEmoji{Name: "💀"},
			Style:    discordgo.DangerButton,
			CustomID: "sheet:adv:toggle",
		}
	default:
		return discordgo.Button{
			Label:    "Normal (Vant/Desv)",
			Emoji:    &discordgo.ComponentEmoji{Name: "⚖️"},
			Style:    discordgo.SecondaryButton,
			CustomID: "sheet:adv:toggle",
		}
	}
}

// Helper para medidor de slots de magia
func renderSlotPips(max, used int) string {
	avail := max - used
	if avail < 0 {
		avail = 0
	}
	pips := ""
	for i := 0; i < avail; i++ {
		pips += "▰"
	}
	for i := 0; i < used; i++ {
		pips += "▱"
	}
	return fmt.Sprintf("%s `(%d/%d)`", pips, avail, max)
}

// getHPColor define a cor do embed baseado no estado de vida
func getHPColor(sheet *character.CharacterSheet) int {
	if sheet.CurrentHP <= 0 {
		return 0x2B2D42 // Quase preto/cinza mórbido
	}
	ratio := float64(sheet.CurrentHP) / float64(sheet.MaxHP)
	if ratio > 0.5 {
		return 0x2ECC71 // Verde vívido
	} else if ratio > 0.25 {
		return 0xF1C40F // Amarelo/Ouro alerta
	}
	return 0xE74C3C // Vermelho perigo crítico
}

// BuildCharacterOverviewEmbed gera a visão geral de combate e status
func BuildCharacterOverviewEmbed(sheet *character.CharacterSheet) *discordgo.MessageEmbed {
	color := getHPColor(sheet)

	subclassStr := ""
	if sheet.Subclass != "" {
		subclassStr = fmt.Sprintf(" (%s)", sheet.Subclass)
	}

	title := fmt.Sprintf("🛡️ %s — Nível %d %s%s", sheet.Name, sheet.Level, sheet.Class, subclassStr)
	desc := fmt.Sprintf("**Raça/Espécie:** %s | **Antecedente:** %s | **Alinhamento:** %s",
		sheet.Race, sheet.Background, sheet.Alignment)

	fields := []*discordgo.MessageEmbedField{
		{
			Name:   "❤️ Pontos de Vida (HP)",
			Value:  renderHPBar(sheet.CurrentHP, sheet.MaxHP, sheet.TempHP),
			Inline: false,
		},
		{
			Name: "⚔️ Combate & Defesa",
			Value: fmt.Sprintf("🛡️ **CA:** `%d` • ⚡ **Iniciativa:** `%+d`\n👟 **Deslocamento:** `%s`\n🎲 **Rolagem:** %s",
				sheet.ArmorClass, sheet.Initiative, sheet.Speed, formatAdvantageLabel(sheet.GetAdvantageMode())),
			Inline: true,
		},
		{
			Name: "🎯 Eficácia & Dados",
			Value: fmt.Sprintf("✨ **Proficiência:** `+%d`\n🎲 **Dados de Vida:** `%d/%d`\n👁️ **Percepção Passiva:** `%d`",
				sheet.ProficiencyBonus, sheet.HitDiceCurrent, sheet.Level, sheet.PassivePerception),
			Inline: true,
		},
		{
			Name: "📊 Atributos Principais",
			Value: fmt.Sprintf("`FOR %2d (%+d)` • `DES %2d (%+d)` • `CON %2d (%+d)`\n`INT %2d (%+d)` • `SAB %2d (%+d)` • `CAR %2d (%+d)`",
				sheet.Attributes.STR.Score, sheet.Attributes.STR.Modifier,
				sheet.Attributes.DEX.Score, sheet.Attributes.DEX.Modifier,
				sheet.Attributes.CON.Score, sheet.Attributes.CON.Modifier,
				sheet.Attributes.INT.Score, sheet.Attributes.INT.Modifier,
				sheet.Attributes.WIS.Score, sheet.Attributes.WIS.Modifier,
				sheet.Attributes.CHA.Score, sheet.Attributes.CHA.Modifier),
			Inline: false,
		},
	}

	// Salvaguardas com destaque nas proficientes
	var savesStr []string
	saveNames := []struct {
		Key  string
		Name string
	}{
		{"str", "FOR"}, {"dex", "DES"}, {"con", "CON"},
		{"int", "INT"}, {"wis", "SAB"}, {"cha", "CAR"},
	}
	for _, sn := range saveNames {
		s, has := sheet.SavingThrows[sn.Key]
		if has && s.Proficient {
			savesStr = append(savesStr, fmt.Sprintf("⭐ **%s:** `%+d`", sn.Name, s.Modifier))
		} else {
			mod := sheet.GetAttributeModifier(sn.Key)
			savesStr = append(savesStr, fmt.Sprintf("%s: `%+d`", sn.Name, mod))
		}
	}
	fields = append(fields, &discordgo.MessageEmbedField{
		Name:   "🛡️ Salvaguardas (Resistências)",
		Value:  strings.Join(savesStr, " • "),
		Inline: false,
	})

	// Recursos limitados diários
	if len(sheet.LimitedUses) > 0 {
		var resList []string
		for _, r := range sheet.LimitedUses {
			pips := renderSlotPips(r.MaxUses, r.MaxUses-r.CurrentUses)
			resList = append(resList, fmt.Sprintf("• **%s:** %s *(Recarrega em %s)*", r.Name, pips, r.ResetOn))
		}
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   "⏳ Recursos & Usos Diários",
			Value:  strings.Join(resList, "\n"),
			Inline: false,
		})
	}

	// Resumo de Slots de Magia se houver
	if len(sheet.Spellcasting.Slots) > 0 {
		var slotLines []string
		// Ordenar níveis de magia
		var lvls []int
		for k := range sheet.Spellcasting.Slots {
			if n, err := strconv.Atoi(k); err == nil {
				lvls = append(lvls, n)
			}
		}
		sort.Ints(lvls)
		for _, l := range lvls {
			slot := sheet.Spellcasting.Slots[fmt.Sprintf("%d", l)]
			if slot.Max > 0 {
				slotLines = append(slotLines, fmt.Sprintf("**%dº Nível:** %s", l, renderSlotPips(slot.Max, slot.Used)))
			}
		}
		if len(slotLines) > 0 {
			fields = append(fields, &discordgo.MessageEmbedField{
				Name:   "✨ Espaços de Magia Diários",
				Value:  strings.Join(slotLines, " | "),
				Inline: false,
			})
		}
	}

	// Ataques principais resumo
	if len(sheet.Attacks) > 0 {
		var atkLines []string
		for i, at := range sheet.Attacks {
			if i >= 3 {
				break
			}
			atkLines = append(atkLines, fmt.Sprintf("• **%s:** `%s` para acertar | `%s`", at.Name, at.AttackBonus, at.Damage))
		}
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   "⚔️ Ataques Rápidos",
			Value:  strings.Join(atkLines, "\n"),
			Inline: false,
		})
	}

	footerText := "Ficha ativa no Discord • Use os botões abaixo para gerenciar PV, magias e rolagens!"
	if sheet.CurrentHP <= 0 {
		footerText = "💀 PERSONAGEM CAÍDO (0 PV)! Faça Salvaguardas contra a Morte ou receba cura urgente!"
	}

	return &discordgo.MessageEmbed{
		Title:       title,
		Description: desc,
		Color:       color,
		Fields:      fields,
		Footer: &discordgo.MessageEmbedFooter{
			Text: footerText,
		},
	}
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// BuildCharacterSkillsEmbed gera a visão detalhada de perícias, salvaguardas e proficiências da ficha
func BuildCharacterSkillsEmbed(sheet *character.CharacterSheet) *discordgo.MessageEmbed {
	title := fmt.Sprintf("🎯 Perícias & Salvaguardas — %s", sheet.Name)

	fields := []*discordgo.MessageEmbedField{
		{
			Name: "📊 Atributos & Modificadores",
			Value: fmt.Sprintf("`FOR %2d (%+d)` • `DES %2d (%+d)` • `CON %2d (%+d)`\n`INT %2d (%+d)` • `SAB %2d (%+d)` • `CAR %2d (%+d)`",
				sheet.Attributes.STR.Score, sheet.Attributes.STR.Modifier,
				sheet.Attributes.DEX.Score, sheet.Attributes.DEX.Modifier,
				sheet.Attributes.CON.Score, sheet.Attributes.CON.Modifier,
				sheet.Attributes.INT.Score, sheet.Attributes.INT.Modifier,
				sheet.Attributes.WIS.Score, sheet.Attributes.WIS.Modifier,
				sheet.Attributes.CHA.Score, sheet.Attributes.CHA.Modifier),
			Inline: false,
		},
	}

	// 1. Salvaguardas
	var savesStr []string
	saveNames := []struct {
		Key  string
		Name string
	}{
		{"str", "FOR"}, {"dex", "DES"}, {"con", "CON"},
		{"int", "INT"}, {"wis", "SAB"}, {"cha", "CAR"},
	}
	for _, sn := range saveNames {
		s, has := sheet.SavingThrows[sn.Key]
		if has && s.Proficient {
			savesStr = append(savesStr, fmt.Sprintf("⭐ **%s:** `%+d`", sn.Name, s.Modifier))
		} else {
			mod := sheet.GetAttributeModifier(sn.Key)
			savesStr = append(savesStr, fmt.Sprintf("%s: `%+d`", sn.Name, mod))
		}
	}
	fields = append(fields, &discordgo.MessageEmbedField{
		Name:   "🛡️ Salvaguardas (Testes de Resistência)",
		Value:  strings.Join(savesStr, " • "),
		Inline: false,
	})

	// 2. Sentidos & Passivas
	fields = append(fields, &discordgo.MessageEmbedField{
		Name: "👁️ Sentidos & Passivas",
		Value: fmt.Sprintf("• **Percepção Passiva:** `%d`\n• **Investigação Passiva:** `%d`\n• **Intuição Passiva:** `%d`",
			sheet.PassivePerception, 10+sheet.Attributes.INT.Modifier, 10+sheet.Attributes.WIS.Modifier),
		Inline: true,
	})

	// 3. 18 Perícias Canônicas do D&D 5e
	canonicalSkills := []struct {
		Name     string
		Attr     string
		AttrKey  string
	}{
		{"Acrobacia", "DES", "dex"},
		{"Adestrar Animais", "SAB", "wis"},
		{"Arcanismo", "INT", "int"},
		{"Atletismo", "FOR", "str"},
		{"Enganação", "CAR", "cha"},
		{"História", "INT", "int"},
		{"Intuição", "SAB", "wis"},
		{"Intimidação", "CAR", "cha"},
		{"Investigação", "INT", "int"},
		{"Medicina", "SAB", "wis"},
		{"Natureza", "INT", "int"},
		{"Percepção", "SAB", "wis"},
		{"Atuação", "CAR", "cha"},
		{"Persuasão", "CAR", "cha"},
		{"Religião", "INT", "int"},
		{"Prestidigitação", "DES", "dex"},
		{"Furtividade", "DES", "dex"},
		{"Sobrevivência", "SAB", "wis"},
	}

	var skillsCol1 []string
	var skillsCol2 []string

	for i, csk := range canonicalSkills {
		sk, exists := sheet.Skills[csk.Name]
		tag := "▫️"
		mod := sheet.GetAttributeModifier(csk.AttrKey)
		if exists {
			mod = sk.Modifier
			if sk.Expertise {
				tag = "🌟"
			} else if sk.Proficient {
				tag = "⭐"
			}
		}
		line := fmt.Sprintf("%s **%s** (%s): `%+d`", tag, csk.Name, csk.Attr, mod)
		if i < 9 {
			skillsCol1 = append(skillsCol1, line)
		} else {
			skillsCol2 = append(skillsCol2, line)
		}
	}

	fields = append(fields, &discordgo.MessageEmbedField{
		Name:   "🎯 Perícias (A - I)",
		Value:  strings.Join(skillsCol1, "\n"),
		Inline: true,
	})
	fields = append(fields, &discordgo.MessageEmbedField{
		Name:   "🎯 Perícias (J - Z)",
		Value:  strings.Join(skillsCol2, "\n"),
		Inline: true,
	})

	// 4. Proficiências de Ferramentas, Idiomas, Armas e Armaduras
	var profStrings []string
	if len(sheet.Proficiencies.Languages) > 0 {
		profStrings = append(profStrings, fmt.Sprintf("🗣️ **Idiomas:** %s", strings.Join(sheet.Proficiencies.Languages, ", ")))
	}
	if len(sheet.Proficiencies.Tools) > 0 {
		profStrings = append(profStrings, fmt.Sprintf("🛠️ **Ferramentas:** %s", strings.Join(sheet.Proficiencies.Tools, ", ")))
	}
	if len(sheet.Proficiencies.Armor) > 0 {
		profStrings = append(profStrings, fmt.Sprintf("🛡️ **Armaduras:** %s", strings.Join(sheet.Proficiencies.Armor, ", ")))
	}
	if len(sheet.Proficiencies.Weapons) > 0 {
		profStrings = append(profStrings, fmt.Sprintf("⚔️ **Armas:** %s", strings.Join(sheet.Proficiencies.Weapons, ", ")))
	}

	if len(profStrings) > 0 {
		profText := truncateString(strings.Join(profStrings, "\n"), 1000)
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   "📜 Proficiências & Treinamento",
			Value:  profText,
			Inline: false,
		})
	}

	return &discordgo.MessageEmbed{
		Title:  title,
		Color:  ColorDNDBlue,
		Fields: fields,
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Legenda: ⭐ = Proficiente | 🌟 = Especialização | ▫️ = Normal",
		},
	}
}

// BuildCharacterAttributesEmbed mantém compatibilidade chamando a visão de perícias e atributos
func BuildCharacterAttributesEmbed(sheet *character.CharacterSheet) *discordgo.MessageEmbed {
	return BuildCharacterSkillsEmbed(sheet)
}

// BuildCharacterAttacksEmbed gera a visão detalhada da tabela oficial de ataques e combate
func BuildCharacterAttacksEmbed(sheet *character.CharacterSheet) *discordgo.MessageEmbed {
	title := fmt.Sprintf("⚔️ Ataques & Conjuração — %s", sheet.Name)
	desc := "Tabela oficial de ataques, bônus de acerto e dados de dano extraídos da ficha."

	var fields []*discordgo.MessageEmbedField

	// Resumo de combate
	advStatus := "⚖️ Rolagem Normal"
	if sheet.GetAdvantageMode() == "advantage" {
		advStatus = "✨ **Com VANTAGEM ativada!** (Rola 2d20 e usa o maior)"
	} else if sheet.GetAdvantageMode() == "disadvantage" {
		advStatus = "💀 **Com DESVANTAGEM ativada!** (Rola 2d20 e usa o menor)"
	}

	fields = append(fields, &discordgo.MessageEmbedField{
		Name: "🛡️ Defesa & Movimento",
		Value: fmt.Sprintf("• **Modo de Rolagem:** %s\n• **Classe de Armadura (CA):** `%d`\n• **Iniciativa:** `%+d`\n• **Deslocamento:** `%s`\n• **Bônus de Proficiência:** `+%d`",
			advStatus, sheet.ArmorClass, sheet.Initiative, sheet.Speed, sheet.ProficiencyBonus),
		Inline: false,
	})

	if len(sheet.Attacks) == 0 {
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   "Nenhum Ataque Cadastrado",
			Value:  "Esta ficha ainda não possui armas ou ataques configurados em *Ataques e Conjuração*.",
			Inline: false,
		})
	} else {
		for _, atk := range sheet.Attacks {
			val := fmt.Sprintf("🎯 **Bônus de Ataque:** `%s`\n💥 **Dano / Tipo:** `%s`", atk.AttackBonus, atk.Damage)
			if atk.Notes != "" {
				val += fmt.Sprintf("\n📝 *%s*", atk.Notes)
			}
			fields = append(fields, &discordgo.MessageEmbedField{
				Name:   fmt.Sprintf("⚔️ %s", atk.Name),
				Value:  val,
				Inline: true,
			})
		}
	}

	// Se a classe for Monge, exibe características de combate marcial
	if strings.Contains(strings.ToLower(sheet.Class), "monk") || strings.Contains(strings.ToLower(sheet.Class), "monge") {
		fields = append(fields, &discordgo.MessageEmbedField{
			Name: "🥋 Recursos Marciais de Monge",
			Value: "• **Defesa Sábia:** CA = 10 + Mod DES + Mod SAB (sem armadura)\n" +
				"• **Artes Marciais:** Pode usar DES para ataques desarmados e armas de monge\n" +
				"• **Enxurrada de Golpes:** Gasta 1 Foco para 2 ataques desarmados como Ação Bônus\n" +
				"• **Defesa Paciente:** Gasta 1 Foco para Esquivar como Ação Bônus\n" +
				"• **Passo do Vento:** Gasta 1 Foco para Desengajar ou Disparar como Ação Bônus",
			Inline: false,
		})
	}

	return &discordgo.MessageEmbed{
		Title:       title,
		Description: desc,
		Color:       ColorDNDRed,
		Fields:      fields,
		Footer: &discordgo.MessageEmbedFooter{
			Text: "D&D 5e • Use os botões abaixo para rolar ataques automaticamente!",
		},
	}
}

// BuildCharacterSpellsEmbed gera a aba de conjuração e grimório
func BuildCharacterSpellsEmbed(sheet *character.CharacterSheet) *discordgo.MessageEmbed {
	title := fmt.Sprintf("✨ Grimório & Espaços de Magia — %s", sheet.Name)

	if sheet.Spellcasting.Ability == "" && len(sheet.Spellcasting.Spells) == 0 && len(sheet.Spellcasting.Slots) == 0 {
		return &discordgo.MessageEmbed{
			Title:       title,
			Description: fmt.Sprintf("**%s** não é um conjurador de magias nesta ficha.\nPara adicionar habilidades mágicas ou slots, use `/ficha` ou edite sua classe.", sheet.Name),
			Color:       ColorDNDBlue,
		}
	}

	desc := fmt.Sprintf("🔮 **Atributo de Conjuração:** %s | 🛡️ **CD de Resistência:** `%d` | ⚔️ **Bônus de Ataque:** `+%d`",
		sheet.Spellcasting.Ability, sheet.Spellcasting.SaveDC, sheet.Spellcasting.AttackBonus)

	fields := []*discordgo.MessageEmbedField{}

	// Slots de magia por nível com medidores visuais
	var slotLines []string
	var lvls []int
	for k := range sheet.Spellcasting.Slots {
		if n, err := strconv.Atoi(k); err == nil {
			lvls = append(lvls, n)
		}
	}
	sort.Ints(lvls)

	for _, l := range lvls {
		slot := sheet.Spellcasting.Slots[fmt.Sprintf("%d", l)]
		if slot.Max > 0 {
			slotLines = append(slotLines, fmt.Sprintf("• **%dº Nível:** %s", l, renderSlotPips(slot.Max, slot.Used)))
		}
	}

	if len(slotLines) > 0 {
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   "📊 Espaços de Magia Diários (Slots)",
			Value:  strings.Join(slotLines, "\n"),
			Inline: false,
		})
	}

	// Agrupar magias por nível (0 = Truques, 1, 2, ...)
	spellsByLevel := make(map[int][]string)
	for _, sp := range sheet.Spellcasting.Spells {
		spellsByLevel[sp.Level] = append(spellsByLevel[sp.Level], sp.Name)
	}

	var spellLevels []int
	for l := range spellsByLevel {
		spellLevels = append(spellLevels, l)
	}
	sort.Ints(spellLevels)

	for _, l := range spellLevels {
		list := spellsByLevel[l]
		lvlName := fmt.Sprintf("%dº Nível", l)
		if l == 0 {
			lvlName = "Truques (Cantrips)"
		}
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   fmt.Sprintf("📖 %s (%d magias)", lvlName, len(list)),
			Value:  strings.Join(list, " • "),
			Inline: false,
		})
	}

	return &discordgo.MessageEmbed{
		Title:       title,
		Description: desc,
		Color:       ColorDNDBlue,
		Fields:      fields,
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Dica: Use /magia usar <nível> para gastar um espaço durante o combate!",
		},
	}
}

// BuildCharacterFeaturesEmbed gera a aba de características e habilidades com suporte a paginação
func BuildCharacterFeaturesEmbed(sheet *character.CharacterSheet, pageOptional ...int) *discordgo.MessageEmbed {
	page := 0
	if len(pageOptional) > 0 && pageOptional[0] >= 0 {
		page = pageOptional[0]
	}

	totalFeats := len(sheet.Features)
	pageSize := 4
	totalPages := (totalFeats + pageSize - 1) / pageSize
	if totalPages == 0 {
		totalPages = 1
	}
	if page >= totalPages {
		page = totalPages - 1
	}

	title := fmt.Sprintf("📜 Habilidades & Traços — %s (Pág. %d/%d)", sheet.Name, page+1, totalPages)
	fields := []*discordgo.MessageEmbedField{}

	// Recursos limitados
	if len(sheet.LimitedUses) > 0 {
		var resList []string
		for _, r := range sheet.LimitedUses {
			pips := renderSlotPips(r.MaxUses, r.MaxUses-r.CurrentUses)
			resList = append(resList, fmt.Sprintf("• **%s:** %s `(%d/%d)` *(Recarrega em %s)*",
				r.Name, pips, r.CurrentUses, r.MaxUses, r.ResetOn))
		}
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   "⏳ Recursos da Ficha (Gaste com 1 clique nos botões abaixo)",
			Value:  strings.Join(resList, "\n"),
			Inline: false,
		})
	}

	// Características e Traços paginados
	if totalFeats > 0 {
		start := page * pageSize
		end := start + pageSize
		if end > totalFeats {
			end = totalFeats
		}

		for _, f := range sheet.Features[start:end] {
			val := f.Description
			if val == "" {
				val = "Característica ativa do personagem."
			}
			if len(val) > 400 {
				val = truncateString(val, 400)
			}

			sourceStr := ""
			if f.Source != "" {
				sourceStr = fmt.Sprintf(" *(%s)*", f.Source)
			}

			fields = append(fields, &discordgo.MessageEmbedField{
				Name:   fmt.Sprintf("✨ %s%s", f.Name, sourceStr),
				Value:  val,
				Inline: false,
			})
		}
	} else {
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   "Nenhuma habilidade cadastrada",
			Value:  "Você pode importar do seu PDF ou cadastrar habilidades adicionais.",
			Inline: false,
		})
	}

	footerText := fmt.Sprintf("Página %d de %d • %d habilidades catalogadas • Use ◀ e ▶ abaixo para folhear!", page+1, totalPages, totalFeats)

	return &discordgo.MessageEmbed{
		Title:  title,
		Color:  ColorDNDGold,
		Fields: fields,
		Footer: &discordgo.MessageEmbedFooter{
			Text: footerText,
		},
	}
}

// BuildCharacterInventoryEmbed gera a aba de inventário e equipamentos
func BuildCharacterInventoryEmbed(sheet *character.CharacterSheet) *discordgo.MessageEmbed {
	title := fmt.Sprintf("🎒 Inventário & Armas — %s", sheet.Name)

	fields := []*discordgo.MessageEmbedField{
		{
			Name: "💰 Bolsa de Moedas",
			Value: fmt.Sprintf("🟡 **%d PO** (Ouro) | ⚪ **%d PP** (Prata) | 🟤 **%d PC** (Cobre)\n🔵 **%d PE** (Electro) | ⚪ **%d PL** (Platina)",
				sheet.Coins.GP, sheet.Coins.SP, sheet.Coins.CP, sheet.Coins.EP, sheet.Coins.PP),
			Inline: false,
		},
	}

	// Ataques e Armas
	if len(sheet.Attacks) > 0 {
		var atkLines []string
		for _, at := range sheet.Attacks {
			atkLines = append(atkLines, fmt.Sprintf("• **%s:** `%s` no acerto | `%s` de dano", at.Name, at.AttackBonus, at.Damage))
		}
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   "⚔️ Armas & Ataques",
			Value:  strings.Join(atkLines, "\n"),
			Inline: false,
		})
	}

	// Equipamentos
	if len(sheet.Inventory) > 0 {
		var itemLines []string
		for _, it := range sheet.Inventory {
			itemLines = append(itemLines, fmt.Sprintf("• `%dx` %s", it.Quantity, it.Name))
		}
		// Agrupar se muito longo
		fullStr := strings.Join(itemLines, "\n")
		if len(fullStr) > 1024 {
			fullStr = fullStr[:1020] + "..."
		}
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   fmt.Sprintf("📦 Equipamentos (%d itens)", len(sheet.Inventory)),
			Value:  fullStr,
			Inline: false,
		})
	} else {
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   "📦 Equipamentos",
			Value:  "Nenhum item cadastrado no inventário.",
			Inline: false,
		})
	}

	return &discordgo.MessageEmbed{
		Title:  title,
		Color:  ColorDNDGreen,
		Fields: fields,
	}
}

// BuildCharacterActionComponents cria as linhas de botões e abas interativas do Discord
func BuildCharacterActionComponents(sheet *character.CharacterSheet, activeTab string, pageOptional ...int) []discordgo.MessageComponent {
	currentPage := 0
	if len(pageOptional) > 0 && pageOptional[0] >= 0 {
		currentPage = pageOptional[0]
	}

	// Linha 1: Abas de navegação (Geral, Ataques, Perícias, Habilidades, Inventário)
	tabOverviewStyle := discordgo.SecondaryButton
	tabAttacksStyle := discordgo.SecondaryButton
	tabSkillsStyle := discordgo.SecondaryButton
	tabFeaturesStyle := discordgo.SecondaryButton
	tabInvStyle := discordgo.SecondaryButton

	switch activeTab {
	case "attacks":
		tabAttacksStyle = discordgo.PrimaryButton
	case "skills", "attributes":
		tabSkillsStyle = discordgo.PrimaryButton
	case "features":
		tabFeaturesStyle = discordgo.PrimaryButton
	case "inventory":
		tabInvStyle = discordgo.PrimaryButton
	default:
		tabOverviewStyle = discordgo.PrimaryButton
	}

	row1 := discordgo.ActionsRow{
		Components: []discordgo.MessageComponent{
			discordgo.Button{
				Label:    "Geral",
				Emoji:    &discordgo.ComponentEmoji{Name: "🛡️"},
				Style:    tabOverviewStyle,
				CustomID: "sheet:tab:overview",
			},
			discordgo.Button{
				Label:    "Ataques",
				Emoji:    &discordgo.ComponentEmoji{Name: "⚔️"},
				Style:    tabAttacksStyle,
				CustomID: "sheet:tab:attacks",
			},
			discordgo.Button{
				Label:    "Perícias",
				Emoji:    &discordgo.ComponentEmoji{Name: "🎯"},
				Style:    tabSkillsStyle,
				CustomID: "sheet:tab:skills",
			},
			discordgo.Button{
				Label:    "Habilidades",
				Emoji:    &discordgo.ComponentEmoji{Name: "📜"},
				Style:    tabFeaturesStyle,
				CustomID: "sheet:tab:features",
			},
			discordgo.Button{
				Label:    "Inventário",
				Emoji:    &discordgo.ComponentEmoji{Name: "🎒"},
				Style:    tabInvStyle,
				CustomID: "sheet:tab:inventory",
			},
		},
	}

	var rows []discordgo.MessageComponent
	rows = append(rows, row1)

	// Linha 2: Contextual dependendo da aba
	if activeTab == "attacks" && len(sheet.Attacks) > 0 {
		advBtn := buildAdvantageButton(sheet)
		var atkButtons []discordgo.MessageComponent
		atkButtons = append(atkButtons, advBtn)

		var extraAtkButtons []discordgo.MessageComponent
		for idx, atk := range sheet.Attacks {
			label := truncateString(atk.Name, 15)
			btn := discordgo.Button{
				Label:    fmt.Sprintf("🎲 %s (%s)", label, atk.AttackBonus),
				Style:    discordgo.DangerButton,
				CustomID: fmt.Sprintf("sheet:atk:roll:%d", idx),
			}
			if len(atkButtons) < 5 {
				atkButtons = append(atkButtons, btn)
			} else if len(extraAtkButtons) < 5 {
				extraAtkButtons = append(extraAtkButtons, btn)
			}
		}
		rows = append(rows, discordgo.ActionsRow{Components: atkButtons})
		if len(extraAtkButtons) > 0 {
			rows = append(rows, discordgo.ActionsRow{Components: extraAtkButtons})
		}
	} else if activeTab == "features" {
		totalFeats := len(sheet.Features)
		pageSize := 4
		totalPages := (totalFeats + pageSize - 1) / pageSize
		if totalPages == 0 {
			totalPages = 1
		}

		prevDisabled := currentPage <= 0
		nextDisabled := currentPage >= totalPages-1

		var navButtons []discordgo.MessageComponent
		navButtons = append(navButtons, discordgo.Button{
			Label:    "◀ Anterior",
			Style:    discordgo.SecondaryButton,
			Disabled: prevDisabled,
			CustomID: fmt.Sprintf("sheet:feat:page:%d", currentPage-1),
		})
		navButtons = append(navButtons, discordgo.Button{
			Label:    fmt.Sprintf("Pág %d de %d", currentPage+1, totalPages),
			Style:    discordgo.SecondaryButton,
			Disabled: true,
			CustomID: "sheet:feat:curr",
		})
		navButtons = append(navButtons, discordgo.Button{
			Label:    "Próxima ▶",
			Style:    discordgo.SecondaryButton,
			Disabled: nextDisabled,
			CustomID: fmt.Sprintf("sheet:feat:page:%d", currentPage+1),
		})

		rows = append(rows, discordgo.ActionsRow{Components: navButtons})
	} else {
		// Linha padrão de ações rápidas de PV
		row2 := discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "-1 HP",
					Style:    discordgo.DangerButton,
					CustomID: "sheet:hp:dmg:1",
				},
				discordgo.Button{
					Label:    "-5 HP",
					Style:    discordgo.DangerButton,
					CustomID: "sheet:hp:dmg:5",
				},
				discordgo.Button{
					Label:    "+1 HP",
					Style:    discordgo.SuccessButton,
					CustomID: "sheet:hp:heal:1",
				},
				discordgo.Button{
					Label:    "+5 HP",
					Style:    discordgo.SuccessButton,
					CustomID: "sheet:hp:heal:5",
				},
				discordgo.Button{
					Label:    "Dano Custom",
					Emoji:    &discordgo.ComponentEmoji{Name: "💀"},
					Style:    discordgo.DangerButton,
					CustomID: "sheet:hp:custom_dmg",
				},
			},
		}
		rows = append(rows, row2)
	}

	// Linha 3: Recursos Limitados com 1 clique (Foco, Uivos, Metabolismo, etc.)
	if len(sheet.LimitedUses) > 0 {
		var resButtons []discordgo.MessageComponent
		for idx, r := range sheet.LimitedUses {
			if idx >= 5 {
				break
			}
			emojiName := "⏳"
			rLow := strings.ToLower(r.Name)
			if strings.Contains(rLow, "foco") {
				emojiName = "⚡"
			} else if strings.Contains(rLow, "uivo") {
				emojiName = "🐺"
			} else if strings.Contains(rLow, "metabol") {
				emojiName = "🧬"
			} else if strings.Contains(rLow, "cura") {
				emojiName = "💚"
			}

			style := discordgo.PrimaryButton
			disabled := false
			if r.CurrentUses <= 0 {
				style = discordgo.SecondaryButton
				disabled = true
			}

			label := fmt.Sprintf("-1 %s (%d/%d)", truncateString(r.Name, 10), r.CurrentUses, r.MaxUses)
			resButtons = append(resButtons, discordgo.Button{
				Label:    label,
				Emoji:    &discordgo.ComponentEmoji{Name: emojiName},
				Style:    style,
				Disabled: disabled,
				CustomID: fmt.Sprintf("sheet:res:use:%d", idx),
			})
		}
		if len(resButtons) > 0 {
			rows = append(rows, discordgo.ActionsRow{Components: resButtons})
		}
	}

	// Linha 4: Descansos, Cura Customizada e Magia
	var restBtns []discordgo.MessageComponent
	restBtns = append(restBtns, discordgo.Button{
		Label:    "Descanso Curto",
		Emoji:    &discordgo.ComponentEmoji{Name: "⛺"},
		Style:    discordgo.SecondaryButton,
		CustomID: "sheet:rest:short",
	})
	restBtns = append(restBtns, discordgo.Button{
		Label:    "Descanso Longo",
		Emoji:    &discordgo.ComponentEmoji{Name: "🛌"},
		Style:    discordgo.SecondaryButton,
		CustomID: "sheet:rest:long",
	})
	restBtns = append(restBtns, discordgo.Button{
		Label:    "Curar Custom",
		Emoji:    &discordgo.ComponentEmoji{Name: "💚"},
		Style:    discordgo.SuccessButton,
		CustomID: "sheet:hp:custom_heal",
	})

	if len(sheet.Spellcasting.Slots) > 0 || len(sheet.Spellcasting.Spells) > 0 {
		restBtns = append(restBtns, discordgo.Button{
			Label:    "Gastar Slot",
			Emoji:    &discordgo.ComponentEmoji{Name: "✨"},
			Style:    discordgo.SecondaryButton,
			CustomID: "sheet:spell:use_menu",
		})
	}

	// Disponibiliza o botão de alternar Vantagem/Desvantagem também nas demais abas
	if activeTab != "attacks" && len(restBtns) < 5 {
		restBtns = append(restBtns, buildAdvantageButton(sheet))
	}

	rows = append(rows, discordgo.ActionsRow{Components: restBtns})

	return rows
}
