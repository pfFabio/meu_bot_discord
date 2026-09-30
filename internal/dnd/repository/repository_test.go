package repository_test

import (
	"testing"

	"kenobi-bot/internal/dnd/model"
	"kenobi-bot/internal/dnd/repository"
)

func sampleRepository() repository.DNDRepository {
	classes := []model.Class{
		{
			ID:     "monk",
			Name:   "Monk",
			NamePT: "Monge",
			Subclasses: []model.Subclass{
				{ID: "open_hand", Name: "Warrior of the Open Hand", ClassName: "Monk", UnlockLevel: 3},
				{ID: "shadow", Name: "Warrior of Shadow", ClassName: "Monk", UnlockLevel: 3},
			},
		},
		{
			ID:     "wizard",
			Name:   "Wizard",
			NamePT: "Mago",
			Subclasses: []model.Subclass{
				{ID: "evoker", Name: "Evoker", ClassName: "Wizard", UnlockLevel: 3},
			},
		},
	}

	spells := []model.Spell{
		{ID: "fireball", Name: "Fireball", Level: 3, School: "Evocation", Classes: []string{"Wizard", "Sorcerer"}},
		{ID: "magic_missile", Name: "Magic Missile", Level: 1, School: "Evocation", Classes: []string{"Wizard", "Sorcerer"}},
		{ID: "shield", Name: "Shield", Level: 1, School: "Abjuration", Classes: []string{"Wizard", "Sorcerer"}},
		{ID: "light", Name: "Light", Level: 0, School: "Evocation", Classes: []string{"Cleric", "Wizard"}},
	}

	monsters := []model.Monster{
		{ID: "goblin", Name: "Goblin", Type: "Humanoid", ChallengeRating: "1/4", HitPoints: "7 (2d6)"},
		{ID: "red_dragon", Name: "Adult Red Dragon", Type: "Dragon", ChallengeRating: "17", HitPoints: "256 (19d12 + 133)"},
		{ID: "aboleth", Name: "Aboleth", Type: "Aberration", ChallengeRating: "10", HitPoints: "150 (20d10 + 40)"},
	}

	items := []model.Item{
		{ID: "greatsword", Name: "Greatsword", Category: "Arma Marcial Corpo a Corpo", Mastery: "Graze", Cost: "50 PO"},
		{ID: "potion_healing", Name: "Potion of Healing", Category: "Poção", Rarity: "Comum", Cost: "50 PO"},
	}

	rules := []model.Rule{
		{ID: "invisible", Name: "Invisible", Summary: "Condição Invisível concede vantagem e os outros têm desvantagem."},
		{ID: "grappled", Name: "Grappled", Summary: "Condição Agarrado reduz deslocamento a 0."},
	}

	species := []model.Species{
		{
			ID:           "dhampir",
			Name:         "Dhampir",
			NamePT:       "Dhampiro",
			Category:     "Ravenloft",
			Source:       "Ravenloft - The Horrors Within",
			CreatureType: "Humanoid",
			Size:         "Medium or Small",
			Speed:        "35 ft",
			Darkvision:   "60 ft",
			Traits: []model.SpeciesTrait{
				{Name: "Vampiric Bite", Description: "Mordida vampírica com dano perfurante e cura."},
			},
		},
		{
			ID:           "elf",
			Name:         "Elf",
			NamePT:       "Elfo",
			Category:     "Common",
			Source:       "PHB 2024",
			CreatureType: "Humanoid",
			Size:         "Medium",
			Speed:        "30 ft",
			Darkvision:   "60 ft",
		},
	}

	return repository.NewMemoryRepository(classes, species, spells, monsters, items, rules)
}

func TestClassRepository(t *testing.T) {
	repo := sampleRepository()

	t.Run("GetClasses retorna todas as classes registradas", func(t *testing.T) {
		classes := repo.GetClasses()
		if len(classes) != 2 {
			t.Errorf("GetClasses() retornou %d classes, esperado 2", len(classes))
		}
	})

	t.Run("GetClass por ID exato", func(t *testing.T) {
		class, ok := repo.GetClass("monk")
		if !ok || class == nil {
			t.Fatalf("GetClass('monk') falhou em encontrar a classe")
		}
		if class.Name != "Monk" {
			t.Errorf("class.Name = %q, esperado 'Monk'", class.Name)
		}
	})

	t.Run("GetClass por Nome em Português ou Inglês (Case-Insensitive)", func(t *testing.T) {
		classPT, okPT := repo.GetClass("Monge")
		if !okPT || classPT == nil {
			t.Errorf("GetClass('Monge') falhou em encontrar por nome em PT")
		}

		classEN, okEN := repo.GetClass("wizard")
		if !okEN || classEN == nil {
			t.Errorf("GetClass('wizard') falhou em encontrar por nome em EN")
		}
	})

	t.Run("GetClass inexistente retorna falso", func(t *testing.T) {
		_, ok := repo.GetClass("astronauta")
		if ok {
			t.Errorf("GetClass('astronauta') deveria retornar falso")
		}
	})

	t.Run("GetSubclasses retorna lista correta", func(t *testing.T) {
		subs := repo.GetSubclasses("monk")
		if len(subs) != 2 {
			t.Errorf("GetSubclasses('monk') retornou %d subclasses, esperado 2", len(subs))
		}
	})

	t.Run("GetSubclass encontra subclasse específica", func(t *testing.T) {
		sub, ok := repo.GetSubclass("monk", "open_hand")
		if !ok || sub == nil {
			t.Fatalf("GetSubclass('monk', 'open_hand') não encontrou a subclasse")
		}
		if sub.Name != "Warrior of the Open Hand" {
			t.Errorf("sub.Name = %q, esperado 'Warrior of the Open Hand'", sub.Name)
		}
	})
}

func TestSpellRepository(t *testing.T) {
	repo := sampleRepository()

	t.Run("SearchSpells com filtro de nível", func(t *testing.T) {
		spells := repo.SearchSpells("", 1, "", "", 10)
		if len(spells) != 2 {
			t.Errorf("SearchSpells(level: 1) retornou %d magias, esperado 2 (Magic Missile, Shield)", len(spells))
		}
	})

	t.Run("SearchSpells com filtro de escola", func(t *testing.T) {
		spells := repo.SearchSpells("", -1, "Abjuration", "", 10)
		if len(spells) != 1 || spells[0].Name != "Shield" {
			t.Errorf("SearchSpells(school: 'Abjuration') retornou %+v, esperado apenas 'Shield'", spells)
		}
	})

	t.Run("SearchSpells com busca de nome parcial", func(t *testing.T) {
		spells := repo.SearchSpells("fire", -1, "", "", 10)
		if len(spells) != 1 || spells[0].Name != "Fireball" {
			t.Errorf("SearchSpells(name: 'fire') retornou %+v, esperado 'Fireball'", spells)
		}
	})

	t.Run("GetSpell encontra magia exata", func(t *testing.T) {
		spell, ok := repo.GetSpell("Fireball")
		if !ok || spell == nil {
			t.Fatalf("GetSpell('Fireball') não encontrou a magia")
		}
		if spell.Level != 3 {
			t.Errorf("spell.Level = %d, esperado 3", spell.Level)
		}
	})
}

func TestMonsterRepository(t *testing.T) {
	repo := sampleRepository()

	t.Run("SearchMonsters por busca textual", func(t *testing.T) {
		monsters := repo.SearchMonsters("dragon", "", "", 10)
		if len(monsters) != 1 || monsters[0].Name != "Adult Red Dragon" {
			t.Errorf("SearchMonsters(name: 'dragon') retornou %+v, esperado 'Adult Red Dragon'", monsters)
		}
	})

	t.Run("SearchMonsters por Tier de CR", func(t *testing.T) {
		monsters := repo.SearchMonsters("", "16+", "", 10)
		if len(monsters) != 1 || monsters[0].Name != "Adult Red Dragon" {
			t.Errorf("SearchMonsters(CR: '16+') retornou %+v, esperado 'Adult Red Dragon'", monsters)
		}
	})

	t.Run("GetMonster encontra monstro específico", func(t *testing.T) {
		m, ok := repo.GetMonster("goblin")
		if !ok || m == nil {
			t.Fatalf("GetMonster('goblin') não encontrou o monstro")
		}
		if m.HitPoints != "7 (2d6)" {
			t.Errorf("m.HitPoints = %q, esperado '7 (2d6)'", m.HitPoints)
		}
	})
}

func TestItemAndRuleRepository(t *testing.T) {
	repo := sampleRepository()

	t.Run("SearchItems com filtro de categoria", func(t *testing.T) {
		items := repo.SearchItems("", "Poção", "", "", 10)
		if len(items) != 1 || items[0].Name != "Potion of Healing" {
			t.Errorf("SearchItems(category: 'Poção') retornou %+v, esperado 'Potion of Healing'", items)
		}
	})

	t.Run("SearchRules por nome ou resumo", func(t *testing.T) {
		rules := repo.SearchRules("invisivel", 10)
		if len(rules) != 1 || rules[0].Name != "Invisible" {
			rulesEN := repo.SearchRules("invisible", 10)
			if len(rulesEN) != 1 {
				t.Errorf("SearchRules('invisible') não retornou a regra")
			}
		}
	})
}

func TestSpeciesRepository(t *testing.T) {
	repo := sampleRepository()

	t.Run("GetSpeciesList retorna lista filtrada por categoria", func(t *testing.T) {
		ravenloftList := repo.GetSpeciesList("Ravenloft")
		if len(ravenloftList) != 1 || ravenloftList[0].ID != "dhampir" {
			t.Errorf("GetSpeciesList('Ravenloft') = %+v, esperado dhampir", ravenloftList)
		}

		allList := repo.GetSpeciesList("all")
		if len(allList) != 2 {
			t.Errorf("GetSpeciesList('all') retornou %d itens, esperado 2", len(allList))
		}
	})

	t.Run("GetSpecies busca por ID e Nome em PT/EN", func(t *testing.T) {
		sp, found := repo.GetSpecies("Dhampir")
		if !found || sp.ID != "dhampir" {
			t.Errorf("GetSpecies('Dhampir') falhou: encontrado=%v", found)
		}

		spPT, foundPT := repo.GetSpecies("dhampiro")
		if !foundPT || spPT.ID != "dhampir" {
			t.Errorf("GetSpecies('dhampiro') falhou: encontrado=%v", foundPT)
		}
	})
}

