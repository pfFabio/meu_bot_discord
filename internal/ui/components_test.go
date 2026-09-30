package ui_test

import (
	"fmt"
	"testing"

	"github.com/bwmarrin/discordgo"
	"kenobi-bot/internal/dnd/model"
	"kenobi-bot/internal/ui"
)

func TestComponents(t *testing.T) {
	t.Run("BuildHubComponents cria botões interativos", func(t *testing.T) {
		rows := ui.BuildHubComponents()
		if len(rows) == 0 {
			t.Fatal("BuildHubComponents() não retornou linhas de componentes")
		}
	})

	t.Run("BuildClassSelectorComponents respeita limite de 25 opções do Discord", func(t *testing.T) {
		var dummyClasses []model.Class
		for i := 1; i <= 30; i++ {
			dummyClasses = append(dummyClasses, model.Class{
				ID:     fmt.Sprintf("class_%d", i),
				Name:   fmt.Sprintf("Class %d", i),
				NamePT: fmt.Sprintf("Classe %d", i),
			})
		}

		rows := ui.BuildClassSelectorComponents(dummyClasses, &dummyClasses[0])
		if len(rows) == 0 {
			t.Fatal("BuildClassSelectorComponents() retornou 0 linhas")
		}

		// Valida que o primeiro SelectMenu não ultrapassa 25 opções
		if actionRow, ok := rows[0].(discordgo.ActionsRow); ok {
			if selectMenu, ok := actionRow.Components[0].(discordgo.SelectMenu); ok {
				if len(selectMenu.Options) > 25 {
					t.Errorf("SelectMenu de classes possui %d opções (máximo permitido pelo Discord é 25)", len(selectMenu.Options))
				}
			}
		}
	})

	t.Run("BuildSpellFilterComponents cria menu com no máximo 25 opções", func(t *testing.T) {
		var dummySpells []model.Spell
		for i := 1; i <= 40; i++ {
			dummySpells = append(dummySpells, model.Spell{
				ID:    fmt.Sprintf("spell_%d", i),
				Name:  fmt.Sprintf("Spell %d", i),
				Level: 1,
			})
		}

		rows := ui.BuildSpellFilterComponents(dummySpells, 1, "")
		if len(rows) == 0 {
			t.Fatal("BuildSpellFilterComponents() retornou 0 linhas")
		}

		for _, row := range rows {
			if actionRow, ok := row.(discordgo.ActionsRow); ok {
				for _, comp := range actionRow.Components {
					if selectMenu, ok := comp.(discordgo.SelectMenu); ok {
						if len(selectMenu.Options) > 25 {
							t.Errorf("SelectMenu de magias possui %d opções (máximo permitido pelo Discord é 25)", len(selectMenu.Options))
						}
					}
				}
			}
		}
	})

	t.Run("BuildMonsterFilterComponents cria menus de monstros com limite 25", func(t *testing.T) {
		var dummyMonsters []model.Monster
		for i := 1; i <= 50; i++ {
			dummyMonsters = append(dummyMonsters, model.Monster{
				ID:              fmt.Sprintf("monster_%d", i),
				Name:            fmt.Sprintf("Monster %d", i),
				ChallengeRating: "1",
			})
		}

		rows := ui.BuildMonsterFilterComponents(dummyMonsters, "1-4")
		if len(rows) == 0 {
			t.Fatal("BuildMonsterFilterComponents() retornou 0 linhas")
		}

		for _, row := range rows {
			if actionRow, ok := row.(discordgo.ActionsRow); ok {
				for _, comp := range actionRow.Components {
					if selectMenu, ok := comp.(discordgo.SelectMenu); ok {
						if len(selectMenu.Options) > 25 {
							t.Errorf("SelectMenu de monstros possui %d opções (máximo é 25)", len(selectMenu.Options))
						}
					}
				}
			}
		}
	})

	t.Run("BuildSpeciesSelectorComponents respeita limite de 25 opções", func(t *testing.T) {
		var dummySpecies []model.Species
		for i := 1; i <= 30; i++ {
			dummySpecies = append(dummySpecies, model.Species{
				ID:       fmt.Sprintf("sp_%d", i),
				Name:     fmt.Sprintf("Species %d", i),
				NamePT:   fmt.Sprintf("Espécie %d", i),
				Category: "Ravenloft",
			})
		}

		rows := ui.BuildSpeciesSelectorComponents(dummySpecies, &dummySpecies[0], "Ravenloft")
		if len(rows) == 0 {
			t.Fatal("BuildSpeciesSelectorComponents() retornou 0 linhas")
		}

		if actionRow, ok := rows[1].(discordgo.ActionsRow); ok {
			if selectMenu, ok := actionRow.Components[0].(discordgo.SelectMenu); ok {
				if len(selectMenu.Options) > 25 {
					t.Errorf("SelectMenu de espécies possui %d opções (máximo é 25)", len(selectMenu.Options))
				}
			}
		}
	})
}

