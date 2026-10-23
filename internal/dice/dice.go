package dice

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

var (
	// diceRegex captura:
	// Grupo 1: Quantidade de dados (opcional, padrão 1)
	// Grupo 2: Lados do dado (1 a 100)
	// Grupo 3: Flag de keep (k ou kh) (opcional)
	// Grupo 4: Quantidade de dados a manter (opcional)
	// Grupo 5: Sinal do modificador (+ ou -) (opcional)
	// Grupo 6: Valor do modificador (opcional)
	// Grupo 7: Flag explícita de reroll (ex: r1, ro1) (opcional)
	diceRegex = regexp.MustCompile(`^(?i)\s*(\d*)d(\d+)(?:\s*(k(?:h)?)\s*(\d*))?(?:\s*([+-])\s*(\d+))?(?:\s*(r1|ro1))?\s*$`)

	// sucsDiceRegex captura expressões para rolagem de sucessos:
	// Formato 1: "6d10", "8d6" ou "d10"
	// Formato 2: apenas número "6" (default d10)
	// Opcional: flag de reroll "r1"
	// Opcional: alvo explícito como "5", "alvo 5", "dif 6", ">= 5"
	sucsDiceRegex = regexp.MustCompile(`^(?i)\s*(?:(\d*)d(\d+)|(\d+))(?:\s*(r1|ro1))?(?:\s*(?:alvo|dif|limite|>=?)\s*(\d+)|\s+(\d+))?\s*$`)

	ErrInvalidFormat    = errors.New("formato de dado inválido (use ex: d20, 2d10, 3d20k2, 1d8+2, 4d6-1)")
	ErrMaxSides         = errors.New("o dado máximo permitido é d100")
	ErrMaxCount         = errors.New("a quantidade máxima permitida é de 100 dados por rolagem")
	ErrZeroDice         = errors.New("a quantidade de dados e lados deve ser no mínimo 1")
	ErrKeepTooHigh      = errors.New("a quantidade de dados a manter (k) não pode ser maior que o total de dados rolados")
	ErrThresholdTooHigh = errors.New("o limite de sucesso não pode ser maior que o número de lados do dado")
	ErrThresholdTooLow  = errors.New("o limite de sucesso deve ser no mínimo 1")
)

// DieDetail detalha a rolagem individual de cada dado
type DieDetail struct {
	InitialRoll int  // Primeiro resultado rolado
	FinalRoll   int  // Resultado final considerado
	Rerolled    bool // Se foi rerolado por tirar 1
	Kept        bool // Se este dado foi mantido (para regras de keep k/kh)
}

// RollOptions define opções avançadas de rolagem (ex: Sorte Halfling / Great Weapon Fighting)
type RollOptions struct {
	RerollOn1 bool // Se true, rerrola 1s automaticamente uma vez
}

// RollResult armazena todos os detalhes da rolagem comum
type RollResult struct {
	Notation string      // ex: "3d20k2+3"
	Count    int         // ex: 3
	Sides    int         // ex: 20
	Keep     int         // ex: 2 (0 significa todos mantidos)
	Modifier int         // ex: +3
	Rolls    []int       // ex: [18, 5, 12]
	Details  []DieDetail // ex: [{18, 18, false, true}, ...]
	DiceSum  int         // ex: 30 (soma apenas dos dados mantidos)
	Total    int         // ex: 33
}

// DicePoolGroup armazena a rolagem e resultado de um grupo de dados (ex: 3d6 ou 1d20)
type DicePoolGroup struct {
	Notation  string      // ex: "3d6"
	Count     int         // quantidade de dados neste grupo
	Sides     int         // lados por dado neste grupo
	Rolls     []int       // resultados individuais deste grupo
	Details   []DieDetail // detalhamento com rerolls deste grupo
	Successes int         // total de sucessos deste grupo
	Failures  int         // total de falhas deste grupo
}

// SuccessRollResult armazena os detalhes de uma rolagem com contagem de sucessos e falhas
type SuccessRollResult struct {
	Notation  string          // ex: "3d6 + 1d20" ou "6d10"
	Count     int             // quantidade total de dados rolados
	Sides     int             // lados por dado (se homogêneo) ou 0 se heterogêneo
	Threshold int             // valor mínimo para contar como sucesso (>= Threshold)
	Rolls     []int           // todos os resultados individuais
	Details   []DieDetail     // todos os detalhamentos com rerolls
	Groups    []DicePoolGroup // grupos individuais de dados
	Successes int             // total geral de sucessos
	Failures  int             // total geral de falhas
}

// DiceRoller define o contrato para rolagem de dados
type DiceRoller interface {
	Roll(notation string) (*RollResult, error)
	RollWithOptions(notation string, opts RollOptions) (*RollResult, error)
	RollSuccess(notation string, threshold int, opts RollOptions) (*SuccessRollResult, error)
	IsDiceNotation(text string) bool
	IsSuccessDiceNotation(text string) bool
}

// Roller implementa DiceRoller com gerador criptograficamente seguro
type Roller struct{}

// NewRoller cria um novo rolador de dados
func NewRoller() *Roller {
	return &Roller{}
}

// IsDiceNotation verifica rapidamente se uma string é uma notação de dados válida
func (r *Roller) IsDiceNotation(text string) bool {
	return diceRegex.MatchString(text)
}

// IsSuccessDiceNotation verifica se uma string é uma notação válida para rolagem de sucessos
func (r *Roller) IsSuccessDiceNotation(text string) bool {
	groups, _, _, err := parseSuccessNotation(text)
	return err == nil && len(groups) > 0
}

// Roll interpreta a notação e executa a rolagem padrão
func (r *Roller) Roll(notation string) (*RollResult, error) {
	return r.RollWithOptions(notation, RollOptions{})
}

// RollWithOptions interpreta a notação e executa a rolagem com opções (como rerrolar 1s e manter N maiores)
func (r *Roller) RollWithOptions(notation string, opts RollOptions) (*RollResult, error) {
	matches := diceRegex.FindStringSubmatch(notation)
	if matches == nil {
		return nil, ErrInvalidFormat
	}

	countStr := matches[1]
	sidesStr := matches[2]
	keepFlag := matches[3]
	keepStr := matches[4]
	signStr := matches[5]
	modStr := matches[6]
	rerollFlag := matches[7]

	// Se a notação tiver tag explícita "r1" ou "ro1", ativa reroll
	if rerollFlag != "" {
		opts.RerollOn1 = true
	}

	// 1. Quantidade de dados (padrão 1 se omitido, ex: "d20" -> 1)
	count := 1
	if countStr != "" {
		c, err := strconv.Atoi(countStr)
		if err != nil || c < 1 {
			return nil, ErrZeroDice
		}
		if c > 100 {
			return nil, ErrMaxCount
		}
		count = c
	}

	// 2. Número de lados do dado (1 a 100)
	sides, err := strconv.Atoi(sidesStr)
	if err != nil || sides < 1 {
		return nil, ErrZeroDice
	}
	if sides > 100 {
		return nil, ErrMaxSides
	}

	// 3. Quantidade de dados a manter (k / kh)
	keep := 0
	if keepFlag != "" {
		if keepStr != "" {
			k, err := strconv.Atoi(keepStr)
			if err != nil || k < 1 {
				return nil, errors.New("a quantidade de dados a manter (k) deve ser no mínimo 1")
			}
			if k > count {
				return nil, ErrKeepTooHigh
			}
			keep = k
		} else {
			// Se digitou apenas 'k' ou 'kh' sem número (ex: 2d20k), mantém 1
			keep = 1
		}
	}

	// 4. Modificador opcional (+ ou -)
	modifier := 0
	if modStr != "" {
		m, err := strconv.Atoi(modStr)
		if err == nil {
			if signStr == "-" {
				modifier = -m
			} else {
				modifier = m
			}
		}
	}

	// 5. Rolar cada dado de forma aleatória e uniforme no intervalo [1, sides]
	var rolls []int
	var details []DieDetail

	for i := 0; i < count; i++ {
		initial := rollSingleDie(sides)
		final := initial
		rerolled := false

		// Se a opção de rerrolar 1 estiver ativa e o dado tirou 1
		if opts.RerollOn1 && initial == 1 && sides > 1 {
			final = rollSingleDie(sides)
			rerolled = true
		}

		rolls = append(rolls, final)
		details = append(details, DieDetail{
			InitialRoll: initial,
			FinalRoll:   final,
			Rerolled:    rerolled,
			Kept:        true,
		})
	}

	// 6. Aplicar regra de Keep (manter os N maiores valores)
	diceSum := 0
	if keep > 0 && keep < count {
		type indexedRoll struct {
			index int
			val   int
		}
		indexed := make([]indexedRoll, count)
		for i, d := range details {
			indexed[i] = indexedRoll{index: i, val: d.FinalRoll}
		}

		// Ordena decrescente pelo valor rolado
		sort.SliceStable(indexed, func(i, j int) bool {
			return indexed[i].val > indexed[j].val
		})

		// Marca todos como descartados inicialmente
		for i := range details {
			details[i].Kept = false
		}
		// Marca apenas os 'keep' maiores como mantidos
		for i := 0; i < keep; i++ {
			idx := indexed[i].index
			details[idx].Kept = true
		}
	}

	// Somar apenas os dados mantidos
	for _, d := range details {
		if d.Kept {
			diceSum += d.FinalRoll
		}
	}

	total := diceSum + modifier

	return &RollResult{
		Notation: strings.TrimSpace(notation),
		Count:    count,
		Sides:    sides,
		Keep:     keep,
		Modifier: modifier,
		Rolls:    rolls,
		Details:  details,
		DiceSum:  diceSum,
		Total:    total,
	}, nil
}

// rollSingleDie gera um número natural aleatório entre 1 e sides usando crypto/rand
func rollSingleDie(sides int) int {
	if sides <= 1 {
		return 1
	}
	nBig, err := rand.Int(rand.Reader, big.NewInt(int64(sides)))
	if err != nil {
		return 1
	}
	return int(nBig.Int64()) + 1
}

// Format gera a mensagem estilizada para o Discord
func (res *RollResult) Format() string {
	var rollsStr []string
	hasReroll := false
	isKeep := res.Keep > 0 && res.Keep < res.Count

	for _, d := range res.Details {
		isDropped := isKeep && !d.Kept
		if d.Rerolled {
			hasReroll = true
			if isDropped {
				rollsStr = append(rollsStr, fmt.Sprintf("~~%d ➔ %d~~", d.InitialRoll, d.FinalRoll))
			} else {
				rollsStr = append(rollsStr, fmt.Sprintf("~~%d~~ ➔ **%d**", d.InitialRoll, d.FinalRoll))
			}
		} else {
			if isDropped {
				rollsStr = append(rollsStr, fmt.Sprintf("~~%d~~", d.FinalRoll))
			} else {
				rollsStr = append(rollsStr, fmt.Sprintf("**%d**", d.FinalRoll))
			}
		}
	}
	rollsFormatted := fmt.Sprintf("[%s]", strings.Join(rollsStr, ", "))

	rerollTag := ""
	if hasReroll {
		rerollTag = " *(🍀 1 Rerrolado!)*"
	}

	// Caso 1: Apenas 1 dado sem modificador (ex: d20 -> [15] = 15)
	if res.Count == 1 && res.Modifier == 0 {
		return fmt.Sprintf("🎲 Rolando `%s`%s:\nResultado: %s ➔ **Total: %d**",
			res.Notation, rerollTag, rollsFormatted, res.Total)
	}

	// Caso 2: Apenas 1 dado com modificador (ex: 1d8+2 -> [6] + 2 = 8)
	if res.Count == 1 && res.Modifier != 0 {
		modSign := "+"
		modVal := res.Modifier
		if res.Modifier < 0 {
			modSign = "-"
			modVal = -res.Modifier
		}
		return fmt.Sprintf("🎲 Rolando `%s`%s:\nResultado: %s %s %d ➔ **Total: %d**",
			res.Notation, rerollTag, rollsFormatted, modSign, modVal, res.Total)
	}

	// Rótulo da soma dos dados
	sumLabel := fmt.Sprintf("Soma: %d", res.DiceSum)
	if isKeep {
		sumLabel = fmt.Sprintf("%d maiores: %d", res.Keep, res.DiceSum)
	}

	// Caso 3: Múltiplos dados sem modificador (ex: 3d20k2 -> [**18**, ~~5~~, **12**] (2 maiores: 30) = 30)
	if res.Count > 1 && res.Modifier == 0 {
		return fmt.Sprintf("🎲 Rolando `%s`%s:\nDados: %s (%s) ➔ **Total: %d**",
			res.Notation, rerollTag, rollsFormatted, sumLabel, res.Total)
	}

	// Caso 4: Múltiplos dados com modificador (ex: 3d20k2+5 -> [**18**, ~~5~~, **12**] (2 maiores: 30) + 5 = 35)
	modSign := "+"
	modVal := res.Modifier
	if res.Modifier < 0 {
		modSign = "-"
		modVal = -res.Modifier
	}
	return fmt.Sprintf("🎲 Rolando `%s`%s:\nDados: %s (%s) %s %d ➔ **Total: %d**",
		res.Notation, rerollTag, rollsFormatted, sumLabel, modSign, modVal, res.Total)
}

type diceGroupSpec struct {
	Count int
	Sides int
}

func parseSuccessNotation(notation string) ([]diceGroupSpec, int, bool, error) {
	notation = strings.TrimSpace(notation)
	if notation == "" {
		return nil, 0, false, ErrInvalidFormat
	}

	hasReroll := false
	rerollRegex := regexp.MustCompile(`(?i)\b(r1|ro1)\b`)
	if rerollRegex.MatchString(notation) {
		hasReroll = true
		notation = rerollRegex.ReplaceAllString(notation, " ")
	}

	explicitTarget := 0
	targetRegex := regexp.MustCompile(`(?i)(?:alvo|dif|limite|>=?)\s*(\d+)`)
	if match := targetRegex.FindStringSubmatch(notation); match != nil {
		if t, err := strconv.Atoi(match[1]); err == nil && t > 0 {
			explicitTarget = t
		}
		notation = targetRegex.ReplaceAllString(notation, " ")
	}

	// Pré-processa dados colados sem espaço (ex: "3d61d20" -> "3d6 1d20")
	compactDiceRegex := regexp.MustCompile(`(?i)d(100|20|12|10|[2-8])(\d*d)`)
	for compactDiceRegex.MatchString(notation) {
		notation = compactDiceRegex.ReplaceAllString(notation, "d$1 $2")
	}

	dieRegex := regexp.MustCompile(`(?i)(\d*)d(\d+)`)
	indices := dieRegex.FindAllStringSubmatchIndex(notation, -1)

	var groups []diceGroupSpec
	totalCount := 0

	if len(indices) > 0 {
		var remainderBuilder strings.Builder
		lastIdx := 0

		for _, loc := range indices {
			remainderBuilder.WriteString(notation[lastIdx:loc[0]])
			lastIdx = loc[1]

			countStr := ""
			if loc[2] != -1 && loc[3] != -1 {
				countStr = notation[loc[2]:loc[3]]
			}
			sidesStr := notation[loc[4]:loc[5]]

			count := 1
			if countStr != "" {
				c, err := strconv.Atoi(countStr)
				if err != nil || c < 1 {
					return nil, 0, false, ErrZeroDice
				}
				count = c
			}

			sides, err := strconv.Atoi(sidesStr)
			if err != nil || sides < 1 {
				return nil, 0, false, ErrZeroDice
			}
			if sides > 100 {
				return nil, 0, false, ErrMaxSides
			}

			totalCount += count
			if totalCount > 100 {
				return nil, 0, false, ErrMaxCount
			}

			groups = append(groups, diceGroupSpec{Count: count, Sides: sides})
		}
		remainderBuilder.WriteString(notation[lastIdx:])
		remainder := strings.TrimSpace(remainderBuilder.String())

		cleanedRemainder := strings.Trim(remainder, "+,; \t")
		if cleanedRemainder != "" {
			if t, err := strconv.Atoi(cleanedRemainder); err == nil && t > 0 {
				if explicitTarget == 0 {
					explicitTarget = t
				}
			} else {
				return nil, 0, false, ErrInvalidFormat
			}
		}
	} else {
		cleaned := strings.TrimSpace(notation)
		fields := strings.Fields(cleaned)
		if len(fields) == 0 {
			return nil, 0, false, ErrInvalidFormat
		}

		c, err := strconv.Atoi(fields[0])
		if err != nil || c < 1 {
			return nil, 0, false, ErrInvalidFormat
		}
		if c > 100 {
			return nil, 0, false, ErrMaxCount
		}

		if len(fields) > 1 {
			if t, err := strconv.Atoi(fields[1]); err == nil && t > 0 {
				explicitTarget = t
			}
		}

		groups = append(groups, diceGroupSpec{Count: c, Sides: 10})
	}

	return groups, explicitTarget, hasReroll, nil
}

// RollSuccess executa uma rolagem de pool de dados (único ou múltiplos ex: 3d61d20) e contabiliza sucessos e falhas
func (r *Roller) RollSuccess(notation string, threshold int, opts RollOptions) (*SuccessRollResult, error) {
	groupsSpec, explicitTarget, hasReroll, err := parseSuccessNotation(notation)
	if err != nil {
		return nil, err
	}

	if hasReroll {
		opts.RerollOn1 = true
	}

	if explicitTarget > 0 {
		threshold = explicitTarget
	}
	if threshold <= 0 {
		threshold = 5 // Padrão
	}

	if threshold < 1 {
		return nil, ErrThresholdTooLow
	}

	maxSides := 0
	totalCount := 0
	for _, g := range groupsSpec {
		if g.Sides > maxSides {
			maxSides = g.Sides
		}
		totalCount += g.Count
	}

	if threshold > maxSides {
		return nil, fmt.Errorf("o limite de sucesso (%d) não pode ser maior que o maior dado da rolagem (d%d)", threshold, maxSides)
	}

	var allRolls []int
	var allDetails []DieDetail
	var groups []DicePoolGroup
	totalSuccesses := 0
	totalFailures := 0
	var notationParts []string

	for _, g := range groupsSpec {
		groupNotation := fmt.Sprintf("%dd%d", g.Count, g.Sides)
		notationParts = append(notationParts, groupNotation)

		var gRolls []int
		var gDetails []DieDetail
		gSuccesses := 0
		gFailures := 0

		for i := 0; i < g.Count; i++ {
			initial := rollSingleDie(g.Sides)
			final := initial
			rerolled := false

			if opts.RerollOn1 && initial == 1 && g.Sides > 1 {
				final = rollSingleDie(g.Sides)
				rerolled = true
			}

			if final >= threshold {
				gSuccesses++
			} else {
				gFailures++
			}

			detail := DieDetail{
				InitialRoll: initial,
				FinalRoll:   final,
				Rerolled:    rerolled,
				Kept:        true,
			}
			gRolls = append(gRolls, final)
			gDetails = append(gDetails, detail)

			allRolls = append(allRolls, final)
			allDetails = append(allDetails, detail)
		}

		totalSuccesses += gSuccesses
		totalFailures += gFailures

		groups = append(groups, DicePoolGroup{
			Notation:  groupNotation,
			Count:     g.Count,
			Sides:     g.Sides,
			Rolls:     gRolls,
			Details:   gDetails,
			Successes: gSuccesses,
			Failures:  gFailures,
		})
	}

	fullNotation := strings.Join(notationParts, " + ")
	sides := 0
	if len(groupsSpec) == 1 {
		sides = groupsSpec[0].Sides
	}

	return &SuccessRollResult{
		Notation:  fullNotation,
		Count:     totalCount,
		Sides:     sides,
		Threshold: threshold,
		Rolls:     allRolls,
		Details:   allDetails,
		Groups:    groups,
		Successes: totalSuccesses,
		Failures:  totalFailures,
	}, nil
}

// Format gera a mensagem estilizada para o Discord para rolagem de sucessos
func (res *SuccessRollResult) Format() string {
	hasReroll := false
	for _, d := range res.Details {
		if d.Rerolled {
			hasReroll = true
			break
		}
	}

	rerollTag := ""
	if hasReroll {
		rerollTag = " *(🍀 1 Rerrolado!)*"
	}

	formatGroupDice := func(details []DieDetail, threshold int) string {
		var diceStrs []string
		for _, d := range details {
			isSuccess := d.FinalRoll >= threshold
			var str string
			if d.Rerolled {
				if isSuccess {
					str = fmt.Sprintf("~~%d~~ ➔ **%d**", d.InitialRoll, d.FinalRoll)
				} else {
					str = fmt.Sprintf("~~%d~~ ➔ %d", d.InitialRoll, d.FinalRoll)
				}
			} else {
				if isSuccess {
					str = fmt.Sprintf("**%d**", d.FinalRoll)
				} else {
					str = fmt.Sprintf("%d", d.FinalRoll)
				}
			}
			diceStrs = append(diceStrs, str)
		}
		return "[" + strings.Join(diceStrs, ", ") + "]"
	}

	outcome := ""
	if res.Successes > 0 {
		if res.Successes == 1 {
			outcome = "🎉 **Resultado: 1 Sucesso!**"
		} else {
			outcome = fmt.Sprintf("🎉 **Resultado: %d Sucessos!**", res.Successes)
		}
	} else {
		outcome = "💀 **Resultado: Falha! (Nenhum sucesso)**"
	}

	pluralFalha := "Falhas"
	if res.Failures == 1 {
		pluralFalha = "Falha"
	}
	pluralSucesso := "Sucessos"
	if res.Successes == 1 {
		pluralSucesso = "Sucesso"
	}

	// Caso A: Único tipo de dado (ex: 6d10)
	if len(res.Groups) <= 1 {
		diceFormatted := formatGroupDice(res.Details, res.Threshold)
		return fmt.Sprintf("🎯 **Rolagem de Sucessos** (`%s` • Alvo: **≥ %d**)%s:\n"+
			"🎲 Dados: %s\n"+
			"✨ **%d %s** | 💀 **%d %s**\n"+
			"%s",
			res.Notation, res.Threshold, rerollTag,
			diceFormatted,
			res.Successes, pluralSucesso,
			res.Failures, pluralFalha,
			outcome,
		)
	}

	// Caso B: Múltiplos tipos de dados (ex: 3d6 + 1d20)
	var groupLines []string
	for _, g := range res.Groups {
		gDiceFormatted := formatGroupDice(g.Details, res.Threshold)
		gPluralSucesso := "Sucessos"
		if g.Successes == 1 {
			gPluralSucesso = "Sucesso"
		}
		gPluralFalha := "Falhas"
		if g.Failures == 1 {
			gPluralFalha = "Falha"
		}
		groupLines = append(groupLines, fmt.Sprintf("• **%s:** %s ➔ **%d** %s, **%d** %s",
			g.Notation, gDiceFormatted, g.Successes, gPluralSucesso, g.Failures, gPluralFalha))
	}

	return fmt.Sprintf("🎯 **Rolagem de Sucessos** (`%s` • Alvo: **≥ %d**)%s:\n"+
		"🎲 **Dados por grupo:**\n%s\n"+
		"✨ **Total: %d %s** | 💀 **%d %s**\n"+
		"%s",
		res.Notation, res.Threshold, rerollTag,
		strings.Join(groupLines, "\n"),
		res.Successes, pluralSucesso,
		res.Failures, pluralFalha,
		outcome,
	)
}

