package ui_test

import (
	"testing"

	"kenobi-bot/internal/dnd/model"
	"kenobi-bot/internal/ui"
)

func TestEmbeds(t *testing.T) {
	t.Run("BuildHubEmbed constrói embed válido", func(t *testing.T) {
		embed := ui.BuildHubEmbed()
		if embed == nil {
			t.Fatal("BuildHubEmbed() retornou nil")
		}
		if embed.Title == "" {
			t.Error("Embed Hub não possui título")
		}
		if len(embed.Fields) == 0 {
			t.Error("Embed Hub não possui campos informativos")
		}
	})

	t.Run("BuildClassEmbed constrói ficha completa da classe", func(t *testing.T) {
		class := &model.Class{
			ID:       "monk",
			Name:     "Monk",
			NamePT:   "Monge",
			Emoji:    "🥋",
			Color:    0xE9C46A,
			Overview: "Mestre da disciplina física",
			Role:     "Combatente Veloz",
			Traits: model.ClassTraits{
				PrimaryAbility: "Destreza e Sabedoria",
				HitDie:         "d8",
				SavingThrows:   "Força e Destreza",
			},
			Features: []model.ClassFeature{
				{Level: 1, Name: "Artes Marciais", Description: "Dano desarmado"},
			},
		}

		embed := ui.BuildClassEmbed(class)
		if embed == nil {
			t.Fatal("BuildClassEmbed() retornou nil")
		}
		if embed.Color != 0xE9C46A {
			t.Errorf("embed.Color = %x, esperado %x", embed.Color, 0xE9C46A)
		}
		if len(embed.Fields) < 3 {
			t.Errorf("BuildClassEmbed() retornou apenas %d campos, esperado pelo menos 3", len(embed.Fields))
		}
	})

	t.Run("BuildSubclassEmbed constrói ficha da subclasse", func(t *testing.T) {
		class := &model.Class{ID: "monk", Name: "Monk", Emoji: "🥋", Color: 0xE9C46A}
		sub := &model.Subclass{
			ID:          "open_hand",
			Name:        "Warrior of the Open Hand [PHB'24]",
			ClassName:   "Monk",
			UnlockLevel: 3,
			Synopsis:    "Mestre supremo das artes marciais",
			Features: []model.SubclassFeature{
				{Level: 3, Name: "Técnica da Palma Aberta", Description: "Derruba inimigos"},
			},
		}

		embed := ui.BuildSubclassEmbed(sub, class)
		if embed == nil {
			t.Fatal("BuildSubclassEmbed() retornou nil")
		}
		if len(embed.Fields) == 0 {
			t.Error("BuildSubclassEmbed() não incluiu habilidades da subclasse")
		}
	})

	t.Run("BuildSpellEmbed formata propriedades da magia", func(t *testing.T) {
		spell := &model.Spell{
			ID:            "fireball",
			Name:          "Fireball",
			Level:         3,
			School:        "Evocation",
			CastingTime:   "1 Action",
			Range:         "150 feet",
			Components:    "V, S, M",
			Duration:      "Instantaneous",
			Classes:       []string{"Sorcerer", "Wizard"},
			Description:   "A bright streak flashes from your pointing finger...",
			HigherLevels:  "When you cast this spell using a spell slot of 4th level or higher...",
			Concentration: false,
			Ritual:        false,
		}

		embed := ui.BuildSpellEmbed(spell)
		if embed == nil {
			t.Fatal("BuildSpellEmbed() retornou nil")
		}
		if len(embed.Fields) < 3 {
			t.Errorf("BuildSpellEmbed() retornou %d campos, esperado pelo menos 3", len(embed.Fields))
		}
	})

	t.Run("BuildMonsterEmbed formata ficha de combate", func(t *testing.T) {
		monster := &model.Monster{
			ID:              "goblin",
			Name:            "Goblin",
			Type:            "Small Humanoid (Goblinoid)",
			Alignment:       "Neutral Evil",
			ArmorClass:      "15 (Leather Armor, Shield)",
			HitPoints:       "7 (2d6)",
			Speed:           "30 ft.",
			ChallengeRating: "1/4",
			Actions: []model.MonsterAction{
				{Name: "Scimitar", Description: "Melee Weapon Attack: +4 to hit, reach 5 ft., one target. Hit: 5 (1d6 + 2) slashing damage."},
			},
		}

		embed := ui.BuildMonsterEmbed(monster)
		if embed == nil {
			t.Fatal("BuildMonsterEmbed() retornou nil")
		}
		if len(embed.Fields) == 0 {
			t.Error("BuildMonsterEmbed() não incluiu campos de estatísticas")
		}
	})

	t.Run("BuildSpeciesEmbed formata ficha de espécie/raça de Ravenloft", func(t *testing.T) {
		sp := &model.Species{
			ID:           "dhampir",
			Name:         "Dhampir",
			NamePT:       "Dhampiro",
			Category:     "Ravenloft",
			Source:       "Ravenloft - The Horrors Within",
			CreatureType: "Humanoid",
			Size:         "Medium or Small",
			Speed:        "35 ft",
			Darkvision:   "60 ft",
			Overview:     "Criatura entre os vivos e mortos-vivos",
			Traits: []model.SpeciesTrait{
				{Name: "Spider Climb", Description: "Escalada pelas paredes"},
				{Name: "Vampiric Bite", Description: "Mordida vampírica com cura"},
			},
		}

		embed := ui.BuildSpeciesEmbed(sp)
		if embed == nil {
			t.Fatal("BuildSpeciesEmbed() retornou nil")
		}
		if embed.Color != 0x581845 {
			t.Errorf("embed.Color = %x, esperado 0x581845 para Ravenloft", embed.Color)
		}
		if len(embed.Fields) < 4 {
			t.Errorf("BuildSpeciesEmbed() retornou apenas %d campos, esperado pelo menos 4", len(embed.Fields))
		}
	})
}

