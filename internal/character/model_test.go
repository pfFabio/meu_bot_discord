package character

import (
	"testing"
)

func TestCharacterCalculations(t *testing.T) {
	sheet := &CharacterSheet{
		Name:  "Test Wizard",
		Level: 3,
		Attributes: AbilityScores{
			STR: AttributeScore{Score: 8},
			DEX: AttributeScore{Score: 14},
			CON: AttributeScore{Score: 14},
			INT: AttributeScore{Score: 17},
			WIS: AttributeScore{Score: 12},
			CHA: AttributeScore{Score: 10},
		},
		Spellcasting: Spellcasting{
			Ability: "Inteligência",
			Slots: map[string]SpellSlotInfo{
				"1": {Level: 1, Max: 4, Used: 0},
				"2": {Level: 2, Max: 2, Used: 0},
			},
		},
		MaxHP:     20,
		CurrentHP: 20,
	}

	sheet.CalculateDerived()

	// Checa modificadores
	if sheet.Attributes.STR.Modifier != -1 {
		t.Errorf("Esperado STR mod -1, obtido %d", sheet.Attributes.STR.Modifier)
	}
	if sheet.Attributes.DEX.Modifier != 2 {
		t.Errorf("Esperado DEX mod +2, obtido %d", sheet.Attributes.DEX.Modifier)
	}
	if sheet.Attributes.INT.Modifier != 3 {
		t.Errorf("Esperado INT mod +3, obtido %d", sheet.Attributes.INT.Modifier)
	}

	// Bônus de Proficiência nível 3 deve ser +2
	if sheet.ProficiencyBonus != 2 {
		t.Errorf("Esperado Prof Bonus +2, obtido %d", sheet.ProficiencyBonus)
	}

	// CD de Magia = 8 + prof(2) + int(3) = 13
	if sheet.Spellcasting.SaveDC != 13 {
		t.Errorf("Esperado SaveDC 13, obtido %d", sheet.Spellcasting.SaveDC)
	}

	// Bônus de Ataque de Magia = 2 + 3 = 5
	if sheet.Spellcasting.AttackBonus != 5 {
		t.Errorf("Esperado AttackBonus 5, obtido %d", sheet.Spellcasting.AttackBonus)
	}
}

func TestDamageAndHealing(t *testing.T) {
	sheet := &CharacterSheet{
		Name:      "Fighter",
		MaxHP:     30,
		CurrentHP: 30,
		TempHP:    5,
	}

	// Dano de 8: 5 absorvidos por TempHP, 3 causam perda de HP real (30 -> 27)
	lost, isDown := sheet.TakeDamage(8)
	if lost != 3 {
		t.Errorf("Esperado 3 de HP perdido, obtido %d", lost)
	}
	if sheet.CurrentHP != 27 {
		t.Errorf("Esperado 27 HP atual, obtido %d", sheet.CurrentHP)
	}
	if sheet.TempHP != 0 {
		t.Errorf("Esperado 0 TempHP, obtido %d", sheet.TempHP)
	}
	if isDown {
		t.Errorf("Personagem não deveria estar caído")
	}

	// Cura de 5 (27 -> 30, respeita MaxHP)
	healed := sheet.Heal(5)
	if healed != 3 {
		t.Errorf("Esperado 3 pontos curados até o limite de 30, obtido %d", healed)
	}
	if sheet.CurrentHP != 30 {
		t.Errorf("Esperado 30 HP, obtido %d", sheet.CurrentHP)
	}

	// Dano letal
	lost2, isDown2 := sheet.TakeDamage(50)
	if lost2 != 30 {
		t.Errorf("Esperado 30 de HP perdido, obtido %d", lost2)
	}
	if sheet.CurrentHP != 0 {
		t.Errorf("Esperado 0 HP, obtido %d", sheet.CurrentHP)
	}
	if !isDown2 {
		t.Errorf("Personagem deveria estar caído (0 HP)")
	}
}

func TestSpellSlotsAndRest(t *testing.T) {
	sheet := &CharacterSheet{
		MaxHP:     24,
		CurrentHP: 10,
		Level:     3,
		Spellcasting: Spellcasting{
			Slots: map[string]SpellSlotInfo{
				"1": {Level: 1, Max: 4, Used: 0},
			},
		},
		LimitedUses: []LimitedResource{
			{Name: "Canalizar Divindade", MaxUses: 2, CurrentUses: 0, ResetOn: "Descanso Curto"},
			{Name: "Clarão Protetor", MaxUses: 3, CurrentUses: 1, ResetOn: "Descanso Longo"},
		},
	}

	// Usa slot nível 1
	err := sheet.UseSpellSlot(1)
	if err != nil {
		t.Fatalf("Erro inesperado ao gastar slot: %v", err)
	}
	if sheet.Spellcasting.Slots["1"].Used != 1 {
		t.Errorf("Esperado 1 slot usado, obtido %d", sheet.Spellcasting.Slots["1"].Used)
	}

	// Descanso Curto
	sheet.ShortRest()
	if sheet.LimitedUses[0].CurrentUses != 2 {
		t.Errorf("Canalizar Divindade deveria ter resetado para 2 no Descanso Curto")
	}
	if sheet.LimitedUses[1].CurrentUses != 1 {
		t.Errorf("Clarão Protetor NÃO deveria resetar no Descanso Curto")
	}

	// Descanso Longo
	sheet.LongRest()
	if sheet.CurrentHP != 24 {
		t.Errorf("HP deveria estar no máximo após Descanso Longo")
	}
	if sheet.Spellcasting.Slots["1"].Used != 0 {
		t.Errorf("Slots de magia deveriam estar zerados após Descanso Longo")
	}
	if sheet.LimitedUses[1].CurrentUses != 3 {
		t.Errorf("Clarão Protetor deveria ter resetado para 3 no Descanso Longo")
	}
}

func TestAdvantageMode(t *testing.T) {
	sheet := &CharacterSheet{Name: "Samanta Jaci"}

	// Padrão deve ser "normal"
	if mode := sheet.GetAdvantageMode(); mode != "normal" {
		t.Errorf("Esperado modo padrão 'normal', obtido '%s'", mode)
	}

	// 1º Toggle -> "advantage"
	sheet.ToggleAdvantageMode()
	if mode := sheet.GetAdvantageMode(); mode != "advantage" {
		t.Errorf("Esperado 'advantage', obtido '%s'", mode)
	}

	// 2º Toggle -> "disadvantage"
	sheet.ToggleAdvantageMode()
	if mode := sheet.GetAdvantageMode(); mode != "disadvantage" {
		t.Errorf("Esperado 'disadvantage', obtido '%s'", mode)
	}

	// 3º Toggle -> "normal"
	sheet.ToggleAdvantageMode()
	if mode := sheet.GetAdvantageMode(); mode != "normal" {
		t.Errorf("Esperado 'normal', obtido '%s'", mode)
	}
}
