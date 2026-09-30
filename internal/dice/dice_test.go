package dice_test

import (
	"strings"
	"testing"

	"kenobi-bot/internal/dice"
)

func TestParseAndRoll(t *testing.T) {
	roller := dice.NewRoller()

	t.Run("Rolar d20 simples sem modificador", func(t *testing.T) {
		res, err := roller.Roll("d20")
		if err != nil {
			t.Fatalf("Erro inesperado ao rolar d20: %v", err)
		}
		if res.Count != 1 {
			t.Errorf("res.Count = %d, esperado 1", res.Count)
		}
		if res.Sides != 20 {
			t.Errorf("res.Sides = %d, esperado 20", res.Sides)
		}
		if res.Modifier != 0 {
			t.Errorf("res.Modifier = %d, esperado 0", res.Modifier)
		}
		if len(res.Rolls) != 1 {
			t.Fatalf("len(res.Rolls) = %d, esperado 1", len(res.Rolls))
		}
		if res.Rolls[0] < 1 || res.Rolls[0] > 20 {
			t.Errorf("Dado rolou %d, fora do intervalo [1, 20]", res.Rolls[0])
		}
		if res.Total != res.Rolls[0] {
			t.Errorf("res.Total = %d, esperado %d", res.Total, res.Rolls[0])
		}
	})

	t.Run("Rolar 2d10 mostrando rolagem individual e soma", func(t *testing.T) {
		res, err := roller.Roll("2d10")
		if err != nil {
			t.Fatalf("Erro inesperado ao rolar 2d10: %v", err)
		}
		if res.Count != 2 {
			t.Errorf("res.Count = %d, esperado 2", res.Count)
		}
		if res.Sides != 10 {
			t.Errorf("res.Sides = %d, esperado 10", res.Sides)
		}
		if len(res.Rolls) != 2 {
			t.Fatalf("len(res.Rolls) = %d, esperado 2", len(res.Rolls))
		}
		for i, roll := range res.Rolls {
			if roll < 1 || roll > 10 {
				t.Errorf("Dado #%d rolou %d, fora de [1, 10]", i+1, roll)
			}
		}
		diceSum := res.Rolls[0] + res.Rolls[1]
		if res.DiceSum != diceSum {
			t.Errorf("res.DiceSum = %d, esperado %d", res.DiceSum, diceSum)
		}
		if res.Total != diceSum {
			t.Errorf("res.Total = %d, esperado %d", res.Total, diceSum)
		}
	})

	t.Run("Rolar 1d8+2 com modificador positivo", func(t *testing.T) {
		res, err := roller.Roll("1d8+2")
		if err != nil {
			t.Fatalf("Erro inesperado ao rolar 1d8+2: %v", err)
		}
		if res.Count != 1 {
			t.Errorf("res.Count = %d, esperado 1", res.Count)
		}
		if res.Sides != 8 {
			t.Errorf("res.Sides = %d, esperado 8", res.Sides)
		}
		if res.Modifier != 2 {
			t.Errorf("res.Modifier = %d, esperado 2", res.Modifier)
		}
		if res.Total != res.Rolls[0]+2 {
			t.Errorf("res.Total = %d, esperado %d", res.Total, res.Rolls[0]+2)
		}
	})

	t.Run("Rolar com modificador negativo e espaços (ex: 3d6 - 1)", func(t *testing.T) {
		res, err := roller.Roll("3d6 - 1")
		if err != nil {
			t.Fatalf("Erro inesperado: %v", err)
		}
		if res.Modifier != -1 {
			t.Errorf("res.Modifier = %d, esperado -1", res.Modifier)
		}
		expectedTotal := res.DiceSum - 1
		if res.Total != expectedTotal {
			t.Errorf("res.Total = %d, esperado %d", res.Total, expectedTotal)
		}
	})

	t.Run("Rolar d100 válido", func(t *testing.T) {
		res, err := roller.Roll("d100")
		if err != nil {
			t.Fatalf("Erro inesperado ao rolar d100: %v", err)
		}
		if res.Sides != 100 {
			t.Errorf("res.Sides = %d, esperado 100", res.Sides)
		}
	})

	t.Run("Rejeitar dados acima de d100 (ex: d101)", func(t *testing.T) {
		_, err := roller.Roll("d101")
		if err == nil {
			t.Errorf("Esperado erro ao tentar rolar d101 (limite é d100)")
		}
	})

	t.Run("Rejeitar quantidade de dados excessiva (ex: 101d6)", func(t *testing.T) {
		_, err := roller.Roll("101d6")
		if err == nil {
			t.Errorf("Esperado erro ao tentar rolar mais de 100 dados")
		}
	})

	t.Run("Rejeitar formato inválido (ex: abc, d0, 0d6)", func(t *testing.T) {
		invalidInputs := []string{"abc", "d0", "0d6", "2d", "d", "5"}
		for _, input := range invalidInputs {
			_, err := roller.Roll(input)
			if err == nil {
				t.Errorf("Esperado erro para entrada inválida %q", input)
			}
		}
	})

	t.Run("Rerrolar 1 uma única vez quando opção RerollOn1 está ativada", func(t *testing.T) {
		opts := dice.RollOptions{RerollOn1: true}
		res, err := roller.RollWithOptions("d20", opts)
		if err != nil {
			t.Fatalf("Erro inesperado: %v", err)
		}
		if len(res.Details) != 1 {
			t.Fatalf("len(res.Details) = %d, esperado 1", len(res.Details))
		}
		// Se o dado inicial foi 1, ele deve marcar Rerolled como true
		if res.Details[0].InitialRoll == 1 {
			if !res.Details[0].Rerolled {
				t.Errorf("Dado que tirou 1 deveria ter sido rerolado")
			}
		}
	})

	t.Run("FormatResult exibe dados riscados quando rerolados", func(t *testing.T) {
		res := &dice.RollResult{
			Notation: "d20",
			Count:    1,
			Sides:    20,
			Modifier: 0,
			Rolls:    []int{15},
			DiceSum:  15,
			Total:    15,
			Details: []dice.DieDetail{
				{InitialRoll: 1, FinalRoll: 15, Rerolled: true},
			},
		}

		formatted := res.Format()
		if !strings.Contains(formatted, "~~1~~") {
			t.Errorf("Resultado formatado deveria conter ~~1~~ riscado, obteve: %q", formatted)
		}
		if !strings.Contains(formatted, "15") {
			t.Errorf("Resultado formatado deveria conter o valor final 15")
		}
	})

	t.Run("Rolar 3d20k2 mantendo apenas os 2 maiores valores", func(t *testing.T) {
		res, err := roller.Roll("3d20k2")
		if err != nil {
			t.Fatalf("Erro ao rolar 3d20k2: %v", err)
		}
		if res.Count != 3 {
			t.Errorf("res.Count = %d, esperado 3", res.Count)
		}
		if res.Sides != 20 {
			t.Errorf("res.Sides = %d, esperado 20", res.Sides)
		}
		if res.Keep != 2 {
			t.Errorf("res.Keep = %d, esperado 2", res.Keep)
		}
		if len(res.Rolls) != 3 {
			t.Fatalf("len(res.Rolls) = %d, esperado 3", len(res.Rolls))
		}

		// Encontrar os 2 maiores valores dos dados rolados
		sortedVals := make([]int, 3)
		copy(sortedVals, res.Rolls)
		for i := 0; i < len(sortedVals); i++ {
			for j := i + 1; j < len(sortedVals); j++ {
				if sortedVals[j] > sortedVals[i] {
					sortedVals[i], sortedVals[j] = sortedVals[j], sortedVals[i]
				}
			}
		}
		expectedSum := sortedVals[0] + sortedVals[1]
		if res.DiceSum != expectedSum {
			t.Errorf("res.DiceSum = %d, esperado %d (soma dos 2 maiores: %d + %d dos dados %v)",
				res.DiceSum, expectedSum, sortedVals[0], sortedVals[1], res.Rolls)
		}
		if res.Total != expectedSum {
			t.Errorf("res.Total = %d, esperado %d", res.Total, expectedSum)
		}

		// Verificar que exatamente 2 dados foram marcados como Kept
		keptCount := 0
		for _, d := range res.Details {
			if d.Kept {
				keptCount++
			}
		}
		if keptCount != 2 {
			t.Errorf("Quantidade de dados mantidos = %d, esperado 2", keptCount)
		}

		formatted := res.Format()
		if !strings.Contains(formatted, "2 maiores") {
			t.Errorf("Formatação deveria mencionar '2 maiores', obteve: %s", formatted)
		}
	})

	t.Run("Rolar 4d6k3 com modificador positivo e descartado riscado", func(t *testing.T) {
		res, err := roller.Roll("4d6k3+5")
		if err != nil {
			t.Fatalf("Erro ao rolar 4d6k3+5: %v", err)
		}
		if res.Count != 4 || res.Keep != 3 || res.Modifier != 5 {
			t.Errorf("Parâmetros incorretos: Count=%d, Keep=%d, Mod=%d", res.Count, res.Keep, res.Modifier)
		}

		if res.Total != res.DiceSum+5 {
			t.Errorf("res.Total (%d) != res.DiceSum (%d) + 5", res.Total, res.DiceSum)
		}

		// 1 dado deve estar riscado (descartado)
		formatted := res.Format()
		if !strings.Contains(formatted, "~~") {
			t.Errorf("Deveria conter dado descartado com strikethrough '~~': %s", formatted)
		}
	})

	t.Run("Rejeitar k maior que a quantidade de dados (ex: 2d20k3)", func(t *testing.T) {
		_, err := roller.Roll("2d20k3")
		if err == nil {
			t.Errorf("Esperado erro ao tentar manter mais dados do que o total rolado")
		}
	})

	t.Run("Rejeitar k zero (ex: 3d20k0)", func(t *testing.T) {
		_, err := roller.Roll("3d20k0")
		if err == nil {
			t.Errorf("Esperado erro para k0")
		}
	})
}

