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

func TestSuccessRoll(t *testing.T) {
	roller := dice.NewRoller()

	t.Run("Rolar 6d10 com alvo 5 padrão", func(t *testing.T) {
		res, err := roller.RollSuccess("6d10", 5, dice.RollOptions{})
		if err != nil {
			t.Fatalf("Erro inesperado: %v", err)
		}
		if res.Count != 6 {
			t.Errorf("Count = %d, esperado 6", res.Count)
		}
		if res.Sides != 10 {
			t.Errorf("Sides = %d, esperado 10", res.Sides)
		}
		if res.Threshold != 5 {
			t.Errorf("Threshold = %d, esperado 5", res.Threshold)
		}
		if len(res.Rolls) != 6 {
			t.Fatalf("len(Rolls) = %d, esperado 6", len(res.Rolls))
		}
		if res.Successes+res.Failures != 6 {
			t.Errorf("Total de sucessos (%d) + falhas (%d) deve ser 6", res.Successes, res.Failures)
		}

		actualSuccess := 0
		actualFail := 0
		for _, r := range res.Rolls {
			if r >= 5 {
				actualSuccess++
			} else {
				actualFail++
			}
		}
		if res.Successes != actualSuccess {
			t.Errorf("Successes = %d, esperado %d", res.Successes, actualSuccess)
		}
		if res.Failures != actualFail {
			t.Errorf("Failures = %d, esperado %d", res.Failures, actualFail)
		}

		fmtStr := res.Format()
		if !strings.Contains(fmtStr, "Rolagem de Sucessos") {
			t.Errorf("Format() deve conter cabeçalho, obteve: %s", fmtStr)
		}
		if !strings.Contains(fmtStr, "Sucesso") {
			t.Errorf("Format() deve conter Sucesso(s), obteve: %s", fmtStr)
		}
	})

	t.Run("Rolar apenas com quantidade numérica '6' (padrão d10)", func(t *testing.T) {
		res, err := roller.RollSuccess("6", 5, dice.RollOptions{})
		if err != nil {
			t.Fatalf("Erro ao rolar apenas número: %v", err)
		}
		if res.Count != 6 || res.Sides != 10 {
			t.Errorf("Esperado 6d10, obteve %dd%d", res.Count, res.Sides)
		}
		if res.Threshold != 5 {
			t.Errorf("Threshold = %d, esperado 5", res.Threshold)
		}
	})

	t.Run("Rolar 8d6 com alvo explícito na notação '8d6 alvo 4'", func(t *testing.T) {
		res, err := roller.RollSuccess("8d6 alvo 4", 0, dice.RollOptions{})
		if err != nil {
			t.Fatalf("Erro ao rolar com alvo na notação: %v", err)
		}
		if res.Count != 8 || res.Sides != 6 {
			t.Errorf("Esperado 8d6, obteve %dd%d", res.Count, res.Sides)
		}
		if res.Threshold != 4 {
			t.Errorf("Threshold = %d, esperado 4", res.Threshold)
		}
	})

	t.Run("Rolar com reroll em 1 (Sorte Halfling)", func(t *testing.T) {
		res, err := roller.RollSuccess("10d10", 5, dice.RollOptions{RerollOn1: true})
		if err != nil {
			t.Fatalf("Erro inesperado: %v", err)
		}
		// Se algum dado rolou 1 inicialmente, deve ter sido rerrolado
		for _, d := range res.Details {
			if d.InitialRoll == 1 && !d.Rerolled {
				t.Errorf("Dado que tirou 1 deveria ter sido rerrolado")
			}
		}
	})

	t.Run("Validações e rejeições de erro", func(t *testing.T) {
		// Alvo maior que lados
		_, err := roller.RollSuccess("6d6", 7, dice.RollOptions{})
		if err == nil {
			t.Errorf("Esperado erro quando threshold (7) > lados (6)")
		}

		// Alvo menor que 1
		_, err = roller.RollSuccess("6d6", -1, dice.RollOptions{})
		// -1 faz fallback para padrão 5 (válido no d6)
		if err != nil {
			t.Errorf("Threshold <= 0 deve adotar fallback de 5")
		}

		// Notação inválida
		_, err = roller.RollSuccess("abc", 5, dice.RollOptions{})
		if err == nil {
			t.Errorf("Esperado erro para notação inválida 'abc'")
		}
	})

	t.Run("Rolar dados diferentes combinados tipo 3d61d20", func(t *testing.T) {
		res, err := roller.RollSuccess("3d61d20", 5, dice.RollOptions{})
		if err != nil {
			t.Fatalf("Erro ao rolar 3d61d20: %v", err)
		}
		if res.Count != 4 {
			t.Errorf("Count total = %d, esperado 4 (3 de 3d6 + 1 de 1d20)", res.Count)
		}
		if len(res.Groups) != 2 {
			t.Fatalf("len(Groups) = %d, esperado 2", len(res.Groups))
		}
		if res.Groups[0].Count != 3 || res.Groups[0].Sides != 6 {
			t.Errorf("Grupo 0 esperado 3d6, obteve %dd%d", res.Groups[0].Count, res.Groups[0].Sides)
		}
		if res.Groups[1].Count != 1 || res.Groups[1].Sides != 20 {
			t.Errorf("Grupo 1 esperado 1d20, obteve %dd%d", res.Groups[1].Count, res.Groups[1].Sides)
		}
		if res.Successes+res.Failures != 4 {
			t.Errorf("Total de sucessos + falhas deve ser 4")
		}

		fmtStr := res.Format()
		if !strings.Contains(fmtStr, "3d6") || !strings.Contains(fmtStr, "1d20") {
			t.Errorf("Format() deve listar 3d6 e 1d20, obteve: %s", fmtStr)
		}
		if !strings.Contains(fmtStr, "Dados por grupo:") {
			t.Errorf("Format() deve conter seção 'Dados por grupo:', obteve: %s", fmtStr)
		}
	})

	t.Run("Rolar dados diferentes com separadores '3d6 1d20' e '3d6 + 1d20 6'", func(t *testing.T) {
		res1, err := roller.RollSuccess("3d6 1d20", 5, dice.RollOptions{})
		if err != nil {
			t.Fatalf("Erro ao rolar '3d6 1d20': %v", err)
		}
		if res1.Count != 4 || len(res1.Groups) != 2 {
			t.Errorf("Esperado 4 dados e 2 grupos")
		}

		res2, err := roller.RollSuccess("3d6 + 1d20 6", 0, dice.RollOptions{})
		if err != nil {
			t.Fatalf("Erro ao rolar '3d6 + 1d20 6': %v", err)
		}
		if res2.Threshold != 6 {
			t.Errorf("Threshold = %d, esperado 6", res2.Threshold)
		}
	})

	t.Run("Rolar múltiplos dados separados por '+' tipo 4d6+5d20+2d10", func(t *testing.T) {
		res, err := roller.RollSuccess("4d6+5d20+2d10", 5, dice.RollOptions{})
		if err != nil {
			t.Fatalf("Erro ao rolar 4d6+5d20+2d10: %v", err)
		}
		if res.Count != 11 { // 4 + 5 + 2 = 11
			t.Errorf("Count total = %d, esperado 11", res.Count)
		}
		if len(res.Groups) != 3 {
			t.Fatalf("len(Groups) = %d, esperado 3", len(res.Groups))
		}
		if res.Groups[0].Count != 4 || res.Groups[0].Sides != 6 {
			t.Errorf("Grupo 0 esperado 4d6, obteve %dd%d", res.Groups[0].Count, res.Groups[0].Sides)
		}
		if res.Groups[1].Count != 5 || res.Groups[1].Sides != 20 {
			t.Errorf("Grupo 1 esperado 5d20, obteve %dd%d", res.Groups[1].Count, res.Groups[1].Sides)
		}
		if res.Groups[2].Count != 2 || res.Groups[2].Sides != 10 {
			t.Errorf("Grupo 2 esperado 2d10, obteve %dd%d", res.Groups[2].Count, res.Groups[2].Sides)
		}
		if res.Successes+res.Failures != 11 {
			t.Errorf("Total de sucessos + falhas deve ser 11")
		}

		fmtStr := res.Format()
		if !strings.Contains(fmtStr, "4d6") || !strings.Contains(fmtStr, "5d20") || !strings.Contains(fmtStr, "2d10") {
			t.Errorf("Format() deve listar 4d6, 5d20 e 2d10, obteve: %s", fmtStr)
		}
	})

	t.Run("Alvo explícito no final da notação sobrescreve threshold padrão", func(t *testing.T) {
		res1, err := roller.RollSuccess("4d6+5d20 6", 5, dice.RollOptions{})
		if err != nil {
			t.Fatalf("Erro ao rolar '4d6+5d20 6': %v", err)
		}
		if res1.Threshold != 6 {
			t.Errorf("Threshold deveria ser 6 (passado no final), obteve %d", res1.Threshold)
		}

		res2, err := roller.RollSuccess("4d6+5d20+2d10 7", 5, dice.RollOptions{})
		if err != nil {
			t.Fatalf("Erro ao rolar '4d6+5d20+2d10 7': %v", err)
		}
		if res2.Threshold != 7 {
			t.Errorf("Threshold deveria ser 7 (passado no final), obteve %d", res2.Threshold)
		}

		res3, err := roller.RollSuccess("6d10 8", 5, dice.RollOptions{})
		if err != nil {
			t.Fatalf("Erro ao rolar '6d10 8': %v", err)
		}
		if res3.Threshold != 8 {
			t.Errorf("Threshold deveria ser 8 (passado no final), obteve %d", res3.Threshold)
		}

		res4, err := roller.RollSuccess("6 9", 5, dice.RollOptions{})
		if err != nil {
			t.Fatalf("Erro ao rolar '6 9': %v", err)
		}
		if res4.Threshold != 9 {
			t.Errorf("Threshold deveria ser 9 (passado no final), obteve %d", res4.Threshold)
		}
	})

	t.Run("IsSuccessDiceNotation reconhece formatos válidos", func(t *testing.T) {
		valid := []string{"6d10", "8d6", "6", "d10", "10d10", "6d10 r1", "6d10 5", "6d10 alvo 5", "6 5", "3d61d20", "3d6 1d20", "3d6+1d20", "4d6+5d20+2d10"}
		for _, s := range valid {
			if !roller.IsSuccessDiceNotation(s) {
				t.Errorf("IsSuccessDiceNotation deveria ser true para %q", s)
			}
		}

		invalid := []string{"abc", "d", "d0", "d999"}
		for _, s := range invalid {
			if roller.IsSuccessDiceNotation(s) {
				t.Errorf("IsSuccessDiceNotation deveria ser false para %q", s)
			}
		}
	})
}



